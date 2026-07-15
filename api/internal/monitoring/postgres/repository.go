package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/monitoring"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
}

var _ monitoring.Repository = (*Repository)(nil)

func NewRepository(pool *pgxpool.Pool, q *queries.Queries) *Repository {
	return &Repository{pool: pool, queries: q}
}

func (r *Repository) FindShop(ctx context.Context, shopID int32) (monitoring.Shop, error) {
	row, err := r.queries.GetShopByID(ctx, shopID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return monitoring.Shop{}, monitoring.ErrShopNotFound
		}
		return monitoring.Shop{}, fmt.Errorf("get shop: %w", err)
	}
	return monitoring.Shop{
		ID:                   row.ID,
		OrganizationID:       row.OrganizationID,
		Name:                 row.Name,
		Description:          row.Description,
		GitURL:               row.GitUrl,
		DefaultEnvironmentID: row.DefaultEnvironmentID,
	}, nil
}

func (r *Repository) FindEnvironment(ctx context.Context, environmentID int32) (monitoring.Environment, error) {
	row, err := r.queries.GetEnvironmentByID(ctx, environmentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return monitoring.Environment{}, monitoring.ErrEnvironmentNotFound
		}
		return monitoring.Environment{}, fmt.Errorf("get environment: %w", err)
	}

	var ignores []string
	if row.Ignores != nil {
		if err := json.Unmarshal(row.Ignores, &ignores); err != nil {
			return monitoring.Environment{}, fmt.Errorf("decode environment ignores: %w", err)
		}
	}

	return monitoring.Environment{
		ID:                    row.ID,
		OrganizationID:        row.OrganizationID,
		ShopID:                row.ShopID,
		Name:                  row.Name,
		URL:                   row.Url,
		ClientID:              row.ClientID,
		EncryptedClientSecret: row.ClientSecret,
		ShopwareVersion:       row.ShopwareVersion,
		EnvironmentToken:      row.EnvironmentToken,
		Ignores:               ignores,
	}, nil
}

func (r *Repository) CreateShopWithEnvironment(ctx context.Context, shop monitoring.CreateShopRecord, environment monitoring.CreateEnvironmentRecord) (int32, int32, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, 0, fmt.Errorf("begin create shop transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	txq := r.queries.WithTx(tx)

	shopID, err := txq.CreateShop(ctx, queries.CreateShopParams{
		OrganizationID: shop.OrganizationID,
		Name:           shop.Name,
		Description:    shop.Description,
		GitUrl:         shop.GitURL,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("create shop: %w", err)
	}
	environment.ShopID = shopID
	environmentID, err := createEnvironment(ctx, txq, environment)
	if err != nil {
		return 0, 0, err
	}
	if err := txq.SetShopDefaultEnvironment(ctx, queries.SetShopDefaultEnvironmentParams{
		DefaultEnvironmentID: &environmentID,
		ID:                   shopID,
		OrganizationID:       shop.OrganizationID,
	}); err != nil {
		return 0, 0, fmt.Errorf("set default environment: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("commit create shop transaction: %w", err)
	}
	return shopID, environmentID, nil
}

func (r *Repository) CreateEnvironment(ctx context.Context, environment monitoring.CreateEnvironmentRecord) (int32, error) {
	return createEnvironment(ctx, r.queries, environment)
}

func createEnvironment(ctx context.Context, q *queries.Queries, environment monitoring.CreateEnvironmentRecord) (int32, error) {
	environmentID, err := q.CreateEnvironment(ctx, queries.CreateEnvironmentParams{
		OrganizationID:   environment.OrganizationID,
		ShopID:           environment.ShopID,
		Name:             environment.Name,
		Url:              environment.URL,
		ClientID:         environment.ClientID,
		ClientSecret:     environment.EncryptedClientSecret,
		ShopwareVersion:  environment.ShopwareVersion,
		EnvironmentToken: environment.EnvironmentToken,
	})
	if err != nil {
		return 0, fmt.Errorf("create environment: %w", err)
	}
	return environmentID, nil
}

func (r *Repository) UpdateShop(ctx context.Context, shop monitoring.Shop, defaultEnvironmentID *int32) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin update shop transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	txq := r.queries.WithTx(tx)

	if err := txq.UpdateShop(ctx, queries.UpdateShopParams{
		Name:           shop.Name,
		Description:    shop.Description,
		GitUrl:         shop.GitURL,
		ID:             shop.ID,
		OrganizationID: shop.OrganizationID,
	}); err != nil {
		return fmt.Errorf("update shop: %w", err)
	}
	if defaultEnvironmentID != nil {
		if err := txq.SetShopDefaultEnvironment(ctx, queries.SetShopDefaultEnvironmentParams{
			DefaultEnvironmentID: defaultEnvironmentID,
			ID:                   shop.ID,
			OrganizationID:       shop.OrganizationID,
		}); err != nil {
			return fmt.Errorf("set default environment: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit update shop transaction: %w", err)
	}
	return nil
}

func (r *Repository) CountShopEnvironments(ctx context.Context, shopID int32) (int32, error) {
	count, err := r.queries.CountEnvironmentsInShop(ctx, shopID)
	if err != nil {
		return 0, fmt.Errorf("count environments in shop: %w", err)
	}
	return count, nil
}

func (r *Repository) DeleteShop(ctx context.Context, shopID int32, organizationID string) error {
	if err := r.queries.DeleteShop(ctx, queries.DeleteShopParams{ID: shopID, OrganizationID: organizationID}); err != nil {
		return fmt.Errorf("delete shop: %w", err)
	}
	return nil
}

func (r *Repository) UpdateEnvironment(ctx context.Context, environment monitoring.Environment) error {
	ignores, err := json.Marshal(environment.Ignores)
	if err != nil {
		return fmt.Errorf("encode environment ignores: %w", err)
	}
	if err := r.queries.UpdateEnvironment(ctx, queries.UpdateEnvironmentParams{
		Name:         environment.Name,
		Url:          environment.URL,
		ClientID:     environment.ClientID,
		ClientSecret: environment.EncryptedClientSecret,
		Ignores:      ignores,
		ShopID:       environment.ShopID,
		ID:           environment.ID,
	}); err != nil {
		return fmt.Errorf("update environment: %w", err)
	}
	return nil
}

func (r *Repository) DeleteEnvironment(ctx context.Context, environmentID int32) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin delete environment transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	txq := r.queries.WithTx(tx)
	if err := txq.ReassignShopDefaultEnvironment(ctx, environmentID); err != nil {
		return fmt.Errorf("reassign shop default environment: %w", err)
	}
	if err := txq.DeleteEnvironment(ctx, environmentID); err != nil {
		return fmt.Errorf("delete environment: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete environment transaction: %w", err)
	}
	return nil
}

func (r *Repository) ListDeploymentIDs(ctx context.Context, environmentID int32) ([]int32, error) {
	rows, err := r.queries.ListDeployments(ctx, queries.ListDeploymentsParams{
		EnvironmentID: environmentID,
		Limit:         1000,
		Offset:        0,
	})
	if err != nil {
		return nil, fmt.Errorf("list deployments: %w", err)
	}
	ids := make([]int32, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids, nil
}

func (r *Repository) UpdateSitespeedSettings(ctx context.Context, environmentID int32, enabled bool, urls []string) error {
	urlsJSON, err := json.Marshal(urls)
	if err != nil {
		return fmt.Errorf("encode sitespeed urls: %w", err)
	}
	if err := r.queries.UpdateEnvironmentSitespeedSettings(ctx, queries.UpdateEnvironmentSitespeedSettingsParams{
		SitespeedEnabled: enabled,
		SitespeedUrls:    urlsJSON,
		ID:               environmentID,
	}); err != nil {
		return fmt.Errorf("update environment sitespeed settings: %w", err)
	}
	return nil
}

func (r *Repository) UpdateUserEnvironmentSubscriptions(ctx context.Context, userID string, subscriptions []string) error {
	payload, err := json.Marshal(subscriptions)
	if err != nil {
		return fmt.Errorf("encode environment subscriptions: %w", err)
	}
	if err := r.queries.UpdateUserNotifications(ctx, queries.UpdateUserNotificationsParams{
		Notifications: payload,
		ID:            userID,
	}); err != nil {
		return fmt.Errorf("update user environment subscriptions: %w", err)
	}
	return nil
}
