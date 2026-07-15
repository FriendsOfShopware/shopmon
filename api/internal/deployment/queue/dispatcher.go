package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/deployment"
	"github.com/friendsofshopware/shopmon/api/internal/jobs"
	goqueue "github.com/shyim/go-queue"
)

type Dispatcher struct {
	bus *goqueue.Bus
}

var _ deployment.PostDeploymentDispatcher = (*Dispatcher)(nil)

func NewDispatcher(bus *goqueue.Bus) *Dispatcher {
	return &Dispatcher{bus: bus}
}

func (d *Dispatcher) DispatchPostDeploymentScrape(ctx context.Context, environmentID int32, delay time.Duration) error {
	if err := jobs.Dispatch(ctx, d.bus, jobs.EnvironmentScrape{EnvironmentID: environmentID}, goqueue.WithDelay(delay)); err != nil {
		return fmt.Errorf("dispatch post-deployment scrape: %w", err)
	}
	return nil
}
