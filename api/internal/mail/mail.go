// Package mail provides SMTP email sending and HTML email templates.
package mail

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"
)

// bodyPartSeparator delimits the plain-text and HTML parts inside the body
// string handed to Send. It is intentionally unlikely to occur in rendered
// email content.
const bodyPartSeparator = "\x00--shopmon-body-part--\x00"

// smtpTimeout bounds the entire SMTP exchange so a stalled server cannot hang
// the caller indefinitely.
const smtpTimeout = 30 * time.Second

// Sender sends an HTML email to a recipient.
type Sender interface {
	Send(ctx context.Context, to, subject, body string) error
}

// SMTPConfig holds SMTP connection settings.
type SMTPConfig struct {
	Host      string
	Port      string
	Secure    bool
	User      string
	Pass      string
	From      string
	ReplyTo   string
	TLSConfig *tls.Config
}

// Service sends emails via SMTP.
type Service struct {
	cfg SMTPConfig
}

// NewService creates a new mail service.
func NewService(cfg SMTPConfig) *Service {
	return &Service{cfg: cfg}
}

// NoopSender silently discards emails. Useful in tests.
type NoopSender struct{}

// Send implements Sender.
func (NoopSender) Send(_ context.Context, to, subject, body string) error {
	slog.Debug("noop mail send", "to", to, "subject", subject)
	return nil
}

// Send sends an email with both a text/plain and a text/html part. The body may
// carry the two parts packed around bodyPartSeparator (as produced by the
// template builders); otherwise it is treated as HTML.
// tlsClientConfig returns the TLS configuration for SMTP connections, defaulting
// ServerName to the configured host when not otherwise pinned.
func (s *Service) tlsClientConfig() *tls.Config {
	var tlsConfig *tls.Config
	if s.cfg.TLSConfig != nil {
		tlsConfig = s.cfg.TLSConfig.Clone()
	} else {
		tlsConfig = &tls.Config{}
	}
	if tlsConfig.ServerName == "" && !tlsConfig.InsecureSkipVerify {
		tlsConfig.ServerName = s.cfg.Host
	}
	return tlsConfig
}

func (s *Service) Send(ctx context.Context, to, subject, body string) error {
	text, html := splitBody(body)

	msg, err := s.buildMessage(to, subject, text, html)
	if err != nil {
		return fmt.Errorf("build message: %w", err)
	}

	var auth smtp.Auth
	if s.cfg.User != "" {
		auth = smtp.PlainAuth("", s.cfg.User, s.cfg.Pass, s.cfg.Host)
	}

	addr := net.JoinHostPort(s.cfg.Host, s.cfg.Port)
	dialer := net.Dialer{Timeout: smtpTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if err := conn.SetDeadline(time.Now().Add(smtpTimeout)); err != nil {
		return fmt.Errorf("set deadline: %w", err)
	}

	if s.cfg.Secure {
		tlsConn := tls.Client(conn, s.tlsClientConfig())
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			return fmt.Errorf("tls handshake: %w", err)
		}
		conn = tlsConn
	}

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer func() { _ = client.Close() }()

	// On a plain connection, upgrade to TLS via STARTTLS when the server
	// advertises it. Submission relays (port 587) require this before AUTH,
	// and smtp.PlainAuth refuses to send credentials over an unencrypted
	// connection to a non-localhost host.
	if !s.cfg.Secure {
		if ok, _ := client.Extension("STARTTLS"); ok {
			tlsConfig := s.tlsClientConfig()
			if err := client.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("smtp starttls: %w", err)
			}
		}
	}

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}

	return client.Quit()
}

// buildMessage assembles an RFC 5322 multipart/alternative message with a
// text/plain and a text/html part.
func (s *Service) buildMessage(to, subject, text, html string) ([]byte, error) {
	var msg strings.Builder
	body := &strings.Builder{}
	mw := multipart.NewWriter(body)

	if text != "" {
		part, err := mw.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {"text/plain; charset=UTF-8"},
			"Content-Transfer-Encoding": {"8bit"},
		})
		if err != nil {
			return nil, err
		}
		if _, err := part.Write([]byte(text)); err != nil {
			return nil, err
		}
	}

	part, err := mw.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"text/html; charset=UTF-8"},
		"Content-Transfer-Encoding": {"8bit"},
	})
	if err != nil {
		return nil, err
	}
	if _, err := part.Write([]byte(html)); err != nil {
		return nil, err
	}

	if err := mw.Close(); err != nil {
		return nil, err
	}

	headers := []struct{ k, v string }{
		{"From", s.cfg.From},
		{"To", to},
		{"Subject", subject},
		{"Date", time.Now().Format(time.RFC1123Z)},
		{"Message-ID", s.messageID()},
		{"MIME-Version", "1.0"},
		{"Content-Type", fmt.Sprintf("multipart/alternative; boundary=%q", mw.Boundary())},
	}
	for _, h := range headers {
		fmt.Fprintf(&msg, "%s: %s\r\n", h.k, h.v)
	}
	if s.cfg.ReplyTo != "" {
		fmt.Fprintf(&msg, "Reply-To: %s\r\n", s.cfg.ReplyTo)
	}
	msg.WriteString("\r\n")
	msg.WriteString(body.String())

	return []byte(msg.String()), nil
}

// messageID returns a unique RFC 5322 Message-ID using the From domain.
func (s *Service) messageID() string {
	var buf [16]byte
	_, _ = rand.Read(buf[:])

	domain := s.cfg.Host
	if at := strings.LastIndex(s.cfg.From, "@"); at != -1 {
		domain = strings.Trim(s.cfg.From[at+1:], "<> ")
	}
	if domain == "" {
		domain = "localhost"
	}

	return fmt.Sprintf("<%s.%d@%s>", hex.EncodeToString(buf[:]), time.Now().UnixNano(), domain)
}

// splitBody separates a packed body into its plain-text and HTML parts. A body
// without the separator is treated as HTML only.
func splitBody(body string) (text, html string) {
	if idx := strings.Index(body, bodyPartSeparator); idx != -1 {
		return body[:idx], body[idx+len(bodyPartSeparator):]
	}
	return "", body
}
