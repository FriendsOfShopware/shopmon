package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"slices"
	"sync"
	"time"

	cron "github.com/robfig/cron/v3"
)

const (
	environmentScrapeSchedule = "0 * * * *"
	// Planning tick only. Each environment is then delayed across
	// sitespeedSpreadWindow so the sitespeed service is not hit all at once.
	sitespeedScrapeSchedule   = "0 3 * * *"
	lockCleanupSchedule       = "0 4 * * *"
	invitationCleanupSchedule = "0 5 * * *"
	oldDataCleanupSchedule    = "30 4 * * *"
	changelogSyncSchedule     = "15 * * * *"
	// Offset from the hour so we do not pile onto Packagist's :00 peaks.
	advisorySyncSchedule = "17 * * * *"
	// Offset from the advisory sync so the backport map is refreshed before the
	// next advisory pass rather than racing it.
	securityPluginSyncSchedule = "37 * * * *"

	// sitespeedSpreadWindow is how far the daily sitespeed fan-out is stretched.
	// It stays inside the 24h cron interval so one day's jobs are due before
	// the next 03:00 enqueue.
	sitespeedSpreadWindow = 23 * time.Hour
)

type ScheduledEnvironment struct {
	ID                   int32
	ConnectionIssueCount int32
}

type ScheduleRepository interface {
	ListEnvironmentScrapeTargets(ctx context.Context) ([]ScheduledEnvironment, error)
	ListSitespeedScrapeTargets(ctx context.Context) ([]int32, error)
}

type ScheduleDispatcher interface {
	EnqueueEnvironmentScrape(ctx context.Context, environmentID int32) error
	EnqueueSitespeedScrape(ctx context.Context, environmentID int32, delay time.Duration) error
	EnqueueLockCleanup(ctx context.Context) error
	EnqueueInvitationCleanup(ctx context.Context) error
	EnqueueOldDataCleanup(ctx context.Context) error
	EnqueueShopwareChangelogSync(ctx context.Context) error
	EnqueueComposerAdvisorySync(ctx context.Context) error
	EnqueueSecurityPluginSync(ctx context.Context) error
}

type SchedulerConfig struct {
	SitespeedEnabled bool
}

type Scheduler struct {
	cron       *cron.Cron
	repository ScheduleRepository
	dispatcher ScheduleDispatcher
	config     SchedulerConfig
	randMu     sync.Mutex
	rand       *rand.Rand
}

func NewScheduler(repository ScheduleRepository, dispatcher ScheduleDispatcher, config SchedulerConfig) (*Scheduler, error) {
	if repository == nil {
		return nil, errors.New("create job scheduler: repository is required")
	}
	if dispatcher == nil {
		return nil, errors.New("create job scheduler: dispatcher is required")
	}

	scheduler := &Scheduler{
		cron:       cron.New(),
		repository: repository,
		dispatcher: dispatcher,
		config:     config,
		rand:       rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
	}
	if err := scheduler.register(); err != nil {
		return nil, err
	}
	return scheduler, nil
}

func (s *Scheduler) register() error {
	registrations := []struct {
		name     string
		schedule string
		run      func()
	}{
		{
			name:     "environment scrape",
			schedule: environmentScrapeSchedule,
			run: func() {
				s.enqueueEnvironmentScrapes(context.Background())
			},
		},
		{
			name:     "sitespeed scrape",
			schedule: sitespeedScrapeSchedule,
			run: func() {
				s.enqueueSitespeedScrapes(context.Background())
			},
		},
		{
			name:     "lock cleanup",
			schedule: lockCleanupSchedule,
			run: func() {
				ctx := context.Background()
				s.logDispatchError(ctx, "lock cleanup", s.dispatcher.EnqueueLockCleanup(ctx))
			},
		},
		{
			name:     "invitation cleanup",
			schedule: invitationCleanupSchedule,
			run: func() {
				ctx := context.Background()
				s.logDispatchError(ctx, "invitation cleanup", s.dispatcher.EnqueueInvitationCleanup(ctx))
			},
		},
		{
			name:     "old data cleanup",
			schedule: oldDataCleanupSchedule,
			run: func() {
				ctx := context.Background()
				s.logDispatchError(ctx, "old data cleanup", s.dispatcher.EnqueueOldDataCleanup(ctx))
			},
		},
		{
			name:     "shopware changelog sync",
			schedule: changelogSyncSchedule,
			run: func() {
				ctx := context.Background()
				s.logDispatchError(ctx, "shopware changelog sync", s.dispatcher.EnqueueShopwareChangelogSync(ctx))
			},
		},
		{
			name:     "security plugin sync",
			schedule: securityPluginSyncSchedule,
			run: func() {
				ctx := context.Background()
				s.logDispatchError(ctx, "security plugin sync", s.dispatcher.EnqueueSecurityPluginSync(ctx))
			},
		},
		{
			name:     "composer advisory sync",
			schedule: advisorySyncSchedule,
			run: func() {
				ctx := context.Background()
				s.logDispatchError(ctx, "composer advisory sync", s.dispatcher.EnqueueComposerAdvisorySync(ctx))
			},
		},
	}

	for _, registration := range registrations {
		if _, err := s.cron.AddFunc(registration.schedule, registration.run); err != nil {
			return fmt.Errorf("register %s schedule: %w", registration.name, err)
		}
	}
	return nil
}

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Stop() context.Context {
	return s.cron.Stop()
}

func (s *Scheduler) enqueueEnvironmentScrapes(ctx context.Context) {
	targets, err := s.repository.ListEnvironmentScrapeTargets(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list environments for scrape", "error", err)
		return
	}
	dispatched := 0
	skipped := 0
	failed := 0
	for _, target := range targets {
		// Back off instead of hammering an environment that has repeatedly
		// failed to connect.
		if target.ConnectionIssueCount >= 3 {
			skipped++
			continue
		}
		if err := s.dispatcher.EnqueueEnvironmentScrape(ctx, target.ID); err != nil {
			failed++
			slog.ErrorContext(ctx, "failed to dispatch environment scrape", "environmentId", target.ID, "error", err)
			continue
		}
		dispatched++
	}
	slog.InfoContext(ctx, "enqueued environment scrapes",
		"targets", len(targets),
		"dispatched", dispatched,
		"skippedConnectionIssues", skipped,
		"failed", failed,
	)
}

func (s *Scheduler) enqueueSitespeedScrapes(ctx context.Context) {
	if !s.config.SitespeedEnabled {
		return
	}
	targets, err := s.repository.ListSitespeedScrapeTargets(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list environments for sitespeed scrape", "error", err)
		return
	}

	s.randMu.Lock()
	delays := sitespeedScrapeDelays(targets, sitespeedSpreadWindow, s.rand)
	s.randMu.Unlock()

	dispatched := 0
	failed := 0
	for _, environmentID := range targets {
		delay := delays[environmentID]
		if err := s.dispatcher.EnqueueSitespeedScrape(ctx, environmentID, delay); err != nil {
			failed++
			slog.ErrorContext(ctx, "failed to dispatch sitespeed scrape", "environmentId", environmentID, "delay", delay.String(), "error", err)
			continue
		}
		dispatched++
	}
	slog.InfoContext(ctx, "enqueued sitespeed scrapes",
		"targets", len(targets),
		"dispatched", dispatched,
		"failed", failed,
		"spreadWindow", sitespeedSpreadWindow.String(),
	)
}

// sitespeedScrapeDelays assigns each environment a delay in [0, window).
// Slots are evenly spaced, then jittered inside the first half of the slot,
// after a shuffle. Runs stay apart, and a shop is not measured at the same
// time every day.
func sitespeedScrapeDelays(environmentIDs []int32, window time.Duration, rng *rand.Rand) map[int32]time.Duration {
	delays := make(map[int32]time.Duration, len(environmentIDs))
	n := len(environmentIDs)
	if n == 0 || window <= 0 || rng == nil {
		return delays
	}

	order := slices.Clone(environmentIDs)
	slices.Sort(order)
	rng.Shuffle(n, func(i, j int) {
		order[i], order[j] = order[j], order[i]
	})

	if n == 1 {
		delays[order[0]] = time.Duration(rng.Int64N(int64(window)))
		return delays
	}

	spacing := int64(window) / int64(n)
	if spacing < 2 {
		for _, id := range order {
			delays[id] = time.Duration(rng.Int64N(int64(window)))
		}
		return delays
	}

	jitterMax := spacing / 2
	for i, id := range order {
		jitter := rng.Int64N(jitterMax)
		delays[id] = time.Duration(int64(i)*spacing + jitter)
	}
	return delays
}

func (s *Scheduler) logDispatchError(ctx context.Context, job string, err error) {
	if err != nil {
		slog.ErrorContext(ctx, "failed to dispatch scheduled job", "job", job, "error", err)
	}
}
