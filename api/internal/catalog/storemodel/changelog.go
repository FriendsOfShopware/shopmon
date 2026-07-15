package storemodel

import (
	"sort"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/shopwareaccount"
	"github.com/friendsofshopware/shopmon/api/internal/version"
)

type Data struct {
	English *shopwareaccount.StorePlugin
	German  *shopwareaccount.StorePlugin
}

func (d *Data) Primary() *shopwareaccount.StorePlugin {
	if d == nil {
		return nil
	}
	if d.English != nil {
		return d.English
	}
	return d.German
}

type MergedChangelog struct {
	Version    string
	English    string
	German     string
	ReleasedAt string
}

type Changelog struct {
	Version      string    `json:"version"`
	Text         string    `json:"text"`
	TextDe       string    `json:"textDe,omitempty"`
	CreationDate time.Time `json:"creationDate"`
}

func IndexPlugins(plugins []shopwareaccount.StorePlugin) map[string]*shopwareaccount.StorePlugin {
	result := make(map[string]*shopwareaccount.StorePlugin, len(plugins))
	for i := range plugins {
		result[plugins[i].Name] = &plugins[i]
	}
	return result
}

func (d *Data) MergedChangelogs() []MergedChangelog {
	byVersion := make(map[string]*MergedChangelog)
	order := make([]string, 0)
	add := func(changelogs []shopwareaccount.StoreChangelog, german bool) {
		for _, changelog := range changelogs {
			merged, ok := byVersion[changelog.Version]
			if !ok {
				merged = &MergedChangelog{Version: changelog.Version}
				byVersion[changelog.Version] = merged
				order = append(order, changelog.Version)
			}
			if german {
				merged.German = changelog.Text
			} else {
				merged.English = changelog.Text
			}
			if merged.ReleasedAt == "" && changelog.CreationDate.Date != "" {
				if parsed := ParseDate(changelog.CreationDate.Date); !parsed.IsZero() {
					merged.ReleasedAt = parsed.Format(time.RFC3339)
				}
			}
		}
	}
	if d.English != nil {
		add(d.English.Changelogs, false)
	}
	if d.German != nil {
		add(d.German.Changelogs, true)
	}
	result := make([]MergedChangelog, 0, len(order))
	for _, value := range order {
		result = append(result, *byVersion[value])
	}
	return result
}

func (d *Data) GlobalLatestVersion() string {
	latest := ""
	for _, changelog := range d.MergedChangelogs() {
		if latest == "" || version.Compare(changelog.Version, latest) > 0 {
			latest = changelog.Version
		}
	}
	if latest == "" {
		if primary := d.Primary(); primary != nil {
			latest = primary.Version
		}
	}
	return latest
}

func Between(changelogs []MergedChangelog, oldVersion, newVersion string) []Changelog {
	var result []Changelog
	for _, changelog := range changelogs {
		if version.Compare(changelog.Version, oldVersion) > 0 && version.Compare(changelog.Version, newVersion) <= 0 {
			entry := Changelog{
				Version: changelog.Version,
				Text:    changelog.English,
				TextDe:  changelog.German,
			}
			if changelog.ReleasedAt != "" {
				if parsed, err := time.Parse(time.RFC3339, changelog.ReleasedAt); err == nil {
					entry.CreationDate = parsed
				}
			}
			result = append(result, entry)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return version.Compare(result[i].Version, result[j].Version) < 0
	})
	return result
}

func ParseDate(value string) time.Time {
	formats := []string{
		"2006-01-02 15:04:05.000000",
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02",
	}
	for _, format := range formats {
		if parsed, err := time.Parse(format, value); err == nil {
			return parsed
		}
	}
	return time.Time{}
}
