package mail

import (
	"context"
	"errors"

	gomailer "github.com/shyim/go-mailer"
	"github.com/shyim/go-mailer/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Low-cardinality span attributes for outbound mail failures. These extend the
// attributes already set by go-mailer's otelmw (messaging.gomailer.outcome, …).
// Never attach recipient, subject, or body — those explode cardinality.
const (
	attrSMTPCode  = "messaging.gomailer.smtp_code"
	attrRetryable = "mail.retryable"
)

// IsRetryableSMTPCode reports whether an SMTP reply code is a soft/transient
// failure that is typically worth retrying.
//
// SMTP reply classes (RFC 5321):
//   - 4xx — transient negative (e.g. 421, 450, 451, 452) → retryable
//   - 5xx — permanent negative → not retryable
//
// Codes outside 400–599 (including 0 / unknown) are treated as not retryable
// so callers only flip mail.retryable=true for known soft failures.
func IsRetryableSMTPCode(code int) bool {
	return code >= 400 && code < 500
}

// smtpFailureAttrs extracts low-cardinality attributes from a send error when
// it wraps a *gomailer.TransportError with an SMTP response code. Returns nil
// when no SMTP code is available.
func smtpFailureAttrs(err error) []attribute.KeyValue {
	if err == nil {
		return nil
	}
	var te *gomailer.TransportError
	if !errors.As(err, &te) || te.Code == 0 {
		return nil
	}
	return []attribute.KeyValue{
		attribute.Int(attrSMTPCode, te.Code),
		attribute.Bool(attrRetryable, IsRetryableSMTPCode(te.Code)),
	}
}

// smtpSpanAttrs returns AfterSend middleware that attaches SMTP failure
// attributes to the active OpenTelemetry span. It must sit *inside* otelmw so
// the gomailer.send span is still open when the hook runs.
func smtpSpanAttrs() middleware.Middleware {
	return middleware.AfterSend(func(ctx context.Context, _ *gomailer.SentMessage, err error) {
		attrs := smtpFailureAttrs(err)
		if len(attrs) == 0 {
			return
		}
		span := trace.SpanFromContext(ctx)
		if !span.IsRecording() {
			return
		}
		span.SetAttributes(attrs...)
	})
}
