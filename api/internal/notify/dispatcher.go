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
	// slack is the organization-level channel; nil when not configured.
	slack *slackChannel
}

// NewDispatcher wires the default in-app and email channels, plus the
// organization Slack channel when slackCfg is usable. mailSvc may be nil
// (email delivery becomes a no-op).
func NewDispatcher(q *queries.Queries, mailSvc mail.Sender, slackCfg SlackConfig) *Dispatcher {
	tr := NewTranslator()
	d := &Dispatcher{
		tr:     tr,
		prefs:  queryResolver{q: q},
		locker: queryLocker{q: q},
		channels: map[ChannelName]Channel{
			ChannelInApp: &inAppChannel{q: q},
			ChannelEmail: &emailChannel{mail: mailSvc, tr: tr},
		},
	}
	if slackCfg.AppSecret != "" && slackCfg.Poster != nil {
		d.slack = &slackChannel{
			store:     q,
			poster:    slackCfg.Poster,
			appSecret: slackCfg.AppSecret,
			baseURL:   slackCfg.FrontendURL,
			tr:        tr,
		}
	}
	return d
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
	meta := metaFor(ev.Type)

	// The organization channel does not depend on anyone watching the
	// environment, so it runs before the recipient check.
	d.dispatchSlack(ctx, ev, meta, &res)

	if len(recipients) == 0 {
		return res
	}

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

// dispatchSlack posts the event once to each configured Slack webhook (the
// organization's and the shop's), deduped per webhook with the same window as
// email so a flapping check does not flood the channel.
func (d *Dispatcher) dispatchSlack(ctx context.Context, ev Event, meta eventMeta, res *DispatchResult) {
	if d.slack == nil {
		return
	}

	targets, err := d.slack.targets(ctx, ev)
	if err != nil {
		slog.Warn("notify: failed to load slack webhook", "event", ev.Type, "organizationId", ev.OrgID, "error", err)
		metrics.RecordNotifyDelivery(ctx, metrics.OutcomeError, string(ChannelSlack))
		res.Failed++
	}

	for _, target := range targets {
		lockKey := target.lockPrefix + ev.DedupKey
		if meta.emailDedupTTL > 0 && !d.locker.acquire(ctx, lockKey, meta.emailDedupTTL) {
			res.Skipped++
			continue
		}

		if err := d.slack.Send(ctx, target.url, ev); err != nil {
			// Without the release, the caller's retry would hit the lock, count as
			// skipped, and record the alert as delivered although Slack never got it.
			if meta.emailDedupTTL > 0 {
				d.locker.release(ctx, lockKey)
			}
			slog.Warn("notify: channel delivery failed",
				"channel", ChannelSlack, "event", ev.Type, "organizationId", ev.OrgID, "error", err)
			metrics.RecordNotifyDelivery(ctx, metrics.OutcomeError, string(ChannelSlack))
			res.Failed++
			continue
		}
		metrics.RecordNotifyDelivery(ctx, metrics.OutcomeOK, string(ChannelSlack))
		res.Delivered++
	}
}
