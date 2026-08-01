package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedAdvisory(t *testing.T, q *queries.Queries, id string, packages []string, visible bool, notesInternal *string) {
	t.Helper()
	ctx := context.Background()
	require.NoError(t, q.UpsertComposerAdvisory(ctx, queries.UpsertComposerAdvisoryParams{
		AdvisoryID: id,
		Title:      "Seeded advisory " + id,
		Link:       ptrStr("https://example.com/" + id),
		Cve:        ptrStr(id),
		GhsaID:     ptrStr("GHSA-" + id),
		Severity:   ptrStr("medium"),
		Sources:    []byte(`[{"name":"GitHub","remoteId":"GHSA-` + id + `"}]`),
		ReportedAt: pgtype.Timestamp{Time: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC), Valid: true},
	}))
	for i, pkg := range packages {
		require.NoError(t, q.UpsertComposerAdvisoryPackage(ctx, queries.UpsertComposerAdvisoryPackageParams{
			AdvisoryID:          id,
			PackageName:         pkg,
			PackagistAdvisoryID: fmt.Sprintf("PKSA-%s-%d", id, i),
			AffectedVersions:    ">=6.7.0.0,<6.7.10.1",
		}))
	}
	if !visible || notesInternal != nil {
		_, err := q.UpdateComposerAdvisoryEnrichment(ctx, queries.UpdateComposerAdvisoryEnrichmentParams{
			AdvisoryID:         id,
			IsVisible:          visible,
			NotesInternal:      notesInternal,
			AffectedComponents: []string{},
			Tags:               []string{"seed"},
			EnrichedBy:         ptrStr("test"),
		})
		require.NoError(t, err)
	}
}

func ptrStr(s string) *string { return &s }

func TestListAdvisoriesHidesInvisibleAndInternalNotes(t *testing.T) {
	env := testutil.Setup(t)
	seedAdvisory(t, env.Queries, "CVE-VISIBLE", []string{"shopware/core", "shopware/platform"}, true, ptrStr("secret-internal"))
	seedAdvisory(t, env.Queries, "CVE-HIDDEN", []string{"shopware/core"}, false, ptrStr("hidden-internal"))

	token := env.SeedUser(t, "user-1", "Regular User", "user@example.com", "user")

	// scope=all: this asserts visibility filtering, independent of whether any
	// shop inventory matches (the default scope is "affected").
	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/advisories?scope=all", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body struct {
		Advisories []map[string]any `json:"advisories"`
		Total      int              `json:"total"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Equal(t, 1, body.Total)
	require.Len(t, body.Advisories, 1)
	assert.Equal(t, "CVE-VISIBLE", body.Advisories[0]["advisoryId"])
	pkgs, ok := body.Advisories[0]["packages"].([]any)
	require.True(t, ok)
	assert.Len(t, pkgs, 2)
	_, hasInternal := body.Advisories[0]["notesInternal"]
	assert.False(t, hasInternal)

	req = testutil.NewRequest(t, "GET", env.Server.URL+"/api/advisories/CVE-HIDDEN", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAdminCanUpdateEnrichmentAndSeeHidden(t *testing.T) {
	env := testutil.Setup(t)
	seedAdvisory(t, env.Queries, "CVE-ADMIN", []string{"shopware/core"}, true, nil)

	adminToken := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")

	payload := map[string]any{
		"severityOverride":   "critical",
		"isVisible":          false,
		"notesPublic":        "Public note",
		"notesInternal":      "Internal note",
		"remediationSummary": "Upgrade now",
		"recommendedUpgrade": "6.7.10.1",
		"tags":               []string{"critical", "core"},
		"affectedComponents": []string{"admin"},
	}
	body, _ := json.Marshal(payload)
	req := testutil.NewRequest(t, "PATCH", env.Server.URL+"/api/admin/advisories/CVE-ADMIN", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var updated map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&updated))
	assert.Equal(t, "critical", updated["severityOverride"])
	assert.Equal(t, false, updated["isVisible"])
	assert.Equal(t, "Internal note", updated["notesInternal"])

	userToken := env.SeedUser(t, "user-2", "Regular User", "user2@example.com", "user")
	req = testutil.NewRequest(t, "GET", env.Server.URL+"/api/advisories/CVE-ADMIN", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAdminAdvisoryRequiresAdmin(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Regular User", "user@example.com", "user")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/admin/advisories", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// seedAdvisoryMatch links an advisory to an environment as if a scrape had
// matched it against the environment's SBOM.
func seedAdvisoryMatch(t *testing.T, q *queries.Queries, environmentID int32, advisoryID, pkg, installed string) {
	t.Helper()
	require.NoError(t, q.UpsertEnvironmentAdvisoryMatch(context.Background(), queries.UpsertEnvironmentAdvisoryMatchParams{
		EnvironmentID:    environmentID,
		AdvisoryID:       advisoryID,
		PackageName:      pkg,
		InstalledVersion: installed,
		AffectedVersions: ">=6.7.0.0,<6.7.10.1",
	}))
}

func TestListAdvisoriesScopeAndCounts(t *testing.T) {
	env := testutil.Setup(t)

	seedAdvisory(t, env.Queries, "CVE-HITS-ME", []string{"shopware/core"}, true, nil)
	seedAdvisory(t, env.Queries, "CVE-NOT-MINE", []string{"other/package"}, true, nil)

	token := env.SeedUser(t, "user-1", "Regular User", "user@example.com", "user")
	env.SeedOrganization(t, "org-1", "Org One", "org-one", "user-1")
	shopID := env.SeedShop(t, "org-1", "Shop One")
	envID := env.SeedEnvironment(t, "org-1", shopID, "Production", "https://shop.example.com")
	seedAdvisoryMatch(t, env.Queries, int32(envID), "CVE-HITS-ME", "shopware/core", "6.7.8.2")

	get := func(query string) (int, []map[string]any, map[string]any) {
		t.Helper()
		req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/advisories"+query, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var body struct {
			Advisories  []map[string]any `json:"advisories"`
			Total       int              `json:"total"`
			ScopeCounts map[string]any   `json:"scopeCounts"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		return body.Total, body.Advisories, body.ScopeCounts
	}

	// Default scope is "affected": only the advisory matching this org's shop.
	total, items, counts := get("")
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, "CVE-HITS-ME", items[0]["advisoryId"])
	assert.Equal(t, float64(1), items[0]["affectedEnvironmentCount"])
	// Tab badges report both scopes regardless of the active one.
	assert.Equal(t, float64(2), counts["all"])
	assert.Equal(t, float64(1), counts["affected"])

	// scope=all returns the full visible catalog.
	total, items, _ = get("?scope=all")
	assert.Equal(t, 2, total)
	assert.Len(t, items, 2)

	// Sorting by affected puts the matched advisory first even in "all".
	_, items, _ = get("?scope=all&sort=affected")
	require.Len(t, items, 2)
	assert.Equal(t, "CVE-HITS-ME", items[0]["advisoryId"])

	// Non-scope filters still apply within a scope.
	total, _, _ = get("?scope=affected&package=other/package")
	assert.Equal(t, 0, total)
}

// A user who belongs to no organization can have no matches, so the affected
// scope is empty while the catalog count still guides them to "all".
func TestListAdvisoriesAffectedScopeEmptyWithoutOrganizations(t *testing.T) {
	env := testutil.Setup(t)
	seedAdvisory(t, env.Queries, "CVE-SOMETHING", []string{"shopware/core"}, true, nil)
	token := env.SeedUser(t, "user-1", "Regular User", "user@example.com", "user")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/advisories", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body struct {
		Advisories  []map[string]any `json:"advisories"`
		Total       int              `json:"total"`
		ScopeCounts map[string]any   `json:"scopeCounts"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, 0, body.Total)
	assert.Empty(t, body.Advisories)
	assert.Equal(t, float64(1), body.ScopeCounts["all"])
	assert.Equal(t, float64(0), body.ScopeCounts["affected"])
}

// Suppressing an advisory moves it out of the "affecting my shops" scope and
// into the suppressed scope, for every environment of the shop at once.
func TestSuppressAdvisoryMovesItBetweenScopes(t *testing.T) {
	env := testutil.Setup(t)

	seedAdvisory(t, env.Queries, "CVE-SUPPRESS-ME", []string{"shopware/core"}, true, nil)
	seedAdvisory(t, env.Queries, "CVE-STAYS", []string{"shopware/core"}, true, nil)

	token := env.SeedUser(t, "user-1", "Owner User", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Org One", "org-one", "user-1")
	shopID := env.SeedShop(t, "org-1", "Shop One")
	prodID := env.SeedEnvironment(t, "org-1", shopID, "Production", "https://prod.example.com")
	stagingID := env.SeedEnvironment(t, "org-1", shopID, "Staging", "https://staging.example.com")

	for _, envID := range []int{prodID, stagingID} {
		seedAdvisoryMatch(t, env.Queries, int32(envID), "CVE-SUPPRESS-ME", "shopware/core", "6.7.8.2")
		seedAdvisoryMatch(t, env.Queries, int32(envID), "CVE-STAYS", "shopware/core", "6.7.8.2")
	}

	counts := func(scope string) (int, map[string]any) {
		t.Helper()
		req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/advisories?scope="+scope, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var body struct {
			Advisories  []map[string]any `json:"advisories"`
			Total       int              `json:"total"`
			ScopeCounts map[string]any   `json:"scopeCounts"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		// The tab badge must agree with the list it labels.
		assert.Len(t, body.Advisories, body.Total)
		return body.Total, body.ScopeCounts
	}

	affected, sc := counts("affected")
	assert.Equal(t, 2, affected)
	assert.Equal(t, float64(0), sc["suppressed"])

	payload, _ := json.Marshal(map[string]any{
		"shopId": shopID,
		"reason": "Mitigated by WAF rule",
	})
	req := testutil.NewRequest(t, "POST", env.Server.URL+"/api/advisories/CVE-SUPPRESS-ME/suppressions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	assert.Equal(t, "Mitigated by WAF rule", created["reason"])

	// One shop-level suppression covers both environments.
	affected, sc = counts("affected")
	assert.Equal(t, 1, affected)
	assert.Equal(t, float64(1), sc["affected"])
	assert.Equal(t, float64(1), sc["suppressed"])

	suppressed, _ := counts("suppressed")
	assert.Equal(t, 1, suppressed)

	// Revoking brings it back.
	suppressionID, ok := created["id"].(float64)
	require.True(t, ok)
	req = testutil.NewRequest(t, "DELETE", fmt.Sprintf("%s/api/suppressions/%d", env.Server.URL, int64(suppressionID)), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	affected, _ = counts("affected")
	assert.Equal(t, 2, affected)
}

// A duplicate suppression is a 409, not a silent overwrite -- the failure mode
// the legacy whole-array ignores PATCH could not avoid.
func TestSuppressAdvisoryRejectsDuplicateAndEmptyReason(t *testing.T) {
	env := testutil.Setup(t)
	seedAdvisory(t, env.Queries, "CVE-DUP", []string{"shopware/core"}, true, nil)

	token := env.SeedUser(t, "user-1", "Owner User", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Org One", "org-one", "user-1")
	shopID := env.SeedShop(t, "org-1", "Shop One")

	post := func(body map[string]any) int {
		t.Helper()
		payload, _ := json.Marshal(body)
		req := testutil.NewRequest(t, "POST", env.Server.URL+"/api/advisories/CVE-DUP/suppressions", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		return resp.StatusCode
	}

	assert.Equal(t, http.StatusBadRequest, post(map[string]any{"shopId": shopID, "reason": "   "}))
	assert.Equal(t, http.StatusCreated, post(map[string]any{"shopId": shopID, "reason": "accepted"}))
	assert.Equal(t, http.StatusConflict, post(map[string]any{"shopId": shopID, "reason": "accepted again"}))
}

// An expired suppression is inactive on every read path, so it must not block
// re-suppressing the same scope: the stale row is retired and a fresh one
// created, instead of a permanent 409.
func TestSuppressAdvisoryRenewableAfterExpiry(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()
	seedAdvisory(t, env.Queries, "CVE-EXPIRED", []string{"shopware/core"}, true, nil)

	token := env.SeedUser(t, "user-1", "Owner User", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Org One", "org-one", "user-1")
	shopID := env.SeedShop(t, "org-1", "Shop One")

	// Seed an already-expired, unrevoked suppression directly: the API itself
	// rejects a past expiresAt, which is exactly why this state only arises
	// from the passage of time.
	_, err := env.Pool.Exec(ctx, `
		INSERT INTO advisory_suppression (organization_id, shop_id, advisory_id, reason, expires_at, created_by, created_at)
		VALUES ('org-1', $1, 'CVE-EXPIRED', 'temporary', NOW() - INTERVAL '1 day', 'user-1', NOW() - INTERVAL '31 days')
	`, shopID)
	require.NoError(t, err)

	payload, _ := json.Marshal(map[string]any{"shopId": shopID, "reason": "still accepted"})
	req := testutil.NewRequest(t, "POST", env.Server.URL+"/api/advisories/CVE-EXPIRED/suppressions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// The contract declares shopName required on the created suppression.
	var created struct {
		ShopName string `json:"shopName"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	assert.Equal(t, "Shop One", created.ShopName)

	// The expired row is retired, not deleted: the audit trail survives.
	var revoked int
	require.NoError(t, env.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM advisory_suppression WHERE advisory_id = 'CVE-EXPIRED' AND revoked_at IS NOT NULL`,
	).Scan(&revoked))
	assert.Equal(t, 1, revoked)
}

// Revoking carries the same role requirement as creating. A member who could
// never have accepted a shop-wide risk must not be able to withdraw that
// acceptance either — otherwise the audit trail records an end to a decision
// nobody with the authority to make it chose to end.
func TestRevokeShopWideSuppressionRequiresElevatedRole(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()
	seedAdvisory(t, env.Queries, "CVE-REVOKE", []string{"shopware/core"}, true, nil)

	ownerToken := env.SeedUser(t, "user-1", "Owner User", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Org One", "org-one", "user-1")
	shopID := env.SeedShop(t, "org-1", "Shop One")

	memberToken := env.SeedUser(t, "user-2", "Member User", "member@example.com", "user")
	env.SeedMember(t, "org-1", "user-2", "member")

	// Owner creates a shop-wide suppression (member could not have).
	payload, _ := json.Marshal(map[string]any{"shopId": shopID, "reason": "accepted"})
	req := testutil.NewRequest(t, "POST", env.Server.URL+"/api/advisories/CVE-REVOKE/suppressions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	var created struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	_ = resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotZero(t, created.ID)

	revoke := func(token string) int {
		t.Helper()
		req := testutil.NewRequest(t, "DELETE",
			fmt.Sprintf("%s/api/suppressions/%d", env.Server.URL, created.ID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		return resp.StatusCode
	}

	assert.Equal(t, http.StatusForbidden, revoke(memberToken),
		"a member must not revoke a shop-wide suppression")

	var revokedAt *time.Time
	require.NoError(t, env.Pool.QueryRow(ctx,
		`SELECT revoked_at FROM advisory_suppression WHERE id = $1`, created.ID).Scan(&revokedAt))
	assert.Nil(t, revokedAt, "the suppression must still be active after the denied revoke")

	assert.Equal(t, http.StatusNoContent, revoke(ownerToken))
}

// A member of another organization must not be able to suppress for this shop.
func TestSuppressAdvisoryRejectsForeignOrganization(t *testing.T) {
	env := testutil.Setup(t)
	seedAdvisory(t, env.Queries, "CVE-FOREIGN", []string{"shopware/core"}, true, nil)

	env.SeedUser(t, "user-1", "Owner User", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Org One", "org-one", "user-1")
	shopID := env.SeedShop(t, "org-1", "Shop One")

	outsiderToken := env.SeedUser(t, "user-2", "Outsider", "outsider@example.com", "user")

	payload, _ := json.Marshal(map[string]any{"shopId": shopID, "reason": "not mine"})
	req := testutil.NewRequest(t, "POST", env.Server.URL+"/api/advisories/CVE-FOREIGN/suppressions", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+outsiderToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// The admin update endpoint is a PATCH with every field optional, so a body
// that sets one field must not clear the others — and must never republish an
// advisory an admin deliberately hid.
func TestAdminUpdateAdvisoryPreservesOmittedFields(t *testing.T) {
	env := testutil.Setup(t)
	seedAdvisory(t, env.Queries, "CVE-PATCH", []string{"shopware/core"}, true, nil)

	token := env.SeedUser(t, "admin-1", "Admin", "admin@example.com", "admin")

	patch := func(body map[string]any) *http.Response {
		t.Helper()
		payload, _ := json.Marshal(body)
		req := testutil.NewRequest(t, "PATCH", env.Server.URL+"/api/admin/advisories/CVE-PATCH", bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		return resp
	}

	// Full enrichment, including hiding the advisory.
	resp := patch(map[string]any{
		"isVisible":     false,
		"notesInternal": "under investigation",
		"notesPublic":   "mitigated upstream",
		"tags":          []string{"ssrf"},
	})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	// A partial follow-up touching only tags.
	resp = patch(map[string]any{"tags": []string{"ssrf", "triaged"}})
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var updated struct {
		IsVisible     bool     `json:"isVisible"`
		NotesInternal *string  `json:"notesInternal"`
		NotesPublic   *string  `json:"notesPublic"`
		Tags          []string `json:"tags"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&updated))
	_ = resp.Body.Close()

	assert.False(t, updated.IsVisible, "omitting isVisible must not republish a hidden advisory")
	require.NotNil(t, updated.NotesInternal)
	assert.Equal(t, "under investigation", *updated.NotesInternal)
	require.NotNil(t, updated.NotesPublic)
	assert.Equal(t, "mitigated upstream", *updated.NotesPublic)
	assert.Equal(t, []string{"ssrf", "triaged"}, updated.Tags)

	// An explicit empty string still clears a field. Decoded into a fresh
	// struct: reusing `updated` would keep the previous value for a field the
	// response omits, hiding exactly what this asserts.
	resp = patch(map[string]any{"notesPublic": ""})
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var cleared struct {
		NotesPublic   *string `json:"notesPublic"`
		NotesInternal *string `json:"notesInternal"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&cleared))
	_ = resp.Body.Close()
	assert.Nil(t, cleared.NotesPublic, "an explicit empty value must clear the column")
	require.NotNil(t, cleared.NotesInternal, "clearing one field must not clear its neighbours")
	assert.Equal(t, "under investigation", *cleared.NotesInternal)

	// An out-of-enum severity is rejected rather than stored and silently
	// dropped on read.
	resp = patch(map[string]any{"severityOverride": "catastrophic"})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	_ = resp.Body.Close()
}
