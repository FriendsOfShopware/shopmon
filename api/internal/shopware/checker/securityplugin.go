package checker

import (
	"slices"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/shopware/sbom"
	"github.com/friendsofshopware/shopmon/api/internal/version"
)

const SecurityPluginName = "SwagPlatformSecurity"

// PluginFix is the minimum SwagPlatformSecurity version on one branch that
// backports a given advisory.
type PluginFix struct {
	GhsaID        string
	Branch        string
	PluginVersion string
}

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
	// CoverageUnverified: the installed plugin has no confirmed backport in the catalog.
	CoverageUnverified CoverageState = "unverified"
	// CoverageUnknown: the advisory cannot be evaluated against the plugin.
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

// PluginFixIndex indexes backports by GHSA and plugin branch.
type PluginFixIndex map[string]map[string]string

func NewPluginFixIndex(fixes []PluginFix) PluginFixIndex {
	index := make(PluginFixIndex, len(fixes))
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

// ExcludeMitigatedMatches removes first-party package matches covered by the
// installed security plugin. It reuses the backing array of matches.
func (index PluginFixIndex) ExcludeMitigatedMatches(matches []sbom.Match, shopwareVersion string, ext *Extension) []sbom.Match {
	if ext == nil || !ext.Active || !ext.Installed || ext.Version == "" {
		return matches
	}
	return slices.DeleteFunc(matches, func(match sbom.Match) bool {
		return isShopwarePackage(match.PackageName) &&
			resolveByPlugin(match.GHSA, shopwareVersion, ext, index).State == CoverageCovered
	})
}

// resolveByPlugin evaluates release coverage for an installed security plugin.
func resolveByPlugin(ghsa, shopwareVersion string, ext *Extension, index PluginFixIndex) PluginResolution {
	ghsa = strings.ToUpper(strings.TrimSpace(ghsa))
	branch := pluginBranchForShopware(shopwareVersion)
	if ghsa == "" || branch == "" {
		return PluginResolution{State: CoverageUnknown}
	}

	required := index[ghsa][branch]
	if ext == nil || !ext.Active || !ext.Installed || ext.Version == "" {
		if required != "" {
			return PluginResolution{State: CoverageNotInstalled, RequiredVersion: required}
		}
		return PluginResolution{State: CoverageUnknown}
	}
	if pluginMajor(ext.Version) != branch {
		return PluginResolution{State: CoverageUnknown}
	}
	if required == "" {
		return PluginResolution{State: CoverageUnverified}
	}

	covered, err := version.Satisfies(ext.Version, ">="+required)
	if err != nil {
		return PluginResolution{State: CoverageUnverified}
	}
	state := CoverageOutdated
	if covered {
		state = CoverageCovered
	}
	return PluginResolution{State: state, RequiredVersion: required, InstalledVersion: ext.Version}
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

func FindSecurityPlugin(extensions []Extension) *Extension {
	for i := range extensions {
		if extensions[i].Name == SecurityPluginName {
			return &extensions[i]
		}
	}
	return nil
}
