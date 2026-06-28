package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListSecurityAdvisories(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "u1", "User One", "user1@example.com", "user")

	ctx := context.Background()
	// Insert two advisories with different reported_at to assert ordering
	// (most recent first, NULLs last).
	_, err := env.Pool.Exec(ctx, `
		INSERT INTO security_advisory
			(advisory_id, origin, package_name, title, link, cve, affected_versions, source_name, severity, reported_at)
		VALUES
			('PKSA-old', 'packagist', 'shopware/core', 'Old advisory', 'https://example.com/old', 'CVE-2020-0001', '<6.4.0.0', 'FriendsOfPHP/security-advisories', 'medium', '2020-01-01 12:00:00'),
			('PKSA-new', 'packagist', 'shopware/storefront', 'New advisory', NULL, NULL, '>=6.5.0.0,<6.5.8.0', NULL, 'high', '2023-06-01 12:00:00')
	`)
	require.NoError(t, err)

	req := env.AuthRequest(t, http.MethodGet, "/api/info/security-advisories", token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var advisories []api.SecurityAdvisory
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&advisories))
	require.Len(t, advisories, 2)

	// Most recent first.
	assert.Equal(t, "PKSA-new", advisories[0].AdvisoryId)
	assert.Equal(t, "shopware/storefront", advisories[0].PackageName)
	assert.Equal(t, "high", *advisories[0].Severity)
	assert.Nil(t, advisories[0].Cve)
	assert.Nil(t, advisories[0].Link)

	assert.Equal(t, "PKSA-old", advisories[1].AdvisoryId)
	assert.Equal(t, "CVE-2020-0001", *advisories[1].Cve)
	assert.Equal(t, "https://example.com/old", *advisories[1].Link)
	require.NotNil(t, advisories[1].ReportedAt)
}

func TestListSecurityAdvisories_RequiresAuth(t *testing.T) {
	env := testutil.Setup(t)

	resp, err := testutil.Get(t, env.Server.URL+"/api/info/security-advisories")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
