package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCompatibleChangelog(t *testing.T) {
	versions := []changelogVersion{
		{Version: "1.9.0", TextEn: "en 1.9.0", TextDe: "de 1.9.0"},
		{Version: "2.0.0", TextEn: "en 2.0.0"},
		{Version: "2.0.5", TextEn: "en 2.0.5", TextDe: "de 2.0.5", ReleasedAt: "2024-01-15T08:30:00Z"},
		{Version: "2.1.0", TextEn: "en 2.1.0", TextDe: "de 2.1.0"},
		{Version: "2.2.0", TextEn: "en 2.2.0"}, // above latest compatible -> excluded
	}

	// Installed 2.0.0, latest compatible 2.1.0 -> 2.0.5 and 2.1.0, oldest-first.
	got := buildCompatibleChangelog(versions, "2.0.0", "2.1.0")
	require.NotNil(t, got)
	require.Len(t, *got, 2)

	first := (*got)[0]
	assert.Equal(t, "2.0.5", first.Version)
	assert.Equal(t, "en 2.0.5", first.Text)
	require.NotNil(t, first.TextDe)
	assert.Equal(t, "de 2.0.5", *first.TextDe)
	assert.False(t, first.CreationDate.IsZero())

	second := (*got)[1]
	assert.Equal(t, "2.1.0", second.Version)
	require.NotNil(t, second.TextDe)
	assert.Equal(t, "de 2.1.0", *second.TextDe)

	// No German text -> TextDe stays nil.
	got = buildCompatibleChangelog(versions, "1.9.0", "2.0.0")
	require.NotNil(t, got)
	require.Len(t, *got, 1)
	assert.Equal(t, "2.0.0", (*got)[0].Version)
	assert.Nil(t, (*got)[0].TextDe)

	// Nothing newer than installed -> nil.
	assert.Nil(t, buildCompatibleChangelog(versions, "2.2.0", ""))
}
