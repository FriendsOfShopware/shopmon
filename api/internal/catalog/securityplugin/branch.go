package securityplugin

import "strings"

// pluginBranchToShopware maps a SwagPlatformSecurity major version to the
// Shopware line it serves. Confirmed by probing the store's pluginsByName
// endpoint: 6.7 resolves to 4.x, 6.6 to 3.x, 6.5 to 2.x.
//
// This is a constant rather than an inference. The store returns changelogs for
// every branch regardless of the Shopware version queried, so nothing in the
// response identifies which line a given entry belongs to.
var pluginBranchToShopware = map[string]string{
	"4": "6.7",
	"3": "6.6",
	"2": "6.5",
}

// ShopwareBranch returns the Shopware line a plugin branch serves, or "" when
// the branch is unknown — a future 5.x, say. Callers must treat "" as unknown
// coverage rather than as "not backported": crediting an unmapped branch with
// fixes, or condemning it, would both be guesses about a shop's exposure.
func ShopwareBranch(pluginBranch string) string {
	return pluginBranchToShopware[pluginBranch]
}

// PluginBranchForShopware is the inverse: the plugin branch serving a Shopware
// version such as "6.7.8.2". Returns "" when the line is unknown.
func PluginBranchForShopware(shopwareVersion string) string {
	line := shopwareLine(shopwareVersion)
	if line == "" {
		return ""
	}
	for branch, served := range pluginBranchToShopware {
		if served == line {
			return branch
		}
	}
	return ""
}

// shopwareLine reduces a Shopware version to its major.minor line, which is the
// granularity at which both the plugin branches and GitHub's patched versions
// are organised.
func shopwareLine(shopwareVersion string) string {
	shopwareVersion = strings.TrimPrefix(strings.TrimSpace(shopwareVersion), "v")
	if shopwareVersion == "" {
		return ""
	}
	parts := strings.SplitN(shopwareVersion, ".", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return ""
	}
	return parts[0] + "." + parts[1]
}

// ShopwareLine exposes shopwareLine for callers that key data by Shopware line.
func ShopwareLine(shopwareVersion string) string {
	return shopwareLine(shopwareVersion)
}
