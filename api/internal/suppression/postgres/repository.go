package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/friendsofshopware/shopmon/api/internal/database"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/suppression"
)

// uniqueViolation is the PostgreSQL error code for a unique constraint breach.
// The partial unique indexes on advisory_suppression turn a double-submit into
// this rather than the lost-update clobber the legacy ignores PATCH suffers.
const uniqueViolation = "23505"

type Repository struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
}

var _ suppression.Repository = (*Repository)(nil)

func NewRepository(pool *pgxpool.Pool, q *queries.Queries) *Repository {
	return &Repository{pool: pool, queries: q}
}

func (r *Repository) FindShopScope(ctx context.Context, shopID int32) (suppression.ShopScope, error) {
	row, err := r.queries.GetShopByID(ctx, shopID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return suppression.ShopScope{}, suppression.ErrShopNotFound
		}
		return suppression.ShopScope{}, fmt.Errorf("get shop: %w", err)
	}
	return suppression.ShopScope{ShopID: row.ID, OrganizationID: row.OrganizationID}, nil
}

func (r *Repository) EnvironmentBelongsToShop(ctx context.Context, environmentID, shopID int32) (bool, error) {
	row, err := r.queries.GetEnvironmentByID(ctx, environmentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("get environment: %w", err)
	}
	return row.ShopID == shopID, nil
}

func (r *Repository) AdvisoryExists(ctx context.Context, advisoryID string) (bool, error) {
	if _, err := r.queries.GetComposerAdvisory(ctx, advisoryID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("get advisory: %w", err)
	}
	return true, nil
}

func (r *Repository) Create(ctx context.Context, record suppression.Suppression) (suppression.Suppression, error) {
	row, err := r.insert(ctx, record)
	if err == nil {
		out := toDomain(row)
		r.fillDisplayNames(ctx, &out)
		return out, nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != uniqueViolation {
		return suppression.Suppression{}, fmt.Errorf("insert suppression: %w", err)
	}

	// The partial unique indexes cannot express expiry (NOW() is not usable in
	// an index predicate), so the conflicting row may be an expired suppression
	// that every read path already treats as inactive. Retire it — a soft
	// revoke, preserving the audit trail — and retry once. If nothing expired
	// was blocking, the scope really is actively suppressed.
	retired, retireErr := r.queries.RevokeExpiredAdvisorySuppressions(ctx, queries.RevokeExpiredAdvisorySuppressionsParams{
		ShopID:        record.ShopID,
		AdvisoryID:    record.AdvisoryID,
		EnvironmentID: record.EnvironmentID,
		RevokedBy:     record.CreatedBy,
	})
	if retireErr != nil {
		return suppression.Suppression{}, fmt.Errorf("retire expired suppression: %w", retireErr)
	}
	if retired == 0 {
		return suppression.Suppression{}, suppression.ErrAlreadySuppressed
	}

	row, err = r.insert(ctx, record)
	if err != nil {
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return suppression.Suppression{}, suppression.ErrAlreadySuppressed
		}
		return suppression.Suppression{}, fmt.Errorf("insert suppression: %w", err)
	}
	out := toDomain(row)
	r.fillDisplayNames(ctx, &out)
	return out, nil
}

// fillDisplayNames resolves the shop (and environment) names the API contract
// requires on a created suppression; the bare INSERT ... RETURNING cannot
// provide them. Best-effort: a lookup failure leaves the field empty rather
// than failing a write that already succeeded.
func (r *Repository) fillDisplayNames(ctx context.Context, s *suppression.Suppression) {
	if shop, err := r.queries.GetShopByID(ctx, s.ShopID); err == nil {
		s.ShopName = shop.Name
	}
	if s.EnvironmentID != nil {
		if env, err := r.queries.GetEnvironmentByID(ctx, *s.EnvironmentID); err == nil {
			s.EnvironmentName = &env.Name
		}
	}
}

func (r *Repository) insert(ctx context.Context, record suppression.Suppression) (queries.AdvisorySuppression, error) {
	return r.queries.CreateAdvisorySuppression(ctx, queries.CreateAdvisorySuppressionParams{
		OrganizationID: record.OrganizationID,
		ShopID:         record.ShopID,
		EnvironmentID:  record.EnvironmentID,
		AdvisoryID:     record.AdvisoryID,
		Reason:         record.Reason,
		ExpiresAt:      database.TimestampFromPtr(record.ExpiresAt),
		CreatedBy:      record.CreatedBy,
	})
}

func (r *Repository) Find(ctx context.Context, id int64, organizationIDs []string) (suppression.Suppression, error) {
	row, err := r.queries.GetAdvisorySuppressionInOrgs(ctx, queries.GetAdvisorySuppressionInOrgsParams{
		ID:              id,
		OrganizationIds: organizationIDs,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return suppression.Suppression{}, suppression.ErrNotFound
		}
		return suppression.Suppression{}, fmt.Errorf("get suppression: %w", err)
	}
	return toDomain(row), nil
}

func (r *Repository) Revoke(ctx context.Context, id int64, userID string, organizationIDs []string) (suppression.Suppression, error) {
	row, err := r.queries.RevokeAdvisorySuppression(ctx, queries.RevokeAdvisorySuppressionParams{
		ID:              id,
		RevokedBy:       &userID,
		OrganizationIds: organizationIDs,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return suppression.Suppression{}, suppression.ErrNotFound
		}
		return suppression.Suppression{}, fmt.Errorf("revoke suppression: %w", err)
	}
	return toDomain(row), nil
}

func (r *Repository) ListForOrganizations(ctx context.Context, organizationIDs []string, includeInactive bool) ([]suppression.Suppression, error) {
	rows, err := r.queries.ListSuppressionsForOrganizations(ctx, queries.ListSuppressionsForOrganizationsParams{
		OrganizationIds: organizationIDs,
		IncludeInactive: includeInactive,
	})
	if err != nil {
		return nil, fmt.Errorf("list suppressions: %w", err)
	}

	out := make([]suppression.Suppression, 0, len(rows))
	for _, row := range rows {
		out = append(out, suppression.Suppression{
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

func toDomain(row queries.AdvisorySuppression) suppression.Suppression {
	return suppression.Suppression{
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
