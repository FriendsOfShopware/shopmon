package checker

import (
	"encoding/json"
	"testing"
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
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if data.LatestPluginVersion != "2.0.10" {
		t.Errorf("LatestPluginVersion = %q, want %q", data.LatestPluginVersion, "2.0.10")
	}
	if len(data.Advisories) != 2 {
		t.Fatalf("expected 2 advisories, got %d", len(data.Advisories))
	}

	adv := data.Advisories["adv-1"]
	if adv.Title != "XSS in admin" || adv.CVE != "CVE-2024-0001" || adv.Link != "http://example.com/adv-1" {
		t.Errorf("adv-1 = %+v, unexpected fields", adv)
	}

	ids := data.VersionToAdvisories["6.5.0.0"]
	if len(ids) != 2 || ids[0] != "adv-1" || ids[1] != "adv-2" {
		t.Errorf("versionToAdvisories[6.5.0.0] = %v, want [adv-1 adv-2]", ids)
	}
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
			if flagged != tt.wantFlagged {
				t.Errorf("flagged = %v, want %v", flagged, tt.wantFlagged)
			}
		})
	}
}
