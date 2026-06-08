package jobs

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/api"
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
					IsCompatible: true,
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
	assert.True(t, entry.IsCompatible)
	assert.True(t, created.Equal(entry.CreationDate))
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
