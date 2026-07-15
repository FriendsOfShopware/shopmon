package scrape

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

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
func (h *Service) loadExistingExtensions(ctx context.Context, envID int32) []existingExtension {
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

	// The scrape only reads name/version/state here, not localized text, so the
	// language is irrelevant — request English.
	store, err := h.queries.GetEnvironmentStoreExtensions(ctx, queries.GetEnvironmentStoreExtensionsParams{
		EnvironmentID: envID,
		Language:      "en",
	})
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
// Changelogs for updated store extensions are attached separately from the
// catalog (attachStoreChangelogs).
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

		diffs = append(diffs, extensionDiff{
			Name:       newExt.Name,
			Label:      newExt.Label,
			State:      state,
			OldVersion: strPtr(old.Version),
			NewVersion: strPtr(newExt.Version),
			Active:     newExt.Active,
		})
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
