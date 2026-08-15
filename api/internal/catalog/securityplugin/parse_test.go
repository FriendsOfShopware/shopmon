package securityplugin

import (
	"sort"
	"testing"
)

func TestExtractGHSAs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		changelog string
		want      []string
	}{
		{
			name:      "href only",
			changelog: `<a href="https://github.com/shopware/shopware/security/advisories/GHSA-abcd-1234-wxyz">Advisory</a>`,
			want:      []string{"GHSA-ABCD-1234-WXYZ"},
		},
		{
			name:      "anchor text only",
			changelog: `<li>Fixed <a href="/x">GHSA-abcd-1234-wxyz</a></li>`,
			want:      []string{"GHSA-ABCD-1234-WXYZ"},
		},
		{
			// The real changelog puts the id in both href and text; counting it
			// twice would inflate the coverage statistics.
			name:      "href and text deduplicate",
			changelog: `<a href="https://example.com/GHSA-abcd-1234-wxyz">GHSA-abcd-1234-wxyz</a>`,
			want:      []string{"GHSA-ABCD-1234-WXYZ"},
		},
		{
			name:      "bare prose",
			changelog: `Added fix for GHSA-gq96-5pfx-f4vc in the media endpoint.`,
			want:      []string{"GHSA-GQ96-5PFX-F4VC"},
		},
		{
			name:      "multiple ids preserve first-seen order",
			changelog: `<ul><li>GHSA-f8q6-3g5w-jjr6</li><li>GHSA-9v5m-39wh-5chq</li></ul>`,
			want:      []string{"GHSA-F8Q6-3G5W-JJR6", "GHSA-9V5M-39WH-5CHQ"},
		},
		{
			// Guards against a looser pattern matching CVE ids or truncated
			// identifiers, which would create fixes for advisories that do not exist.
			name:      "near misses are not matched",
			changelog: `GHSA-abc-1234-wxyz and CVE-2024-1234 and GHSA-abcd-1234`,
			want:      nil,
		},
		{
			name:      "no identifiers",
			changelog: `<ul><li>Improved performance of the SVG validator.</li></ul>`,
			want:      nil,
		},
		{
			name:      "empty string does not panic",
			changelog: "",
			want:      nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ExtractGHSAs(tc.changelog)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("got %v, want %v", got, tc.want)
					break
				}
			}
		})
	}
}

func TestBranch(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"4.0.12":   "4",
		"3.0.15":   "3",
		"2.0.20":   "2",
		"10.1.0":   "10",
		"":         "",
		"dev-main": "",
		"4":        "",
	}
	for input, want := range cases {
		if got := Branch(input); got != want {
			t.Errorf("Branch(%q) = %q, want %q", input, got, want)
		}
	}
}

// The store returns changelogs for every branch at once, so a GHSA fixed on all
// three lines must yield three rows with three different minimums. A global
// minimum would tell a 6.7 shop that 2.0.20 fixes it, which is wrong.
func TestDeriveFixesGroupsByBranch(t *testing.T) {
	t.Parallel()

	entries := []ChangelogEntry{
		{Version: "4.0.10", Changelog: "fix for GHSA-xvhc-gm7j-mhmc"},
		{Version: "3.0.15", Changelog: "fix for GHSA-xvhc-gm7j-mhmc"},
		{Version: "2.0.20", Changelog: "fix for GHSA-xvhc-gm7j-mhmc"},
	}

	fixes, _ := DeriveFixes(entries)
	byBranch := map[string]string{}
	for _, fix := range fixes {
		byBranch[fix.Branch] = fix.PluginVersion
	}

	want := map[string]string{"4": "4.0.10", "3": "3.0.15", "2": "2.0.20"}
	for branch, version := range want {
		if byBranch[branch] != version {
			t.Errorf("branch %s = %q, want %q", branch, byBranch[branch], version)
		}
	}
}

// A later entry refining an existing fix must not raise the required version.
// This is the single most consequential rule in the parser: getting it backwards
// would tell every shop to upgrade past the release that actually fixed them.
func TestDeriveFixesKeepsMinimumNotLatest(t *testing.T) {
	t.Parallel()

	entries := []ChangelogEntry{
		{Version: "4.0.10", Changelog: "Added fix for GHSA-xvhc-gm7j-mhmc"},
		{Version: "4.0.12", Changelog: "Improved the SVG validator for GHSA-xvhc-gm7j-mhmc"},
	}

	fixes, _ := DeriveFixes(entries)
	if len(fixes) != 1 {
		t.Fatalf("got %d fixes, want 1", len(fixes))
	}
	if fixes[0].PluginVersion != "4.0.10" {
		t.Errorf("PluginVersion = %q, want 4.0.10 (the first fix, not the refinement)", fixes[0].PluginVersion)
	}
}

// Version ordering must be numeric: string comparison puts "4.0.10" before
// "4.0.9", which would understate the required version.
func TestDeriveFixesUsesNumericVersionOrder(t *testing.T) {
	t.Parallel()

	entries := []ChangelogEntry{
		{Version: "4.0.10", Changelog: "GHSA-aaaa-bbbb-cccc"},
		{Version: "4.0.9", Changelog: "GHSA-aaaa-bbbb-cccc"},
	}

	fixes, _ := DeriveFixes(entries)
	if len(fixes) != 1 || fixes[0].PluginVersion != "4.0.9" {
		t.Fatalf("got %+v, want minimum 4.0.9", fixes)
	}
}

func TestDeriveFixesIsOrderIndependent(t *testing.T) {
	t.Parallel()

	forward := []ChangelogEntry{
		{Version: "4.0.9", Changelog: "GHSA-aaaa-bbbb-cccc"},
		{Version: "4.0.12", Changelog: "GHSA-aaaa-bbbb-cccc"},
	}
	reverse := []ChangelogEntry{forward[1], forward[0]}

	a, _ := DeriveFixes(forward)
	b, _ := DeriveFixes(reverse)
	if len(a) != 1 || len(b) != 1 || a[0].PluginVersion != b[0].PluginVersion {
		t.Fatalf("order changed the result: %+v vs %+v", a, b)
	}
}

// A GHSA present only on the newest branch means older Shopware lines cannot fix
// it via the plugin at all -- they must upgrade the core. Verified against the
// live store data for GHSA-gq96-5pfx-f4vc.
func TestDeriveFixesBranchOnlyGhsaProducesNoOtherBranches(t *testing.T) {
	t.Parallel()

	entries := []ChangelogEntry{
		{Version: "4.0.10", Changelog: "fix for GHSA-gq96-5pfx-f4vc"},
		{Version: "3.0.15", Changelog: "unrelated improvements"},
		{Version: "2.0.20", Changelog: "unrelated improvements"},
	}

	fixes, _ := DeriveFixes(entries)
	if len(fixes) != 1 {
		t.Fatalf("got %d fixes, want 1", len(fixes))
	}
	if fixes[0].Branch != "4" {
		t.Errorf("Branch = %q, want 4", fixes[0].Branch)
	}
}

// Coverage must record a branch that was parsed but contains no identifiers, so
// the checker can tell that apart from a branch it never saw.
func TestDeriveFixesCoverageRecordsParsedBranchesWithoutGhsas(t *testing.T) {
	t.Parallel()

	entries := []ChangelogEntry{
		{Version: "1.0.5", Changelog: "Initial release"},
		{Version: "1.0.9", Changelog: "Bug fixes"},
	}

	fixes, coverage := DeriveFixes(entries)
	if len(fixes) != 0 {
		t.Fatalf("got %d fixes, want 0", len(fixes))
	}
	if len(coverage) != 1 {
		t.Fatalf("got %d coverage rows, want 1", len(coverage))
	}
	if coverage[0].Branch != "1" || coverage[0].GhsaCount != 0 {
		t.Errorf("coverage = %+v, want branch 1 with zero ids", coverage[0])
	}
	if coverage[0].MaxParsedVersion != "1.0.9" {
		t.Errorf("MaxParsedVersion = %q, want 1.0.9", coverage[0].MaxParsedVersion)
	}
	// The branch was parsed end to end even though no entry named a GHSA, so
	// both bounds must be populated. Leaving the minimum empty would read as
	// "never parsed" to anyone debugging why the branch resolves to unknown.
	if coverage[0].MinParsedVersion != "1.0.5" {
		t.Errorf("MinParsedVersion = %q, want 1.0.5", coverage[0].MinParsedVersion)
	}
}

// Releases that predate GHSA labelling still count as parsed: the coverage
// range must span them rather than starting at the first advisory.
func TestDeriveFixesCoverageSpansEntriesWithoutGhsas(t *testing.T) {
	t.Parallel()

	entries := []ChangelogEntry{
		{Version: "1.0.1", Changelog: "Initial release"},
		{Version: "1.0.5", Changelog: "Fixes GHSA-xvhc-gm7j-mhmc"},
		{Version: "1.0.9", Changelog: "Bug fixes"},
	}

	_, coverage := DeriveFixes(entries)
	if len(coverage) != 1 {
		t.Fatalf("got %d coverage rows, want 1", len(coverage))
	}
	if coverage[0].MinParsedVersion != "1.0.1" {
		t.Errorf("MinParsedVersion = %q, want 1.0.1 (the earliest parsed release)", coverage[0].MinParsedVersion)
	}
	if coverage[0].MaxParsedVersion != "1.0.9" {
		t.Errorf("MaxParsedVersion = %q, want 1.0.9", coverage[0].MaxParsedVersion)
	}
	if coverage[0].GhsaCount != 1 {
		t.Errorf("GhsaCount = %d, want 1", coverage[0].GhsaCount)
	}
}

func TestDeriveFixesSkipsUnparseableVersions(t *testing.T) {
	t.Parallel()

	fixes, coverage := DeriveFixes([]ChangelogEntry{
		{Version: "", Changelog: "GHSA-aaaa-bbbb-cccc"},
		{Version: "dev-main", Changelog: "GHSA-aaaa-bbbb-cccc"},
	})
	if len(fixes) != 0 || len(coverage) != 0 {
		t.Fatalf("got %d fixes / %d coverage, want none", len(fixes), len(coverage))
	}
}

func TestDeriveFixesMultipleIdsInOneEntry(t *testing.T) {
	t.Parallel()

	// Mirrors the real 4.0.10 entry, which lists nine advisories at once.
	ids := []string{
		"GHSA-4x3x-869w-xx3m", "GHSA-7w52-7jvm-m9vw", "GHSA-8v9p-g828-v98f",
		"GHSA-9v5m-39wh-5chq", "GHSA-f8q6-3g5w-jjr6", "GHSA-gq96-5pfx-f4vc",
		"GHSA-gv8p-48fr-4fxg", "GHSA-v39m-97p8-gqg7", "GHSA-xvhc-gm7j-mhmc",
	}
	changelog := ""
	for _, id := range ids {
		changelog += `<li>Added fix for <a href="https://example.com/` + id + `">` + id + `</a></li>`
	}

	fixes, coverage := DeriveFixes([]ChangelogEntry{{Version: "4.0.10", Changelog: changelog}})
	if len(fixes) != len(ids) {
		t.Fatalf("got %d fixes, want %d", len(fixes), len(ids))
	}
	if coverage[0].GhsaCount != len(ids) {
		t.Errorf("GhsaCount = %d, want %d", coverage[0].GhsaCount, len(ids))
	}

	got := make([]string, 0, len(fixes))
	for _, fix := range fixes {
		got = append(got, fix.PluginVersion)
	}
	sort.Strings(got)
	for _, version := range got {
		if version != "4.0.10" {
			t.Fatalf("unexpected version %q", version)
		}
	}
}

func TestShopwareBranchMapping(t *testing.T) {
	t.Parallel()

	if got := ShopwareBranch("4"); got != "6.7" {
		t.Errorf("ShopwareBranch(4) = %q, want 6.7", got)
	}
	// An unmapped future branch must read as unknown, not as a wrong line.
	if got := ShopwareBranch("9"); got != "" {
		t.Errorf("ShopwareBranch(9) = %q, want empty", got)
	}
	if got := PluginBranchForShopware("6.7.8.2"); got != "4" {
		t.Errorf("PluginBranchForShopware(6.7.8.2) = %q, want 4", got)
	}
	if got := PluginBranchForShopware("6.4.20.0"); got != "" {
		t.Errorf("PluginBranchForShopware(6.4.20.0) = %q, want empty", got)
	}
	if got := ShopwareLine("6.6.10.18"); got != "6.6" {
		t.Errorf("ShopwareLine = %q, want 6.6", got)
	}
}
