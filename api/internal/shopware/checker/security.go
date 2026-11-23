package checker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/version"
)

// SecurityAdvisory is a catalog entry used to flag vulnerable Shopware
// installs. Populated from the local composer_advisory table (Packagist sync,
// plus advisories authored by Shopmon admins). One entry is one logical
// CVE/GHSA, which may cover multiple packages.
type SecurityAdvisory struct {
	ID       string
	Title    string
	CVE      string
	GHSA     string
	Link     string
	Severity string
	// Origin is "shopmon" for Shopmon-authored advisories and "upstream" for
	// synced ones; it selects the message keys that mark Shopmon reports.
	Origin   string
	Packages []SecurityAdvisoryPackage
}

// SecurityAdvisoryPackage is a package-specific version range for an advisory.
type SecurityAdvisoryPackage struct {
	PackageName      string
	AffectedVersions string
}

const (
	// shopwarePackagePrefix marks first-party Composer packages, whose version
	// is known from /_info/config even when no SBOM is available.
	shopwarePackagePrefix = "shopware/"
)

func checkSecurity(ctx context.Context, input Input, output *Output) {
	_ = ctx
	// Plugin resolution below needs the installed SwagPlatformSecurity version,
	// so without the extension list the advisories cannot be judged either way.
	// Emitting no checks would look like they were resolved, so the source is
	// marked unavailable and the caller keeps the previous findings.
	if input.Missing.Extensions {
		output.MarkUnavailable(SourceSecurity, prefixSecurity)
		return
	}

	if len(input.SecurityAdvisories) == 0 {
		return
	}

	shopwareVersion := strings.TrimSpace(input.Config.Version)
	if shopwareVersion == "" {
		return
	}

	// Evaluate documented backports against the active plugin's release.
	securityPlugin := FindSecurityPlugin(input.Extensions)
	fixIndex := NewPluginFixIndex(input.SecurityPluginFixes)

	matched := matchAdvisories(shopwareVersion, input.InstalledPackages, input.SecurityAdvisories)
	for _, adv := range matched {
		id := securityCheckID(adv.SecurityAdvisory)
		displayID := adv.CVE
		if displayID == "" {
			displayID = adv.GHSA
		}
		params := map[string]any{"title": adv.Title, "cve": displayID, "id": adv.ID}
		messageKey := advisoryMessageKey("check.security.advisory", adv.Origin)
		if adv.CVE != "" {
			messageKey = advisoryMessageKey("check.security.advisoryCve", adv.Origin)
		}
		// When the hit came from a dependency rather than the core itself, name
		// the package so the check is actionable.
		isDependency := adv.PackageName != "" && !isShopwarePackage(adv.PackageName)
		if isDependency {
			params["package"] = adv.PackageName
			params["installedVersion"] = adv.InstalledVersion
			messageKey = advisoryMessageKey("check.security.advisoryPackage", adv.Origin)
		}
		link := adv.Link

		// The plugin only backports core fixes, so a vulnerable dependency is
		// never resolved by it.
		resolution := PluginResolution{State: CoverageUnknown}
		if !isDependency {
			resolution = resolveByPlugin(adv.GHSA, shopwareVersion, securityPlugin, fixIndex)
		}

		switch resolution.State {
		case CoverageCovered:
			params["securityPluginVersion"] = resolution.InstalledVersion
			output.Success(id, advisoryMessageKey("check.security.advisoryMitigated", adv.Origin), params, SourceSecurity, link)
			continue
		case CoverageOutdated:
			// Exposed, but one plugin update away rather than a core upgrade.
			params["securityPluginVersion"] = resolution.InstalledVersion
			params["securityPluginFixVersion"] = resolution.RequiredVersion
			output.Warning(id, advisoryMessageKey("check.security.advisoryPluginOutdated", adv.Origin), params, SourceSecurity, link)
			continue
		case CoverageNotInstalled:
			params["securityPluginFixVersion"] = resolution.RequiredVersion
			messageKey = advisoryMessageKey("check.security.advisoryInstallPlugin", adv.Origin)
		case CoverageUnverified:
			messageKey = advisoryMessageKey("check.security.advisoryPluginUnverified", adv.Origin)
		}

		switch strings.ToLower(adv.Severity) {
		case "low", "none":
			output.Warning(id, messageKey, params, SourceSecurity, link)
		default:
			output.Error(id, messageKey, params, SourceSecurity, link)
		}
	}
}

// matchedAdvisory pairs an advisory with the specific package that triggered it,
// so the rendered check can name the vulnerable dependency.
type matchedAdvisory struct {
	SecurityAdvisory
	PackageName      string
	InstalledVersion string
}

// matchAdvisories returns advisories affecting this shop.
//
// When an SBOM is available (FroshTools exposes the Composer package
// inventory), every installed package is checked against its constraints, which
// catches vulnerabilities in dependencies like symfony/* or twig/*. Without an
// SBOM it degrades to the previous behaviour: shopware/* constraints evaluated
// against the Shopware core version only.
func matchAdvisories(shopwareVersion string, installed map[string]string, advisories []SecurityAdvisory) []matchedAdvisory {
	out := make([]matchedAdvisory, 0)
	for _, adv := range advisories {
		if pkg, installedVersion, ok := advisoryMatch(shopwareVersion, installed, adv); ok {
			out = append(out, matchedAdvisory{
				SecurityAdvisory: adv,
				PackageName:      pkg,
				InstalledVersion: installedVersion,
			})
		}
	}
	return out
}

// advisoryMatch prefers a vulnerable dependency when both it and a core package
// match, since a Security Plugin backport only mitigates the core package.
func advisoryMatch(shopwareVersion string, installed map[string]string, adv SecurityAdvisory) (string, string, bool) {
	var corePackage, coreVersion string
	for _, pkg := range adv.Packages {
		constraint := strings.TrimSpace(pkg.AffectedVersions)
		// Both empty and "*" mean "everything" to version.Satisfies, and "*" is
		// what ingest stores when Packagist omits the range. Treating it as a
		// real constraint would report every shop as vulnerable on the strength
		// of an incomplete advisory row. sbom.affects rejects it for the same
		// reason; the two matchers must agree.
		if constraint == "" || constraint == "*" {
			continue
		}

		// Prefer the version the SBOM reports for this exact package; fall back
		// to the core version for shopware/* so shops without an SBOM keep
		// getting core advisories.
		installedVersion, known := lookupInstalledVersion(installed, pkg.PackageName)
		if !known {
			if !isShopwarePackage(pkg.PackageName) {
				continue
			}
			installedVersion = shopwareVersion
		}
		if installedVersion == "" {
			continue
		}

		ok, err := version.Satisfies(installedVersion, constraint)
		if err != nil {
			slog.Debug("skip advisory package with unparseable constraint",
				"advisoryId", adv.ID, "package", pkg.PackageName,
				"constraint", constraint, "error", err)
			continue
		}
		if !ok {
			continue
		}
		if !isShopwarePackage(pkg.PackageName) {
			return pkg.PackageName, installedVersion, true
		}
		if corePackage == "" {
			corePackage, coreVersion = pkg.PackageName, installedVersion
		}
	}
	return corePackage, coreVersion, corePackage != ""
}

func isShopwarePackage(name string) bool {
	return strings.HasPrefix(strings.ToLower(name), shopwarePackagePrefix)
}

// lookupInstalledVersion resolves a package name against the SBOM inventory.
// Composer package names are case-insensitive, and the inventory is stored
// lowercased.
func lookupInstalledVersion(installed map[string]string, packageName string) (string, bool) {
	if len(installed) == 0 {
		return "", false
	}
	if v, ok := installed[packageName]; ok {
		return v, true
	}
	v, ok := installed[strings.ToLower(strings.TrimSpace(packageName))]
	return v, ok
}

// advisoryMessageKey picks the message key variant that marks a
// Shopmon-authored advisory as a Shopmon report — users must be able to tell
// a Shopmon early warning apart from an official CVE at a glance. Every
// advisory* locale key has a *Shopmon twin carrying that marker.
func advisoryMessageKey(base, origin string) string {
	if origin == "shopmon" {
		return base + "Shopmon"
	}
	return base
}

// SecurityCheckID is the stable check id for an advisory. It is exported so
// callers that suppress by advisory (rather than by check) can derive the same
// id without duplicating the CVE > GHSA > advisory-id precedence rule.
func SecurityCheckID(cve, ghsa, advisoryID string) string {
	return securityCheckID(SecurityAdvisory{ID: advisoryID, CVE: cve, GHSA: ghsa})
}

func securityCheckID(adv SecurityAdvisory) string {
	if cve := strings.TrimSpace(adv.CVE); cve != "" {
		return fmt.Sprintf("security.%s", strings.ToUpper(cve))
	}
	if ghsa := strings.TrimSpace(adv.GHSA); ghsa != "" {
		return fmt.Sprintf("security.%s", ghsa)
	}
	return fmt.Sprintf("security.%s", adv.ID)
}
