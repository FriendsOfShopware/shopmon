package jobs

import (
	"context"
	"errors"
	"math/rand/v2"
	"slices"
	"testing"
	"time"

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
	enabled.rand = rand.New(rand.NewPCG(1, 2))
	enabled.enqueueSitespeedScrapes(context.Background())
	assert.Equal(t, 1, repository.sitespeedCalls)
	assert.Equal(t, []int32{4, 5}, dispatcher.sitespeedScrapes)
	assertSitespeedDelaysSpread(t, dispatcher.sitespeedDelays, sitespeedSpreadWindow)
}

func TestSchedulerSpreadsSitespeedScrapesAcrossTheDay(t *testing.T) {
	ids := []int32{4, 5, 6, 7}
	repository := &scheduleRepositoryStub{sitespeed: ids}
	dispatcher := &scheduleDispatcherStub{}
	scheduler, err := NewScheduler(repository, dispatcher, SchedulerConfig{SitespeedEnabled: true})
	require.NoError(t, err)
	scheduler.rand = rand.New(rand.NewPCG(1, 2))

	scheduler.enqueueSitespeedScrapes(context.Background())

	expected := sitespeedScrapeDelays(ids, sitespeedSpreadWindow, rand.New(rand.NewPCG(1, 2)))
	require.Len(t, dispatcher.sitespeedScrapes, len(ids))
	for i, environmentID := range dispatcher.sitespeedScrapes {
		assert.Equal(t, expected[environmentID], dispatcher.sitespeedDelays[i])
	}
	assertSitespeedDelaysSpread(t, dispatcher.sitespeedDelays, sitespeedSpreadWindow)
}

func TestSitespeedScrapeDelaysStayInsideTheWindow(t *testing.T) {
	ids := make([]int32, 100)
	for i := range ids {
		ids[i] = int32(i + 1)
	}
	window := sitespeedSpreadWindow

	delays := sitespeedScrapeDelays(ids, window, rand.New(rand.NewPCG(7, 11)))

	require.Len(t, delays, len(ids))
	sorted := make([]time.Duration, 0, len(ids))
	for _, id := range ids {
		delay, ok := delays[id]
		require.True(t, ok)
		assert.GreaterOrEqual(t, delay, time.Duration(0))
		assert.Less(t, delay, window)
		sorted = append(sorted, delay)
	}
	slices.Sort(sorted)
	minGap := window / time.Duration(len(ids)) / 2
	for i := 1; i < len(sorted); i++ {
		assert.GreaterOrEqual(t, sorted[i]-sorted[i-1], minGap)
	}
}

func TestSitespeedScrapeDelaysChangeBetweenDays(t *testing.T) {
	ids := []int32{10, 20, 30, 40, 50}
	first := sitespeedScrapeDelays(ids, sitespeedSpreadWindow, rand.New(rand.NewPCG(1, 2)))
	second := sitespeedScrapeDelays(ids, sitespeedSpreadWindow, rand.New(rand.NewPCG(3, 4)))

	same := 0
	for _, id := range ids {
		if first[id] == second[id] {
			same++
		}
	}
	assert.Less(t, same, len(ids))
}

func TestSitespeedScrapeDelaysIgnoreInputOrder(t *testing.T) {
	window := sitespeedSpreadWindow
	forward := sitespeedScrapeDelays([]int32{3, 1, 2}, window, rand.New(rand.NewPCG(9, 9)))
	backward := sitespeedScrapeDelays([]int32{2, 1, 3}, window, rand.New(rand.NewPCG(9, 9)))
	assert.Equal(t, forward, backward)
}

func TestSitespeedScrapeDelayForASingleEnvironmentUsesTheWholeWindow(t *testing.T) {
	seen := map[time.Duration]struct{}{}
	for seed := uint64(1); seed <= 20; seed++ {
		delays := sitespeedScrapeDelays([]int32{8}, sitespeedSpreadWindow, rand.New(rand.NewPCG(seed, seed)))
		delay := delays[8]
		assert.GreaterOrEqual(t, delay, time.Duration(0))
		assert.Less(t, delay, sitespeedSpreadWindow)
		seen[delay] = struct{}{}
	}
	assert.Greater(t, len(seen), 1)
}

func TestSitespeedScrapeDelaysEmpty(t *testing.T) {
	assert.Empty(t, sitespeedScrapeDelays(nil, sitespeedSpreadWindow, rand.New(rand.NewPCG(1, 2))))
	assert.Empty(t, sitespeedScrapeDelays([]int32{1}, 0, rand.New(rand.NewPCG(1, 2))))
	assert.Empty(t, sitespeedScrapeDelays([]int32{1}, sitespeedSpreadWindow, nil))
}

func assertSitespeedDelaysSpread(t *testing.T, delays []time.Duration, window time.Duration) {
	t.Helper()
	require.NotEmpty(t, delays)
	for _, delay := range delays {
		assert.GreaterOrEqual(t, delay, time.Duration(0))
		assert.Less(t, delay, window)
	}
	if len(delays) < 2 {
		return
	}
	sorted := slices.Clone(delays)
	slices.Sort(sorted)
	assert.Positive(t, sorted[len(sorted)-1]-sorted[0])
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
	sitespeedDelays    []time.Duration
}

func (d *scheduleDispatcherStub) EnqueueEnvironmentScrape(_ context.Context, environmentID int32) error {
	d.environmentScrapes = append(d.environmentScrapes, environmentID)
	return nil
}

func (d *scheduleDispatcherStub) EnqueueSitespeedScrape(_ context.Context, environmentID int32, delay time.Duration) error {
	d.sitespeedScrapes = append(d.sitespeedScrapes, environmentID)
	d.sitespeedDelays = append(d.sitespeedDelays, delay)
	return nil
}

func (d *scheduleDispatcherStub) EnqueueLockCleanup(context.Context) error { return nil }

func (d *scheduleDispatcherStub) EnqueueInvitationCleanup(context.Context) error { return nil }

func (d *scheduleDispatcherStub) EnqueueOldDataCleanup(context.Context) error { return nil }

func (d *scheduleDispatcherStub) EnqueueShopwareChangelogSync(context.Context) error { return nil }

func (d *scheduleDispatcherStub) EnqueueComposerAdvisorySync(context.Context) error { return nil }
func (d *scheduleDispatcherStub) EnqueueSecurityPluginSync(context.Context) error   { return nil }
