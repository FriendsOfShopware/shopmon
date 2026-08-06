package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/crypto"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// encryptEnvironmentSecret replaces an environment's stored client_secret with an
// AES-GCM encrypted value so that the monitoring Shopware gateway can decrypt it.
// SeedEnvironment stores a plaintext secret, which the Shopware-calling handlers
// cannot decrypt, so tests that actually reach Shopware must call this.
func encryptEnvironmentSecret(t *testing.T, env *testutil.TestEnv, environmentID int, secret string) {
	t.Helper()
	encrypted, err := crypto.Encrypt(secret, env.Cfg.AppSecret)
	require.NoError(t, err)
	_, err = env.Pool.Exec(t.Context(),
		`UPDATE environment SET client_secret = $1 WHERE id = $2`, encrypted, environmentID)
	require.NoError(t, err)
}

func TestGetOrganizationEnvironments(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	env.SeedEnvironment(t, "org-1", shopID, "Environment A", "https://a.example.com")
	env.SeedEnvironment(t, "org-1", shopID, "Environment B", "https://b.example.com")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/organizations/org-1/environments", nil)
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
	// The detail payload reports only the changelog count; the entries themselves
	// are served paginated by GetEnvironmentChangelogs.
	seedChangelogs(t, env, environmentID, 12)

	req := testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var environment api.EnvironmentDetail
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&environment))
	assert.Equal(t, 12, environment.ChangelogsCount)
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

	req := testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// seedChangelogs inserts count changelog entries for an environment, oldest
// first, so that the newest entry is the last one inserted.
func seedChangelogs(t *testing.T, env *testutil.TestEnv, environmentID, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		extensions := fmt.Sprintf(
			`[{"name":"SwagPayPal","label":"PayPal","state":"updated","oldVersion":"1.%d.0","newVersion":"1.%d.0","active":true}]`,
			i, i+1,
		)
		_, err := env.Pool.Exec(t.Context(),
			`INSERT INTO environment_changelog (environment_id, extensions, date)
			 VALUES ($1, $2, NOW() - make_interval(hours => $3))`,
			environmentID, extensions, count-i,
		)
		require.NoError(t, err)
	}
}

func TestGetEnvironmentChangelogs_Paginates(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")
	seedChangelogs(t, env, environmentID, 25)

	fetch := func(query string) api.EnvironmentChangelogsResponse {
		t.Helper()
		req := testutil.NewRequest(t, "GET",
			fmt.Sprintf("%s/api/environments/%d/changelogs%s", env.Server.URL, environmentID, query), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result api.EnvironmentChangelogsResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
		return result
	}

	// The total always reports the full history, independent of the page size.
	first := fetch("?limit=10&offset=0")
	assert.Equal(t, 25, first.Total)
	require.Len(t, first.Entries, 10)

	second := fetch("?limit=10&offset=10")
	assert.Equal(t, 25, second.Total)
	require.Len(t, second.Entries, 10)

	last := fetch("?limit=10&offset=20")
	assert.Equal(t, 25, last.Total)
	assert.Len(t, last.Entries, 5)

	// Entries are newest first and pages must not overlap.
	assert.True(t, first.Entries[0].Date.After(first.Entries[9].Date))
	assert.True(t, first.Entries[9].Date.After(second.Entries[0].Date))

	seen := make(map[int]bool)
	for _, page := range []api.EnvironmentChangelogsResponse{first, second, last} {
		for _, entry := range page.Entries {
			require.False(t, seen[entry.Id], "entry %d returned on more than one page", entry.Id)
			seen[entry.Id] = true
			assert.Equal(t, environmentID, entry.EnvironmentId)
			assert.Equal(t, "My Environment", entry.EnvironmentName)
			assert.Len(t, entry.Extensions, 1)
		}
	}
	assert.Len(t, seen, 25)

	// Omitting the query parameters falls back to the first page.
	assert.Len(t, fetch("").Entries, 10)

	// An offset past the end yields no entries but still reports the total.
	beyond := fetch("?limit=10&offset=25")
	assert.Equal(t, 25, beyond.Total)
	assert.Empty(t, beyond.Entries)
}

// Inserts share a single NOW() per transaction, so entries with identical dates
// are normal. Ordering must still be total or offset pages overlap/skip entries.
func TestGetEnvironmentChangelogs_IdenticalDatesPaginateWithoutOverlap(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	for i := 0; i < 15; i++ {
		_, err := env.Pool.Exec(t.Context(),
			`INSERT INTO environment_changelog (environment_id, extensions, date)
			 VALUES ($1, '[]'::jsonb, '2026-01-01 12:00:00')`, environmentID)
		require.NoError(t, err)
	}

	fetchIDs := func(limit, offset int) []int {
		t.Helper()
		req := testutil.NewRequest(t, "GET",
			fmt.Sprintf("%s/api/environments/%d/changelogs?limit=%d&offset=%d", env.Server.URL, environmentID, limit, offset), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var page api.EnvironmentChangelogsResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&page))
		ids := make([]int, 0, len(page.Entries))
		for _, entry := range page.Entries {
			ids = append(ids, entry.Id)
		}
		return ids
	}

	// With every date equal, a total ordering is what makes paging coherent: the
	// pages must concatenate into exactly the same sequence as one unpaged read.
	unpaged := fetchIDs(15, 0)
	require.Len(t, unpaged, 15)
	assert.IsDecreasing(t, unpaged, "entries sharing a date must fall back to a deterministic id order")

	var paged []int
	for offset := 0; offset < 15; offset += 5 {
		page := fetchIDs(5, offset)
		require.Len(t, page, 5)
		paged = append(paged, page...)
	}
	assert.Equal(t, unpaged, paged, "paging must not overlap, skip, or reorder entries")
}

func TestGetEnvironmentChangelogs_RejectsOutOfRangePagination(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	for _, query := range []string{
		"?limit=0",
		"?limit=-1",
		"?limit=101",
		"?limit=2147483648",
		"?offset=-1",
	} {
		req := testutil.NewRequest(t, "GET",
			fmt.Sprintf("%s/api/environments/%d/changelogs%s", env.Server.URL, environmentID, query), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "query %q must be rejected", query)
		_ = resp.Body.Close()
	}
}

func TestGetEnvironmentChangelogs_NotMember(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedUser(t, "user-2", "Other User", "other@example.com", "user")
	env.SeedOrganization(t, "org-2", "Other Org", "other-org", "user-2")
	shopID := env.SeedShop(t, "org-2", "Other Shop")
	environmentID := env.SeedEnvironment(t, "org-2", shopID, "Other Environment", "https://other.example.com")
	seedChangelogs(t, env, environmentID, 3)

	req := testutil.NewRequest(t, "GET",
		fmt.Sprintf("%s/api/environments/%d/changelogs", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestCreateEnvironment(t *testing.T) {
	mockShopware := testutil.NewMockShopwareServer(t)
	defer mockShopware.Close()

	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	body, _ := json.Marshal(api.CreateEnvironmentRequest{
		Name:         "Staging",
		ShopUrl:      mockShopware.URL,
		ClientId:     "client-id",
		ClientSecret: "client-secret",
		ShopId:       shopID,
	})

	req := testutil.NewRequest(t, http.MethodPost, env.Server.URL+"/api/environments", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result struct {
		ID int32 `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.NotZero(t, result.ID)

	var encryptedSecret, version string
	err = env.Pool.QueryRow(t.Context(), `SELECT client_secret, shopware_version FROM environment WHERE id = $1`, result.ID).Scan(&encryptedSecret, &version)
	require.NoError(t, err)
	decryptedSecret, err := crypto.Decrypt(encryptedSecret, env.Cfg.AppSecret)
	require.NoError(t, err)
	assert.Equal(t, "client-secret", decryptedSecret)
	assert.Equal(t, "6.5.0.0", version)
}

func TestCreateEnvironmentWithToken(t *testing.T) {
	mockShopware := testutil.NewMockShopwareServer(t)
	defer mockShopware.Close()

	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	providedToken := "aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899"
	body, _ := json.Marshal(api.CreateEnvironmentRequest{
		Name:             "Staging",
		ShopUrl:          mockShopware.URL,
		ClientId:         "client-id",
		ClientSecret:     "client-secret",
		ShopId:           shopID,
		EnvironmentToken: &providedToken,
	})

	req := testutil.NewRequest(t, http.MethodPost, env.Server.URL+"/api/environments", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result struct {
		ID int32 `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

	var environmentToken string
	err = env.Pool.QueryRow(t.Context(), `SELECT environment_token FROM environment WHERE id = $1`, result.ID).Scan(&environmentToken)
	require.NoError(t, err)
	assert.Equal(t, providedToken, environmentToken)

	// Token must appear on environment detail for the bypass-auth header UI.
	detailReq := testutil.NewRequest(t, http.MethodGet, fmt.Sprintf("%s/api/environments/%d", env.Server.URL, result.ID), nil)
	detailReq.Header.Set("Authorization", "Bearer "+token)
	detailResp, err := http.DefaultClient.Do(detailReq)
	require.NoError(t, err)
	defer func() { _ = detailResp.Body.Close() }()
	require.Equal(t, http.StatusOK, detailResp.StatusCode)

	var detail api.EnvironmentDetail
	require.NoError(t, json.NewDecoder(detailResp.Body).Decode(&detail))
	assert.Equal(t, providedToken, detail.EnvironmentToken)
}

func TestDeleteEnvironment(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "To Delete", "https://del.example.com")

	req := testutil.NewRequest(t, "DELETE", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify it's gone
	req = testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Environment is deleted, so org membership check on a non-existent environment returns not found
	assert.True(t, resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden)
}

// TestDeleteDefaultEnvironment_Reassigns covers the case from issue #702: deleting an
// environment that is its shop's default must not fail on the RESTRICT foreign key. The
// shop's default is moved to another environment of the same shop.
func TestDeleteDefaultEnvironment_Reassigns(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	// First environment becomes the shop default.
	defaultEnvID := env.SeedEnvironment(t, "org-1", shopID, "Default", "https://default.example.com")
	otherEnvID := env.SeedEnvironment(t, "org-1", shopID, "Other", "https://other.example.com")

	var currentDefault int
	require.NoError(t, env.Pool.QueryRow(t.Context(),
		`SELECT default_environment_id FROM shop WHERE id = $1`, shopID).Scan(&currentDefault))
	require.Equal(t, defaultEnvID, currentDefault)

	req := testutil.NewRequest(t, "DELETE", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, defaultEnvID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// The shop default is reassigned to the remaining environment.
	require.NoError(t, env.Pool.QueryRow(t.Context(),
		`SELECT default_environment_id FROM shop WHERE id = $1`, shopID).Scan(&currentDefault))
	assert.Equal(t, otherEnvID, currentDefault)
}

// TestDeleteLastEnvironment_ClearsDefault covers deleting the only environment of a shop:
// the shop's default is cleared to NULL instead of violating the RESTRICT foreign key.
func TestDeleteLastEnvironment_ClearsDefault(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "Only", "https://only.example.com")

	req := testutil.NewRequest(t, "DELETE", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	var defaultEnvID *int
	require.NoError(t, env.Pool.QueryRow(t.Context(),
		`SELECT default_environment_id FROM shop WHERE id = $1`, shopID).Scan(&defaultEnvID))
	assert.Nil(t, defaultEnvID)

	// The API must expose the missing default as null rather than coercing it to a
	// bogus environment id of 0, which would render broken links in the shop lists.
	shopsReq := testutil.NewRequest(t, "GET", env.Server.URL+"/api/account/shops", nil)
	shopsReq.Header.Set("Authorization", "Bearer "+token)
	shopsResp, err := http.DefaultClient.Do(shopsReq)
	require.NoError(t, err)
	defer func() { _ = shopsResp.Body.Close() }()

	var shops []api.AccountShop
	require.NoError(t, json.NewDecoder(shopsResp.Body).Decode(&shops))
	require.Len(t, shops, 1)
	assert.Nil(t, shops[0].DefaultEnvironmentId)
}

func TestGetEnvironmentSubscription(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	// Subscription status is exposed on the environment detail payload, and a
	// freshly seeded environment is not subscribed.
	assert.False(t, getEnvironmentDetail(t, env.Server.URL, token, environmentID).Subscribed)
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

	req := testutil.NewRequest(t, "PATCH", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify the name changed by fetching the environment
	req = testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
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

	// Subscription status is exposed on the environment detail payload, so the
	// frontend doesn't need a separate request. Initially not subscribed.
	assert.False(t, getEnvironmentDetail(t, env.Server.URL, token, environmentID).Subscribed)

	// Subscribe
	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/subscribe", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify subscribed
	assert.True(t, getEnvironmentDetail(t, env.Server.URL, token, environmentID).Subscribed)

	// Unsubscribe
	req = testutil.NewRequest(t, "DELETE", fmt.Sprintf("%s/api/environments/%d/subscribe", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify unsubscribed
	assert.False(t, getEnvironmentDetail(t, env.Server.URL, token, environmentID).Subscribed)
}

// getEnvironmentDetail fetches the full environment detail payload for the given
// environment as the authenticated user.
func getEnvironmentDetail(t *testing.T, serverURL, token string, environmentID int) api.EnvironmentDetail {
	t.Helper()

	req := testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d", serverURL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var environment api.EnvironmentDetail
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&environment))
	return environment
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

	req := testutil.NewRequest(t, "PUT", fmt.Sprintf("%s/api/environments/%d/sitespeed-settings", env.Server.URL, environmentID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify settings by fetching the environment
	req = testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
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

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/account/extensions", nil)
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

func TestRefreshEnvironment(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/refresh", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Only enqueues a job on the in-memory bus; does not call Shopware.
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestRefreshEnvironment_WithSitespeed(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	sitespeed := true
	body, _ := json.Marshal(api.RefreshEnvironmentJSONRequestBody{Sitespeed: &sitespeed})

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/refresh", env.Server.URL, environmentID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestRefreshEnvironment_NotMember(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedUser(t, "user-2", "Other User", "other@example.com", "user")
	env.SeedOrganization(t, "org-2", "Other Org", "other-org", "user-2")
	shopID := env.SeedShop(t, "org-2", "Other Shop")
	environmentID := env.SeedEnvironment(t, "org-2", shopID, "Other Environment", "https://other.example.com")

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/refresh", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRefreshEnvironment_Unauthenticated(t *testing.T) {
	env := testutil.Setup(t)
	env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/refresh", env.Server.URL, environmentID), nil)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestClearEnvironmentCache(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"mock-token","expires_in":600}`))
	})
	var cacheCleared bool
	mux.HandleFunc("/api/_action/cache", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		cacheCleared = true
		w.WriteHeader(http.StatusNoContent)
	})
	shopware := httptest.NewServer(mux)
	defer shopware.Close()

	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", shopware.URL)
	encryptEnvironmentSecret(t, env, environmentID, "test-secret")

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/clear-cache", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.True(t, cacheCleared, "expected DELETE /_action/cache to be called on Shopware")
}

func TestClearEnvironmentCache_ShopwareError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"mock-token","expires_in":600}`))
	})
	mux.HandleFunc("/api/_action/cache", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	shopware := httptest.NewServer(mux)
	defer shopware.Close()

	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", shopware.URL)
	encryptEnvironmentSecret(t, env, environmentID, "test-secret")

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/clear-cache", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Shopware failure is surfaced as a bad gateway.
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

func TestClearEnvironmentCache_NotMember(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedUser(t, "user-2", "Other User", "other@example.com", "user")
	env.SeedOrganization(t, "org-2", "Other Org", "other-org", "user-2")
	shopID := env.SeedShop(t, "org-2", "Other Shop")
	environmentID := env.SeedEnvironment(t, "org-2", shopID, "Other Environment", "https://other.example.com")

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/clear-cache", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRescheduleTask(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"mock-token","expires_in":600}`))
	})
	var patched, runTriggered bool
	mux.HandleFunc("/api/scheduled-task/task-123", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		patched = true
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/_action/scheduled-task/run", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		runTriggered = true
		w.WriteHeader(http.StatusOK)
	})
	shopware := httptest.NewServer(mux)
	defer shopware.Close()

	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", shopware.URL)
	encryptEnvironmentSecret(t, env, environmentID, "test-secret")

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/tasks/task-123/reschedule", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.True(t, patched, "expected PATCH /scheduled-task/task-123 to be called")
	assert.True(t, runTriggered, "expected POST /_action/scheduled-task/run to be called")
}

func TestRescheduleTask_ShopwareError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"mock-token","expires_in":600}`))
	})
	mux.HandleFunc("/api/scheduled-task/task-123", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	shopware := httptest.NewServer(mux)
	defer shopware.Close()

	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", shopware.URL)
	encryptEnvironmentSecret(t, env, environmentID, "test-secret")

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/tasks/task-123/reschedule", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

func TestRescheduleTask_NotMember(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedUser(t, "user-2", "Other User", "other@example.com", "user")
	env.SeedOrganization(t, "org-2", "Other Org", "other-org", "user-2")
	shopID := env.SeedShop(t, "org-2", "Other Shop")
	environmentID := env.SeedEnvironment(t, "org-2", shopID, "Other Environment", "https://other.example.com")

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/tasks/task-123/reschedule", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
