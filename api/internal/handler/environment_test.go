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

func TestGetOrganizationEnvironments(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	env.SeedEnvironment(t, "org-1", shopID, "Environment A", "https://a.example.com")
	env.SeedEnvironment(t, "org-1", shopID, "Environment B", "https://b.example.com")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/organizations/org-1/environments", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var environments []json.RawMessage
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&environments))
	assert.Len(t, environments, 2)
}

func TestGetEnvironment(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var environment api.EnvironmentDetail
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&environment))
	assert.Equal(t, "My Environment", environment.Name)
	assert.Equal(t, "https://env.example.com", environment.Url)
	assert.Equal(t, "6.5.0.0", environment.ShopwareVersion)
	assert.Equal(t, "green", environment.Status)
	assert.Equal(t, "Test Org", environment.OrganizationName)
	assert.NotNil(t, environment.Extensions)
	assert.NotNil(t, environment.ScheduledTasks)
	assert.NotNil(t, environment.Queues)
	assert.NotNil(t, environment.Checks)
}

func TestGetEnvironment_NotMember(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	// Create another user's org and environment
	env.SeedUser(t, "user-2", "Other User", "other@example.com", "user")
	env.SeedOrganization(t, "org-2", "Other Org", "other-org", "user-2")
	shopID := env.SeedShop(t, "org-2", "Other Shop")
	environmentID := env.SeedEnvironment(t, "org-2", shopID, "Other Environment", "https://other.example.com")

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestDeleteEnvironment(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "To Delete", "https://del.example.com")

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify it's gone
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Environment is deleted, so org membership check on a non-existent environment returns not found
	assert.True(t, resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden)
}

func TestGetEnvironmentSubscription(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	// Initially not subscribed
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/environments/%d/subscribe", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]bool
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.False(t, result["subscribed"])
}

func TestUpdateEnvironment(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "Old Name", "https://old.example.com")

	newName := "New Name"
	body, _ := json.Marshal(api.UpdateEnvironmentRequest{
		Name:   &newName,
		ShopId: shopID,
	})

	req, _ := http.NewRequest("PATCH", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify the name changed by fetching the environment
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var environment api.EnvironmentDetail
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&environment))
	assert.Equal(t, "New Name", environment.Name)
}

func TestSubscribeAndUnsubscribeEnvironment(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	// Initially not subscribed
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/environments/%d/subscribe", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var result map[string]bool
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.False(t, result["subscribed"])

	// Subscribe
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/api/environments/%d/subscribe", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify subscribed
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/environments/%d/subscribe", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.True(t, result["subscribed"])

	// Unsubscribe
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("%s/api/environments/%d/subscribe", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify unsubscribed
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/environments/%d/subscribe", env.Server.URL, environmentID), nil)
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
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	urls := []string{"https://env.example.com/", "https://env.example.com/products"}
	body, _ := json.Marshal(api.SitespeedSettingsRequest{
		Enabled: true,
		Urls:    &urls,
	})

	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/api/environments/%d/sitespeed-settings", env.Server.URL, environmentID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify settings by fetching the environment
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var environment api.EnvironmentDetail
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&environment))
	assert.True(t, environment.SitespeedEnabled)
	require.NotNil(t, environment.SitespeedUrls)
	assert.Len(t, *environment.SitespeedUrls, 2)
	assert.Contains(t, *environment.SitespeedUrls, "https://env.example.com/")
	assert.Contains(t, *environment.SitespeedUrls, "https://env.example.com/products")
}

func TestGetAccountExtensions_WithData(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	// Seed extensions
	_, err := env.Pool.Exec(t.Context(), `
		INSERT INTO environment_extension (environment_id, name, label, active, version, latest_version, installed)
		VALUES ($1, 'SwagPayPal', 'PayPal', true, '5.0.0', '5.1.0', true),
		       ($1, 'SwagCmsExtensions', 'CMS Extensions', true, '3.2.0', '3.2.0', true)
	`, environmentID)
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
	require.Len(t, paypal.Environments, 1)
	assert.Equal(t, environmentID, paypal.Environments[0].EnvironmentId)
	assert.Equal(t, "My Environment", paypal.Environments[0].EnvironmentName)
}
