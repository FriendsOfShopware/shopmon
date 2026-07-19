package environment

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/ptr"
	"github.com/friendsofshopware/shopmon/api/internal/version"
)

// toAccountEnvironment maps a single organization environment row to its API DTO.
func toAccountEnvironment(row queries.ListEnvironmentsByOrganizationRow) api.AccountEnvironment {
	var shopID *int
	if row.ShopID > 0 {
		v := int(row.ShopID)
		shopID = &v
	}
	return api.AccountEnvironment{
		Id:               int(row.ID),
		Name:             row.Name,
		Url:              row.Url,
		Favicon:          row.Favicon,
		Status:           row.Status,
		ShopwareVersion:  row.ShopwareVersion,
		LastScrapedAt:    database.TimePtr(row.LastScrapedAt),
		LastScrapedError: row.LastScrapedError,
		OrganizationId:   row.OrganizationID,
		OrganizationName: row.OrganizationName,
		ShopId:           shopID,
		ShopName:         row.ShopName,
	}
}

// toAccountEnvironments maps a slice of organization environment rows to API DTOs.
func toAccountEnvironments(rows []queries.ListEnvironmentsByOrganizationRow) []api.AccountEnvironment {
	result := make([]api.AccountEnvironment, 0, len(rows))
	for _, row := range rows {
		result = append(result, toAccountEnvironment(row))
	}
	return result
}

// mergeEnvironmentExtensions assembles the unified extension list for an
// environment from the normalized store tables and the unknown-extension table.
// Store extensions carry catalog metadata (store link, rating) and a
// compatibility-capped changelog (already resolved to the requested language)
// built from the version catalog; unknown extensions carry only the
// shop-reported fields.
func mergeEnvironmentExtensions(
	storeRows []queries.GetEnvironmentStoreExtensionsRow,
	unknownRows []queries.EnvironmentExtension,
	changelogRows []queries.GetEnvironmentStoreExtensionChangelogsRow,
) []api.EnvironmentExtension {
	changelogByExt := make(map[string][]changelogVersion)
	for _, r := range changelogRows {
		changelogByExt[r.ExtensionName] = append(changelogByExt[r.ExtensionName], changelogVersion{
			Version:    r.Version,
			Text:       ptr.Deref(r.Changelog),
			ReleasedAt: ptr.Deref(r.ReleasedAt),
		})
	}

	extensions := make([]api.EnvironmentExtension, 0, len(storeRows)+len(unknownRows))

	for _, row := range storeRows {
		var ratingAvg *float32
		if row.RatingAverage != nil {
			v := float32(*row.RatingAverage)
			ratingAvg = &v
		}

		latestVersion := ptr.Deref(row.LatestVersion)
		changelog := buildCompatibleChangelog(changelogByExt[row.ExtensionName], row.Version, latestVersion)

		extensions = append(extensions, api.EnvironmentExtension{
			Name:          row.ExtensionName,
			Label:         row.Label,
			Version:       row.Version,
			LatestVersion: latestVersion,
			Active:        row.Active,
			Installed:     row.Installed,
			StoreLink:     row.StoreLink,
			RatingAverage: ratingAvg,
			Changelog:     changelog,
			InstalledAt:   ptr.ParseTime(row.InstalledAt, time.RFC3339),
		})
	}

	for _, row := range unknownRows {
		extensions = append(extensions, api.EnvironmentExtension{
			Name:          row.Name,
			Label:         row.Label,
			Version:       row.Version,
			LatestVersion: ptr.Deref(row.LatestVersion),
			Active:        row.Active,
			Installed:     row.Installed,
			InstalledAt:   ptr.ParseTime(row.InstalledAt, time.RFC3339),
		})
	}

	sort.Slice(extensions, func(i, j int) bool {
		return extensions[i].Name < extensions[j].Name
	})
	return extensions
}

// changelogVersion is an intermediate store changelog entry whose text is
// already resolved to the requested language.
type changelogVersion struct {
	Version    string
	Text       string
	ReleasedAt string
}

// buildCompatibleChangelog returns changelog entries strictly newer than the
// installed version and not newer than the latest compatible version (when
// known), ordered oldest-first. Returns nil when there is nothing to show.
func buildCompatibleChangelog(versions []changelogVersion, installedVersion, latestVersion string) *[]api.ExtensionChangelogEntry {
	entries := make([]api.ExtensionChangelogEntry, 0, len(versions))
	for _, v := range versions {
		if version.Compare(v.Version, installedVersion) <= 0 {
			continue
		}
		if latestVersion != "" && version.Compare(v.Version, latestVersion) > 0 {
			continue
		}
		entry := api.ExtensionChangelogEntry{
			Version: v.Version,
			Text:    v.Text,
		}
		if v.ReleasedAt != "" {
			if t, err := time.Parse(time.RFC3339, v.ReleasedAt); err == nil {
				entry.CreationDate = t
			}
		}
		entries = append(entries, entry)
	}
	if len(entries) == 0 {
		return nil
	}
	sort.Slice(entries, func(i, j int) bool {
		return version.Compare(entries[i].Version, entries[j].Version) < 0
	})
	return &entries
}

func mapEnvironmentScheduledTasks(rows []queries.EnvironmentScheduledTask) []api.ScheduledTask {
	tasks := make([]api.ScheduledTask, 0, len(rows))
	for _, row := range rows {
		var lastExecutionTime *time.Time
		if row.LastExecutionTime != nil {
			if parsed, err := time.Parse(time.RFC3339, *row.LastExecutionTime); err == nil {
				lastExecutionTime = &parsed
			}
		}

		var nextExecutionTime *time.Time
		if row.NextExecutionTime != nil {
			if parsed, err := time.Parse(time.RFC3339, *row.NextExecutionTime); err == nil {
				nextExecutionTime = &parsed
			}
		}

		tasks = append(tasks, api.ScheduledTask{
			Id:                row.TaskID,
			Name:              row.Name,
			Status:            row.Status,
			RunInterval:       int(row.Interval),
			Overdue:           row.Overdue,
			LastExecutionTime: lastExecutionTime,
			NextExecutionTime: nextExecutionTime,
		})
	}

	return tasks
}

func mapEnvironmentQueues(rows []queries.EnvironmentQueue) []api.Queue {
	queues := make([]api.Queue, 0, len(rows))
	for _, row := range rows {
		queues = append(queues, api.Queue{
			Name: row.Name,
			Size: int(row.Size),
		})
	}

	return queues
}

func mapEnvironmentCache(row *queries.EnvironmentCache) *api.CacheInfo {
	if row == nil {
		return nil
	}

	id := int(row.ID)
	return &api.CacheInfo{
		Id:           &id,
		Environment:  &row.Environment,
		HttpCache:    &row.HttpCache,
		CacheAdapter: &row.CacheAdapter,
	}
}

func mapEnvironmentChecks(rows []queries.EnvironmentCheck) []api.EnvironmentCheck {
	checks := make([]api.EnvironmentCheck, 0, len(rows))
	for _, row := range rows {
		check := api.EnvironmentCheck{
			Id:         row.CheckID,
			Level:      row.Level,
			Message:    row.Message,
			MessageKey: row.MessageKey,
			Link:       row.Link,
		}

		if len(row.Params) > 0 {
			var params map[string]interface{}
			if err := json.Unmarshal(row.Params, &params); err == nil && len(params) > 0 {
				check.Params = &params
			}
		}

		checks = append(checks, check)
	}

	return checks
}

func mapEnvironmentSitespeeds(rows []queries.EnvironmentSitespeed) []api.Sitespeed {
	sitespeeds := make([]api.Sitespeed, 0, len(rows))
	for _, row := range rows {
		createdAt := database.Time(row.CreatedAt)

		var ttfb, fullyLoaded, largestContentfulPaint, firstContentfulPaint, transferSize *float32
		var cumulativeLayoutShift *float32
		if row.Ttfb != nil {
			v := float32(*row.Ttfb)
			ttfb = &v
		}
		if row.FullyLoaded != nil {
			v := float32(*row.FullyLoaded)
			fullyLoaded = &v
		}
		if row.LargestContentfulPaint != nil {
			v := float32(*row.LargestContentfulPaint)
			largestContentfulPaint = &v
		}
		if row.FirstContentfulPaint != nil {
			v := float32(*row.FirstContentfulPaint)
			firstContentfulPaint = &v
		}
		if row.CumulativeLayoutShift != nil {
			cumulativeLayoutShift = row.CumulativeLayoutShift
		}
		if row.TransferSize != nil {
			v := float32(*row.TransferSize)
			transferSize = &v
		}

		var deployment *api.SitespeedDeployment
		if row.DeploymentID != nil {
			deployment = &api.SitespeedDeployment{
				Id:   int(*row.DeploymentID),
				Name: fmt.Sprintf("Deployment #%d", *row.DeploymentID),
			}
		}

		sitespeeds = append(sitespeeds, api.Sitespeed{
			CreatedAt:              createdAt,
			Ttfb:                   ttfb,
			FullyLoaded:            fullyLoaded,
			LargestContentfulPaint: largestContentfulPaint,
			FirstContentfulPaint:   firstContentfulPaint,
			CumulativeLayoutShift:  cumulativeLayoutShift,
			TransferSize:           transferSize,
			Deployment:             deployment,
		})
	}

	return sitespeeds
}

func mapEnvironmentChangelogs(environment *queries.GetEnvironmentByIDRow, rows []queries.EnvironmentChangelog) ([]api.AccountChangelog, error) {
	changelogs := make([]api.AccountChangelog, 0, len(rows))
	for _, row := range rows {
		environmentID := 0
		if row.EnvironmentID != nil {
			environmentID = int(*row.EnvironmentID)
		}

		var extensions []api.ExtensionDiff
		if row.Extensions != nil {
			if err := json.Unmarshal(row.Extensions, &extensions); err != nil {
				return nil, fmt.Errorf("decode environment changelog %d extensions: %w", row.ID, err)
			}
		}
		if extensions == nil {
			extensions = []api.ExtensionDiff{}
		}

		changelogs = append(changelogs, api.AccountChangelog{
			Id:                          int(row.ID),
			EnvironmentId:               environmentID,
			EnvironmentName:             environment.Name,
			EnvironmentShopName:         ptr.Deref(environment.ShopName),
			EnvironmentOrganizationName: environment.OrganizationName,
			EnvironmentOrganizationId:   environment.OrganizationID,
			Extensions:                  extensions,
			OldShopwareVersion:          row.OldShopwareVersion,
			NewShopwareVersion:          row.NewShopwareVersion,
			Date:                        database.Time(row.Date),
		})
	}

	return changelogs, nil
}
