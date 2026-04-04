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

func TestListShops(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	env.SeedShop(t, "org-1", "Shop A")
	env.SeedShop(t, "org-1", "Shop B")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/organizations/org-1/shops", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var shops []api.Shop
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&shops))
	assert.Len(t, shops, 2)
}

func TestListShops_NotMember(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	// Create org without user-1 as member
	_, err := env.Pool.Exec(t.Context(), `INSERT INTO organization (id, name, slug, created_at) VALUES ('org-other', 'Other Org', 'other-org', NOW())`)
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/organizations/org-other/shops", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestCreateShop(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")

	body, _ := json.Marshal(api.CreateShopRequest{
		Name:        "New Shop",
		Description: strPtr("A test shop"),
	})

	req, _ := http.NewRequest("POST", env.Server.URL+"/api/organizations/org-1/shops", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotNil(t, result["id"])
}

func TestUpdateShop(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Old Name")

	newName := "Updated Name"
	body, _ := json.Marshal(api.UpdateShopRequest{
		Name: &newName,
	})

	req, _ := http.NewRequest("PATCH", fmt.Sprintf("%s/api/organizations/org-1/shops/%d", env.Server.URL, shopID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestDeleteShop(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "To Delete")

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/organizations/org-1/shops/%d", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify it's gone
	req, _ = http.NewRequest("GET", env.Server.URL+"/api/organizations/org-1/shops", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var shops []api.Shop
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&shops))
	assert.Empty(t, shops)
}

func TestDeleteShop_WithEnvironments_Fails(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Has Environments")
	env.SeedEnvironment(t, "org-1", shopID, "Environment", "https://env.example.com")

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/organizations/org-1/shops/%d", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func strPtr(s string) *string {
	return &s
}
