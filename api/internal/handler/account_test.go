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

func TestGetAccountMe_Unauthenticated(t *testing.T) {
	env := testutil.Setup(t)

	resp, err := http.Get(env.Server.URL + "/api/account/me")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetAccountMe(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/account/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var user api.UserProfile
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&user))
	assert.Equal(t, "user-1", user.Id)
	assert.Equal(t, "Test User", user.DisplayName)
	assert.Equal(t, "test@example.com", string(user.Email))
}

func TestGetAccountOrganizations(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/account/organizations", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var orgs []api.AccountOrganization
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&orgs))
	require.Len(t, orgs, 1)
	assert.Equal(t, "org-1", orgs[0].Id)
	assert.Equal(t, "Test Org", orgs[0].Name)
	assert.Equal(t, "Test Org", orgs[0].Name)
}

func TestGetAccountEnvironments(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/account/environments", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var environments []api.AccountEnvironment
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&environments))
	require.Len(t, environments, 1)
	assert.Equal(t, "My Environment", environments[0].Name)
	assert.Equal(t, "https://env.example.com", environments[0].Url)
	assert.Equal(t, "green", environments[0].Status)
}

func TestGetAccountShops(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	env.SeedShop(t, "org-1", "My Shop")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/account/shops", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var shops []api.AccountShop
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&shops))
	require.Len(t, shops, 1)
	assert.Equal(t, "My Shop", shops[0].Name)
	assert.Equal(t, "org-1", shops[0].OrganizationId)
}

func TestGetAccountChangelogs_Empty(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/account/changelogs", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var changelogs []api.AccountChangelog
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&changelogs))
	assert.Empty(t, changelogs)
}

func TestGetAccountSubscribedEnvironments_Empty(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/account/subscribed-environments", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var environments []api.SubscribedEnvironment
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&environments))
	assert.Empty(t, environments)
}
