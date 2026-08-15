package advisory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Advisories published on a repository are never mirrored into the global
// advisory database, so the global endpoint 404s for them. They are also absent
// from the cve_id and affects indexes, which leaves the advisory link as the
// only route to the data.
func TestFetchAdvisoryFallsBackToRepositoryEndpoint(t *testing.T) {
	t.Parallel()

	var globalHits, repoHits int
	mux := http.NewServeMux()
	mux.HandleFunc("/advisories/GHSA-7m52-jw36-44r3", func(w http.ResponseWriter, _ *http.Request) {
		globalHits++
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	})
	// Verbatim shape of the repository-scoped response: patched_versions instead
	// of first_patched_version, and no references key at all.
	mux.HandleFunc("/repos/modelcontextprotocol/php-sdk/security-advisories/GHSA-7m52-jw36-44r3",
		func(w http.ResponseWriter, r *http.Request) {
			repoHits++
			assert.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ghsa_id":     "GHSA-7m52-jw36-44r3",
				"cve_id":      "CVE-2026-53965",
				"summary":     "DoS: client HttpTransport SSE buffer grows unbounded",
				"description": "## Summary\r\n\r\nThe HTTP client transport reads an SSE stream.",
				"html_url":    "https://github.com/modelcontextprotocol/php-sdk/security/advisories/GHSA-7m52-jw36-44r3",
				"cvss":        map[string]any{"score": nil, "vector_string": nil},
				"cvss_severities": map[string]any{
					"cvss_v3": map[string]any{"score": nil, "vector_string": nil},
				},
				"cwes": []map[string]string{
					{"cwe_id": "CWE-400", "name": "Uncontrolled Resource Consumption"},
				},
				"vulnerabilities": []map[string]any{
					{
						"package":                  map[string]string{"ecosystem": "composer", "name": "mcp/sdk"},
						"vulnerable_version_range": ">= 0.5.0, <= 0.7.0",
						"patched_versions":         "0.7.1",
					},
				},
			})
		})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newGitHubClient("tok", defaultUserAgent, server.Client())
	client.baseURL = server.URL

	link := "https://github.com/modelcontextprotocol/php-sdk/security/advisories/GHSA-7m52-jw36-44r3"
	details, err := client.fetchAdvisoryWithLink(context.Background(), "GHSA-7m52-jw36-44r3", link)
	require.NoError(t, err)
	require.NotNil(t, details)

	assert.Equal(t, 1, globalHits, "global endpoint should be tried first")
	assert.Equal(t, 1, repoHits, "repository endpoint should be the fallback")
	assert.Contains(t, details.Summary, "SSE buffer grows unbounded")
	assert.Contains(t, details.Description, "HTTP client transport")
	require.Len(t, details.CWEs, 1)
	assert.Equal(t, "CWE-400", details.CWEs[0].ID)
	// A null cvss score must not be recorded as 0.
	assert.Nil(t, details.CVSSScore)
	// The repo endpoint omits references; html_url still seeds the list.
	require.NotEmpty(t, details.References)
	assert.Equal(t, link, details.References[0])
}

// Without a repository link there is nothing to fall back to, so a 404 must
// surface rather than being silently swallowed.
func TestFetchAdvisoryWithoutLinkKeepsNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	}))
	t.Cleanup(server.Close)

	client := newGitHubClient("tok", defaultUserAgent, server.Client())
	client.baseURL = server.URL

	_, err := client.fetchAdvisoryWithLink(context.Background(), "GHSA-7m52-jw36-44r3", "")
	require.Error(t, err)
	assert.ErrorIs(t, err, errGitHubAdvisoryNotFound)
}

// A global advisory must not pay for an extra repository request.
func TestFetchAdvisorySkipsFallbackOnSuccess(t *testing.T) {
	t.Parallel()

	var repoHits int
	mux := http.NewServeMux()
	mux.HandleFunc("/advisories/GHSA-xvhc-gm7j-mhmc", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ghsa_id": "GHSA-xvhc-gm7j-mhmc",
			"summary": "Global advisory",
		})
	})
	mux.HandleFunc("/repos/", func(w http.ResponseWriter, _ *http.Request) {
		repoHits++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newGitHubClient("tok", defaultUserAgent, server.Client())
	client.baseURL = server.URL

	details, err := client.fetchAdvisoryWithLink(context.Background(), "GHSA-xvhc-gm7j-mhmc",
		"https://github.com/shopware/platform/security/advisories/GHSA-xvhc-gm7j-mhmc")
	require.NoError(t, err)
	assert.Equal(t, "Global advisory", details.Summary)
	assert.Zero(t, repoHits, "repository endpoint must not be called when the global one answers")
}

// The repository endpoint reports the patched release under a different key.
func TestFirstPatchedByLineAcceptsPatchedVersions(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"ghsa_id":"GHSA-repo","vulnerabilities":[
      {"package":{"name":"shopware/core","ecosystem":"composer"},"vulnerable_version_range":">= 6.7.0.0, <= 6.7.10.0","patched_versions":"6.7.10.1"},
      {"package":{"name":"mcp/sdk","ecosystem":"composer"},"patched_versions":"0.7.1"}]}`)
	var resp githubAdvisoryResponse
	require.NoError(t, json.Unmarshal(raw, &resp))

	got := firstPatchedByLine(resp)
	assert.Equal(t, "6.7.10.1", got["6.7"])
	assert.NotContains(t, got, "0.7", "non-shopware package must be ignored")
}

// The GHSA is often reachable only through the advisory link: FriendsOfPHP rows
// carry a remoteId like "mcp/sdk/CVE-2026-53965.yaml" and no GHSA in Sources.
func TestGHSAFromLink(t *testing.T) {
	t.Parallel()

	cases := []struct {
		link string
		want string
	}{
		{"https://github.com/modelcontextprotocol/php-sdk/security/advisories/GHSA-7m52-jw36-44r3", "GHSA-7m52-jw36-44r3"},
		{"https://github.com/advisories/GHSA-xvhc-gm7j-mhmc", "GHSA-xvhc-gm7j-mhmc"},
		{"https://example.com/CVE-2026-53965", ""},
		{"", ""},
		{"https://github.com/owner/repo/security/advisories/not-a-ghsa", ""},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, ghsaFromLink(tc.link), "link %q", tc.link)
	}
}

// owner/repo is interpolated into a token-bearing URL, so anything that could
// reshape the path must not survive parsing.
func TestRepoFromAdvisoryLink(t *testing.T) {
	t.Parallel()

	owner, repo := repoFromAdvisoryLink("https://github.com/modelcontextprotocol/php-sdk/security/advisories/GHSA-7m52-jw36-44r3")
	assert.Equal(t, "modelcontextprotocol", owner)
	assert.Equal(t, "php-sdk", repo)

	// Global advisory links carry no repository.
	owner, repo = repoFromAdvisoryLink("https://github.com/advisories/GHSA-xvhc-gm7j-mhmc")
	assert.Empty(t, owner)
	assert.Empty(t, repo)

	for _, link := range []string{
		"https://evil.com/owner/repo/security/advisories/GHSA-aaaa-bbbb-cccc",
		"https://github.com/owner/repo/security/advisories/GHSA-aaaa-bbbb-cccc?x=1",
		"https://github.com/../../x/security/advisories/GHSA-aaaa-bbbb-cccc",
		"http://github.com/owner/repo/security/advisories/GHSA-aaaa-bbbb-cccc",
	} {
		owner, repo = repoFromAdvisoryLink(link)
		assert.Empty(t, owner, "link %q must not yield an owner", link)
		assert.Empty(t, repo, "link %q must not yield a repo", link)
	}
}
