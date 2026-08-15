// Package metrics exposes low-cardinality business-outcome counters for
// Datadog monitors. Instruments are registered after telemetry.Setup installs
// a MeterProvider; until then Record* calls are no-ops.
//
// Attribute sets are intentionally tiny enums. Never add environment URLs,
// shop URLs, email addresses, or extension name lists as attributes.
package metrics

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const meterName = "shopmon"

// Outcome values shared across counters. Keep this set small so Datadog
// monitors stay cheap and cardinality cannot explode.
const (
	OutcomeOK             = "ok"
	OutcomeError          = "error"
	OutcomeAuthError      = "auth_error"
	OutcomeDataFetchError = "data_fetch_error"
	OutcomeRateLimited    = "rate_limited"
	OutcomeSkipped        = "skipped"
	OutcomeUnknown        = "unknown"
)

// Channel values for notify delivery. Matches notify.ChannelName.
const (
	ChannelEmail = "email"
	ChannelInApp = "in_app"
)

var (
	mu               sync.RWMutex
	scrapeOutcome    metric.Int64Counter
	storeSyncOutcome metric.Int64Counter
	sitespeedOutcome metric.Int64Counter
	mailSend         metric.Int64Counter
	notifyDelivery   metric.Int64Counter
	uptimeProbe      metric.Int64Counter
	uptimeTransition metric.Int64Counter
)

// Register creates the shopmon outcome instruments on the global MeterProvider.
// Call it from telemetry.Setup after otel.SetMeterProvider so counters bind to
// the real exporter rather than the default no-op provider.
func Register() {
	m := otel.Meter(meterName)

	scrape, _ := m.Int64Counter("shopmon.scrape.outcome",
		metric.WithDescription("Environment scrape business outcomes"),
		metric.WithUnit("{scrape}"),
	)
	storeSync, _ := m.Int64Counter("shopmon.store_sync.outcome",
		metric.WithDescription("Shopware Store API catalog sync outcomes"),
		metric.WithUnit("{sync}"),
	)
	sitespeed, _ := m.Int64Counter("shopmon.sitespeed.outcome",
		metric.WithDescription("Sitespeed scrape outcomes"),
		metric.WithUnit("{scrape}"),
	)
	mail, _ := m.Int64Counter("shopmon.mail.send",
		metric.WithDescription("Outbound email send attempts"),
		metric.WithUnit("{email}"),
	)
	notify, _ := m.Int64Counter("shopmon.notify.delivery",
		metric.WithDescription("Notification channel delivery attempts"),
		metric.WithUnit("{delivery}"),
	)
	uptime, _ := m.Int64Counter("shopmon.uptime.probe",
		metric.WithDescription("External uptime probe outcomes"),
		metric.WithUnit("{probe}"),
	)
	uptimeTransitions, _ := m.Int64Counter("shopmon.uptime.transition",
		metric.WithDescription("Uptime state transitions; a burst of downs indicates probe-infrastructure trouble"),
		metric.WithUnit("{transition}"),
	)

	mu.Lock()
	defer mu.Unlock()
	scrapeOutcome = scrape
	storeSyncOutcome = storeSync
	sitespeedOutcome = sitespeed
	mailSend = mail
	notifyDelivery = notify
	uptimeProbe = uptime
	uptimeTransition = uptimeTransitions
}

// RecordScrapeOutcome increments shopmon.scrape.outcome.
// Allowed outcomes: ok, error, auth_error, data_fetch_error.
func RecordScrapeOutcome(ctx context.Context, outcome string) {
	add(ctx, scrapeCounter(), "outcome", NormalizeScrapeOutcome(outcome))
}

// RecordStoreSyncOutcome increments shopmon.store_sync.outcome.
// Allowed outcomes: ok, rate_limited, error.
func RecordStoreSyncOutcome(ctx context.Context, outcome string) {
	add(ctx, storeSyncCounter(), "outcome", NormalizeStoreSyncOutcome(outcome))
}

// RecordSitespeedOutcome increments shopmon.sitespeed.outcome.
// Allowed outcomes: ok, error, skipped.
func RecordSitespeedOutcome(ctx context.Context, outcome string) {
	add(ctx, sitespeedCounter(), "outcome", NormalizeSitespeedOutcome(outcome))
}

// RecordMailSend increments shopmon.mail.send.
// Allowed outcomes: ok, error. Never label with recipient addresses.
func RecordMailSend(ctx context.Context, outcome string) {
	add(ctx, mailCounter(), "outcome", NormalizeMailOutcome(outcome))
}

// RecordNotifyDelivery increments shopmon.notify.delivery.
// Allowed outcomes: ok, error. Channel must be a tiny enum (email, in_app).
func RecordNotifyDelivery(ctx context.Context, outcome, channel string) {
	c := notifyCounter()
	if c == nil {
		return
	}
	c.Add(ctx, 1, metric.WithAttributes(
		attribute.String("outcome", NormalizeMailOutcome(outcome)),
		attribute.String("channel", NormalizeNotifyChannel(channel)),
	))
}

// Transition values for uptime state changes.
const (
	TransitionDown = "down"
	TransitionUp   = "up"
)

// RecordUptimeProbe increments shopmon.uptime.probe.
// Allowed outcomes: ok, error.
func RecordUptimeProbe(ctx context.Context, outcome string) {
	add(ctx, uptimeProbeCounter(), "outcome", NormalizeMailOutcome(outcome))
}

// RecordUptimeTransition increments shopmon.uptime.transition.
// Allowed transitions: down, up.
func RecordUptimeTransition(ctx context.Context, transition string) {
	add(ctx, uptimeTransitionCounter(), "transition", NormalizeUptimeTransition(transition))
}

// NormalizeUptimeTransition maps arbitrary strings onto down|up|unknown.
func NormalizeUptimeTransition(transition string) string {
	switch transition {
	case TransitionDown, TransitionUp:
		return transition
	default:
		return OutcomeUnknown
	}
}

// NormalizeScrapeOutcome maps arbitrary strings onto the scrape outcome enum.
func NormalizeScrapeOutcome(outcome string) string {
	switch outcome {
	case OutcomeOK, OutcomeError, OutcomeAuthError, OutcomeDataFetchError:
		return outcome
	default:
		return OutcomeUnknown
	}
}

// NormalizeStoreSyncOutcome maps arbitrary strings onto the store sync enum.
func NormalizeStoreSyncOutcome(outcome string) string {
	switch outcome {
	case OutcomeOK, OutcomeError, OutcomeRateLimited:
		return outcome
	default:
		return OutcomeUnknown
	}
}

// NormalizeSitespeedOutcome maps arbitrary strings onto the sitespeed enum.
func NormalizeSitespeedOutcome(outcome string) string {
	switch outcome {
	case OutcomeOK, OutcomeError, OutcomeSkipped:
		return outcome
	default:
		return OutcomeUnknown
	}
}

// NormalizeMailOutcome maps arbitrary strings onto ok|error|unknown.
func NormalizeMailOutcome(outcome string) string {
	switch outcome {
	case OutcomeOK, OutcomeError:
		return outcome
	default:
		return OutcomeUnknown
	}
}

// NormalizeNotifyChannel maps arbitrary strings onto email|in_app|unknown.
func NormalizeNotifyChannel(channel string) string {
	switch channel {
	case ChannelEmail, ChannelInApp:
		return channel
	default:
		return OutcomeUnknown
	}
}

func add(ctx context.Context, c metric.Int64Counter, key, value string) {
	if c == nil {
		return
	}
	c.Add(ctx, 1, metric.WithAttributes(attribute.String(key, value)))
}

func scrapeCounter() metric.Int64Counter {
	mu.RLock()
	defer mu.RUnlock()
	return scrapeOutcome
}

func storeSyncCounter() metric.Int64Counter {
	mu.RLock()
	defer mu.RUnlock()
	return storeSyncOutcome
}

func sitespeedCounter() metric.Int64Counter {
	mu.RLock()
	defer mu.RUnlock()
	return sitespeedOutcome
}

func mailCounter() metric.Int64Counter {
	mu.RLock()
	defer mu.RUnlock()
	return mailSend
}

func notifyCounter() metric.Int64Counter {
	mu.RLock()
	defer mu.RUnlock()
	return notifyDelivery
}

func uptimeProbeCounter() metric.Int64Counter {
	mu.RLock()
	defer mu.RUnlock()
	return uptimeProbe
}

func uptimeTransitionCounter() metric.Int64Counter {
	mu.RLock()
	defer mu.RUnlock()
	return uptimeTransition
}
