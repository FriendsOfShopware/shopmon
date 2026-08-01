package sbom

import (
	"sort"
	"testing"
)

func TestIndexMatch(t *testing.T) {
	idx := NewIndex([]AdvisoryPackage{
		{AdvisoryID: "CVE-2024-1", PackageName: "shopware/core", AffectedVersions: ">=6.6.0.0,<6.6.10.1"},
		{AdvisoryID: "CVE-2024-2", PackageName: "symfony/http-kernel", AffectedVersions: ">=6.4.0,<6.4.3"},
		{AdvisoryID: "CVE-2024-3", PackageName: "twig/twig", AffectedVersions: "<3.0.0"},
		{AdvisoryID: "CVE-2024-4", PackageName: "shopware/core", AffectedVersions: ""},
	})

	matches := idx.Match(map[string]string{
		"shopware/core":       "6.6.5.0",
		"symfony/http-kernel": "6.4.2",
		"twig/twig":           "3.8.0",
		"unknown/package":     "1.0.0",
	})

	ids := make([]string, 0, len(matches))
	for _, m := range matches {
		ids = append(ids, m.AdvisoryID)
	}
	sort.Strings(ids)

	want := []string{"CVE-2024-1", "CVE-2024-2"}
	if len(ids) != len(want) {
		t.Fatalf("got %v, want %v", ids, want)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Errorf("got %v, want %v", ids, want)
			break
		}
	}
}

func TestIndexMatchRecordsInstalledVersion(t *testing.T) {
	idx := NewIndex([]AdvisoryPackage{
		{AdvisoryID: "CVE-1", PackageName: "shopware/core", AffectedVersions: "<6.7.0.0"},
	})

	matches := idx.Match(map[string]string{"shopware/core": "6.6.5.0"})
	if len(matches) != 1 {
		t.Fatalf("got %d matches", len(matches))
	}
	m := matches[0]
	if m.InstalledVersion != "6.6.5.0" {
		t.Errorf("InstalledVersion = %q", m.InstalledVersion)
	}
	if m.AffectedVersions != "<6.7.0.0" {
		t.Errorf("AffectedVersions = %q", m.AffectedVersions)
	}
	// The catalog's original casing is preserved for display.
	if m.PackageName != "shopware/core" {
		t.Errorf("PackageName = %q", m.PackageName)
	}
}

func TestIndexMatchIsCaseInsensitive(t *testing.T) {
	idx := NewIndex([]AdvisoryPackage{
		{AdvisoryID: "CVE-1", PackageName: "Twig/Twig", AffectedVersions: "<4.0.0"},
	})

	if got := idx.Match(map[string]string{"twig/twig": "3.8.0"}); len(got) != 1 {
		t.Fatalf("got %d matches, want 1", len(got))
	}
}

func TestAffects(t *testing.T) {
	cases := []struct {
		name       string
		version    string
		constraint string
		want       bool
	}{
		{"in range", "6.6.5.0", ">=6.6.0.0,<6.6.10.1", true},
		{"above range", "6.6.10.5", ">=6.6.0.0,<6.6.10.1", false},
		{"or branch first", "6.5.0.0", "<6.6.10.18|>=6.7.0.0,<6.7.10.1", true},
		{"or branch second", "6.7.1.0", "<6.6.10.18|>=6.7.0.0,<6.7.10.1", true},
		{"outside both branches", "6.9.9.9", "<6.6.10.18|>=6.7.0.0,<6.7.10.1", false},
		{"v prefix", "v1.2.3", ">=1.0.0,<2.0.0", true},
		{"empty constraint", "1.0.0", "", false},
		{"wildcard is rejected", "1.0.0", "*", false},
		{"empty version", "", "<2.0.0", false},
		{"unparseable constraint", "1.0.0", "not-a-constraint", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := affects(tc.version, tc.constraint); got != tc.want {
				t.Errorf("affects(%q, %q) = %v, want %v", tc.version, tc.constraint, got, tc.want)
			}
		})
	}
}

func TestNewIndexPackageNames(t *testing.T) {
	idx := NewIndex([]AdvisoryPackage{
		{AdvisoryID: "A", PackageName: "shopware/core", AffectedVersions: "<1.0"},
		{AdvisoryID: "B", PackageName: "SHOPWARE/CORE", AffectedVersions: "<2.0"},
		{AdvisoryID: "C", PackageName: "  ", AffectedVersions: "<1.0"},
	})

	names := idx.PackageNames()
	if len(names) != 1 || names[0] != "shopware/core" {
		t.Errorf("PackageNames() = %v, want [shopware/core]", names)
	}
}

func TestNilIndexIsSafe(t *testing.T) {
	var idx *Index
	if got := idx.Match(map[string]string{"a/b": "1.0"}); got != nil {
		t.Errorf("Match on nil index = %v", got)
	}
	if got := idx.PackageNames(); got != nil {
		t.Errorf("PackageNames on nil index = %v", got)
	}
}
