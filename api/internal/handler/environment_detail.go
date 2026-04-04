package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
)

type environmentDetailAggregate struct {
	extensions  []queries.EnvironmentExtension
	tasks       []queries.EnvironmentScheduledTask
	queues      []queries.EnvironmentQueue
	cache       *queries.EnvironmentCache
	checks      []queries.EnvironmentCheck
	sitespeeds  []queries.EnvironmentSitespeed
	changelogs  []queries.EnvironmentChangelog
	deployCount int32
}

func (h *Handler) loadEnvironmentDetailAggregate(ctx context.Context, environmentID int32) environmentDetailAggregate {
	aggregate := environmentDetailAggregate{}

	aggregate.extensions = h.loadEnvironmentExtensions(ctx, environmentID)
	aggregate.tasks = h.loadEnvironmentScheduledTasks(ctx, environmentID)
	aggregate.queues = h.loadEnvironmentQueues(ctx, environmentID)
	aggregate.cache = h.loadEnvironmentCache(ctx, environmentID)
	aggregate.checks = h.loadEnvironmentChecks(ctx, environmentID)
	aggregate.sitespeeds = h.loadEnvironmentSitespeeds(ctx, environmentID)
	aggregate.changelogs = h.loadEnvironmentChangelogs(ctx, environmentID)

	deployCount, err := h.queries.CountEnvironmentDeployments(ctx, environmentID)
	if err == nil {
		aggregate.deployCount = deployCount
	}

	return aggregate
}

func (h *Handler) loadEnvironmentExtensions(ctx context.Context, environmentID int32) []queries.EnvironmentExtension {
	rows, err := h.queries.GetEnvironmentExtensions(ctx, environmentID)
	if err != nil {
		slog.Error("failed to get environment extensions", "error", err)
		return nil
	}

	return rows
}

func (h *Handler) loadEnvironmentScheduledTasks(ctx context.Context, environmentID int32) []queries.EnvironmentScheduledTask {
	rows, err := h.queries.GetEnvironmentScheduledTasks(ctx, environmentID)
	if err != nil {
		slog.Error("failed to get environment scheduled tasks", "error", err)
		return nil
	}

	return rows
}

func (h *Handler) loadEnvironmentQueues(ctx context.Context, environmentID int32) []queries.EnvironmentQueue {
	rows, err := h.queries.GetEnvironmentQueues(ctx, environmentID)
	if err != nil {
		slog.Error("failed to get environment queues", "error", err)
		return nil
	}

	return rows
}

func (h *Handler) loadEnvironmentCache(ctx context.Context, environmentID int32) *queries.EnvironmentCache {
	row, err := h.queries.GetEnvironmentCache(ctx, environmentID)
	if err != nil {
		return nil
	}

	return &row
}

func (h *Handler) loadEnvironmentChecks(ctx context.Context, environmentID int32) []queries.EnvironmentCheck {
	rows, err := h.queries.GetEnvironmentChecks(ctx, environmentID)
	if err != nil {
		slog.Error("failed to get environment checks", "error", err)
		return nil
	}

	return rows
}

func (h *Handler) loadEnvironmentSitespeeds(ctx context.Context, environmentID int32) []queries.EnvironmentSitespeed {
	rows, err := h.queries.GetEnvironmentSitespeeds(ctx, &environmentID)
	if err != nil {
		slog.Error("failed to get environment sitespeeds", "error", err)
		return nil
	}

	return rows
}

func (h *Handler) loadEnvironmentChangelogs(ctx context.Context, environmentID int32) []queries.EnvironmentChangelog {
	rows, err := h.queries.GetEnvironmentChangelogs(ctx, &environmentID)
	if err != nil {
		slog.Error("failed to get environment changelogs", "error", err)
		return nil
	}

	return rows
}

func (h *Handler) buildEnvironmentDetail(environment *queries.GetEnvironmentByIDRow, aggregate environmentDetailAggregate) api.EnvironmentDetail {
	var ignores *[]string
	if len(environment.Ignores) > 0 {
		var ign []string
		if err := json.Unmarshal(environment.Ignores, &ign); err == nil {
			ignores = &ign
		}
	}

	var sitespeedURLs *[]string
	if len(environment.SitespeedUrls) > 0 {
		var urls []string
		if err := json.Unmarshal(environment.SitespeedUrls, &urls); err == nil {
			sitespeedURLs = &urls
		}
	}

	var lastChangelog *api.AccountChangelog
	if len(environment.LastChangelog) > 0 {
		var changelog api.AccountChangelog
		if err := json.Unmarshal(environment.LastChangelog, &changelog); err == nil && !changelog.Date.IsZero() {
			lastChangelog = &changelog
		}
	}

	shopID := int(environment.ShopID)
	var shopIDPtr *int
	if environment.ShopID > 0 {
		shopIDPtr = &shopID
	}

	return api.EnvironmentDetail{
		Id:                 int(environment.ID),
		Name:               environment.Name,
		Url:                environment.Url,
		Favicon:            environment.Favicon,
		Status:             environment.Status,
		ShopwareVersion:    environment.ShopwareVersion,
		LastScrapedAt:      pgtimeToTimePtr(environment.LastScrapedAt),
		LastScrapedError:   environment.LastScrapedError,
		OrganizationId:     environment.OrganizationID,
		OrganizationName:   environment.OrganizationName,
		ShopId:             shopIDPtr,
		ShopName:           environment.ShopName,
		ShopDescription:    environment.ShopDescription,
		EnvironmentImage:   environment.EnvironmentImage,
		EnvironmentToken:   environment.EnvironmentToken,
		Ignores:            ignores,
		CreatedAt:          pgtimeToTime(environment.CreatedAt),
		SitespeedEnabled:   environment.SitespeedEnabled,
		SitespeedDetailUrl: sitespeedDetailUrl(h.cfg, environment.ID, environment.SitespeedEnabled),
		SitespeedUrls:      sitespeedURLs,
		Extensions:         mapEnvironmentExtensions(aggregate.extensions),
		ScheduledTasks:     mapEnvironmentScheduledTasks(aggregate.tasks),
		Queues:             mapEnvironmentQueues(aggregate.queues),
		Cache:              mapEnvironmentCache(aggregate.cache),
		Checks:             mapEnvironmentChecks(aggregate.checks),
		Sitespeeds:         mapEnvironmentSitespeeds(aggregate.sitespeeds),
		Changelogs:         mapEnvironmentChangelogs(environment, aggregate.changelogs),
		DeploymentsCount:   int(aggregate.deployCount),
		LastChangelog:      lastChangelog,
	}
}

func mapEnvironmentExtensions(rows []queries.EnvironmentExtension) []api.EnvironmentExtension {
	extensions := make([]api.EnvironmentExtension, 0, len(rows))
	for _, row := range rows {
		var ratingAvg *float32
		if row.RatingAverage != nil {
			v := float32(*row.RatingAverage)
			ratingAvg = &v
		}

		var changelog *string
		if len(row.Changelog) > 0 {
			value := string(row.Changelog)
			changelog = &value
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
