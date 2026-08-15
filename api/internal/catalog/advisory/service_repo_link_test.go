package advisory_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	catalogadvisory "github.com/friendsofshopware/shopmon/api/internal/catalog/advisory"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// canonicalAdvisoryID prefers the CVE, so two advisories carrying different
// CVEs but the same GHSA become two rows that the enrichment query groups back
// together — each keeping its own disclosure link. When those links disagree,
// the query must hand the GitHub pass the repository-scoped one: it is the only
// shape that names a repository to fall back to. Picking by sort order chooses
// the global github.com/advisories link, which 404s for a repository advisory
// and names no repository, stranding both rows unenriched.
func TestSyncPrefersRepositoryScopedLinkAmongSiblings(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()

	const (
		ghsa      = "GHSA-7m52-jw36-44r3"
		globalURL = "https://github.com/advisories/" + ghsa
		repoURL   = "https://github.com/modelcontextprotocol/php-sdk/security/advisories/" + ghsa
	)

	var repoHits int
	mux := http.NewServeMux()
	mux.HandleFunc("/packages.json", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"metadata-url":        "/p2/%package%.json",
			"security-advisories": map[string]any{"api-url": "/api/security-advisories/"},
		})
	})
	mux.HandleFunc("/packages/list.json", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"packageNames": []string{"shopware/core"}})
	})
	// Two CVEs sharing one GHSA become two rows, each with its own link. The
	// global link sorts before the repository one.
	mux.HandleFunc("/api/security-advisories/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			require.NoError(t, r.ParseForm())
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"advisories": map[string]any{
				"shopware/core": []map[string]any{
					{
						"advisoryId": "PKSA-a", "packageName": "shopware/core",
						"title": "SSE buffer DoS", "cve": "CVE-2026-0100",
						"affectedVersions": ">=6.7.0.0,<6.7.10.1",
						"reportedAt":       "2026-06-04 19:36:07", "link": globalURL,
						"sources": []map[string]string{{"name": "GitHub", "remoteId": ghsa}},
					},
					{
						"advisoryId": "PKSA-b", "packageName": "shopware/core",
						"title": "SSE buffer DoS", "cve": "CVE-2026-0101",
						"affectedVersions": ">=6.7.0.0,<6.7.10.1",
						"reportedAt":       "2026-06-04 19:36:07", "link": repoURL,
						"sources": []map[string]string{{"name": "GitHub", "remoteId": ghsa}},
					},
				},
			},
		})
	})
	// Repository-only advisory: absent from the global database.
	mux.HandleFunc("/advisories/"+ghsa, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"message":"Not Found"}`)
	})
	mux.HandleFunc("/repos/modelcontextprotocol/php-sdk/security-advisories/"+ghsa,
		func(w http.ResponseWriter, _ *http.Request) {
			repoHits++
			_, _ = io.WriteString(w, `{"ghsa_id":"`+ghsa+`","summary":"SSE buffer DoS",
              "description":"Unbounded buffer growth.","vulnerabilities":[
              {"package":{"name":"shopware/core","ecosystem":"composer"},
               "vulnerable_version_range":">= 6.7.0.0, <= 6.7.10.0","patched_versions":"6.7.10.1"}]}`)
		})
	mux.HandleFunc("/rest/json/cves/2.0", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"vulnerabilities": []map[string]any{}})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.NotFound(w, r)
			return
		}
		_, _ = io.WriteString(w, "{}")
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	svc := catalogadvisory.NewServiceWithClient(env.Pool, env.Queries, server.URL, server.Client())
	require.NoError(t, svc.Sync(ctx))

	assert.Positive(t, repoHits, "the repository endpoint must be reached via the repo-scoped sibling link")

	// Both rows share the GHSA, so the single fetch must enrich each of them.
	for _, id := range []string{"CVE-2026-0100", "CVE-2026-0101"} {
		row, err := env.Queries.GetComposerAdvisory(ctx, id)
		require.NoError(t, err, "advisory %s", id)

		require.NotNil(t, row.Description, "advisory %s has no description", id)
		assert.Contains(t, *row.Description, "Unbounded buffer growth")

		var patched map[string]string
		require.NoError(t, json.Unmarshal(row.FirstPatchedVersions, &patched))
		assert.Equal(t, "6.7.10.1", patched["6.7"],
			"patched versions must survive a divergent-link group (%s)", id)
	}
}
