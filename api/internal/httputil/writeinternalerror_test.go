package httputil

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestWriteInternalErrorRecordsSpanError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/environments/1", nil)

	tp := sdktrace.NewTracerProvider()
	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(req.Context(), "test-span")
	req = req.WithContext(ctx)

	WriteInternalError(rec, req, errors.New("database exploded"), "failed to load environment", slog.String("operation", "get environment detail"))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var body ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "failed to load environment", body.Message)
	assert.NotContains(t, rec.Body.String(), "database exploded")

	ro := span.(sdktrace.ReadOnlySpan)
	errs := ro.Events()
	require.Len(t, errs, 1)
	assert.Equal(t, "exception", errs[0].Name)
	assert.Equal(t, "exception.type", string(errs[0].Attributes[0].Key))

	var hasStacktrace bool
	for _, attr := range errs[0].Attributes {
		if string(attr.Key) == "exception.stacktrace" && attr.Value.Type() != 0 {
			hasStacktrace = true
			break
		}
	}
	assert.True(t, hasStacktrace, "expected exception.stacktrace attribute to be set")

	assert.Contains(t, ro.Status().Description, "database exploded")
	assert.Equal(t, codes.Error, ro.Status().Code)
}
