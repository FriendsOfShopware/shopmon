package mail

import (
	"context"
	"errors"
	"fmt"
	"testing"

	gomailer "github.com/shyim/go-mailer"
	"github.com/shyim/go-mailer/middleware"
	"github.com/shyim/go-mailer/middleware/otelmw"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestIsRetryableSMTPCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code int
		want bool
	}{
		// Transient soft failures (SES 451 is the production case).
		{421, true},
		{450, true},
		{451, true},
		{452, true},
		{499, true},

		// Permanent hard failures.
		{500, false},
		{550, false},
		{554, false},
		{599, false},

		// Non-SMTP / success / unknown — not classified as retryable.
		{0, false},
		{250, false},
		{354, false},
		{399, false},
		{600, false},
		{-1, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("code_%d", tt.code), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, IsRetryableSMTPCode(tt.code))
		})
	}
}

func TestSMTPFailureAttrs(t *testing.T) {
	t.Parallel()

	t.Run("nil error", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, smtpFailureAttrs(nil))
	})

	t.Run("non-transport error", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, smtpFailureAttrs(errors.New("boom")))
	})

	t.Run("transport error without code", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, smtpFailureAttrs(gomailer.NewTransportError("dial failed")))
	})

	t.Run("wrapped 451 soft failure", func(t *testing.T) {
		t.Parallel()
		te := gomailer.NewTransportError(`Expected response code "250" but got code "451"`)
		te.Code = 451
		attrs := smtpFailureAttrs(fmt.Errorf("send mail: %w", te))
		require.Len(t, attrs, 2)
		assert.Equal(t, attribute.Int(attrSMTPCode, 451), attrs[0])
		assert.Equal(t, attribute.Bool(attrRetryable, true), attrs[1])
	})

	t.Run("550 hard failure", func(t *testing.T) {
		t.Parallel()
		te := gomailer.NewTransportError("mailbox unavailable")
		te.Code = 550
		attrs := smtpFailureAttrs(te)
		require.Len(t, attrs, 2)
		assert.Equal(t, attribute.Int(attrSMTPCode, 550), attrs[0])
		assert.Equal(t, attribute.Bool(attrRetryable, false), attrs[1])
	})
}

// failingTransport returns a fixed TransportError from Send.
type failingTransport struct {
	err error
}

func (f *failingTransport) String() string { return "smtp://test" }

func (f *failingTransport) Send(context.Context, gomailer.RawMessage, *gomailer.Envelope) (*gomailer.SentMessage, error) {
	return nil, f.err
}

func TestSMTPSpanAttrsOnGomailerSend(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))

	te := gomailer.NewTransportError(`Expected response code "250" but got code "451"`)
	te.Code = 451
	leaf := &failingTransport{err: te}

	// Same order as NewService: otelmw outer, smtp attrs inner.
	tr := middleware.Wrap(leaf, otelmw.New(tp, nil), smtpSpanAttrs())

	from := gomailer.MustAddress("from@example.com", "")
	to := gomailer.MustAddress("to@example.com", "")
	env, err := gomailer.NewEnvelope(from, []gomailer.Address{to})
	require.NoError(t, err)

	_, sendErr := tr.Send(context.Background(), nil, env)
	require.Error(t, sendErr)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	span := spans[0]
	assert.Equal(t, "gomailer.send", span.Name())

	attrs := span.Attributes()
	code, ok := findAttr(attrs, attrSMTPCode)
	require.True(t, ok, "missing %s", attrSMTPCode)
	assert.Equal(t, int64(451), code.AsInt64())

	retryable, ok := findAttr(attrs, attrRetryable)
	require.True(t, ok, "missing %s", attrRetryable)
	assert.True(t, retryable.AsBool())

	// Existing otelmw outcome attribute is preserved (we extend, not replace).
	outcome, ok := findAttr(attrs, "messaging.gomailer.outcome")
	require.True(t, ok)
	assert.Equal(t, "error", outcome.AsString())

	// No high-cardinality recipient/subject/body attributes.
	for _, kv := range attrs {
		key := string(kv.Key)
		assert.NotEqual(t, "mail.recipient", key)
		assert.NotEqual(t, "mail.to", key)
		assert.NotEqual(t, "mail.subject", key)
		assert.NotEqual(t, "mail.body", key)
		assert.NotContains(t, key, "email.address")
	}
}

func findAttr(attrs []attribute.KeyValue, key string) (attribute.Value, bool) {
	for _, kv := range attrs {
		if string(kv.Key) == key {
			return kv.Value, true
		}
	}
	return attribute.Value{}, false
}
