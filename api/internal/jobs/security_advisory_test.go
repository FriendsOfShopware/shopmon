package jobs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const samplePackagistResponse = `{
  "advisories": {
    "shopware/core": [
      {
        "advisoryId": "PKSA-abc",
        "packageName": "shopware/core",
        "title": "SQL injection in core",
        "link": "https://github.com/shopware/shopware/security/advisories/GHSA-xxxx",
        "cve": "CVE-2023-1234",
        "affectedVersions": ">=6.4.0.0,<6.4.20.0",
        "reportedAt": "2023-05-10 14:30:00",
        "severity": "high",
        "sources": [
          { "name": "GitHub", "remoteId": "GHSA-xxxx" }
        ]
      }
    ],
    "shopware/storefront": []
  }
}`

func TestSecurityAdvisoryFetch(t *testing.T) {
	var gotPackages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPackages = r.URL.Query()["packages[]"]
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(samplePackagistResponse))
	}))
	defer srv.Close()

	h := &SecurityAdvisorySyncHandler{baseURL: srv.URL + "/"}

	resp, err := h.fetch(context.Background())
	require.NoError(t, err)

	// All tracked packages are requested.
	assert.ElementsMatch(t, securityAdvisoryPackages, gotPackages)

	advisories := resp.Advisories["shopware/core"]
	require.Len(t, advisories, 1)
	assert.Equal(t, "PKSA-abc", advisories[0].AdvisoryID)
	assert.Empty(t, resp.Advisories["shopware/storefront"])
}

func TestToUpsertParams(t *testing.T) {
	var resp packagistAdvisoriesResponse
	require.NoError(t, json.Unmarshal([]byte(samplePackagistResponse), &resp))

	params := toUpsertParams(resp.Advisories["shopware/core"][0])

	assert.Equal(t, "PKSA-abc", params.AdvisoryID)
	assert.Equal(t, "packagist", params.Origin)
	assert.Equal(t, "shopware/core", params.PackageName)
	assert.Equal(t, ">=6.4.0.0,<6.4.20.0", params.AffectedVersions)
	require.NotNil(t, params.Cve)
	assert.Equal(t, "CVE-2023-1234", *params.Cve)
	require.NotNil(t, params.SourceName)
	assert.Equal(t, "GitHub", *params.SourceName)
	require.NotNil(t, params.SourceRemoteID)
	assert.Equal(t, "GHSA-xxxx", *params.SourceRemoteID)
	require.True(t, params.ReportedAt.Valid)
	assert.Equal(t, 2023, params.ReportedAt.Time.Year())
}

func TestParsePackagistTime(t *testing.T) {
	assert.False(t, parsePackagistTime("").Valid)
	assert.False(t, parsePackagistTime("not-a-date").Valid)

	ts := parsePackagistTime("2023-05-10 14:30:00")
	require.True(t, ts.Valid)
	assert.Equal(t, 5, int(ts.Time.Month()))
}
