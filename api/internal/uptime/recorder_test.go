package uptime

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRecorderRepo is an in-memory RecorderRepository.
type fakeRecorderRepo struct {
	state         queries.UpdateUptimeProbeStateParams
	stateCalls    int
	insertedEvent int32
	nextEventID   int32
	openEvent     queries.EnvironmentUptimeEvent
	openEventErr  error
	resolved      []queries.ResolveUptimeEventsParams
	insertErr     error
	updateErr     error
}

func (f *fakeRecorderRepo) UpdateUptimeProbeState(_ context.Context, arg queries.UpdateUptimeProbeStateParams) error {
	f.state = arg
	f.stateCalls++
	return f.updateErr
}

func (f *fakeRecorderRepo) InsertUptimeEvent(_ context.Context, _ queries.InsertUptimeEventParams) (int32, error) {
	if f.insertErr != nil {
		return 0, f.insertErr
	}
	f.nextEventID++
	f.insertedEvent = f.nextEventID
	return f.insertedEvent, nil
}

func (f *fakeRecorderRepo) GetOpenUptimeEvent(_ context.Context, _ int32) (queries.EnvironmentUptimeEvent, error) {
	return f.openEvent, f.openEventErr
}

func (f *fakeRecorderRepo) ResolveUptimeEvents(_ context.Context, arg queries.ResolveUptimeEventsParams) error {
	f.resolved = append(f.resolved, arg)
	return nil
}

func healthyMonitor() Monitor {
	return Monitor{
		EnvironmentID:     7,
		URL:               "https://shop.example.com",
		IntervalSeconds:   60,
		FailureThreshold:  3,
		RecoveryThreshold: 2,
		Status:            StatusUp,
		EnvironmentName:   "Production",
		OrganizationID:    "org-1",
	}
}

func TestRecorderStaysUpBelowFailureThreshold(t *testing.T) {
	repo := &fakeRecorderRepo{}
	recorder := NewRecorder(repo)
	monitor := healthyMonitor()

	outcome, err := recorder.Record(context.Background(), monitor, ProbeResult{OK: false, StatusCode: 503})
	require.NoError(t, err)

	assert.Equal(t, TransitionNone, outcome.Transition)
	assert.Equal(t, StatusUp, repo.state.Status, "state must not flap before the threshold")
	assert.EqualValues(t, 1, repo.state.ConsecutiveFailures)
	assert.EqualValues(t, 0, repo.state.ConsecutiveSuccesses)
	assert.Zero(t, repo.insertedEvent)
}

func TestRecorderOpensIncidentAtThreshold(t *testing.T) {
	repo := &fakeRecorderRepo{}
	recorder := NewRecorder(repo)
	monitor := healthyMonitor()
	monitor.ConsecutiveFailures = 2 // two failures already recorded

	outcome, err := recorder.Record(context.Background(), monitor, ProbeResult{OK: false, StatusCode: 502, Err: "status 502"})
	require.NoError(t, err)

	assert.Equal(t, TransitionDown, outcome.Transition)
	assert.EqualValues(t, 1, outcome.EventID)
	assert.Equal(t, StatusDown, repo.state.Status)
	assert.EqualValues(t, 3, repo.state.ConsecutiveFailures)
}

func TestRecorderResetsFailureStreakOnSuccess(t *testing.T) {
	repo := &fakeRecorderRepo{}
	recorder := NewRecorder(repo)
	monitor := healthyMonitor()
	monitor.ConsecutiveFailures = 2

	outcome, err := recorder.Record(context.Background(), monitor, ProbeResult{OK: true, StatusCode: 200})
	require.NoError(t, err)

	assert.Equal(t, TransitionNone, outcome.Transition)
	assert.Equal(t, StatusUp, repo.state.Status)
	assert.EqualValues(t, 0, repo.state.ConsecutiveFailures)
	assert.EqualValues(t, 1, repo.state.ConsecutiveSuccesses)
}

func TestRecorderRecoversAtRecoveryThreshold(t *testing.T) {
	started := time.Now().Add(-7 * time.Minute)
	repo := &fakeRecorderRepo{
		openEvent: queries.EnvironmentUptimeEvent{
			ID:        42,
			StartedAt: pgtype.Timestamp{Time: started, Valid: true},
		},
	}
	recorder := NewRecorder(repo)
	monitor := healthyMonitor()
	monitor.Status = StatusDown
	monitor.ConsecutiveSuccesses = 1 // one success already recorded

	outcome, err := recorder.Record(context.Background(), monitor, ProbeResult{OK: true, StatusCode: 200})
	require.NoError(t, err)

	assert.Equal(t, TransitionUp, outcome.Transition)
	assert.EqualValues(t, 42, outcome.EventID)
	assert.InDelta(t, (7 * time.Minute).Seconds(), outcome.DownFor.Seconds(), 1)
	assert.Equal(t, StatusUp, repo.state.Status)
	require.Len(t, repo.resolved, 1)
	assert.EqualValues(t, 7, repo.resolved[0].EnvironmentID)
}

func TestRecorderStaysDownBelowRecoveryThreshold(t *testing.T) {
	repo := &fakeRecorderRepo{}
	recorder := NewRecorder(repo)
	monitor := healthyMonitor()
	monitor.Status = StatusDown

	outcome, err := recorder.Record(context.Background(), monitor, ProbeResult{OK: true, StatusCode: 200})
	require.NoError(t, err)

	assert.Equal(t, TransitionNone, outcome.Transition)
	assert.Equal(t, StatusDown, repo.state.Status, "incident stays open until recovery threshold")
	assert.EqualValues(t, 1, repo.state.ConsecutiveSuccesses)
	assert.Empty(t, repo.resolved)
}

func TestRecorderRecoveryWithoutOpenEventRow(t *testing.T) {
	repo := &fakeRecorderRepo{openEventErr: pgx.ErrNoRows}
	recorder := NewRecorder(repo)
	monitor := healthyMonitor()
	monitor.Status = StatusDown
	monitor.ConsecutiveSuccesses = 1

	outcome, err := recorder.Record(context.Background(), monitor, ProbeResult{OK: true, StatusCode: 200})
	require.NoError(t, err)

	assert.Equal(t, TransitionUp, outcome.Transition)
	assert.Zero(t, outcome.DownFor)
	assert.Equal(t, StatusUp, repo.state.Status)
}

func TestRecorderUnknownBecomesUpOnSuccess(t *testing.T) {
	repo := &fakeRecorderRepo{}
	recorder := NewRecorder(repo)
	monitor := healthyMonitor()
	monitor.Status = StatusUnknown

	outcome, err := recorder.Record(context.Background(), monitor, ProbeResult{OK: true, StatusCode: 200})
	require.NoError(t, err)

	assert.Equal(t, TransitionNone, outcome.Transition)
	assert.Equal(t, StatusUp, repo.state.Status)
}

func TestRecorderUnknownCanGoDown(t *testing.T) {
	repo := &fakeRecorderRepo{}
	recorder := NewRecorder(repo)
	monitor := healthyMonitor()
	monitor.Status = StatusUnknown
	monitor.ConsecutiveFailures = 2

	outcome, err := recorder.Record(context.Background(), monitor, ProbeResult{OK: false, Err: "timeout"})
	require.NoError(t, err)

	assert.Equal(t, TransitionDown, outcome.Transition)
	assert.Equal(t, StatusDown, repo.state.Status)
}

func TestRecorderPropagatesInsertError(t *testing.T) {
	repo := &fakeRecorderRepo{insertErr: errors.New("boom")}
	recorder := NewRecorder(repo)
	monitor := healthyMonitor()
	monitor.ConsecutiveFailures = 2

	_, err := recorder.Record(context.Background(), monitor, ProbeResult{OK: false})
	require.Error(t, err)
	assert.Zero(t, repo.stateCalls, "state must not advance when the incident insert failed")
}

func TestRecorderPropagatesUpdateError(t *testing.T) {
	repo := &fakeRecorderRepo{updateErr: errors.New("boom")}
	recorder := NewRecorder(repo)

	_, err := recorder.Record(context.Background(), healthyMonitor(), ProbeResult{OK: true, StatusCode: 200})
	require.Error(t, err)
}
