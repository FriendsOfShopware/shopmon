package otelx

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestWrapClientTransportDowngradesExpectedStatus(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = io.WriteString(w, "slow down")
	}))
	t.Cleanup(upstream.Close)

	client := &http.Client{
		Transport: WrapClientTransport(
			otelhttp.NewTransport(
				CaptureClientSpan(http.DefaultTransport),
				otelhttp.WithTracerProvider(tp),
			),
		),
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, upstream.URL, nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	_, _ = io.Copy(io.Discard, resp.Body)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

	require.NoError(t, tp.Shutdown(context.Background()))
	require.NotEmpty(t, sr.Ended())
	span := sr.Ended()[0]
	assert.Equal(t, codes.Ok, span.Status().Code, "429 must not remain status=Error")
	assert.True(t, hasBoolAttr(span.Attributes(), AttrErrorExpected, true))
}

func TestWrapClientTransportKeepsHardStatus(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(upstream.Close)

	client := &http.Client{
		Transport: WrapClientTransport(
			otelhttp.NewTransport(
				CaptureClientSpan(http.DefaultTransport),
				otelhttp.WithTracerProvider(tp),
			),
		),
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, upstream.URL, nil)
	require.NoError(t, err)
	resp, err := client.Do(req)
	require.NoError(t, err)
	_, _ = io.Copy(io.Discard, resp.Body)
	require.NoError(t, resp.Body.Close())

	require.NoError(t, tp.Shutdown(context.Background()))
	require.NotEmpty(t, sr.Ended())
	span := sr.Ended()[0]
	assert.Equal(t, codes.Error, span.Status().Code)
	assert.False(t, hasBoolAttr(span.Attributes(), AttrErrorExpected, true))
}
