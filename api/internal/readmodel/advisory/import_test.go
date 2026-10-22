package advisory_test

import (
	"context"
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/ghsa"
	advisoryread "github.com/friendsofshopware/shopmon/api/internal/readmodel/advisory"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/friendsofshopware/shopmon/api/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ghsaFetcherStub struct {
	advisory *ghsa.Advisory
	err      error
}

func (s ghsaFetcherStub) FetchGHSA(context.Context, string) (*ghsa.Advisory, error) {
	return s.advisory, s.err
}

func fetchedStub() *ghsa.Advisory {
	published := time.Date(2026, 10, 19, 7, 55, 51, 0, time.UTC)
	score := 7.1
	return &ghsa.Advisory{
		GHSAID:      "GHSA-vj8m-rx9g-45fq",
		Title:       "Missing ACL enforcement permits unauthorized access to archived email",
		Description: "## Summary\n\nDetails",
		Severity:    "high",
		CVSSScore:   &score,
		HTMLURL:     "https://github.com/FriendsOfShopware/FroshPlatformMailArchive/security/advisories/GHSA-vj8m-rx9g-45fq",
		PublishedAt: &published,
		Packages: []ghsa.Package{
			{Name: "frosh/mail-platform-archive", Ecosystem: "composer", VulnerableVersionRange: ">= 2.0.0, <= 3.6.0", PatchedVersion: "3.7.0"},
			{Name: "unrelated", Ecosystem: "npm", VulnerableVersionRange: "<1.0.0", PatchedVersion: "1.0.0"},
		},
	}
}

func TestPreviewGHSAImportProjectsComposerPackagesOnly(t *testing.T) {
	env := testutil.Setup(t)
	svc := advisoryread.NewService(env.Pool, env.Queries, ghsaFetcherStub{advisory: fetchedStub()})

	preview, err := svc.PreviewGHSAImport(context.Background(), "GHSA-vj8m-rx9g-45fq")
	require.NoError(t, err)
	assert.Equal(t, "GHSA-vj8m-rx9g-45fq", preview.GhsaID)
	assert.Equal(t, "", preview.ExistingAdvisoryID)
	require.Len(t, preview.Packages, 1, "non-Composer packages can never match an SBOM and are dropped")
	assert.Equal(t, "frosh/mail-platform-archive", preview.Packages[0].PackageName)
	assert.Equal(t, "3.7.0", preview.Packages[0].PatchedVersion)
	require.NotNil(t, preview.Severity)
	assert.Equal(t, "high", *preview.Severity)
}

func TestImportGHSAIdentityAndLifecycle(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()
	env.SeedUser(t, "admin-1", "Admin User", "admin@example.com", "admin")
	svc := advisoryread.NewService(env.Pool, env.Queries, ghsaFetcherStub{advisory: fetchedStub()})

	input := advisoryread.CreateInput{
		Title:    "Missing ACL enforcement permits unauthorized access to archived email",
		GhsaID:   new("GHSA-vj8m-rx9g-45fq"),
		Severity: new("high"),
		Packages: []advisoryread.PackageInput{
			{PackageName: "frosh/mail-platform-archive", AffectedVersions: ">=2.0.0,<=3.6.0"},
		},
		CreatedBy: "admin-1",
	}

	view, err := svc.ImportGHSA(ctx, input)
	require.NoError(t, err)
	assert.Equal(t, "GHSA-vj8m-rx9g-45fq", view.Advisory.AdvisoryID, "the GHSA id is the stable advisory id")
	assert.Equal(t, "github", view.Advisory.Origin)
	assert.True(t, view.Advisory.DetailsSyncedAt.Valid, "imported content came from GitHub; enrichment must not churn it")

	// The preview now reports the import instead of offering a duplicate.
	preview, err := svc.PreviewGHSAImport(ctx, "GHSA-vj8m-rx9g-45fq")
	require.NoError(t, err)
	assert.Equal(t, "GHSA-vj8m-rx9g-45fq", preview.ExistingAdvisoryID)

	_, err = svc.ImportGHSA(ctx, input)
	assert.ErrorIs(t, err, advisoryread.ErrConflict)

	// Content stays editable while only the GHSA exists …
	newTitle := "Corrected title"
	_, err = svc.UpdateEnrichment(ctx, view.Advisory.AdvisoryID, advisoryread.EnrichmentUpdate{
		EnrichedBy: "admin-1",
		CoreFields: &advisoryread.CoreFieldsUpdate{Title: &newTitle},
	})
	require.NoError(t, err)

	// … the pending CVE links onto the same row …
	linked, err := svc.LinkOfficial(ctx, view.Advisory.AdvisoryID, advisoryread.LinkOfficialInput{
		Cve: new("CVE-2026-5555"),
	})
	require.NoError(t, err)
	assert.Equal(t, "GHSA-vj8m-rx9g-45fq", linked.Advisory.AdvisoryID)
	require.NotNil(t, linked.Advisory.Cve)
	assert.Equal(t, "CVE-2026-5555", *linked.Advisory.Cve)

	// … and from then on the sync owns the content.
	_, err = svc.UpdateEnrichment(ctx, view.Advisory.AdvisoryID, advisoryread.EnrichmentUpdate{
		EnrichedBy: "admin-1",
		CoreFields: &advisoryread.CoreFieldsUpdate{Title: &newTitle},
	})
	assert.ErrorIs(t, err, advisoryread.ErrConflict)
}

func TestImportGHSARequiresGHSAID(t *testing.T) {
	env := testutil.Setup(t)
	svc := advisoryread.NewService(env.Pool, env.Queries, nil)

	_, err := svc.ImportGHSA(context.Background(), advisoryread.CreateInput{
		Title: "No GHSA",
		Packages: []advisoryread.PackageInput{
			{PackageName: "frosh/mail-platform-archive", AffectedVersions: "<3.7.0"},
		},
		CreatedBy: "admin-1",
	})
	assert.ErrorIs(t, err, advisoryread.ErrInvalidInput)

	_, err = svc.PreviewGHSAImport(context.Background(), "GHSA-vj8m-rx9g-45fq")
	assert.ErrorIs(t, err, advisoryread.ErrGHSAImportUnavailable)
}

func TestPreviewGHSACombinesPackageRanges(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()
	env.SeedUser(t, "admin", "Admin", "admin@example.com", "admin")
	fetched := fetchedStub()
	fetched.Packages = []ghsa.Package{
		{Name: "vendor/package", Ecosystem: "composer", VulnerableVersionRange: ">=1.0.0,<1.2.0", PatchedVersion: "1.2.0"},
		{Name: "Vendor/Package", Ecosystem: "composer", VulnerableVersionRange: ">=2.0.0,<2.1.0", PatchedVersion: "2.1.0"},
		{Name: "vendor/package", Ecosystem: "composer", VulnerableVersionRange: ">=1.0.0,<1.2.0", PatchedVersion: "1.2.0"},
	}
	svc := advisoryread.NewService(env.Pool, env.Queries, ghsaFetcherStub{advisory: fetched})
	preview, err := svc.PreviewGHSAImport(ctx, fetched.GHSAID)
	require.NoError(t, err)
	require.Len(t, preview.Packages, 1)
	pkg := preview.Packages[0]
	require.Equal(t, ">=1.0.0,<1.2.0 || >=2.0.0,<2.1.0", pkg.AffectedVersions)
	require.Empty(t, pkg.PatchedVersion, "different branches must not produce misleading single-version advice")
	for installed, want := range map[string]bool{"1.1.0": true, "2.0.5": true, "1.5.0": false, "2.1.0": false} {
		matched, err := version.Satisfies(installed, pkg.AffectedVersions)
		require.NoError(t, err)
		assert.Equal(t, want, matched, installed)
	}
	imported, err := svc.ImportGHSA(ctx, advisoryread.CreateInput{
		Title: preview.Title, GhsaID: &preview.GhsaID, CreatedBy: "admin",
		Packages: []advisoryread.PackageInput{{PackageName: pkg.PackageName, AffectedVersions: pkg.AffectedVersions}},
	})
	require.NoError(t, err)
	require.Len(t, imported.Packages, 1)
	assert.Equal(t, pkg.AffectedVersions, imported.Packages[0].AffectedVersions)
}
