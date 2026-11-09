package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/friendsofshopware/shopmon/api/internal/notification"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

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

func (r *Repository) Save(ctx context.Context, target destination.Target, row destination.Record, create bool) error {
	var count int64
	var err error
	if create {
		count, err = r.queries.CreateNotificationDestination(ctx, queries.CreateNotificationDestinationParams{ID: row.ID, OrganizationID: target.OrganizationID, ShopID: target.ShopID, Name: row.Name, Dsn: row.DSN})
	} else {
		count, err = r.queries.UpdateNotificationDestination(ctx, queries.UpdateNotificationDestinationParams{ID: row.ID, OrganizationID: target.OrganizationID, ShopID: target.ShopID, Name: row.Name, Dsn: row.DSN})
	}
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

func (r *Repository) ExistsInOrganization(ctx context.Context, orgID, id string) (bool, error) {
	_, err := r.queries.GetNotificationDestinationInOrganization(ctx, queries.GetNotificationDestinationInOrganizationParams{ID: id, OrganizationID: orgID})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}
func (r *Repository) EnvironmentMatches(ctx context.Context, id string, environmentID int32) (bool, error) {
	return r.queries.EnvironmentMatchesNotificationDestination(ctx, queries.EnvironmentMatchesNotificationDestinationParams{ID: id, ID_2: environmentID})
}
func (r *Repository) ListPreferences(ctx context.Context, id string) ([]notification.Preference, error) {
	rows, err := r.queries.ListDestinationNotificationPreferences(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list destination preferences: %w", err)
	}
	result := make([]notification.Preference, 0, len(rows))
	for _, p := range rows {
		result = append(result, notification.Preference{ScopeType: p.ScopeType, ScopeID: p.ScopeID, EventType: p.EventType, Channel: p.Channel, Enabled: p.Enabled})
	}
	return result, nil
}
func (r *Repository) SetPreference(ctx context.Context, id string, p notification.Preference) error {
	err := r.queries.UpsertDestinationNotificationPreference(ctx, queries.UpsertDestinationNotificationPreferenceParams{DestinationID: id, ScopeType: p.ScopeType, ScopeID: p.ScopeID, EventType: p.EventType, Enabled: p.Enabled})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		if pgErr.ConstraintName == "destination_notification_preference_environment_id_fkey" {
			return destination.ErrShopNotFound
		}
		return destination.ErrNotConfigured
	}
	return err
}
func (r *Repository) DeletePreference(ctx context.Context, id string, p notification.Preference) error {
	return r.queries.DeleteDestinationNotificationPreference(ctx, queries.DeleteDestinationNotificationPreferenceParams{DestinationID: id, ScopeType: p.ScopeType, ScopeID: p.ScopeID, EventType: p.EventType})
}
