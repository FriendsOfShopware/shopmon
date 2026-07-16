package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/organization"
	"github.com/jackc/pgx/v5"
)

// AuthorizationRepository adapts sqlc queries to organization access decisions.
type AuthorizationRepository struct {
	queries *queries.Queries
}

var _ organization.AuthorizationRepository = (*AuthorizationRepository)(nil)

func NewAuthorizationRepository(q *queries.Queries) *AuthorizationRepository {
	return &AuthorizationRepository{queries: q}
}

func (r *AuthorizationRepository) OrganizationExists(ctx context.Context, organizationID string) (bool, error) {
	if _, err := r.queries.GetOrganizationByID(ctx, organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("get organization: %w", err)
	}
	return true, nil
}

func (r *AuthorizationRepository) FindRole(ctx context.Context, userID, organizationID string) (organization.Role, error) {
	value, err := r.queries.GetMemberRole(ctx, queries.GetMemberRoleParams{
		OrganizationID: organizationID,
		UserID:         userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", organization.ErrMembershipNotFound
		}
		return "", fmt.Errorf("get member role: %w", err)
	}

	role, ok := organization.ParseRole(value)
	if !ok {
		return "", fmt.Errorf("invalid organization role %q", value)
	}
	return role, nil
}
