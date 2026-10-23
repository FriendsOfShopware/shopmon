package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminAdvisoryImpactFromInventory(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()
	adminToken := env.SeedUser(t, "admin", "Admin", "admin@example.com", "admin")
	userToken := env.SeedUser(t, "owner", "Owner", "owner@example.com", "user")
	// The admin belongs to neither organization: impact must cover both.
	env.SeedOrganization(t, "org-1", "One", "one", "owner")
	env.SeedOrganization(t, "org-2", "Two", "two", "owner")
	shop1 := env.SeedShop(t, "org-1", "One")
	shop2 := env.SeedShop(t, "org-2", "Two")
	prod := env.SeedEnvironment(t, "org-1", shop1, "Production", "https://one.example.com")
	stage := env.SeedEnvironment(t, "org-1", shop1, "Staging", "https://stage.example.com")
	other := env.SeedEnvironment(t, "org-2", shop2, "Production", "https://two.example.com")
	patchedShop := env.SeedShop(t, "org-2", "Patched")
	patched := env.SeedEnvironment(t, "org-2", patchedShop, "Production", "https://patched.example.com")
	noInventory := env.SeedShop(t, "org-2", "No inventory")
	env.SeedEnvironment(t, "org-2", noInventory, "Production", "https://unknown.example.com")
	seedAdvisory(t, env.Queries, "CVE-DRAFT", []string{"shopware/core", "shopware/storefront"}, false, nil)
	seedAdvisory(t, env.Queries, "CVE-VISIBLE", []string{"shopware/core", "shopware/storefront"}, true, nil)
	seedAdvisory(t, env.Queries, "CVE-ZERO", []string{"unused/package"}, true, nil)

	seedInventory := func(environmentID int, pkg, version string) {
		t.Helper()
		require.NoError(t, env.Queries.UpsertEnvironmentSbomComponent(ctx, queries.UpsertEnvironmentSbomComponentParams{
			EnvironmentID: int32(environmentID), PackageName: pkg, Version: version,
		}))
	}
	for _, id := range []int{prod, stage, other} {
		seedInventory(id, "shopware/core", "6.7.8.2")
	}
	for _, id := range []int{prod, stage} {
		seedInventory(id, "shopware/storefront", "6.7.8.2")
	}
	seedInventory(patched, "shopware/core", "6.7.10.1")
	// Stale background matches must not inflate current inventory counts.
	seedAdvisoryMatch(t, env.Queries, int32(patched), "CVE-VISIBLE", "shopware/core", "6.7.8.2")

	get := func(path string, target any) {
		t.Helper()
		req := testutil.NewRequest(t, "GET", env.Server.URL+"/api"+path, nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.NoError(t, json.NewDecoder(resp.Body).Decode(target))
	}
	want := api.AdminAdvisoryImpact{ShopCount: 2, EnvironmentCount: 3, ActiveShopCount: 2}
	var list api.AdminAdvisoryListResponse
	get("/admin/advisories", &list)
	require.Len(t, list.Advisories, 3)
	for _, item := range list.Advisories {
		if item.AdvisoryId == "CVE-ZERO" {
			assert.Equal(t, api.AdminAdvisoryImpact{}, item.Impact)
		} else {
			assert.Equal(t, want, item.Impact, item.AdvisoryId)
		}
		var detail api.AdminAdvisory
		get("/admin/advisories/"+item.AdvisoryId, &detail)
		assert.Equal(t, item.Impact, detail.Impact)
	}
	// Aliases resolve to the same impact; pagination does not change counts.
	var alias api.AdminAdvisory
	get("/admin/advisories/PKSA-CVE-DRAFT-0", &alias)
	assert.Equal(t, want, alias.Impact)
	get("/admin/advisories?limit=1&offset=1&package=shopware/core", &list)
	require.Len(t, list.Advisories, 1)
	assert.Equal(t, 2, list.Total)
	assert.Equal(t, want, list.Advisories[0].Impact)

	suppress := func(org string, shop int, environmentID *int32, expires pgtype.Timestamp) int64 {
		t.Helper()
		row, err := env.Queries.CreateAdvisorySuppression(ctx, queries.CreateAdvisorySuppressionParams{
			OrganizationID: org, ShopID: int32(shop), EnvironmentID: environmentID,
			AdvisoryID: "CVE-DRAFT", Reason: "Mitigated", ExpiresAt: expires,
		})
		require.NoError(t, err)
		return row.ID
	}
	checkDraft := func(active, suppressed int) {
		t.Helper()
		var detail api.AdminAdvisory
		get("/admin/advisories/CVE-DRAFT", &detail)
		assert.Equal(t, api.AdminAdvisoryImpact{
			ShopCount: 2, EnvironmentCount: 3, ActiveShopCount: active, SuppressedEnvironmentCount: suppressed,
		}, detail.Impact)
	}
	suppress("org-1", shop1, new(int32(prod)), pgtype.Timestamp{})
	checkDraft(2, 1) // staging keeps the first shop active
	shopSuppression := suppress("org-1", shop1, nil, pgtype.Timestamp{})
	checkDraft(1, 2) // overlapping suppressions still count each environment once
	suppress("org-2", shop2, nil, pgtype.Timestamp{Time: time.Now().UTC().Add(-time.Hour), Valid: true})
	checkDraft(1, 2) // expired suppression has no effect
	_, err := env.Queries.RevokeAdvisorySuppression(ctx, queries.RevokeAdvisorySuppressionParams{
		ID: shopSuppression, OrganizationIds: []string{"org-1"},
	})
	require.NoError(t, err)
	checkDraft(2, 1)

	// Counts update immediately when inventory changes, without a rematch job.
	seedInventory(other, "shopware/core", "6.7.10.1")
	get("/admin/advisories/CVE-DRAFT", &alias)
	assert.Equal(t, api.AdminAdvisoryImpact{
		ShopCount: 1, EnvironmentCount: 2, ActiveShopCount: 1, SuppressedEnvironmentCount: 1,
	}, alias.Impact)
	var matchCount int
	require.NoError(t, env.Pool.QueryRow(ctx, "SELECT count(*) FROM environment_advisory_match WHERE advisory_id = 'CVE-DRAFT'").Scan(&matchCount))
	assert.Zero(t, matchCount, "previewing a draft must not publish matches")

	// Global impact is available only through admin responses.
	for _, path := range []string{"/admin/advisories", "/admin/advisories/CVE-VISIBLE"} {
		for _, token := range []string{"", userToken} {
			req := testutil.NewRequest(t, "GET", env.Server.URL+"/api"+path, nil)
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
			wantStatus := http.StatusUnauthorized
			if token != "" {
				wantStatus = http.StatusForbidden
			}
			assert.Equal(t, wantStatus, resp.StatusCode, fmt.Sprintf("%s token=%t", path, token != ""))
		}
	}
	var public map[string]any
	get("/advisories/CVE-VISIBLE", &public)
	assert.NotContains(t, public, "impact")
}
