package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminGetStats(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	env.SeedUser(t, "user-1", "Regular User", "user@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "admin-1")
	projectID := env.SeedProject(t, "org-1", "Test Project")
	env.SeedShop(t, "org-1", projectID, "Shop 1", "https://shop1.example.com")
	env.SeedShop(t, "org-1", projectID, "Shop 2", "https://shop2.example.com")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/admin/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var stats api.AdminStats
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&stats))
	assert.Equal(t, 2, stats.TotalUsers)
	assert.Equal(t, 1, stats.TotalOrganizations)
	assert.Equal(t, 2, stats.TotalShops)
	assert.Equal(t, 2, stats.ShopsByStatus.Green)
}

func TestAdminGetStats_NonAdmin(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Regular User", "user@example.com", "user")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/admin/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAdminGetOrganizations(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	env.SeedOrganization(t, "org-1", "Alpha Org", "alpha-org", "admin-1")
	env.SeedOrganization(t, "org-2", "Beta Org", "beta-org", "admin-1")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/admin/organizations?limit=10&offset=0", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result api.AdminOrganizationsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, 2, result.Total)
	assert.Len(t, result.Organizations, 2)
}

func TestAdminGetShops(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "admin-1")
	projectID := env.SeedProject(t, "org-1", "Test Project")
	env.SeedShop(t, "org-1", projectID, "Shop 1", "https://shop1.example.com")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/admin/shops?limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result api.AdminShopsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, 1, result.Total)
	assert.Len(t, result.Shops, 1)
}

func TestAdminGetGrowth(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/admin/growth", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var growth api.AdminGrowth
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&growth))
	assert.NotNil(t, growth.Users)
	assert.NotNil(t, growth.Shops)
}

func TestAdminGetShopwareVersions(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "admin-1")
	projectID := env.SeedProject(t, "org-1", "Test Project")
	env.SeedShop(t, "org-1", projectID, "Shop 1", "https://shop1.example.com")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/admin/shopware-versions", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var versions []api.ShopwareVersionCount
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&versions))
	require.Len(t, versions, 1)
	assert.Equal(t, "6.5.0.0", versions[0].Version)
	assert.Equal(t, 1, versions[0].Count)
}
