// Package version compares version strings (e.g. Shopware extension versions
// like "6.6.1.0") using github.com/shyim/go-version. It exposes a small
// string-friendly, error-tolerant helper shared by the scrape write path and the
// read/handler path so both filter changelogs with identical semantics.
package version

import (
	"regexp"
	"strings"

	goversion "github.com/shyim/go-version"
)

// distroSuffixRe matches Linux distribution package suffixes appended to
// runtime versions (e.g. PHP's "8.3.6-0ubuntu0.24.04.1" or
// "8.2.25-1+ubuntu22.04.1+deb.sury.org+1", or MariaDB's
// "10.11.8-MariaDB-0ubuntu0.24.04.1" or "5.5.5-10.11.8-MariaDB-0ubuntu0.24.04.1").
// Standard pre-release tags (alpha, beta, rc, dev, patch) are preserved.
var distroSuffixRe = regexp.MustCompile(`(?i)^(?:5\.5\.5-)?(?:mariadb-)?v?(\d+\.\d+(?:\.\d+)*)-(?:\d|ubuntu|deb|mariadb|el\d|alpine).*$`)

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	lower := strings.ToLower(v)
	switch {
	case strings.HasPrefix(lower, "mariadb "):
		v = strings.TrimSpace(v[len("mariadb "):])
	case strings.HasPrefix(lower, "mysql "):
		v = strings.TrimSpace(v[len("mysql "):])
	}
	if m := distroSuffixRe.FindStringSubmatch(v); m != nil {
		return m[1]
	}
	return v
}

// ParseDatabaseVersion splits a stored or reported database version string
// (such as "MySQL 8.0.36-0ubuntu0.24.04.1", "MariaDB 10.11.8-MariaDB-0ubuntu0.24.04.1",
// or a raw version string) into its advisory package name ("mysql" or "mariadb")
// and exact version string ("8.0.36-0ubuntu0.24.04.1" or
// "10.11.8-MariaDB-0ubuntu0.24.04.1"). Distribution suffixes are preserved for
// storage and display; Compare and Satisfies normalize them during evaluation.
func ParseDatabaseVersion(raw string) (string, string) {
	v := strings.TrimSpace(raw)
	if v == "" || strings.EqualFold(v, "unknown") {
		return "", ""
	}
	lower := strings.ToLower(v)
	switch {
	case strings.HasPrefix(lower, "mariadb "):
		ver := strings.TrimSpace(v[len("mariadb "):])
		if ver == "" || strings.EqualFold(ver, "unknown") {
			return "", ""
		}
		return "mariadb", ver
	case strings.HasPrefix(lower, "mysql "):
		ver := strings.TrimSpace(v[len("mysql "):])
		if ver == "" || strings.EqualFold(ver, "unknown") {
			return "", ""
		}
		return "mysql", ver
	case strings.Contains(lower, "mariadb") || strings.HasPrefix(v, "10.") || strings.HasPrefix(v, "11."):
		return "mariadb", v
	default:
		return "mysql", v
	}
}

// Compare compares two version strings via go-version. It returns 1 if a > b,
// -1 if a < b, and 0 if they are equal. If either value cannot be parsed it
// falls back to a plain string comparison so callers still get a deterministic
// ordering.
func Compare(a, b string) int {
	va, errA := goversion.NewVersion(normalizeVersion(a))
	vb, errB := goversion.NewVersion(normalizeVersion(b))
	if errA != nil || errB != nil {
		return strings.Compare(a, b)
	}
	return va.Compare(vb)
}

// Satisfies reports whether version matches a Composer constraint string
// (e.g. ">=6.7.0.0,<6.7.10.1" or "<6.6.10.18|>=6.7.0.0,<6.7.10.1").
// Invalid version or constraint strings return false with a non-nil error.
func Satisfies(version, constraint string) (bool, error) {
	return goversion.Satisfies(normalizeVersion(version), constraint)
}
