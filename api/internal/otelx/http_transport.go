package otelx

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

// spanBoxKey is an unexported context key. expectedStatusTransport stashes an
// empty box in the request context; spanCaptureTransport (inside otelhttp)
// fills it with the client span so expectedStatusTransport can downgrade
// expected 4xx/5xx from Error→Ok after otelhttp has set status.
type spanBoxKey struct{}

type spanBox struct {
	span trace.Span
}

// WrapClientTransport wraps base (typically an otelhttp.Transport) so expected
// dependency HTTP statuses do not count as APM errors.
//
// Call as:
//
//	otelx.WrapClientTransport(otelhttp.NewTransport(base, ...))
//
// Order matters: this wrapper must sit *outside* otelhttp so it runs after
// otelhttp sets span status from the response code, and *before* the body is
// closed (which ends the span).
func WrapClientTransport(otelTransport http.RoundTripper) http.RoundTripper {
	if otelTransport == nil {
		otelTransport = http.DefaultTransport
	}
	return &expectedStatusTransport{base: otelTransport}
}

// CaptureClientSpan returns a RoundTripper that records the active client span
// from the request context into a box placed by WrapClientTransport. Use it as
// the *base* of otelhttp.NewTransport:
//
//	otelhttp.NewTransport(otelx.CaptureClientSpan(base), ...)
func CaptureClientSpan(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &spanCaptureTransport{base: base}
}

type spanCaptureTransport struct {
	base http.RoundTripper
}

func (t *spanCaptureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if box, ok := req.Context().Value(spanBoxKey{}).(*spanBox); ok && box != nil {
		box.span = trace.SpanFromContext(req.Context())
	}
	return t.base.RoundTrip(req)
}

type expectedStatusTransport struct {
	base http.RoundTripper
}

func (t *expectedStatusTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	box := &spanBox{}
	ctx := context.WithValue(req.Context(), spanBoxKey{}, box)
	req = req.WithContext(ctx)

	resp, err := t.base.RoundTrip(req)
	if err != nil || resp == nil {
		return resp, err
	}
	if !HTTPClientStatusExpected(resp.StatusCode) {
		return resp, nil
	}
	if box.span != nil && box.span.IsRecording() {
		// otelhttp already set status=Error and error.type=<code>. Downgrade
		// status so APM error rate stays clean; keep attributes for forensics.
		RecordExpectedHTTP(box.span, resp.StatusCode, nil)
	}
	return resp, nil
}
