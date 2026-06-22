package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
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
		LastScrapedAt:    pgtimeToTimePtr(row.LastScrapedAt),
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

func mapEnvironmentExtensions(rows []queries.EnvironmentExtension) []api.EnvironmentExtension {
	extensions := make([]api.EnvironmentExtension, 0, len(rows))
	for _, row := range rows {
		var ratingAvg *float32
		if row.RatingAverage != nil {
			v := float32(*row.RatingAverage)
			ratingAvg = &v
		}

		var changelog *[]api.ExtensionChangelogEntry
		if len(row.Changelog) > 0 {
			var entries []api.ExtensionChangelogEntry
			if err := json.Unmarshal(row.Changelog, &entries); err != nil {
				slog.Warn("failed to unmarshal extension changelog", "extension", row.Name, "error", err)
			} else if len(entries) > 0 {
				changelog = &entries
			}
		}

		var installedAt *time.Time
		if row.InstalledAt != nil {
			if parsed, err := time.Parse(time.RFC3339, *row.InstalledAt); err == nil {
				installedAt = &parsed
			}
		}

		latestVersion := ""
		if row.LatestVersion != nil {
			latestVersion = *row.LatestVersion
		}

		extensions = append(extensions, api.EnvironmentExtension{
			Name:          row.Name,
			Label:         row.Label,
			Version:       row.Version,
			LatestVersion: latestVersion,
			Active:        row.Active,
			Installed:     row.Installed,
			StoreLink:     row.StoreLink,
			RatingAverage: ratingAvg,
			Changelog:     changelog,
			InstalledAt:   installedAt,
		})
	}

	return extensions
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
		checks = append(checks, api.EnvironmentCheck{
			Id:      row.CheckID,
			Level:   row.Level,
			Message: row.Message,
			Link:    row.Link,
		})
	}

	return checks
}

func mapEnvironmentSitespeeds(rows []queries.EnvironmentSitespeed) []api.Sitespeed {
	sitespeeds := make([]api.Sitespeed, 0, len(rows))
	for _, row := range rows {
		createdAt := pgtimeToTime(row.CreatedAt)

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

func mapEnvironmentChangelogs(environment *queries.GetEnvironmentByIDRow, rows []queries.EnvironmentChangelog) []api.AccountChangelog {
	changelogs := make([]api.AccountChangelog, 0, len(rows))
	for _, row := range rows {
		environmentID := 0
		if row.EnvironmentID != nil {
			environmentID = int(*row.EnvironmentID)
		}

		var extensions []api.ExtensionDiff
		if row.Extensions != nil {
			_ = json.Unmarshal(row.Extensions, &extensions)
		}
		if extensions == nil {
			extensions = []api.ExtensionDiff{}
		}

		changelogs = append(changelogs, api.AccountChangelog{
			Id:                          int(row.ID),
			EnvironmentId:               environmentID,
			EnvironmentName:             environment.Name,
			EnvironmentOrganizationName: environment.OrganizationName,
			EnvironmentOrganizationId:   environment.OrganizationID,
			Extensions:                  extensions,
			OldShopwareVersion:          row.OldShopwareVersion,
			NewShopwareVersion:          row.NewShopwareVersion,
			Date:                        pgtimeToTime(row.Date),
		})
	}

	return changelogs
}
