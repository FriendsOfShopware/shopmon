// Package mail provides SMTP email sending and HTML email templates.
package mail

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

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

// Send sends an HTML email.
func (s *Service) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port)

	headers := make(map[string]string)
	headers["From"] = s.cfg.From
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"
	if s.cfg.ReplyTo != "" {
		headers["Reply-To"] = s.cfg.ReplyTo
	}

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	var auth smtp.Auth
	if s.cfg.User != "" {
		auth = smtp.PlainAuth("", s.cfg.User, s.cfg.Pass, s.cfg.Host)
	}

	if s.cfg.Secure {
		tlsConfig := s.cfg.TLSConfig
		if tlsConfig == nil {
			tlsConfig = &tls.Config{ServerName: s.cfg.Host}
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("tls dial: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, s.cfg.Host)
		if err != nil {
			return fmt.Errorf("smtp client: %w", err)
		}
		defer client.Close()

		if auth != nil {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("smtp auth: %w", err)
			}
		}

		if err := client.Mail(s.cfg.From); err != nil {
			return err
		}
		if err := client.Rcpt(to); err != nil {
			return err
		}

		w, err := client.Data()
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte(msg.String())); err != nil {
			return err
		}
		return w.Close()
	}

	return smtp.SendMail(addr, auth, s.cfg.From, []string{to}, []byte(msg.String()))
}
