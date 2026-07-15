package jobs

import (
	"context"
	"fmt"
	"time"

	goqueue "github.com/shyim/go-queue"
	queueotel "github.com/shyim/go-queue/middleware/otel"
	"github.com/shyim/go-queue/transport/postgres"

	catalogchangelog "github.com/friendsofshopware/shopmon/api/internal/catalog/changelog"
	catalogsync "github.com/friendsofshopware/shopmon/api/internal/catalog/sync"
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/friendsofshopware/shopmon/api/internal/maintenance"
	monitoringscrape "github.com/friendsofshopware/shopmon/api/internal/monitoring/scrape"
	"github.com/friendsofshopware/shopmon/api/internal/monitoring/sitespeed"
	"github.com/jackc/pgx/v5/pgxpool"
)

const TransportName = "async"

// Message types — plain structs for go-queue's generic dispatch.

type EnvironmentScrape struct {
	EnvironmentID int32 `json:"environment_id"`
}
type SitespeedScrape struct {
	EnvironmentID int32 `json:"environment_id"`
}
type LockCleanup struct{}
type InvitationCleanup struct{}
type OldDataCleanup struct{}
type ShopwareChangelogSync struct{}

// StoreExtensionSync refreshes the shared store extension catalog for a set of
// extension names. ShopwareVersion is the version of the environment that
// requested the sync; it is included in the compatibility probing.
type StoreExtensionSync struct {
	Names           []string `json:"names"`
	ShopwareVersion string   `json:"shopware_version"`
}

// NewBus creates a go-queue Bus backed by PostgreSQL and registers all job handlers.
func NewBus(ctx context.Context, pool *pgxpool.Pool, q *queries.Queries, cfg *config.Config, mailSvc mail.Sender) (*goqueue.Bus, error) {
	transport := postgres.NewTransportFromPool(pool, postgres.Config{
		Table: "queue_messages",
	})

	bus := goqueue.NewBus()

	if cfg.OtelEnabled {
		bus.AddDispatchMiddleware(queueotel.DispatchMiddleware(
			queueotel.WithSpanNameNormalizer(queueotel.DefaultSpanNameNormalizer),
		))
	}

	bus.AddTransport(TransportName, transport)

	storeSync := catalogsync.NewService(pool, q, cfg)
	dispatcher := &scrapeDispatcher{bus: bus}
	environmentScrape := monitoringscrape.NewService(pool, q, cfg, mailSvc, storeSync, dispatcher)
	sitespeedScrape := sitespeed.NewService(q, cfg)
	cleanup := maintenance.NewService(q)
	changelog := catalogchangelog.NewService(q, cfg)

	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, message EnvironmentScrape) error {
		return environmentScrape.Scrape(ctx, message.EnvironmentID)
	})
	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, message StoreExtensionSync) error {
		return storeSync.Sync(ctx, message.Names, message.ShopwareVersion)
	})
	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, message SitespeedScrape) error {
		return sitespeedScrape.Scrape(ctx, message.EnvironmentID)
	})
	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, _ LockCleanup) error {
		return cleanup.CleanupLocks(ctx)
	})
	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, _ InvitationCleanup) error {
		return cleanup.CleanupInvitations(ctx)
	})
	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, _ OldDataCleanup) error {
		return cleanup.CleanupOldData(ctx)
	})
	goqueue.HandleFunc(bus, TransportName, func(ctx context.Context, _ ShopwareChangelogSync) error {
		return changelog.Sync(ctx)
	})

	if err := bus.Setup(ctx); err != nil {
		return nil, fmt.Errorf("setup job bus: %w", err)
	}

	return bus, nil
}

type scrapeDispatcher struct {
	bus *goqueue.Bus
}

var _ monitoringscrape.Dispatcher = (*scrapeDispatcher)(nil)

func (d *scrapeDispatcher) DispatchCatalogSync(ctx context.Context, names []string, shopwareVersion string) error {
	if err := goqueue.Dispatch(ctx, d.bus, StoreExtensionSync{Names: names, ShopwareVersion: shopwareVersion}); err != nil {
		return fmt.Errorf("dispatch store extension sync: %w", err)
	}
	return nil
}

func (d *scrapeDispatcher) DispatchSitespeedScrape(ctx context.Context, environmentID int32, delay time.Duration) error {
	if err := goqueue.Dispatch(ctx, d.bus, SitespeedScrape{EnvironmentID: environmentID}, goqueue.WithDelay(delay)); err != nil {
		return fmt.Errorf("dispatch sitespeed scrape: %w", err)
	}
	return nil
}
