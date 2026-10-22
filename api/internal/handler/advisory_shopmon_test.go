package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func shopmonAdvRequest(t *testing.T, method, url, token string, payload any) (int, map[string]any) {
	t.Helper()
	var body *bytes.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		require.NoError(t, err)
		body = bytes.NewReader(encoded)
	} else {
		body = bytes.NewReader(nil)
	}
	req := testutil.NewRequest(t, method, url, body)
	req.Header.Set("Authorization", "Bearer "+token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	decoded := map[string]any{}
	if resp.ContentLength != 0 {
		_ = json.NewDecoder(resp.Body).Decode(&decoded)
	}
	return resp.StatusCode, decoded
}

func shopmonAdvCreatePayload(visible bool) map[string]any {
	return map[string]any{
		"title":   "Remote code execution in example package",
		"summary": "Shopmon summary",
		"packages": []map[string]any{
			{"packageName": "shopware/core", "affectedVersions": ">=6.7.0.0,<6.7.10.1"},
		},
		"severity":  "high",
		"isVisible": visible,
	}
}

func TestAdminCreateShopmonAdvisoryDraftThenPublish(t *testing.T) {
	env := testutil.Setup(t)
	adminToken := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	userToken := env.SeedUser(t, "user-1", "Regular User", "user@example.com", "user")

	status, created := shopmonAdvRequest(t, "POST", env.Server.URL+"/api/admin/advisories", adminToken, shopmonAdvCreatePayload(false))
	require.Equal(t, http.StatusCreated, status, "%v", created)
	advisoryID, _ := created["advisoryId"].(string)
	assert.True(t, strings.HasPrefix(advisoryID, "SHOPMON-"), "id = %q", advisoryID)
	assert.Equal(t, "shopmon", created["origin"])
	assert.Equal(t, false, created["isVisible"])
	assert.Equal(t, "admin-1", created["createdBy"])

	// Drafts stay invisible to users.
	status, _ = shopmonAdvRequest(t, "GET", env.Server.URL+"/api/advisories/"+advisoryID, userToken, nil)
	assert.Equal(t, http.StatusNotFound, status)

	// Non-admins cannot create advisories.
	status, _ = shopmonAdvRequest(t, "POST", env.Server.URL+"/api/admin/advisories", userToken, shopmonAdvCreatePayload(true))
	assert.Equal(t, http.StatusForbidden, status)

	// Publishing surfaces it to users, marked with its Shopmon origin.
	status, updated := shopmonAdvRequest(t, "PATCH", env.Server.URL+"/api/admin/advisories/"+advisoryID, adminToken, map[string]any{"isVisible": true})
	require.Equal(t, http.StatusOK, status, "%v", updated)

	status, fetched := shopmonAdvRequest(t, "GET", env.Server.URL+"/api/advisories/"+advisoryID, userToken, nil)
	require.Equal(t, http.StatusOK, status, "%v", fetched)
	assert.Equal(t, "shopmon", fetched["origin"])
	assert.Equal(t, "Remote code execution in example package", fetched["title"])
	_, hasCreatedBy := fetched["createdBy"]
	assert.False(t, hasCreatedBy, "createdBy is admin-only")

	// The admin list can be filtered down to Shopmon-authored advisories.
	status, list := shopmonAdvRequest(t, "GET", env.Server.URL+"/api/admin/advisories?origin=shopmon", adminToken, nil)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, float64(1), list["total"])
	status, list = shopmonAdvRequest(t, "GET", env.Server.URL+"/api/admin/advisories?origin=upstream", adminToken, nil)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, float64(0), list["total"])
}

func TestAdminCreateShopmonAdvisoryValidation(t *testing.T) {
	env := testutil.Setup(t)
	adminToken := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	seedAdvisory(t, env.Queries, "CVE-2026-1111", []string{"shopware/core"}, true, nil)

	cases := []struct {
		name    string
		mutate  func(map[string]any)
		want    int
		message string
	}{
		{
			name: "unparseable constraint",
			mutate: func(p map[string]any) {
				p["packages"] = []map[string]any{{"packageName": "shopware/core", "affectedVersions": "not-a-constraint"}}
			},
			want:    http.StatusBadRequest,
			message: "affected versions",
		},
		{
			name: "wildcard constraint would flag every shop",
			mutate: func(p map[string]any) {
				p["packages"] = []map[string]any{{"packageName": "shopware/core", "affectedVersions": "*"}}
			},
			want:    http.StatusBadRequest,
			message: "affected versions",
		},
		{
			name:    "no packages",
			mutate:  func(p map[string]any) { p["packages"] = []map[string]any{} },
			want:    http.StatusBadRequest,
			message: "at least one affected package",
		},
		{
			name: "invalid package name",
			mutate: func(p map[string]any) {
				p["packages"] = []map[string]any{{"packageName": "not a package", "affectedVersions": "<1.0.0"}}
			},
			want:    http.StatusBadRequest,
			message: "package name",
		},
		{
			name:    "invalid CVE format",
			mutate:  func(p map[string]any) { p["cve"] = "CVE-XYZ" },
			want:    http.StatusBadRequest,
			message: "CVE",
		},
		{
			name:    "CVE already tracked",
			mutate:  func(p map[string]any) { p["cve"] = "CVE-2026-1111" },
			want:    http.StatusConflict,
			message: "already tracked",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := shopmonAdvCreatePayload(false)
			tc.mutate(payload)
			status, body := shopmonAdvRequest(t, "POST", env.Server.URL+"/api/admin/advisories", adminToken, payload)
			assert.Equal(t, tc.want, status, "%v", body)
			assert.Contains(t, body["message"], tc.message)
		})
	}
}

func TestAdminUpdateShopmonAdvisoryCoreFieldsGuards(t *testing.T) {
	env := testutil.Setup(t)
	adminToken := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	seedAdvisory(t, env.Queries, "CVE-2026-2222", []string{"shopware/core"}, true, nil)

	// Core fields on an upstream advisory belong to the sync, not to admins.
	status, body := shopmonAdvRequest(t, "PATCH", env.Server.URL+"/api/admin/advisories/CVE-2026-2222", adminToken, map[string]any{"title": "Rewritten"})
	assert.Equal(t, http.StatusBadRequest, status, "%v", body)

	// On an unlinked Shopmon advisory they edit freely.
	status, created := shopmonAdvRequest(t, "POST", env.Server.URL+"/api/admin/advisories", adminToken, shopmonAdvCreatePayload(true))
	require.Equal(t, http.StatusCreated, status, "%v", created)
	advisoryID := created["advisoryId"].(string)

	status, updated := shopmonAdvRequest(t, "PATCH", env.Server.URL+"/api/admin/advisories/"+advisoryID, adminToken, map[string]any{
		"title":    "Corrected title",
		"severity": "critical",
		"packages": []map[string]any{{"packageName": "shopware/core", "affectedVersions": "<6.7.10.2"}},
	})
	require.Equal(t, http.StatusOK, status, "%v", updated)
	assert.Equal(t, "Corrected title", updated["title"])
	pkgs, _ := updated["packages"].([]any)
	require.Len(t, pkgs, 1)
	assert.Equal(t, "<6.7.10.2", pkgs[0].(map[string]any)["affectedVersions"])

	// An explicit empty severity clears it (the admin form's "unknown" choice).
	status, updated = shopmonAdvRequest(t, "PATCH", env.Server.URL+"/api/admin/advisories/"+advisoryID, adminToken, map[string]any{"severity": ""})
	require.Equal(t, http.StatusOK, status, "%v", updated)
	assert.Nil(t, updated["severity"])

	// After linking, the sync owns the content again.
	status, linked := shopmonAdvRequest(t, "POST", env.Server.URL+"/api/admin/advisories/"+advisoryID+"/link-official", adminToken, map[string]any{"cve": "CVE-2026-3333"})
	require.Equal(t, http.StatusOK, status, "%v", linked)
	assert.Equal(t, "CVE-2026-3333", linked["cve"])
	assert.Equal(t, advisoryID, linked["advisoryId"], "the Shopmon id is stable across linking")

	status, body = shopmonAdvRequest(t, "PATCH", env.Server.URL+"/api/admin/advisories/"+advisoryID, adminToken, map[string]any{"title": "Too late"})
	assert.Equal(t, http.StatusConflict, status, "%v", body)

	// Enrichment stays editable after linking.
	status, updated = shopmonAdvRequest(t, "PATCH", env.Server.URL+"/api/admin/advisories/"+advisoryID, adminToken, map[string]any{"notesPublic": "Upgrade to 6.7.10.2"})
	require.Equal(t, http.StatusOK, status, "%v", updated)
	assert.Equal(t, "Upgrade to 6.7.10.2", updated["notesPublic"])
}

func TestAdminUpdateShopmonAdvisoryCVSSScore(t *testing.T) {
	env := testutil.Setup(t)
	adminToken := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	payload := shopmonAdvCreatePayload(false)
	payload["cvssScore"] = 7.5
	status, created := shopmonAdvRequest(t, "POST", env.Server.URL+"/api/admin/advisories", adminToken, payload)
	require.Equal(t, http.StatusCreated, status, "%v", created)
	url := env.Server.URL + "/api/admin/advisories/" + created["advisoryId"].(string)

	for _, tc := range []struct {
		name   string
		patch  map[string]any
		want   any
		status int
	}{
		{"omitted preserves score", map[string]any{"summary": "Edited summary"}, 7.5, http.StatusOK},
		{"null clears score", map[string]any{"cvssScore": nil}, nil, http.StatusOK},
		{"zero is a score", map[string]any{"cvssScore": 0}, float64(0), http.StatusOK},
		{"number updates score", map[string]any{"cvssScore": 8.1}, 8.1, http.StatusOK},
		{"out of range is rejected", map[string]any{"cvssScore": 11}, 8.1, http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, updated := shopmonAdvRequest(t, "PATCH", url, adminToken, tc.patch)
			require.Equal(t, tc.status, status, "%v", updated)
			status, fetched := shopmonAdvRequest(t, "GET", url, adminToken, nil)
			require.Equal(t, http.StatusOK, status, "%v", fetched)
			assert.Equal(t, tc.want, fetched["cvssScore"])
		})
	}

	status, linked := shopmonAdvRequest(t, "POST", url+"/link-official", adminToken, map[string]any{"cve": "CVE-2026-4444"})
	require.Equal(t, http.StatusOK, status, "%v", linked)
	status, body := shopmonAdvRequest(t, "PATCH", url, adminToken, map[string]any{"cvssScore": nil})
	assert.Equal(t, http.StatusConflict, status, "%v", body)
}

func TestAdminLinkShopmonAdvisoryMergesOfficialRow(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()
	adminToken := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	userToken := env.SeedUser(t, "user-1", "Regular User", "user@example.com", "user")
	env.SeedOrganization(t, "org-1", "Org One", "org-one", "user-1")
	shopID := env.SeedShop(t, "org-1", "Shop One")
	envID := env.SeedEnvironment(t, "org-1", shopID, "Production", "https://shop.example.com")

	status, created := shopmonAdvRequest(t, "POST", env.Server.URL+"/api/admin/advisories", adminToken, shopmonAdvCreatePayload(true))
	require.Equal(t, http.StatusCreated, status, "%v", created)
	shopmonID := created["advisoryId"].(string)

	// The official advisory syncs in as its own canonical-CVE row first.
	seedAdvisory(t, env.Queries, "CVE-2026-4444", []string{"shopware/core"}, true, nil)

	// Both rows matched the shop and the owner suppressed the Shopmon one;
	// the merge must keep exactly one match and the suppression.
	seedAdvisoryMatch(t, env.Queries, int32(envID), shopmonID, "shopware/core", "6.7.8.2")
	seedAdvisoryMatch(t, env.Queries, int32(envID), "CVE-2026-4444", "shopware/core", "6.7.8.2")
	require.NoError(t, env.Queries.MarkEnvironmentAdvisoriesNotified(ctx, queries.MarkEnvironmentAdvisoriesNotifiedParams{
		EnvironmentID: int32(envID),
		Column2:       []string{shopmonID},
	}))
	_, err := env.Queries.CreateAdvisorySuppression(ctx, queries.CreateAdvisorySuppressionParams{
		OrganizationID: "org-1",
		ShopID:         int32(shopID),
		AdvisoryID:     shopmonID,
		Reason:         "mitigated by WAF",
		CreatedBy:      new("user-1"),
	})
	require.NoError(t, err)

	status, linked := shopmonAdvRequest(t, "POST", env.Server.URL+"/api/admin/advisories/"+shopmonID+"/link-official", adminToken, map[string]any{"cve": "CVE-2026-4444"})
	require.Equal(t, http.StatusOK, status, "%v", linked)
	assert.Equal(t, shopmonID, linked["advisoryId"])
	assert.Equal(t, "CVE-2026-4444", linked["cve"])
	assert.Equal(t, "shopmon", linked["origin"])
	assert.Equal(t, "Seeded advisory CVE-2026-4444", linked["title"], "upstream content wins on merge")

	// One catalog row remains, and the official id resolves to it.
	_, err = env.Queries.GetComposerAdvisory(ctx, "CVE-2026-4444")
	assert.Error(t, err, "the canonical CVE row must be merged away")
	status, fetched := shopmonAdvRequest(t, "GET", env.Server.URL+"/api/advisories/CVE-2026-4444", userToken, nil)
	require.Equal(t, http.StatusOK, status, "%v", fetched)
	assert.Equal(t, shopmonID, fetched["advisoryId"])

	matches, err := env.Queries.GetEnvironmentAdvisoryMatches(ctx, int32(envID))
	require.NoError(t, err)
	require.Len(t, matches, 1)
	assert.Equal(t, shopmonID, matches[0].AdvisoryID)

	notified, err := env.Queries.ListNotifiedAdvisoryIDsForEnvironment(ctx, int32(envID))
	require.NoError(t, err)
	assert.Equal(t, []string{shopmonID}, notified, "already-alerted shops must not be alerted again")

	suppressions, err := env.Queries.ListSuppressionsForOrganizations(ctx, queries.ListSuppressionsForOrganizationsParams{
		OrganizationIds: []string{"org-1"},
	})
	require.NoError(t, err)
	require.Len(t, suppressions, 1)
	assert.Equal(t, shopmonID, suppressions[0].AdvisoryID)
}

func TestAdminImportGhsaAdvisory(t *testing.T) {
	env := testutil.Setup(t)
	adminToken := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	userToken := env.SeedUser(t, "user-1", "Regular User", "user@example.com", "user")

	// The test environment wires no GitHub fetcher, so the preview answers
	// 503 instead of pretending to have fetched anything.
	status, _ := shopmonAdvRequest(t, "POST", env.Server.URL+"/api/admin/advisories/import/ghsa/preview", adminToken, map[string]any{
		"input": "GHSA-vj8m-rx9g-45fq",
	})
	assert.Equal(t, http.StatusServiceUnavailable, status)
	status, _ = shopmonAdvRequest(t, "POST", env.Server.URL+"/api/admin/advisories/import/ghsa/preview", userToken, map[string]any{
		"input": "GHSA-vj8m-rx9g-45fq",
	})
	assert.Equal(t, http.StatusForbidden, status)

	// The reviewed import accepts the advisory URL in the GHSA field and
	// stores the advisory under its GHSA id with the github origin.
	payload := shopmonAdvCreatePayload(true)
	payload["ghsaId"] = "https://github.com/FriendsOfShopware/FroshPlatformMailArchive/security/advisories/GHSA-vj8m-rx9g-45fq"
	payload["packages"] = []map[string]any{
		{"packageName": "frosh/mail-platform-archive", "affectedVersions": ">=2.0.0,<=3.6.0"},
	}
	status, created := shopmonAdvRequest(t, "POST", env.Server.URL+"/api/admin/advisories/import/ghsa", adminToken, payload)
	require.Equal(t, http.StatusCreated, status, "%v", created)
	assert.Equal(t, "GHSA-vj8m-rx9g-45fq", created["advisoryId"])
	assert.Equal(t, "github", created["origin"])
	assert.Equal(t, "GHSA-vj8m-rx9g-45fq", created["ghsaId"])

	// Users see it as an official advisory: no Shopmon origin marking.
	status, fetched := shopmonAdvRequest(t, "GET", env.Server.URL+"/api/advisories/GHSA-vj8m-rx9g-45fq", userToken, nil)
	require.Equal(t, http.StatusOK, status, "%v", fetched)
	assert.Equal(t, "github", fetched["origin"])

	// A second import of the same GHSA conflicts instead of duplicating.
	status, body := shopmonAdvRequest(t, "POST", env.Server.URL+"/api/admin/advisories/import/ghsa", adminToken, payload)
	assert.Equal(t, http.StatusConflict, status, "%v", body)

	// The admin list can filter down to GitHub imports.
	status, list := shopmonAdvRequest(t, "GET", env.Server.URL+"/api/admin/advisories?origin=github", adminToken, nil)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, float64(1), list["total"])

	// When the pending CVE is assigned, it links via the pasted GHSA URL too
	// and the advisory answers to the CVE from then on.
	status, linked := shopmonAdvRequest(t, "POST", env.Server.URL+"/api/admin/advisories/GHSA-vj8m-rx9g-45fq/link-official", adminToken, map[string]any{
		"cve":    "CVE-2026-6666",
		"ghsaId": "https://github.com/FriendsOfShopware/FroshPlatformMailArchive/security/advisories/GHSA-vj8m-rx9g-45fq",
	})
	require.Equal(t, http.StatusOK, status, "%v", linked)
	assert.Equal(t, "CVE-2026-6666", linked["cve"])

	status, fetched = shopmonAdvRequest(t, "GET", env.Server.URL+"/api/advisories/CVE-2026-6666", userToken, nil)
	require.Equal(t, http.StatusOK, status, "%v", fetched)
	assert.Equal(t, "GHSA-vj8m-rx9g-45fq", fetched["advisoryId"])
}
