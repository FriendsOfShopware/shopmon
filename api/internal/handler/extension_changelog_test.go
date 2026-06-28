package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCompatibleChangelog(t *testing.T) {
	// Text is already resolved to the requested language before this point.
	versions := []changelogVersion{
		{Version: "1.9.0", Text: "1.9.0"},
		{Version: "2.0.0", Text: "2.0.0"},
		{Version: "2.0.5", Text: "2.0.5", ReleasedAt: "2024-01-15T08:30:00Z"},
		{Version: "2.1.0", Text: "2.1.0"},
		{Version: "2.2.0", Text: "2.2.0"}, // above latest compatible -> excluded
	}

	// Installed 2.0.0, latest compatible 2.1.0 -> 2.0.5 and 2.1.0, oldest-first.
	got := buildCompatibleChangelog(versions, "2.0.0", "2.1.0")
	require.NotNil(t, got)
	require.Len(t, *got, 2)

	first := (*got)[0]
	assert.Equal(t, "2.0.5", first.Version)
	assert.Equal(t, "2.0.5", first.Text)
	assert.False(t, first.CreationDate.IsZero())

	second := (*got)[1]
	assert.Equal(t, "2.1.0", second.Version)
	assert.Equal(t, "2.1.0", second.Text)

	// Only the version strictly newer than installed and not above latest.
	got = buildCompatibleChangelog(versions, "1.9.0", "2.0.0")
	require.NotNil(t, got)
	require.Len(t, *got, 1)
	assert.Equal(t, "2.0.0", (*got)[0].Version)

	// Nothing newer than installed -> nil.
	assert.Nil(t, buildCompatibleChangelog(versions, "2.2.0", ""))
}
