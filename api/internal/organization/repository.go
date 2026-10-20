package organization

import (
	"context"
	"errors"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type store struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
}

func (r *store) CreateWithOwner(ctx context.Context, record Organization, memberID, ownerID, sessionToken string) error {
	return r.withTx(ctx, func(txq *queries.Queries) error {
		if _, err := txq.CreateOrganization(ctx, queries.CreateOrganizationParams{
			ID:   record.ID,
			Name: record.Name,
			Slug: record.Slug,
		}); err != nil {
			return fmt.Errorf("insert organization: %w", err)
		}
		if err := txq.CreateMember(ctx, queries.CreateMemberParams{
			ID:             memberID,
			OrganizationID: record.ID,
			UserID:         ownerID,
			Role:           string(RoleOwner),
		}); err != nil {
			return fmt.Errorf("insert owner membership: %w", err)
		}
		if err := txq.SetActiveOrganization(ctx, queries.SetActiveOrganizationParams{
			ActiveOrganizationID: &record.ID,
			Token:                sessionToken,
		}); err != nil {
			return fmt.Errorf("activate organization: %w", err)
		}
		return nil
	})
}

func (r *store) Find(ctx context.Context, organizationID string) (Organization, error) {
	row, err := r.queries.GetOrganizationByID(ctx, organizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Organization{}, ErrOrganizationNotFound
		}
		return Organization{}, fmt.Errorf("query organization: %w", err)
	}
	return Organization{
		ID:        row.ID,
		Name:      row.Name,
		Slug:      row.Slug,
		Logo:      row.Logo,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (r *store) Update(ctx context.Context, record Organization) error {
	if err := r.queries.UpdateOrganization(ctx, queries.UpdateOrganizationParams{
		Name: record.Name,
		Slug: record.Slug,
		Logo: record.Logo,
		ID:   record.ID,
	}); err != nil {
		return fmt.Errorf("execute organization update: %w", err)
	}
	return nil
}

func (r *store) Delete(ctx context.Context, organizationID string) error {
	if err := r.queries.DeleteOrganization(ctx, organizationID); err != nil {
		return fmt.Errorf("execute organization delete: %w", err)
	}
	return nil
}

func (r *store) CreateInvitation(ctx context.Context, invitation Invitation) error {
	var role *string
	if invitation.Role != nil {
		value := string(*invitation.Role)
		role = &value
	}
	if _, err := r.queries.CreateInvitation(ctx, queries.CreateInvitationParams{
		ID:             invitation.ID,
		OrganizationID: invitation.OrganizationID,
		Email:          invitation.Email,
		Role:           role,
		ExpiresAt:      pgtype.Timestamp{Time: invitation.ExpiresAt, Valid: true},
		InviterID:      invitation.InviterID,
	}); err != nil {
		return fmt.Errorf("insert organization invitation: %w", err)
	}
	return nil
}

func (r *store) FindInvitation(ctx context.Context, invitationID string) (Invitation, error) {
	row, err := r.queries.GetInvitationByID(ctx, invitationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Invitation{}, ErrInvitationNotFound
		}
		return Invitation{}, fmt.Errorf("query organization invitation: %w", err)
	}
	role, err := parseOptionalRole(row.Role)
	if err != nil {
		return Invitation{}, fmt.Errorf("decode organization invitation %q: %w", invitationID, err)
	}
	return Invitation{
		ID:             row.ID,
		OrganizationID: row.OrganizationID,
		Email:          row.Email,
		Role:           role,
		Status:         row.Status,
		ExpiresAt:      row.ExpiresAt.Time,
		InviterID:      row.InviterID,
		InviterName:    row.InviterName,
	}, nil
}

func (r *store) UpdateInvitationStatus(ctx context.Context, invitationID, status string) error {
	if err := r.queries.UpdateInvitationStatus(ctx, queries.UpdateInvitationStatusParams{Status: status, ID: invitationID}); err != nil {
		return fmt.Errorf("execute invitation status update: %w", err)
	}
	return nil
}

func (r *store) AcceptInvitation(ctx context.Context, invitationID, memberID, organizationID, userID string, role Role) error {
	return r.withTx(ctx, func(txq *queries.Queries) error {
		if err := txq.CreateMember(ctx, queries.CreateMemberParams{
			ID:             memberID,
			OrganizationID: organizationID,
			UserID:         userID,
			Role:           string(role),
		}); err != nil {
			return fmt.Errorf("insert invited member: %w", err)
		}
		if err := txq.UpdateInvitationStatus(ctx, queries.UpdateInvitationStatusParams{
			Status: "accepted",
			ID:     invitationID,
		}); err != nil {
			return fmt.Errorf("mark invitation accepted: %w", err)
		}
		return nil
	})
}

func (r *store) DeleteMember(ctx context.Context, organizationID, userID string) error {
	// notification_preference / user_notification reference environments by text
	// scope_id / link params (no foreign key), so they would otherwise linger after
	// the member is gone. Delete them in the same transaction as the membership.
	return r.withTx(ctx, func(txq *queries.Queries) error {
		if err := txq.DeleteMember(ctx, queries.DeleteMemberParams{
			OrganizationID: organizationID,
			UserID:         userID,
		}); err != nil {
			return fmt.Errorf("execute member delete: %w", err)
		}
		if err := txq.DeleteUserOrgNotificationPreferences(ctx, queries.DeleteUserOrgNotificationPreferencesParams{
			UserID:         userID,
			OrganizationID: organizationID,
		}); err != nil {
			return fmt.Errorf("delete org notification preferences: %w", err)
		}
		if err := txq.DeleteUserOrgNotifications(ctx, queries.DeleteUserOrgNotificationsParams{
			UserID:         userID,
			OrganizationID: organizationID,
		}); err != nil {
			return fmt.Errorf("delete org notifications: %w", err)
		}
		return nil
	})
}

func (r *store) CountOwners(ctx context.Context, organizationID string) (int32, error) {
	count, err := r.queries.CountOrgOwners(ctx, organizationID)
	if err != nil {
		return 0, fmt.Errorf("query owner count: %w", err)
	}
	return count, nil
}

func (r *store) UpdateMemberRole(ctx context.Context, organizationID, userID string, role Role) error {
	if err := r.queries.UpdateMemberRole(ctx, queries.UpdateMemberRoleParams{
		Role:           string(role),
		OrganizationID: organizationID,
		UserID:         userID,
	}); err != nil {
		return fmt.Errorf("execute member role update: %w", err)
	}
	return nil
}

func (r *store) ListMembers(ctx context.Context, organizationID string) ([]Member, error) {
	rows, err := r.queries.ListMembers(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("query organization members: %w", err)
	}
	members := make([]Member, 0, len(rows))
	for _, row := range rows {
		role, ok := ParseRole(row.Role)
		if !ok {
			return nil, fmt.Errorf("decode member %q: invalid role %q", row.ID, row.Role)
		}
		members = append(members, Member{
			ID:        row.ID,
			UserID:    row.UserID,
			Role:      role,
			Name:      row.UserName,
			Email:     row.UserEmail,
			Image:     row.UserImage,
			CreatedAt: row.CreatedAt.Time,
		})
	}
	return members, nil
}

func (r *store) ListInvitations(ctx context.Context, organizationID string) ([]Invitation, error) {
	rows, err := r.queries.ListInvitations(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("query organization invitations: %w", err)
	}
	invitations := make([]Invitation, 0, len(rows))
	for _, row := range rows {
		role, err := parseOptionalRole(row.Role)
		if err != nil {
			return nil, fmt.Errorf("decode invitation %q: %w", row.ID, err)
		}
		invitations = append(invitations, Invitation{
			ID:             row.ID,
			OrganizationID: row.OrganizationID,
			Email:          row.Email,
			Role:           role,
			Status:         row.Status,
			ExpiresAt:      row.ExpiresAt.Time,
			CreatedAt:      row.CreatedAt.Time,
			InviterID:      row.InviterID,
			InviterName:    row.InviterName,
		})
	}
	return invitations, nil
}

func (r *store) ListForUser(ctx context.Context, userID string) ([]Membership, error) {
	rows, err := r.queries.ListUserOrganizations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query organizations for user: %w", err)
	}
	memberships := make([]Membership, 0, len(rows))
	for _, row := range rows {
		role, ok := ParseRole(row.Role)
		if !ok {
			return nil, fmt.Errorf("decode organization %q membership: invalid role %q", row.ID, row.Role)
		}
		memberships = append(memberships, Membership{
			Organization: Organization{
				ID:        row.ID,
				Name:      row.Name,
				Slug:      row.Slug,
				Logo:      row.Logo,
				CreatedAt: row.CreatedAt.Time,
			},
			Role: role,
		})
	}
	return memberships, nil
}

func (r *store) DeleteInvitation(ctx context.Context, invitationID string) error {
	if err := r.queries.DeleteInvitationByID(ctx, invitationID); err != nil {
		return fmt.Errorf("execute invitation delete: %w", err)
	}
	return nil
}

func (r *store) SetActiveOrganization(ctx context.Context, sessionToken, organizationID string) error {
	if err := r.queries.SetActiveOrganization(ctx, queries.SetActiveOrganizationParams{
		ActiveOrganizationID: &organizationID,
		Token:                sessionToken,
	}); err != nil {
		return fmt.Errorf("execute active organization update: %w", err)
	}
	return nil
}

func (r *store) withTx(ctx context.Context, operation func(*queries.Queries) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := operation(r.queries.WithTx(tx)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func parseOptionalRole(value *string) (*Role, error) {
	if value == nil {
		return nil, nil
	}
	role, ok := ParseRole(*value)
	if !ok {
		return nil, fmt.Errorf("invalid role %q", *value)
	}
	return &role, nil
}
