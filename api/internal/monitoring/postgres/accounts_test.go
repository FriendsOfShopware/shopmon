package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/monitoring"
	"github.com/friendsofshopware/shopmon/api/internal/testutil/testdb"
)

func setupAccountRepo(t *testing.T) (*Repository, *pgxpool.Pool, int32) {
	t.Helper()
	pool := testdb.Setup(t)
	q := queries.New(pool)
	ctx := context.Background()

	_, err := pool.Exec(ctx,
		`INSERT INTO organization (id, name, slug, created_at) VALUES ('org-1','O','o',NOW())`)
	require.NoError(t, err)
	// acknowledged_by references "user", so the acting account has to exist.
	_, err = pool.Exec(ctx,
		`INSERT INTO "user" (id, name, email, email_verified, created_at, updated_at)
		 VALUES ('user-1','U','u@example.com',true,NOW(),NOW())`)
	require.NoError(t, err)
	var shopID int32
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO shop (organization_id,name,created_at,updated_at) VALUES ('org-1','S',NOW(),NOW()) RETURNING id`).Scan(&shopID))
	var envID int32
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO environment (organization_id, shop_id, name, url, client_id, client_secret,
		                         shopware_version, environment_token, created_at)
		VALUES ('org-1', $1, 'E', 'https://e', 'c', 's', '6.6', 't', NOW()) RETURNING id`, shopID).Scan(&envID))
	require.NoError(t, err)

	// A seeded baseline, as the first scrape would leave it.
	_, err = pool.Exec(ctx,
		`INSERT INTO environment_account_state (environment_id, supported, baseline_at, account_count)
		 VALUES ($1, true, NOW(), 3)`, envID)
	require.NoError(t, err)

	return NewRepository(pool, q), pool, envID
}

func baselineSet(t *testing.T, pool *pgxpool.Pool, envID int32) bool {
	t.Helper()
	var set bool
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT baseline_at IS NOT NULL FROM environment_account_state WHERE environment_id = $1`, envID).Scan(&set))
	return set
}

// Saving the settings form unchanged must not reset the baseline. It would fold
// an administrator added since the last scrape into the new baseline, and that
// account would then never be reported.
func TestSetAccountMonitoringKeepsBaselineWhenAlreadyEnabled(t *testing.T) {
	repo, pool, envID := setupAccountRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.SetAccountMonitoring(ctx, envID, true, "user-1"))

	assert.True(t, baselineSet(t, pool, envID),
		"an unchanged save must leave the baseline alone")
}

// Re-enabling after a pause does reset it, so the next scrape seeds instead of
// reporting everything that changed while monitoring was off.
func TestSetAccountMonitoringResetsBaselineOnReEnable(t *testing.T) {
	repo, pool, envID := setupAccountRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.SetAccountMonitoring(ctx, envID, false, "user-1"))
	assert.True(t, baselineSet(t, pool, envID), "disabling leaves the baseline in place")

	require.NoError(t, repo.SetAccountMonitoring(ctx, envID, true, "user-1"))
	assert.False(t, baselineSet(t, pool, envID), "re-enabling seeds again")
}

// Disabling acknowledges the open findings: the scrape stops emitting their
// checks, and leaving them open would strand them with no tab to clear from.
func TestSetAccountMonitoringAcknowledgesOpenFindingsOnDisable(t *testing.T) {
	repo, pool, envID := setupAccountRepo(t)
	ctx := context.Background()

	_, err := pool.Exec(ctx, `
		INSERT INTO environment_account_event (environment_id, account_type, account_id, label, change, level)
		VALUES ($1,'user','aaaa','a***@e.de','new_admin_user','red'),
		       ($1,'user','bbbb','b***@e.de','admin_revoked','info')`, envID)
	require.NoError(t, err)

	require.NoError(t, repo.SetAccountMonitoring(ctx, envID, false, "user-1"))

	var open int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM environment_account_event
		 WHERE environment_id = $1 AND acknowledged_at IS NULL`, envID).Scan(&open))
	assert.Equal(t, 0, open)

	var by *string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT acknowledged_by FROM environment_account_event
		 WHERE environment_id = $1 AND change = 'new_admin_user'`, envID).Scan(&by))
	require.NotNil(t, by)
	assert.Equal(t, "user-1", *by, "who switched it off is recorded")
}

// Confirming has to name the state the reviewer saw. A scrape refreshing the
// finding in between must make the acknowledgement fail rather than close a
// finding describing something nobody read.
func TestAcknowledgeAccountEventGuardsOnRevision(t *testing.T) {
	repo, pool, envID := setupAccountRepo(t)
	ctx := context.Background()

	var id int64
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO environment_account_event (environment_id, account_type, account_id, label, change, level)
		VALUES ($1,'acl_role','r1','Editor','role_privileges_grown','yellow') RETURNING id`, envID).Scan(&id))

	// A scrape refreshes it with a worse state after the page was rendered.
	_, err := pool.Exec(ctx,
		`UPDATE environment_account_event SET revision = 2 WHERE id = $1`, id)
	require.NoError(t, err)

	err = repo.AcknowledgeAccountEvent(ctx, envID, id, 1, "user-1", nil)
	assert.ErrorIs(t, err, monitoring.ErrAccountEventChanged)

	var open bool
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT acknowledged_at IS NULL FROM environment_account_event WHERE id = $1`, id).Scan(&open))
	assert.True(t, open, "the finding stays open until the newer state is reviewed")

	// Confirming the state that is actually there succeeds.
	require.NoError(t, repo.AcknowledgeAccountEvent(ctx, envID, id, 2, "user-1", nil))

	// And a second submit is a conflict, not a silent success.
	assert.ErrorIs(t, repo.AcknowledgeAccountEvent(ctx, envID, id, 2, "user-1", nil),
		monitoring.ErrAccountEventAlreadyAcknowledged)

	// An unknown event is a plain not-found.
	assert.ErrorIs(t, repo.AcknowledgeAccountEvent(ctx, envID, 999999, 1, "user-1", nil),
		monitoring.ErrAccountEventNotFound)
}
