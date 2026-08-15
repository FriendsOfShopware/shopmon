package jobs

import (
	"context"
	"fmt"
	"time"

	goqueue "github.com/shyim/go-queue"
)

// Dispatcher adapts capability and scheduler dispatch ports to go-queue.
type Dispatcher struct {
	bus *goqueue.Bus
}

var _ ScheduleDispatcher = (*Dispatcher)(nil)

func NewDispatcher(bus *goqueue.Bus) *Dispatcher {
	return &Dispatcher{bus: bus}
}

func (d *Dispatcher) DispatchCatalogSync(ctx context.Context, names []string, shopwareVersion string) error {
	if err := Dispatch(ctx, d.bus, StoreExtensionSync{Names: names, ShopwareVersion: shopwareVersion}); err != nil {
		return fmt.Errorf("dispatch store extension sync: %w", err)
	}
	return nil
}

func (d *Dispatcher) DispatchSitespeedScrape(ctx context.Context, environmentID int32, delay time.Duration) error {
	if err := Dispatch(ctx, d.bus, SitespeedScrape{EnvironmentID: environmentID}, goqueue.WithDelay(delay)); err != nil {
		return fmt.Errorf("dispatch delayed sitespeed scrape: %w", err)
	}
	return nil
}

func (d *Dispatcher) EnqueueEnvironmentScrape(ctx context.Context, environmentID int32) error {
	if err := Dispatch(ctx, d.bus, EnvironmentScrape{EnvironmentID: environmentID}); err != nil {
		return fmt.Errorf("dispatch environment scrape: %w", err)
	}
	return nil
}

func (d *Dispatcher) EnqueueSitespeedScrape(ctx context.Context, environmentID int32) error {
	if err := Dispatch(ctx, d.bus, SitespeedScrape{EnvironmentID: environmentID}); err != nil {
		return fmt.Errorf("dispatch sitespeed scrape: %w", err)
	}
	return nil
}

func (d *Dispatcher) EnqueueLockCleanup(ctx context.Context) error {
	if err := Dispatch(ctx, d.bus, LockCleanup{}); err != nil {
		return fmt.Errorf("dispatch lock cleanup: %w", err)
	}
	return nil
}

func (d *Dispatcher) EnqueueInvitationCleanup(ctx context.Context) error {
	if err := Dispatch(ctx, d.bus, InvitationCleanup{}); err != nil {
		return fmt.Errorf("dispatch invitation cleanup: %w", err)
	}
	return nil
}

func (d *Dispatcher) EnqueueOldDataCleanup(ctx context.Context) error {
	if err := Dispatch(ctx, d.bus, OldDataCleanup{}); err != nil {
		return fmt.Errorf("dispatch old data cleanup: %w", err)
	}
	return nil
}

func (d *Dispatcher) EnqueueShopwareChangelogSync(ctx context.Context) error {
	if err := Dispatch(ctx, d.bus, ShopwareChangelogSync{}); err != nil {
		return fmt.Errorf("dispatch shopware changelog sync: %w", err)
	}
	return nil
}

func (d *Dispatcher) EnqueueComposerAdvisorySync(ctx context.Context) error {
	if err := Dispatch(ctx, d.bus, ComposerAdvisorySync{}); err != nil {
		return fmt.Errorf("dispatch composer advisory sync: %w", err)
	}
	return nil
}

func (d *Dispatcher) EnqueueSecurityPluginSync(ctx context.Context) error {
	if err := Dispatch(ctx, d.bus, SecurityPluginSync{}); err != nil {
		return fmt.Errorf("dispatch security plugin sync: %w", err)
	}
	return nil
}

func (d *Dispatcher) EnqueueUptimeDailyRollup(ctx context.Context) error {
	if err := Dispatch(ctx, d.bus, UptimeDailyRollup{}); err != nil {
		return fmt.Errorf("dispatch uptime daily rollup: %w", err)
	}
	return nil
}

func (d *Dispatcher) DispatchUptimeResume(ctx context.Context, environmentID int32, delay time.Duration) error {
	if err := Dispatch(ctx, d.bus, UptimeResume{EnvironmentID: environmentID}, goqueue.WithDelay(delay)); err != nil {
		return fmt.Errorf("dispatch delayed uptime resume: %w", err)
	}
	return nil
}
