package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedStoreExtension inserts a catalog row plus its per-environment link so the
// extension shows up in the account extension list.
func seedStoreExtension(t *testing.T, env *testutil.TestEnv, name string, environmentID int) {
	t.Helper()
	ctx := context.Background()
	_, err := env.Pool.Exec(ctx, `
		INSERT INTO store_extension (name, store_link, latest_version)
		VALUES ($1, $2, '1.0.0')
	`, name, "https://store.shopware.com/"+name)
	require.NoError(t, err)

	_, err = env.Pool.Exec(ctx, `
		INSERT INTO store_extension_translation (extension_name, language, label)
		VALUES ($1, 'en', $2)
	`, name, name)
	require.NoError(t, err)

	_, err = env.Pool.Exec(ctx, `
		INSERT INTO environment_store_extension (environment_id, extension_name, label, version, latest_version, active, installed)
		VALUES ($1, $2, $3, '1.0.0', '1.0.0', true, true)
	`, environmentID, name, name)
	require.NoError(t, err)
}

func postJSON(t *testing.T, env *testutil.TestEnv, path, token string, body any) *http.Response {
	t.Helper()
	buf, err := json.Marshal(body)
	require.NoError(t, err)
	req := testutil.NewRequest(t, http.MethodPost, env.Server.URL+path, bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func TestReportExtension_Unauthenticated(t *testing.T) {
	env := testutil.Setup(t)
	resp := postJSON(t, env, "/api/account/extensions/Foo/report", "", api.ExtensionReportRequest{
		Category: api.Performance,
		Comment:  "slow",
	})
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestReportExtension_UnknownExtension(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")

	resp := postJSON(t, env, "/api/account/extensions/DoesNotExist/report", token, api.ExtensionReportRequest{
		Category: api.Performance,
		Comment:  "slow",
	})
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestReportExtension_InvalidCategory(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	envID := env.SeedEnvironment(t, "org-1", shopID, "Prod", "https://example.com")
	seedStoreExtension(t, env, "MyPlugin", envID)

	resp := postJSON(t, env, "/api/account/extensions/MyPlugin/report", token, map[string]string{
		"category": "bogus",
		"comment":  "x",
	})
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestReportExtension_EmptyComment(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	envID := env.SeedEnvironment(t, "org-1", shopID, "Prod", "https://example.com")
	seedStoreExtension(t, env, "MyPlugin", envID)

	resp := postJSON(t, env, "/api/account/extensions/MyPlugin/report", token, api.ExtensionReportRequest{
		Category: api.Performance,
		Comment:  "   ",
	})
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

// TestReportExtension_FullFlow covers submit -> admin moderation -> the approved
// report surfacing as a community warning on the account extension list.
func TestReportExtension_FullFlow(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	adminToken := env.SeedUser(t, "admin-1", "Admin", "admin@example.com", "admin")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	envID := env.SeedEnvironment(t, "org-1", shopID, "Prod", "https://example.com")
	seedStoreExtension(t, env, "MyPlugin", envID)

	// Submit a report.
	resp := postJSON(t, env, "/api/account/extensions/MyPlugin/report", token, api.ExtensionReportRequest{
		Category: api.Performance,
		Comment:  "Slows down the storefront",
	})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	_ = resp.Body.Close()

	// A second submission from the same user is a no-op (still 204) and must not
	// create a duplicate pending row.
	resp = postJSON(t, env, "/api/account/extensions/MyPlugin/report", token, api.ExtensionReportRequest{
		Category: api.Performance,
		Comment:  "again",
	})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	_ = resp.Body.Close()

	// Not yet approved: no community warning on the account list.
	require.Nil(t, accountExtensionReports(t, env, token, "MyPlugin"))

	// Non-admin cannot list reports.
	req := env.AuthRequest(t, http.MethodGet, "/api/admin/extension-reports", token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	_ = resp.Body.Close()

	// Admin lists pending reports.
	req = env.AuthRequest(t, http.MethodGet, "/api/admin/extension-reports?status=pending", adminToken)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var listResp api.AdminExtensionReportsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&listResp))
	_ = resp.Body.Close()
	require.Len(t, listResp.Reports, 1)
	reportID := listResp.Reports[0].Id
	assert.Equal(t, "MyPlugin", listResp.Reports[0].ExtensionName)
	assert.Equal(t, "performance", listResp.Reports[0].Category)
	assert.Equal(t, "pending", listResp.Reports[0].Status)

	// Approve it.
	resp = postJSON(t, env, fmt.Sprintf("/api/admin/extension-reports/%d/approve", reportID), adminToken, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	_ = resp.Body.Close()

	// Now the approved report surfaces as a community warning.
	reports := accountExtensionReports(t, env, token, "MyPlugin")
	require.Len(t, reports, 1)
	assert.Equal(t, "performance", reports[0].Category)
	assert.Equal(t, 1, reports[0].Count)

	// An audit log entry was recorded.
	var auditCount int
	require.NoError(t, env.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM audit_log WHERE action = 'extension_report.approved'`).Scan(&auditCount))
	assert.Equal(t, 1, auditCount)
}

func TestRejectExtensionReport(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	adminToken := env.SeedUser(t, "admin-1", "Admin", "admin@example.com", "admin")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	envID := env.SeedEnvironment(t, "org-1", shopID, "Prod", "https://example.com")
	seedStoreExtension(t, env, "MyPlugin", envID)

	resp := postJSON(t, env, "/api/account/extensions/MyPlugin/report", token, api.ExtensionReportRequest{
		Category: api.Security,
		Comment:  "leaks data",
	})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	_ = resp.Body.Close()

	var reportID int
	require.NoError(t, env.Pool.QueryRow(context.Background(),
		`SELECT id FROM store_extension_report WHERE extension_name = 'MyPlugin'`).Scan(&reportID))

	resp = postJSON(t, env, fmt.Sprintf("/api/admin/extension-reports/%d/reject", reportID), adminToken, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	_ = resp.Body.Close()

	// Rejected reports never become community warnings.
	require.Nil(t, accountExtensionReports(t, env, token, "MyPlugin"))

	var status string
	require.NoError(t, env.Pool.QueryRow(context.Background(),
		`SELECT status FROM store_extension_report WHERE id = $1`, reportID).Scan(&status))
	assert.Equal(t, "rejected", status)
}

// accountExtensionReports fetches the account extension list and returns the
// reports attached to the named extension (nil when absent).
func accountExtensionReports(t *testing.T, env *testutil.TestEnv, token, name string) []api.ExtensionReportSummary {
	t.Helper()
	req := env.AuthRequest(t, http.MethodGet, "/api/account/extensions", token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var extensions []api.AccountExtension
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&extensions))
	for _, ext := range extensions {
		if ext.Name == name && ext.Reports != nil {
			return *ext.Reports
		}
	}
	return nil
}
