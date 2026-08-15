package sbom

import (
	"log/slog"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/version"
)

// AdvisoryPackage is one package/constraint pair belonging to an advisory.
type AdvisoryPackage struct {
	AdvisoryID       string
	PackageName      string
	AffectedVersions string
}

// Match is a vulnerable package found in an environment's SBOM.
type Match struct {
	AdvisoryID       string
	PackageName      string
	InstalledVersion string
	AffectedVersions string
}

// Index groups advisory package constraints by (lowercased) package name so a
// full SBOM can be checked without scanning the whole advisory catalog per
// package. Build it once per rematch pass and reuse across environments.
type Index struct {
	byPackage map[string][]AdvisoryPackage
}

// NewIndex builds a lookup index from the advisory package catalog.
func NewIndex(packages []AdvisoryPackage) *Index {
	byPackage := make(map[string][]AdvisoryPackage, len(packages))
	for _, p := range packages {
		name := strings.ToLower(strings.TrimSpace(p.PackageName))
		if name == "" || strings.TrimSpace(p.AffectedVersions) == "" {
			continue
		}
		byPackage[name] = append(byPackage[name], p)
	}
	return &Index{byPackage: byPackage}
}

// PackageNames returns every package name the index knows about.
func (i *Index) PackageNames() []string {
	if i == nil {
		return nil
	}
	out := make([]string, 0, len(i.byPackage))
	for name := range i.byPackage {
		out = append(out, name)
	}
	return out
}

// Match returns every advisory that covers one of the installed packages.
// Installed is package name -> version, as produced by Document.VersionMap.
func (i *Index) Match(installed map[string]string) []Match {
	if i == nil || len(installed) == 0 {
		return nil
	}

	out := make([]Match, 0)
	for name, installedVersion := range installed {
		name = strings.ToLower(strings.TrimSpace(name))
		installedVersion = strings.TrimSpace(installedVersion)
		if name == "" || installedVersion == "" {
			continue
		}

		for _, adv := range i.byPackage[name] {
			if !affects(installedVersion, adv.AffectedVersions) {
				continue
			}
			out = append(out, Match{
				AdvisoryID:       adv.AdvisoryID,
				PackageName:      adv.PackageName,
				InstalledVersion: installedVersion,
				AffectedVersions: adv.AffectedVersions,
			})
		}
	}
	return out
}

// affects reports whether installedVersion falls inside a Composer affected
// version expression such as ">=6.7.0.0,<6.7.8.1" or "<6.6.10.18|>=6.7.0.0".
// version.Satisfies already understands "|"/"||" OR branches and comma/space
// separated AND constraints, so the whole expression is passed through as-is —
// the same call the existing core-version security check makes.
//
// An unparseable constraint yields false rather than a match, so a malformed
// advisory can never raise a false alarm against a customer shop.
func affects(installedVersion, affectedVersions string) bool {
	installedVersion = strings.TrimSpace(installedVersion)
	affectedVersions = strings.TrimSpace(affectedVersions)
	if installedVersion == "" || affectedVersions == "" {
		return false
	}

	// An empty or "*" constraint means "everything" to version.Satisfies, which
	// would flag every shop. Advisories always carry a real range, so treat it
	// as bad data.
	if affectedVersions == "*" {
		return false
	}

	ok, err := version.Satisfies(installedVersion, affectedVersions)
	if err != nil {
		slog.Debug("skip unparseable advisory constraint",
			"constraint", affectedVersions, "version", installedVersion, "error", err)
		return false
	}
	return ok
}
