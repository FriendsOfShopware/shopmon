package mail

import (
	"context"
	"errors"

	"github.com/friendsofshopware/shopmon/api/internal/otelx"
	gomailer "github.com/shyim/go-mailer"
	"github.com/shyim/go-mailer/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Span name matches go-mailer's otelmw default so existing Datadog dashboards
// keep working after we swap in a classifying tracer.
const mailSpanName = "gomailer.send"

// instrumentTransport wraps leaf with an OpenTelemetry client span per Send
// attempt. Soft SMTP failures (421/450/451/452) and retryable network blips
// set error.expected=true and status Ok; other failures keep status Error.
//
// Outcome volume for alerts stays on shopmon.mail.send (package metrics). We
// intentionally do not use otelmw.New here: its Span.SetError path cannot
// classify expected degradations.
func instrumentTransport(leaf gomailer.Transport) gomailer.Transport {
	return middleware.Wrap(leaf, middleware.Observability(
		middleware.WithTracer(&classifyingMailTracer{
			tracer: otel.Tracer("shopmon/mail"),
		}),
		middleware.WithSpanName(mailSpanName),
	))
}

type classifyingMailTracer struct {
	tracer trace.Tracer
}

func (t *classifyingMailTracer) Start(ctx context.Context, name string) (context.Context, middleware.Span) {
	ctx, span := t.tracer.Start(ctx, name, trace.WithSpanKind(trace.SpanKindClient))
	return ctx, &classifyingMailSpan{span: span}
}

type classifyingMailSpan struct {
	span    trace.Span
	lastErr error
}

func (s *classifyingMailSpan) SetAttributes(attrs ...middleware.Attr) {
	if len(attrs) == 0 {
		return
	}
	kvs := make([]attribute.KeyValue, 0, len(attrs))
	for _, a := range attrs {
		switch a.Kind {
		case middleware.KindInt:
			kvs = append(kvs, attribute.Int64(a.Key, a.Int))
		case middleware.KindBool:
			kvs = append(kvs, attribute.Bool(a.Key, a.Bool))
		default:
			kvs = append(kvs, attribute.String(a.Key, a.Str))
		}
	}
	s.span.SetAttributes(kvs...)
}

func (s *classifyingMailSpan) RecordError(err error) {
	// Stash until SetError so we can choose expected vs hard. The observability
	// middleware always calls RecordError then SetError on failure.
	s.lastErr = err
}

func (s *classifyingMailSpan) SetError(description string) {
	err := s.lastErr
	if IsExpectedSMTPError(err) {
		otelx.RecordExpected(s.span, err)
		var te *gomailer.TransportError
		if errors.As(err, &te) && te.Code != 0 {
			s.span.SetAttributes(attribute.Int("smtp.response.code", te.Code))
		}
		return
	}
	if err != nil {
		s.span.RecordError(err)
	}
	s.span.SetStatus(codes.Error, description)
}

func (s *classifyingMailSpan) End() { s.span.End() }
