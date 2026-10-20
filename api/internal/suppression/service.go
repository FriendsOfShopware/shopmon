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

	"github.com/friendsofshopware/shopmon/api/internal/database"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/organization"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

// Authorizer reports a caller's role within an organization.
type Authorizer interface {
	RequireRole(ctx context.Context, userID, organizationID string, allowedRoles ...organization.Role) (organization.Role, error)
}

type Service struct {
	queries    *queries.Queries
	authorizer Authorizer
	now        func() time.Time
}

func NewService(q *queries.Queries, authorizer Authorizer) *Service {
	return &Service{queries: q, authorizer: authorizer, now: time.Now}
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

	shop, err := s.queries.GetShopByID(ctx, cmd.ShopID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Suppression{}, ErrShopNotFound
		}
		return Suppression{}, fmt.Errorf("get shop: %w", err)
	}

	if err := s.authorizeScope(ctx, cmd.UserID, shop.OrganizationID, cmd.EnvironmentID); err != nil {
		return Suppression{}, err
	}

	if cmd.EnvironmentID != nil {
		belongs, err := s.environmentBelongsToShop(ctx, *cmd.EnvironmentID, cmd.ShopID)
		if err != nil {
			return Suppression{}, err
		}
		if !belongs {
			return Suppression{}, ErrEnvironmentScope
		}
	}

	exists, err := s.advisoryExists(ctx, cmd.AdvisoryID)
	if err != nil {
		return Suppression{}, err
	}
	if !exists {
		return Suppression{}, ErrAdvisoryNotFound
	}

	created, err := s.create(ctx, Suppression{
		OrganizationID: shop.OrganizationID,
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
//
// Revoking carries the same role requirement as creating: withdrawing a
// shop-wide risk acceptance is an owner/admin decision, and letting any member
// undo one they could never have made would be a privilege asymmetry — the
// audit trail would record an acceptance nobody with the authority to make it
// chose to end.
func (s *Service) Revoke(ctx context.Context, id int64, userID string, organizationIDs []string) (Suppression, error) {
	if len(organizationIDs) == 0 {
		return Suppression{}, ErrNotFound
	}

	// Load restricted to the caller's organizations, so the revoke can be
	// authorized against its actual scope before mutating it.
	existing, err := s.find(ctx, id, organizationIDs)
	if err != nil {
		return Suppression{}, err
	}
	if err := s.authorizeScope(ctx, userID, existing.OrganizationID, existing.EnvironmentID); err != nil {
		return Suppression{}, err
	}

	revoked, err := s.revoke(ctx, id, userID, organizationIDs)
	if err != nil {
		return Suppression{}, err
	}
	return revoked, nil
}

// authorizeScope applies the role policy a suppression's blast radius demands:
// shop-wide is owner/admin, environment-scoped additionally allows members.
//
// Only genuine membership/role denials become ErrNotAuthorized. A database
// or context failure is returned unchanged, so an outage surfaces as an error
// rather than as a misleading 403.
func (s *Service) authorizeScope(ctx context.Context, userID, organizationID string, environmentID *int32) error {
	allowed := []organization.Role{organization.RoleOwner, organization.RoleAdmin}
	if environmentID != nil {
		allowed = append(allowed, organization.RoleMember)
	}
	if _, err := s.authorizer.RequireRole(ctx, userID, organizationID, allowed...); err != nil {
		if errors.Is(err, organization.ErrMembershipNotFound) || errors.Is(err, organization.ErrRoleNotAllowed) {
			return ErrNotAuthorized
		}
		return err
	}
	return nil
}

// List returns the suppressions visible to the caller's organizations.
func (s *Service) List(ctx context.Context, organizationIDs []string, includeInactive bool) ([]Suppression, error) {
	if len(organizationIDs) == 0 {
		return nil, nil
	}
	rows, err := s.queries.ListSuppressionsForOrganizations(ctx, queries.ListSuppressionsForOrganizationsParams{
		OrganizationIds: organizationIDs,
		IncludeInactive: includeInactive,
	})
	if err != nil {
		return nil, fmt.Errorf("list suppressions: %w", err)
	}

	out := make([]Suppression, 0, len(rows))
	for _, row := range rows {
		out = append(out, Suppression{
			ID:              row.ID,
			OrganizationID:  row.OrganizationID,
			ShopID:          row.ShopID,
			EnvironmentID:   row.EnvironmentID,
			AdvisoryID:      row.AdvisoryID,
			Reason:          row.Reason,
			ExpiresAt:       database.TimePtr(row.ExpiresAt),
			CreatedBy:       row.CreatedBy,
			CreatedAt:       row.CreatedAt.Time,
			RevokedAt:       database.TimePtr(row.RevokedAt),
			RevokedBy:       row.RevokedBy,
			ShopName:        row.ShopName,
			EnvironmentName: row.EnvironmentName,
			CreatedByName:   row.CreatedByName,
		})
	}
	return out, nil
}

// Active reports whether a suppression is currently in force.
func (s Suppression) Active(now time.Time) bool {
	if s.RevokedAt != nil {
		return false
	}
	return s.ExpiresAt == nil || s.ExpiresAt.After(now)
}

// uniqueViolation is the PostgreSQL error code for a unique constraint breach.
// The partial unique indexes on advisory_suppression turn a double-submit into
// this rather than the lost-update clobber the legacy ignores PATCH suffers.
const uniqueViolation = "23505"

func (s *Service) environmentBelongsToShop(ctx context.Context, environmentID, shopID int32) (bool, error) {
	row, err := s.queries.GetEnvironmentByID(ctx, environmentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("get environment: %w", err)
	}
	return row.ShopID == shopID, nil
}

func (s *Service) advisoryExists(ctx context.Context, advisoryID string) (bool, error) {
	if _, err := s.queries.GetComposerAdvisory(ctx, advisoryID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("get advisory: %w", err)
	}
	return true, nil
}

func (s *Service) create(ctx context.Context, record Suppression) (Suppression, error) {
	row, err := s.insert(ctx, record)
	if err == nil {
		out := toDomain(row)
		s.fillDisplayNames(ctx, &out)
		return out, nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != uniqueViolation {
		return Suppression{}, fmt.Errorf("insert suppression: %w", err)
	}

	// The partial unique indexes cannot express expiry (NOW() is not usable in
	// an index predicate), so the conflicting row may be an expired suppression
	// that every read path already treats as inactive. Retire it — a soft
	// revoke, preserving the audit trail — and retry once. If nothing expired
	// was blocking, the scope really is actively suppressed.
	retired, retireErr := s.queries.RevokeExpiredAdvisorySuppressions(ctx, queries.RevokeExpiredAdvisorySuppressionsParams{
		ShopID:        record.ShopID,
		AdvisoryID:    record.AdvisoryID,
		EnvironmentID: record.EnvironmentID,
		RevokedBy:     record.CreatedBy,
	})
	if retireErr != nil {
		return Suppression{}, fmt.Errorf("retire expired suppression: %w", retireErr)
	}
	if retired == 0 {
		return Suppression{}, ErrAlreadySuppressed
	}

	row, err = s.insert(ctx, record)
	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return Suppression{}, ErrAlreadySuppressed
		}
		return Suppression{}, fmt.Errorf("insert suppression: %w", err)
	}
	out := toDomain(row)
	s.fillDisplayNames(ctx, &out)
	return out, nil
}

// fillDisplayNames resolves the shop (and environment) names the API contract
// requires on a created suppression; the bare INSERT ... RETURNING cannot
// provide them. Best-effort: a lookup failure leaves the field empty rather
// than failing a write that already succeeded.
func (s *Service) fillDisplayNames(ctx context.Context, record *Suppression) {
	if shop, err := s.queries.GetShopByID(ctx, record.ShopID); err == nil {
		record.ShopName = shop.Name
	}
	if record.EnvironmentID != nil {
		if env, err := s.queries.GetEnvironmentByID(ctx, *record.EnvironmentID); err == nil {
			record.EnvironmentName = &env.Name
		}
	}
}

func (s *Service) insert(ctx context.Context, record Suppression) (queries.AdvisorySuppression, error) {
	return s.queries.CreateAdvisorySuppression(ctx, queries.CreateAdvisorySuppressionParams{
		OrganizationID: record.OrganizationID,
		ShopID:         record.ShopID,
		EnvironmentID:  record.EnvironmentID,
		AdvisoryID:     record.AdvisoryID,
		Reason:         record.Reason,
		ExpiresAt:      database.TimestampFromPtr(record.ExpiresAt),
		CreatedBy:      record.CreatedBy,
	})
}

func (s *Service) find(ctx context.Context, id int64, organizationIDs []string) (Suppression, error) {
	row, err := s.queries.GetAdvisorySuppressionInOrgs(ctx, queries.GetAdvisorySuppressionInOrgsParams{
		ID:              id,
		OrganizationIds: organizationIDs,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Suppression{}, ErrNotFound
		}
		return Suppression{}, fmt.Errorf("get suppression: %w", err)
	}
	return toDomain(row), nil
}

func (s *Service) revoke(ctx context.Context, id int64, userID string, organizationIDs []string) (Suppression, error) {
	row, err := s.queries.RevokeAdvisorySuppression(ctx, queries.RevokeAdvisorySuppressionParams{
		ID:              id,
		RevokedBy:       &userID,
		OrganizationIds: organizationIDs,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Suppression{}, ErrNotFound
		}
		return Suppression{}, fmt.Errorf("revoke suppression: %w", err)
	}
	return toDomain(row), nil
}

func toDomain(row queries.AdvisorySuppression) Suppression {
	return Suppression{
		ID:             row.ID,
		OrganizationID: row.OrganizationID,
		ShopID:         row.ShopID,
		EnvironmentID:  row.EnvironmentID,
		AdvisoryID:     row.AdvisoryID,
		Reason:         row.Reason,
		ExpiresAt:      database.TimePtr(row.ExpiresAt),
		CreatedBy:      row.CreatedBy,
		CreatedAt:      row.CreatedAt.Time,
		RevokedAt:      database.TimePtr(row.RevokedAt),
		RevokedBy:      row.RevokedBy,
	}
}
