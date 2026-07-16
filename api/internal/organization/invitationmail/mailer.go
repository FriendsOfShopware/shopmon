package invitationmail

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

func (m *Mailer) SendInvitation(ctx context.Context, recipient, inviterName, organizationName, invitationID string) error {
	acceptURL := m.frontendURL + "/app/organizations/accept/" + invitationID
	rejectURL := m.frontendURL + "/app/organizations/reject/" + invitationID
	email := m.sender.BuildOrgInviteEmail(inviterName, organizationName, acceptURL, rejectURL)
	if err := m.sender.Send(ctx, recipient, email); err != nil {
		return fmt.Errorf("send organization invitation email: %w", err)
	}
	return nil
}
