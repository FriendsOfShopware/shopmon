package checker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchAdvisories(t *testing.T) {
	t.Parallel()

	advisories := []SecurityAdvisory{
		{
			ID: "CVE-2026-1", Title: "SSRF", CVE: "CVE-2026-1", Severity: "medium",
			Packages: []SecurityAdvisoryPackage{
				{PackageName: "shopware/core", AffectedVersions: ">=6.7.0.0,<6.7.10.1"},
				{PackageName: "shopware/platform", AffectedVersions: ">=6.7.0.0,<6.7.10.1"},
			},
		},
		{
			ID: "CVE-2020-1", Title: "Old issue", CVE: "CVE-2020-1", Severity: "high",
			Packages: []SecurityAdvisoryPackage{
				{PackageName: "shopware/core", AffectedVersions: "<6.4.0.0"},
			},
		},
		{
			ID: "OTHER", Title: "Other", Severity: "high",
			Packages: []SecurityAdvisoryPackage{
				{PackageName: "acme/plugin", AffectedVersions: "*"},
			},
		},
	}

	matched := matchAdvisories("6.7.5.0", nil, advisories)
	require.Len(t, matched, 1)
	assert.Equal(t, "CVE-2026-1", matched[0].ID)

	assert.Empty(t, matchAdvisories("6.7.10.1", nil, advisories))
}

// Without an SBOM only shopware/* constraints are evaluated. Once the package
// inventory is known, a vulnerable dependency is matched at its own version.
func TestMatchAdvisoriesUsesSbomPackages(t *testing.T) {
	t.Parallel()

	advisories := []SecurityAdvisory{
		{
			ID: "CVE-2026-2", Title: "Symfony issue", CVE: "CVE-2026-2", Severity: "high",
			Packages: []SecurityAdvisoryPackage{
				{PackageName: "symfony/http-kernel", AffectedVersions: ">=6.4.0,<6.4.3"},
			},
		},
	}

	// No SBOM: a non-shopware package cannot be evaluated at all.
	assert.Empty(t, matchAdvisories("6.7.5.0", nil, advisories))

	matched := matchAdvisories("6.7.5.0", map[string]string{
		"symfony/http-kernel": "6.4.2",
	}, advisories)
	require.Len(t, matched, 1)
	assert.Equal(t, "CVE-2026-2", matched[0].ID)
	assert.Equal(t, "symfony/http-kernel", matched[0].PackageName)
	assert.Equal(t, "6.4.2", matched[0].InstalledVersion)

	// Patched version is no longer affected.
	assert.Empty(t, matchAdvisories("6.7.5.0", map[string]string{
		"symfony/http-kernel": "6.4.3",
	}, advisories))
}

// The SBOM reports the real shopware/core version, which should win over the
// core version string from /_info/config when both are present.
func TestMatchAdvisoriesPrefersSbomVersionForShopwarePackages(t *testing.T) {
	t.Parallel()

	advisories := []SecurityAdvisory{
		{
			ID: "CVE-2026-3", Title: "Core issue", CVE: "CVE-2026-3", Severity: "high",
			Packages: []SecurityAdvisoryPackage{
				{PackageName: "shopware/core", AffectedVersions: "<6.7.10.1"},
			},
		},
	}

	// Config reports a vulnerable version but the SBOM shows core is patched.
	assert.Empty(t, matchAdvisories("6.7.5.0", map[string]string{
		"shopware/core": "6.7.10.2",
	}, advisories))
}

func TestMatchAdvisoriesPackageLookupIsCaseInsensitive(t *testing.T) {
	t.Parallel()

	advisories := []SecurityAdvisory{
		{
			ID: "CVE-2026-4", Title: "Twig issue", CVE: "CVE-2026-4", Severity: "medium",
			Packages: []SecurityAdvisoryPackage{
				{PackageName: "Twig/Twig", AffectedVersions: "<4.0.0"},
			},
		},
	}

	matched := matchAdvisories("6.7.5.0", map[string]string{"twig/twig": "3.8.0"}, advisories)
	require.Len(t, matched, 1)
	assert.Equal(t, "CVE-2026-4", matched[0].ID)
}

// The security plugin backports individual advisories, so resolution is
// per-GHSA. Versions below mirror the real store data: GHSA-xvhc-gm7j-mhmc is
// fixed in 4.0.10 on the 6.7 line, and GHSA-gq96-5pfx-f4vc exists only there.
func TestResolveByPluginStates(t *testing.T) {
	t.Parallel()

	index := buildPluginFixIndex([]PluginFix{
		{GhsaID: "GHSA-XVHC-GM7J-MHMC", Branch: "4", PluginVersion: "4.0.10"},
		{GhsaID: "GHSA-XVHC-GM7J-MHMC", Branch: "3", PluginVersion: "3.0.15"},
		{GhsaID: "GHSA-GQ96-5PFX-F4VC", Branch: "4", PluginVersion: "4.0.10"},
	})
	coverage := PluginCoverage{"4": 26, "3": 21, "2": 22}
	adv := SecurityAdvisory{ID: "CVE-1", GHSA: "GHSA-xvhc-gm7j-mhmc"}

	plugin := func(version string) *Extension {
		return &Extension{Name: swagPlatformSecurity, Active: true, Installed: true, Version: version}
	}

	cases := []struct {
		name            string
		shopwareVersion string
		advisory        SecurityAdvisory
		ext             *Extension
		coverage        PluginCoverage
		want            CoverageState
		wantRequired    string
	}{
		{
			name:            "plugin at or above the fix version covers it",
			shopwareVersion: "6.7.8.2", advisory: adv, ext: plugin("4.0.10"),
			coverage: coverage, want: CoverageCovered, wantRequired: "4.0.10",
		},
		{
			name:            "a newer plugin still covers it",
			shopwareVersion: "6.7.8.2", advisory: adv, ext: plugin("4.0.12"),
			coverage: coverage, want: CoverageCovered, wantRequired: "4.0.10",
		},
		{
			name:            "an older plugin leaves the shop exposed",
			shopwareVersion: "6.7.8.2", advisory: adv, ext: plugin("4.0.9"),
			coverage: coverage, want: CoverageOutdated, wantRequired: "4.0.10",
		},
		{
			name:            "an inactive plugin backports nothing",
			shopwareVersion: "6.7.8.2", advisory: adv,
			ext:      &Extension{Name: swagPlatformSecurity, Active: false, Installed: true, Version: "4.0.12"},
			coverage: coverage, want: CoverageNotInstalled, wantRequired: "4.0.10",
		},
		{
			name:            "no plugin at all, but a backport exists",
			shopwareVersion: "6.7.8.2", advisory: adv, ext: nil,
			coverage: coverage, want: CoverageNotInstalled, wantRequired: "4.0.10",
		},
		{
			// The decisive case: this advisory is fixed only on the 4.x branch,
			// so a 6.6 shop cannot close it with the plugin at all.
			name:            "branch-only advisory is not backportable on an older line",
			shopwareVersion: "6.6.10.0",
			advisory:        SecurityAdvisory{ID: "CVE-2", GHSA: "GHSA-gq96-5pfx-f4vc"},
			ext:             plugin("3.0.16"), coverage: coverage, want: CoverageNotBackported,
		},
		{
			name:            "an unmapped Shopware line is unknown, never covered",
			shopwareVersion: "6.4.20.0", advisory: adv, ext: plugin("4.0.12"),
			coverage: coverage, want: CoverageUnknown,
		},
		{
			// The 4.x plugin does not contain the 6.6 line's backports, however
			// high its version number compares against the 3.x fix release.
			name:            "a plugin from a newer branch is never credited as coverage",
			shopwareVersion: "6.6.10.0", advisory: adv, ext: plugin("4.0.10"),
			coverage: coverage, want: CoverageUnknown,
		},
		{
			name:            "a plugin from an older branch is unknown, not outdated",
			shopwareVersion: "6.7.8.2", advisory: adv, ext: plugin("3.0.20"),
			coverage: coverage, want: CoverageUnknown,
		},
		{
			name:            "an unparsed branch is unknown, never not-backported",
			shopwareVersion: "6.7.8.2", advisory: adv, ext: plugin("4.0.12"),
			coverage: PluginCoverage{}, want: CoverageUnknown,
		},
		{
			// A branch parsed with zero references predates GHSA labelling, so
			// its silence is not evidence either way.
			name:            "a branch with no references is unknown",
			shopwareVersion: "6.7.8.2", advisory: adv, ext: plugin("4.0.12"),
			coverage: PluginCoverage{"4": 0}, want: CoverageUnknown,
		},
		{
			name:            "an advisory without a GHSA cannot be resolved",
			shopwareVersion: "6.7.8.2",
			advisory:        SecurityAdvisory{ID: "CVE-3", CVE: "CVE-3"},
			ext:             plugin("4.0.12"), coverage: coverage, want: CoverageUnknown,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := resolveByPlugin(tc.advisory, tc.shopwareVersion, tc.ext, index, tc.coverage)
			assert.Equal(t, tc.want, got.State)
			if tc.wantRequired != "" {
				assert.Equal(t, tc.wantRequired, got.RequiredVersion)
			}
		})
	}
}

// A covered advisory is reported as a green check naming the mitigation, not
// dropped: the reasoning must stay visible and must not degrade shop status.
func TestCheckSecurityReportsPluginMitigation(t *testing.T) {
	t.Parallel()

	input := Input{
		Config: ShopConfig{Version: "6.7.8.2"},
		Extensions: []Extension{
			{Name: swagPlatformSecurity, Active: true, Installed: true, Version: "4.0.12"},
		},
		SecurityAdvisories: []SecurityAdvisory{{
			ID: "CVE-M", Title: "Mitigated", CVE: "CVE-M", GHSA: "GHSA-xvhc-gm7j-mhmc", Severity: "high",
			Packages: []SecurityAdvisoryPackage{{PackageName: "shopware/core", AffectedVersions: "<6.7.10.1"}},
		}},
		SecurityPluginFixes:    []PluginFix{{GhsaID: "GHSA-XVHC-GM7J-MHMC", Branch: "4", PluginVersion: "4.0.10"}},
		SecurityPluginCoverage: PluginCoverage{"4": 26},
	}

	output := NewOutput(nil)
	checkSecurity(context.Background(), input, output)
	result := output.Result()

	require.Len(t, result.Checks, 1)
	assert.Equal(t, StatusGreen, result.Checks[0].Level)
	assert.Equal(t, "check.security.advisoryMitigated", result.Checks[0].MessageKey)
	assert.Equal(t, "4.0.10", result.Checks[0].MessageParams["securityPluginFixVersion"])
	assert.Equal(t, StatusGreen, result.Status, "a mitigated advisory must not degrade shop status")
}

// An outdated plugin is a warning naming the target version, not a bare error:
// the remedy is a plugin update rather than a core upgrade.
func TestCheckSecurityReportsOutdatedPlugin(t *testing.T) {
	t.Parallel()

	input := Input{
		Config: ShopConfig{Version: "6.7.8.2"},
		Extensions: []Extension{
			{Name: swagPlatformSecurity, Active: true, Installed: true, Version: "4.0.9"},
		},
		SecurityAdvisories: []SecurityAdvisory{{
			ID: "CVE-O", Title: "Outdated", CVE: "CVE-O", GHSA: "GHSA-xvhc-gm7j-mhmc", Severity: "high",
			Packages: []SecurityAdvisoryPackage{{PackageName: "shopware/core", AffectedVersions: "<6.7.10.1"}},
		}},
		SecurityPluginFixes:    []PluginFix{{GhsaID: "GHSA-XVHC-GM7J-MHMC", Branch: "4", PluginVersion: "4.0.10"}},
		SecurityPluginCoverage: PluginCoverage{"4": 26},
	}

	output := NewOutput(nil)
	checkSecurity(context.Background(), input, output)
	result := output.Result()

	require.Len(t, result.Checks, 1)
	assert.Equal(t, StatusYellow, result.Checks[0].Level)
	assert.Equal(t, "check.security.advisoryPluginOutdated", result.Checks[0].MessageKey)
}

// Without a synced backport map nothing may be claimed, so behaviour matches
// what it was before the plugin was modelled at all.
func TestCheckSecurityWithoutPluginMapReportsAsBefore(t *testing.T) {
	t.Parallel()

	input := Input{
		Config: ShopConfig{Version: "6.7.8.2"},
		Extensions: []Extension{
			{Name: swagPlatformSecurity, Active: true, Installed: true, Version: "4.0.12"},
		},
		SecurityAdvisories: []SecurityAdvisory{{
			ID: "CVE-U", Title: "Unknown", CVE: "CVE-U", GHSA: "GHSA-xvhc-gm7j-mhmc", Severity: "high",
			Packages: []SecurityAdvisoryPackage{{PackageName: "shopware/core", AffectedVersions: "<6.7.10.1"}},
		}},
	}

	output := NewOutput(nil)
	checkSecurity(context.Background(), input, output)
	result := output.Result()

	require.Len(t, result.Checks, 1)
	assert.Equal(t, StatusRed, result.Checks[0].Level)
	assert.Equal(t, "check.security.advisoryCve", result.Checks[0].MessageKey)
}

func TestCheckSecurityEmitsErrorsAndWarnings(t *testing.T) {
	t.Parallel()

	input := Input{
		Config: ShopConfig{Version: "6.7.5.0"},
		SecurityAdvisories: []SecurityAdvisory{
			{
				ID: "CVE-HIGH", Title: "High issue", CVE: "CVE-HIGH", Severity: "high",
				Link: "https://example.com/high",
				Packages: []SecurityAdvisoryPackage{
					{PackageName: "shopware/core", AffectedVersions: ">=6.7.0.0,<6.7.10.1"},
				},
			},
			{
				ID: "CVE-LOW", Title: "Low issue", CVE: "CVE-LOW", Severity: "low",
				Link: "https://example.com/low",
				Packages: []SecurityAdvisoryPackage{
					{PackageName: "shopware/core", AffectedVersions: ">=6.7.0.0,<6.7.10.1"},
				},
			},
		},
	}
	output := NewOutput(nil)
	checkSecurity(context.Background(), input, output)
	result := output.Result()

	require.Len(t, result.Checks, 2)
	assert.Equal(t, StatusRed, result.Status)

	byID := map[string]Check{}
	for _, c := range result.Checks {
		byID[c.ID] = c
	}
	assert.Equal(t, StatusRed, byID["security.CVE-HIGH"].Level)
	assert.Equal(t, StatusYellow, byID["security.CVE-LOW"].Level)
}

func TestSecurityCheckID(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "security.CVE-2026-1", securityCheckID(SecurityAdvisory{CVE: "cve-2026-1", ID: "x"}))
	assert.Equal(t, "security.GHSA-abc", securityCheckID(SecurityAdvisory{GHSA: "GHSA-abc", ID: "x"}))
	assert.Equal(t, "security.PKSA-1", securityCheckID(SecurityAdvisory{ID: "PKSA-1"}))
}

// Ingest stores "*" when Packagist omits an affected range. Both matchers must
// reject it: treated as a constraint it means "every version", which would
// report every shop as vulnerable on the strength of an incomplete row.
func TestAdvisoryMatchIgnoresWildcardConstraint(t *testing.T) {
	t.Parallel()

	adv := SecurityAdvisory{
		ID: "CVE-WILDCARD",
		Packages: []SecurityAdvisoryPackage{
			{PackageName: "shopware/core", AffectedVersions: "*"},
		},
	}

	_, _, matched := advisoryMatch("6.7.8.2", map[string]string{"shopware/core": "6.7.8.2"}, adv)
	assert.False(t, matched, "a wildcard range must never flag a shop")

	// A real range on the same package still matches.
	adv.Packages[0].AffectedVersions = ">=6.7.0.0,<6.7.10.1"
	pkg, installed, matched := advisoryMatch("6.7.8.2", map[string]string{"shopware/core": "6.7.8.2"}, adv)
	assert.True(t, matched)
	assert.Equal(t, "shopware/core", pkg)
	assert.Equal(t, "6.7.8.2", installed)
}
