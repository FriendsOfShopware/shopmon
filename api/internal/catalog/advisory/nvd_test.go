package advisory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNVDFetchCVE(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/rest/json/cves/2.0", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "CVE-2026-48013", r.URL.Query().Get("cveId"))
		assert.Equal(t, "test-key", r.Header.Get("apiKey"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"vulnerabilities": []map[string]any{
				{
					"cve": map[string]any{
						"id": "CVE-2026-48013",
						"descriptions": []map[string]string{
							{"lang": "en", "value": "Shopware prior to fixed versions allows SSRF via media external-link."},
						},
						"metrics": map[string]any{
							"cvssMetricV31": []map[string]any{
								{
									"cvssData": map[string]any{
										"baseScore":    4.1,
										"vectorString": "CVSS:3.1/AV:N/AC:L/PR:H/UI:N/S:C/C:L/I:N/A:N",
									},
								},
							},
						},
						"weaknesses": []map[string]any{
							{"description": []map[string]string{{"lang": "en", "value": "CWE-918"}}},
						},
						"references": []map[string]string{
							{"url": "https://github.com/advisories/GHSA-gq96-5pfx-f4vc"},
						},
					},
				},
			},
		})
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newNVDClient("test-key", defaultUserAgent, server.Client())
	client.baseURL = server.URL + "/rest/json/cves/2.0"

	details, err := client.fetchCVE(context.Background(), "CVE-2026-48013")
	require.NoError(t, err)
	require.NotNil(t, details)
	assert.Contains(t, details.Description, "SSRF")
	assert.NotEmpty(t, details.Summary)
	require.NotNil(t, details.CVSSScore)
	assert.InDelta(t, 4.1, *details.CVSSScore, 0.01)
	require.Len(t, details.CWEs, 1)
	assert.Equal(t, "CWE-918", details.CWEs[0].ID)
	assert.Equal(t, "https://nvd.nist.gov/vuln/detail/CVE-2026-48013", details.References[0])
}

func TestTruncateSummary(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "Short.", truncateSummary("Short."))
	long := "This is a long first sentence that should be kept whole. And more text after."
	assert.Equal(t, "This is a long first sentence that should be kept whole.", truncateSummary(long))
}

// NVD descriptions are arbitrary UTF-8 and this client falls back to
// non-English text, so truncation must cut on a rune boundary: invalid UTF-8
// is rejected by Postgres and would fail enrichment for that CVE entirely.
func TestTruncateSummaryKeepsValidUTF8(t *testing.T) {
	t.Parallel()

	// Multi-byte runes straddling the 197-byte cut, with no sentence break.
	text := strings.Repeat("ä", 300)
	got := truncateSummary(text)

	if !utf8.ValidString(got) {
		t.Fatalf("truncated summary is not valid UTF-8: %q", got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected an ellipsis, got %q", got)
	}
	if len(got) > 201 {
		t.Errorf("truncated summary too long: %d bytes", len(got))
	}
}
