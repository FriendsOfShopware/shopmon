package jobs

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/shopwareaccount"
	"github.com/friendsofshopware/shopmon/api/internal/version"
)

// enrichExtensionsFromStore fetches store metadata for the given extensions in
// both English (en_GB) and German (de_DE) and attaches it to the entries the
// store knows about, recording the latest compatible version per entry. The
// locale must use the underscore form; the hyphenated form is silently ignored
// by the API. It returns false only when both locale calls fail, meaning store
// membership is unknown for this scrape.
func (h *EnvironmentScrapeHandler) enrichExtensionsFromStore(extensions []extensionEntry, shopwareVersion string) bool {
	technicalNames := make([]string, 0, len(extensions))
	for _, ext := range extensions {
		technicalNames = append(technicalNames, ext.Name)
	}

	client := shopwareaccount.NewClient("", nil)

	var (
		enPlugins, dePlugins []shopwareaccount.StorePlugin
		enErr, deErr         error
		wg                   sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		enPlugins, enErr = client.PluginsByName(context.Background(), "en_GB", shopwareVersion, technicalNames)
	}()
	go func() {
		defer wg.Done()
		dePlugins, deErr = client.PluginsByName(context.Background(), "de_DE", shopwareVersion, technicalNames)
	}()
	wg.Wait()

	if enErr != nil {
		slog.Warn("failed to fetch store plugins", "locale", "en_GB", "error", enErr)
	}
	if deErr != nil {
		slog.Warn("failed to fetch store plugins", "locale", "de_DE", "error", deErr)
	}

	enMap := indexStorePlugins(enPlugins)
	deMap := indexStorePlugins(dePlugins)

	for i := range extensions {
		en := enMap[extensions[i].Name]
		de := deMap[extensions[i].Name]
		if en == nil && de == nil {
			continue
		}
		data := &storeExtensionData{en: en, de: de}
		extensions[i].Store = data
		// latest_version is the latest release the store reports as compatible
		// with this environment's Shopware version (the shopwareVersion query
		// param caps it), so it is recorded per environment on the link row.
		if p := data.primary(); p != nil && p.Version != "" {
			v := p.Version
			extensions[i].LatestVersion = &v
		}
	}

	return enErr == nil || deErr == nil
}

func indexStorePlugins(plugins []shopwareaccount.StorePlugin) map[string]*shopwareaccount.StorePlugin {
	m := make(map[string]*shopwareaccount.StorePlugin, len(plugins))
	for i := range plugins {
		m[plugins[i].Name] = &plugins[i]
	}
	return m
}

// mergedChangelog is a single store changelog version with both language texts.
type mergedChangelog struct {
	Version    string
	En         string
	De         string
	ReleasedAt string
}

// mergedChangelogs combines the English and German changelog entries by version.
func (s *storeExtensionData) mergedChangelogs() []mergedChangelog {
	byVersion := make(map[string]*mergedChangelog)
	order := make([]string, 0)
	add := func(cls []shopwareaccount.StoreChangelog, isDe bool) {
		for _, cl := range cls {
			mc, ok := byVersion[cl.Version]
			if !ok {
				mc = &mergedChangelog{Version: cl.Version}
				byVersion[cl.Version] = mc
				order = append(order, cl.Version)
			}
			if isDe {
				mc.De = cl.Text
			} else {
				mc.En = cl.Text
			}
			if mc.ReleasedAt == "" && cl.CreationDate.Date != "" {
				if t := parseStoreDate(cl.CreationDate.Date); !t.IsZero() {
					mc.ReleasedAt = t.Format(time.RFC3339)
				}
			}
		}
	}
	if s.en != nil {
		add(s.en.Changelogs, false)
	}
	if s.de != nil {
		add(s.de.Changelogs, true)
	}
	result := make([]mergedChangelog, 0, len(order))
	for _, v := range order {
		result = append(result, *byVersion[v])
	}
	return result
}

// globalLatestVersion returns the newest version known from the store changelog,
// which (unlike the compatible Version field) is not capped to a specific
// Shopware version and is therefore environment-independent. It falls back to the
// compatible version when no changelog is available.
func (s *storeExtensionData) globalLatestVersion() string {
	latest := ""
	for _, mc := range s.mergedChangelogs() {
		if latest == "" || version.Compare(mc.Version, latest) > 0 {
			latest = mc.Version
		}
	}
	if latest == "" {
		if p := s.primary(); p != nil {
			latest = p.Version
		}
	}
	return latest
}

// changelogsBetween returns store changelog entries (both languages) strictly
// newer than oldVersion and not newer than newVersion, ordered oldest-first.
func (s *storeExtensionData) changelogsBetween(oldVersion, newVersion string) []extensionChangelog {
	var out []extensionChangelog
	for _, mc := range s.mergedChangelogs() {
		if version.Compare(mc.Version, oldVersion) > 0 && version.Compare(mc.Version, newVersion) <= 0 {
			entry := extensionChangelog{Version: mc.Version, Text: mc.En, TextDe: mc.De}
			if mc.ReleasedAt != "" {
				if t, err := time.Parse(time.RFC3339, mc.ReleasedAt); err == nil {
					entry.CreationDate = t
				}
			}
			out = append(out, entry)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return version.Compare(out[i].Version, out[j].Version) < 0
	})
	return out
}

// existingExtension is the unified prior state of an extension (store-known or
// unknown) used to diff against the freshly scraped extensions.
type existingExtension struct {
	Name      string
	Label     string
	Version   string
	Active    bool
	Installed bool
	// IsStore reports whether the prior row came from environment_store_extension.
	IsStore bool
	// LatestVersion is the prior link's compatible latest version (store only),
	// preserved when the store API is unavailable for a scrape.
	LatestVersion *string
}

// loadExistingExtensions returns the prior persisted state of all extensions of
// an environment, merging the unknown (environment_extension) and store-known
// (environment_store_extension) tables.
func (h *EnvironmentScrapeHandler) loadExistingExtensions(ctx context.Context, envID int32) []existingExtension {
	var result []existingExtension

	unknown, err := h.queries.GetEnvironmentExtensions(ctx, envID)
	if err != nil {
		slog.Warn("failed to get old unknown extensions", "environmentId", envID, "error", err)
	}
	for _, e := range unknown {
		result = append(result, existingExtension{
			Name: e.Name, Label: e.Label, Version: e.Version, Active: e.Active, Installed: e.Installed,
		})
	}

	store, err := h.queries.GetEnvironmentStoreExtensions(ctx, envID)
	if err != nil {
		slog.Warn("failed to get old store extensions", "environmentId", envID, "error", err)
	}
	for _, e := range store {
		result = append(result, existingExtension{
			Name: e.ExtensionName, Label: e.Label, Version: e.Version, Active: e.Active, Installed: e.Installed,
			IsStore: true, LatestVersion: e.LatestVersion,
		})
	}

	return result
}

// calculateExtensionDiff compares old (from DB) and new (from scrape) extensions.
// For updated store extensions the changelog between the old and new version is
// taken from the freshly fetched store data (both languages).
func calculateExtensionDiff(oldExtensions []existingExtension, newExtensions []extensionEntry) []extensionDiff {
	if len(oldExtensions) == 0 {
		return nil
	}

	newByName := make(map[string]*extensionEntry, len(newExtensions))
	for i := range newExtensions {
		newByName[newExtensions[i].Name] = &newExtensions[i]
	}

	var diffs []extensionDiff
	seen := make(map[string]struct{}, len(oldExtensions))

	for _, old := range oldExtensions {
		seen[old.Name] = struct{}{}

		newExt, found := newByName[old.Name]
		if !found {
			diffs = append(diffs, extensionDiff{
				Name:       old.Name,
				Label:      old.Label,
				State:      "removed",
				OldVersion: strPtr(old.Version),
				Active:     old.Active,
			})
			continue
		}

		var state string
		if old.Version != newExt.Version {
			state = "updated"
		} else if old.Active && !newExt.Active {
			state = "deactivated"
		} else if !old.Active && newExt.Active {
			state = "activated"
		}
		if state == "" {
			continue
		}

		diff := extensionDiff{
			Name:       newExt.Name,
			Label:      newExt.Label,
			State:      state,
			OldVersion: strPtr(old.Version),
			NewVersion: strPtr(newExt.Version),
			Active:     newExt.Active,
		}
		if state == "updated" && newExt.Store != nil {
			diff.Changelog = newExt.Store.changelogsBetween(old.Version, newExt.Version)
		}
		diffs = append(diffs, diff)
	}

	for i := range newExtensions {
		if _, found := seen[newExtensions[i].Name]; found {
			continue
		}
		ne := &newExtensions[i]
		diffs = append(diffs, extensionDiff{
			Name:       ne.Name,
			Label:      ne.Label,
			State:      "installed",
			NewVersion: strPtr(ne.Version),
			Active:     ne.Active,
		})
	}

	return diffs
}

// isScheduledTaskOverdue determines if a scheduled task is overdue based on its next execution time.
func isScheduledTaskOverdue(task shopwareScheduledTask) bool {
	if task.NextExecutionTime == nil {
		return false
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000+00:00",
		"2006-01-02T15:04:05+00:00",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		t, err := time.Parse(format, *task.NextExecutionTime)
		if err == nil {
			return time.Now().After(t)
		}
	}
	return false
}

// getFavicon fetches the shop URL and parses the HTML for a favicon link.
func getFavicon(ctx context.Context, shopURL string) *string {
	httpClient := httputil.NewHTTPClient(httputil.WithTimeout(10*time.Second), func(c *http.Client) {
		c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		}
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, shopURL, nil)
	if err != nil {
		return nil
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return nil
	}

	// Match <link rel="icon" ...> or <link rel="shortcut icon" ...>
	re := regexp.MustCompile(`(?i)<link[^>]+rel=["']?(?:shortcut\s+)?icon["']?[^>]*>`)
	match := re.Find(body)
	if match == nil {
		return nil
	}

	// Extract href from the matched tag
	hrefRe := regexp.MustCompile(`(?i)href=["']([^"']+)["']`)
	hrefMatch := hrefRe.FindSubmatch(match)
	if len(hrefMatch) < 2 {
		return nil
	}

	iconURL := string(hrefMatch[1])

	if strings.HasPrefix(iconURL, "http") {
		return &iconURL
	}

	if strings.HasPrefix(iconURL, "/") {
		// Convert relative to absolute
		parsed, err := url.Parse(shopURL)
		if err != nil {
			return nil
		}
		absoluteURL := fmt.Sprintf("%s://%s%s", parsed.Scheme, parsed.Host, iconURL)
		return &absoluteURL
	}

	return nil
}

func strPtr(s string) *string {
	return &s
}

// parseStoreDate parses the date string returned by the Shopware store API
// (e.g. "2024-01-15 00:00:00.000000") into a time.Time. It returns the zero
// time if the value cannot be parsed.
func parseStoreDate(s string) time.Time {
	formats := []string{
		"2006-01-02 15:04:05.000000",
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
