package advisory

import (
	"context"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/sbom"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The rematch commits its match rows before notifying, because those rows
// drive the UI and shop status whether or not the alert succeeds. Dedup must
// therefore key on "already alerted", not "a match row exists" — otherwise a
// transient failure in the notify path silences the advisory forever.
func TestRematchRetriesUnnotifiedAdvisories(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()

	env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Org One", "org-one", "user-1")
	shopID := env.SeedShop(t, "org-1", "Shop One")
	envID := int32(env.SeedEnvironment(t, "org-1", shopID, "prod", "https://example.com"))

	_, err := env.Pool.Exec(ctx, `
		INSERT INTO composer_advisory (advisory_id, title, severity, is_visible)
		VALUES ('CVE-NOTIFY-1', 'Vuln one', 'high', true)
	`)
	require.NoError(t, err)

	_, err = env.Pool.Exec(ctx, `
		INSERT INTO environment_sbom_component (environment_id, package_name, version)
		VALUES ($1, 'symfony/http-kernel', '6.4.2')
	`, envID)
	require.NoError(t, err)

	// notifier is nil, standing in for a notify path that does not complete:
	// the match rows must still be written, and nothing may be marked.
	svc := &Service{pool: env.Pool, queries: env.Queries}
	index := sbom.NewIndex([]sbom.AdvisoryPackage{{
		AdvisoryID:       "CVE-NOTIFY-1",
		PackageName:      "symfony/http-kernel",
		AffectedVersions: ">=6.4.0,<6.4.3",
	}})

	count, err := svc.rematchEnvironment(ctx, envID, index)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	matches, err := env.Queries.GetEnvironmentAdvisoryMatches(ctx, envID)
	require.NoError(t, err)
	require.Len(t, matches, 1, "match rows must persist regardless of the alert")

	notified, err := env.Queries.ListNotifiedAdvisoryIDsForEnvironment(ctx, envID)
	require.NoError(t, err)
	assert.Empty(t, notified, "an advisory that was never dispatched must stay unmarked")

	// Second pass: the advisory is still un-notified, so it is offered again
	// rather than being treated as already seen.
	_, err = svc.rematchEnvironment(ctx, envID, index)
	require.NoError(t, err)

	notified, err = env.Queries.ListNotifiedAdvisoryIDsForEnvironment(ctx, envID)
	require.NoError(t, err)
	assert.Empty(t, notified)

	// Once recorded, the advisory is not re-offered.
	require.NoError(t, env.Queries.MarkEnvironmentAdvisoriesNotified(ctx, queries.MarkEnvironmentAdvisoriesNotifiedParams{
		EnvironmentID: envID,
		Column2:       []string{"CVE-NOTIFY-1"},
	}))
	notified, err = env.Queries.ListNotifiedAdvisoryIDsForEnvironment(ctx, envID)
	require.NoError(t, err)
	assert.Equal(t, []string{"CVE-NOTIFY-1"}, notified)
}

// A marker outliving its match would silence the advisory if it ever returned
// (package downgraded, affected range widened), so stale markers are pruned.
func TestRematchPrunesStaleNotificationMarkers(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()

	env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Org One", "org-one", "user-1")
	shopID := env.SeedShop(t, "org-1", "Shop One")
	envID := int32(env.SeedEnvironment(t, "org-1", shopID, "prod", "https://example.com"))

	_, err := env.Pool.Exec(ctx, `
		INSERT INTO composer_advisory (advisory_id, title, severity, is_visible)
		VALUES ('CVE-NOTIFY-2', 'Vuln two', 'high', true)
	`)
	require.NoError(t, err)

	// The shop has upgraded out of the affected range: a component that no
	// longer matches, plus a leftover marker from when it did.
	_, err = env.Pool.Exec(ctx, `
		INSERT INTO environment_sbom_component (environment_id, package_name, version)
		VALUES ($1, 'symfony/http-kernel', '6.4.9')
	`, envID)
	require.NoError(t, err)
	require.NoError(t, env.Queries.MarkEnvironmentAdvisoriesNotified(ctx, queries.MarkEnvironmentAdvisoriesNotifiedParams{
		EnvironmentID: envID,
		Column2:       []string{"CVE-NOTIFY-2"},
	}))

	svc := &Service{pool: env.Pool, queries: env.Queries}
	index := sbom.NewIndex([]sbom.AdvisoryPackage{{
		AdvisoryID:       "CVE-NOTIFY-2",
		PackageName:      "symfony/http-kernel",
		AffectedVersions: ">=6.4.0,<6.4.3",
	}})

	_, err = svc.rematchEnvironment(ctx, envID, index)
	require.NoError(t, err)

	notified, err := env.Queries.ListNotifiedAdvisoryIDsForEnvironment(ctx, envID)
	require.NoError(t, err)
	assert.Empty(t, notified, "marker must not outlive the match it belongs to")
}

func TestRematchMatchesEnvironmentRuntimeVersions(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()

	env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Org One", "org-one", "user-1")
	shopID := env.SeedShop(t, "org-1", "Shop One")
	envID := int32(env.SeedEnvironment(t, "org-1", shopID, "prod", "https://example.com"))

	_, err := env.Pool.Exec(ctx, `
		UPDATE environment SET php_version = '8.3.14', database_version = '8.0.36' WHERE id = $1
	`, envID)
	require.NoError(t, err)

	_, err = env.Pool.Exec(ctx, `
		INSERT INTO composer_advisory (advisory_id, title, severity, is_visible, origin)
		VALUES ('SHOPMON-2026-0001', 'PHP PDO parser vulnerability', 'high', true, 'shopmon'),
		       ('SHOPMON-2026-0002', 'MySQL server vulnerability', 'medium', true, 'shopmon')
	`)
	require.NoError(t, err)

	_, err = env.Pool.Exec(ctx, `
		INSERT INTO composer_advisory_package (advisory_id, package_name, affected_versions)
		VALUES ('SHOPMON-2026-0001', 'php', '<8.4.0'),
		       ('SHOPMON-2026-0002', 'mysql', '>=8.0.0,<8.0.40')
	`)
	require.NoError(t, err)

	svc := &Service{pool: env.Pool, queries: env.Queries}
	matched, err := svc.rematchEnvironments(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, matched)

	matches, err := env.Queries.GetEnvironmentAdvisoryMatches(ctx, envID)
	require.NoError(t, err)
	require.Len(t, matches, 2)
	assert.Equal(t, "SHOPMON-2026-0001", matches[0].AdvisoryID)
	assert.Equal(t, "php", matches[0].PackageName)
	assert.Equal(t, "8.3.14", matches[0].InstalledVersion)
	assert.Equal(t, "SHOPMON-2026-0002", matches[1].AdvisoryID)
	assert.Equal(t, "mysql", matches[1].PackageName)
	assert.Equal(t, "8.0.36", matches[1].InstalledVersion)
}
