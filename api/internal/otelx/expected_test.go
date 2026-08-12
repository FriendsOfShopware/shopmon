package otelx

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestHTTPClientStatusExpected(t *testing.T) {
	expected := []int{401, 403, 408, 425, 429, 502, 503, 504}
	for _, code := range expected {
		assert.True(t, HTTPClientStatusExpected(code), "status %d", code)
	}

	hard := []int{0, 200, 301, 400, 404, 418, 500, 501, 505}
	for _, code := range hard {
		assert.False(t, HTTPClientStatusExpected(code), "status %d", code)
	}
}

type stubAPIError struct {
	code int
}

func (e *stubAPIError) Error() string      { return fmt.Sprintf("status %d", e.code) }
func (e *stubAPIError) HTTPStatusCode() int { return e.code }

func TestIsExpectedDependencyError(t *testing.T) {
	assert.False(t, IsExpectedDependencyError(nil))
	assert.True(t, IsExpectedDependencyError(&stubAPIError{code: 429}))
	assert.True(t, IsExpectedDependencyError(&stubAPIError{code: 401}))
	assert.False(t, IsExpectedDependencyError(&stubAPIError{code: 500}))
	assert.False(t, IsExpectedDependencyError(&stubAPIError{code: 0}))

	assert.True(t, IsExpectedDependencyError(fmt.Errorf("wrap: %w", &stubAPIError{code: 503})))
	assert.True(t, IsExpectedDependencyError(&net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: syscall.ECONNREFUSED,
	}))
	assert.False(t, IsExpectedDependencyError(errors.New("boom")))
}

func TestRecordExpectedSetsOkAndAttribute(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	ctx, span := tp.Tracer("test").Start(context.Background(), "op")

	err := &stubAPIError{code: 429}
	RecordExpected(span, err)
	span.End()
	require.NoError(t, tp.Shutdown(ctx))

	require.Len(t, sr.Ended(), 1)
	got := sr.Ended()[0]
	assert.Equal(t, codes.Ok, got.Status().Code)
	assert.True(t, hasBoolAttr(got.Attributes(), AttrErrorExpected, true))
	require.NotEmpty(t, got.Events(), "RecordError should add an event")
}

func TestRecordHardSetsErrorWithoutExpected(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	ctx, span := tp.Tracer("test").Start(context.Background(), "op")

	RecordHard(span, errors.New("bug"))
	span.End()
	require.NoError(t, tp.Shutdown(ctx))

	require.Len(t, sr.Ended(), 1)
	got := sr.Ended()[0]
	assert.Equal(t, codes.Error, got.Status().Code)
	assert.False(t, hasBoolAttr(got.Attributes(), AttrErrorExpected, true))
}

func TestRecordDependencyClassifies(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	tr := tp.Tracer("test")

	_, span := tr.Start(context.Background(), "expected")
	RecordDependency(span, &stubAPIError{code: http.StatusTooManyRequests})
	span.End()

	_, span = tr.Start(context.Background(), "hard")
	RecordDependency(span, errors.New("unexpected"))
	span.End()

	require.NoError(t, tp.Shutdown(context.Background()))
	require.Len(t, sr.Ended(), 2)
	assert.Equal(t, codes.Ok, sr.Ended()[0].Status().Code)
	assert.Equal(t, codes.Error, sr.Ended()[1].Status().Code)
}

func hasBoolAttr(attrs []attribute.KeyValue, key string, want bool) bool {
	for _, a := range attrs {
		if string(a.Key) == key && a.Value.Type() == attribute.BOOL && a.Value.AsBool() == want {
			return true
		}
	}
	return false
}
