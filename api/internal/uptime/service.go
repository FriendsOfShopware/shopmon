package uptime

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotAuthorized       = errors.New("not authorized for organization")
	ErrEnvironmentNotFound = errors.New("environment not found")
	ErrInvalidSettings     = errors.New("invalid uptime settings")
)

// Retention bounds: hourly rollups are superseded by daily ones after ~a
// month; resolved incidents stay for over a year so yearly availability
// remains explainable.
const (
	hourlyRollupRetention = 35 * 24 * time.Hour
	eventRetention        = 400 * 24 * time.Hour
)

// Settings is the monitor configuration a user can change.
type Settings struct {
	Enabled           bool
	URL               string // empty probes the environment's own URL
	IntervalSeconds   int32
	ExpectedStatus    int32 // 0 = any 2xx/3xx
	ContentMatch      string
	FailureThreshold  int32
	RecoveryThreshold int32
}

// SettingsView is the monitor configuration and live probe state returned by
// the API.
type SettingsView struct {
	Enabled           bool       `json:"enabled"`
	URL               *string    `json:"url"`
	IntervalSeconds   int32      `json:"intervalSeconds"`
	ExpectedStatus    int32      `json:"expectedStatus"`
	ContentMatch      *string    `json:"contentMatch"`
	FailureThreshold  int32      `json:"failureThreshold"`
	RecoveryThreshold int32      `json:"recoveryThreshold"`
	Status            string     `json:"status"`
	LastCheckedAt     *time.Time `json:"lastCheckedAt"`
	LastStatusCode    *int32     `json:"lastStatusCode"`
	LastLatencyMs     *int32     `json:"lastLatencyMs"`
	LastError         *string    `json:"lastError"`
}

// Incident is one downtime period.
type Incident struct {
	ID              int32      `json:"id"`
	StartedAt       time.Time  `json:"startedAt"`
	ResolvedAt      *time.Time `json:"resolvedAt"`
	DurationSeconds int64      `json:"durationSeconds"`
	StatusCode      *int32     `json:"statusCode"`
	LatencyMs       *int32     `json:"latencyMs"`
	Error           *string    `json:"error"`
}

// DayPoint is one day of the availability history strip.
type DayPoint struct {
	Day          string   `json:"day"` // YYYY-MM-DD
	Availability *float64 `json:"availability"`
	DownSeconds  int64    `json:"downSeconds"`
}

// LatencyPoint is one hourly latency sample for the 24h chart.
type LatencyPoint struct {
	Timestamp time.Time `json:"timestamp"`
	AvgMs     *int32    `json:"avgMs"`
	P95Ms     *int32    `json:"p95Ms"`
	ProbesRun int32     `json:"probesRun"`
}

// View is the full uptime payload for an environment.
type View struct {
	Settings     SettingsView   `json:"settings"`
	OpenIncident *Incident      `json:"openIncident"`
	Availability *float64       `json:"availability"`
	Days         []DayPoint     `json:"days"`
	Incidents    []Incident     `json:"incidents"`
	Latency      []LatencyPoint `json:"latency"`
}

// Service is the API- and job-facing surface of uptime monitoring. The probe
// loop lives in MonitorLoop; this service handles user reads/writes and the
// scheduled rollup/retention work.
type Service struct {
	pool *pgxpool.Pool
	q    *queries.Queries
	now  func() time.Time
}

func NewService(pool *pgxpool.Pool, q *queries.Queries) *Service {
	return &Service{pool: pool, q: q, now: time.Now}
}

// authorize resolves the environment's organization and checks membership.
func (s *Service) authorize(ctx context.Context, userID string, environmentID int32) error {
	orgID, err := s.q.GetEnvironmentOrganizationID(ctx, environmentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrEnvironmentNotFound
	}
	if err != nil {
		return fmt.Errorf("get environment organization: %w", err)
	}
	member, err := s.q.IsOrganizationMember(ctx, queries.IsOrganizationMemberParams{
		OrganizationID: orgID,
		UserID:         userID,
	})
	if err != nil {
		return fmt.Errorf("check organization membership: %w", err)
	}
	if !member {
		return ErrNotAuthorized
	}
	return nil
}

func validateSettings(st Settings) error {
	if st.IntervalSeconds < 30 || st.IntervalSeconds > 3600 {
		return fmt.Errorf("%w: intervalSeconds must be between 30 and 3600", ErrInvalidSettings)
	}
	if st.ExpectedStatus != 0 && (st.ExpectedStatus < 100 || st.ExpectedStatus > 599) {
		return fmt.Errorf("%w: expectedStatus must be 0 or between 100 and 599", ErrInvalidSettings)
	}
	if st.FailureThreshold < 1 || st.FailureThreshold > 10 {
		return fmt.Errorf("%w: failureThreshold must be between 1 and 10", ErrInvalidSettings)
	}
	if st.RecoveryThreshold < 1 || st.RecoveryThreshold > 10 {
		return fmt.Errorf("%w: recoveryThreshold must be between 1 and 10", ErrInvalidSettings)
	}
	if st.URL != "" {
		parsed, err := url.Parse(st.URL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return fmt.Errorf("%w: url must be a valid http(s) URL", ErrInvalidSettings)
		}
		if err := validatePublicHost(parsed.Host); err != nil {
			return fmt.Errorf("%w: url %v", ErrInvalidSettings, err)
		}
	}
	return nil
}

// validatePublicHost rejects obviously internal targets at settings time so
// users get immediate feedback. The authoritative guard is the dial-time
// address check in the prober, which also covers redirects and DNS drift.
func validatePublicHost(host string) error {
	hostname := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	}
	if strings.EqualFold(hostname, "localhost") || strings.HasSuffix(strings.ToLower(hostname), ".localhost") {
		return errors.New("host must be a public address")
	}
	if ip := net.ParseIP(strings.Trim(hostname, "[]")); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
			return errors.New("host must be a public address")
		}
	}
	return nil
}

// UpdateSettings creates or updates the monitor for an environment. Disabling
// resolves any open incident, since the system stops making claims about the
// shop's availability; the upsert and the resolve run in one transaction so a
// monitor can never end up disabled with an incident that nothing will close.
func (s *Service) UpdateSettings(ctx context.Context, userID string, environmentID int32, st Settings) error {
	if err := s.authorize(ctx, userID, environmentID); err != nil {
		return err
	}
	if err := validateSettings(st); err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin uptime settings transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	qtx := s.q.WithTx(tx)

	err = qtx.UpsertUptimeSettings(ctx, queries.UpsertUptimeSettingsParams{
		EnvironmentID:     environmentID,
		Enabled:           st.Enabled,
		Url:               nilIfEmpty(st.URL),
		IntervalSeconds:   st.IntervalSeconds,
		ExpectedStatus:    st.ExpectedStatus,
		ContentMatch:      nilIfEmpty(st.ContentMatch),
		FailureThreshold:  st.FailureThreshold,
		RecoveryThreshold: st.RecoveryThreshold,
	})
	if err != nil {
		return fmt.Errorf("upsert uptime settings: %w", err)
	}

	if !st.Enabled {
		if err := qtx.ResolveUptimeEvents(ctx, queries.ResolveUptimeEventsParams{
			EnvironmentID: environmentID,
			ResolvedAt:    pgTimestamp(s.now()),
		}); err != nil {
			return fmt.Errorf("resolve open uptime events: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit uptime settings: %w", err)
	}
	return nil
}

// GetView returns the monitor state, availability and history for a range.
// Supported ranges: "24h", "7d", "30d", "90d".
func (s *Service) GetView(ctx context.Context, userID string, environmentID int32, rangeKey string) (View, error) {
	var view View
	if err := s.authorize(ctx, userID, environmentID); err != nil {
		return view, err
	}
	days, err := rangeDays(rangeKey)
	if err != nil {
		return view, err
	}

	now := s.now().UTC()

	settings, err := s.q.GetUptimeSettings(ctx, environmentID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		view.Settings = SettingsView{IntervalSeconds: 60, ExpectedStatus: 0, FailureThreshold: 3, RecoveryThreshold: 2, Status: StatusUnknown}
	case err != nil:
		return view, fmt.Errorf("get uptime settings: %w", err)
	default:
		view.Settings = SettingsView{
			Enabled:           settings.Enabled,
			URL:               settings.Url,
			IntervalSeconds:   settings.IntervalSeconds,
			ExpectedStatus:    settings.ExpectedStatus,
			ContentMatch:      settings.ContentMatch,
			FailureThreshold:  settings.FailureThreshold,
			RecoveryThreshold: settings.RecoveryThreshold,
			Status:            settings.Status,
			LastCheckedAt:     timePtr(settings.LastCheckedAt),
			LastStatusCode:    settings.LastStatusCode,
			LastLatencyMs:     settings.LastLatencyMs,
			LastError:         settings.LastError,
		}
	}

	// The open incident is range-independent so an outage that started before
	// the window is still visible.
	open, err := s.q.GetOpenUptimeEvent(ctx, environmentID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
	case err != nil:
		return view, fmt.Errorf("get open uptime event: %w", err)
	default:
		incident := incidentFromRow(open, now)
		view.OpenIncident = &incident
	}

	// The 24h range is a rolling now-24h window so availability, incidents
	// and the latency series all cover the same period; multi-day ranges are
	// UTC-day aligned to match the day strip.
	rangeStart := startOfDay(now).AddDate(0, 0, -(days - 1))
	if days == 1 {
		rangeStart = now.Add(-24 * time.Hour)
	}

	if days == 1 {
		if err := s.getViewRolling24h(ctx, environmentID, now, rangeStart, &view); err != nil {
			return view, err
		}
		return view, nil
	}

	// Availability: downtime always comes from incidents (the source of
	// truth); daily rollups refine the denominator by excluding recorded
	// no-data time. A day with incidents but no rollup yet (the nightly
	// aggregation has not run) counts fully, so incidents are never lost.
	todayStart := startOfDay(now)

	rollups, err := s.q.ListUptimeDailyRollups(ctx, queries.ListUptimeDailyRollupsParams{
		EnvironmentID: environmentID,
		Day:           pgDate(rangeStart),
		Day_2:         pgDate(todayStart.AddDate(0, 0, -1)),
	})
	if err != nil {
		return view, fmt.Errorf("list daily rollups: %w", err)
	}
	rollupByDay := make(map[string]queries.EnvironmentUptimeRollupDaily, len(rollups))
	for _, r := range rollups {
		rollupByDay[r.Day.Time.Format("2006-01-02")] = r
	}

	var upSeconds, downSeconds int64
	fullDayEvents, err := s.q.ListUptimeEventsOverlapping(ctx, queries.ListUptimeEventsOverlappingParams{
		EnvironmentID: environmentID,
		RangeFrom:     pgTimestamp(rangeStart),
		RangeTo:       pgTimestamp(todayStart),
	})
	if err != nil {
		return view, fmt.Errorf("list uptime events: %w", err)
	}

	for day := rangeStart; day.Before(todayStart); day = day.AddDate(0, 0, 1) {
		if r, ok := rollupByDay[day.Format("2006-01-02")]; ok {
			upSeconds += int64(r.UpSeconds)
			downSeconds += int64(r.DownSeconds)
			continue
		}
		// No rollup row: count the day only when an incident proves the
		// monitor was running.
		dayDown := overlapSeconds(fullDayEvents, day, day.AddDate(0, 0, 1))
		if dayDown > 0 {
			upSeconds += 24*3600 - dayDown
			downSeconds += dayDown
		}
	}
	todayDown, err := s.downSecondsInRange(ctx, environmentID, todayStart, now)
	if err != nil {
		return view, err
	}
	elapsedToday := int64(now.Sub(todayStart).Seconds())
	upSeconds += elapsedToday - todayDown
	downSeconds += todayDown
	if total := upSeconds + downSeconds; total > 0 {
		availability := float64(upSeconds) / float64(total)
		view.Availability = &availability
	}

	// Incidents within the range.
	events, err := s.q.ListUptimeEvents(ctx, queries.ListUptimeEventsParams{
		EnvironmentID: environmentID,
		StartedAt:     pgTimestamp(rangeStart),
	})
	if err != nil {
		return view, fmt.Errorf("list uptime events: %w", err)
	}
	view.Incidents = make([]Incident, 0, len(events))
	for _, e := range events {
		view.Incidents = append(view.Incidents, incidentFromRow(e, now))
	}

	// Day strip, reusing the rollups fetched for availability.
	view.Days = make([]DayPoint, 0, days)
	for i := 0; i < days; i++ {
		day := rangeStart.AddDate(0, 0, i)
		key := day.Format("2006-01-02")
		if i == days-1 {
			// Today is synthesized from incidents; elapsed time is the
			// denominator.
			elapsed := int64(now.Sub(day).Seconds())
			if elapsed <= 0 {
				view.Days = append(view.Days, DayPoint{Day: key})
				continue
			}
			availability := float64(elapsed-todayDown) / float64(elapsed)
			view.Days = append(view.Days, DayPoint{Day: key, Availability: &availability, DownSeconds: todayDown})
			continue
		}
		r, ok := rollupByDay[key]
		if !ok {
			view.Days = append(view.Days, DayPoint{Day: key})
			continue
		}
		if total := r.UpSeconds + r.DownSeconds; total > 0 {
			availability := float64(r.UpSeconds) / float64(total)
			view.Days = append(view.Days, DayPoint{Day: key, Availability: &availability, DownSeconds: int64(r.DownSeconds)})
		} else {
			view.Days = append(view.Days, DayPoint{Day: key})
		}
	}

	return view, nil
}

// getViewRolling24h fills availability, incidents and the latency series for
// the rolling 24h view. The availability denominator starts at the earliest
// evidence of monitoring inside the window (rollups or incidents), so time
// before the monitor existed is not silently counted as up.
func (s *Service) getViewRolling24h(ctx context.Context, environmentID int32, now, rangeStart time.Time, view *View) error {
	hourlies, err := s.q.ListUptimeHourlyRollups(ctx, queries.ListUptimeHourlyRollupsParams{
		EnvironmentID: environmentID,
		Hour:          pgTimestamp(rangeStart),
		Hour_2:        pgTimestamp(now),
	})
	if err != nil {
		return fmt.Errorf("list hourly rollups: %w", err)
	}
	view.Latency = make([]LatencyPoint, 0, len(hourlies))
	for _, h := range hourlies {
		view.Latency = append(view.Latency, LatencyPoint{
			Timestamp: h.Hour.Time,
			AvgMs:     h.LatencyAvgMs,
			P95Ms:     h.LatencyP95Ms,
			ProbesRun: h.ProbesRun,
		})
	}

	events, err := s.q.ListUptimeEvents(ctx, queries.ListUptimeEventsParams{
		EnvironmentID: environmentID,
		StartedAt:     pgTimestamp(rangeStart),
	})
	if err != nil {
		return fmt.Errorf("list uptime events: %w", err)
	}
	view.Incidents = make([]Incident, 0, len(events))
	for _, e := range events {
		view.Incidents = append(view.Incidents, incidentFromRow(e, now))
	}

	// Downtime must include incidents that started before the window and are
	// still open (or resolved inside it).
	overlapping, err := s.q.ListUptimeEventsOverlapping(ctx, queries.ListUptimeEventsOverlappingParams{
		EnvironmentID: environmentID,
		RangeFrom:     pgTimestamp(rangeStart),
		RangeTo:       pgTimestamp(now),
	})
	if err != nil {
		return fmt.Errorf("list overlapping uptime events: %w", err)
	}

	// Bound the denominator by the earliest evidence of monitoring inside
	// the window (rollups or incidents); with no evidence at all there is no
	// availability to report and pre-monitor time is never counted as up.
	dataStart := rangeStart
	evidence := false
	for _, e := range overlapping {
		evidence = true
		if e.StartedAt.Time.After(rangeStart) && e.StartedAt.Time.Before(dataStart) {
			dataStart = e.StartedAt.Time
		}
	}
	for _, h := range hourlies {
		evidence = true
		if h.Hour.Time.After(rangeStart) && h.Hour.Time.Before(dataStart) {
			dataStart = h.Hour.Time
		}
	}

	if !evidence {
		return nil
	}
	elapsed := int64(now.Sub(dataStart).Seconds())
	if elapsed > 0 {
		downSeconds := overlapSeconds(overlapping, dataStart, now)
		availability := float64(elapsed-downSeconds) / float64(elapsed)
		view.Availability = &availability
	}
	return nil
}

// downSecondsInRange sums incident time clipped to [from, to).
func (s *Service) downSecondsInRange(ctx context.Context, environmentID int32, from, to time.Time) (int64, error) {
	events, err := s.q.ListUptimeEventsOverlapping(ctx, queries.ListUptimeEventsOverlappingParams{
		EnvironmentID: environmentID,
		RangeFrom:     pgTimestamp(from),
		RangeTo:       pgTimestamp(to),
	})
	if err != nil {
		return 0, fmt.Errorf("list overlapping uptime events: %w", err)
	}
	return overlapSeconds(events, from, to), nil
}

func overlapSeconds(events []queries.EnvironmentUptimeEvent, from, to time.Time) int64 {
	var total int64
	for _, e := range events {
		start := e.StartedAt.Time
		if start.Before(from) {
			start = from
		}
		end := to
		if e.ResolvedAt.Valid && e.ResolvedAt.Time.Before(end) {
			end = e.ResolvedAt.Time
		}
		if end.After(start) {
			total += int64(end.Sub(start).Seconds())
		}
	}
	return total
}

// PauseMonitor suspends probing without touching configuration; used when a
// deployment is rolling out. Paused time is excluded from availability.
func (s *Service) PauseMonitor(ctx context.Context, environmentID int32) error {
	if err := s.q.PauseUptimeMonitor(ctx, environmentID); err != nil {
		return fmt.Errorf("pause uptime monitor: %w", err)
	}
	return nil
}

// ResumeMonitor re-arms a paused monitor.
func (s *Service) ResumeMonitor(ctx context.Context, environmentID int32) error {
	if err := s.q.ResumeUptimeMonitor(ctx, environmentID); err != nil {
		return fmt.Errorf("resume uptime monitor: %w", err)
	}
	return nil
}

// RunDailyMaintenance aggregates the previous UTC day into daily rollups and
// enforces retention. Scheduled once a day via the queue.
func (s *Service) RunDailyMaintenance(ctx context.Context) error {
	now := s.now().UTC()
	yesterday := startOfDay(now).AddDate(0, 0, -1)
	if err := s.AggregateDay(ctx, yesterday); err != nil {
		return fmt.Errorf("aggregate daily uptime rollups: %w", err)
	}

	if err := s.q.DeleteUptimeHourlyRollupsOlderThan(ctx, pgTimestamp(now.Add(-hourlyRollupRetention))); err != nil {
		return fmt.Errorf("delete old hourly rollups: %w", err)
	}
	if err := s.q.DeleteUptimeEventsOlderThan(ctx, pgTimestamp(now.Add(-eventRetention))); err != nil {
		return fmt.Errorf("delete old uptime events: %w", err)
	}
	return nil
}

// AggregateDay computes the daily rollup rows for one UTC day from incident
// events and hourly rollups.
func (s *Service) AggregateDay(ctx context.Context, day time.Time) error {
	from := startOfDay(day)
	to := from.AddDate(0, 0, 1)

	targets, err := s.q.ListUptimeDailyAggregateTargets(ctx, queries.ListUptimeDailyAggregateTargetsParams{
		RangeFrom: pgTimestamp(from),
		RangeTo:   pgTimestamp(to),
	})
	if err != nil {
		return fmt.Errorf("list aggregate targets: %w", err)
	}

	for _, environmentID := range targets {
		if err := s.aggregateEnvironmentDay(ctx, environmentID, from, to); err != nil {
			return fmt.Errorf("aggregate environment %d: %w", environmentID, err)
		}
	}
	return nil
}

func (s *Service) aggregateEnvironmentDay(ctx context.Context, environmentID int32, from, to time.Time) error {
	events, err := s.q.ListUptimeEventsOverlapping(ctx, queries.ListUptimeEventsOverlappingParams{
		EnvironmentID: environmentID,
		RangeFrom:     pgTimestamp(from),
		RangeTo:       pgTimestamp(to),
	})
	if err != nil {
		return fmt.Errorf("list events: %w", err)
	}
	hourlies, err := s.q.ListUptimeHourlyRollups(ctx, queries.ListUptimeHourlyRollupsParams{
		EnvironmentID: environmentID,
		Hour:          pgTimestamp(from),
		Hour_2:        pgTimestamp(to),
	})
	if err != nil {
		return fmt.Errorf("list hourly rollups: %w", err)
	}
	hourByTime := make(map[time.Time]queries.EnvironmentUptimeRollupHourly, len(hourlies))
	for _, h := range hourlies {
		hourByTime[h.Hour.Time] = h
	}

	downSeconds := overlapSeconds(events, from, to)

	// Hours with no probes at all carry no data; they leave the availability
	// denominator instead of counting as up. Partial hours count as data.
	var nodataSeconds int64
	for i := 0; i < 24; i++ {
		hourStart := from.Add(time.Duration(i) * time.Hour)
		h, ok := hourByTime[hourStart]
		if ok && h.ProbesRun > 0 {
			continue
		}
		nodataSeconds += 3600 - overlapSeconds(events, hourStart, hourStart.Add(time.Hour))
	}
	if nodataSeconds < 0 {
		nodataSeconds = 0
	}

	upSeconds := 24*3600 - downSeconds - nodataSeconds
	if upSeconds < 0 {
		upSeconds = 0
	}

	// Latency: probe-weighted average of hourly averages; p95 is the worst
	// hourly p95 (a documented approximation).
	var latencyWeightedSum int64
	var latencyWeight int64
	var p95 *int32
	for _, h := range hourlies {
		if h.LatencyAvgMs != nil && h.ProbesRun > 0 {
			latencyWeightedSum += int64(*h.LatencyAvgMs) * int64(h.ProbesRun)
			latencyWeight += int64(h.ProbesRun)
		}
		if h.LatencyP95Ms != nil && (p95 == nil || *h.LatencyP95Ms > *p95) {
			v := *h.LatencyP95Ms
			p95 = &v
		}
	}
	var avg *int32
	if latencyWeight > 0 {
		v := int32(latencyWeightedSum / latencyWeight)
		avg = &v
	}

	err = s.q.UpsertUptimeDailyRollup(ctx, queries.UpsertUptimeDailyRollupParams{
		EnvironmentID: environmentID,
		Day:           pgDate(from),
		UpSeconds:     int32(upSeconds),
		DownSeconds:   int32(downSeconds),
		PausedSeconds: 0,
		NodataSeconds: int32(nodataSeconds),
		LatencyAvgMs:  avg,
		LatencyP95Ms:  p95,
	})
	if err != nil {
		return fmt.Errorf("upsert daily rollup: %w", err)
	}
	return nil
}

func incidentFromRow(e queries.EnvironmentUptimeEvent, now time.Time) Incident {
	incident := Incident{
		ID:         e.ID,
		StartedAt:  e.StartedAt.Time,
		StatusCode: e.StatusCode,
		LatencyMs:  e.LatencyMs,
		Error:      e.Error,
	}
	if e.ResolvedAt.Valid {
		resolved := e.ResolvedAt.Time
		incident.ResolvedAt = &resolved
		incident.DurationSeconds = int64(resolved.Sub(e.StartedAt.Time).Seconds())
	} else {
		incident.DurationSeconds = int64(now.Sub(e.StartedAt.Time).Seconds())
	}
	return incident
}

func rangeDays(rangeKey string) (int, error) {
	switch rangeKey {
	case "24h", "":
		return 1, nil
	case "7d":
		return 7, nil
	case "30d":
		return 30, nil
	case "90d":
		return 90, nil
	default:
		return 0, fmt.Errorf("%w: unsupported range %q", ErrInvalidSettings, rangeKey)
	}
}

func startOfDay(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func nilIfEmpty(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
