package notify

import (
	"context"
	"log/slog"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/friendsofshopware/shopmon/api/internal/metrics"
)

// Dispatcher resolves recipients' channel preferences, renders each event into
// their locale, and fans out delivery. It is the single entry point job code
// uses to notify users.
type Dispatcher struct {
	tr       *Translator
	prefs    PreferenceResolver
	locker   lockAcquirer
	channels map[ChannelName]Channel
}

// NewDispatcher wires the default in-app and email channels. mailSvc may be nil
// (email delivery becomes a no-op).
func NewDispatcher(q *queries.Queries, mailSvc mail.Sender) *Dispatcher {
	tr := NewTranslator()
	return &Dispatcher{
		tr:     tr,
		prefs:  queryResolver{q: q},
		locker: queryLocker{q: q},
		channels: map[ChannelName]Channel{
			ChannelInApp: &inAppChannel{q: q},
			ChannelEmail: &emailChannel{mail: mailSvc, tr: tr},
		},
	}
}

// Translator exposes the shared catalog so callers can render text outside the
// dispatch path if needed.
func (d *Dispatcher) Translator() *Translator { return d.tr }

// DispatchResult summarises what a Dispatch actually delivered. Callers that
// persist a "these recipients have been told" marker must consult it: without
// it a failed delivery is indistinguishable from a successful one, and marking
// regardless would suppress the alert forever.
type DispatchResult struct {
	// Delivered counts successful channel sends across all recipients.
	Delivered int
	// Failed counts channel sends that returned an error.
	Failed int
	// Skipped counts channels that were not attempted (email held by the dedup
	// lock, or a channel the recipient has switched off). A skip is not a
	// failure: nothing went wrong and there is nothing to retry.
	Skipped int
}

// Delivery reports whether the event reached at least one recipient and no
// channel failed. Only then is it safe to record the event as delivered.
func (r DispatchResult) Delivery() bool {
	return r.Delivered > 0 && r.Failed == 0
}

// Dispatch delivers an event to all recipients on their resolved channels. It
// never returns an error: per-recipient, per-channel failures are logged so one
// bad delivery cannot abort the rest (or the surrounding scrape). The returned
// DispatchResult reports what was delivered, for callers that record delivery
// state and must not mark a failed alert as sent.
func (d *Dispatcher) Dispatch(ctx context.Context, ev Event, recipients []Recipient) DispatchResult {
	var res DispatchResult
	if len(recipients) == 0 {
		return res
	}

	meta := metaFor(ev.Type)

	// Email dedup is decided once per event (not per recipient): acquire the
	// lock up front and skip the email channel for everyone when it is held.
	emailAllowed := true
	if meta.emailDedupTTL > 0 {
		emailAllowed = d.locker.acquire(ctx, "alert_"+ev.DedupKey, meta.emailDedupTTL)
	}

	for _, r := range recipients {
		channels := d.prefs.Channels(ctx, r, ev, meta.defaultChannels)

		locale := r.Locale
		if locale == "" {
			locale = DefaultLocale
		}
		msg := RenderedMessage{
			Title: d.tr.T(locale, ev.TitleKey, ev.Params),
			Body:  d.tr.T(locale, ev.MessageKey, ev.Params),
		}

		for _, name := range channels {
			if name == ChannelEmail && !emailAllowed {
				res.Skipped++
				continue
			}
			ch, ok := d.channels[name]
			if !ok {
				res.Skipped++
				continue
			}
			if err := ch.Send(ctx, r, ev, msg); err != nil {
				slog.Warn("notify: channel delivery failed",
					"channel", name, "event", ev.Type, "userId", r.ID, "error", err)
				metrics.RecordNotifyDelivery(ctx, metrics.OutcomeError, string(name))
				res.Failed++
				continue
			}
			metrics.RecordNotifyDelivery(ctx, metrics.OutcomeOK, string(name))
			res.Delivered++
		}
	}

	return res
}
