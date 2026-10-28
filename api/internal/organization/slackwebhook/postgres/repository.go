package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/organization/slackwebhook"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	queries *queries.Queries
}

func NewRepository(q *queries.Queries) *Repository {
	return &Repository{queries: q}
}

func (r *Repository) Find(ctx context.Context, target slackwebhook.Target) (string, error) {
	var (
		encrypted string
		err       error
	)
	if target.ShopID != 0 {
		encrypted, err = r.queries.GetShopSlackWebhook(ctx, queries.GetShopSlackWebhookParams{
			ShopID:         target.ShopID,
			OrganizationID: target.OrganizationID,
		})
	} else {
		encrypted, err = r.queries.GetOrganizationSlackWebhook(ctx, target.OrganizationID)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return "", slackwebhook.ErrNotConfigured
	}
	if err != nil {
		return "", fmt.Errorf("query slack webhook: %w", err)
	}
	return encrypted, nil
}

func (r *Repository) Upsert(ctx context.Context, target slackwebhook.Target, encryptedURL string) error {
	if target.ShopID != 0 {
		return r.queries.UpsertShopSlackWebhook(ctx, queries.UpsertShopSlackWebhookParams{
			ShopID:     target.ShopID,
			WebhookUrl: encryptedURL,
		})
	}
	return r.queries.UpsertOrganizationSlackWebhook(ctx, queries.UpsertOrganizationSlackWebhookParams{
		OrganizationID: target.OrganizationID,
		WebhookUrl:     encryptedURL,
	})
}

func (r *Repository) Delete(ctx context.Context, target slackwebhook.Target) error {
	if target.ShopID != 0 {
		return r.queries.DeleteShopSlackWebhook(ctx, target.ShopID)
	}
	return r.queries.DeleteOrganizationSlackWebhook(ctx, target.OrganizationID)
}

func (r *Repository) ShopBelongsToOrganization(ctx context.Context, organizationID string, shopID int32) (bool, error) {
	return r.queries.ShopBelongsToOrganization(ctx, queries.ShopBelongsToOrganizationParams{
		ID:             shopID,
		OrganizationID: organizationID,
	})
}
