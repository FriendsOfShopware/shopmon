package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	cron "github.com/robfig/cron/v3"
)

const (
	environmentScrapeSchedule = "0 * * * *"
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
	// Aggregate the previous UTC day's uptime probes into daily availability
	// rollups and enforce retention. Runs shortly after midnight UTC.
	uptimeDailyRollupSchedule = "10 0 * * *"
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
	EnqueueSitespeedScrape(ctx context.Context, environmentID int32) error
	EnqueueLockCleanup(ctx context.Context) error
	EnqueueInvitationCleanup(ctx context.Context) error
	EnqueueOldDataCleanup(ctx context.Context) error
	EnqueueShopwareChangelogSync(ctx context.Context) error
	EnqueueComposerAdvisorySync(ctx context.Context) error
	EnqueueSecurityPluginSync(ctx context.Context) error
	EnqueueUptimeDailyRollup(ctx context.Context) error
}

type SchedulerConfig struct {
	SitespeedEnabled bool
}

type Scheduler struct {
	cron       *cron.Cron
	repository ScheduleRepository
	dispatcher ScheduleDispatcher
	config     SchedulerConfig
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
		{
			name:     "uptime daily rollup",
			schedule: uptimeDailyRollupSchedule,
			run: func() {
				ctx := context.Background()
				s.logDispatchError(ctx, "uptime daily rollup", s.dispatcher.EnqueueUptimeDailyRollup(ctx))
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
	for _, target := range targets {
		// Back off instead of hammering an environment that has repeatedly
		// failed to connect.
		if target.ConnectionIssueCount >= 3 {
			continue
		}
		if err := s.dispatcher.EnqueueEnvironmentScrape(ctx, target.ID); err != nil {
			slog.ErrorContext(ctx, "failed to dispatch environment scrape", "environmentId", target.ID, "error", err)
		}
	}
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
	for _, environmentID := range targets {
		if err := s.dispatcher.EnqueueSitespeedScrape(ctx, environmentID); err != nil {
			slog.ErrorContext(ctx, "failed to dispatch sitespeed scrape", "environmentId", environmentID, "error", err)
		}
	}
}

func (s *Scheduler) logDispatchError(ctx context.Context, job string, err error) {
	if err != nil {
		slog.ErrorContext(ctx, "failed to dispatch scheduled job", "job", job, "error", err)
	}
}
