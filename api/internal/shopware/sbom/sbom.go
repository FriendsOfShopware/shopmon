// Package sbom fetches and parses the CycloneDX software bill of materials
// exposed by the FroshTools plugin, giving Shopmon the full Composer package
// inventory of a shop instead of just its Shopware core version.
package sbom

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/friendsofshopware/shopmon/api/internal/shopware"
)

// Path is the FroshTools SBOM endpoint. It is served by newer FroshTools
// releases only; older plugins answer 404 and callers must treat that as
// "unsupported", not as a scrape failure.
const Path = "/_action/frosh-tools/security/sbom"

// Fetcher is the subset of the Shopware admin client this package needs.
type Fetcher interface {
	Get(ctx context.Context, path string) ([]byte, error)
}

// Document is the parsed CycloneDX BOM. Only the fields Shopmon consumes are
// modelled; CycloneDX is additive, so unknown fields are ignored.
type Document struct {
	BomFormat    string      `json:"bomFormat"`
	SpecVersion  string      `json:"specVersion"`
	SerialNumber string      `json:"serialNumber"`
	Metadata     Metadata    `json:"metadata"`
	Components   []Component `json:"components"`
}

type Metadata struct {
	Timestamp string `json:"timestamp"`
}

// Component is one CycloneDX component. For Composer packages produced by
// FroshTools, Group holds the vendor and Name the package, so the Composer
// name is "group/name".
type Component struct {
	Type    string `json:"type"`
	BomRef  string `json:"bom-ref"`
	Group   string `json:"group"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Purl    string `json:"purl"`
}

// Package is a flattened Composer package from the SBOM.
type Package struct {
	Name    string
	Version string
	Type    string
	Purl    string
}

// ErrUnsupported reports that the shop's FroshTools does not expose the SBOM
// endpoint. Callers should record this and skip SBOM ingestion for the shop
// rather than failing the whole scrape.
type ErrUnsupported struct {
	Reason string
}

func (e *ErrUnsupported) Error() string {
	return fmt.Sprintf("sbom endpoint unsupported: %s", e.Reason)
}

// Fetch retrieves and parses the SBOM for one shop.
func Fetch(ctx context.Context, client Fetcher) (*Document, error) {
	if client == nil {
		return nil, &ErrUnsupported{Reason: "no client"}
	}

	raw, err := client.Get(ctx, Path)
	if err != nil {
		// Only a missing route means "this shop cannot serve an SBOM". A 403
		// means the integration lacks the frosh_tools:read ACL — a fixable
		// configuration problem that must surface as an error with last_error,
		// not be filed away as a permanent capability gap nobody is told about.
		if isMissingRoute(err) {
			return nil, &ErrUnsupported{Reason: "endpoint not found"}
		}
		return nil, fmt.Errorf("fetch sbom: %w", err)
	}

	return Parse(raw)
}

// maxDocumentBytes caps the SBOM payload. The document is shop-controlled and
// the admin client reads the body unbounded, so a hostile or runaway response
// could otherwise allocate freely here and again when flattening components.
// A real Shopware SBOM is a few hundred KB.
const maxDocumentBytes = 8 << 20

// Parse decodes a CycloneDX JSON document.
func Parse(raw []byte) (*Document, error) {
	if len(raw) > maxDocumentBytes {
		return nil, fmt.Errorf("parse sbom: payload too large (%d bytes)", len(raw))
	}
	var doc Document
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse sbom: %w", err)
	}
	if !strings.EqualFold(doc.BomFormat, "CycloneDX") {
		// The endpoint answered, so the shop is not "unsupported" — the payload
		// is wrong. A real error keeps supported=true and records last_error,
		// so a broken proxy or unexpected body is visible instead of looking
		// like an absent plugin.
		return nil, fmt.Errorf("parse sbom: unexpected bom format %q", doc.BomFormat)
	}
	return &doc, nil
}

// Packages flattens the document's components into Composer packages, dropping
// entries without a usable name or version. Names are lowercased because
// Composer package names are case-insensitive and advisory matching keys on
// them.
func (d *Document) Packages() []Package {
	if d == nil {
		return nil
	}

	out := make([]Package, 0, len(d.Components))
	seen := make(map[string]bool, len(d.Components))

	for _, c := range d.Components {
		name := c.composerName()
		version := strings.TrimPrefix(strings.TrimSpace(c.Version), "v")
		if name == "" || !validVersion(version) {
			continue
		}
		if seen[name] {
			continue
		}
		seen[name] = true

		out = append(out, Package{
			Name:    name,
			Version: version,
			Type:    SanitizeMeta(c.Type),
			Purl:    SanitizeMeta(c.Purl),
		})
	}
	return out
}

// VersionMap returns package name -> installed version, the shape expected by
// go-composer's Advisories.Affecting.
func (d *Document) VersionMap() map[string]string {
	pkgs := d.Packages()
	out := make(map[string]string, len(pkgs))
	for _, p := range pkgs {
		out[p.Name] = p.Version
	}
	return out
}

// GeneratedAt parses the BOM metadata timestamp, reporting false when absent or
// malformed.
func (d *Document) GeneratedAt() (time.Time, bool) {
	if d == nil {
		return time.Time{}, false
	}
	ts := strings.TrimSpace(d.Metadata.Timestamp)
	if ts == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return time.Time{}, false
	}
	return t.UTC(), true
}

// composerNameRe is Composer's own package-name rule: exactly one vendor
// segment and one package segment of lowercase word characters. The SBOM is
// shop-controlled input, and names that fail this rule — platform requirements,
// multi-segment paths, control characters — cannot be matched against
// advisories and must not reach the database or the Packagist fan-out.
var composerNameRe = regexp.MustCompile(`^[a-z0-9]([_.-]?[a-z0-9]+)*/[a-z0-9](([_.]|-{1,2})?[a-z0-9]+)*$`)

const (
	maxPackageNameLen = 190
	maxVersionLen     = 100
	maxMetaLen        = 512
)

func (c Component) composerName() string {
	group := strings.TrimSpace(c.Group)
	name := strings.TrimSpace(c.Name)
	if name == "" {
		return ""
	}
	full := name
	if group != "" {
		full = group + "/" + name
	} else if !strings.Contains(name, "/") {
		// Composer packages are always vendor/name; a groupless component is a
		// platform requirement (php, ext-*) we cannot match advisories against.
		return ""
	}
	full = strings.ToLower(full)
	if len(full) > maxPackageNameLen || !composerNameRe.MatchString(full) {
		return ""
	}
	return full
}

// validVersion accepts only bounded, printable version strings. The version is
// written verbatim into the database, so hostile bytes (NUL aborts the whole
// scrape transaction in Postgres) are rejected here at the trust boundary.
func validVersion(v string) bool {
	if v == "" || len(v) > maxVersionLen || !utf8.ValidString(v) {
		return false
	}
	for _, r := range v {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}

// SanitizeMeta bounds optional shop-supplied metadata (type, purl, spec
// version, serial number); anything unprintable or oversized is dropped rather
// than stored.
func SanitizeMeta(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > maxMetaLen || !utf8.ValidString(s) {
		return ""
	}
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return ""
		}
	}
	return s
}

// SanitizeError bounds error text that may embed a shop's response body before
// it lands in a user-visible column: invalid UTF-8 and control characters are
// stripped (a NUL would abort the whole Postgres transaction), and the result
// is truncated on a rune boundary.
func SanitizeError(s string) string {
	const maxErrLen = 500
	s = strings.ToValidUTF8(s, "�")
	s = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return ' '
		}
		return r
	}, s)
	if len(s) > maxErrLen {
		cut := maxErrLen
		for cut > 0 && !utf8.RuneStart(s[cut]) {
			cut--
		}
		s = s[:cut] + "…"
	}
	return strings.TrimSpace(s)
}

// isMissingRoute reports whether the shop answered 404, which is how a
// FroshTools release predating the SBOM endpoint replies. Deliberately not 403:
// a permission failure is a configuration problem an operator can fix, and
// classifying it as an absent capability would hide it behind a Debug log while
// package-level advisory coverage stays silently disabled.
func isMissingRoute(err error) bool {
	if apiErr, ok := errors.AsType[*shopware.ApiError](err); ok {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}
