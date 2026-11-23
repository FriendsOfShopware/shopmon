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

func TestResolveByPluginStates(t *testing.T) {
	t.Parallel()

	const advisory = "GHSA-4frp-8844-m589"
	index := NewPluginFixIndex([]PluginFix{
		{GhsaID: advisory, Branch: "4", PluginVersion: "4.0.17"},
	})
	plugin := func(ver string, active bool) *Extension {
		return &Extension{Name: SecurityPluginName, Active: active, Installed: true, Version: ver}
	}

	cases := []struct {
		name string
		ghsa string
		ext  *Extension
		want CoverageState
	}{
		{"verified release covers the advisory", advisory, plugin("4.0.17", true), CoverageCovered},
		{"later release covers the advisory", advisory, plugin("4.0.18", true), CoverageCovered},
		{"earlier release needs an update", advisory, plugin("4.0.16", true), CoverageOutdated},
		{"undocumented backport stays unverified", "GHSA-ABCD-1234-WXYZ", plugin("4.0.17", true), CoverageUnverified},
		{"incompatible plugin branch", advisory, plugin("3.0.21", true), CoverageUnknown},
		{"plugin is inactive and a backport exists", advisory, plugin("4.0.17", false), CoverageNotInstalled},
		{"plugin is absent and a backport exists", advisory, nil, CoverageNotInstalled},
		{"advisory has no GHSA", "", plugin("4.0.17", true), CoverageUnknown},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, resolveByPlugin(tc.ghsa, "6.7.8.2", tc.ext, index).State)
		})
	}
}

// Without the extension list the installed SwagPlatformSecurity version is
// unknown, so no advisory can be judged. Emitting zero checks would read as
// "all clear", so the source is marked unavailable and the caller keeps the
// findings from the previous run.
func TestCheckSecurityUnavailableWithoutExtensions(t *testing.T) {
	t.Parallel()

	input := Input{
		Config:  ShopConfig{Version: "6.7.8.2"},
		Missing: MissingData{Extensions: true},
		SecurityAdvisories: []SecurityAdvisory{{
			ID: "CVE-U", Title: "Unjudgeable", CVE: "CVE-U", Severity: "high",
			Packages: []SecurityAdvisoryPackage{{PackageName: "shopware/core", AffectedVersions: "<6.7.10.1"}},
		}},
	}

	output := NewOutput(nil)
	checkSecurity(context.Background(), input, output)
	result := output.Result()

	assert.Empty(t, result.Checks)
	assert.Contains(t, result.Unavailable, UnavailableSource{Source: SourceSecurity, IDPrefix: prefixSecurity})
}

// A covered advisory is reported as a green check naming the mitigation, not
// dropped: the reasoning must stay visible and must not degrade shop status.
func TestCheckSecurityReportsPluginMitigation(t *testing.T) {
	t.Parallel()

	input := Input{
		Config: ShopConfig{Version: "6.7.8.2"},
		Extensions: []Extension{
			{Name: SecurityPluginName, Active: true, Installed: true, Version: "4.0.12"},
		},
		SecurityAdvisories: []SecurityAdvisory{{
			ID: "CVE-M", Title: "Mitigated", CVE: "CVE-M", GHSA: "GHSA-xvhc-gm7j-mhmc", Severity: "high",
			Packages: []SecurityAdvisoryPackage{{PackageName: "shopware/core", AffectedVersions: "<6.7.10.1"}},
		}},
		SecurityPluginFixes: []PluginFix{{GhsaID: "GHSA-XVHC-GM7J-MHMC", Branch: "4", PluginVersion: "4.0.10"}},
	}

	output := NewOutput(nil)
	checkSecurity(context.Background(), input, output)
	result := output.Result()

	require.Len(t, result.Checks, 1)
	assert.Equal(t, StatusGreen, result.Checks[0].Level)
	assert.Equal(t, "check.security.advisoryMitigated", result.Checks[0].MessageKey)
	assert.Equal(t, "4.0.12", result.Checks[0].MessageParams["securityPluginVersion"])
	assert.Equal(t, StatusGreen, result.Status, "a mitigated advisory must not degrade shop status")
}

// An outdated plugin is a warning naming the target version, not a bare error:
// the remedy is a plugin update rather than a core upgrade.
func TestCheckSecurityReportsOutdatedPlugin(t *testing.T) {
	t.Parallel()

	input := Input{
		Config: ShopConfig{Version: "6.7.8.2"},
		Extensions: []Extension{
			{Name: SecurityPluginName, Active: true, Installed: true, Version: "4.0.9"},
		},
		SecurityAdvisories: []SecurityAdvisory{{
			ID: "CVE-O", Title: "Outdated", CVE: "CVE-O", GHSA: "GHSA-xvhc-gm7j-mhmc", Severity: "high",
			Packages: []SecurityAdvisoryPackage{{PackageName: "shopware/core", AffectedVersions: "<6.7.10.1"}},
		}},
		SecurityPluginFixes: []PluginFix{{GhsaID: "GHSA-XVHC-GM7J-MHMC", Branch: "4", PluginVersion: "4.0.10"}},
	}

	output := NewOutput(nil)
	checkSecurity(context.Background(), input, output)
	result := output.Result()

	require.Len(t, result.Checks, 1)
	assert.Equal(t, StatusYellow, result.Checks[0].Level)
	assert.Equal(t, "check.security.advisoryPluginOutdated", result.Checks[0].MessageKey)
}

func TestCheckSecurityReportsUnverifiedPluginCoverage(t *testing.T) {
	t.Parallel()

	input := Input{
		Config: ShopConfig{Version: "6.7.8.2"},
		Extensions: []Extension{
			{Name: SecurityPluginName, Active: true, Installed: true, Version: "4.0.12"},
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
	assert.Equal(t, "check.security.advisoryPluginUnverified", result.Checks[0].MessageKey)
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

// A Shopmon-authored advisory must render with the marked message keys so
// users can tell a Shopmon early warning apart from an official CVE. The
// check id stays the stable Shopmon id while no CVE is linked.
func TestCheckSecurityMarksShopmonAdvisories(t *testing.T) {
	t.Parallel()

	input := Input{
		Config: ShopConfig{Version: "6.7.8.2"},
		SecurityAdvisories: []SecurityAdvisory{{
			ID: "SHOPMON-2026-0001", Title: "Early warning", Severity: "high", Origin: "shopmon",
			Packages: []SecurityAdvisoryPackage{{PackageName: "shopware/core", AffectedVersions: "<6.7.10.1"}},
		}},
	}

	output := NewOutput(nil)
	checkSecurity(context.Background(), input, output)
	result := output.Result()

	require.Len(t, result.Checks, 1)
	assert.Equal(t, "security.SHOPMON-2026-0001", result.Checks[0].ID)
	assert.Equal(t, "check.security.advisoryShopmon", result.Checks[0].MessageKey)
}
