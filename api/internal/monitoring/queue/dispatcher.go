package queue

import (
	"context"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/jobs"
	"github.com/friendsofshopware/shopmon/api/internal/monitoring"
	goqueue "github.com/shyim/go-queue"
)

type Dispatcher struct {
	bus *goqueue.Bus
}

var _ monitoring.JobDispatcher = (*Dispatcher)(nil)

func NewDispatcher(bus *goqueue.Bus) *Dispatcher {
	return &Dispatcher{bus: bus}
}

func (d *Dispatcher) DispatchEnvironmentScrape(ctx context.Context, environmentID int32) error {
	if err := goqueue.Dispatch(ctx, d.bus, jobs.EnvironmentScrape{EnvironmentID: environmentID}); err != nil {
		return fmt.Errorf("dispatch environment scrape: %w", err)
	}
	return nil
}

func (d *Dispatcher) DispatchSitespeedScrape(ctx context.Context, environmentID int32) error {
	if err := goqueue.Dispatch(ctx, d.bus, jobs.SitespeedScrape{EnvironmentID: environmentID}); err != nil {
		return fmt.Errorf("dispatch sitespeed scrape: %w", err)
	}
	return nil
}
