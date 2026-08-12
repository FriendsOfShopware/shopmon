// Package otelx classifies expected/transient dependency failures for OpenTelemetry
// spans so Datadog APM error rate reflects shopmon bugs rather than tenant-side
// or retryable upstream noise.
//
// Taxonomy (keep cardinality low — only these stable attributes):
//
//   - error.expected=true  — this failure is an expected dependency degradation
//     (rate limit, tenant auth, retryable upstream 5xx, etc.). Prefer leaving
//     span status Ok for these so they do not count toward APM error rate.
//     Alerts for real errors:
//
//     status:error -@error.expected:true
//
//   - Hard failures (panics, unexpected 5xx from our API, exhausted non-retryable
//     job errors) keep status=Error and MUST NOT set error.expected.
//
// Outcome counters from package metrics (shopmon.*.outcome) remain the cheap
// signal for rate_limited / auth_error volumes; this package only adjusts spans.
//
// SMTP soft-failure classification lives in package mail (keeps go-mailer out of
// the shared HTTP client dependency graph).
package otelx

import (
	"errors"
	"net"
	"net/http"
	"os"
	"syscall"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// AttrErrorExpected is the stable boolean span attribute used to mark
// expected/degraded dependency failures. Datadog query to find real errors:
//
//	status:error -@error.expected:true
const AttrErrorExpected = "error.expected"

// ErrorExpectedAttr is the attribute.KeyValue form of AttrErrorExpected=true.
var ErrorExpectedAttr = attribute.Bool(AttrErrorExpected, true)

// HTTPClientStatusExpected reports whether an outbound HTTP status code is an
// expected dependency degradation rather than a shopmon bug.
//
// Included:
//   - 401/403 — tenant shop credentials / ACL (not our bug)
//   - 408/425/429 — timeouts / rate limits (retried or aborted for later retry)
//   - 502/503/504 — retryable upstream gateway failures (e.g. Sitespeed)
//
// Excluded (still Error): 400/404/500 and other unexpected statuses.
func HTTPClientStatusExpected(code int) bool {
	switch code {
	case http.StatusUnauthorized, // 401
		http.StatusForbidden,          // 403
		http.StatusRequestTimeout,     // 408
		http.StatusTooEarly,           // 425
		http.StatusTooManyRequests,    // 429
		http.StatusBadGateway,         // 502
		http.StatusServiceUnavailable, // 503
		http.StatusGatewayTimeout:     // 504
		return true
	default:
		return false
	}
}

// IsRetryableNetError reports connection-refused / reset / timeout style errors
// that background jobs will retry (Sitespeed down, brief relay blip, etc.).
func IsRetryableNetError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ETIMEDOUT) {
		return true
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	var syscallErr *os.SyscallError
	if errors.As(err, &syscallErr) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return false
}

// httpStatusCoder is implemented by dependency API errors (shopware.ApiError,
// shopwareaccount.APIError) so call sites can classify without switching on
// concrete types here.
type httpStatusCoder interface {
	HTTPStatusCode() int
}

// DependencyHTTPStatus extracts an HTTP status from err when it implements
// HTTPStatusCode() with a positive code.
func DependencyHTTPStatus(err error) (int, bool) {
	var h httpStatusCoder
	if errors.As(err, &h) {
		if code := h.HTTPStatusCode(); code > 0 {
			return code, true
		}
	}
	return 0, false
}

// IsExpectedDependencyError reports whether err is a classified expected
// dependency degradation (HTTP status or retryable network error).
// Soft SMTP failures are classified in package mail via IsExpectedSMTPError.
func IsExpectedDependencyError(err error) bool {
	if err == nil {
		return false
	}
	if code, ok := DependencyHTTPStatus(err); ok {
		return HTTPClientStatusExpected(code)
	}
	return IsRetryableNetError(err)
}

// RecordExpected marks a span as an expected dependency degradation: records the
// error as an event, sets error.expected=true, and forces status Ok so Datadog
// error rate is not inflated. Use for intermediate retries and tenant-side
// failures.
//
// codes.Ok is required (not Unset): the Go OTel SDK will not downgrade Error→Unset.
func RecordExpected(span trace.Span, err error) {
	if !span.IsRecording() {
		return
	}
	span.SetAttributes(ErrorExpectedAttr)
	if err != nil {
		span.RecordError(err)
	}
	span.SetStatus(codes.Ok, "")
}

// RecordExpectedHTTP is like RecordExpected when the HTTP status is known.
func RecordExpectedHTTP(span trace.Span, statusCode int, err error) {
	if !span.IsRecording() {
		return
	}
	span.SetAttributes(
		ErrorExpectedAttr,
		attribute.Int("http.response.status_code", statusCode),
	)
	if err != nil {
		span.RecordError(err)
	}
	span.SetStatus(codes.Ok, "")
}

// RecordHard marks a span as a real failure: RecordError + status Error, without
// error.expected. Use for panics, unexpected bugs, and final hard failures.
func RecordHard(span trace.Span, err error) {
	if !span.IsRecording() || err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// RecordDependency records err on span using expected-vs-hard classification.
func RecordDependency(span trace.Span, err error) {
	if err == nil {
		return
	}
	if IsExpectedDependencyError(err) {
		RecordExpected(span, err)
		return
	}
	RecordHard(span, err)
}
