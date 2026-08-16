package uptime

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/testutil/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPercentile(t *testing.T) {
	assert.EqualValues(t, -1, percentile(nil, 95))
	assert.EqualValues(t, 5, percentile([]int32{5}, 95))

	values := make([]int32, 100)
	for i := range values {
		values[i] = int32(i + 1) // 1..100
	}
	assert.EqualValues(t, 95, percentile(values, 95))
	assert.EqualValues(t, 50, percentile(values, 50))
}

func TestMonitorLoopMassDownBreaker(t *testing.T) {
	loop := NewMonitorLoop(MonitorConfig{
		BreakerWindow:    time.Minute,
		BreakerThreshold: 3,
	}, nil, nil, nil, nil, nil)

	now := time.Now()
	assert.False(t, loop.trackMassDown(1, now))
	assert.False(t, loop.trackMassDown(2, now.Add(10*time.Second)))
	assert.True(t, loop.trackMassDown(3, now.Add(20*time.Second)), "third distinct down within the window trips the breaker")

	// Recovered monitors leave the window, un-tripping the breaker even
	// while new failures arrive.
	loop.clearMassDown(2)
	loop.clearMassDown(3)
	assert.False(t, loop.trackMassDown(4, now.Add(30*time.Second)), "recoveries must remove monitors from the window")

	// Downs outside the window do not count.
	loop2 := NewMonitorLoop(MonitorConfig{BreakerWindow: time.Minute, BreakerThreshold: 3}, nil, nil, nil, nil, nil)
	base := time.Now()
	loop2.trackMassDown(1, base)
	loop2.trackMassDown(2, base.Add(10*time.Second))
	assert.False(t, loop2.trackMassDown(3, base.Add(90*time.Second)), "expired downs must be pruned")

	// One flapping monitor never accumulates towards the threshold.
	loop3 := NewMonitorLoop(MonitorConfig{BreakerWindow: time.Minute, BreakerThreshold: 3}, nil, nil, nil, nil, nil)
	loop3.trackMassDown(7, base)
	loop3.trackMassDown(7, base.Add(10*time.Second))
	assert.False(t, loop3.trackMassDown(7, base.Add(20*time.Second)), "the same monitor must count once")
}

// stubProber returns a fixed result for every probe.
type stubProber struct {
	mu     sync.Mutex
	result ProbeResult
	calls  int
	// delay, when set, is how long each probe takes.
	delay time.Duration
}

func (s *stubProber) Probe(context.Context, ProbeConfig) ProbeResult {
	s.mu.Lock()
	s.calls++
	delay := s.delay
	s.mu.Unlock()
	if delay > 0 {
		time.Sleep(delay)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.result
}

// stubAlerter records alert invocations.
type stubAlerter struct {
	mu         sync.Mutex
	downs      []int32
	recoveries []int32
}

func (s *stubAlerter) AlertDown(_ context.Context, m Monitor, _ int32, _ ProbeResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.downs = append(s.downs, m.EnvironmentID)
}

func (s *stubAlerter) AlertRecovered(_ context.Context, m Monitor, _ int32, _ time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recoveries = append(s.recoveries, m.EnvironmentID)
}

// claimRepo returns a fixed set of monitors once, then nothing, mimicking
// the atomic claim semantics.
type claimRepo struct {
	mu       sync.Mutex
	monitors []queries.ClaimDueUptimeMonitorsRow
	calls    int
}

func (c *claimRepo) ClaimDueUptimeMonitors(context.Context, queries.ClaimDueUptimeMonitorsParams) ([]queries.ClaimDueUptimeMonitorsRow, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls++
	out := c.monitors
	c.monitors = nil
	return out, nil
}

// rollRepo records flushed hourly rollups.
type rollRepo struct {
	mu   sync.Mutex
	rows []queries.UpsertUptimeHourlyRollupParams
}

func (r *rollRepo) UpsertUptimeHourlyRollup(_ context.Context, arg queries.UpsertUptimeHourlyRollupParams) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows = append(r.rows, arg)
	return nil
}

func TestMonitorLoopProbesAndAlerts(t *testing.T) {
	prober := &stubProber{result: ProbeResult{OK: false, StatusCode: 503, Err: "status 503"}}
	alerter := &stubAlerter{}
	repo := &fakeRecorderRepo{}
	recorder := NewRecorder(repo)

	monitor := queries.ClaimDueUptimeMonitorsRow{
		EnvironmentID:     7,
		IntervalSeconds:   60,
		FailureThreshold:  1,
		RecoveryThreshold: 1,
		Status:            StatusUp,
		EnvironmentUrl:    "https://shop.example.com",
		EnvironmentName:   "Production",
		OrganizationID:    "org-1",
	}
	due := &claimRepo{monitors: []queries.ClaimDueUptimeMonitorsRow{monitor}}
	rolls := &rollRepo{}

	loop := NewMonitorLoop(MonitorConfig{
		TickInterval:     5 * time.Millisecond,
		MaxConcurrent:    4,
		BreakerThreshold: 10,
	}, due, rolls, prober, recorder, alerter)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	loop.Start(ctx)

	require.Eventually(t, func() bool {
		alerter.mu.Lock()
		defer alerter.mu.Unlock()
		return len(alerter.downs) == 1
	}, 2*time.Second, 5*time.Millisecond)

	assert.EqualValues(t, 7, alerter.downs[0])
	loop.Stop()
}

// TestMonitorLoopStopWaitsForInFlightProbes verifies that Stop lets a slow
// probe finish recording instead of cancelling it mid-flight.
func TestMonitorLoopStopWaitsForInFlightProbes(t *testing.T) {
	prober := &stubProber{result: ProbeResult{OK: true, StatusCode: 200}, delay: 150 * time.Millisecond}
	repo := &fakeRecorderRepo{}
	recorder := NewRecorder(repo)

	monitor := queries.ClaimDueUptimeMonitorsRow{
		EnvironmentID:     7,
		IntervalSeconds:   60,
		FailureThreshold:  3,
		RecoveryThreshold: 2,
		Status:            StatusUp,
		EnvironmentUrl:    "https://shop.example.com",
	}
	due := &claimRepo{monitors: []queries.ClaimDueUptimeMonitorsRow{monitor}}

	loop := NewMonitorLoop(MonitorConfig{
		TickInterval: 5 * time.Millisecond,
	}, due, &rollRepo{}, prober, recorder, &stubAlerter{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	loop.Start(ctx)

	// Wait until a probe is in flight, then stop immediately.
	require.Eventually(t, func() bool {
		prober.mu.Lock()
		defer prober.mu.Unlock()
		return prober.calls > 0
	}, 2*time.Second, 5*time.Millisecond)
	loop.Stop()

	assert.Equal(t, 1, repo.stateCalls, "the in-flight probe must have recorded its state")
	assert.Equal(t, StatusUp, repo.state.Status)
}

func TestMonitorLoopSuppressesRecoveryAfterMassDown(t *testing.T) {
	loop := NewMonitorLoop(MonitorConfig{BreakerThreshold: 1}, nil, nil, nil, nil, nil)

	// A single down trips the threshold-1 breaker and marks the monitor suppressed.
	tripped := loop.trackMassDown(7, time.Now())
	require.True(t, tripped)
	loop.mu.Lock()
	loop.suppressed[7] = struct{}{}
	loop.mu.Unlock()

	alerter := &stubAlerter{}
	loop.alerter = alerter

	// Simulate the recovery branch of handleTransition directly.
	mon := Monitor{EnvironmentID: 7}
	loop.handleTransition(context.Background(), mon, RecordOutcome{Transition: TransitionUp}, ProbeResult{OK: true})

	alerter.mu.Lock()
	defer alerter.mu.Unlock()
	assert.Empty(t, alerter.recoveries, "recovery must stay silent when the down alert was suppressed")
	loop.mu.Lock()
	_, stillTracked := loop.suppressed[7]
	loop.mu.Unlock()
	assert.False(t, stillTracked, "suppression marker is cleared after recovery")
	loop.mu.Lock()
	_, inBreakerWindow := loop.recentDowns[7]
	loop.mu.Unlock()
	assert.False(t, inBreakerWindow, "recovery must leave the mass-down window")
}

func TestMonitorLoopGatesShortRecoveryNotifications(t *testing.T) {
	loop := NewMonitorLoop(MonitorConfig{BreakerThreshold: 100}, nil, nil, nil, nil, nil)
	alerter := &stubAlerter{}
	loop.alerter = alerter
	mon := Monitor{EnvironmentID: 7}

	// A short blip recovers silently (no customer notification).
	loop.handleTransition(context.Background(), mon, RecordOutcome{Transition: TransitionUp, DownFor: 30 * time.Second}, ProbeResult{OK: true})

	// A sustained outage recovering does notify.
	loop.handleTransition(context.Background(), mon, RecordOutcome{Transition: TransitionUp, DownFor: 10 * time.Minute}, ProbeResult{OK: true})

	alerter.mu.Lock()
	defer alerter.mu.Unlock()
	assert.Equal(t, []int32{7}, alerter.recoveries, "only the long incident should notify recovery")
}

func TestMonitorLoopAccumulatesAndFlushes(t *testing.T) {
	loop := NewMonitorLoop(MonitorConfig{}, nil, nil, nil, nil, nil)

	hour := time.Now().UTC().Truncate(time.Hour)
	monitor := Monitor{EnvironmentID: 7, IntervalSeconds: 60}
	loop.accumulate(monitor, ProbeResult{OK: true, Latency: 100 * time.Millisecond}, hour.Add(time.Minute))
	loop.accumulate(monitor, ProbeResult{OK: true, Latency: 200 * time.Millisecond}, hour.Add(2*time.Minute))
	loop.accumulate(monitor, ProbeResult{OK: false, Latency: 300 * time.Millisecond}, hour.Add(3*time.Minute))

	rolls := &rollRepo{}
	loop.rollRepo = rolls
	loop.flushAccumulators(context.Background(), loop.takeAccumulators())

	require.Len(t, rolls.rows, 1)
	row := rolls.rows[0]
	assert.EqualValues(t, 7, row.EnvironmentID)
	assert.EqualValues(t, hour, row.Hour.Time, "the rollup must land in the hour the probes started")
	assert.EqualValues(t, 3, row.ProbesRun)
	assert.EqualValues(t, 2, row.ProbesOk)
	assert.EqualValues(t, 60, row.ProbesExpected, "3600/60 probes expected per hour")
	require.NotNil(t, row.LatencyAvgMs)
	assert.EqualValues(t, 200, *row.LatencyAvgMs)
	require.NotNil(t, row.LatencyP95Ms)
	assert.EqualValues(t, 300, *row.LatencyP95Ms)
}

// TestMonitorLoopAccumulateBucketsByProbeStart verifies that a probe started
// in an earlier hour flushes that hour instead of silently merging into the
// current one.
func TestMonitorLoopAccumulateBucketsByProbeStart(t *testing.T) {
	loop := NewMonitorLoop(MonitorConfig{}, nil, nil, nil, nil, nil)
	rolls := &rollRepo{}
	loop.rollRepo = rolls

	currentHour := time.Now().UTC().Truncate(time.Hour)
	previousHour := currentHour.Add(-time.Hour)
	monitor := Monitor{EnvironmentID: 7, IntervalSeconds: 60}

	// A probe for the previous hour arrives first (e.g. rolled over mid-probe).
	loop.accumulate(monitor, ProbeResult{OK: true, Latency: 100 * time.Millisecond}, previousHour)
	// Then a regular probe for the current hour.
	loop.accumulate(monitor, ProbeResult{OK: true, Latency: 200 * time.Millisecond}, currentHour)

	require.Len(t, rolls.rows, 1, "the previous hour must have been flushed when the hour changed")
	assert.Equal(t, previousHour, rolls.rows[0].Hour.Time)

	loop.flushAccumulators(context.Background(), loop.takeAccumulators())
	require.Len(t, rolls.rows, 2)
	assert.Equal(t, currentHour, rolls.rows[1].Hour.Time)
}

// TestClaimDueUptimeMonitorsIsAtomic is the regression test for the
// double-probe race: claiming the same due monitor a second time must return
// nothing because the first claim already pushed next_check_at forward.
func TestClaimDueUptimeMonitorsIsAtomic(t *testing.T) {
	pool := testdb.Setup(t)
	q := queries.New(pool)
	service := NewService(pool, q)
	userID, envID := seedUptimeFixture(t, pool)

	require.NoError(t, service.UpdateSettings(t.Context(), userID, envID, Settings{
		Enabled: true, IntervalSeconds: 60, FailureThreshold: 3, RecoveryThreshold: 2,
	}))
	// UpdateSettings arms next_check_at = NOW(); it is due immediately.
	// Simulate a slow first claim that has not recorded its probe yet.
	params := queries.ClaimDueUptimeMonitorsParams{
		Now:        pgTimestamp(time.Now().UTC()),
		ShardTotal: 1,
		ShardIndex: 0,
	}

	first, err := q.ClaimDueUptimeMonitors(t.Context(), params)
	require.NoError(t, err)
	require.Len(t, first, 1, "due monitor must be claimed")

	second, err := q.ClaimDueUptimeMonitors(t.Context(), params)
	require.NoError(t, err)
	assert.Empty(t, second, "an already-claimed monitor must not be handed out again")
}
