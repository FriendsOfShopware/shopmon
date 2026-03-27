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

type shopDetailAggregate struct {
	extensions  []queries.ShopExtension
	tasks       []queries.ShopScheduledTask
	queues      []queries.ShopQueue
	cache       *queries.ShopCache
	checks      []queries.ShopCheck
	sitespeeds  []queries.ShopSitespeed
	changelogs  []queries.ShopChangelog
	deployCount int32
}

func (h *Handler) loadShopDetailAggregate(ctx context.Context, shopID int32) shopDetailAggregate {
	aggregate := shopDetailAggregate{}

	aggregate.extensions = h.loadShopExtensions(ctx, shopID)
	aggregate.tasks = h.loadShopScheduledTasks(ctx, shopID)
	aggregate.queues = h.loadShopQueues(ctx, shopID)
	aggregate.cache = h.loadShopCache(ctx, shopID)
	aggregate.checks = h.loadShopChecks(ctx, shopID)
	aggregate.sitespeeds = h.loadShopSitespeeds(ctx, shopID)
	aggregate.changelogs = h.loadShopChangelogs(ctx, shopID)

	deployCount, err := h.queries.CountShopDeployments(ctx, shopID)
	if err == nil {
		aggregate.deployCount = deployCount
	}

	return aggregate
}

func (h *Handler) loadShopExtensions(ctx context.Context, shopID int32) []queries.ShopExtension {
	rows, err := h.queries.GetShopExtensions(ctx, shopID)
	if err != nil {
		slog.Error("failed to get shop extensions", "error", err)
		return nil
	}

	return rows
}

func (h *Handler) loadShopScheduledTasks(ctx context.Context, shopID int32) []queries.ShopScheduledTask {
	rows, err := h.queries.GetShopScheduledTasks(ctx, shopID)
	if err != nil {
		slog.Error("failed to get shop scheduled tasks", "error", err)
		return nil
	}

	return rows
}

func (h *Handler) loadShopQueues(ctx context.Context, shopID int32) []queries.ShopQueue {
	rows, err := h.queries.GetShopQueues(ctx, shopID)
	if err != nil {
		slog.Error("failed to get shop queues", "error", err)
		return nil
	}

	return rows
}

func (h *Handler) loadShopCache(ctx context.Context, shopID int32) *queries.ShopCache {
	row, err := h.queries.GetShopCache(ctx, shopID)
	if err != nil {
		return nil
	}

	return &row
}

func (h *Handler) loadShopChecks(ctx context.Context, shopID int32) []queries.ShopCheck {
	rows, err := h.queries.GetShopChecks(ctx, shopID)
	if err != nil {
		slog.Error("failed to get shop checks", "error", err)
		return nil
	}

	return rows
}

func (h *Handler) loadShopSitespeeds(ctx context.Context, shopID int32) []queries.ShopSitespeed {
	rows, err := h.queries.GetShopSitespeeds(ctx, &shopID)
	if err != nil {
		slog.Error("failed to get shop sitespeeds", "error", err)
		return nil
	}

	return rows
}

func (h *Handler) loadShopChangelogs(ctx context.Context, shopID int32) []queries.ShopChangelog {
	rows, err := h.queries.GetShopChangelogs(ctx, &shopID)
	if err != nil {
		slog.Error("failed to get shop changelogs", "error", err)
		return nil
	}

	return rows
}

func (h *Handler) buildShopDetail(shop *queries.GetShopByIDRow, aggregate shopDetailAggregate) api.ShopDetail {
	var ignores *[]string
	if shop.Ignores != nil && len(shop.Ignores) > 0 {
		var ign []string
		if err := json.Unmarshal(shop.Ignores, &ign); err == nil {
			ignores = &ign
		}
	}

	var sitespeedURLs *[]string
	if shop.SitespeedUrls != nil && len(shop.SitespeedUrls) > 0 {
		var urls []string
		if err := json.Unmarshal(shop.SitespeedUrls, &urls); err == nil {
			sitespeedURLs = &urls
		}
	}

	var lastChangelog *api.AccountChangelog
	if shop.LastChangelog != nil && len(shop.LastChangelog) > 0 {
		var changelog api.AccountChangelog
		if err := json.Unmarshal(shop.LastChangelog, &changelog); err == nil && !changelog.Date.IsZero() {
			lastChangelog = &changelog
		}
	}

	projectID := int(shop.ProjectID)
	var projectIDPtr *int
	if shop.ProjectID > 0 {
		projectIDPtr = &projectID
	}

	return api.ShopDetail{
		Id:                 int(shop.ID),
		Name:               shop.Name,
		Url:                shop.Url,
		Favicon:            shop.Favicon,
		Status:             shop.Status,
		ShopwareVersion:    shop.ShopwareVersion,
		LastScrapedAt:      pgtimeToTimePtr(shop.LastScrapedAt),
		LastScrapedError:   shop.LastScrapedError,
		OrganizationId:     shop.OrganizationID,
		OrganizationName:   shop.OrganizationName,
		ProjectId:          projectIDPtr,
		ProjectName:        shop.ProjectName,
		ProjectDescription: shop.ProjectDescription,
		ShopImage:          shop.ShopImage,
		ShopToken:          shop.ShopToken,
		Ignores:            ignores,
		CreatedAt:          pgtimeToTime(shop.CreatedAt),
		SitespeedEnabled:   shop.SitespeedEnabled,
		SitespeedDetailUrl: sitespeedDetailUrl(h.cfg, shop.ID, shop.SitespeedEnabled),
		SitespeedUrls:      sitespeedURLs,
		Extensions:         mapShopExtensions(aggregate.extensions),
		ScheduledTasks:     mapShopScheduledTasks(aggregate.tasks),
		Queues:             mapShopQueues(aggregate.queues),
		Cache:              mapShopCache(aggregate.cache),
		Checks:             mapShopChecks(aggregate.checks),
		Sitespeeds:         mapShopSitespeeds(aggregate.sitespeeds),
		Changelogs:         mapShopChangelogs(shop, aggregate.changelogs),
		DeploymentsCount:   int(aggregate.deployCount),
		LastChangelog:      lastChangelog,
	}
}

func mapShopExtensions(rows []queries.ShopExtension) []api.ShopExtension {
	extensions := make([]api.ShopExtension, 0, len(rows))
	for _, row := range rows {
		var ratingAvg *float32
		if row.RatingAverage != nil {
			v := float32(*row.RatingAverage)
			ratingAvg = &v
		}

		var changelog *string
		if row.Changelog != nil && len(row.Changelog) > 0 {
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

		extensions = append(extensions, api.ShopExtension{
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

func mapShopScheduledTasks(rows []queries.ShopScheduledTask) []api.ScheduledTask {
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

func mapShopQueues(rows []queries.ShopQueue) []api.Queue {
	queues := make([]api.Queue, 0, len(rows))
	for _, row := range rows {
		queues = append(queues, api.Queue{
			Name: row.Name,
			Size: int(row.Size),
		})
	}

	return queues
}

func mapShopCache(row *queries.ShopCache) *api.CacheInfo {
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

func mapShopChecks(rows []queries.ShopCheck) []api.ShopCheck {
	checks := make([]api.ShopCheck, 0, len(rows))
	for _, row := range rows {
		checks = append(checks, api.ShopCheck{
			Id:      row.CheckID,
			Level:   row.Level,
			Message: row.Message,
			Link:    row.Link,
		})
	}

	return checks
}

func mapShopSitespeeds(rows []queries.ShopSitespeed) []api.Sitespeed {
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

func mapShopChangelogs(shop *queries.GetShopByIDRow, rows []queries.ShopChangelog) []api.AccountChangelog {
	changelogs := make([]api.AccountChangelog, 0, len(rows))
	for _, row := range rows {
		shopID := 0
		if row.ShopID != nil {
			shopID = int(*row.ShopID)
		}

		var extensions []api.ExtensionDiff
		if row.Extensions != nil {
			_ = json.Unmarshal(row.Extensions, &extensions)
		}
		if extensions == nil {
			extensions = []api.ExtensionDiff{}
		}

		changelogs = append(changelogs, api.AccountChangelog{
			Id:                   int(row.ID),
			ShopId:               shopID,
			ShopName:             shop.Name,
			ShopOrganizationName: shop.OrganizationName,
			ShopOrganizationId:   shop.OrganizationID,
			Extensions:           extensions,
			OldShopwareVersion:   row.OldShopwareVersion,
			NewShopwareVersion:   row.NewShopwareVersion,
			Date:                 pgtimeToTime(row.Date),
		})
	}

	return changelogs
}
