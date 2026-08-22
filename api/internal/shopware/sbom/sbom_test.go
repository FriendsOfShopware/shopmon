package sbom

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/friendsofshopware/shopmon/api/internal/shopware"
)

const sampleBOM = `{
  "bomFormat": "CycloneDX",
  "specVersion": "1.7",
  "serialNumber": "urn:uuid:1234",
  "version": 1,
  "metadata": {
    "timestamp": "2026-07-20T10:11:12+00:00",
    "component": {"type": "application", "bom-ref": "app:shopware/production", "name": "shopware/production"}
  },
  "components": [
    {"type": "library", "bom-ref": "pkg:composer/shopware/core@6.6.10.0", "group": "shopware", "name": "core", "version": "6.6.10.0", "purl": "pkg:composer/shopware/core@6.6.10.0"},
    {"type": "library", "group": "symfony", "name": "http-kernel", "version": "v6.4.2", "purl": "pkg:composer/symfony/http-kernel@v6.4.2"},
    {"type": "library", "group": "Twig", "name": "Twig", "version": "3.8.0"},
    {"type": "library", "name": "php", "version": "8.2.0"},
    {"type": "library", "group": "vendor", "name": "noversion"}
  ]
}`

type stubFetcher struct {
	body []byte
	err  error
}

func (s stubFetcher) Get(context.Context, string) ([]byte, error) {
	return s.body, s.err
}

func TestParsePackages(t *testing.T) {
	doc, err := Parse([]byte(sampleBOM))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if doc.SpecVersion != "1.7" {
		t.Errorf("specVersion = %q, want 1.7", doc.SpecVersion)
	}

	pkgs := doc.Packages()
	if len(pkgs) != 3 {
		t.Fatalf("got %d packages, want 3: %+v", len(pkgs), pkgs)
	}

	byName := map[string]Package{}
	for _, p := range pkgs {
		byName[p.Name] = p
	}

	if got := byName["shopware/core"].Version; got != "6.6.10.0" {
		t.Errorf("shopware/core version = %q", got)
	}
	// Leading "v" is stripped so versions compare consistently.
	if got := byName["symfony/http-kernel"].Version; got != "6.4.2" {
		t.Errorf("symfony/http-kernel version = %q, want 6.4.2", got)
	}
	// Composer names are case-insensitive; we normalize to lowercase.
	if _, ok := byName["twig/twig"]; !ok {
		t.Errorf("expected twig/twig, got %v", byName)
	}
	// Platform requirements (php, ext-*) carry no vendor and cannot be matched.
	if _, ok := byName["php"]; ok {
		t.Error("php should be dropped")
	}
	// Components without a version are unusable for constraint matching.
	if _, ok := byName["vendor/noversion"]; ok {
		t.Error("component without version should be dropped")
	}
}

// The SBOM is shop-controlled input: hostile or malformed component fields
// must be dropped at the parse boundary, never reach the database (a NUL byte
// aborts the whole scrape transaction) or the Packagist fan-out.
func TestPackagesRejectsHostileInput(t *testing.T) {
	doc := &Document{Components: []Component{
		{Group: "evil", Name: "nul\x00byte", Version: "1.0.0"},
		{Group: "evil", Name: "pkg", Version: "1.0\x00.0"},
		{Group: "a/b", Name: "c", Version: "1.0.0"},
		{Group: "  spaced  ", Name: " name ", Version: "1.0.0"},
		{Group: "vendor", Name: "in ner", Version: "1.0.0"},
		{Group: "long", Name: strings.Repeat("a", 300), Version: "1.0.0"},
		{Group: "vendor", Name: "longver", Version: strings.Repeat("9", 200)},
		{Group: "ok", Name: "pkg", Version: "2.0.0", Type: "lib\x00rary", Purl: "pkg:composer/ok/pkg@2.0.0"},
	}}

	pkgs := doc.Packages()
	if len(pkgs) != 2 {
		t.Fatalf("got %d packages, want 2: %+v", len(pkgs), pkgs)
	}

	byName := map[string]Package{}
	for _, p := range pkgs {
		byName[p.Name] = p
	}
	// Padded group/name segments trim to a valid single-slash name.
	if _, ok := byName["spaced/name"]; !ok {
		t.Errorf("expected spaced/name to survive trimming, got %+v", pkgs)
	}
	// Hostile metadata is dropped while the package itself survives.
	ok := byName["ok/pkg"]
	if ok.Type != "" {
		t.Errorf("Type with NUL should be dropped, got %q", ok.Type)
	}
	if ok.Purl != "pkg:composer/ok/pkg@2.0.0" {
		t.Errorf("valid purl should be kept, got %q", ok.Purl)
	}
}

// A bare "v" must not slip past the empty-version check and become an empty
// stored version that can never match an advisory.
func TestPackagesDropsBareVPrefixVersion(t *testing.T) {
	doc := &Document{Components: []Component{
		{Group: "vendor", Name: "pkg", Version: "v"},
	}}
	if pkgs := doc.Packages(); len(pkgs) != 0 {
		t.Fatalf("version \"v\" should be dropped, got %+v", pkgs)
	}
}

func TestSanitizeError(t *testing.T) {
	got := SanitizeError("bad\x00body\ttext")
	if strings.ContainsRune(got, 0) || strings.ContainsRune(got, '\t') {
		t.Errorf("control characters not stripped: %q", got)
	}
	long := SanitizeError(strings.Repeat("ä", 600))
	if len(long) > 510 {
		t.Errorf("not truncated: %d bytes", len(long))
	}
	if !utf8.ValidString(long) {
		t.Errorf("truncation broke UTF-8: %q", long)
	}
}

func TestVersionMapAndGeneratedAt(t *testing.T) {
	doc, err := Parse([]byte(sampleBOM))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	vm := doc.VersionMap()
	if vm["shopware/core"] != "6.6.10.0" {
		t.Errorf("VersionMap[shopware/core] = %q", vm["shopware/core"])
	}

	ts, ok := doc.GeneratedAt()
	if !ok {
		t.Fatal("GeneratedAt: expected timestamp")
	}
	if ts.Year() != 2026 || ts.Month() != 7 || ts.Day() != 20 {
		t.Errorf("GeneratedAt = %v", ts)
	}
}

// The endpoint answered, so the shop is not "unsupported" — the payload is
// wrong. A real error keeps supported=true and records last_error instead of
// silently filing the shop away as incapable.
func TestParseRejectsNonCycloneDXAsError(t *testing.T) {
	_, err := Parse([]byte(`{"bomFormat":"SPDX","components":[]}`))
	if err == nil {
		t.Fatal("expected an error for a non-CycloneDX document")
	}
	if _, ok := errors.AsType[*ErrUnsupported](err); ok {
		t.Fatalf("err = %v, want a plain parse error, not ErrUnsupported", err)
	}
}

func TestParseRejectsOversizedPayload(t *testing.T) {
	_, err := Parse(make([]byte, maxDocumentBytes+1))
	if err == nil {
		t.Fatal("expected an error for an oversized payload")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("err = %v, want a size error", err)
	}
}

// Only 404 means "this FroshTools predates the SBOM route". A 403 is a missing
// frosh_tools:read ACL — fixable configuration that must surface as an error,
// not be hidden as an absent capability.
func TestFetchTreatsOnly404AsUnsupported(t *testing.T) {
	client := stubFetcher{err: &shopware.ApiError{StatusCode: http.StatusNotFound, Body: "nope"}}
	_, err := Fetch(context.Background(), client)
	var unsupported *ErrUnsupported
	if !errors.As(err, &unsupported) {
		t.Errorf("404: err = %v, want ErrUnsupported", err)
	}

	client = stubFetcher{err: &shopware.ApiError{StatusCode: http.StatusForbidden, Body: "missing acl"}}
	_, err = Fetch(context.Background(), client)
	if err == nil || errors.As(err, &unsupported) {
		t.Errorf("403: err = %v, want a real error so the ACL problem is visible", err)
	}
}

func TestFetchPropagatesRealErrors(t *testing.T) {
	client := stubFetcher{err: &shopware.ApiError{StatusCode: http.StatusInternalServerError, Body: "boom"}}
	_, err := Fetch(context.Background(), client)

	var unsupported *ErrUnsupported
	if err == nil || errors.As(err, &unsupported) {
		t.Fatalf("err = %v, want a real error", err)
	}
}

func TestFetchParsesBody(t *testing.T) {
	doc, err := Fetch(context.Background(), stubFetcher{body: []byte(sampleBOM)})
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(doc.Packages()) != 3 {
		t.Errorf("got %d packages", len(doc.Packages()))
	}
}
