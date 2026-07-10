package jobs

import (
	"context"

	goqueue "github.com/shyim/go-queue"
	queueotel "github.com/shyim/go-queue/middleware/otel"
	"github.com/shyim/go-queue/transport/postgres"

	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
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

	storeSync := NewStoreExtensionSyncHandler(pool, q, cfg)
	envScrape := NewEnvironmentScrapeHandler(pool, q, cfg, bus, mailSvc, storeSync)
	sitespeed := NewSitespeedScrapeHandler(pool, q, cfg)
	cleanup := NewCleanupHandler(q)
	changelog := NewShopwareChangelogHandler(q, cfg)

	goqueue.HandleFunc(bus, TransportName, envScrape.HandleScrape)
	goqueue.HandleFunc(bus, TransportName, storeSync.HandleSync)
	goqueue.HandleFunc(bus, TransportName, sitespeed.HandleScrape)
	goqueue.HandleFunc(bus, TransportName, cleanup.HandleLockCleanup)
	goqueue.HandleFunc(bus, TransportName, cleanup.HandleInvitationCleanup)
	goqueue.HandleFunc(bus, TransportName, cleanup.HandleOldDataCleanup)
	goqueue.HandleFunc(bus, TransportName, changelog.HandleSync)

	if err := bus.Setup(ctx); err != nil {
		return nil, err
	}

	return bus, nil
}
