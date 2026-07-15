package credentialmail

import (
	"context"
	"fmt"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/mail"
)

type Mailer struct {
	sender      mail.Sender
	frontendURL string
}

func New(sender mail.Sender, frontendURL string) *Mailer {
	return &Mailer{sender: sender, frontendURL: strings.TrimSuffix(frontendURL, "/")}
}

func (m *Mailer) SendConfirmation(ctx context.Context, recipient, name, token string) error {
	message := m.sender.BuildConfirmationEmail(name, m.frontendURL+"/account/confirm/"+token)
	if err := m.sender.Send(ctx, recipient, message); err != nil {
		return fmt.Errorf("send confirmation email: %w", err)
	}
	return nil
}

func (m *Mailer) SendPasswordReset(ctx context.Context, recipient, name, token string) error {
	message := m.sender.BuildPasswordResetEmail(name, m.frontendURL+"/account/forgot-password/"+token)
	if err := m.sender.Send(ctx, recipient, message); err != nil {
		return fmt.Errorf("send password reset email: %w", err)
	}
	return nil
}
