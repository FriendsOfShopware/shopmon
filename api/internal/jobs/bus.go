package jobs

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	goqueue "github.com/shyim/go-queue"
	queueotel "github.com/shyim/go-queue/middleware/otel"
	"github.com/shyim/go-queue/transport/postgres"
)

type BusConfig struct {
	OTelEnabled bool
}

// NewBus creates and sets up a PostgreSQL-backed bus. It intentionally does
// not register handlers: API and fixture processes only dispatch messages,
// while the worker adds executable handlers through RegisterHandlers.
func NewBus(ctx context.Context, pool *pgxpool.Pool, config BusConfig) (*goqueue.Bus, error) {
	transport := postgres.NewTransportFromPool(pool, postgres.Config{
		Table: "queue_messages",
	})

	bus := goqueue.NewBus()
	if config.OTelEnabled {
		bus.AddDispatchMiddleware(queueotel.DispatchMiddleware(
			queueotel.WithSpanNameNormalizer(queueotel.DefaultSpanNameNormalizer),
		))
	}
	bus.AddTransport(TransportName, transport)

	if err := bus.Setup(ctx); err != nil {
		return nil, fmt.Errorf("setup job bus: %w", err)
	}
	return bus, nil
}

// Dispatch routes a Shopmon job explicitly to the stable async transport.
// go-queue otherwise derives the route from a registered handler, which would
// force dispatch-only processes to construct the complete worker graph.
func Dispatch[T any](ctx context.Context, bus *goqueue.Bus, message T, options ...goqueue.Option) error {
	if bus == nil {
		return fmt.Errorf("dispatch job: bus is required")
	}
	routedOptions := make([]goqueue.Option, 0, len(options)+1)
	routedOptions = append(routedOptions, options...)
	routedOptions = append(routedOptions, goqueue.WithQueue(TransportName))
	return goqueue.Dispatch(ctx, bus, message, routedOptions...)
}
