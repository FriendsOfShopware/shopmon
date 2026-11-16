// Package delivery adapts go-notifier transports to Shopmon's secret-safe delivery boundary.
package delivery

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	notifier "github.com/shyim/go-notifier"
	"github.com/shyim/go-notifier/transport/discord"
	"github.com/shyim/go-notifier/transport/gotify"
	"github.com/shyim/go-notifier/transport/microsoftteams"
	"github.com/shyim/go-notifier/transport/slack"
	"github.com/shyim/go-notifier/transport/telegram"
)

var ErrInvalidDSN = errors.New("invalid notification DSN; check the provider, credentials and destination")
var ErrDelivery = errors.New("notification delivery failed; check the credentials and destination")

type Client struct {
	factories []notifier.TransportFactoryInterface
}

func NewClient(timeout time.Duration) *Client {
	client := httputil.NewSecretHTTPClient(timeout)
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return newClient(client)
}

func newClient(client *http.Client) *Client {
	return &Client{factories: []notifier.TransportFactoryInterface{
		slack.NewTransportFactory(client), telegram.NewTransportFactory(client),
		discord.NewTransportFactory(client), gotify.NewTransportFactory(client), microsoftteams.NewTransportFactory(client),
	}}
}

// Details contains only explicitly allowlisted, non-credential display fields.
type Details struct {
	Provider string
	Summary  string
}

func Describe(raw string) (Details, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Fragment != "" {
		return Details{}, ErrInvalidDSN
	}
	dsn, err := notifier.NewDSN(raw)
	if err != nil {
		return Details{}, ErrInvalidDSN
	}
	result := Details{Provider: dsn.GetScheme()}
	switch result.Provider {
	case "slack":
		if hook := dsn.GetOption("webhook_url"); hook != "" {
			parsed, err := url.Parse(hook)
			if err != nil || parsed.Scheme != "https" || parsed.Host != "hooks.slack.com" || parsed.User != nil || parsed.Fragment != "" || !strings.HasPrefix(parsed.Path, "/services/") {
				return Details{}, ErrInvalidDSN
			}
			result.Summary = "Incoming webhook"
		} else {
			if dsn.GetOption("channel") == "" {
				return Details{}, ErrInvalidDSN
			}
			result.Summary = dsn.GetOption("channel")
		}
	case "telegram":
		if dsn.GetOption("channel") == "" {
			return Details{}, ErrInvalidDSN
		}
		result.Summary = dsn.GetOption("channel")
	case "discord":
		result.Summary = dsn.GetOption("webhook_id")
	case "gotify":
		if dsn.GetHost() == "default" {
			return Details{}, ErrInvalidDSN
		}
		result.Summary = dsn.GetHost()
	case "microsoftteams":
		result.Summary = "Incoming webhook"
	default:
		return Details{}, ErrInvalidDSN
	}
	return result, nil
}

func (c *Client) transport(raw string) (notifier.TransportInterface, Details, error) {
	details, err := Describe(raw)
	if err != nil {
		return nil, Details{}, err
	}
	dsn, err := notifier.NewDSN(raw)
	if err != nil {
		return nil, Details{}, ErrInvalidDSN
	}
	for _, factory := range c.factories {
		if factory.Supports(dsn) {
			transport, err := factory.Create(dsn)
			// Library errors may embed the DSN, tokens or webhook URL. Never wrap them.
			if err != nil {
				return nil, Details{}, ErrInvalidDSN
			}
			return transport, details, nil
		}
	}
	return nil, Details{}, ErrInvalidDSN
}

func Validate(raw string) (Details, error) {
	_, details, err := newClient(http.DefaultClient).transport(raw)
	return details, err
}

func (c *Client) Post(ctx context.Context, dsn, text string) error {
	transport, details, err := c.transport(dsn)
	if err != nil {
		return err
	}
	// Telegram escapes MarkdownV2 inside go-notifier; Gotify uses plain text.
	// Escape the other providers so shop-controlled text cannot create mentions.
	switch details.Provider {
	case "slack":
		text = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(text)
	case "discord", "microsoftteams":
		text = strings.NewReplacer("\\", "\\\\", "*", "\\*", "_", "\\_", "~", "\\~", "`", "\\`", "[", "\\[", "]", "\\]", "<", "\\<", ">", "\\>", "@", "@\u200b").Replace(text)
	}
	if details.Provider == "discord" {
		text = limitDiscordMessage(text)
	}
	_, err = transport.Send(ctx, notifier.NewChatMessage(text))
	if err != nil {
		return ErrDelivery
	}
	return nil
}

// Identity normalizes DSN encoding for deduplication. It contains credentials
// and must only be used internally (or hashed), never logged or returned.
func Identity(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if u.Scheme == "slack" && u.Query().Get("webhook_url") != "" {
		return "slack:" + u.Query().Get("webhook_url")
	}
	u.RawQuery = u.Query().Encode()
	return u.String()
}

// Discord counts UTF-16 units. Bound the escaped payload, retaining the first
// lines and the final action link when possible.
func limitDiscordMessage(text string) string {
	const limit = 2000
	units := func(s string) int { return len(utf16.Encode([]rune(s))) }
	if units(text) <= limit {
		return text
	}
	tail := ""
	if i := strings.LastIndex(text, "\n"); i >= 0 && strings.Contains(text[i:], "https://") && units(text[i:]) < limit/2 {
		tail = text[i:]
	}
	budget := limit - units(tail) - 2
	prefix := []rune{}
	for _, r := range text {
		n := 1
		if r > 0xffff {
			n = 2
		}
		if budget < n {
			break
		}
		prefix = append(prefix, r)
		budget -= n
	}
	return strings.TrimRight(string(prefix), "\\") + " …" + tail
}
