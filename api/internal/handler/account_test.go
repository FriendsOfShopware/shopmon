package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAccountMe_Unauthenticated(t *testing.T) {
	env := testutil.Setup(t)

	resp, err := http.Get(env.Server.URL + "/api/account/me")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetAccountMe(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/account/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

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
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var orgs []api.AccountOrganization
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&orgs))
	require.Len(t, orgs, 1)
	assert.Equal(t, "org-1", orgs[0].Id)
	assert.Equal(t, "Test Org", orgs[0].Name)
	assert.Equal(t, "Test Org", orgs[0].Name)
}

func TestGetAccountShops(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	projectID := env.SeedProject(t, "org-1", "Test Project")
	env.SeedShop(t, "org-1", projectID, "My Shop", "https://shop.example.com")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/account/shops", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var shops []api.AccountShop
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&shops))
	require.Len(t, shops, 1)
	assert.Equal(t, "My Shop", shops[0].Name)
	assert.Equal(t, "https://shop.example.com", shops[0].Url)
	assert.Equal(t, "green", shops[0].Status)
}

func TestGetAccountProjects(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	env.SeedProject(t, "org-1", "My Project")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/account/projects", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var projects []api.AccountProject
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&projects))
	require.Len(t, projects, 1)
	assert.Equal(t, "My Project", projects[0].Name)
	assert.Equal(t, "org-1", projects[0].OrganizationId)
}

func TestGetAccountChangelogs_Empty(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/account/changelogs", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var changelogs []api.AccountChangelog
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&changelogs))
	assert.Empty(t, changelogs)
}

func TestGetAccountSubscribedShops_Empty(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/account/subscribed-shops", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var shops []api.SubscribedShop
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&shops))
	assert.Empty(t, shops)
}
