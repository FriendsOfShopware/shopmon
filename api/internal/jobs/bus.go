package jobs

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	goqueue "github.com/shyim/go-queue"
	queueotel "github.com/shyim/go-queue/middleware/otel"
	"github.com/shyim/go-queue/transport/amqp"
	"github.com/shyim/go-queue/transport/postgres"
)

const (
	// DriverPostgres keeps jobs in the queue_messages table of the app database.
	DriverPostgres = "postgres"
	// DriverAMQP publishes jobs to a RabbitMQ/LavinMQ broker.
	DriverAMQP = "amqp"
)

type BusConfig struct {
	OTelEnabled bool
	// Driver selects the queue backend, DriverPostgres when empty.
	Driver string
	// AMQP is only read when Driver is DriverAMQP.
	AMQP AMQPConfig
}

// AMQPConfig describes the broker Shopmon jobs are published to and consumed
// from when the AMQP driver is selected.
type AMQPConfig struct {
	DSN           string
	Exchange      string
	Queue         string
	PrefetchCount int
	// DelayedExchange declares the exchange as x-delayed-message so the broker
	// holds back delayed jobs instead of delivering them immediately. Requires
	// LavinMQ (native) or the RabbitMQ delayed-message plugin.
	DelayedExchange bool
}

// NewBus creates and sets up the job bus for the configured driver. It
// intentionally does not register handlers: API and fixture processes only
// dispatch messages, while the worker adds executable handlers through
// RegisterHandlers.
//
// Two transports are registered: the legacy async/catalog queue and the
// shop-facing scrape queue. See ScrapeTransportName for why they are split.
//
// The returned close function releases resources the bus owns itself (the AMQP
// connections). It never touches the caller-owned pgx pool, so callers keep
// closing that themselves.
func NewBus(ctx context.Context, pool *pgxpool.Pool, config BusConfig) (*goqueue.Bus, func() error, error) {
	transports, closeTransport, err := newTransports(pool, config)
	if err != nil {
		return nil, nil, err
	}

	bus := goqueue.NewBus()
	if config.OTelEnabled {
		bus.AddDispatchMiddleware(queueotel.DispatchMiddleware(
			queueotel.WithSpanNameNormalizer(queueotel.DefaultSpanNameNormalizer),
		))
	}
	for name, transport := range transports {
		bus.AddTransport(name, transport)
	}

	if err := bus.Setup(ctx); err != nil {
		_ = closeTransport()
		return nil, nil, fmt.Errorf("setup job bus: %w", err)
	}
	return bus, closeTransport, nil
}

func newTransport(pool *pgxpool.Pool, config BusConfig) (goqueue.Transport, func() error, error) {
	transports, closeTransport, err := newTransports(pool, config)
	if err != nil {
		return nil, nil, err
	}
	return transports[TransportName], closeTransport, nil
}

func newTransports(pool *pgxpool.Pool, config BusConfig) (map[string]goqueue.Transport, func() error, error) {
	noopClose := func() error { return nil }

	switch config.Driver {
	case "", DriverPostgres:
		if pool == nil {
			return nil, nil, fmt.Errorf("job bus: postgres driver requires a database pool")
		}
		// The pool is shared with the rest of the process, so the transports must
		// not be closed with the bus — postgres.Transport.Close() closes the pool.
		return map[string]goqueue.Transport{
			TransportName: postgres.NewTransportFromPool(pool, postgres.Config{
				Table:   "queue_messages",
				Channel: "queue_notify",
			}),
			ScrapeTransportName: postgres.NewTransportFromPool(pool, postgres.Config{
				Table:   "queue_messages_scrape",
				Channel: "queue_notify_scrape",
			}),
		}, noopClose, nil
	case DriverAMQP:
		if config.AMQP.DSN == "" {
			return nil, nil, fmt.Errorf("job bus: amqp driver requires a broker DSN")
		}
		catalog := newAMQPTransport(config.AMQP, config.AMQP.Queue)
		scrapeQueue := config.AMQP.Queue + "-scrape"
		scrape := newAMQPTransport(config.AMQP, scrapeQueue)
		return map[string]goqueue.Transport{
			TransportName:       catalog,
			ScrapeTransportName: scrape,
		}, func() error {
			err := catalog.Close()
			if scrapeErr := scrape.Close(); scrapeErr != nil && err == nil {
				err = scrapeErr
			}
			return err
		}, nil
	default:
		return nil, nil, fmt.Errorf("job bus: unknown queue driver %q (supported: %s, %s)", config.Driver, DriverPostgres, DriverAMQP)
	}
}

func newAMQPTransport(config AMQPConfig, queue string) *amqp.Transport {
	return amqp.NewTransport(amqp.Config{
		DSN:           config.DSN,
		Exchange:      config.Exchange,
		Queue:         queue,
		RoutingKey:    queue,
		PrefetchCount: config.PrefetchCount,
		Durable:       true,
		// Re-declare exchange, queue, and binding after a broker restart.
		AutoSetup:         true,
		PublisherConfirms: true,
		DelayedExchange:   config.DelayedExchange,
	})
}

// transportFor returns the queue a Shopmon job type is published to.
// StoreExtensionSync stays on the legacy async transport; every other job
// uses the scrape transport so catalog work cannot occupy the shop-refresh path.
func transportFor[T any]() string {
	var zero T
	if _, ok := any(zero).(StoreExtensionSync); ok {
		return TransportName
	}
	return ScrapeTransportName
}

// Dispatch routes a Shopmon job to its dedicated transport.
// go-queue otherwise derives the route from a registered handler, which would
// force dispatch-only processes to construct the complete worker graph.
func Dispatch[T any](ctx context.Context, bus *goqueue.Bus, message T, options ...goqueue.Option) error {
	if bus == nil {
		return fmt.Errorf("dispatch job: bus is required")
	}
	routedOptions := make([]goqueue.Option, 0, len(options)+1)
	routedOptions = append(routedOptions, options...)
	routedOptions = append(routedOptions, goqueue.WithQueue(transportFor[T]()))
	return goqueue.Dispatch(ctx, bus, message, routedOptions...)
}
