package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	env.SeedEnvironment(t, "org-1", shopID, "Environment 1", "https://env1.example.com")
	env.SeedEnvironment(t, "org-1", shopID, "Environment 2", "https://env2.example.com")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/admin/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var stats api.AdminStats
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&stats))
	assert.Equal(t, 2, stats.TotalUsers)
	assert.Equal(t, 1, stats.TotalOrganizations)
	assert.Equal(t, 2, stats.TotalEnvironments)
	assert.Equal(t, 2, stats.EnvironmentsByStatus.Green)
}

func TestAdminGetStats_NonAdmin(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Regular User", "user@example.com", "user")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/admin/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAdminGetOrganizations(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	env.SeedOrganization(t, "org-1", "Alpha Org", "alpha-org", "admin-1")
	env.SeedOrganization(t, "org-2", "Beta Org", "beta-org", "admin-1")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/admin/organizations?limit=10&offset=0", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result api.AdminOrganizationsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, 2, result.Total)
	assert.Len(t, result.Organizations, 2)
}

func TestAdminGetOrganizations_Search(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	env.SeedOrganization(t, "org-1", "Alpha Org", "alpha-org", "admin-1")
	env.SeedOrganization(t, "org-2", "Beta Org", "beta-org", "admin-1")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/admin/organizations?searchValue=alpha", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result api.AdminOrganizationsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, 1, result.Total)
	require.Len(t, result.Organizations, 1)
	assert.Equal(t, "Alpha Org", result.Organizations[0].Name)
}

func TestAdminGetOrganizations_SortByName(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	env.SeedOrganization(t, "org-1", "Beta Org", "beta-org", "admin-1")
	env.SeedOrganization(t, "org-2", "Alpha Org", "alpha-org", "admin-1")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/admin/organizations?sortBy=name&sortDirection=asc", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result api.AdminOrganizationsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.Len(t, result.Organizations, 2)
	assert.Equal(t, "Alpha Org", result.Organizations[0].Name)
	assert.Equal(t, "Beta Org", result.Organizations[1].Name)
}

func TestAdminGetEnvironments(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "admin-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	env.SeedEnvironment(t, "org-1", shopID, "Environment 1", "https://env1.example.com")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/admin/environments?limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result api.AdminEnvironmentsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, 1, result.Total)
	require.Len(t, result.Environments, 1)
	require.NotNil(t, result.Environments[0].ShopName)
	assert.Equal(t, "Test Shop", *result.Environments[0].ShopName)
}

func TestAdminGetEnvironments_Search(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "admin-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	env.SeedEnvironment(t, "org-1", shopID, "Production", "https://prod.example.com")
	env.SeedEnvironment(t, "org-1", shopID, "Staging", "https://staging.example.com")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/admin/environments?searchValue=staging", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result api.AdminEnvironmentsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, 1, result.Total)
	require.Len(t, result.Environments, 1)
	assert.Equal(t, "Staging", result.Environments[0].Name)
}

func TestAdminGetGrowth(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/admin/growth", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var growth api.AdminGrowth
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&growth))
	assert.NotNil(t, growth.Users)
	assert.NotNil(t, growth.Environments)
}

func TestAdminGetShopwareVersions(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "admin-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	env.SeedEnvironment(t, "org-1", shopID, "Environment 1", "https://env1.example.com")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/admin/shopware-versions", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var versions []api.ShopwareVersionCount
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&versions))
	require.Len(t, versions, 1)
	assert.Equal(t, "6.5.0.0", versions[0].Version)
	assert.Equal(t, 1, versions[0].Count)
}

func TestAdminGetOrganizationDetail(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	env.SeedOrganization(t, "org-1", "Acme Org", "acme-org", "admin-1")
	shopID := env.SeedShop(t, "org-1", "Acme Shop")
	env.SeedEnvironment(t, "org-1", shopID, "Production", "https://prod.example.com")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/admin/organizations/org-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var detail api.AdminOrganizationDetail
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&detail))
	assert.Equal(t, "Acme Org", detail.Name)
	assert.Equal(t, 1, detail.EnvironmentCount)
	require.Len(t, detail.Members, 1)
	assert.Equal(t, "admin@example.com", detail.Members[0].Email)
	require.Len(t, detail.Environments, 1)
	assert.Equal(t, "Production", detail.Environments[0].Name)
}

func TestAdminGetOrganizationDetail_NotFound(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/admin/organizations/missing", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAdminGetEnvironmentDetail(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	env.SeedOrganization(t, "org-1", "Acme Org", "acme-org", "admin-1")
	shopID := env.SeedShop(t, "org-1", "Acme Shop")
	envID := env.SeedEnvironment(t, "org-1", shopID, "Production", "https://prod.example.com")

	req := testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/admin/environments/%d", env.Server.URL, envID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var detail api.AdminEnvironmentDetail
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&detail))
	assert.Equal(t, "Production", detail.Name)
	assert.Equal(t, "Acme Org", detail.OrganizationName)
	assert.Equal(t, "Acme Shop", detail.ShopName)
	// Secrets must never be exposed.
	body, _ := json.Marshal(detail)
	assert.NotContains(t, string(body), "client_secret")
	assert.NotContains(t, string(body), "environment_token")
}

func TestAdminGetEnvironmentDetail_NotFound(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/admin/environments/999999", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAdminGetAuditLog(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	env.SeedUser(t, "user-1", "Target User", "target@example.com", "user")

	// Trigger an audited admin action: set a role.
	roleReq := testutil.NewRequest(t, "PATCH", env.Server.URL+"/api/auth/admin/users/user-1/role", bytes.NewReader([]byte(`{"role":"admin"}`)))
	roleReq.Header.Set("Authorization", "Bearer "+token)
	roleReq.Header.Set("Content-Type", "application/json")
	roleResp, err := http.DefaultClient.Do(roleReq)
	require.NoError(t, err)
	_ = roleResp.Body.Close()
	require.Equal(t, http.StatusOK, roleResp.StatusCode)

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/admin/audit-log", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result api.AdminAuditLogResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.GreaterOrEqual(t, result.Total, 1)
	require.NotEmpty(t, result.Entries)
	assert.Equal(t, "admin.set_role", result.Entries[0].Action)
	require.NotNil(t, result.Entries[0].ActorName)
	assert.Equal(t, "Admin User", *result.Entries[0].ActorName)
}

func TestAdminGetAuditLog_NonAdmin(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Regular User", "user@example.com", "user")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/admin/audit-log", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
