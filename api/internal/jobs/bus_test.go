package jobs

import (
	"context"
	"testing"
	"time"

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
