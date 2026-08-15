package advisory

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shyim/go-composer/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractGHSAID(t *testing.T) {
	t.Parallel()

	adv := repository.SecurityAdvisory{
		RemoteID: "ignored",
		Sources: []repository.AdvisorySource{
			{Name: "GitHub", RemoteID: "GHSA-gq96-5pfx-f4vc"},
		},
	}
	assert.Equal(t, "GHSA-gq96-5pfx-f4vc", extractGHSAID(adv))

	adv = repository.SecurityAdvisory{RemoteID: "GHSA-aaaa-bbbb-cccc"}
	assert.Equal(t, "GHSA-aaaa-bbbb-cccc", extractGHSAID(adv))

	adv = repository.SecurityAdvisory{RemoteID: "CVE-2026-1"}
	assert.Empty(t, extractGHSAID(adv))
}

func TestParseReportedAt(t *testing.T) {
	t.Parallel()

	ts, ok := parseReportedAt("2026-06-04 19:36:07")
	require.True(t, ok)
	assert.Equal(t, 2026, ts.Year())
	assert.Equal(t, time.June, ts.Month())
	assert.Equal(t, 4, ts.Day())

	_, ok = parseReportedAt("")
	assert.False(t, ok)

	_, ok = parseReportedAt("not-a-date")
	assert.False(t, ok)
}

func TestFilterShopwarePackages(t *testing.T) {
	t.Parallel()

	in := []string{"shopware/core", "acme/plugin", "shopware/platform"}
	assert.Equal(t, []string{"shopware/core", "shopware/platform"}, filterShopwarePackages(in))
}

func TestCanonicalAdvisoryID(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "CVE-2026-1", canonicalAdvisoryID("cve-2026-1", "GHSA-x", "PKSA-1"))
	assert.Equal(t, "GHSA-x", canonicalAdvisoryID("", "GHSA-x", "PKSA-1"))
	assert.Equal(t, "PKSA-1", canonicalAdvisoryID("", "", "PKSA-1"))
}

// A monitored shop's in-house Composer packages are the tenant's information.
// Only names Packagist already publishes may be sent to it, and publicness is
// established from vendor listings so an unpublished package name is never
// transmitted — not even to ask whether it exists.
func TestPublicPackagesOfExcludesPrivateNames(t *testing.T) {
	t.Parallel()

	var probed []string
	mux := http.NewServeMux()
	mux.HandleFunc("/packages/list.json", func(w http.ResponseWriter, r *http.Request) {
		vendor := r.URL.Query().Get("vendor")
		probed = append(probed, r.URL.RequestURI())
		switch vendor {
		case "symfony":
			_, _ = io.WriteString(w, `{"packageNames":["symfony/http-kernel","symfony/console"]}`)
		case "acme":
			// Vendor exists on Packagist, but this shop's package does not.
			_, _ = io.WriteString(w, `{"packageNames":["acme/public-widget"]}`)
		default:
			_, _ = io.WriteString(w, `{"packageNames":[]}`)
		}
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newPackagistClient(server.URL, defaultUserAgent, server.Client())
	got := client.publicPackagesOf(context.Background(), []string{
		"symfony/http-kernel",
		"acme/secret-billing",
		"acme/public-widget",
		"internal/tooling",
	})

	assert.ElementsMatch(t, []string{"symfony/http-kernel", "acme/public-widget"}, got)
	assert.NotContains(t, got, "acme/secret-billing")
	assert.NotContains(t, got, "internal/tooling")

	// The private package segment must not appear in any outbound request.
	for _, uri := range probed {
		assert.NotContains(t, uri, "secret-billing")
		assert.NotContains(t, uri, "tooling")
	}
}

// An unverifiable vendor is excluded rather than assumed public: the guarantee
// is that nothing unconfirmed is transmitted.
func TestPublicPackagesOfExcludesUnverifiableVendors(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/packages/list.json", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := newPackagistClient(server.URL, defaultUserAgent, server.Client())
	got := client.publicPackagesOf(context.Background(), []string{"symfony/http-kernel"})
	assert.Empty(t, got)
}
