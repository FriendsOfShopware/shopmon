// Package slack posts messages to Slack incoming webhooks.
package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// webhookHost is the only host incoming webhooks live on. Restricting to it
// keeps a tenant-supplied URL from turning the worker into a request proxy.
const webhookHost = "hooks.slack.com"

var ErrInvalidWebhookURL = errors.New("webhook URL must start with https://hooks.slack.com/")

// ValidateWebhookURL accepts only HTTPS URLs on the Slack webhook host.
func ValidateWebhookURL(raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme != "https" || u.Host != webhookHost || u.User != nil || len(u.Path) <= 1 {
		return ErrInvalidWebhookURL
	}
	return nil
}

// Client posts messages to incoming webhooks.
type Client struct {
	http     *http.Client
	validate func(string) error
}

func NewClient(timeout time.Duration) *Client {
	return &Client{http: &http.Client{
		Timeout: timeout,
		// Webhooks answer directly; a redirect would leave the validated host.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}, validate: ValidateWebhookURL}
}

// Post sends a plain mrkdwn message to the webhook.
func (c *Client) Post(ctx context.Context, webhookURL, text string) error {
	if err := c.validate(webhookURL); err != nil {
		return err
	}

	body, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return fmt.Errorf("marshal slack message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimSpace(webhookURL), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		// The URL is the credential; never let it reach logs via *url.Error.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			return fmt.Errorf("post slack message: %w", urlErr.Err)
		}
		return fmt.Errorf("post slack message: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return fmt.Errorf("slack responded %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
	}
	return nil
}

// Escape makes text safe for mrkdwn, where &, < and > start control sequences.
func Escape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}
