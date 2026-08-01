// Package suppression records a shop owner's decision that a known security
// advisory is accepted or mitigated outside Shopmon — a WAF rule, an unused
// component, or another compensating control.
//
// It is deliberately separate from environment.ignores: that column is a flat
// array of check ids with no room for a reason, an actor, an expiry, or a
// shop-wide scope, and it is rewritten wholesale by the generic environment
// PATCH. Both mechanisms coexist and are unioned where checks are evaluated.
package suppression

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/organization"
)

var (
	ErrNotAuthorized     = errors.New("not authorized for organization")
	ErrAdvisoryNotFound  = errors.New("advisory not found")
	ErrShopNotFound      = errors.New("shop not found")
	ErrEnvironmentScope  = errors.New("environment does not belong to shop")
	ErrAlreadySuppressed = errors.New("advisory is already suppressed for this scope")
	ErrNotFound          = errors.New("suppression not found")
	ErrReasonRequired    = errors.New("a reason is required")
	ErrExpiryInPast      = errors.New("expiry must be in the future")
)

// Suppression is one recorded risk acceptance.
type Suppression struct {
	ID             int64
	OrganizationID string
	ShopID         int32
	// EnvironmentID nil means every environment of the shop.
	EnvironmentID *int32
	AdvisoryID    string
	Reason        string
	ExpiresAt     *time.Time
	CreatedBy     *string
	CreatedAt     time.Time
	RevokedAt     *time.Time
	RevokedBy     *string

	// Display fields, populated by list queries that join the related rows.
	// Empty on the records returned by Create/Revoke.
	ShopName        string
	EnvironmentName *string
	CreatedByName   *string
}

// CreateCommand is the input for recording a suppression.
type CreateCommand struct {
	UserID        string
	ShopID        int32
	EnvironmentID *int32
	AdvisoryID    string
	Reason        string
	ExpiresAt     *time.Time
}

// ShopScope is the ownership information the service needs to authorize and
// validate a suppression without importing persistence types.
type ShopScope struct {
	ShopID         int32
	OrganizationID string
}

// Repository is the persistence port.
type Repository interface {
	FindShopScope(ctx context.Context, shopID int32) (ShopScope, error)
	EnvironmentBelongsToShop(ctx context.Context, environmentID, shopID int32) (bool, error)
	AdvisoryExists(ctx context.Context, advisoryID string) (bool, error)
	Create(ctx context.Context, record Suppression) (Suppression, error)
	Revoke(ctx context.Context, id int64, userID string, organizationIDs []string) (Suppression, error)
	ListForOrganizations(ctx context.Context, organizationIDs []string, includeInactive bool) ([]Suppression, error)
}

// Authorizer reports a caller's role within an organization.
type Authorizer interface {
	RequireRole(ctx context.Context, userID, organizationID string, allowedRoles ...organization.Role) (organization.Role, error)
}

type Service struct {
	repository Repository
	authorizer Authorizer
	now        func() time.Time
}

func NewService(repository Repository, authorizer Authorizer) *Service {
	return &Service{repository: repository, authorizer: authorizer, now: time.Now}
}

// Create records a suppression after validating the scope and the caller's role.
//
// Suppressing is a risk-acceptance decision with blast radius across a whole
// shop, so it requires owner or admin. Narrowing to a single environment is the
// same decision at a smaller scale and is left to any member.
func (s *Service) Create(ctx context.Context, cmd CreateCommand) (Suppression, error) {
	reason := strings.TrimSpace(cmd.Reason)
	if reason == "" {
		return Suppression{}, ErrReasonRequired
	}
	if cmd.ExpiresAt != nil && !cmd.ExpiresAt.After(s.now().UTC()) {
		return Suppression{}, ErrExpiryInPast
	}

	scope, err := s.repository.FindShopScope(ctx, cmd.ShopID)
	if err != nil {
		return Suppression{}, err
	}

	allowed := []organization.Role{organization.RoleOwner, organization.RoleAdmin}
	if cmd.EnvironmentID != nil {
		allowed = append(allowed, organization.RoleMember)
	}
	if _, err := s.authorizer.RequireRole(ctx, cmd.UserID, scope.OrganizationID, allowed...); err != nil {
		return Suppression{}, ErrNotAuthorized
	}

	if cmd.EnvironmentID != nil {
		belongs, err := s.repository.EnvironmentBelongsToShop(ctx, *cmd.EnvironmentID, cmd.ShopID)
		if err != nil {
			return Suppression{}, err
		}
		if !belongs {
			return Suppression{}, ErrEnvironmentScope
		}
	}

	exists, err := s.repository.AdvisoryExists(ctx, cmd.AdvisoryID)
	if err != nil {
		return Suppression{}, err
	}
	if !exists {
		return Suppression{}, ErrAdvisoryNotFound
	}

	created, err := s.repository.Create(ctx, Suppression{
		OrganizationID: scope.OrganizationID,
		ShopID:         cmd.ShopID,
		EnvironmentID:  cmd.EnvironmentID,
		AdvisoryID:     cmd.AdvisoryID,
		Reason:         reason,
		ExpiresAt:      cmd.ExpiresAt,
		CreatedBy:      &cmd.UserID,
	})
	if err != nil {
		return Suppression{}, fmt.Errorf("create suppression: %w", err)
	}
	return created, nil
}

// Revoke soft-deletes a suppression, keeping the row as an audit record. The
// organization scope is applied in the query so a caller cannot revoke another
// tenant's suppression by guessing an id.
func (s *Service) Revoke(ctx context.Context, id int64, userID string, organizationIDs []string) (Suppression, error) {
	if len(organizationIDs) == 0 {
		return Suppression{}, ErrNotFound
	}
	revoked, err := s.repository.Revoke(ctx, id, userID, organizationIDs)
	if err != nil {
		return Suppression{}, err
	}
	return revoked, nil
}

// List returns the suppressions visible to the caller's organizations.
func (s *Service) List(ctx context.Context, organizationIDs []string, includeInactive bool) ([]Suppression, error) {
	if len(organizationIDs) == 0 {
		return nil, nil
	}
	return s.repository.ListForOrganizations(ctx, organizationIDs, includeInactive)
}

// Active reports whether a suppression is currently in force.
func (s Suppression) Active(now time.Time) bool {
	if s.RevokedAt != nil {
		return false
	}
	return s.ExpiresAt == nil || s.ExpiresAt.After(now)
}
