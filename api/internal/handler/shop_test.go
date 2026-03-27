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

func TestGetOrganizationShops(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	projectID := env.SeedProject(t, "org-1", "Test Project")
	env.SeedShop(t, "org-1", projectID, "Shop A", "https://a.example.com")
	env.SeedShop(t, "org-1", projectID, "Shop B", "https://b.example.com")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/organizations/org-1/shops", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var shops []json.RawMessage
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&shops))
	assert.Len(t, shops, 2)
}

func TestGetShop(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	projectID := env.SeedProject(t, "org-1", "Test Project")
	shopID := env.SeedShop(t, "org-1", projectID, "My Shop", "https://shop.example.com")

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/shops/%d", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var shop api.ShopDetail
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&shop))
	assert.Equal(t, "My Shop", shop.Name)
	assert.Equal(t, "https://shop.example.com", shop.Url)
	assert.Equal(t, "6.5.0.0", shop.ShopwareVersion)
	assert.Equal(t, "green", shop.Status)
	assert.Equal(t, "Test Org", shop.OrganizationName)
	assert.NotNil(t, shop.Extensions)
	assert.NotNil(t, shop.ScheduledTasks)
	assert.NotNil(t, shop.Queues)
	assert.NotNil(t, shop.Checks)
}

func TestGetShop_NotMember(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	// Create another user's org and shop
	env.SeedUser(t, "user-2", "Other User", "other@example.com", "user")
	env.SeedOrganization(t, "org-2", "Other Org", "other-org", "user-2")
	projectID := env.SeedProject(t, "org-2", "Other Project")
	shopID := env.SeedShop(t, "org-2", projectID, "Other Shop", "https://other.example.com")

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/shops/%d", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestDeleteShop(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	projectID := env.SeedProject(t, "org-1", "Test Project")
	shopID := env.SeedShop(t, "org-1", projectID, "To Delete", "https://del.example.com")

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/shops/%d", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify it's gone
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/shops/%d", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Shop is deleted, so org membership check on a non-existent shop returns not found
	assert.True(t, resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden)
}

func TestGetShopSubscription(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	projectID := env.SeedProject(t, "org-1", "Test Project")
	shopID := env.SeedShop(t, "org-1", projectID, "My Shop", "https://shop.example.com")

	// Initially not subscribed
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/shops/%d/subscribe", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]bool
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.False(t, result["subscribed"])
}

func TestUpdateShop(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	projectID := env.SeedProject(t, "org-1", "Test Project")
	shopID := env.SeedShop(t, "org-1", projectID, "Old Name", "https://old.example.com")

	newName := "New Name"
	body, _ := json.Marshal(api.UpdateShopRequest{
		Name:      &newName,
		ProjectId: projectID,
	})

	req, _ := http.NewRequest("PATCH", fmt.Sprintf("%s/api/shops/%d", env.Server.URL, shopID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify the name changed by fetching the shop
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/shops/%d", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var shop api.ShopDetail
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&shop))
	assert.Equal(t, "New Name", shop.Name)
}

func TestSubscribeAndUnsubscribeShop(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	projectID := env.SeedProject(t, "org-1", "Test Project")
	shopID := env.SeedShop(t, "org-1", projectID, "My Shop", "https://shop.example.com")

	// Initially not subscribed
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/shops/%d/subscribe", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var result map[string]bool
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.False(t, result["subscribed"])

	// Subscribe
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/api/shops/%d/subscribe", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify subscribed
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/shops/%d/subscribe", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.True(t, result["subscribed"])

	// Unsubscribe
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("%s/api/shops/%d/subscribe", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify unsubscribed
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/shops/%d/subscribe", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.False(t, result["subscribed"])
}

func TestUpdateSitespeedSettings(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	projectID := env.SeedProject(t, "org-1", "Test Project")
	shopID := env.SeedShop(t, "org-1", projectID, "My Shop", "https://shop.example.com")

	urls := []string{"https://shop.example.com/", "https://shop.example.com/products"}
	body, _ := json.Marshal(api.SitespeedSettingsRequest{
		Enabled: true,
		Urls:    &urls,
	})

	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/api/shops/%d/sitespeed-settings", env.Server.URL, shopID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify settings by fetching the shop
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/shops/%d", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var shop api.ShopDetail
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&shop))
	assert.True(t, shop.SitespeedEnabled)
	require.NotNil(t, shop.SitespeedUrls)
	assert.Len(t, *shop.SitespeedUrls, 2)
	assert.Contains(t, *shop.SitespeedUrls, "https://shop.example.com/")
	assert.Contains(t, *shop.SitespeedUrls, "https://shop.example.com/products")
}

func TestGetAccountExtensions_WithData(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	projectID := env.SeedProject(t, "org-1", "Test Project")
	shopID := env.SeedShop(t, "org-1", projectID, "My Shop", "https://shop.example.com")

	// Seed extensions
	_, err := env.Pool.Exec(t.Context(), `
		INSERT INTO shop_extension (shop_id, name, label, active, version, latest_version, installed)
		VALUES ($1, 'SwagPayPal', 'PayPal', true, '5.0.0', '5.1.0', true),
		       ($1, 'SwagCmsExtensions', 'CMS Extensions', true, '3.2.0', '3.2.0', true)
	`, shopID)
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/account/extensions", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var extensions []api.AccountExtension
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&extensions))
	require.Len(t, extensions, 2)

	// Find the PayPal extension
	var paypal *api.AccountExtension
	for i := range extensions {
		if extensions[i].Name == "SwagPayPal" {
			paypal = &extensions[i]
			break
		}
	}
	require.NotNil(t, paypal, "SwagPayPal extension should be present")
	assert.Equal(t, "PayPal", paypal.Label)
	assert.Equal(t, "5.0.0", paypal.Version)
	assert.Equal(t, "5.1.0", paypal.LatestVersion)
	assert.True(t, paypal.Active)
	assert.True(t, paypal.Installed)
	require.Len(t, paypal.Shops, 1)
	assert.Equal(t, shopID, paypal.Shops[0].ShopId)
	assert.Equal(t, "My Shop", paypal.Shops[0].ShopName)
}
