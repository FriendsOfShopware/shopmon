package metrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestNormalizeOutcomes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, OutcomeOK, NormalizeScrapeOutcome(OutcomeOK))
	assert.Equal(t, OutcomeAuthError, NormalizeScrapeOutcome(OutcomeAuthError))
	assert.Equal(t, OutcomeDataFetchError, NormalizeScrapeOutcome(OutcomeDataFetchError))
	assert.Equal(t, OutcomeUnknown, NormalizeScrapeOutcome("https://shop.example/admin"))
	assert.Equal(t, OutcomeUnknown, NormalizeScrapeOutcome("user@example.com"))

	assert.Equal(t, OutcomeRateLimited, NormalizeStoreSyncOutcome(OutcomeRateLimited))
	assert.Equal(t, OutcomeUnknown, NormalizeStoreSyncOutcome("SwagPayPal"))

	assert.Equal(t, OutcomeSkipped, NormalizeSitespeedOutcome(OutcomeSkipped))
	assert.Equal(t, OutcomeUnknown, NormalizeSitespeedOutcome("5xx"))

	assert.Equal(t, OutcomeOK, NormalizeMailOutcome(OutcomeOK))
	assert.Equal(t, OutcomeUnknown, NormalizeMailOutcome("bounce"))

	assert.Equal(t, ChannelEmail, NormalizeNotifyChannel(ChannelEmail))
	assert.Equal(t, ChannelInApp, NormalizeNotifyChannel(ChannelInApp))
	assert.Equal(t, OutcomeUnknown, NormalizeNotifyChannel("webhook"))
}

func TestRecordHelpersEmitLowCardinalityAttributes(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	prev := otel.GetMeterProvider()
	otel.SetMeterProvider(mp)
	t.Cleanup(func() {
		otel.SetMeterProvider(prev)
		_ = mp.Shutdown(context.Background())
	})

	Register()

	ctx := context.Background()
	RecordScrapeOutcome(ctx, OutcomeOK)
	RecordScrapeOutcome(ctx, "https://evil.example") // coerced to unknown
	RecordStoreSyncOutcome(ctx, OutcomeRateLimited)
	RecordSitespeedOutcome(ctx, OutcomeError)
	RecordMailSend(ctx, OutcomeOK)
	RecordNotifyDelivery(ctx, OutcomeError, ChannelEmail)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(ctx, &rm))

	counts := collectSumCounts(t, rm)
	assert.Equal(t, int64(1), counts["shopmon.scrape.outcome|outcome=ok"])
	assert.Equal(t, int64(1), counts["shopmon.scrape.outcome|outcome=unknown"])
	assert.Equal(t, int64(1), counts["shopmon.store_sync.outcome|outcome=rate_limited"])
	assert.Equal(t, int64(1), counts["shopmon.sitespeed.outcome|outcome=error"])
	assert.Equal(t, int64(1), counts["shopmon.mail.send|outcome=ok"])
	// Attribute iteration is sorted by key, so channel precedes outcome.
	assert.Equal(t, int64(1), counts["shopmon.notify.delivery|channel=email|outcome=error"])
}

func TestRecordWithoutRegisterIsNoop(t *testing.T) {
	// Isolate from other tests that may have registered instruments.
	mu.Lock()
	scrapeOutcome = nil
	storeSyncOutcome = nil
	sitespeedOutcome = nil
	mailSend = nil
	notifyDelivery = nil
	mu.Unlock()

	assert.NotPanics(t, func() {
		RecordScrapeOutcome(context.Background(), OutcomeOK)
		RecordStoreSyncOutcome(context.Background(), OutcomeError)
		RecordSitespeedOutcome(context.Background(), OutcomeSkipped)
		RecordMailSend(context.Background(), OutcomeError)
		RecordNotifyDelivery(context.Background(), OutcomeOK, ChannelInApp)
	})
}

func collectSumCounts(t *testing.T, rm metricdata.ResourceMetrics) map[string]int64 {
	t.Helper()
	out := make(map[string]int64)
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			sum, ok := m.Data.(metricdata.Sum[int64])
			require.Truef(t, ok, "expected int64 sum for %s", m.Name)
			for _, dp := range sum.DataPoints {
				key := m.Name
				for _, attr := range dp.Attributes.ToSlice() {
					key += "|" + string(attr.Key) + "=" + attr.Value.AsString()
				}
				out[key] = dp.Value
			}
		}
	}
	return out
}
