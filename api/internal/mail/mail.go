// Package mail provides email sending (via the go-mailer SMTP transport) and
// HTML email templates.
package mail

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"

	gomailer "github.com/shyim/go-mailer"
	"github.com/shyim/go-mailer/middleware"
	"github.com/shyim/go-mailer/middleware/otelmw"
	"github.com/shyim/go-mailer/transport"
	smtptransport "github.com/shyim/go-mailer/transport/smtp"
)

// Email is a fully composed message: its subject line and rendered bodies. Text
// is the plain-text alternative; HTML is the rich representation. The template
// builders populate all three.
type Email struct {
	Subject string
	Text    string
	HTML    string
}

// Sender composes and sends emails. The Build* methods render the message
// content (their product link points at the configured frontend); Send delivers
// a composed Email to a recipient.
type Sender interface {
	Send(ctx context.Context, to string, email Email) error

	BuildConfirmationEmail(name, verifyURL string) Email
	BuildPasswordResetEmail(name, resetURL string) Email
	BuildOrgInviteEmail(inviterName, orgName, acceptURL, rejectURL string) Email
	BuildStatusChangeEmail(userName, envName, newStatus, alertMessage string) Email
	BuildConnectionFailedEmail(userName, envName, alertMessage string) Email
}

// Config holds the settings for the SMTP mailer.
type Config struct {
	// DSN is a go-mailer SMTP connection string, e.g.
	// "smtp://user:pass@host:587" or "smtps://user:pass@host:465". Supported
	// options include verify_peer, require_tls and auto_tls.
	DSN string
	// From is the envelope/header From address. It may be a bare address
	// ("noreply@example.com") or include a display name
	// ("Shopmon <noreply@example.com>").
	From string
	// ReplyTo is an optional Reply-To address (bare or with display name).
	ReplyTo string
	// FrontendURL is the base URL of the web app, used as the product link in
	// rendered emails (e.g. the logo/header link).
	FrontendURL string
}

// Service sends emails via the go-mailer SMTP transport.
type Service struct {
	mailer  *gomailer.Mailer
	from    gomailer.Address
	replyTo *gomailer.Address
	// frontendURL is the product link embedded in rendered emails.
	frontendURL string
	// closer releases the leaf transport. It is captured before any middleware
	// wrapping, because the observability middleware does not forward Close, so
	// Mailer.Close would otherwise silently skip the pooled SMTP connection.
	closer io.Closer
}

// NewService creates a mail service from cfg. It returns an error when the DSN
// or the From/ReplyTo addresses cannot be parsed.
func NewService(cfg Config) (*Service, error) {
	leaf, err := transport.FromDSN(cfg.DSN, transport.Deps{})
	if err != nil {
		return nil, fmt.Errorf("mail transport from dsn: %w", err)
	}

	// The DSN factory refuses to send AUTH credentials over an unencrypted
	// connection. Local relays (e.g. the dev MailHog/Mailpit on localhost)
	// commonly advertise AUTH without STARTTLS; allow plaintext auth there so
	// development and tests keep working, while remote relays stay protected.
	if st, ok := leaf.(*smtptransport.Transport); ok && isLocalDSNHost(cfg.DSN) {
		st.SetAllowPlaintextAuth(true)
	}

	// Instrument each delivery attempt with an OpenTelemetry span and metrics.
	// Passing nil providers makes otelmw fall back to the globals configured by
	// the telemetry package; when telemetry is disabled those are no-ops, so the
	// middleware degrades to a cheap pass-through. We wrap the leaf transport so
	// every delivery attempt (including retries) gets its own span, and keep the
	// leaf's closer so shutdown still QUITs the pooled connection.
	tr := middleware.Wrap(leaf, otelmw.New(nil, nil))

	return newService(tr, leaf, cfg.From, cfg.ReplyTo, cfg.FrontendURL)
}

// NewServiceWithTransport creates a mail service over an arbitrary go-mailer
// transport. Production uses NewService (which builds an instrumented SMTP
// transport from a DSN); tests inject a mailertest.RecordingTransport to
// capture messages. It returns an error when the from/replyTo addresses cannot
// be parsed.
func NewServiceWithTransport(tr gomailer.Transport, from, replyTo, frontendURL string) (*Service, error) {
	return newService(tr, tr, from, replyTo, frontendURL)
}

// newService builds a Service from the (possibly wrapped) transport used for
// sending and the underlying transport used for Close.
func newService(tr, closeTarget gomailer.Transport, from, replyTo, frontendURL string) (*Service, error) {
	fromAddr, err := gomailer.ParseAddress(from)
	if err != nil {
		return nil, fmt.Errorf("mail from address: %w", err)
	}

	svc := &Service{
		mailer:      gomailer.NewMailer(tr),
		from:        fromAddr,
		frontendURL: frontendURL,
	}
	if c, ok := closeTarget.(io.Closer); ok {
		svc.closer = c
	}

	if replyTo != "" {
		replyToAddr, err := gomailer.ParseAddress(replyTo)
		if err != nil {
			return nil, fmt.Errorf("mail reply-to address: %w", err)
		}
		svc.replyTo = &replyToAddr
	}

	return svc, nil
}

// Close releases the underlying SMTP transport, sending QUIT and closing any
// pooled connection. It closes the leaf transport directly (rather than via the
// Mailer) so the observability wrapper, which does not forward Close, cannot
// suppress it. It is safe to call more than once.
func (s *Service) Close() error {
	if s.closer == nil {
		return nil
	}
	return s.closer.Close()
}

// Send sends an email with a text/plain and a text/html part to a single
// recipient.
func (s *Service) Send(ctx context.Context, to string, email Email) error {
	toAddr, err := gomailer.ParseAddress(to)
	if err != nil {
		return fmt.Errorf("recipient address: %w", err)
	}

	msg := gomailer.NewMessage().
		SetFrom(s.from).
		SetTo(toAddr).
		SetSubject(email.Subject)

	if email.Text != "" {
		msg.SetText([]byte(email.Text))
	}
	if email.HTML != "" {
		msg.SetHTML([]byte(email.HTML))
	}
	if s.replyTo != nil {
		msg.SetReplyTo(*s.replyTo)
	}

	// A nil envelope derives the sender and recipients from the message
	// headers set above.
	if err := s.mailer.Send(ctx, msg, nil); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	return nil
}

// isLocalDSNHost reports whether the SMTP DSN points at a loopback host, where
// plaintext AUTH over an unencrypted connection is acceptable.
func isLocalDSNHost(dsn string) bool {
	u, err := url.Parse(dsn)
	if err != nil {
		return false
	}
	host := u.Hostname()
	switch host {
	case "localhost", "127.0.0.1", "::1":
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}
