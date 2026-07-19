package mail

import (
	"fmt"
	"html"
	"log/slog"

	"github.com/matcornic/hermes"
	"github.com/russross/blackfriday/v2"
)

// defaultProductLink is used as the email product link when the service has no
// configured frontend URL (e.g. zero-value services in tests). It mirrors the
// FRONTEND_URL config default.
const defaultProductLink = "http://localhost:3000"

func (s *Service) newHermes() hermes.Hermes {
	link := s.frontendURL
	if link == "" {
		link = defaultProductLink
	}
	return hermes.Hermes{
		Theme: &shopmonTheme{},
		Product: hermes.Product{
			Name:      "Shopmon",
			Link:      link,
			Copyright: "Shopmon · FriendsOfShopware",
			Logo:      fmt.Sprintf("%s/shopmon-logo.png", link),
		},
	}
}

// BuildConfirmationEmail returns the email verification email.
func (s *Service) BuildConfirmationEmail(name, verifyURL string) Email {
	return s.generate("Confirm your email address", hermes.Email{
		Body: hermes.Body{
			Name:      name,
			Signature: "Best regards,",
			Intros: []string{
				"Thank you for registering with us. Please confirm your email address by clicking the button below.",
			},
			Actions: []hermes.Action{
				{
					Button: hermes.Button{
						Color: "#2563eb",
						Text:  "Confirm Email",
						Link:  verifyURL,
					},
				},
			},
			Outros: []string{
				"If you didn't create an account, you can safely ignore this email.",
			},
		},
	})
}

// BuildPasswordResetEmail returns the password reset email.
func (s *Service) BuildPasswordResetEmail(name, resetURL string) Email {
	return s.generate("Reset your password", hermes.Email{
		Body: hermes.Body{
			Name:      name,
			Signature: "Best regards,",
			Intros: []string{
				"We received a request to reset your password. Please click the button below to set a new password.",
			},
			Actions: []hermes.Action{
				{
					Button: hermes.Button{
						Color: "#2563eb",
						Text:  "Reset Password",
						Link:  resetURL,
					},
				},
			},
			Outros: []string{
				"This link will expire in 1 hour. If you didn't request a password reset, you can safely ignore this email.",
			},
		},
	})
}

// BuildOrgInviteEmail returns the organization invitation email.
func (s *Service) BuildOrgInviteEmail(inviterName, orgName, acceptURL, rejectURL string) Email {
	return s.generate("You have been invited to join "+orgName+" at Shopmon", hermes.Email{
		Body: hermes.Body{
			Signature: "Best regards,",
			Intros:    []string{inviterName + " has invited you to join **" + orgName + "** on Shopmon."},
			Actions: []hermes.Action{
				{
					Button: hermes.Button{
						Color: "#2563eb",
						Text:  "Accept Invitation",
						Link:  acceptURL,
					},
				},
				{
					Button: hermes.Button{
						Color: "#64748b",
						Text:  "Decline",
						Link:  rejectURL,
					},
				},
			},
		},
	})
}

// BuildStatusChangeEmail returns the alert email sent when an environment's
// status changes.
func (s *Service) BuildStatusChangeEmail(userName, envName, newStatus, alertMessage string) Email {
	return s.generate(
		"Environment "+envName+" status changed to "+newStatus,
		shopAlertBody(userName, envName, alertMessage),
	)
}

// BuildConnectionFailedEmail returns the alert email sent when an environment
// cannot be reached.
func (s *Service) BuildConnectionFailedEmail(userName, envName, alertMessage string) Email {
	return s.generate(
		"Environment "+envName+" connection failed",
		shopAlertBody(userName, envName, alertMessage),
	)
}

// shopAlertBody builds the shared body for environment alert emails.
func shopAlertBody(userName, shopName, alertMessage string) hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Name:      userName,
			Signature: "Best regards,",
			Intros: []string{
				"There is an alert for environment **" + shopName + "**:",
				alertMessage,
			},
		},
	}
}

// BuildAlertEmail renders a notification alert email. The subject, intro and
// message are already localized by the caller; this only wraps them in the
// branded template.
func (s *Service) BuildAlertEmail(userName, subject, intro, alertMessage string) Email {
	return s.generate(subject, hermes.Email{
		Body: hermes.Body{
			Name:      userName,
			Signature: "Best regards,",
			Intros: []string{
				intro,
				alertMessage,
			},
		},
	})
}

// generate renders an email with the given subject into its plain-text and HTML
// representations. On render errors it logs and falls back to whatever was
// produced.
func (s *Service) generate(subject string, email hermes.Email) Email {
	h := s.newHermes()

	// Make a shallow copy of email and deep copy Intros/Outros for HTML generation.
	// This ensures blackfriday HTML transformations do not contaminate the plain-text output.
	htmlEmail := email
	htmlEmail.Body.Intros = make([]string, len(email.Body.Intros))
	copy(htmlEmail.Body.Intros, email.Body.Intros)
	htmlEmail.Body.Outros = make([]string, len(email.Body.Outros))
	copy(htmlEmail.Body.Outros, email.Body.Outros)

	// Pre-process markdown formatting in Intros and Outros for rich HTML email rendering.
	// We escape HTML first to prevent raw HTML injection from user inputs (e.g. orgName, shopName).
	for i, intro := range htmlEmail.Body.Intros {
		escaped := html.EscapeString(intro)
		htmlEmail.Body.Intros[i] = string(blackfriday.Run([]byte(escaped)))
	}
	for i, outro := range htmlEmail.Body.Outros {
		escaped := html.EscapeString(outro)
		htmlEmail.Body.Outros[i] = string(blackfriday.Run([]byte(escaped)))
	}

	html, err := h.GenerateHTML(htmlEmail)
	if err != nil {
		slog.Error("generate html email", "error", err)
	}

	text, err := h.GeneratePlainText(email)
	if err != nil {
		slog.Error("generate plain text email", "error", err)
	}

	return Email{Subject: subject, Text: text, HTML: html}
}
