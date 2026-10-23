package advisory_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	catalogadvisory "github.com/friendsofshopware/shopmon/api/internal/catalog/advisory"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	advisoryread "github.com/friendsofshopware/shopmon/api/internal/readmodel/advisory"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A Shopmon-authored advisory linked to a CVE must absorb the upstream sync
// instead of gaining a canonical-CVE twin: two rows would double-match shops
// and re-alert everyone who already heard about the Shopmon report.
func TestSyncMergesIntoLinkedShopmonAdvisory(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()

	env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Org One", "org-one", "user-1")
	shopID := env.SeedShop(t, "org-1", "Shop One")
	envID := int32(env.SeedEnvironment(t, "org-1", shopID, "prod", "https://example.com"))

	// The Shopmon report a shop was already matched and alerted against,
	// since linked to the CVE the sync is about to deliver.
	_, err := env.Queries.InsertManualAdvisory(ctx, queries.InsertManualAdvisoryParams{
		AdvisoryID:         "SHOPMON-2026-0001",
		Title:              "Shopmon early warning",
		Cve:                new("CVE-2026-0007"),
		Severity:           new("high"),
		Sources:            []byte(`[{"name":"Shopmon","remoteId":"SHOPMON-2026-0001"}]`),
		Cwes:               []byte("[]"),
		ExternalReferences: []byte("[]"),
		ReportedAt:         pgtype.Timestamp{Valid: true, Time: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)},
		DetailsSource:      new("shopmon"),
		Summary:            new("Shopmon summary"),
		IsVisible:          true,
		NotesPublic:        new("Shopmon guidance"),
		AffectedComponents: []string{},
		Tags:               []string{},
		Origin:             "shopmon",
		CreatedBy:          new("user-1"),
	})
	require.NoError(t, err)
	require.NoError(t, env.Queries.UpsertComposerAdvisoryPackage(ctx, queries.UpsertComposerAdvisoryPackageParams{
		AdvisoryID:       "SHOPMON-2026-0001",
		PackageName:      "shopware/core",
		AffectedVersions: ">=6.7.0.0,<6.7.10.1",
	}))
	// A mistaken manual package must disappear when official ranges arrive.
	require.NoError(t, env.Queries.UpsertComposerAdvisoryPackage(ctx, queries.UpsertComposerAdvisoryPackageParams{
		AdvisoryID: "SHOPMON-2026-0001", PackageName: "qa/incorrect-package", AffectedVersions: "<2.0.0",
	}))
	_, err = env.Pool.Exec(ctx, `
		INSERT INTO environment_sbom_component (environment_id, package_name, version)
		VALUES ($1, 'shopware/core', '6.7.8.0')
	`, envID)
	require.NoError(t, err)
	require.NoError(t, env.Queries.MarkEnvironmentAdvisoriesNotified(ctx, queries.MarkEnvironmentAdvisoriesNotifiedParams{
		EnvironmentID: envID,
		Column2:       []string{"SHOPMON-2026-0001"},
	}))

	server := advisorySyncServer(t)

	svc := catalogadvisory.NewServiceWithClient(env.Pool, env.Queries, server.URL, server.Client())
	require.NoError(t, svc.Sync(ctx))

	// No canonical-CVE twin.
	_, err = env.Queries.GetComposerAdvisory(ctx, "CVE-2026-0007")
	assert.ErrorIs(t, err, pgx.ErrNoRows)

	// The Shopmon row absorbed the upstream content and kept its marking and
	// admin enrichment.
	row, err := env.Queries.GetComposerAdvisory(ctx, "SHOPMON-2026-0001")
	require.NoError(t, err)
	assert.Equal(t, "Official SSRF title", row.Title)
	assert.Equal(t, "shopmon", row.Origin)
	require.NotNil(t, row.Cve)
	assert.Equal(t, "CVE-2026-0007", *row.Cve)
	require.NotNil(t, row.NotesPublic)
	assert.Equal(t, "Shopmon guidance", *row.NotesPublic)

	pkgs, err := env.Queries.ListComposerAdvisoryPackagesForAdvisory(ctx, "SHOPMON-2026-0001")
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	require.NotNil(t, pkgs[0].PackagistAdvisoryID)
	assert.Equal(t, "PKSA-core-7", *pkgs[0].PackagistAdvisoryID)

	// The environment still matches exactly once, under the Shopmon id, and
	// its alert marker survived: no second alert for the same vulnerability.
	matches, err := env.Queries.GetEnvironmentAdvisoryMatches(ctx, envID)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	assert.Equal(t, "SHOPMON-2026-0001", matches[0].AdvisoryID)

	notified, err := env.Queries.ListNotifiedAdvisoryIDsForEnvironment(ctx, envID)
	require.NoError(t, err)
	assert.Equal(t, []string{"SHOPMON-2026-0001"}, notified)
}

// A GitHub-imported advisory (GHSA id, CVE still pending at import time)
// must absorb the later Packagist sync of the same vulnerability under its
// GHSA id, exactly like a Shopmon-authored one: no canonical-CVE twin, no
// second alert.
func TestSyncMergesIntoGitHubImportedAdvisory(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()

	env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Org One", "org-one", "user-1")
	shopID := env.SeedShop(t, "org-1", "Shop One")
	envID := int32(env.SeedEnvironment(t, "org-1", shopID, "prod", "https://example.com"))

	_, err := env.Queries.InsertManualAdvisory(ctx, queries.InsertManualAdvisoryParams{
		AdvisoryID:         "GHSA-test-0007",
		Title:              "Imported GHSA title",
		GhsaID:             new("GHSA-test-0007"),
		Severity:           new("high"),
		Sources:            []byte(`[{"name":"GitHub","remoteId":"GHSA-test-0007"}]`),
		Cwes:               []byte("[]"),
		ExternalReferences: []byte("[]"),
		ReportedAt:         pgtype.Timestamp{Valid: true, Time: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)},
		DetailsSource:      new("github"),
		DetailsSyncedAt:    pgtype.Timestamp{Valid: true, Time: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)},
		GithubSyncedAt:     pgtype.Timestamp{Valid: true, Time: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)},
		Summary:            new("Imported summary"),
		IsVisible:          true,
		AffectedComponents: []string{},
		Tags:               []string{},
		Origin:             "github",
		CreatedBy:          new("user-1"),
	})
	require.NoError(t, err)
	require.NoError(t, env.Queries.UpsertComposerAdvisoryPackage(ctx, queries.UpsertComposerAdvisoryPackageParams{
		AdvisoryID:       "GHSA-test-0007",
		PackageName:      "shopware/core",
		AffectedVersions: ">=6.7.0.0,<6.7.10.1",
	}))
	_, err = env.Pool.Exec(ctx, `
		INSERT INTO environment_sbom_component (environment_id, package_name, version)
		VALUES ($1, 'shopware/core', '6.7.8.0')
	`, envID)
	require.NoError(t, err)
	require.NoError(t, env.Queries.MarkEnvironmentAdvisoriesNotified(ctx, queries.MarkEnvironmentAdvisoriesNotifiedParams{
		EnvironmentID: envID,
		Column2:       []string{"GHSA-test-0007"},
	}))

	server := advisorySyncServer(t)
	svc := catalogadvisory.NewServiceWithClient(env.Pool, env.Queries, server.URL, server.Client())
	require.NoError(t, svc.Sync(ctx))

	_, err = env.Queries.GetComposerAdvisory(ctx, "CVE-2026-0007")
	assert.ErrorIs(t, err, pgx.ErrNoRows)

	row, err := env.Queries.GetComposerAdvisory(ctx, "GHSA-test-0007")
	require.NoError(t, err)
	assert.Equal(t, "Official SSRF title", row.Title)
	assert.Equal(t, "github", row.Origin)
	require.NotNil(t, row.Cve)
	assert.Equal(t, "CVE-2026-0007", *row.Cve)

	matches, err := env.Queries.GetEnvironmentAdvisoryMatches(ctx, envID)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	assert.Equal(t, "GHSA-test-0007", matches[0].AdvisoryID)

	notified, err := env.Queries.ListNotifiedAdvisoryIDsForEnvironment(ctx, envID)
	require.NoError(t, err)
	assert.Equal(t, []string{"GHSA-test-0007"}, notified)
}

// advisorySyncServer serves one Packagist advisory (CVE-2026-0007 /
// GHSA-test-0007 on shopware/core) the way the sync consumes it.
func advisorySyncServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/packages.json", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"metadata-url":        "/p2/%package%.json",
			"security-advisories": map[string]any{"api-url": "/api/security-advisories/"},
		})
	})
	mux.HandleFunc("/packages/list.json", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"packageNames": []string{"shopware/core"},
		})
	})
	mux.HandleFunc("/api/security-advisories/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			require.NoError(t, r.ParseForm())
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"advisories": map[string]any{
				"shopware/core": []map[string]any{
					{
						"advisoryId": "PKSA-core-7", "packageName": "shopware/core",
						"title": "Official SSRF title", "cve": "CVE-2026-0007",
						"affectedVersions": ">=6.7.0.0,<6.7.10.1", "severity": "medium",
						"reportedAt": "2026-06-04 19:36:07",
						"sources":    []map[string]string{{"name": "GitHub", "remoteId": "GHSA-test-0007"}},
					},
				},
			},
		})
	})
	mux.HandleFunc("/rest/json/cves/2.0", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"vulnerabilities": []map[string]any{}})
	})
	mux.HandleFunc("/advisories/GHSA-test-0007", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{}`)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "{}")
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func TestSyncAndLinkSerializeCatalogIdentity(t *testing.T) {
	for _, first := range []string{"sync", "link"} {
		t.Run(first+" first", func(t *testing.T) {
			env := testutil.Setup(t)
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			env.SeedUser(t, "admin", "Admin", "admin@example.com", "admin")
			server := advisorySyncServer(t)
			syncService := catalogadvisory.NewServiceWithClient(env.Pool, env.Queries, server.URL, server.Client())
			require.NoError(t, syncService.Sync(ctx))
			manualService := advisoryread.NewService(env.Pool, env.Queries, nil)
			manual, err := manualService.Create(ctx, advisoryread.CreateInput{
				Title: "Early warning", CreatedBy: "admin",
				Packages: []advisoryread.PackageInput{{PackageName: "shopware/core", AffectedVersions: "<6.7.10.1"}},
			})
			require.NoError(t, err)
			gate, err := env.Pool.Begin(ctx)
			require.NoError(t, err)
			defer func() { _ = gate.Rollback(ctx) }()
			require.NoError(t, env.Queries.WithTx(gate).LockComposerAdvisoryCatalog(ctx))

			done := make(chan error, 2)
			start := func(operation string) {
				go func() {
					if operation == "sync" {
						done <- syncService.Sync(ctx)
						return
					}
					_, err := manualService.LinkOfficial(ctx, manual.Advisory.AdvisoryID, advisoryread.LinkOfficialInput{Cve: new("CVE-2026-0007")})
					done <- err
				}()
			}
			waitForBlocked := func(want int) {
				require.Eventually(t, func() bool {
					var count int
					err := env.Pool.QueryRow(ctx, `SELECT count(*) FROM pg_locks WHERE locktype = 'advisory' AND NOT granted AND database = (SELECT oid FROM pg_database WHERE datname = current_database())`).Scan(&count)
					return err == nil && count == want
				}, 5*time.Second, 10*time.Millisecond)
			}
			start(first)
			waitForBlocked(1)
			if first == "sync" {
				start("link")
			} else {
				start("sync")
			}
			waitForBlocked(2)
			require.NoError(t, gate.Commit(ctx))
			for range 2 {
				select {
				case err := <-done:
					require.NoError(t, err)
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				}
			}
			_, err = env.Queries.GetComposerAdvisory(ctx, "CVE-2026-0007")
			require.ErrorIs(t, err, pgx.ErrNoRows)
			row, err := env.Queries.GetComposerAdvisory(ctx, manual.Advisory.AdvisoryID)
			require.NoError(t, err)
			require.Equal(t, "CVE-2026-0007", *row.Cve)
			packages, err := env.Queries.ListComposerAdvisoryPackagesForAdvisory(ctx, row.AdvisoryID)
			require.NoError(t, err)
			require.Len(t, packages, 1)
			require.Equal(t, "PKSA-core-7", *packages[0].PackagistAdvisoryID)
		})
	}
}
