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
	"github.com/friendsofshopware/shopmon/api/internal/delivery"
)

type DestinationConfig struct {
	AppSecret   string
	FrontendURL string
	Poster      DestinationPoster
}

const destinationTimeout = 10 * time.Second

func NewDestinationConfig(appSecret, frontendURL string) DestinationConfig {
	return DestinationConfig{AppSecret: appSecret, FrontendURL: frontendURL, Poster: delivery.NewClient(destinationTimeout)}
}

type DestinationPoster interface {
	Post(context.Context, string, string) error
}
type destinationStore interface {
	GetEventNotificationDestinations(context.Context, queries.GetEventNotificationDestinationsParams) ([]queries.NotificationDestination, error)
}
type destinationTarget struct{ dsn string }

func (t destinationTarget) lockKey(dedupKey string) string {
	sum := sha256.Sum256([]byte(t.dsn))
	return "destination_" + hex.EncodeToString(sum[:]) + "_" + dedupKey
}

type destinationChannel struct {
	store     destinationStore
	poster    DestinationPoster
	appSecret string
	baseURL   string
	tr        *Translator
}

func (c *destinationChannel) targets(ctx context.Context, ev Event) ([]destinationTarget, error) {
	if ev.OrgID == "" {
		return nil, nil
	}
	var envID int64
	if ev.ScopeType == ScopeEnvironment {
		parsed, err := strconv.ParseInt(ev.ScopeID, 10, 32)
		if err == nil && parsed > 0 {
			envID = parsed
		}
	}
	rows, err := c.store.GetEventNotificationDestinations(ctx, queries.GetEventNotificationDestinationsParams{OrganizationID: ev.OrgID, EnvironmentID: int32(envID)})
	if err != nil {
		return nil, fmt.Errorf("load notification destinations: %w", err)
	}
	var result []destinationTarget
	var errs []error
	seen := map[string]bool{}
	for _, row := range rows {
		dsn, err := crypto.Decrypt(row.Dsn, c.appSecret)
		if err != nil {
			errs = append(errs, fmt.Errorf("decrypt destination %s: %w", row.ID, err))
			continue
		}
		if !seen[dsn] {
			result = append(result, destinationTarget{dsn: dsn})
			seen[dsn] = true
		}
	}
	return result, errors.Join(errs...)
}
func (c *destinationChannel) Send(ctx context.Context, dsn string, ev Event) error {
	return c.poster.Post(ctx, dsn, c.render(ev))
}

// Render neutral text; the delivery adapter escapes each provider's markup.
func (c *destinationChannel) render(ev Event) string {
	var b strings.Builder
	b.WriteString(levelEmoji(ev.Level) + " " + c.tr.T(DefaultLocale, ev.TitleKey, ev.Params) + "\n")
	b.WriteString(c.tr.T(DefaultLocale, ev.MessageKey, ev.Params))
	for _, reason := range ev.Reasons {
		b.WriteString("\n• " + c.tr.RenderCheck(DefaultLocale, reason.Key, reason.Params))
	}
	if url := frontendURL(c.baseURL, ev.Link); url != "" {
		b.WriteString("\n" + c.tr.T(DefaultLocale, actionTextKey(ev.Link.Name), nil) + ": " + url)
	}
	return b.String()
}
func levelEmoji(level Level) string {
	switch level {
	case LevelError:
		return "🚨"
	case LevelWarning:
		return "⚠️"
	default:
		return "✅"
	}
}
