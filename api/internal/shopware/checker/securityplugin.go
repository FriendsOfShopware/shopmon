package checker

import (
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/version"
)

// PluginFix is the minimum SwagPlatformSecurity version on one branch that
// backports a given advisory.
type PluginFix struct {
	GhsaID        string
	Branch        string
	PluginVersion string
}

// PluginCoverage records which plugin branches have been parsed, and how many
// advisory references each yielded. A branch with zero references was still
// parsed; a branch that is absent never was. The checker must not conflate them.
type PluginCoverage map[string]int

// CoverageState is how an advisory relates to the security plugin for one shop.
type CoverageState string

const (
	// CoverageCovered: the installed plugin backports this advisory. The check
	// is reported as a success naming the mitigation rather than being dropped,
	// so the reasoning stays visible and auditable.
	CoverageCovered CoverageState = "covered"
	// CoverageOutdated: the plugin backports it, but only from a later release
	// than the one installed. The shop is exposed, yet the fix is a plugin
	// update rather than a core upgrade.
	CoverageOutdated CoverageState = "outdated"
	// CoverageNotInstalled: a backport exists for this shop's line but the
	// plugin is absent or inactive. Installing it is a valid remediation.
	CoverageNotInstalled CoverageState = "notInstalled"
	// CoverageNotBackported: the branch was parsed and does not carry this
	// advisory, so only a core upgrade closes it.
	CoverageNotBackported CoverageState = "notBackported"
	// CoverageUnknown: no GHSA, an unmapped Shopware line, or an unparsed
	// branch. Nothing can be claimed, so the advisory is reported as before.
	CoverageUnknown CoverageState = "unknown"
)

// PluginResolution is the outcome of evaluating one advisory against the
// security plugin.
type PluginResolution struct {
	State CoverageState
	// RequiredVersion is the minimum plugin version backporting the advisory on
	// this shop's branch, when one is known.
	RequiredVersion string
	// InstalledVersion is the plugin version the shop currently runs.
	InstalledVersion string
}

// pluginFixIndex is fixes keyed by GHSA then branch.
type pluginFixIndex map[string]map[string]string

func buildPluginFixIndex(fixes []PluginFix) pluginFixIndex {
	index := make(pluginFixIndex, len(fixes))
	for _, fix := range fixes {
		ghsa := strings.ToUpper(strings.TrimSpace(fix.GhsaID))
		if ghsa == "" || fix.Branch == "" {
			continue
		}
		if index[ghsa] == nil {
			index[ghsa] = map[string]string{}
		}
		index[ghsa][fix.Branch] = fix.PluginVersion
	}
	return index
}

// resolveByPlugin decides how the security plugin relates to one advisory.
//
// Every ambiguous branch resolves toward reporting rather than suppressing.
// Claiming a shop is protected is a statement about its security posture, and
// the cost of being wrong is asymmetric: a redundant warning is noise, while a
// wrongly suppressed advisory hides a live vulnerability.
func resolveByPlugin(
	adv SecurityAdvisory,
	shopwareVersion string,
	ext *Extension,
	index pluginFixIndex,
	coverage PluginCoverage,
) PluginResolution {
	ghsa := strings.ToUpper(strings.TrimSpace(adv.GHSA))
	if ghsa == "" {
		// Without a GHSA there is no join key to the plugin's changelog. Do not
		// infer one from the CVE: a wrong join here would suppress a real
		// vulnerability.
		return PluginResolution{State: CoverageUnknown}
	}

	branch := pluginBranchForShopware(shopwareVersion)
	if branch == "" {
		return PluginResolution{State: CoverageUnknown}
	}

	// A branch we have never parsed tells us nothing. An old plugin release
	// whose changelogs predate GHSA labelling likewise cannot be credited.
	parsedRefs, parsed := coverage[branch]
	if !parsed || parsedRefs == 0 {
		return PluginResolution{State: CoverageUnknown}
	}

	required, hasFix := index[ghsa][branch]
	if !hasFix {
		return PluginResolution{State: CoverageNotBackported}
	}

	if ext == nil || !ext.Active || !ext.Installed || ext.Version == "" {
		return PluginResolution{State: CoverageNotInstalled, RequiredVersion: required}
	}

	// The plugin's major version is its branch. A plugin from another line — a
	// 4.x build on a 6.6 shop, say — does not contain this branch's backports,
	// however high its version number compares, so a cross-branch comparison
	// must never be credited as coverage.
	if pluginMajor(ext.Version) != branch {
		return PluginResolution{State: CoverageUnknown}
	}

	if version.Compare(ext.Version, required) >= 0 {
		return PluginResolution{
			State:            CoverageCovered,
			RequiredVersion:  required,
			InstalledVersion: ext.Version,
		}
	}
	return PluginResolution{
		State:            CoverageOutdated,
		RequiredVersion:  required,
		InstalledVersion: ext.Version,
	}
}

// pluginBranchForShopware maps a Shopware version onto the plugin branch that
// serves it. Kept in sync with internal/catalog/securityplugin; duplicated here
// rather than imported so the checker stays free of catalog dependencies.
func pluginBranchForShopware(shopwareVersion string) string {
	shopwareVersion = strings.TrimPrefix(strings.TrimSpace(shopwareVersion), "v")
	parts := strings.SplitN(shopwareVersion, ".", 3)
	if len(parts) < 2 {
		return ""
	}
	switch parts[0] + "." + parts[1] {
	case "6.7":
		return "4"
	case "6.6":
		return "3"
	case "6.5":
		return "2"
	default:
		return ""
	}
}

// pluginMajor extracts the branch a plugin release belongs to: its major
// version, tolerating a leading "v".
func pluginMajor(pluginVersion string) string {
	pluginVersion = strings.TrimPrefix(strings.TrimSpace(pluginVersion), "v")
	if i := strings.IndexByte(pluginVersion, '.'); i > 0 {
		return pluginVersion[:i]
	}
	return pluginVersion
}

func findExtension(extensions []Extension, name string) *Extension {
	for i := range extensions {
		if extensions[i].Name == name {
			return &extensions[i]
		}
	}
	return nil
}
