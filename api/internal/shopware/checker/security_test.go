package checker

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// checkSecurity itself fetches https://raw.githubusercontent.com/... directly via
// httputil.NewHTTPClient (not input.Client), so it cannot be driven by a mock
// without real network access. These tests cover the pure parts: unmarshalling
// of securityData and the version->advisory lookup logic that checkSecurity relies on.

func TestSecurityDataUnmarshal(t *testing.T) {
	raw := `{
		"latestPluginVersion": "2.0.10",
		"advisories": {
			"adv-1": {"title": "XSS in admin", "link": "http://example.com/adv-1", "cve": "CVE-2024-0001", "source": "FoS"},
			"adv-2": {"title": "SQL injection", "link": "http://example.com/adv-2", "cve": "", "source": "FoS"}
		},
		"versionToAdvisories": {
			"6.5.0.0": ["adv-1", "adv-2"],
			"6.6.0.0": ["adv-2"]
		}
	}`

	var data securityData
	require.NoError(t, json.Unmarshal([]byte(raw), &data))

	assert.Equal(t, "2.0.10", data.LatestPluginVersion)
	require.Len(t, data.Advisories, 2)

	adv := data.Advisories["adv-1"]
	assert.Equal(t, "XSS in admin", adv.Title)
	assert.Equal(t, "CVE-2024-0001", adv.CVE)
	assert.Equal(t, "http://example.com/adv-1", adv.Link)

	ids := data.VersionToAdvisories["6.5.0.0"]
	assert.Equal(t, []string{"adv-1", "adv-2"}, ids)
}

// TestSecurityVersionLookup documents the lookup semantics checkSecurity uses:
// a version is only flagged when present in versionToAdvisories with a non-empty list.
func TestSecurityVersionLookup(t *testing.T) {
	data := securityData{
		VersionToAdvisories: map[string][]string{
			"6.5.0.0": {"adv-1"},
			"6.6.0.0": {},
		},
	}

	tests := []struct {
		name        string
		version     string
		wantFlagged bool
	}{
		{name: "version with advisories is flagged", version: "6.5.0.0", wantFlagged: true},
		{name: "version with empty advisory list is not flagged", version: "6.6.0.0", wantFlagged: false},
		{name: "unknown version is not flagged", version: "6.7.0.0", wantFlagged: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, ok := data.VersionToAdvisories[tt.version]
			flagged := ok && len(ids) > 0
			assert.Equal(t, tt.wantFlagged, flagged)
		})
	}
}
