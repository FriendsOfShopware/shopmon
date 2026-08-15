package uptime

import (
	"context"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/metrics"
)

// minRecoveryNotifyDuration gates recovery notifications: incidents shorter
// than this flip back silently to avoid up/down notification ping-pong. The
// state transition and incident resolution happen regardless.
const minRecoveryNotifyDuration = 2 * time.Minute

// MonitorConfig tunes the probe loop.
type MonitorConfig struct {
	// TickInterval is how often the loop looks for due monitors. Keep this
	// well below the smallest probe interval so probes stay on cadence.
	TickInterval time.Duration
	// ProbeTimeout bounds a single probe.
	ProbeTimeout time.Duration
	// MaxConcurrent bounds in-flight probes.
	MaxConcurrent int
	// ShardTotal/ShardIndex split monitors across worker replicas: a monitor
	// is owned by the worker where environment_id % ShardTotal == ShardIndex.
	ShardTotal int32
	ShardIndex int32
	// BreakerWindow/BreakerThreshold implement the mass-down circuit breaker:
	// if at least BreakerThreshold monitors go down within BreakerWindow, the
	// cause is almost certainly the probe infrastructure, not the shops, and
	// customer notifications are suppressed.
	BreakerWindow    time.Duration
	BreakerThreshold int
	// ListTimeout bounds the due-monitors query.
	ListTimeout time.Duration
}

func (c *MonitorConfig) withDefaults() MonitorConfig {
	out := *c
	if out.TickInterval <= 0 {
		out.TickInterval = 10 * time.Second
	}
	if out.ProbeTimeout <= 0 {
		out.ProbeTimeout = 10 * time.Second
	}
	if out.MaxConcurrent <= 0 {
		out.MaxConcurrent = 40
	}
	if out.ShardTotal <= 0 {
		out.ShardTotal = 1
	}
	if out.ShardIndex < 0 {
		out.ShardIndex = 0
	}
	if out.BreakerWindow <= 0 {
		out.BreakerWindow = 2 * time.Minute
	}
	if out.BreakerThreshold <= 0 {
		out.BreakerThreshold = 10
	}
	if out.ListTimeout <= 0 {
		out.ListTimeout = 30 * time.Second
	}
	return out
}

// DueRepository lists monitors that are due for a probe.
type DueRepository interface {
	ListDueUptimeMonitors(ctx context.Context, arg queries.ListDueUptimeMonitorsParams) ([]queries.ListDueUptimeMonitorsRow, error)
}

// RollupRepository receives flushed hourly aggregates.
type RollupRepository interface {
	UpsertUptimeHourlyRollup(ctx context.Context, arg queries.UpsertUptimeHourlyRollupParams) error
}

// Alerter delivers incident notifications. The monitor consults it on
// transitions; the circuit breaker can suppress calls.
type Alerter interface {
	AlertDown(ctx context.Context, m Monitor, eventID int32, res ProbeResult)
	AlertRecovered(ctx context.Context, m Monitor, eventID int32, downFor time.Duration)
}

// hourAcc accumulates one hour of probes for one monitor.
type hourAcc struct {
	probesRun  int32
	probesOK   int32
	latencySum int64 // milliseconds
	latencies  []int32
	interval   int32 // last-seen probe interval, for probes_expected
}

// MonitorLoop drives probes: list due monitors, probe them concurrently with a
// bounded pool, record state, accumulate hourly rollups and alert on
// transitions.
type MonitorLoop struct {
	cfg      MonitorConfig
	dueRepo  DueRepository
	rollRepo RollupRepository
	prober   Prober
	recorder *Recorder
	alerter  Alerter
	now      func() time.Time

	sem chan struct{}

	mu           sync.Mutex
	currentHour  time.Time
	accumulators map[int32]*hourAcc
	recentDowns  []time.Time
	// suppressed tracks monitors whose down-alert was swallowed by the
	// circuit breaker, so the matching recovery notification is skipped too.
	suppressed map[int32]struct{}

	wg     sync.WaitGroup
	cancel context.CancelFunc
}

func NewMonitorLoop(cfg MonitorConfig, dueRepo DueRepository, rollRepo RollupRepository, prober Prober, recorder *Recorder, alerter Alerter) *MonitorLoop {
	return &MonitorLoop{
		cfg:          cfg.withDefaults(),
		dueRepo:      dueRepo,
		rollRepo:     rollRepo,
		prober:       prober,
		recorder:     recorder,
		alerter:      alerter,
		now:          time.Now,
		sem:          make(chan struct{}, cfg.withDefaults().MaxConcurrent),
		currentHour:  time.Now().UTC().Truncate(time.Hour),
		accumulators: make(map[int32]*hourAcc),
		suppressed:   make(map[int32]struct{}),
	}
}

// Start launches the probe loop; Stop terminates it and flushes pending data.
func (m *MonitorLoop) Start(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	m.cancel = cancel

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(m.cfg.TickInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.tick(ctx)
			}
		}
	}()
	slog.Info("uptime monitor loop started",
		"shardIndex", m.cfg.ShardIndex, "shardTotal", m.cfg.ShardTotal,
		"maxConcurrent", m.cfg.MaxConcurrent, "tickInterval", m.cfg.TickInterval)
}

// Stop halts the loop and flushes the in-flight hourly accumulator.
func (m *MonitorLoop) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
	m.wg.Wait()
	// Flush the partial hour so probes already counted are not lost.
	m.mu.Lock()
	hour := m.currentHour
	accs := m.accumulators
	m.accumulators = make(map[int32]*hourAcc)
	m.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	m.flushAccumulators(ctx, hour, accs)
}

func (m *MonitorLoop) tick(ctx context.Context) {
	m.rollHourIfNeeded(ctx)

	listCtx, cancel := context.WithTimeout(ctx, m.cfg.ListTimeout)
	due, err := m.dueRepo.ListDueUptimeMonitors(listCtx, queries.ListDueUptimeMonitorsParams{
		Now:        pgTimestamp(m.now().UTC()),
		ShardTotal: m.cfg.ShardTotal,
		ShardIndex: m.cfg.ShardIndex,
	})
	cancel()
	if err != nil {
		slog.Error("uptime: failed to list due monitors", "error", err)
		return
	}

	for _, row := range due {
		mon := monitorFromRow(row)
		select {
		case m.sem <- struct{}{}:
		case <-ctx.Done():
			return
		}
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			defer func() { <-m.sem }()
			m.probeAndRecord(ctx, mon)
		}()
	}
}

func (m *MonitorLoop) probeAndRecord(ctx context.Context, mon Monitor) {
	res := m.prober.Probe(ctx, ProbeConfig{
		URL:            mon.URL,
		ExpectedStatus: int(mon.ExpectedStatus),
		ContentMatch:   mon.ContentMatch,
		Timeout:        m.cfg.ProbeTimeout,
	})

	if res.OK {
		metrics.RecordUptimeProbe(ctx, metrics.OutcomeOK)
	} else {
		metrics.RecordUptimeProbe(ctx, metrics.OutcomeError)
	}

	outcome, err := m.recorder.Record(ctx, mon, res)
	if err != nil {
		slog.Error("uptime: failed to record probe", "environmentId", mon.EnvironmentID, "error", err)
		return
	}

	m.accumulate(mon, res)
	m.handleTransition(ctx, mon, outcome, res)
}

func (m *MonitorLoop) handleTransition(ctx context.Context, mon Monitor, outcome RecordOutcome, res ProbeResult) {
	switch outcome.Transition {
	case TransitionDown:
		metrics.RecordUptimeTransition(ctx, metrics.TransitionDown)
		if m.trackMassDown(m.now()) {
			m.mu.Lock()
			m.suppressed[mon.EnvironmentID] = struct{}{}
			m.mu.Unlock()
			slog.Warn("uptime: down alert suppressed by mass-down circuit breaker",
				"environmentId", mon.EnvironmentID)
			return
		}
		slog.Error("uptime: environment is down",
			"environmentId", mon.EnvironmentID, "url", mon.URL,
			"statusCode", res.StatusCode, "error", res.Err)
		m.alerter.AlertDown(ctx, mon, outcome.EventID, res)

	case TransitionUp:
		metrics.RecordUptimeTransition(ctx, metrics.TransitionUp)
		m.mu.Lock()
		_, wasSuppressed := m.suppressed[mon.EnvironmentID]
		delete(m.suppressed, mon.EnvironmentID)
		m.mu.Unlock()
		if wasSuppressed {
			return
		}
		if outcome.DownFor < minRecoveryNotifyDuration {
			slog.Info("uptime: environment recovered below notify threshold",
				"environmentId", mon.EnvironmentID, "downFor", outcome.DownFor)
			return
		}
		slog.Info("uptime: environment recovered",
			"environmentId", mon.EnvironmentID, "downFor", outcome.DownFor)
		m.alerter.AlertRecovered(ctx, mon, outcome.EventID, outcome.DownFor)
	}
}

// trackMassDown records a down transition and reports whether the mass-down
// circuit breaker tripped.
func (m *MonitorLoop) trackMassDown(now time.Time) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := now.Add(-m.cfg.BreakerWindow)
	kept := m.recentDowns[:0]
	for _, t := range m.recentDowns {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	kept = append(kept, now)
	m.recentDowns = kept
	return len(m.recentDowns) >= m.cfg.BreakerThreshold
}

// accumulate folds one probe into the current hour's rollup data.
func (m *MonitorLoop) accumulate(mon Monitor, res ProbeResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	acc, ok := m.accumulators[mon.EnvironmentID]
	if !ok {
		acc = &hourAcc{}
		m.accumulators[mon.EnvironmentID] = acc
	}
	acc.probesRun++
	if res.OK {
		acc.probesOK++
	}
	ms := int32(res.Latency.Milliseconds())
	acc.latencySum += int64(ms)
	// Bound memory for pathological intervals; p95 degrades gracefully.
	if len(acc.latencies) < 1000 {
		acc.latencies = append(acc.latencies, ms)
	}
	acc.interval = mon.IntervalSeconds
}

// rollHourIfNeeded flushes the accumulator whenever the UTC hour changes.
func (m *MonitorLoop) rollHourIfNeeded(ctx context.Context) {
	hour := m.now().UTC().Truncate(time.Hour)

	m.mu.Lock()
	if hour.Equal(m.currentHour) {
		m.mu.Unlock()
		return
	}
	prevHour := m.currentHour
	accs := m.accumulators
	m.currentHour = hour
	m.accumulators = make(map[int32]*hourAcc)
	m.mu.Unlock()

	m.flushAccumulators(ctx, prevHour, accs)
}

func (m *MonitorLoop) flushAccumulators(ctx context.Context, hour time.Time, accs map[int32]*hourAcc) {
	for envID, acc := range accs {
		if acc.probesRun == 0 {
			continue
		}
		expected := int32(0)
		if acc.interval > 0 {
			expected = 3600 / acc.interval
		}
		var avg, p95 *int32
		avgVal := int32(acc.latencySum / int64(acc.probesRun))
		avg = &avgVal
		if p := percentile(acc.latencies, 95); p >= 0 {
			p95 = &p
		}
		err := m.rollRepo.UpsertUptimeHourlyRollup(ctx, queries.UpsertUptimeHourlyRollupParams{
			EnvironmentID:  envID,
			Hour:           pgTimestamp(hour),
			ProbesExpected: expected,
			ProbesRun:      acc.probesRun,
			ProbesOk:       acc.probesOK,
			PausedSeconds:  0,
			LatencyAvgMs:   avg,
			LatencyP95Ms:   p95,
		})
		if err != nil {
			slog.Error("uptime: failed to flush hourly rollup", "environmentId", envID, "hour", hour, "error", err)
		}
	}
}

func percentile(values []int32, pct int) int32 {
	if len(values) == 0 {
		return -1
	}
	sorted := append([]int32(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := (len(sorted)*pct+99)/100 - 1 // nearest-rank: ceil(pct*n)th smallest
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return sorted[idx]
}

func monitorFromRow(row queries.ListDueUptimeMonitorsRow) Monitor {
	url := row.EnvironmentUrl
	if row.Url != nil && *row.Url != "" {
		url = *row.Url
	}
	contentMatch := ""
	if row.ContentMatch != nil {
		contentMatch = *row.ContentMatch
	}
	return Monitor{
		EnvironmentID:        row.EnvironmentID,
		URL:                  url,
		IntervalSeconds:      row.IntervalSeconds,
		ExpectedStatus:       row.ExpectedStatus,
		ContentMatch:         contentMatch,
		FailureThreshold:     row.FailureThreshold,
		RecoveryThreshold:    row.RecoveryThreshold,
		Status:               row.Status,
		ConsecutiveFailures:  row.ConsecutiveFailures,
		ConsecutiveSuccesses: row.ConsecutiveSuccesses,
		EnvironmentName:      row.EnvironmentName,
		OrganizationID:       row.OrganizationID,
	}
}
