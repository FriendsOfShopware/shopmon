package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/organization"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
}

var _ organization.Repository = (*Repository)(nil)

func NewRepository(pool *pgxpool.Pool, q *queries.Queries) *Repository {
	return &Repository{pool: pool, queries: q}
}

func (r *Repository) CreateWithOwner(ctx context.Context, record organization.Organization, memberID, ownerID, sessionToken string) error {
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
			Role:           string(organization.RoleOwner),
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

func (r *Repository) Find(ctx context.Context, organizationID string) (organization.Organization, error) {
	row, err := r.queries.GetOrganizationByID(ctx, organizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return organization.Organization{}, organization.ErrOrganizationNotFound
		}
		return organization.Organization{}, fmt.Errorf("query organization: %w", err)
	}
	return organization.Organization{
		ID:        row.ID,
		Name:      row.Name,
		Slug:      row.Slug,
		Logo:      row.Logo,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (r *Repository) Update(ctx context.Context, record organization.Organization) error {
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

func (r *Repository) Delete(ctx context.Context, organizationID string) error {
	if err := r.queries.DeleteOrganization(ctx, organizationID); err != nil {
		return fmt.Errorf("execute organization delete: %w", err)
	}
	return nil
}

func (r *Repository) CreateInvitation(ctx context.Context, invitation organization.Invitation) error {
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

func (r *Repository) FindInvitation(ctx context.Context, invitationID string) (organization.Invitation, error) {
	row, err := r.queries.GetInvitationByID(ctx, invitationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return organization.Invitation{}, organization.ErrInvitationNotFound
		}
		return organization.Invitation{}, fmt.Errorf("query organization invitation: %w", err)
	}
	role, err := parseOptionalRole(row.Role)
	if err != nil {
		return organization.Invitation{}, fmt.Errorf("decode organization invitation %q: %w", invitationID, err)
	}
	return organization.Invitation{
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

func (r *Repository) UpdateInvitationStatus(ctx context.Context, invitationID, status string) error {
	if err := r.queries.UpdateInvitationStatus(ctx, queries.UpdateInvitationStatusParams{Status: status, ID: invitationID}); err != nil {
		return fmt.Errorf("execute invitation status update: %w", err)
	}
	return nil
}

func (r *Repository) AcceptInvitation(ctx context.Context, invitationID, memberID, organizationID, userID string, role organization.Role) error {
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

func (r *Repository) DeleteMember(ctx context.Context, organizationID, userID string) error {
	if err := r.queries.DeleteMember(ctx, queries.DeleteMemberParams{OrganizationID: organizationID, UserID: userID}); err != nil {
		return fmt.Errorf("execute member delete: %w", err)
	}
	return nil
}

func (r *Repository) CountOwners(ctx context.Context, organizationID string) (int32, error) {
	count, err := r.queries.CountOrgOwners(ctx, organizationID)
	if err != nil {
		return 0, fmt.Errorf("query owner count: %w", err)
	}
	return count, nil
}

func (r *Repository) UpdateMemberRole(ctx context.Context, organizationID, userID string, role organization.Role) error {
	if err := r.queries.UpdateMemberRole(ctx, queries.UpdateMemberRoleParams{
		Role:           string(role),
		OrganizationID: organizationID,
		UserID:         userID,
	}); err != nil {
		return fmt.Errorf("execute member role update: %w", err)
	}
	return nil
}

func (r *Repository) ListMembers(ctx context.Context, organizationID string) ([]organization.Member, error) {
	rows, err := r.queries.ListMembers(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("query organization members: %w", err)
	}
	members := make([]organization.Member, 0, len(rows))
	for _, row := range rows {
		role, ok := organization.ParseRole(row.Role)
		if !ok {
			return nil, fmt.Errorf("decode member %q: invalid role %q", row.ID, row.Role)
		}
		members = append(members, organization.Member{
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

func (r *Repository) ListInvitations(ctx context.Context, organizationID string) ([]organization.Invitation, error) {
	rows, err := r.queries.ListInvitations(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("query organization invitations: %w", err)
	}
	invitations := make([]organization.Invitation, 0, len(rows))
	for _, row := range rows {
		role, err := parseOptionalRole(row.Role)
		if err != nil {
			return nil, fmt.Errorf("decode invitation %q: %w", row.ID, err)
		}
		invitations = append(invitations, organization.Invitation{
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

func (r *Repository) ListForUser(ctx context.Context, userID string) ([]organization.Membership, error) {
	rows, err := r.queries.ListUserOrganizations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query organizations for user: %w", err)
	}
	memberships := make([]organization.Membership, 0, len(rows))
	for _, row := range rows {
		role, ok := organization.ParseRole(row.Role)
		if !ok {
			return nil, fmt.Errorf("decode organization %q membership: invalid role %q", row.ID, row.Role)
		}
		memberships = append(memberships, organization.Membership{
			Organization: organization.Organization{
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

func (r *Repository) DeleteInvitation(ctx context.Context, invitationID string) error {
	if err := r.queries.DeleteInvitationByID(ctx, invitationID); err != nil {
		return fmt.Errorf("execute invitation delete: %w", err)
	}
	return nil
}

func (r *Repository) SetActiveOrganization(ctx context.Context, sessionToken, organizationID string) error {
	if err := r.queries.SetActiveOrganization(ctx, queries.SetActiveOrganizationParams{
		ActiveOrganizationID: &organizationID,
		Token:                sessionToken,
	}); err != nil {
		return fmt.Errorf("execute active organization update: %w", err)
	}
	return nil
}

func (r *Repository) withTx(ctx context.Context, operation func(*queries.Queries) error) error {
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

func parseOptionalRole(value *string) (*organization.Role, error) {
	if value == nil {
		return nil, nil
	}
	role, ok := organization.ParseRole(*value)
	if !ok {
		return nil, fmt.Errorf("invalid role %q", *value)
	}
	return &role, nil
}
