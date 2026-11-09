package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComposerAdvisoryFeed(t *testing.T) {
	env := testutil.Setup(t)
	const pkg = "store.shopware.com/example"
	seedAdvisory(t, env.Queries, "CVE-PUBLIC", []string{pkg, "shopware/core", "store.shopware.com.evil/example"}, true, new("secret-internal"))
	seedAdvisory(t, env.Queries, "CVE-HIDDEN", []string{pkg}, false, nil)
	seedAdvisory(t, env.Queries, "SHOPMON-2026-0001", []string{pkg, "store.shopware.com/second"}, true, nil)
	_, err := env.Pool.Exec(context.Background(), `
		UPDATE composer_advisory SET sources = '[]', reported_at = NULL, severity_override = 'critical'
		WHERE advisory_id = 'SHOPMON-2026-0001'`)
	require.NoError(t, err)
	_, err = env.Pool.Exec(context.Background(), `
		UPDATE composer_advisory_package SET packagist_advisory_id = NULL
		WHERE advisory_id = 'SHOPMON-2026-0001'`)
	require.NoError(t, err)
	// Exceed the normal advisory listing's maximum page size.
	for i := range 101 {
		seedAdvisory(t, env.Queries, fmt.Sprintf("CVE-BULK-%03d", i), []string{pkg}, true, nil)
	}
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		t.Run(method, func(t *testing.T) {
			req, err := http.NewRequestWithContext(t.Context(), method, env.Server.URL+"/api/advisories/composer", strings.NewReader("packages%5B%5D=store.shopware.com%2Fexample"))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()
			require.Equal(t, http.StatusOK, resp.StatusCode)
			raw, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.NotContains(t, string(raw), "secret-internal")
			assert.NotContains(t, string(raw), "CVE-HIDDEN")
			var body api.ComposerAdvisoryResponse
			require.NoError(t, json.Unmarshal(raw, &body))
			if method == http.MethodGet {
				require.Len(t, body.Advisories, 2)
			} else {
				require.Len(t, body.Advisories, 1)
			}
			require.Len(t, body.Advisories[pkg], 103)
			upstream := body.Advisories[pkg][101]
			assert.Equal(t, "PKSA-CVE-PUBLIC-0", upstream.AdvisoryID)
			assert.Equal(t, pkg, upstream.PackageName)
			assert.Equal(t, ">=6.7.0.0,<6.7.10.1", upstream.AffectedVersions)
			assert.Equal(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC), upstream.ReportedAt)
			manual := body.Advisories[pkg][102]
			assert.Equal(t, "SHOPMON-2026-0001", manual.AdvisoryID)
			assert.Equal(t, new("critical"), manual.Severity)
			assert.False(t, manual.ReportedAt.IsZero())
			assert.Equal(t, []api.AdvisorySource{{Name: "Shopmon", RemoteId: manual.AdvisoryID}}, manual.Sources)
		})
	}
}

func TestComposerAdvisoryFeedEmpty(t *testing.T) {
	env := testutil.Setup(t)
	seedAdvisory(t, env.Queries, "CVE-HIDDEN", []string{"store.shopware.com/example"}, false, nil)
	req := testutil.NewRequest(t, http.MethodGet, env.Server.URL+"/api/advisories/composer", nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var body map[string]json.RawMessage
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.JSONEq(t, `{}`, string(body["advisories"]))
}

func TestComposerAdvisoryPackageFilters(t *testing.T) {
	env := testutil.Setup(t)
	const first = "store.shopware.com/first"
	const second = "store.shopware.com/second"
	seedAdvisory(t, env.Queries, "CVE-MULTI", []string{first, second}, true, nil)
	seedAdvisory(t, env.Queries, "CVE-UNRELATED", []string{"store.shopware.com/unrelated"}, true, nil)
	for _, tt := range []struct {
		name string
		form string
		want []string
	}{
		{"brackets", "packages%5B%5D=store.shopware.com%2Ffirst", []string{first}},
		{"indexed", "packages%5B0%5D=store.shopware.com%2Ffirst", []string{first}},
		{"multiple brackets", "packages[]=store.shopware.com/first&packages[]=store.shopware.com/second", []string{first, second}},
		{"multiple indexed", "packages[0]=store.shopware.com/first&packages[1]=store.shopware.com/second", []string{first, second}},
		{"duplicates", "packages[]=store.shopware.com/first&packages[]=store.shopware.com/first", []string{first}},
		{"uppercase", "packages[]=STORE.SHOPWARE.COM/FIRST", []string{first}},
		{"unknown", "packages[]=store.shopware.com/unknown", nil},
		{"empty", "", nil},
		{"empty name", "packages[]=", nil},
		{"unrelated field", "other[]=store.shopware.com/first", nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := testutil.NewRequest(t, http.MethodPost, env.Server.URL+"/api/advisories/composer", strings.NewReader(tt.form))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()
			require.Equal(t, http.StatusOK, resp.StatusCode)
			var body api.ComposerAdvisoryResponse
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
			require.NotNil(t, body.Advisories)
			require.Len(t, body.Advisories, len(tt.want))
			for _, name := range tt.want {
				require.Len(t, body.Advisories[name], 1)
				assert.Equal(t, name, body.Advisories[name][0].PackageName)
			}
		})
	}
	t.Run("malformed form", func(t *testing.T) {
		req := testutil.NewRequest(t, http.MethodPost, env.Server.URL+"/api/advisories/composer", strings.NewReader("packages[]=%zz"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
