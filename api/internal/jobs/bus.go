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
// The returned close function releases resources the bus owns itself (the AMQP
// connection). It never touches the caller-owned pgx pool, so callers keep
// closing that themselves.
func NewBus(ctx context.Context, pool *pgxpool.Pool, config BusConfig) (*goqueue.Bus, func() error, error) {
	transport, closeTransport, err := newTransport(pool, config)
	if err != nil {
		return nil, nil, err
	}

	bus := goqueue.NewBus()
	if config.OTelEnabled {
		bus.AddDispatchMiddleware(queueotel.DispatchMiddleware(
			queueotel.WithSpanNameNormalizer(queueotel.DefaultSpanNameNormalizer),
		))
	}
	bus.AddTransport(TransportName, transport)

	if err := bus.Setup(ctx); err != nil {
		_ = closeTransport()
		return nil, nil, fmt.Errorf("setup job bus: %w", err)
	}
	return bus, closeTransport, nil
}

func newTransport(pool *pgxpool.Pool, config BusConfig) (goqueue.Transport, func() error, error) {
	noopClose := func() error { return nil }

	switch config.Driver {
	case "", DriverPostgres:
		if pool == nil {
			return nil, nil, fmt.Errorf("job bus: postgres driver requires a database pool")
		}
		// The pool is shared with the rest of the process, so the transport must
		// not be closed with the bus — postgres.Transport.Close() closes the pool.
		return postgres.NewTransportFromPool(pool, postgres.Config{
			Table: "queue_messages",
		}), noopClose, nil
	case DriverAMQP:
		if config.AMQP.DSN == "" {
			return nil, nil, fmt.Errorf("job bus: amqp driver requires a broker DSN")
		}
		transport := amqp.NewTransport(amqp.Config{
			DSN:           config.AMQP.DSN,
			Exchange:      config.AMQP.Exchange,
			Queue:         config.AMQP.Queue,
			RoutingKey:    config.AMQP.Queue,
			PrefetchCount: config.AMQP.PrefetchCount,
			Durable:       true,
			// Re-declare exchange, queue, and binding after a broker restart.
			AutoSetup:         true,
			PublisherConfirms: true,
			DelayedExchange:   config.AMQP.DelayedExchange,
		})
		return transport, transport.Close, nil
	default:
		return nil, nil, fmt.Errorf("job bus: unknown queue driver %q (supported: %s, %s)", config.Driver, DriverPostgres, DriverAMQP)
	}
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
