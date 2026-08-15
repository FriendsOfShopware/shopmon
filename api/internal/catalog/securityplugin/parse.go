// Package securityplugin derives which SwagPlatformSecurity release backports
// which security advisory.
//
// The Shopware Security Plugin closes core vulnerabilities without a core
// upgrade. Its documentation states that "every fix in the plugin corresponds
// to a published security advisory and is identified by its GHSA id", and those
// ids appear in the store changelog Shopmon already syncs — so the mapping is a
// parsing problem, not an integration one.
package securityplugin

import (
	"regexp"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/version"
)

// ghsaRe matches a GHSA identifier anywhere in the changelog HTML: an anchor
// href, the anchor text, or bare prose.
//
// A regex over the raw text is deliberate. The store changelog is hand-written
// HTML with no stable structure, so matching the identifier itself survives an
// entry being reformatted from a list into a paragraph, which DOM traversal
// would not.
var ghsaRe = regexp.MustCompile(`(?i)GHSA-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}`)

// ChangelogEntry is one released plugin version and its changelog text.
type ChangelogEntry struct {
	Version    string
	Changelog  string
	ReleasedAt *time.Time
}

// Fix is the minimum plugin version on one branch that backports a GHSA.
type Fix struct {
	GhsaID        string
	Branch        string
	PluginVersion string
	ReleasedAt    *time.Time
}

// Coverage records what was parsed for a branch, so callers can distinguish a
// branch that genuinely backports nothing from one that was never parsed.
type Coverage struct {
	Branch           string
	MinParsedVersion string
	MaxParsedVersion string
	GhsaCount        int
}

// ExtractGHSAs returns the distinct GHSA ids in one changelog entry, upper-cased
// and in first-seen order. An anchor contributes its id twice (href and text),
// so de-duplication is required rather than incidental.
func ExtractGHSAs(changelog string) []string {
	matches := ghsaRe.FindAllString(changelog, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]bool, len(matches))
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		id := strings.ToUpper(match)
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

// Branch returns the plugin's major version, which identifies its release
// branch. The store returns changelogs for every branch regardless of the
// Shopware version queried, so the branch must be derived from the version
// string itself.
func Branch(pluginVersion string) string {
	pluginVersion = strings.TrimSpace(pluginVersion)
	if pluginVersion == "" {
		return ""
	}
	major, _, found := strings.Cut(pluginVersion, ".")
	if !found || major == "" {
		return ""
	}
	for _, r := range major {
		if r < '0' || r > '9' {
			return ""
		}
	}
	return major
}

// DeriveFixes folds changelog entries into one fix per (GHSA, branch), keeping
// the lowest version on each branch that names the GHSA.
//
// The minimum is the correct answer, not the maximum: a later entry such as
// "Improved the SVG validator for GHSA-xvhc-gm7j-mhmc" refines a fix shipped
// earlier, so taking the newest mention would tell shops to upgrade further
// than they need to. Ordering uses version comparison rather than slice order,
// so unordered input yields the same result.
func DeriveFixes(entries []ChangelogEntry) ([]Fix, []Coverage) {
	type key struct {
		ghsa   string
		branch string
	}

	best := make(map[key]Fix)
	coverage := make(map[string]*Coverage)

	for _, entry := range entries {
		branch := Branch(entry.Version)
		if branch == "" {
			continue
		}

		cov, ok := coverage[branch]
		if !ok {
			cov = &Coverage{Branch: branch}
			coverage[branch] = cov
		}
		if cov.MaxParsedVersion == "" || version.Compare(entry.Version, cov.MaxParsedVersion) > 0 {
			cov.MaxParsedVersion = entry.Version
		}

		for _, ghsa := range ExtractGHSAs(entry.Changelog) {
			cov.GhsaCount++
			if cov.MinParsedVersion == "" || version.Compare(entry.Version, cov.MinParsedVersion) < 0 {
				cov.MinParsedVersion = entry.Version
			}

			k := key{ghsa: ghsa, branch: branch}
			current, exists := best[k]
			if exists && version.Compare(current.PluginVersion, entry.Version) <= 0 {
				continue
			}
			best[k] = Fix{
				GhsaID:        ghsa,
				Branch:        branch,
				PluginVersion: entry.Version,
				ReleasedAt:    entry.ReleasedAt,
			}
		}
	}

	fixes := make([]Fix, 0, len(best))
	for _, fix := range best {
		fixes = append(fixes, fix)
	}

	coverages := make([]Coverage, 0, len(coverage))
	for _, cov := range coverage {
		coverages = append(coverages, *cov)
	}
	return fixes, coverages
}
