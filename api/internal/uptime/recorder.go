// Package uptime implements external storefront uptime monitoring: a probe
// loop in the worker, incident tracking with flap suppression, availability
// rollups, and alerting through the notify package.
//
// Probes run outside the queue (they are timing-critical and discardable);
// only durable side effects (notifications, daily rollups) take the queue or
// are otherwise scheduled.
package uptime

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Monitor states persisted on environment_uptime.status.
const (
	StatusUnknown = "unknown"
	StatusUp      = "up"
	StatusDown    = "down"
	StatusPaused  = "paused"
)

// Transition reports what a recorded probe changed.
type Transition int

const (
	TransitionNone Transition = iota
	// TransitionDown fired when the failure threshold was crossed and an
	// incident was opened.
	TransitionDown
	// TransitionUp fired when the recovery threshold was crossed and the open
	// incident was resolved.
	TransitionUp
)

// Monitor carries the configuration and live state of one monitor, as loaded
// for a due probe.
type Monitor struct {
	EnvironmentID        int32
	URL                  string
	IntervalSeconds      int32
	ExpectedStatus       int32
	ContentMatch         string
	FailureThreshold     int32
	RecoveryThreshold    int32
	Status               string
	ConsecutiveFailures  int32
	ConsecutiveSuccesses int32

	// Notification context.
	EnvironmentName string
	OrganizationID  string
}

// RecorderRepository is the persistence surface the recorder needs. It is an
// interface so the state machine can be tested against fakes.
type RecorderRepository interface {
	UpdateUptimeProbeState(ctx context.Context, arg queries.UpdateUptimeProbeStateParams) error
	InsertUptimeEvent(ctx context.Context, arg queries.InsertUptimeEventParams) (int32, error)
	GetOpenUptimeEvent(ctx context.Context, environmentID int32) (queries.EnvironmentUptimeEvent, error)
	ResolveUptimeEvents(ctx context.Context, arg queries.ResolveUptimeEventsParams) error
}

// Recorder applies probe results to monitor state with flap suppression and
// maintains the incident (event) rows.
type Recorder struct {
	repo RecorderRepository
	now  func() time.Time
}

func NewRecorder(repo RecorderRepository) *Recorder {
	return &Recorder{repo: repo, now: time.Now}
}

// RecordOutcome reports what a recorded probe changed.
type RecordOutcome struct {
	Transition Transition
	// EventID is the incident opened by a TransitionDown, or the incident
	// resolved by a TransitionUp.
	EventID int32
	// DownFor is the duration of the incident closed by a TransitionUp.
	DownFor time.Duration
}

// Record applies one probe result. It writes the incident rows on transitions
// and always updates the persisted probe state (counters, last_* fields).
func (r *Recorder) Record(ctx context.Context, m Monitor, res ProbeResult) (RecordOutcome, error) {
	var outcome RecordOutcome
	now := r.now()

	failures := m.ConsecutiveFailures
	successes := m.ConsecutiveSuccesses
	status := m.Status

	if !res.OK {
		failures++
		successes = 0
		// The threshold gates the transition: single flaky probes must not
		// open incidents. While below it, the previously healthy state stands.
		if m.Status != StatusDown && failures >= m.FailureThreshold {
			status = StatusDown
			eventID, err := r.repo.InsertUptimeEvent(ctx, queries.InsertUptimeEventParams{
				EnvironmentID: m.EnvironmentID,
				StartedAt:     pgtype.Timestamp{Time: now.UTC(), Valid: true},
				StatusCode:    nullInt32(int32(res.StatusCode)),
				LatencyMs:     int32Ptr(int32(res.Latency.Milliseconds())),
				Error:         nullString(res.Err),
			})
			if err != nil {
				return outcome, fmt.Errorf("insert uptime event: %w", err)
			}
			outcome.Transition = TransitionDown
			outcome.EventID = eventID
		}
	} else {
		successes++
		failures = 0
		switch {
		case m.Status == StatusDown && successes >= m.RecoveryThreshold:
			open, err := r.repo.GetOpenUptimeEvent(ctx, m.EnvironmentID)
			switch {
			case err == nil && open.StartedAt.Valid:
				outcome.DownFor = now.Sub(open.StartedAt.Time)
				outcome.EventID = open.ID
			case err == nil:
				outcome.EventID = open.ID
			case errors.Is(err, pgx.ErrNoRows):
				// No open incident row (data drift); resolve below is a no-op.
			case err != nil:
				return outcome, fmt.Errorf("get open uptime event: %w", err)
			}
			if err := r.repo.ResolveUptimeEvents(ctx, queries.ResolveUptimeEventsParams{
				EnvironmentID: m.EnvironmentID,
				ResolvedAt:    pgtype.Timestamp{Time: now.UTC(), Valid: true},
			}); err != nil {
				return outcome, fmt.Errorf("resolve uptime event: %w", err)
			}
			status = StatusUp
			outcome.Transition = TransitionUp
		case m.Status != StatusDown:
			status = StatusUp
		}
	}

	// next_check_at is not touched here: the claim that scheduled this probe
	// already pushed it forward, which is what prevents double probes.
	err := r.repo.UpdateUptimeProbeState(ctx, queries.UpdateUptimeProbeStateParams{
		EnvironmentID:        m.EnvironmentID,
		Status:               status,
		ConsecutiveFailures:  failures,
		ConsecutiveSuccesses: successes,
		LastStatusCode:       nullInt32(int32(res.StatusCode)),
		LastLatencyMs:        int32Ptr(int32(res.Latency.Milliseconds())),
		LastError:            nullString(res.Err),
	})
	if err != nil {
		return outcome, fmt.Errorf("update uptime probe state: %w", err)
	}

	return outcome, nil
}

// nullInt32 maps 0 to nil. Only meaningful for status codes, where 0 means
// "no HTTP response was received".
func nullInt32(v int32) *int32 {
	if v == 0 {
		return nil
	}
	return &v
}

// int32Ptr always stores the value: a latency of 0 ms is a real measurement,
// not a missing one.
func int32Ptr(v int32) *int32 {
	return &v
}

func nullString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
