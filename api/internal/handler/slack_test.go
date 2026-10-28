package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testWebhookURL = "https://hooks.slack.com/services/T000/B000/secret"

func slackRequest(t *testing.T, env *testutil.TestEnv, method, path, token string, body any) *http.Response {
	t.Helper()
	return slackRequestAt(t, env, method, "/api/organizations/org-1/slack-webhook"+path, token, body)
}

func slackRequestAt(t *testing.T, env *testutil.TestEnv, method, url, token string, body any) *http.Response {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := testutil.NewRequest(t, method, env.Server.URL+url, reader)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func slackStatusAt(t *testing.T, env *testutil.TestEnv, url, token string) api.SlackWebhookStatus {
	t.Helper()
	resp := slackRequestAt(t, env, http.MethodGet, url, token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var status api.SlackWebhookStatus
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&status))
	return status
}

func slackConfigured(t *testing.T, env *testutil.TestEnv, token string) bool {
	t.Helper()
	return slackStatusAt(t, env, "/api/organizations/org-1/slack-webhook", token).Configured
}

func TestSlackWebhookLifecycle(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")

	status := slackStatusAt(t, env, "/api/organizations/org-1/slack-webhook", token)
	assert.False(t, status.Configured)
	assert.True(t, status.CanManage)

	resp := slackRequest(t, env, http.MethodPut, "", token, api.UpdateSlackWebhookRequest{WebhookUrl: testWebhookURL})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.True(t, slackConfigured(t, env, token))

	// The status response never carries the URL.
	raw := slackRequest(t, env, http.MethodGet, "", token, nil)
	body, err := io.ReadAll(raw.Body)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "hooks.slack.com")

	// Stored encrypted, never in plain text.
	var stored string
	require.NoError(t, env.Pool.QueryRow(t.Context(),
		`SELECT webhook_url FROM organization_slack_webhook WHERE organization_id = 'org-1'`).Scan(&stored))
	assert.NotContains(t, stored, "hooks.slack.com")

	resp = slackRequest(t, env, http.MethodPost, "/test", token, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	require.Len(t, env.Slack.Posts, 1)
	assert.Equal(t, testWebhookURL, env.Slack.Posts[0].URL)

	resp = slackRequest(t, env, http.MethodDelete, "", token, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.False(t, slackConfigured(t, env, token))
}

func TestSlackWebhookRejectsForeignURL(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")

	resp := slackRequest(t, env, http.MethodPut, "", token, api.UpdateSlackWebhookRequest{WebhookUrl: "https://example.com/hook"})
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.False(t, slackConfigured(t, env, token))
}

func TestSlackWebhookTestWithoutWebhook(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")

	resp := slackRequest(t, env, http.MethodPost, "/test", token, nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSlackWebhookTestReportsSlackRejection(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")

	resp := slackRequest(t, env, http.MethodPut, "", token, api.UpdateSlackWebhookRequest{WebhookUrl: testWebhookURL})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	env.Slack.Err = errors.New("slack responded 404: no_service")
	resp = slackRequest(t, env, http.MethodPost, "/test", token, nil)
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

func TestSlackWebhookMemberCanReadButNotManage(t *testing.T) {
	env := testutil.Setup(t)
	env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	memberToken := env.SeedUser(t, "user-2", "Member", "member@example.com", "user")
	env.SeedMember(t, "org-1", "user-2", "member")

	status := slackStatusAt(t, env, "/api/organizations/org-1/slack-webhook", memberToken)
	assert.False(t, status.Configured)
	assert.False(t, status.CanManage)

	resp := slackRequest(t, env, http.MethodPut, "", memberToken, api.UpdateSlackWebhookRequest{WebhookUrl: testWebhookURL})
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp = slackRequest(t, env, http.MethodDelete, "", memberToken, nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp = slackRequest(t, env, http.MethodPost, "/test", memberToken, nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestSlackWebhookOutsiderIsForbidden(t *testing.T) {
	env := testutil.Setup(t)
	env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	outsiderToken := env.SeedUser(t, "user-3", "Outsider", "outsider@example.com", "user")

	resp := slackRequest(t, env, http.MethodGet, "", outsiderToken, nil)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestShopSlackWebhookLifecycle(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Shop")
	base := fmt.Sprintf("/api/organizations/org-1/shops/%d/slack-webhook", shopID)

	assert.False(t, slackStatusAt(t, env, base, token).Configured)

	resp := slackRequestAt(t, env, http.MethodPut, base, token, api.UpdateSlackWebhookRequest{WebhookUrl: testWebhookURL})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.True(t, slackStatusAt(t, env, base, token).Configured)
	// The shop webhook is separate from the organization one.
	assert.False(t, slackConfigured(t, env, token))

	resp = slackRequestAt(t, env, http.MethodPost, base+"/test", token, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	require.Len(t, env.Slack.Posts, 1)
	assert.Contains(t, env.Slack.Posts[0].Text, "this shop")

	resp = slackRequestAt(t, env, http.MethodDelete, base, token, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.False(t, slackStatusAt(t, env, base, token).Configured)
}

func TestShopSlackWebhookRejectsShopOfAnotherOrganization(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	env.SeedUser(t, "user-2", "Other", "other@example.com", "user")
	env.SeedOrganization(t, "org-2", "Other Org", "other-org", "user-2")
	foreignShop := env.SeedShop(t, "org-2", "Foreign")

	// Admin of org-1 addressing org-2's shop through org-1's path.
	url := fmt.Sprintf("/api/organizations/org-1/shops/%d/slack-webhook", foreignShop)
	resp := slackRequestAt(t, env, http.MethodPut, url, token, api.UpdateSlackWebhookRequest{WebhookUrl: testWebhookURL})
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp = slackRequestAt(t, env, http.MethodGet, url, token, nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp = slackRequestAt(t, env, http.MethodDelete, url, token, nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp = slackRequestAt(t, env, http.MethodPost, url+"/test", token, nil)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Empty(t, env.Slack.Posts)

	var count int
	require.NoError(t, env.Pool.QueryRow(t.Context(), `SELECT COUNT(*) FROM shop_slack_webhook`).Scan(&count))
	assert.Zero(t, count)
}

// Zero would address the organization webhook and values beyond int32 would
// wrap onto another shop, so both must be rejected before reaching storage.
func TestShopSlackWebhookRejectsOutOfRangeShopIDs(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	env.SeedShop(t, "org-1", "Shop")

	resp := slackRequest(t, env, http.MethodPut, "", token, api.UpdateSlackWebhookRequest{WebhookUrl: testWebhookURL})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	for _, shopID := range []string{"0", "4294967296", "4294967297"} {
		url := "/api/organizations/org-1/shops/" + shopID + "/slack-webhook"
		resp = slackRequestAt(t, env, http.MethodDelete, url, token, nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, shopID)
		resp = slackRequestAt(t, env, http.MethodGet, url, token, nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, shopID)
		resp = slackRequestAt(t, env, http.MethodPut, url, token, api.UpdateSlackWebhookRequest{WebhookUrl: testWebhookURL})
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, shopID)
	}

	// The organization webhook survived the DELETE on shop 0.
	assert.True(t, slackConfigured(t, env, token))
	var count int
	require.NoError(t, env.Pool.QueryRow(t.Context(), `SELECT COUNT(*) FROM shop_slack_webhook`).Scan(&count))
	assert.Zero(t, count)
}
