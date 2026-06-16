package jobs

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
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

func TestBuildCompatibleChangelogs(t *testing.T) {
	mkEntry := func(v string) shopwareaccount.StoreChangelog {
		cl := shopwareaccount.StoreChangelog{Version: v, Text: "notes " + v}
		cl.CreationDate.Date = "2024-01-15 08:30:00"
		return cl
	}

	sp := &shopwareaccount.StorePlugin{
		Name:    "AcmeExt",
		Version: "2.1.0", // latest version compatible with shop's Shopware version
		Changelogs: []shopwareaccount.StoreChangelog{
			mkEntry("1.9.0"), // older than installed -> excluded
			mkEntry("2.0.0"), // == installed       -> excluded
			mkEntry("2.0.5"), // in range            -> included
			mkEntry("2.1.0"), // == cap              -> included
			mkEntry("2.2.0"), // requires newer SW   -> excluded
			mkEntry("3.0.0"), // requires newer SW   -> excluded
		},
	}

	got := buildCompatibleChangelogs(sp, "2.0.0")

	require.Len(t, got, 2)
	assert.Equal(t, "2.0.5", got[0].Version)
	assert.Equal(t, "2.1.0", got[1].Version)

	// When installed == sp.Version there's nothing newer compatible.
	assert.Nil(t, buildCompatibleChangelogs(sp, "2.1.0"))
}

func TestCalculateExtensionDiffCapsChangelogToNewVersion(t *testing.T) {
	// Stored changelog at scrape time was capped at the latest compatible
	// version (2.2.4). Shop actually updates 2.1.0 -> 2.2.1, so the diff
	// must drop entries above 2.2.1.
	stored := []extensionChangelog{
		{Version: "2.1.1", Text: "patch 1"},
		{Version: "2.2.0", Text: "minor"},
		{Version: "2.2.1", Text: "target"},
		{Version: "2.2.2", Text: "above target"},
		{Version: "2.2.4", Text: "latest compatible"},
	}
	raw, err := json.Marshal(stored)
	require.NoError(t, err)

	old := []queries.EnvironmentExtension{
		{
			Name:      "AcmeExt",
			Label:     "Acme",
			Active:    true,
			Version:   "2.1.0",
			Installed: true,
			Changelog: raw,
		},
	}
	newExts := []extensionEntry{
		{
			Name:      "AcmeExt",
			Label:     "Acme",
			Active:    true,
			Version:   "2.2.1",
			Installed: true,
		},
	}

	diffs := calculateExtensionDiff(old, newExts)
	require.Len(t, diffs, 1)
	d := diffs[0]
	assert.Equal(t, "updated", d.State)
	require.NotNil(t, d.OldVersion)
	require.NotNil(t, d.NewVersion)
	assert.Equal(t, "2.1.0", *d.OldVersion)
	assert.Equal(t, "2.2.1", *d.NewVersion)

	require.Len(t, d.Changelog, 3)
	got := []string{d.Changelog[0].Version, d.Changelog[1].Version, d.Changelog[2].Version}
	assert.Equal(t, []string{"2.1.1", "2.2.0", "2.2.1"}, got)
}

func TestParseStoreDate(t *testing.T) {
	want := time.Date(2024, 1, 15, 8, 30, 0, 0, time.UTC)

	cases := []string{
		"2024-01-15 08:30:00.000000",
		"2024-01-15 08:30:00",
		"2024-01-15T08:30:00Z",
	}
	for _, in := range cases {
		assert.True(t, want.Equal(parseStoreDate(in)), "input %q", in)
	}

	assert.True(t, parseStoreDate("not-a-date").IsZero())
}
