package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *AuthHandler) createOrganizationFlow(ctx context.Context, userID, sessionToken, name string) (string, error) {
	orgID := uuid.New().String()

	err := h.withTx(ctx, func(txq *queries.Queries) error {
		_, err := txq.CreateOrganization(ctx, queries.CreateOrganizationParams{
			ID:   orgID,
			Name: name,
			Slug: orgID,
		})
		if err != nil {
			return fmt.Errorf("create organization: %w", err)
		}

		if err := txq.CreateMember(ctx, queries.CreateMemberParams{
			ID:             uuid.New().String(),
			OrganizationID: orgID,
			UserID:         userID,
			Role:           "owner",
		}); err != nil {
			return fmt.Errorf("create owner membership: %w", err)
		}

		if err := txq.SetActiveOrganization(ctx, queries.SetActiveOrganizationParams{
			ActiveOrganizationID: &orgID,
			Token:                sessionToken,
		}); err != nil {
			return fmt.Errorf("set active organization: %w", err)
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return orgID, nil
}

type inviteMemberCommand struct {
	OrganizationID string
	InviterID      string
	InviterName    string
	Email          string
	Role           string
}

func (h *AuthHandler) inviteMemberFlow(ctx context.Context, cmd inviteMemberCommand) (string, error) {
	invitationID := uuid.New().String()
	role := cmd.Role

	_, err := h.queries.CreateInvitation(ctx, queries.CreateInvitationParams{
		ID:             invitationID,
		OrganizationID: cmd.OrganizationID,
		Email:          cmd.Email,
		Role:           &role,
		ExpiresAt:      pgtype.Timestamp{Time: time.Now().Add(7 * 24 * time.Hour), Valid: true},
		InviterID:      cmd.InviterID,
	})
	if err != nil {
		return "", fmt.Errorf("create invitation: %w", err)
	}

	orgName := cmd.OrganizationID
	if org, err := h.queries.GetOrganizationByID(ctx, cmd.OrganizationID); err == nil {
		orgName = org.Name
	}

	acceptURL := h.cfg.FrontendURL + "/app/organizations/accept/" + invitationID
	rejectURL := h.cfg.FrontendURL + "/app/organizations/reject/" + invitationID
	if err := h.mail.Send(cmd.Email,
		"You have been invited to join "+orgName+" at Shopmon",
		mail.BuildOrgInviteEmail(cmd.InviterName, orgName, acceptURL, rejectURL)); err != nil {
		slog.Error("failed to send organization invite email", "error", err, "email", cmd.Email)
	}

	return invitationID, nil
}

func (h *AuthHandler) acceptInvitationFlow(ctx context.Context, invitationID, organizationID, userID string, invitationRole *string) error {
	role := "member"
	if invitationRole != nil {
		role = *invitationRole
	}

	return h.withTx(ctx, func(txq *queries.Queries) error {
		if err := txq.CreateMember(ctx, queries.CreateMemberParams{
			ID:             uuid.New().String(),
			OrganizationID: organizationID,
			UserID:         userID,
			Role:           role,
		}); err != nil {
			return fmt.Errorf("create member: %w", err)
		}

		if err := txq.UpdateInvitationStatus(ctx, queries.UpdateInvitationStatusParams{
			Status: "accepted",
			ID:     invitationID,
		}); err != nil {
			return fmt.Errorf("update invitation status: %w", err)
		}

		return nil
	})
}
