package jobs

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestRunEnvironmentJobSetsConsumerSpanAttributes(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "EnvironmentScrape process")

	err := runEnvironmentJob(ctx, 42, func(context.Context) error {
		return nil
	})
	require.NoError(t, err)
	span.End()

	attrs := spanAttributes(span.(sdktrace.ReadOnlySpan))
	assert.Equal(t, int64(42), attrs["environment.id"].AsInt64())
	assert.Equal(t, "ok", attrs["outcome"].AsString())
}

func TestRunEnvironmentJobSetsErrorOutcome(t *testing.T) {
	tp := sdktrace.NewTracerProvider()
	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "SitespeedScrape process")

	jobErr := errors.New("sitespeed failed")
	err := runEnvironmentJob(ctx, 7, func(context.Context) error {
		return jobErr
	})
	assert.ErrorIs(t, err, jobErr)
	span.End()

	attrs := spanAttributes(span.(sdktrace.ReadOnlySpan))
	assert.Equal(t, int64(7), attrs["environment.id"].AsInt64())
	assert.Equal(t, "error", attrs["outcome"].AsString())
}

func TestRunEnvironmentJobAnnotatesParentNotChild(t *testing.T) {
	// Handlers receive the Consumer span in ctx; job body may start Internal
	// children. Attributes must stay on the Consumer entry span.
	tp := sdktrace.NewTracerProvider()
	tracer := tp.Tracer("test")
	ctx, consumer := tracer.Start(context.Background(), "EnvironmentScrape process")

	err := runEnvironmentJob(ctx, 99, func(ctx context.Context) error {
		_, child := tracer.Start(ctx, "environment.scrape")
		child.End()
		return nil
	})
	require.NoError(t, err)
	consumer.End()

	attrs := spanAttributes(consumer.(sdktrace.ReadOnlySpan))
	assert.Equal(t, int64(99), attrs["environment.id"].AsInt64())
	assert.Equal(t, "ok", attrs["outcome"].AsString())
}

func spanAttributes(span sdktrace.ReadOnlySpan) map[string]attribute.Value {
	out := make(map[string]attribute.Value, len(span.Attributes()))
	for _, attr := range span.Attributes() {
		out[string(attr.Key)] = attr.Value
	}
	return out
}
