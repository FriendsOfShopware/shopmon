// Package ghsa carries the shared vocabulary for advisories fetched from
// GitHub: the projected advisory shape the catalog's GitHub client produces
// and the advisory read model consumes, plus the sentinel errors both sides
// agree on. It is a leaf package on purpose — the catalog and the read model
// must not import each other (their test packages meet in testutil).
package ghsa

import (
	"errors"
	"time"
)

// ErrInvalidReference marks an input that is neither a GHSA id nor a GitHub
// advisory URL.
var ErrInvalidReference = errors.New("invalid GHSA reference")

// ErrAdvisoryNotFound marks a GHSA id GitHub does not know (or has not
// published).
var ErrAdvisoryNotFound = errors.New("github advisory not found")

// CWE is one weakness entry of a fetched GitHub advisory.
type CWE struct {
	ID   string
	Name string
}

// Package is one affected package of a fetched GitHub advisory, as published:
// the range keeps GitHub's spelling and the patched version is the advisory's
// own claim, both for the admin to review before import.
type Package struct {
	Name                   string
	Ecosystem              string
	VulnerableVersionRange string
	PatchedVersion         string
}

// Advisory is a GitHub security advisory projected for admin import:
// everything the create form needs, nothing persisted yet.
type Advisory struct {
	GHSAID      string
	CVE         string
	Title       string
	Description string
	// Severity is GitHub's level (critical/high/medium/low), lowercased.
	Severity    string
	CVSSScore   *float64
	CVSSVector  *string
	CWEs        []CWE
	References  []string
	HTMLURL     string
	PublishedAt *time.Time
	Withdrawn   bool
	Packages    []Package
}
