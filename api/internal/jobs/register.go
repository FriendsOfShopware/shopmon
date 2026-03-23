package jobs

import (
	"context"

	goqueue "github.com/shyim/go-queue"
	"github.com/shyim/go-queue/transport/postgres"

	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/jackc/pgx/v5/pgxpool"
)

const TransportName = "async"

// Message types — plain structs for go-queue's generic dispatch.

type ShopScrapeAll struct{}
type ShopScrape struct {
	ShopID int32 `json:"shop_id"`
}
type SitespeedScrapeAll struct{}
type SitespeedScrape struct {
	ShopID int32 `json:"shop_id"`
}
type LockCleanup struct{}
type InvitationCleanup struct{}

// NewBus creates a go-queue Bus backed by PostgreSQL and registers all job handlers.
func NewBus(ctx context.Context, pool *pgxpool.Pool, q *queries.Queries, cfg *config.Config, mailSvc *mail.Service) (*goqueue.Bus, error) {
	transport := postgres.NewTransportFromPool(pool, postgres.Config{
		Table: "queue_messages",
	})

	bus := goqueue.NewBus()
	bus.AddTransport(TransportName, transport)

	shopScrape := NewShopScrapeHandler(pool, q, cfg, bus, mailSvc)
	sitespeed := NewSitespeedScrapeHandler(pool, q, cfg)
	cleanup := NewCleanupHandler(q)

	goqueue.HandleFunc(bus, TransportName, shopScrape.HandleScrapeAll)
	goqueue.HandleFunc(bus, TransportName, shopScrape.HandleScrape)
	goqueue.HandleFunc(bus, TransportName, sitespeed.HandleScrapeAll)
	goqueue.HandleFunc(bus, TransportName, sitespeed.HandleScrape)
	goqueue.HandleFunc(bus, TransportName, cleanup.HandleLockCleanup)
	goqueue.HandleFunc(bus, TransportName, cleanup.HandleInvitationCleanup)

	if err := bus.Setup(ctx); err != nil {
		return nil, err
	}

	return bus, nil
}
