package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
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

	preview := func(token string, packages []api.AdvisoryPackageInput, status int) api.AdminAdvisoryImpact {
		t.Helper()
		body, err := json.Marshal(api.AdminPreviewAdvisoryImpactRequest{Packages: packages})
		require.NoError(t, err)
		req := testutil.NewRequest(t, "POST", env.Server.URL+"/api/admin/advisories/impact", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, status, resp.StatusCode)
		var result api.AdminAdvisoryImpact
		if status == http.StatusOK {
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
		}
		return result
	}
	packages := []api.AdvisoryPackageInput{
		{PackageName: " SHOPWARE/CORE ", AffectedVersions: " >=6.7.0.0,<6.7.10.1 "},
		{PackageName: "shopware/storefront", AffectedVersions: ">=6.7.0.0,<6.7.10.1"},
	}
	result := preview(adminToken, packages, http.StatusOK)
	assert.Equal(t, 2, result.ShopCount)
	assert.Equal(t, 3, result.EnvironmentCount)
	require.Len(t, result.Shops, 2)
	assert.Equal(t, shop1, result.Shops[0].ShopId)
	assert.Equal(t, "One", result.Shops[0].ShopName)
	assert.Equal(t, "org-1", result.Shops[0].OrganizationId)
	assert.Equal(t, "One", result.Shops[0].OrganizationName)
	require.Len(t, result.Shops[0].Matches, 4)
	assert.Equal(t, api.AdminAdvisoryImpactMatch{
		EnvironmentId: prod, EnvironmentName: "Production", PackageName: "shopware/core",
		InstalledVersion: "6.7.8.2", AffectedVersions: ">=6.7.0.0,<6.7.10.1",
	}, result.Shops[0].Matches[0])
	assert.Equal(t, shop2, result.Shops[1].ShopId)
	assert.Equal(t, "org-2", result.Shops[1].OrganizationId)
	require.Len(t, result.Shops[1].Matches, 1)
	assert.Equal(t, other, result.Shops[1].Matches[0].EnvironmentId)

	// Suppression does not change inventory-based previews of arbitrary constraints.
	_, err := env.Queries.CreateAdvisorySuppression(ctx, queries.CreateAdvisorySuppressionParams{
		OrganizationID: "org-1", ShopID: int32(shop1), AdvisoryID: "CVE-DRAFT", Reason: "Mitigated",
	})
	require.NoError(t, err)
	assert.Equal(t, result, preview(adminToken, packages, http.StatusOK))

	// Submitted ranges, not the saved advisory, determine the preview.
	wider := []api.AdvisoryPackageInput{{PackageName: "shopware/core", AffectedVersions: "<6.8.0.0"}}
	expanded := preview(adminToken, wider, http.StatusOK)
	assert.Equal(t, 3, expanded.ShopCount)
	assert.Equal(t, 4, expanded.EnvironmentCount)
	none := preview(adminToken, []api.AdvisoryPackageInput{{PackageName: "shopware/core", AffectedVersions: "<6.7.0.0"}}, http.StatusOK)
	assert.Equal(t, api.AdminAdvisoryImpact{Shops: []api.AdminAdvisoryImpactShop{}}, none)

	seedInventory(other, "shopware/core", "6.7.10.1")
	updated := preview(adminToken, packages, http.StatusOK)
	assert.Equal(t, 1, updated.ShopCount)
	assert.Equal(t, 2, updated.EnvironmentCount)

	// Runtime versions (php, mysql, mariadb) stored on the environment also participate in impact previews.
	_, err = env.Pool.Exec(ctx, `UPDATE environment SET php_version = '8.3.14', mysql_version = '8.0.36' WHERE id = $1`, prod)
	require.NoError(t, err)
	_, err = env.Pool.Exec(ctx, `UPDATE environment SET php_version = '8.4.2', mysql_version = '10.11.6' WHERE id = $1`, other)
	require.NoError(t, err)

	runtimeImpact := preview(adminToken, []api.AdvisoryPackageInput{
		{PackageName: "php", AffectedVersions: "<8.4.0"},
		{PackageName: "mysql", AffectedVersions: ">=8.0.0,<8.0.40"},
		{PackageName: "mariadb", AffectedVersions: ">=10.11.0,<10.11.10"},
	}, http.StatusOK)
	assert.Equal(t, 2, runtimeImpact.ShopCount)
	assert.Equal(t, 2, runtimeImpact.EnvironmentCount)
	require.Len(t, runtimeImpact.Shops, 2)
	assert.Equal(t, []api.AdminAdvisoryImpactMatch{
		{EnvironmentId: prod, EnvironmentName: "Production", PackageName: "mysql", InstalledVersion: "8.0.36", AffectedVersions: ">=8.0.0,<8.0.40"},
		{EnvironmentId: prod, EnvironmentName: "Production", PackageName: "php", InstalledVersion: "8.3.14", AffectedVersions: "<8.4.0"},
	}, runtimeImpact.Shops[0].Matches)
	assert.Equal(t, []api.AdminAdvisoryImpactMatch{
		{EnvironmentId: other, EnvironmentName: "Production", PackageName: "mariadb", InstalledVersion: "10.11.6", AffectedVersions: ">=10.11.0,<10.11.10"},
	}, runtimeImpact.Shops[1].Matches)

	// Previewing neither saves package ranges nor publishes match rows.
	saved, err := env.Queries.ListComposerAdvisoryPackagesForAdvisory(ctx, "CVE-DRAFT")
	require.NoError(t, err)
	require.Len(t, saved, 2)
	assert.Equal(t, ">=6.7.0.0,<6.7.10.1", saved[0].AffectedVersions)
	var matchCount int
	require.NoError(t, env.Pool.QueryRow(ctx, "SELECT count(*) FROM environment_advisory_match WHERE advisory_id = 'CVE-DRAFT'").Scan(&matchCount))
	assert.Zero(t, matchCount)

	preview("", packages, http.StatusUnauthorized)
	preview(userToken, packages, http.StatusForbidden)
	for _, invalid := range [][]api.AdvisoryPackageInput{
		{},
		{{PackageName: "invalid", AffectedVersions: "<2.0"}},
		{{PackageName: "shopware/core", AffectedVersions: "*"}},
		{{PackageName: "shopware/core", AffectedVersions: "not-a-range"}},
		{{PackageName: "shopware/core", AffectedVersions: ""}},
		{{PackageName: "shopware/core", AffectedVersions: "<2.0"}, {PackageName: "SHOPWARE/CORE", AffectedVersions: "<3.0"}},
	} {
		preview(adminToken, invalid, http.StatusBadRequest)
	}
	tooMany := make([]api.AdvisoryPackageInput, 101)
	for i := range tooMany {
		tooMany[i] = api.AdvisoryPackageInput{PackageName: fmt.Sprintf("vendor/package-%d", i), AffectedVersions: "<2.0"}
	}
	preview(adminToken, tooMany, http.StatusBadRequest)
}
