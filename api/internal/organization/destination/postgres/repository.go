package postgres

import (
	"context"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/organization/destination"
)

type Repository struct{ queries *queries.Queries }

func NewRepository(q *queries.Queries) *Repository { return &Repository{queries: q} }

func (r *Repository) List(ctx context.Context, target destination.Target) ([]destination.Record, error) {
	rows, err := r.queries.ListNotificationDestinations(ctx, queries.ListNotificationDestinationsParams{OrganizationID: target.OrganizationID, ShopID: target.ShopID})
	if err != nil {
		return nil, fmt.Errorf("query destinations: %w", err)
	}
	result := make([]destination.Record, 0, len(rows))
	for _, row := range rows {
		result = append(result, destination.Record{ID: row.ID, Name: row.Name, DSN: row.Dsn})
	}
	return result, nil
}

func (r *Repository) Save(ctx context.Context, target destination.Target, row destination.Record) error {
	count, err := r.queries.SaveNotificationDestination(ctx, queries.SaveNotificationDestinationParams{ID: row.ID, OrganizationID: target.OrganizationID, ShopID: target.ShopID, Name: row.Name, Dsn: row.DSN})
	if err != nil {
		return fmt.Errorf("save destination: %w", err)
	}
	if count == 0 {
		return destination.ErrNotConfigured
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, target destination.Target, id string) error {
	count, err := r.queries.DeleteNotificationDestination(ctx, queries.DeleteNotificationDestinationParams{ID: id, OrganizationID: target.OrganizationID, ShopID: target.ShopID})
	if err != nil {
		return fmt.Errorf("delete destination: %w", err)
	}
	if count == 0 {
		return destination.ErrNotConfigured
	}
	return nil
}

func (r *Repository) ShopBelongsToOrganization(ctx context.Context, orgID string, shopID int32) (bool, error) {
	return r.queries.ShopBelongsToOrganization(ctx, queries.ShopBelongsToOrganizationParams{ID: shopID, OrganizationID: orgID})
}
