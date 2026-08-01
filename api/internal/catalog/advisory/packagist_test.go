package advisory

import (
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
