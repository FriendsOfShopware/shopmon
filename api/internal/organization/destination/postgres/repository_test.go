package postgres_test

import (
	"fmt"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/organization/destination"
	"github.com/friendsofshopware/shopmon/api/internal/organization/destination/postgres"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenamePreservesRotatedConnectionAndCannotRecreateDeletedDestination(t *testing.T) {
	env := testutil.Setup(t)
	env.SeedUser(t, "owner", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org", "Org", "org", "owner")
	repo := postgres.NewRepository(env.Queries)
	target := destination.Target{OrganizationID: "org"}
	require.NoError(t, repo.Save(t.Context(), target, destination.Record{ID: "d", Name: "Original", DSN: "old"}, true))
	// One editor loaded the original name while another rotated credentials.
	require.NoError(t, repo.Save(t.Context(), target, destination.Record{ID: "d", Name: "Original", DSN: "rotated"}, false))
	require.NoError(t, repo.Save(t.Context(), target, destination.Record{ID: "d", Name: "Renamed"}, false))
	rows, err := repo.List(t.Context(), target)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "rotated", rows[0].DSN)
	require.NoError(t, repo.Delete(t.Context(), target, "d"))
	require.ErrorIs(t, repo.Save(t.Context(), target, destination.Record{ID: "d", Name: "Stale", DSN: "old"}, false), destination.ErrNotConfigured)
}

func TestEnvironmentDeletionRejectsLatePreferenceWrites(t *testing.T) {
	env := testutil.Setup(t)
	env.SeedUser(t, "owner", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org", "Org", "org", "owner")
	shop := env.SeedShop(t, "org", "Shop")
	environment := env.SeedEnvironment(t, "org", shop, "Production", "https://example.com")
	_, err := env.Queries.CreateNotificationDestination(t.Context(), queries.CreateNotificationDestinationParams{ID: "d", OrganizationID: "org", Name: "Shared", Dsn: "encrypted"})
	require.NoError(t, err)
	pref := queries.UpsertDestinationNotificationPreferenceParams{DestinationID: "d", ScopeType: "environment", ScopeID: fmt.Sprint(environment), Enabled: false}
	require.NoError(t, env.Queries.UpsertDestinationNotificationPreference(t.Context(), pref))
	tx, err := env.Pool.Begin(t.Context())
	require.NoError(t, err)
	defer func() { _ = tx.Rollback(t.Context()) }()
	_, err = tx.Exec(t.Context(), "UPDATE shop SET default_environment_id = NULL WHERE id = $1", shop)
	require.NoError(t, err)
	_, err = tx.Exec(t.Context(), "DELETE FROM environment WHERE id = $1", environment)
	require.NoError(t, err)
	done := make(chan error, 1)
	go func() { done <- env.Queries.UpsertDestinationNotificationPreference(t.Context(), pref) }()
	require.NoError(t, tx.Commit(t.Context()))
	require.Error(t, <-done)
	rows, err := env.Queries.ListDestinationNotificationPreferences(t.Context(), "d")
	require.NoError(t, err)
	assert.Empty(t, rows)
}
