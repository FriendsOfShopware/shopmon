package jobs

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// enrichExtensionsFromStore queries the Shopware Store API and updates extensions
// with latestVersion, ratingAverage, storeLink, and changelog information.
func (h *EnvironmentScrapeHandler) enrichExtensionsFromStore(extensions []extensionEntry, shopwareVersion string) {
	storeURL, err := url.Parse("https://api.shopware.com/pluginStore/pluginsByName")
	if err != nil {
		return
	}

	q := storeURL.Query()
	q.Set("locale", "en-GB")
	q.Set("shopwareVersion", shopwareVersion)
	for _, ext := range extensions {
		q.Add("technicalNames[]", ext.Name)
	}
	storeURL.RawQuery = q.Encode()

	httpClient := httputil.NewHTTPClient(httputil.WithTimeout(30 * time.Second))
	resp, err := httpClient.Get(storeURL.String())
	if err != nil {
		slog.Warn("failed to fetch store plugins", "error", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var storePlugins []storePlugin
	if err := json.Unmarshal(body, &storePlugins); err != nil {
		slog.Warn("failed to parse store plugins response", "error", err)
		return
	}

	// Build a map for quick lookup
	storeMap := make(map[string]*storePlugin, len(storePlugins))
	for i := range storePlugins {
		storeMap[storePlugins[i].Name] = &storePlugins[i]
	}

	for i := range extensions {
		sp, ok := storeMap[extensions[i].Name]
		if !ok {
			continue
		}

		extensions[i].LatestVersion = strPtr(sp.Version)
		extensions[i].RatingAverage = &sp.RatingAverage

		storeLink := strings.ReplaceAll(sp.StoreLink, "http://store.shopware.com:80", "https://store.shopware.com")
		extensions[i].StoreLink = &storeLink

		if changelogs := buildCompatibleChangelogs(sp, extensions[i].Version); len(changelogs) > 0 {
			extensions[i].Changelog = changelogs
		}
	}
}

// buildCompatibleChangelogs returns store changelog entries strictly newer than
// the installed version and not exceeding sp.Version. sp.Version is the latest
// extension version the store reports as compatible with the shop's Shopware
// version (controlled by the shopwareVersion query param to the store API),
// so capping at it excludes entries that require a newer Shopware.
func buildCompatibleChangelogs(sp *storePlugin, installedVersion string) []extensionChangelog {
	var changelogs []extensionChangelog
	for _, cl := range sp.Changelogs {
		if versionCompare(cl.Version, installedVersion) > 0 &&
			versionCompare(cl.Version, sp.Version) <= 0 {
			changelogs = append(changelogs, extensionChangelog{
				Version:      cl.Version,
				Text:         cl.Text,
				CreationDate: parseStoreDate(cl.CreationDate.Date),
			})
		}
	}
	return changelogs
}

// calculateExtensionDiff compares old (from DB) and new (from scrape) extensions.
func calculateExtensionDiff(oldExtensions []queries.EnvironmentExtension, newExtensions []extensionEntry) []extensionDiff {
	if len(oldExtensions) == 0 {
		return nil
	}

	var diffs []extensionDiff

	// Check for updated/deactivated/activated/removed
	for _, old := range oldExtensions {
		found := false
		for _, newExt := range newExtensions {
			if old.Name == newExt.Name {
				found = true

				var state string
				if old.Version != newExt.Version {
					state = "updated"
				} else if old.Active && !newExt.Active {
					state = "deactivated"
				} else if !old.Active && newExt.Active {
					state = "activated"
				}

				if state != "" {
					diff := extensionDiff{
						Name:       newExt.Name,
						Label:      newExt.Label,
						State:      state,
						OldVersion: &old.Version,
						NewVersion: &newExt.Version,
						Active:     newExt.Active,
					}
					if state == "updated" && old.Changelog != nil {
						var cl []extensionChangelog
						if json.Unmarshal(old.Changelog, &cl) == nil {
							// Stored changelog was capped at the latest compatible version
							// at scrape time, which may be ahead of the version the shop
							// actually updated to. Keep only entries up to newExt.Version.
							filtered := cl[:0]
							for _, entry := range cl {
								if versionCompare(entry.Version, newExt.Version) <= 0 {
									filtered = append(filtered, entry)
								}
							}
							diff.Changelog = filtered
						}
					}
					diffs = append(diffs, diff)
				}
				break
			}
		}

		if !found {
			diffs = append(diffs, extensionDiff{
				Name:       old.Name,
				Label:      old.Label,
				State:      "removed",
				OldVersion: &old.Version,
				NewVersion: nil,
				Active:     old.Active,
			})
		}
	}

	// Check for newly installed
	for _, newExt := range newExtensions {
		found := false
		for _, old := range oldExtensions {
			if old.Name == newExt.Name {
				found = true
				break
			}
		}
		if !found {
			diffs = append(diffs, extensionDiff{
				Name:       newExt.Name,
				Label:      newExt.Label,
				State:      "installed",
				OldVersion: nil,
				NewVersion: &newExt.Version,
				Active:     newExt.Active,
			})
		}
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

// versionCompare compares two version strings.
// Returns 1 if a > b, -1 if a < b, 0 if equal.
func versionCompare(a, b string) int {
	ap := strings.Split(a, ".")
	bp := strings.Split(b, ".")

	maxLen := len(ap)
	if len(bp) > maxLen {
		maxLen = len(bp)
	}

	for i := 0; i < maxLen; i++ {
		var an, bn int
		if i < len(ap) {
			an, _ = strconv.Atoi(ap[i])
		}
		if i < len(bp) {
			bn, _ = strconv.Atoi(bp[i])
		}
		if an > bn {
			return 1
		}
		if bn > an {
			return -1
		}
	}
	return 0
}

// getFavicon fetches the shop URL and parses the HTML for a favicon link.
func getFavicon(shopURL string) *string {
	httpClient := httputil.NewHTTPClient(httputil.WithTimeout(10*time.Second), func(c *http.Client) {
		c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		}
	})

	resp, err := httpClient.Get(shopURL)
	if err != nil {
		return nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // Limit to 1MB
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
