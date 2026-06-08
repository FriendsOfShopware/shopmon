package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

// TraceIDHeader writes the current OpenTelemetry trace ID to the response.
func TraceIDHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spanCtx := trace.SpanContextFromContext(r.Context())
		if spanCtx.HasTraceID() {
			w.Header().Set("X-Trace-Id", spanCtx.TraceID().String())
		}

		next.ServeHTTP(w, r)
	})
}
