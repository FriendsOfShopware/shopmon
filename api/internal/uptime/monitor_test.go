package uptime

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
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
	assert.False(t, loop.trackMassDown(now))
	assert.False(t, loop.trackMassDown(now.Add(10*time.Second)))
	assert.True(t, loop.trackMassDown(now.Add(20*time.Second)), "third down within the window trips the breaker")

	// Downs outside the window do not count.
	loop2 := NewMonitorLoop(MonitorConfig{BreakerWindow: time.Minute, BreakerThreshold: 3}, nil, nil, nil, nil, nil)
	base := time.Now()
	loop2.trackMassDown(base)
	loop2.trackMassDown(base.Add(10 * time.Second))
	assert.False(t, loop2.trackMassDown(base.Add(90*time.Second)), "expired downs must be pruned")
}

// stubProber returns a fixed result for every probe.
type stubProber struct {
	mu     sync.Mutex
	result ProbeResult
	calls  int
}

func (s *stubProber) Probe(context.Context, ProbeConfig) ProbeResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
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

// dueRepo returns a fixed set of monitors once, then nothing.
type dueRepo struct {
	mu       sync.Mutex
	monitors []queries.ListDueUptimeMonitorsRow
}

func (d *dueRepo) ListDueUptimeMonitors(context.Context, queries.ListDueUptimeMonitorsParams) ([]queries.ListDueUptimeMonitorsRow, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := d.monitors
	d.monitors = nil
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

	monitor := queries.ListDueUptimeMonitorsRow{
		EnvironmentID:     7,
		IntervalSeconds:   60,
		FailureThreshold:  1,
		RecoveryThreshold: 1,
		Status:            StatusUp,
		EnvironmentUrl:    "https://shop.example.com",
		EnvironmentName:   "Production",
		OrganizationID:    "org-1",
	}
	due := &dueRepo{monitors: []queries.ListDueUptimeMonitorsRow{monitor}}
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
	cancel()
	loop.Stop()
}

func TestMonitorLoopSuppressesRecoveryAfterMassDown(t *testing.T) {
	loop := NewMonitorLoop(MonitorConfig{BreakerThreshold: 1}, nil, nil, nil, nil, nil)

	// A single down trips the threshold-1 breaker and marks the monitor suppressed.
	tripped := loop.trackMassDown(time.Now())
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

	monitor := Monitor{EnvironmentID: 7, IntervalSeconds: 60}
	loop.accumulate(monitor, ProbeResult{OK: true, Latency: 100 * time.Millisecond})
	loop.accumulate(monitor, ProbeResult{OK: true, Latency: 200 * time.Millisecond})
	loop.accumulate(monitor, ProbeResult{OK: false, Latency: 300 * time.Millisecond})

	hour := time.Now().UTC().Truncate(time.Hour)
	loop.mu.Lock()
	accs := loop.accumulators
	loop.accumulators = make(map[int32]*hourAcc)
	loop.mu.Unlock()

	rolls := &rollRepo{}
	loop.rollRepo = rolls
	loop.flushAccumulators(context.Background(), hour, accs)

	require.Len(t, rolls.rows, 1)
	row := rolls.rows[0]
	assert.EqualValues(t, 7, row.EnvironmentID)
	assert.EqualValues(t, 3, row.ProbesRun)
	assert.EqualValues(t, 2, row.ProbesOk)
	assert.EqualValues(t, 60, row.ProbesExpected, "3600/60 probes expected per hour")
	require.NotNil(t, row.LatencyAvgMs)
	assert.EqualValues(t, 200, *row.LatencyAvgMs)
	require.NotNil(t, row.LatencyP95Ms)
	assert.EqualValues(t, 300, *row.LatencyP95Ms)
}
