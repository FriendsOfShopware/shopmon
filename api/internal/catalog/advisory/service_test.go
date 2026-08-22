package advisory_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	catalogadvisory "github.com/friendsofshopware/shopmon/api/internal/catalog/advisory"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncCollapsesPackagesUnderCVE(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()

	mux := http.NewServeMux()
	// go-composer repository root (repo.packagist.org style).
	mux.HandleFunc("/packages.json", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"metadata-url": "/p2/%package%.json",
			"security-advisories": map[string]any{
				"api-url": "/api/security-advisories/",
			},
		})
	})
	mux.HandleFunc("/packages/list.json", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"packageNames": []string{"shopware/core", "shopware/platform"},
		})
	})
	// go-composer POSTs form packages[] to the security-advisories api-url.
	mux.HandleFunc("/api/security-advisories/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			require.NoError(t, r.ParseForm())
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"advisories": map[string]any{
				"shopware/core": []map[string]any{
					{
						"advisoryId": "PKSA-core-1", "packageName": "shopware/core",
						"title": "SSRF issue", "cve": "CVE-2026-0001",
						"affectedVersions": ">=6.7.0.0,<6.7.10.1", "severity": "medium",
						"reportedAt": "2026-06-04 19:36:07",
						"sources":    []map[string]string{{"name": "GitHub", "remoteId": "GHSA-test-0001"}},
					},
				},
				"shopware/platform": []map[string]any{
					{
						"advisoryId": "PKSA-platform-1", "packageName": "shopware/platform",
						"title": "SSRF issue", "cve": "CVE-2026-0001",
						"affectedVersions": ">=6.7.0.0,<6.7.10.1", "severity": "medium",
						"reportedAt": "2026-06-04 19:36:07",
						"sources":    []map[string]string{{"name": "GitHub", "remoteId": "GHSA-test-0001"}},
					},
				},
			},
		})
	})
	mux.HandleFunc("/rest/json/cves/2.0", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"vulnerabilities": []map[string]any{
				{
					"cve": map[string]any{
						"id":           "CVE-2026-0001",
						"descriptions": []map[string]string{{"lang": "en", "value": "NVD description of SSRF."}},
						"metrics": map[string]any{
							"cvssMetricV31": []map[string]any{
								{"cvssData": map[string]any{"baseScore": 5.5, "vectorString": "CVSS:3.1/AV:N"}},
							},
						},
						"weaknesses": []map[string]any{
							{"description": []map[string]string{{"lang": "en", "value": "CWE-918"}}},
						},
						"references": []map[string]string{},
					},
				},
			},
		})
	})
	// GitHub advisory details for the GHSA carried by the Packagist sources.
	mux.HandleFunc("/advisories/GHSA-test-0001", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"ghsa_id":"GHSA-test-0001","summary":"SSRF issue","vulnerabilities":[
          {"package":{"name":"shopware/core","ecosystem":"composer"},"vulnerable_version_range":">= 6.7.0.0, < 6.7.10.1","first_patched_version":"6.7.10.1"}]}`)
	})
	// Swallow unexpected paths with empty OK so go-composer is not noisy.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.NotFound(w, r)
			return
		}
		_, _ = io.WriteString(w, "{}")
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	svc := catalogadvisory.NewServiceWithClient(env.Pool, env.Queries, server.URL, server.Client())
	require.NoError(t, svc.Sync(ctx))

	row, err := env.Queries.GetComposerAdvisory(ctx, "CVE-2026-0001")
	require.NoError(t, err)
	assert.Equal(t, "SSRF issue", row.Title)
	require.NotNil(t, row.Cve)
	assert.Equal(t, "CVE-2026-0001", *row.Cve)
	require.NotNil(t, row.Description)
	assert.Contains(t, *row.Description, "NVD description")

	// NVD enriched the details first, but the GitHub pass must still run for
	// this GHSA: it is the only source of first_patched_versions.
	var patched map[string]string
	require.NoError(t, json.Unmarshal(row.FirstPatchedVersions, &patched))
	assert.Equal(t, "6.7.10.1", patched["6.7"],
		"CVE+GHSA advisory must receive patched versions from GitHub even after NVD enrichment")
	// NVD stays the details source; GitHub must not overwrite it.
	require.NotNil(t, row.DetailsSource)
	assert.Equal(t, "nvd", *row.DetailsSource)

	pkgs, err := env.Queries.ListComposerAdvisoryPackagesForAdvisory(ctx, "CVE-2026-0001")
	require.NoError(t, err)
	require.Len(t, pkgs, 2)
	names := map[string]string{}
	for _, p := range pkgs {
		names[p.PackageName] = p.PackagistAdvisoryID
	}
	assert.Equal(t, "PKSA-core-1", names["shopware/core"])
	assert.Equal(t, "PKSA-platform-1", names["shopware/platform"])

	notes := "shared note"
	_, err = env.Queries.UpdateComposerAdvisoryEnrichment(ctx, queries.UpdateComposerAdvisoryEnrichmentParams{
		AdvisoryID:         "CVE-2026-0001",
		IsVisible:          true,
		NotesPublic:        &notes,
		AffectedComponents: []string{},
		Tags:               []string{"ssrf"},
		EnrichedBy:         new("admin"),
	})
	require.NoError(t, err)
	require.NoError(t, svc.Sync(ctx))

	row, err = env.Queries.GetComposerAdvisory(ctx, "CVE-2026-0001")
	require.NoError(t, err)
	require.NotNil(t, row.NotesPublic)
	assert.Equal(t, notes, *row.NotesPublic)
	assert.Equal(t, []string{"ssrf"}, row.Tags)
}

// A Packagist outage must not hold back alerts for matches that scrapes have
// already stored. The scrape path records match rows but never notifies —
// notification is the rematch's job — so returning early on a sync failure
// would keep those advisories silent for the whole outage rather than the
// usual hour.
func TestSyncRematchesEvenWhenPackagistFails(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()

	env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Org One", "org-one", "user-1")
	shopID := env.SeedShop(t, "org-1", "Shop One")
	envID := int32(env.SeedEnvironment(t, "org-1", shopID, "prod", "https://example.com"))

	// A catalog entry and an inventory that already match: exactly the state a
	// scrape leaves behind between syncs.
	_, err := env.Pool.Exec(ctx, `
		INSERT INTO composer_advisory (advisory_id, title, severity, is_visible)
		VALUES ('CVE-OUTAGE-1', 'Stored advisory', 'high', true)
	`)
	require.NoError(t, err)
	_, err = env.Pool.Exec(ctx, `
		INSERT INTO composer_advisory_package (advisory_id, package_name, packagist_advisory_id, affected_versions)
		VALUES ('CVE-OUTAGE-1', 'symfony/http-kernel', 'PKSA-outage-1', '>=6.4.0,<6.4.3')
	`)
	require.NoError(t, err)
	_, err = env.Pool.Exec(ctx, `
		INSERT INTO environment_sbom_component (environment_id, package_name, version)
		VALUES ($1, 'symfony/http-kernel', '6.4.2')
	`, envID)
	require.NoError(t, err)

	// Every Packagist request fails, so syncFromPackagist returns an error.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "upstream unavailable", http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	svc := catalogadvisory.NewServiceWithClient(env.Pool, env.Queries, server.URL, server.Client())

	// Sync still reports the failure: the outage must stay visible.
	require.Error(t, svc.Sync(ctx))

	// ...but the stored catalog was rematched, so the advisory is on record for
	// this environment and can be alerted on.
	matches, err := env.Queries.GetEnvironmentAdvisoryMatches(ctx, envID)
	require.NoError(t, err)
	require.Len(t, matches, 1, "a Packagist outage must not skip the rematch")
	assert.Equal(t, "CVE-OUTAGE-1", matches[0].AdvisoryID)
}
