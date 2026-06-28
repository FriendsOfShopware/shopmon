package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/jackc/pgx/v5"
)

type environmentDetailAggregate struct {
	extensions  []api.EnvironmentExtension
	tasks       []queries.EnvironmentScheduledTask
	queues      []queries.EnvironmentQueue
	cache       *queries.EnvironmentCache
	checks      []queries.EnvironmentCheck
	sitespeeds  []queries.EnvironmentSitespeed
	changelogs  []queries.EnvironmentChangelog
	deployCount int32
	subscribed  bool
}

func (h *Handler) loadEnvironmentDetailAggregate(ctx context.Context, environmentID int32, language string) environmentDetailAggregate {
	aggregate := environmentDetailAggregate{}

	var wg sync.WaitGroup
	wg.Add(8)

	go func() {
		defer wg.Done()
		aggregate.extensions = h.loadEnvironmentExtensions(ctx, environmentID, language)
	}()
	go func() {
		defer wg.Done()
		aggregate.tasks = h.loadEnvironmentScheduledTasks(ctx, environmentID)
	}()
	go func() {
		defer wg.Done()
		aggregate.queues = h.loadEnvironmentQueues(ctx, environmentID)
	}()
	go func() {
		defer wg.Done()
		aggregate.cache = h.loadEnvironmentCache(ctx, environmentID)
	}()
	go func() {
		defer wg.Done()
		aggregate.checks = h.loadEnvironmentChecks(ctx, environmentID)
	}()
	go func() {
		defer wg.Done()
		aggregate.sitespeeds = h.loadEnvironmentSitespeeds(ctx, environmentID)
	}()
	go func() {
		defer wg.Done()
		aggregate.changelogs = h.loadEnvironmentChangelogs(ctx, environmentID)
	}()
	go func() {
		defer wg.Done()
		deployCount, err := h.queries.CountEnvironmentDeployments(ctx, environmentID)
		if err == nil {
			aggregate.deployCount = deployCount
		}
	}()

	wg.Wait()

	return aggregate
}

func (h *Handler) loadEnvironmentExtensions(ctx context.Context, environmentID int32, language string) []api.EnvironmentExtension {
	// The store and unknown queries together make up the full extension list, so
	// a failure in either must not silently produce a partial list (which would
	// look like a successful empty/short response). Bail out like the sibling
	// load* helpers instead.
	storeRows, err := h.queries.GetEnvironmentStoreExtensions(ctx, queries.GetEnvironmentStoreExtensionsParams{
		EnvironmentID: environmentID,
		Language:      language,
	})
	if err != nil {
		slog.Error("failed to get environment store extensions", "error", err)
		return nil
	}
	unknownRows, err := h.queries.GetEnvironmentExtensions(ctx, environmentID)
	if err != nil {
		slog.Error("failed to get environment extensions", "error", err)
		return nil
	}
	// The changelog is supplementary; on failure we still return the extensions
	// without it rather than dropping the whole list.
	changelogRows, err := h.queries.GetEnvironmentStoreExtensionChangelogs(ctx, queries.GetEnvironmentStoreExtensionChangelogsParams{
		EnvironmentID: environmentID,
		Language:      language,
	})
	if err != nil {
		slog.Error("failed to get environment store extension changelogs", "error", err)
	}

	return mergeEnvironmentExtensions(storeRows, unknownRows, changelogRows)
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
		if !errors.Is(err, pgx.ErrNoRows) {
			slog.Error("failed to get environment cache", "error", err)
		}
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
		Extensions:         aggregate.extensions,
		ScheduledTasks:     mapEnvironmentScheduledTasks(aggregate.tasks),
		Queues:             mapEnvironmentQueues(aggregate.queues),
		Cache:              mapEnvironmentCache(aggregate.cache),
		Checks:             mapEnvironmentChecks(aggregate.checks),
		Sitespeeds:         mapEnvironmentSitespeeds(aggregate.sitespeeds),
		Changelogs:         mapEnvironmentChangelogs(environment, aggregate.changelogs),
		DeploymentsCount:   int(aggregate.deployCount),
		LastChangelog:      lastChangelog,
		Subscribed:         aggregate.subscribed,
	}
}
