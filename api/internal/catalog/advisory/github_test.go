package advisory

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubFetchAdvisory(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/advisories/GHSA-test-0001", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/vnd.github+json", r.Header.Get("Accept"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ghsa_id":     "GHSA-test-0001",
			"summary":     "Shopware SSRF",
			"description": "## Summary\n\nBad endpoint allows SSRF.",
			"html_url":    "https://github.com/advisories/GHSA-test-0001",
			"cvss_severities": map[string]any{
				"cvss_v3": map[string]any{
					"score":         4.1,
					"vector_string": "CVSS:3.1/AV:N/AC:L/PR:H/UI:N/S:C/C:L/I:N/A:N",
				},
			},
			"cwes": []map[string]string{
				{"cwe_id": "CWE-918", "name": "Server-Side Request Forgery (SSRF)"},
			},
			"references": []map[string]string{
				{"url": "https://example.com/advisory"},
			},
		})
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newGitHubClient("tok", defaultUserAgent, server.Client())
	client.baseURL = server.URL

	details, err := client.fetchAdvisory(context.Background(), "GHSA-test-0001")
	require.NoError(t, err)
	require.NotNil(t, details)
	assert.Equal(t, "Shopware SSRF", details.Summary)
	assert.Contains(t, details.Description, "SSRF")
	require.NotNil(t, details.CVSSScore)
	assert.InDelta(t, 4.1, *details.CVSSScore, 0.01)
	require.Len(t, details.CWEs, 1)
	assert.Equal(t, "CWE-918", details.CWEs[0].ID)
	// html_url is prepended when not already listed
	require.GreaterOrEqual(t, len(details.References), 2)
	assert.Equal(t, "https://github.com/advisories/GHSA-test-0001", details.References[0])
}

func TestGitHubFetchAdvisoryStringReferences(t *testing.T) {
	t.Parallel()

	// Real GHSA responses often return references as plain URL strings.
	mux := http.NewServeMux()
	mux.HandleFunc("/advisories/GHSA-string-refs", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ghsa_id":     "GHSA-string-refs",
			"summary":     "String refs advisory",
			"description": "Body",
			"html_url":    "https://github.com/advisories/GHSA-string-refs",
			"references": []string{
				"https://github.com/shopware/platform/security/advisories/GHSA-string-refs",
				"https://example.com/other",
			},
		})
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newGitHubClient("", defaultUserAgent, server.Client())
	client.baseURL = server.URL

	details, err := client.fetchAdvisory(context.Background(), "GHSA-string-refs")
	require.NoError(t, err)
	require.NotNil(t, details)
	assert.Contains(t, details.References, "https://example.com/other")
	assert.Contains(t, details.References, "https://github.com/shopware/platform/security/advisories/GHSA-string-refs")
	assert.Equal(t, "https://github.com/advisories/GHSA-string-refs", details.References[0])
}

// GHSA ids come from Packagist source records — third-party data. A value that
// could reshape the request path must be rejected before it is interpolated
// into a URL carrying our GitHub token.
func TestFetchAdvisoryRejectsMalformedGHSAID(t *testing.T) {
	t.Parallel()

	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		_, _ = io.WriteString(w, "{}")
	}))
	t.Cleanup(server.Close)

	client := newGitHubClient("token", defaultUserAgent, server.Client())
	client.baseURL = server.URL

	for _, id := range []string{
		"GHSA-x/../../user/repo",
		"GHSA-abcd?ref=x",
		"GHSA-ab cd-efgh-ijkl",
		"not-a-ghsa",
		"GHSA-",
	} {
		if _, err := client.fetchAdvisory(context.Background(), id); err == nil {
			t.Errorf("id %q was accepted, want rejection", id)
		}
	}
	if called {
		t.Error("a malformed id reached the network")
	}

	// A well-formed id still works.
	if _, err := client.fetchAdvisory(context.Background(), "GHSA-xvhc-gm7j-mhmc"); err != nil {
		t.Errorf("valid id rejected: %v", err)
	}
	if !called {
		t.Error("a valid id should have been fetched")
	}
}
