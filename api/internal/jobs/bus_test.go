package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	goqueue "github.com/shyim/go-queue"
	"github.com/shyim/go-queue/transport/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const messageTypePrefix = "github.com/friendsofshopware/shopmon/api/internal/jobs."

func TestDispatchRoutesMessagesWithoutRegisteredHandlers(t *testing.T) {
	bus, transport := newMemoryBus()

	tests := []struct {
		name          string
		expectedType  string
		dispatch      func() error
		expectedDelay time.Duration
	}{
		{
			name:         "environment scrape",
			expectedType: messageTypePrefix + "EnvironmentScrape",
			dispatch: func() error {
				return Dispatch(context.Background(), bus, EnvironmentScrape{EnvironmentID: 42})
			},
		},
		{
			name:          "delayed sitespeed scrape",
			expectedType:  messageTypePrefix + "SitespeedScrape",
			expectedDelay: 15 * time.Minute,
			dispatch: func() error {
				return Dispatch(context.Background(), bus, SitespeedScrape{EnvironmentID: 42}, goqueue.WithDelay(15*time.Minute))
			},
		},
		{
			name:         "catalog sync",
			expectedType: messageTypePrefix + "StoreExtensionSync",
			dispatch: func() error {
				return Dispatch(context.Background(), bus, StoreExtensionSync{Names: []string{"Example"}, ShopwareVersion: "6.7"})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.NoError(t, test.dispatch())
			envelope := <-transport.Chan()
			assert.Equal(t, TransportName, envelope.Transport)
			assert.Equal(t, test.expectedType, envelope.Type)
			_, registered := bus.GetHandler(envelope.Type)
			assert.False(t, registered, "dispatch-only bus must not require a worker handler")

			delay, delayed := goqueue.GetStamp[goqueue.DelayStamp](envelope)
			if test.expectedDelay == 0 {
				assert.False(t, delayed)
			} else {
				require.True(t, delayed)
				assert.Equal(t, test.expectedDelay, delay.Delay)
			}
		})
	}
}

func TestRegisterHandlersRegistersEveryStableMessageType(t *testing.T) {
	bus, transport := newMemoryBus()
	require.NoError(t, RegisterHandlers(bus, completeHandlers()))

	dispatches := []func() error{
		func() error { return Dispatch(context.Background(), bus, EnvironmentScrape{}) },
		func() error { return Dispatch(context.Background(), bus, SitespeedScrape{}) },
		func() error { return Dispatch(context.Background(), bus, LockCleanup{}) },
		func() error { return Dispatch(context.Background(), bus, InvitationCleanup{}) },
		func() error { return Dispatch(context.Background(), bus, OldDataCleanup{}) },
		func() error { return Dispatch(context.Background(), bus, ShopwareChangelogSync{}) },
		func() error { return Dispatch(context.Background(), bus, StoreExtensionSync{}) },
	}

	for _, dispatch := range dispatches {
		require.NoError(t, dispatch())
		envelope := <-transport.Chan()
		_, registered := bus.GetHandler(envelope.Type)
		assert.True(t, registered, "worker handler missing for %s", envelope.Type)
	}
}

func TestRegisterHandlersRejectsIncompleteWorkerGraph(t *testing.T) {
	bus, _ := newMemoryBus()
	require.Error(t, RegisterHandlers(bus, Handlers{}))
	require.Error(t, RegisterHandlers(nil, completeHandlers()))
}

func TestNewTransportSelectsDriver(t *testing.T) {
	amqpConfig := AMQPConfig{DSN: "amqp://guest:guest@localhost:5672/", Exchange: "shopmon", Queue: "shopmon"}

	tests := []struct {
		name        string
		config      BusConfig
		expectedErr string
	}{
		{name: "amqp driver", config: BusConfig{Driver: DriverAMQP, AMQP: amqpConfig}},
		{name: "amqp driver without dsn", config: BusConfig{Driver: DriverAMQP}, expectedErr: "requires a broker DSN"},
		// The pool is nil in this test, so only the driver dispatch is asserted:
		// the postgres transport is the one branch that needs it.
		{name: "postgres driver without pool", config: BusConfig{Driver: DriverPostgres}, expectedErr: "requires a database pool"},
		{name: "empty driver defaults to postgres", config: BusConfig{}, expectedErr: "requires a database pool"},
		{name: "unknown driver", config: BusConfig{Driver: "kafka"}, expectedErr: `unknown queue driver "kafka"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transport, closeTransport, err := newTransport(nil, test.config)

			if test.expectedErr != "" {
				assert.ErrorContains(t, err, test.expectedErr)
				assert.Nil(t, transport)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, transport)
			// Constructing the AMQP transport must not dial the broker; only
			// bus.Setup does, so dispatch-only processes fail fast with a clear error.
			require.NoError(t, closeTransport())
		})
	}
}

func TestNewTransportKeepsPostgresPoolOpen(t *testing.T) {
	// postgres.Transport.Close() closes the pool it was handed, which the rest of
	// the process still uses — NewBus must therefore hand back a no-op closer.
	pool, err := pgxpool.New(context.Background(), "postgres://user:pass@127.0.0.1:1/db")
	require.NoError(t, err)
	defer pool.Close()

	_, closeTransport, err := newTransport(pool, BusConfig{Driver: DriverPostgres})
	require.NoError(t, err)
	require.NoError(t, closeTransport())

	// A closed pool rejects acquisition with "closed pool"; a live one gets as far
	// as dialing the (unreachable) address instead.
	_, err = pool.Acquire(context.Background())
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "closed pool")
}

func newMemoryBus() (*goqueue.Bus, *memory.Transport) {
	bus := goqueue.NewBus()
	transport := memory.NewTransport()
	bus.AddTransport(TransportName, transport)
	return bus, transport
}

type scraperStub struct{}

func (scraperStub) Scrape(context.Context, int32) error { return nil }

type storeExtensionSynchronizerStub struct{}

func (storeExtensionSynchronizerStub) Sync(context.Context, []string, string) error { return nil }

type cleanupHandlerStub struct{}

func (cleanupHandlerStub) CleanupLocks(context.Context) error       { return nil }
func (cleanupHandlerStub) CleanupInvitations(context.Context) error { return nil }
func (cleanupHandlerStub) CleanupOldData(context.Context) error     { return nil }

type changelogSynchronizerStub struct{}

func (changelogSynchronizerStub) Sync(context.Context) error { return nil }

func completeHandlers() Handlers {
	return Handlers{
		EnvironmentScraper:         scraperStub{},
		StoreExtensionSynchronizer: storeExtensionSynchronizerStub{},
		SitespeedScraper:           scraperStub{},
		Cleanup:                    cleanupHandlerStub{},
		ChangelogSynchronizer:      changelogSynchronizerStub{},
	}
}
