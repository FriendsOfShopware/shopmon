package scrape

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/catalog/storemodel"
	"github.com/friendsofshopware/shopmon/api/internal/shopwareaccount"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtensionDiffRoundTrip guards against the field-name / type mismatch that
// previously caused plugin update changelogs to lose their version transition
// (from->to) and changelog text: the diff is serialized with the internal
// extensionDiff struct on write, but deserialized into api.ExtensionDiff on
// read. The JSON tags and field types of both sides must stay aligned.
func TestExtensionDiffRoundTrip(t *testing.T) {
	old := "5.0.0"
	updated := "5.1.0"
	created := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	written := []extensionDiff{
		{
			Name:       "SwagPayPal",
			Label:      "PayPal",
			State:      "updated",
			OldVersion: &old,
			NewVersion: &updated,
			Active:     true,
			Changelog: []extensionChangelog{
				{
					Version:      "5.1.0",
					Text:         "<p>Fixed a checkout bug</p>",
					CreationDate: created,
				},
			},
		},
	}

	raw, err := json.Marshal(written)
	require.NoError(t, err)

	var read []api.ExtensionDiff
	require.NoError(t, json.Unmarshal(raw, &read))
	require.Len(t, read, 1)

	got := read[0]
	assert.Equal(t, "SwagPayPal", got.Name)
	assert.Equal(t, "updated", got.State)

	// Version transition must survive the round-trip.
	require.NotNil(t, got.OldVersion, "oldVersion lost in round-trip")
	require.NotNil(t, got.NewVersion, "newVersion lost in round-trip")
	assert.Equal(t, old, *got.OldVersion)
	assert.Equal(t, updated, *got.NewVersion)

	// Changelog text and date must survive the round-trip.
	require.NotNil(t, got.Changelog)
	require.Len(t, *got.Changelog, 1)
	entry := (*got.Changelog)[0]
	assert.Equal(t, "5.1.0", entry.Version)
	assert.Equal(t, "<p>Fixed a checkout bug</p>", entry.Text)
	assert.True(t, created.Equal(entry.CreationDate))
}

func TestStoreChangelogsBetween(t *testing.T) {
	mkEn := func(v string) shopwareaccount.StoreChangelog {
		cl := shopwareaccount.StoreChangelog{Version: v, Text: "en " + v}
		cl.CreationDate.Date = "2024-01-15 08:30:00"
		return cl
	}
	mkDe := func(v string) shopwareaccount.StoreChangelog {
		return shopwareaccount.StoreChangelog{Version: v, Text: "de " + v}
	}

	sd := &storemodel.Data{
		English: &shopwareaccount.StorePlugin{
			Name: "AcmeExt", Version: "2.1.0",
			Changelogs: []shopwareaccount.StoreChangelog{
				mkEn("2.1.0"), // store returns newest-first
				mkEn("2.0.5"),
				mkEn("2.0.0"),
				mkEn("1.9.0"),
			},
		},
		German: &shopwareaccount.StorePlugin{
			Name: "AcmeExt", Version: "2.1.0",
			Changelogs: []shopwareaccount.StoreChangelog{mkDe("2.1.0"), mkDe("2.0.5")},
		},
	}

	// Window (installed 2.0.0, new 2.1.0] -> 2.0.5 and 2.1.0, oldest-first.
	got := storemodel.Between(sd.MergedChangelogs(), "2.0.0", "2.1.0")
	require.Len(t, got, 2)
	assert.Equal(t, "2.0.5", got[0].Version)
	assert.Equal(t, "en 2.0.5", got[0].Text)
	assert.Equal(t, "de 2.0.5", got[0].TextDe, "German changelog text must be merged in")
	assert.Equal(t, "2.1.0", got[1].Version)
	assert.Equal(t, "de 2.1.0", got[1].TextDe)
	assert.True(t, time.Date(2024, 1, 15, 8, 30, 0, 0, time.UTC).Equal(got[0].CreationDate))

	// Nothing newer than the new version.
	assert.Nil(t, storemodel.Between(sd.MergedChangelogs(), "2.1.0", "2.1.0"))
}

func TestGlobalLatestVersion(t *testing.T) {
	// The global latest is the newest changelog version (uncapped), not the
	// Shopware-version-capped Version field.
	sd := &storemodel.Data{
		English: &shopwareaccount.StorePlugin{
			Name:    "AcmeExt",
			Version: "9.12.4", // compatible cap for this environment
			Changelogs: []shopwareaccount.StoreChangelog{
				{Version: "10.6.4"}, // newest, requires newer Shopware
				{Version: "9.12.4"},
				{Version: "9.10.0"},
			},
		},
	}
	if got := sd.GlobalLatestVersion(); got != "10.6.4" {
		t.Fatalf("globalLatestVersion() = %q, want 10.6.4", got)
	}

	// Falls back to the compatible version when no changelog is available.
	sd = &storemodel.Data{English: &shopwareaccount.StorePlugin{Name: "AcmeExt", Version: "1.2.3"}}
	if got := sd.GlobalLatestVersion(); got != "1.2.3" {
		t.Fatalf("globalLatestVersion() fallback = %q, want 1.2.3", got)
	}
}

func TestExtensionUpdateChangelogWindow(t *testing.T) {
	// Shop updates AcmeExt 2.1.0 -> 2.2.1; the diff must record the version
	// transition, and the changelog window attached from the catalog must
	// contain the entries strictly between the two versions, in both languages.
	old := []existingExtension{
		{Name: "AcmeExt", Label: "Acme", Active: true, Version: "2.1.0", Installed: true},
	}
	newExts := []extensionEntry{
		{Name: "AcmeExt", Label: "Acme", Active: true, Version: "2.2.1", Installed: true, isStore: true},
	}

	diffs := calculateExtensionDiff(old, newExts)
	require.Len(t, diffs, 1)
	d := diffs[0]
	assert.Equal(t, "updated", d.State)
	require.NotNil(t, d.OldVersion)
	require.NotNil(t, d.NewVersion)
	assert.Equal(t, "2.1.0", *d.OldVersion)
	assert.Equal(t, "2.2.1", *d.NewVersion)

	// The catalog rows (here as merged changelogs) fill the diff's changelog.
	mcs := []storemodel.MergedChangelog{
		{Version: "2.2.2", English: "above target"},
		{Version: "2.2.1", English: "target", German: "ziel"},
		{Version: "2.2.0", English: "minor"},
		{Version: "2.1.1", English: "patch 1"},
		{Version: "2.1.0", English: "installed"},
	}
	d.Changelog = storemodel.Between(mcs, *d.OldVersion, *d.NewVersion)

	require.Len(t, d.Changelog, 3)
	got := []string{d.Changelog[0].Version, d.Changelog[1].Version, d.Changelog[2].Version}
	assert.Equal(t, []string{"2.1.1", "2.2.0", "2.2.1"}, got)
	assert.Equal(t, "ziel", d.Changelog[2].TextDe)
}

func TestParseStoreDate(t *testing.T) {
	want := time.Date(2024, 1, 15, 8, 30, 0, 0, time.UTC)

	cases := []string{
		"2024-01-15 08:30:00.000000",
		"2024-01-15 08:30:00",
		"2024-01-15T08:30:00Z",
	}
	for _, in := range cases {
		assert.True(t, want.Equal(storemodel.ParseDate(in)), "input %q", in)
	}

	assert.True(t, storemodel.ParseDate("not-a-date").IsZero())
}
