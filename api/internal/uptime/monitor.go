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

// stopGraceMargin is added to ProbeTimeout to bound how long Stop waits for
// in-flight probes before cancelling them.
const stopGraceMargin = 5 * time.Second

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
	// if at least BreakerThreshold distinct monitors go down within
	// BreakerWindow, the cause is almost certainly the probe infrastructure,
	// not the shops, and customer notifications are suppressed.
	BreakerWindow    time.Duration
	BreakerThreshold int
	// ListTimeout bounds the due-monitors claim query.
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

// DueRepository claims monitors that are due for a probe. The claim must be
// atomic (advance next_check_at in the same statement) so two overlapping
// ticks can never probe the same monitor twice.
type DueRepository interface {
	ClaimDueUptimeMonitors(ctx context.Context, arg queries.ClaimDueUptimeMonitorsParams) ([]queries.ClaimDueUptimeMonitorsRow, error)
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
	hour       time.Time
	probesRun  int32
	probesOK   int32
	latencySum int64 // milliseconds
	latencies  []int32
	interval   int32 // last-seen probe interval, for probes_expected
}

func (a *hourAcc) fold(mon Monitor, res ProbeResult) {
	a.probesRun++
	if res.OK {
		a.probesOK++
	}
	ms := int32(res.Latency.Milliseconds())
	a.latencySum += int64(ms)
	// Bound memory for pathological intervals; p95 degrades gracefully.
	if len(a.latencies) < 1000 {
		a.latencies = append(a.latencies, ms)
	}
	a.interval = mon.IntervalSeconds
}

// MonitorLoop drives probes: claim due monitors, probe them concurrently with
// a bounded pool, record state, accumulate hourly rollups and alert on
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
	accumulators map[int32]*hourAcc
	// recentDowns tracks the last down transition per monitor inside the
	// breaker window; recoveries remove their entry so a set of flapping
	// shops cannot keep the breaker tripped forever.
	recentDowns map[int32]time.Time
	// suppressed tracks monitors whose down-alert was swallowed by the
	// circuit breaker, so the matching recovery notification is skipped too.
	suppressed map[int32]struct{}

	wg sync.WaitGroup
	// cancel stops the ticker; probeCtx/probeCancel bound the probes
	// themselves. Probes run on their own context so a shutdown signal does
	// not destroy their results before they are recorded.
	stopMu      sync.Mutex
	cancel      context.CancelFunc
	probeCtx    context.Context
	probeCancel context.CancelFunc
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
		accumulators: make(map[int32]*hourAcc),
		recentDowns:  make(map[int32]time.Time),
		suppressed:   make(map[int32]struct{}),
	}
}

// Start launches the probe loop; Stop terminates it and flushes pending data.
func (m *MonitorLoop) Start(parent context.Context) {
	// The ticker follows the parent so external cancellation halts scheduling
	// immediately; probes get an independent context and rely on Stop's grace
	// period, otherwise a down-transition discovered right at shutdown would
	// lose both its state write and its incident row.
	tickCtx, cancel := context.WithCancel(parent)
	probeCtx, probeCancel := context.WithCancel(context.Background())

	m.stopMu.Lock()
	m.cancel = cancel
	m.probeCtx = probeCtx
	m.probeCancel = probeCancel
	m.stopMu.Unlock()

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(m.cfg.TickInterval)
		defer ticker.Stop()
		for {
			select {
			case <-tickCtx.Done():
				return
			case <-ticker.C:
				m.tick(probeCtx)
			}
		}
	}()
	slog.Info("uptime monitor loop started",
		"shardIndex", m.cfg.ShardIndex, "shardTotal", m.cfg.ShardTotal,
		"maxConcurrent", m.cfg.MaxConcurrent, "tickInterval", m.cfg.TickInterval)
}

// Stop halts the loop, waits for in-flight probes to finish recording (bounded
// by the probe timeout plus a margin), and flushes pending hourly data.
func (m *MonitorLoop) Stop() {
	m.stopMu.Lock()
	cancel := m.cancel
	probeCancel := m.probeCancel
	m.stopMu.Unlock()
	if cancel != nil {
		cancel()
	}

	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(m.cfg.ProbeTimeout + stopGraceMargin):
		// Probes still running are stuck; abandon them so shutdown proceeds.
		if probeCancel != nil {
			probeCancel()
		}
		<-done
	}

	// Flush the partial hour so probes already counted are not lost.
	ctx, cancelFlush := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFlush()
	accs := m.takeAccumulators()
	m.flushAccumulators(ctx, accs)
}

// takeAccumulators detaches all pending accumulators under the lock.
func (m *MonitorLoop) takeAccumulators() map[int32]*hourAcc {
	m.mu.Lock()
	defer m.mu.Unlock()
	accs := m.accumulators
	m.accumulators = make(map[int32]*hourAcc)
	return accs
}

func (m *MonitorLoop) tick(ctx context.Context) {
	m.rollHourIfNeeded(ctx)

	listCtx, cancel := context.WithTimeout(ctx, m.cfg.ListTimeout)
	claimed, err := m.dueRepo.ClaimDueUptimeMonitors(listCtx, queries.ClaimDueUptimeMonitorsParams{
		Now:        pgTimestamp(m.now().UTC()),
		ShardTotal: m.cfg.ShardTotal,
		ShardIndex: m.cfg.ShardIndex,
	})
	cancel()
	if err != nil {
		slog.Error("uptime: failed to claim due monitors", "error", err)
		return
	}

	for _, row := range claimed {
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
	startedAt := m.now()

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

	m.accumulate(mon, res, startedAt)
	m.handleTransition(ctx, mon, outcome, res)
}

func (m *MonitorLoop) handleTransition(ctx context.Context, mon Monitor, outcome RecordOutcome, res ProbeResult) {
	switch outcome.Transition {
	case TransitionDown:
		metrics.RecordUptimeTransition(ctx, metrics.TransitionDown)
		if m.trackMassDown(mon.EnvironmentID, m.now()) {
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
		// A recovery ends the incident: it must also leave the mass-down
		// window so recovered shops stop counting towards the breaker.
		m.clearMassDown(mon.EnvironmentID)
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
// circuit breaker tripped (distinct monitors down within the window).
func (m *MonitorLoop) trackMassDown(envID int32, now time.Time) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := now.Add(-m.cfg.BreakerWindow)
	for id, t := range m.recentDowns {
		if t.Before(cutoff) {
			delete(m.recentDowns, id)
		}
	}
	m.recentDowns[envID] = now
	return len(m.recentDowns) >= m.cfg.BreakerThreshold
}

// clearMassDown removes a recovered monitor from the mass-down window.
func (m *MonitorLoop) clearMassDown(envID int32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.recentDowns, envID)
}

// accumulate folds one probe into the rollup bucket of the hour the probe
// started in, so slow probes completing after an hour rollover are not
// attributed to the wrong bucket.
func (m *MonitorLoop) accumulate(mon Monitor, res ProbeResult, startedAt time.Time) {
	hour := startedAt.UTC().Truncate(time.Hour)

	m.mu.Lock()
	acc, ok := m.accumulators[mon.EnvironmentID]
	if ok && acc.hour.Equal(hour) {
		acc.fold(mon, res)
		m.mu.Unlock()
		return
	}
	if ok {
		// Hour boundary crossed between probes: detach the buffered hour
		// and start a new one. The additive upsert makes flushing the old
		// bucket separately safe.
		delete(m.accumulators, mon.EnvironmentID)
	}
	fresh := &hourAcc{hour: hour}
	fresh.fold(mon, res)
	m.accumulators[mon.EnvironmentID] = fresh
	m.mu.Unlock()

	if ok {
		m.flushAccumulators(m.probeFlushCtx(), map[int32]*hourAcc{mon.EnvironmentID: acc})
	}
}

// probeFlushCtx returns the probe context for incidental flushes outside the
// hour roll; it is cancelled by Stop to avoid writes after shutdown.
func (m *MonitorLoop) probeFlushCtx() context.Context {
	m.stopMu.Lock()
	defer m.stopMu.Unlock()
	if m.probeCtx != nil {
		return m.probeCtx
	}
	return context.Background()
}

// rollHourIfNeeded flushes accumulators whenever the UTC hour changes.
func (m *MonitorLoop) rollHourIfNeeded(ctx context.Context) {
	hour := m.now().UTC().Truncate(time.Hour)

	m.mu.Lock()
	var stale []int32
	for id, acc := range m.accumulators {
		if acc.hour.Before(hour) {
			stale = append(stale, id)
		}
	}
	if len(stale) == 0 {
		m.mu.Unlock()
		return
	}
	detached := make(map[int32]*hourAcc, len(stale))
	for _, id := range stale {
		detached[id] = m.accumulators[id]
		delete(m.accumulators, id)
	}
	m.mu.Unlock()

	m.flushAccumulators(ctx, detached)
}

func (m *MonitorLoop) flushAccumulators(ctx context.Context, accs map[int32]*hourAcc) {
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
			Hour:           pgTimestamp(acc.hour),
			ProbesExpected: expected,
			ProbesRun:      acc.probesRun,
			ProbesOk:       acc.probesOK,
			PausedSeconds:  0,
			LatencyAvgMs:   avg,
			LatencyP95Ms:   p95,
		})
		if err != nil {
			slog.Error("uptime: failed to flush hourly rollup", "environmentId", envID, "hour", acc.hour, "error", err)
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

func monitorFromRow(row queries.ClaimDueUptimeMonitorsRow) Monitor {
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
