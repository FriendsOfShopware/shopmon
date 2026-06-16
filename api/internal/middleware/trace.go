package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
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

// RouteTag sets the http.route attribute on the active span once chi has
// matched the route. otelhttp creates the span before routing, so the route
// pattern is only known after the request has been served. Datadog derives the
// span resource name from http.method + http.route, so without this every
// request collapses to a bare "GET" resource.
func RouteTag(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		rctx := chi.RouteContext(r.Context())
		if rctx == nil || rctx.RoutePattern() == "" {
			return
		}
		trace.SpanFromContext(r.Context()).SetAttributes(
			semconv.HTTPRoute(rctx.RoutePattern()),
		)
	})
}
