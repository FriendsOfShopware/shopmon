package checker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/version"
)

// SecurityAdvisory is a catalog entry used to flag vulnerable Shopware
// installs. Populated from the local composer_advisory table (Packagist sync).
// One entry is one logical CVE/GHSA, which may cover multiple packages.
type SecurityAdvisory struct {
	ID       string
	Title    string
	CVE      string
	GHSA     string
	Link     string
	Severity string
	Packages []SecurityAdvisoryPackage
}

// SecurityAdvisoryPackage is a package-specific version range for an advisory.
type SecurityAdvisoryPackage struct {
	PackageName      string
	AffectedVersions string
}

const (
	swagPlatformSecurity = "SwagPlatformSecurity"
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

	// SwagPlatformSecurity backports individual core fixes. Each is resolved
	// against the installed plugin version rather than suppressing every
	// advisory whenever the plugin happens to be current — that older rule both
	// hid advisories the plugin does not fix and showed ones it does.
	securityPlugin := findExtension(input.Extensions, swagPlatformSecurity)
	fixIndex := buildPluginFixIndex(input.SecurityPluginFixes)

	matched := matchAdvisories(shopwareVersion, input.InstalledPackages, input.SecurityAdvisories)
	for _, adv := range matched {
		id := securityCheckID(adv.SecurityAdvisory)
		params := map[string]any{"title": adv.Title, "cve": adv.CVE}
		messageKey := "check.security.advisory"
		if adv.CVE != "" {
			messageKey = "check.security.advisoryCve"
		}
		// When the hit came from a dependency rather than the core itself, name
		// the package so the check is actionable.
		isDependency := adv.PackageName != "" && !strings.HasPrefix(adv.PackageName, shopwarePackagePrefix)
		if isDependency {
			params["package"] = adv.PackageName
			params["installedVersion"] = adv.InstalledVersion
			messageKey = "check.security.advisoryPackage"
		}
		link := adv.Link

		// The plugin only backports core fixes, so a vulnerable dependency is
		// never resolved by it.
		resolution := PluginResolution{State: CoverageUnknown}
		if !isDependency {
			resolution = resolveByPlugin(adv.SecurityAdvisory, shopwareVersion, securityPlugin, fixIndex, input.SecurityPluginCoverage)
		}

		switch resolution.State {
		case CoverageCovered:
			// Report the mitigation instead of dropping the row: a green check
			// naming the plugin version is auditable, a missing row is not.
			params["securityPluginVersion"] = resolution.InstalledVersion
			params["securityPluginFixVersion"] = resolution.RequiredVersion
			output.Success(id, "check.security.advisoryMitigated", params, SourceSecurity, link)
			continue
		case CoverageOutdated:
			// Exposed, but one plugin update away rather than a core upgrade.
			params["securityPluginVersion"] = resolution.InstalledVersion
			params["securityPluginFixVersion"] = resolution.RequiredVersion
			output.Warning(id, "check.security.advisoryPluginOutdated", params, SourceSecurity, link)
			continue
		case CoverageNotInstalled:
			params["securityPluginFixVersion"] = resolution.RequiredVersion
			messageKey = "check.security.advisoryInstallPlugin"
		case CoverageNotBackported:
			messageKey = "check.security.advisoryNotBackported"
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

// advisoryMatch reports the first package of adv that is installed at a
// vulnerable version, together with that version.
func advisoryMatch(shopwareVersion string, installed map[string]string, adv SecurityAdvisory) (string, string, bool) {
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
			if !strings.HasPrefix(pkg.PackageName, shopwarePackagePrefix) {
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
		if ok {
			return pkg.PackageName, installedVersion, true
		}
	}
	return "", "", false
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
