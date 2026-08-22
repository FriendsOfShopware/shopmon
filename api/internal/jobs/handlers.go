package jobs

import (
	"context"
	"errors"
	"log/slog"

	"github.com/friendsofshopware/shopmon/api/internal/shopwareaccount"
	goqueue "github.com/shyim/go-queue"
)

type EnvironmentScraper interface {
	Scrape(ctx context.Context, environmentID int32) error
}

type StoreExtensionSynchronizer interface {
	Sync(ctx context.Context, names []string, shopwareVersion string) error
}

type SitespeedScraper interface {
	Scrape(ctx context.Context, environmentID int32) error
}

type CleanupHandler interface {
	CleanupLocks(ctx context.Context) error
	CleanupInvitations(ctx context.Context) error
	CleanupOldData(ctx context.Context) error
}

type ChangelogSynchronizer interface {
	Sync(ctx context.Context) error
}

type SecurityPluginSynchronizer interface {
	Sync(ctx context.Context) error
}

type AdvisorySynchronizer interface {
	Sync(ctx context.Context) error
}

// Handlers is the worker-side composition boundary. Construct these services
// in the worker root; dispatch-only processes do not need them.
type Handlers struct {
	EnvironmentScraper         EnvironmentScraper
	StoreExtensionSynchronizer StoreExtensionSynchronizer
	SitespeedScraper           SitespeedScraper
	Cleanup                    CleanupHandler
	ChangelogSynchronizer      ChangelogSynchronizer
	AdvisorySynchronizer       AdvisorySynchronizer
	SecurityPluginSynchronizer SecurityPluginSynchronizer
}

func RegisterHandlers(bus *goqueue.Bus, handlers Handlers) error {
	if bus == nil {
		return errors.New("register job handlers: bus is required")
	}
	if handlers.EnvironmentScraper == nil {
		return errors.New("register job handlers: environment scraper is required")
	}
	if handlers.StoreExtensionSynchronizer == nil {
		return errors.New("register job handlers: store extension synchronizer is required")
	}
	if handlers.SitespeedScraper == nil {
		return errors.New("register job handlers: sitespeed scraper is required")
	}
	if handlers.Cleanup == nil {
		return errors.New("register job handlers: cleanup handler is required")
	}
	if handlers.ChangelogSynchronizer == nil {
		return errors.New("register job handlers: changelog synchronizer is required")
	}
	if handlers.AdvisorySynchronizer == nil {
		return errors.New("register job handlers: advisory synchronizer is required")
	}
	if handlers.SecurityPluginSynchronizer == nil {
		return errors.New("register job handlers: security plugin synchronizer is required")
	}

	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, message EnvironmentScrape) error {
		// Annotate the go-queue otel Consumer span before Scrape starts a child
		// Internal span, so Datadog can facet EnvironmentScrape process by environment.id.
		return runEnvironmentJob(ctx, message.EnvironmentID, func(ctx context.Context) error {
			return handlers.EnvironmentScraper.Scrape(ctx, message.EnvironmentID)
		})
	})
	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, message StoreExtensionSync) error {
		// Names are high-cardinality; keep them off the Consumer entry span.
		return handleStoreExtensionSync(ctx, handlers.StoreExtensionSynchronizer, message)
	})
	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, message SitespeedScrape) error {
		return runEnvironmentJob(ctx, message.EnvironmentID, func(ctx context.Context) error {
			return handlers.SitespeedScraper.Scrape(ctx, message.EnvironmentID)
		})
	})
	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, _ LockCleanup) error {
		return handlers.Cleanup.CleanupLocks(ctx)
	})
	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, _ InvitationCleanup) error {
		return handlers.Cleanup.CleanupInvitations(ctx)
	})
	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, _ OldDataCleanup) error {
		return handlers.Cleanup.CleanupOldData(ctx)
	})
	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, _ ShopwareChangelogSync) error {
		return handlers.ChangelogSynchronizer.Sync(ctx)
	})
	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, _ ComposerAdvisorySync) error {
		return handlers.AdvisorySynchronizer.Sync(ctx)
	})

	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, _ SecurityPluginSync) error {
		return handlers.SecurityPluginSynchronizer.Sync(ctx)
	})
	return nil
}

// handleStoreExtensionSync runs catalog sync and acks Store API rate-limit
// aborts so the worker does not nack/retry into a still-limited API. Incomplete
// names stay eligible for the next scheduled scrape dispatch.
func handleStoreExtensionSync(ctx context.Context, syncer StoreExtensionSynchronizer, message StoreExtensionSync) error {
	err := syncer.Sync(ctx, message.Names, message.ShopwareVersion)
	if shopwareaccount.IsRateLimited(err) {
		slog.Warn("store extension sync rate limited, acknowledging job for later scheduled sync", "error", err)
		return nil
	}
	return err
}
