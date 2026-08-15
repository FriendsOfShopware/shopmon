// Package version compares version strings (e.g. Shopware extension versions
// like "6.6.1.0") using github.com/shyim/go-version. It exposes a small
// string-friendly, error-tolerant helper shared by the scrape write path and the
// read/handler path so both filter changelogs with identical semantics.
package version

import (
	"strings"

	goversion "github.com/shyim/go-version"
)

// Compare compares two version strings via go-version. It returns 1 if a > b,
// -1 if a < b, and 0 if they are equal. If either value cannot be parsed it
// falls back to a plain string comparison so callers still get a deterministic
// ordering.
func Compare(a, b string) int {
	va, errA := goversion.NewVersion(a)
	vb, errB := goversion.NewVersion(b)
	if errA != nil || errB != nil {
		return strings.Compare(a, b)
	}
	return va.Compare(vb)
}

// Satisfies reports whether version matches a Composer constraint string
// (e.g. ">=6.7.0.0,<6.7.10.1" or "<6.6.10.18|>=6.7.0.0,<6.7.10.1").
// Invalid version or constraint strings return false with a non-nil error.
func Satisfies(version, constraint string) (bool, error) {
	return goversion.Satisfies(version, constraint)
}
