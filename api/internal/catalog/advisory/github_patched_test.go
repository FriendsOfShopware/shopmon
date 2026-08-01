package advisory

import (
	"encoding/json"
	"testing"
)

// GitHub reports one patched release per affected branch. Folding them by
// Shopware line is what lets the UI say "upgrade to 6.7.10.1" to a 6.7 shop and
// "6.6.10.18" to a 6.6 one, from a single advisory.
func TestFirstPatchedByLine(t *testing.T) {
	// Verbatim shape of the real GitHub response for GHSA-xvhc-gm7j-mhmc.
	raw := []byte(`{"ghsa_id":"GHSA-xvhc-gm7j-mhmc","vulnerabilities":[
      {"package":{"name":"shopware/core","ecosystem":"composer"},"vulnerable_version_range":">= 6.7.0.0, < 6.7.10.1","first_patched_version":"6.7.10.1"},
      {"package":{"name":"shopware/core","ecosystem":"composer"},"vulnerable_version_range":"< 6.6.10.18","first_patched_version":"6.6.10.18"},
      {"package":{"name":"shopware/platform","ecosystem":"composer"},"vulnerable_version_range":">= 6.7.0.0, < 6.7.10.1","first_patched_version":"6.7.10.1"},
      {"package":{"name":"other/lib","ecosystem":"composer"},"first_patched_version":"9.9.9"}]}`)
	var resp githubAdvisoryResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatal(err)
	}
	got := firstPatchedByLine(resp)
	t.Logf("firstPatchedByLine = %v", got)
	if got["6.7"] != "6.7.10.1" || got["6.6"] != "6.6.10.18" {
		t.Fatalf("got %v, want 6.7=6.7.10.1 and 6.6=6.6.10.18", got)
	}
	if _, ok := got["9.9"]; ok {
		t.Errorf("non-shopware package must be ignored: %v", got)
	}
}
