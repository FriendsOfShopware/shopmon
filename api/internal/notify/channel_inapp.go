package notify

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
)

// inAppChannel persists notifications to user_notification. It stores both the
// translation keys+params (for per-viewer rendering in the UI) and the rendered
// title/message (as a fallback for clients that do not yet key-render).
type inAppChannel struct {
	q *queries.Queries
}

func (c *inAppChannel) Name() ChannelName { return ChannelInApp }

func (c *inAppChannel) Send(ctx context.Context, r Recipient, ev Event, msg RenderedMessage) error {
	// Merge reasons into the stored params so the UI can render the "why" list
	// alongside the message, without mutating the caller's map.
	stored := ev.Params
	if len(ev.Reasons) > 0 {
		stored = make(map[string]any, len(ev.Params)+1)
		for k, v := range ev.Params {
			stored[k] = v
		}
		stored["reasons"] = ev.Reasons
	}

	params := []byte("{}")
	if len(stored) > 0 {
		b, err := json.Marshal(stored)
		if err != nil {
			return fmt.Errorf("marshal params: %w", err)
		}
		params = b
	}

	link, err := json.Marshal(ev.Link)
	if err != nil {
		return fmt.Errorf("marshal link: %w", err)
	}

	titleKey := optional(ev.TitleKey)
	messageKey := optional(ev.MessageKey)

	if err := c.q.UpsertNotification(ctx, queries.UpsertNotificationParams{
		UserID:     r.ID,
		Key:        ev.DedupKey,
		Level:      string(ev.Level),
		Title:      msg.Title,
		Message:    msg.Body,
		TitleKey:   titleKey,
		MessageKey: messageKey,
		Params:     params,
		Link:       link,
	}); err != nil {
		return fmt.Errorf("upsert notification: %w", err)
	}
	return nil
}

func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
