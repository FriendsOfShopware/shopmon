package jobs

import (
	"context"
	"errors"

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

// Handlers is the worker-side composition boundary. Construct these services
// in the worker root; dispatch-only processes do not need them.
type Handlers struct {
	EnvironmentScraper         EnvironmentScraper
	StoreExtensionSynchronizer StoreExtensionSynchronizer
	SitespeedScraper           SitespeedScraper
	Cleanup                    CleanupHandler
	ChangelogSynchronizer      ChangelogSynchronizer
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

	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, message EnvironmentScrape) error {
		return handlers.EnvironmentScraper.Scrape(ctx, message.EnvironmentID)
	})
	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, message StoreExtensionSync) error {
		return handlers.StoreExtensionSynchronizer.Sync(ctx, message.Names, message.ShopwareVersion)
	})
	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, message SitespeedScrape) error {
		return handlers.SitespeedScraper.Scrape(ctx, message.EnvironmentID)
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
	return nil
}
