package jobs

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchedulerRegistersRecurringJobs(t *testing.T) {
	scheduler, err := NewScheduler(&scheduleRepositoryStub{}, &scheduleDispatcherStub{}, SchedulerConfig{})
	require.NoError(t, err)
	assert.Len(t, scheduler.cron.Entries(), 8)
}

func TestSchedulerFiltersRepeatedConnectionFailures(t *testing.T) {
	repository := &scheduleRepositoryStub{environments: []ScheduledEnvironment{
		{ID: 1},
		{ID: 2, ConnectionIssueCount: 3},
		{ID: 3, ConnectionIssueCount: 2},
	}}
	dispatcher := &scheduleDispatcherStub{}
	scheduler, err := NewScheduler(repository, dispatcher, SchedulerConfig{})
	require.NoError(t, err)

	scheduler.enqueueEnvironmentScrapes(context.Background())

	assert.Equal(t, []int32{1, 3}, dispatcher.environmentScrapes)
}

func TestSchedulerOnlyQueriesSitespeedTargetsWhenEnabled(t *testing.T) {
	repository := &scheduleRepositoryStub{sitespeed: []int32{4, 5}}
	dispatcher := &scheduleDispatcherStub{}
	disabled, err := NewScheduler(repository, dispatcher, SchedulerConfig{})
	require.NoError(t, err)

	disabled.enqueueSitespeedScrapes(context.Background())
	assert.Zero(t, repository.sitespeedCalls)
	assert.Empty(t, dispatcher.sitespeedScrapes)

	enabled, err := NewScheduler(repository, dispatcher, SchedulerConfig{SitespeedEnabled: true})
	require.NoError(t, err)
	enabled.enqueueSitespeedScrapes(context.Background())
	assert.Equal(t, 1, repository.sitespeedCalls)
	assert.Equal(t, []int32{4, 5}, dispatcher.sitespeedScrapes)
}

func TestSchedulerStopsFanoutWhenRepositoryFails(t *testing.T) {
	repository := &scheduleRepositoryStub{environmentErr: errors.New("database unavailable")}
	dispatcher := &scheduleDispatcherStub{}
	scheduler, err := NewScheduler(repository, dispatcher, SchedulerConfig{})
	require.NoError(t, err)

	scheduler.enqueueEnvironmentScrapes(context.Background())

	assert.Empty(t, dispatcher.environmentScrapes)
}

func TestNewSchedulerRequiresDependencies(t *testing.T) {
	_, err := NewScheduler(nil, &scheduleDispatcherStub{}, SchedulerConfig{})
	require.Error(t, err)

	_, err = NewScheduler(&scheduleRepositoryStub{}, nil, SchedulerConfig{})
	require.Error(t, err)
}

type scheduleRepositoryStub struct {
	environments   []ScheduledEnvironment
	sitespeed      []int32
	environmentErr error
	sitespeedErr   error
	sitespeedCalls int
}

func (r *scheduleRepositoryStub) ListEnvironmentScrapeTargets(context.Context) ([]ScheduledEnvironment, error) {
	return r.environments, r.environmentErr
}

func (r *scheduleRepositoryStub) ListSitespeedScrapeTargets(context.Context) ([]int32, error) {
	r.sitespeedCalls++
	return r.sitespeed, r.sitespeedErr
}

type scheduleDispatcherStub struct {
	environmentScrapes []int32
	sitespeedScrapes   []int32
}

func (d *scheduleDispatcherStub) EnqueueEnvironmentScrape(_ context.Context, environmentID int32) error {
	d.environmentScrapes = append(d.environmentScrapes, environmentID)
	return nil
}

func (d *scheduleDispatcherStub) EnqueueSitespeedScrape(_ context.Context, environmentID int32) error {
	d.sitespeedScrapes = append(d.sitespeedScrapes, environmentID)
	return nil
}

func (d *scheduleDispatcherStub) EnqueueLockCleanup(context.Context) error { return nil }

func (d *scheduleDispatcherStub) EnqueueInvitationCleanup(context.Context) error { return nil }

func (d *scheduleDispatcherStub) EnqueueOldDataCleanup(context.Context) error { return nil }

func (d *scheduleDispatcherStub) EnqueueShopwareChangelogSync(context.Context) error { return nil }

func (d *scheduleDispatcherStub) EnqueueComposerAdvisorySync(context.Context) error { return nil }
func (d *scheduleDispatcherStub) EnqueueSecurityPluginSync(context.Context) error   { return nil }
