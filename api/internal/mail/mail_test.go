package mail

import (
	"errors"
	"testing"

	"github.com/shyim/go-mailer/mailertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	svc, err := NewService(Config{
		DSN:  "smtp://localhost:587",
		From: "noreply@example.com",
	})
	require.NoError(t, err)
	assert.NotNil(t, svc)
	assert.Equal(t, "noreply@example.com", svc.from.Email())
}

func TestNewServiceInvalidDSN(t *testing.T) {
	_, err := NewService(Config{DSN: "://nope", From: "noreply@example.com"})
	assert.Error(t, err)
}

func TestNewServiceInvalidFrom(t *testing.T) {
	_, err := NewService(Config{DSN: "smtp://localhost:587", From: "not an address"})
	assert.Error(t, err)
}

func TestSend(t *testing.T) {
	rec := mailertest.NewRecordingTransport("")
	svc, err := NewServiceWithTransport(rec, "sender@example.com", "reply@example.com", "")
	require.NoError(t, err)

	err = svc.Send(t.Context(), "recipient@example.com",
		Email{Subject: "Test Subject", HTML: "<p>Hello</p>"})
	require.NoError(t, err)

	mailertest.AssertEmailCount(t, rec, 1)
	sent, ok := rec.Last()
	require.True(t, ok)
	msg := string(sent.Bytes())
	assert.Contains(t, msg, "From: sender@example.com")
	assert.Contains(t, msg, "To: recipient@example.com")
	assert.Contains(t, msg, "Subject: Test Subject")
	assert.Contains(t, msg, "Reply-To: reply@example.com")
	assert.Contains(t, msg, "MIME-Version: 1.0")
	assert.Contains(t, msg, "text/html")
	assert.Contains(t, msg, "<p>Hello</p>")
}

func TestSendNoReplyTo(t *testing.T) {
	rec := mailertest.NewRecordingTransport("")
	svc, err := NewServiceWithTransport(rec, "sender@example.com", "", "")
	require.NoError(t, err)

	err = svc.Send(t.Context(), "recipient@example.com",
		Email{Subject: "Test Subject", HTML: "<p>Hello</p>"})
	require.NoError(t, err)

	sent, ok := rec.Last()
	require.True(t, ok)
	assert.NotContains(t, string(sent.Bytes()), "Reply-To:")
}

func TestSendTextAndHTML(t *testing.T) {
	rec := mailertest.NewRecordingTransport("")
	svc, err := NewServiceWithTransport(rec, "sender@example.com", "", "")
	require.NoError(t, err)

	err = svc.Send(t.Context(), "recipient@example.com",
		Email{Subject: "Subject", Text: "plain body", HTML: "<p>rich body</p>"})
	require.NoError(t, err)

	sent, ok := rec.Last()
	require.True(t, ok)
	msg := string(sent.Bytes())
	assert.Contains(t, msg, "text/plain")
	assert.Contains(t, msg, "plain body")
	assert.Contains(t, msg, "text/html")
	assert.Contains(t, msg, "<p>rich body</p>")
}

func TestSendInvalidRecipient(t *testing.T) {
	rec := mailertest.NewRecordingTransport("")
	svc, err := NewServiceWithTransport(rec, "sender@example.com", "", "")
	require.NoError(t, err)

	err = svc.Send(t.Context(), "not an address", Email{Subject: "Subject", Text: "body"})
	assert.Error(t, err)
	mailertest.AssertNotSent(t, rec)
}

func TestSendTransportError(t *testing.T) {
	rec := mailertest.NewRecordingTransport("")
	svc, err := NewServiceWithTransport(rec, "sender@example.com", "", "")
	require.NoError(t, err)

	rec.FailNext(errors.New("relay unavailable"))
	err = svc.Send(t.Context(), "recipient@example.com", Email{Subject: "Subject", Text: "body"})
	assert.Error(t, err)
	mailertest.AssertNotSent(t, rec)
}

func TestBuildEmailUsesFrontendURL(t *testing.T) {
	rec := mailertest.NewRecordingTransport("")
	svc, err := NewServiceWithTransport(rec, "sender@example.com", "", "https://app.example.test")
	require.NoError(t, err)

	email := svc.BuildConfirmationEmail("Alice", "https://app.example.test/account/confirm/tok")
	assert.Equal(t, "Confirm your email address", email.Subject)
	assert.Contains(t, email.HTML, "https://app.example.test")
}

func TestBuildEmailDefaultProductLink(t *testing.T) {
	rec := mailertest.NewRecordingTransport("")
	svc, err := NewServiceWithTransport(rec, "sender@example.com", "", "")
	require.NoError(t, err)

	email := svc.BuildPasswordResetEmail("Bob", "https://app.example.test/reset/tok")
	assert.Contains(t, email.HTML, defaultProductLink)
}
