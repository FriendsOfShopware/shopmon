package notify

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/crypto"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/slack"
	"github.com/jackc/pgx/v5"
)

// SlackConfig enables the organization Slack channel. An empty AppSecret
// disables it, since stored webhook URLs cannot be decrypted without it.
type SlackConfig struct {
	AppSecret   string
	FrontendURL string
	Poster      SlackPoster
}

// slackTimeout bounds a webhook post so a slow Slack cannot stall a scrape.
const slackTimeout = 10 * time.Second

// NewSlackConfig builds the production Slack configuration.
func NewSlackConfig(appSecret, frontendURL string) SlackConfig {
	return SlackConfig{AppSecret: appSecret, FrontendURL: frontendURL, Poster: slack.NewClient(slackTimeout)}
}

// SlackPoster delivers a mrkdwn message to an incoming webhook.
type SlackPoster interface {
	Post(ctx context.Context, webhookURL, text string) error
}

// slackWebhookStore is the narrow query surface the channel needs (satisfied
// by *queries.Queries).
type slackWebhookStore interface {
	GetOrganizationSlackWebhook(ctx context.Context, organizationID string) (string, error)
	GetEnvironmentShopSlackWebhook(ctx context.Context, arg queries.GetEnvironmentShopSlackWebhookParams) (string, error)
}

// slackTarget is one webhook an event is posted to. lockPrefix keeps the dedup
// of the organization and shop channels apart, so one failing channel does not
// hold back the other.
type slackTarget struct {
	lockPrefix string
	url        string
}

// slackChannel posts an event once per configured webhook: the organization's
// and, for environment events, the shop's. Unlike the per-recipient channels it
// is addressed by scope, so a channel watched by the whole team gets one
// message regardless of subscriber count.
type slackChannel struct {
	store     slackWebhookStore
	poster    SlackPoster
	appSecret string
	baseURL   string
	tr        *Translator
}

// targets returns the configured webhooks for an event. A shop webhook equal
// to the organization one is dropped so the channel is not posted twice. The
// error reports webhooks that could not be loaded; the others are still
// returned.
func (c *slackChannel) targets(ctx context.Context, ev Event) ([]slackTarget, error) {
	var (
		out  []slackTarget
		errs []error
	)

	if ev.OrgID != "" {
		url, err := c.decrypt(c.store.GetOrganizationSlackWebhook(ctx, ev.OrgID))
		if err != nil {
			errs = append(errs, fmt.Errorf("organization webhook: %w", err))
		} else if url != "" {
			out = append(out, slackTarget{lockPrefix: "slack_", url: url})
		}
	}

	// Scoped by organization too, so an event can only ever reach a channel of
	// the organization it belongs to.
	if ev.ScopeType == ScopeEnvironment && ev.OrgID != "" {
		if envID, err := strconv.ParseInt(ev.ScopeID, 10, 32); err == nil {
			url, err := c.decrypt(c.store.GetEnvironmentShopSlackWebhook(ctx, queries.GetEnvironmentShopSlackWebhookParams{
				ID:             int32(envID),
				OrganizationID: ev.OrgID,
			}))
			switch {
			case err != nil:
				errs = append(errs, fmt.Errorf("shop webhook: %w", err))
			case url != "" && (len(out) == 0 || out[0].url != url):
				out = append(out, slackTarget{lockPrefix: "slack_shop_", url: url})
			}
		}
	}

	return out, errors.Join(errs...)
}

// lockKey dedups an event per webhook. Including the webhook means a channel
// replaced within the dedup window still gets the alert.
func (t slackTarget) lockKey(dedupKey string) string {
	sum := sha256.Sum256([]byte(t.url))
	return t.lockPrefix + hex.EncodeToString(sum[:6]) + "_" + dedupKey
}

// decrypt turns a stored webhook into its URL, or "" when none is configured.
func (c *slackChannel) decrypt(encrypted string, err error) (string, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("load slack webhook: %w", err)
	}
	url, err := crypto.Decrypt(encrypted, c.appSecret)
	if err != nil {
		return "", fmt.Errorf("decrypt slack webhook: %w", err)
	}
	return url, nil
}

func (c *slackChannel) Send(ctx context.Context, webhookURL string, ev Event) error {
	if err := c.poster.Post(ctx, webhookURL, c.render(ev)); err != nil {
		return fmt.Errorf("post slack message: %w", err)
	}
	return nil
}

// render builds the mrkdwn text. Organizations have no locale, so the default
// catalog is used.
func (c *slackChannel) render(ev Event) string {
	locale := DefaultLocale

	var b strings.Builder
	b.WriteString(levelEmoji(ev.Level) + " *" + slack.Escape(c.tr.T(locale, ev.TitleKey, ev.Params)) + "*\n")
	b.WriteString(slack.Escape(c.tr.T(locale, ev.MessageKey, ev.Params)))
	for _, reason := range ev.Reasons {
		b.WriteString("\n• " + slack.Escape(c.tr.RenderCheck(locale, reason.Key, reason.Params)))
	}
	if url := frontendURL(c.baseURL, ev.Link); url != "" {
		b.WriteString("\n<" + url + "|" + slack.Escape(c.tr.T(locale, actionTextKey(ev.Link.Name), nil)) + ">")
	}
	return b.String()
}

func levelEmoji(level Level) string {
	switch level {
	case LevelError:
		return ":rotating_light:"
	case LevelWarning:
		return ":warning:"
	default:
		return ":white_check_mark:"
	}
}
