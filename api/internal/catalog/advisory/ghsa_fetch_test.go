package advisory

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/ghsa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGHSAReference(t *testing.T) {
	cases := []struct {
		name     string
		ref      string
		wantID   string
		wantLink string
		wantErr  bool
	}{
		{"bare id", "GHSA-vj8m-rx9g-45fq", "GHSA-vj8m-rx9g-45fq", "", false},
		{"lowercase id", "ghsa-VJ8M-RX9G-45FQ", "GHSA-vj8m-rx9g-45fq", "", false},
		{
			"repository url",
			"https://github.com/FriendsOfShopware/FroshPlatformMailArchive/security/advisories/GHSA-vj8m-rx9g-45fq",
			"GHSA-vj8m-rx9g-45fq",
			"https://github.com/FriendsOfShopware/FroshPlatformMailArchive/security/advisories/GHSA-vj8m-rx9g-45fq",
			false,
		},
		{
			"global url",
			"https://github.com/advisories/GHSA-vj8m-rx9g-45fq",
			"GHSA-vj8m-rx9g-45fq",
			"https://github.com/advisories/GHSA-vj8m-rx9g-45fq",
			false,
		},
		{"empty", "  ", "", "", true},
		{"cve is not a ghsa", "CVE-2026-1234", "", "", true},
		{"path traversal attempt", "GHSA-../../user", "", "", true},
		{"foreign url", "https://example.com/advisories/GHSA-vj8m-rx9g-45fq", "", "", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, link, err := ParseGHSAReference(tc.ref)
			if tc.wantErr {
				require.ErrorIs(t, err, ghsa.ErrInvalidReference)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantID, id)
			assert.Equal(t, tc.wantLink, link)
		})
	}
}

// The advisory that motivated the import lives only on its repository (its
// CVE is still pending), so the fetch must fall back from the global endpoint
// to the repository-scoped one named by the pasted URL.
func TestFetchGHSARepositoryFallback(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/advisories/GHSA-vj8m-rx9g-45fq", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	mux.HandleFunc("/repos/FriendsOfShopware/FroshPlatformMailArchive/security-advisories/GHSA-vj8m-rx9g-45fq", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{
			"ghsa_id": "GHSA-vj8m-rx9g-45fq",
			"cve_id": null,
			"summary": "Missing ACL enforcement permits unauthorized access to archived email",
			"description": "## Summary\n\nDetails…",
			"html_url": "https://github.com/FriendsOfShopware/FroshPlatformMailArchive/security/advisories/GHSA-vj8m-rx9g-45fq",
			"severity": "high",
			"published_at": "2026-10-19T07:55:51Z",
			"cvss_severities": {"cvss_v3": {"score": 7.1, "vector_string": "CVSS:3.1/AV:N/AC:L/PR:L/UI:N/S:U/C:H/I:L/A:N"}},
			"cwes": [{"cwe_id": "CWE-862", "name": "Missing Authorization"}],
			"references": [{"url": "https://example.com/writeup"}],
			"vulnerabilities": [
				{"package": {"ecosystem": "composer", "name": "frosh/mail-platform-archive"},
				 "vulnerable_version_range": ">= 2.0.0, <= 3.6.0", "patched_versions": "3.7.0", "first_patched_version": {"identifier": "3.7.0"}},
				{"package": {"ecosystem": "npm", "name": "unrelated"},
				 "vulnerable_version_range": "<1.0.0", "patched_versions": "1.0.0"}
			]
		}`)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	fetcher := &GHSAFetcher{client: &githubClient{
		baseURL:    server.URL,
		userAgent:  "test",
		httpClient: server.Client(),
	}}

	fetched, err := fetcher.FetchGHSA(context.Background(),
		"https://github.com/FriendsOfShopware/FroshPlatformMailArchive/security/advisories/GHSA-vj8m-rx9g-45fq")
	require.NoError(t, err)

	assert.Equal(t, "GHSA-vj8m-rx9g-45fq", fetched.GHSAID)
	assert.Equal(t, "", fetched.CVE)
	assert.Equal(t, "Missing ACL enforcement permits unauthorized access to archived email", fetched.Title)
	assert.Equal(t, "high", fetched.Severity)
	require.NotNil(t, fetched.CVSSScore)
	assert.Equal(t, 7.1, *fetched.CVSSScore)
	require.Len(t, fetched.CWEs, 1)
	assert.Equal(t, "CWE-862", fetched.CWEs[0].ID)
	require.NotNil(t, fetched.PublishedAt)
	assert.Equal(t, 2026, fetched.PublishedAt.Year())
	assert.False(t, fetched.Withdrawn)
	require.Len(t, fetched.Packages, 2, "the fetcher stays faithful to GitHub; ecosystem filtering happens in the preview")
	assert.Equal(t, "frosh/mail-platform-archive", fetched.Packages[0].Name)
	assert.Equal(t, ">= 2.0.0, <= 3.6.0", fetched.Packages[0].VulnerableVersionRange)
	assert.Equal(t, "3.7.0", fetched.Packages[0].PatchedVersion)

	_, err = fetcher.FetchGHSA(context.Background(), "not-a-reference")
	assert.ErrorIs(t, err, ghsa.ErrInvalidReference)

	_, err = fetcher.FetchGHSA(context.Background(), "GHSA-zzzz-yyyy-xxxx")
	assert.ErrorIs(t, err, ghsa.ErrAdvisoryNotFound)
}
