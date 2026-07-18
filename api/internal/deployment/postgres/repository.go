package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/deployment"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
}

var _ deployment.Repository = (*Repository)(nil)

func NewRepository(pool *pgxpool.Pool, q *queries.Queries) *Repository {
	return &Repository{pool: pool, queries: q}
}

func (r *Repository) FindEnvironment(ctx context.Context, environmentID int32) (deployment.Environment, error) {
	row, err := r.queries.GetEnvironmentByID(ctx, environmentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return deployment.Environment{}, deployment.ErrEnvironmentNotFound
		}
		return deployment.Environment{}, fmt.Errorf("get environment: %w", err)
	}
	return deployment.Environment{ID: row.ID, OrganizationID: row.OrganizationID, ShopID: row.ShopID}, nil
}

func (r *Repository) OrganizationIDForShop(ctx context.Context, shopID int32) (string, error) {
	row, err := r.queries.GetShopByID(ctx, shopID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", deployment.ErrShopNotFound
		}
		return "", fmt.Errorf("get shop: %w", err)
	}
	return row.OrganizationID, nil
}

func (r *Repository) ListDeployments(ctx context.Context, environmentID, limit, offset int32) ([]deployment.Deployment, error) {
	rows, err := r.queries.ListDeployments(ctx, queries.ListDeploymentsParams{
		EnvironmentID: environmentID,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list deployments: %w", err)
	}
	result := make([]deployment.Deployment, 0, len(rows))
	for _, row := range rows {
		result = append(result, deployment.Deployment{
			ID:            row.ID,
			EnvironmentID: row.EnvironmentID,
			Name:          row.Name,
			Command:       row.Command,
			ReturnCode:    row.ReturnCode,
			StartDate:     row.StartDate.Time,
			EndDate:       row.EndDate.Time,
			ExecutionTime: row.ExecutionTime,
			Reference:     row.Reference,
			CreatedAt:     row.CreatedAt.Time,
		})
	}
	return result, nil
}

func (r *Repository) FindDeployment(ctx context.Context, environmentID, deploymentID int32) (deployment.Deployment, error) {
	row, err := r.queries.GetDeploymentByID(ctx, queries.GetDeploymentByIDParams{ID: deploymentID, EnvironmentID: environmentID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return deployment.Deployment{}, deployment.ErrDeploymentNotFound
		}
		return deployment.Deployment{}, fmt.Errorf("get deployment: %w", err)
	}
	return deployment.Deployment{
		ID:            row.ID,
		EnvironmentID: row.EnvironmentID,
		Name:          row.Name,
		Command:       row.Command,
		ReturnCode:    row.ReturnCode,
		StartDate:     row.StartDate.Time,
		EndDate:       row.EndDate.Time,
		ExecutionTime: row.ExecutionTime,
		Composer:      row.Composer,
		Reference:     row.Reference,
		CreatedAt:     row.CreatedAt.Time,
	}, nil
}

func (r *Repository) DeleteDeployment(ctx context.Context, environmentID, deploymentID int32) error {
	if err := r.queries.DeleteDeployment(ctx, queries.DeleteDeploymentParams{ID: deploymentID, EnvironmentID: environmentID}); err != nil {
		return fmt.Errorf("delete deployment: %w", err)
	}
	return nil
}

func (r *Repository) CreateDeployment(ctx context.Context, record deployment.CreateDeploymentRecord) (int32, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin deployment transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	txq := r.queries.WithTx(tx)

	deploymentID, err := txq.CreateDeployment(ctx, queries.CreateDeploymentParams{
		EnvironmentID: record.EnvironmentID,
		Name:          record.Name,
		Command:       record.Command,
		ReturnCode:    record.ReturnCode,
		StartDate:     pgtype.Timestamp{Time: record.StartDate, Valid: true},
		EndDate:       pgtype.Timestamp{Time: record.EndDate, Valid: true},
		ExecutionTime: record.ExecutionTime,
		Composer:      record.Composer,
		Reference:     record.Reference,
	})
	if err != nil {
		return 0, fmt.Errorf("create deployment: %w", err)
	}
	if err := txq.UpdateEnvironmentActiveDeployment(ctx, queries.UpdateEnvironmentActiveDeploymentParams{
		ActiveDeploymentID: &deploymentID,
		ID:                 record.EnvironmentID,
	}); err != nil {
		return 0, fmt.Errorf("set active deployment: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit deployment transaction: %w", err)
	}
	return deploymentID, nil
}

func (r *Repository) ListAPIKeys(ctx context.Context, shopID int32) ([]deployment.APIKey, error) {
	rows, err := r.queries.ListApiKeysByShop(ctx, shopID)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	result := make([]deployment.APIKey, 0, len(rows))
	for _, row := range rows {
		scopes, err := decodeScopes(row.Scopes)
		if err != nil {
			return nil, fmt.Errorf("decode api key %q scopes: %w", row.ID, err)
		}
		result = append(result, deployment.APIKey{
			ID:         row.ID,
			ShopID:     row.ShopID,
			Name:       row.Name,
			Scopes:     scopes,
			CreatedAt:  row.CreatedAt.Time,
			LastUsedAt: database.TimePtr(row.LastUsedAt),
		})
	}
	return result, nil
}

func (r *Repository) CreateAPIKey(ctx context.Context, key deployment.APIKey) error {
	scopes, err := json.Marshal(key.Scopes)
	if err != nil {
		return fmt.Errorf("encode api key scopes: %w", err)
	}
	if _, err := r.queries.CreateApiKey(ctx, queries.CreateApiKeyParams{
		ID:     key.ID,
		ShopID: key.ShopID,
		Name:   key.Name,
		Token:  key.TokenHash,
		Scopes: scopes,
	}); err != nil {
		return fmt.Errorf("create api key: %w", err)
	}
	return nil
}

func (r *Repository) DeleteAPIKey(ctx context.Context, shopID int32, keyID string) error {
	if err := r.queries.DeleteApiKey(ctx, queries.DeleteApiKeyParams{ID: keyID, ShopID: shopID}); err != nil {
		return fmt.Errorf("delete api key: %w", err)
	}
	return nil
}

func (r *Repository) FindAPIKeyByTokenHash(ctx context.Context, tokenHash string) (deployment.APIKey, error) {
	row, err := r.queries.GetApiKeyByToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return deployment.APIKey{}, deployment.ErrAPIKeyInvalid
		}
		return deployment.APIKey{}, fmt.Errorf("get api key: %w", err)
	}
	scopes, err := decodeScopes(row.Scopes)
	if err != nil {
		return deployment.APIKey{}, fmt.Errorf("decode api key scopes: %w", err)
	}
	return deployment.APIKey{
		ID:             row.ID,
		ShopID:         row.ShopID,
		OrganizationID: row.OrganizationID,
		Name:           row.Name,
		TokenHash:      row.Token,
		Scopes:         scopes,
		CreatedAt:      row.CreatedAt.Time,
		LastUsedAt:     database.TimePtr(row.LastUsedAt),
	}, nil
}

func (r *Repository) TouchAPIKey(ctx context.Context, keyID string) error {
	if err := r.queries.UpdateApiKeyLastUsed(ctx, keyID); err != nil {
		return fmt.Errorf("update api key last used: %w", err)
	}
	return nil
}

func decodeScopes(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 {
		return []string{}, nil
	}
	var scopes []string
	if err := json.Unmarshal(raw, &scopes); err != nil {
		return nil, err
	}
	if scopes == nil {
		return []string{}, nil
	}
	return scopes, nil
}
