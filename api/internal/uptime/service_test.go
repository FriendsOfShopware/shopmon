package uptime

import (
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/testutil/testdb"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedUptimeFixture creates a user, an organization with membership, a shop
// and an environment, returning the user and environment IDs.
func seedUptimeFixture(t *testing.T, pool *pgxpool.Pool) (string, int32) {
	t.Helper()
	ctx := t.Context()

	_, err := pool.Exec(ctx, `
		INSERT INTO "user" (id, name, email, created_at, updated_at) VALUES ('user-1', 'Test User', 'user-1@example.com', NOW(), NOW());
		INSERT INTO "user" (id, name, email, created_at, updated_at) VALUES ('stranger', 'Stranger', 'stranger@example.com', NOW(), NOW());
		INSERT INTO organization (id, name, slug, created_at) VALUES ('org-1', 'Test Org', 'test-org', NOW());
		INSERT INTO member (id, organization_id, user_id, created_at) VALUES ('member-1', 'org-1', 'user-1', NOW());
		INSERT INTO shop (organization_id, name, created_at, updated_at) VALUES ('org-1', 'Test Shop', NOW(), NOW());
	`)
	require.NoError(t, err)

	var envID int32
	err = pool.QueryRow(ctx, `
		INSERT INTO environment (organization_id, shop_id, name, status, url, client_id, client_secret, shopware_version, environment_token, created_at)
		VALUES ('org-1', (SELECT id FROM shop LIMIT 1), 'Production', 'green', 'https://shop.example.com', 'client', 'secret', '6.6.0.0', 'token', NOW())
		RETURNING id
	`).Scan(&envID)
	require.NoError(t, err)

	return "user-1", envID
}

func TestUpdateSettingsValidation(t *testing.T) {
	pool := testdb.Setup(t)
	q := queries.New(pool)
	service := NewService(q)
	userID, envID := seedUptimeFixture(t, pool)

	valid := Settings{Enabled: true, IntervalSeconds: 60, ExpectedStatus: 0, FailureThreshold: 3, RecoveryThreshold: 2}

	tests := []struct {
		name    string
		mutate  func(*Settings)
		wantErr bool
	}{
		{name: "valid", mutate: func(*Settings) {}},
		{name: "interval too low", mutate: func(s *Settings) { s.IntervalSeconds = 5 }, wantErr: true},
		{name: "interval too high", mutate: func(s *Settings) { s.IntervalSeconds = 9999 }, wantErr: true},
		{name: "bad expected status", mutate: func(s *Settings) { s.ExpectedStatus = 42 }, wantErr: true},
		{name: "zero failure threshold", mutate: func(s *Settings) { s.FailureThreshold = 0 }, wantErr: true},
		{name: "threshold too high", mutate: func(s *Settings) { s.RecoveryThreshold = 11 }, wantErr: true},
		{name: "invalid url", mutate: func(s *Settings) { s.URL = "not a url" }, wantErr: true},
		{name: "valid override url", mutate: func(s *Settings) { s.URL = "https://status.example.com/health" }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			settings := valid
			test.mutate(&settings)
			err := service.UpdateSettings(t.Context(), userID, envID, settings)
			if test.wantErr {
				assert.ErrorIs(t, err, ErrInvalidSettings)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateSettingsAuthorization(t *testing.T) {
	pool := testdb.Setup(t)
	q := queries.New(pool)
	service := NewService(q)
	_, envID := seedUptimeFixture(t, pool)

	settings := Settings{Enabled: true, IntervalSeconds: 60, FailureThreshold: 3, RecoveryThreshold: 2}

	// Non-member of the organization.
	err := service.UpdateSettings(t.Context(), "stranger", envID, settings)
	assert.ErrorIs(t, err, ErrNotAuthorized)

	// Unknown environment.
	err = service.UpdateSettings(t.Context(), "user-1", 999999, settings)
	assert.ErrorIs(t, err, ErrEnvironmentNotFound)
}

func TestUpdateSettingsDisableResolvesOpenIncident(t *testing.T) {
	pool := testdb.Setup(t)
	q := queries.New(pool)
	service := NewService(q)
	userID, envID := seedUptimeFixture(t, pool)

	enable := Settings{Enabled: true, IntervalSeconds: 60, FailureThreshold: 3, RecoveryThreshold: 2}
	require.NoError(t, service.UpdateSettings(t.Context(), userID, envID, enable))

	// Open an incident directly.
	_, err := q.InsertUptimeEvent(t.Context(), queries.InsertUptimeEventParams{
		EnvironmentID: envID,
		StartedAt:     pgTimestamp(time.Now().Add(-time.Hour)),
	})
	require.NoError(t, err)

	// Disabling the monitor resolves it.
	disable := enable
	disable.Enabled = false
	require.NoError(t, service.UpdateSettings(t.Context(), userID, envID, disable))

	_, err = q.GetOpenUptimeEvent(t.Context(), envID)
	require.Error(t, err, "no open incident should remain")
}

func TestAggregateDayComputesAvailability(t *testing.T) {
	pool := testdb.Setup(t)
	q := queries.New(pool)
	service := NewService(q)
	_, envID := seedUptimeFixture(t, pool)

	day := startOfDay(time.Now().UTC()).AddDate(0, 0, -1)

	// One 1-hour incident from 10:00 to 11:00 (3600s down).
	_, err := q.InsertUptimeEvent(t.Context(), queries.InsertUptimeEventParams{
		EnvironmentID: envID,
		StartedAt:     pgTimestamp(day.Add(10 * time.Hour)),
	})
	require.NoError(t, err)
	err = q.ResolveUptimeEvents(t.Context(), queries.ResolveUptimeEventsParams{
		EnvironmentID: envID,
		ResolvedAt:    pgTimestamp(day.Add(11 * time.Hour)),
	})
	require.NoError(t, err)

	// Hourly rollups for every hour of the day so no hour counts as no-data.
	avg, p95 := int32(120), int32(300)
	for h := 0; h < 24; h++ {
		require.NoError(t, q.UpsertUptimeHourlyRollup(t.Context(), queries.UpsertUptimeHourlyRollupParams{
			EnvironmentID:  envID,
			Hour:           pgTimestamp(day.Add(time.Duration(h) * time.Hour)),
			ProbesExpected: 60,
			ProbesRun:      60,
			ProbesOk:       55,
			LatencyAvgMs:   &avg,
			LatencyP95Ms:   &p95,
		}))
	}

	require.NoError(t, service.AggregateDay(t.Context(), day))

	rollups, err := q.ListUptimeDailyRollups(t.Context(), queries.ListUptimeDailyRollupsParams{
		EnvironmentID: envID,
		Day:           pgDate(day),
		Day_2:         pgDate(day),
	})
	require.NoError(t, err)
	require.Len(t, rollups, 1)

	row := rollups[0]
	assert.EqualValues(t, 3600, row.DownSeconds)
	assert.EqualValues(t, 24*3600-3600, row.UpSeconds)
	assert.EqualValues(t, 0, row.NodataSeconds)
	require.NotNil(t, row.LatencyAvgMs)
	assert.EqualValues(t, 120, *row.LatencyAvgMs)
	require.NotNil(t, row.LatencyP95Ms)
	assert.EqualValues(t, 300, *row.LatencyP95Ms)
}

func TestAggregateDayCountsMissingHoursAsNoData(t *testing.T) {
	pool := testdb.Setup(t)
	q := queries.New(pool)
	service := NewService(q)
	_, envID := seedUptimeFixture(t, pool)

	day := startOfDay(time.Now().UTC()).AddDate(0, 0, -1)

	// Only one hour has probes; the other 23 hours are no-data.
	require.NoError(t, q.UpsertUptimeHourlyRollup(t.Context(), queries.UpsertUptimeHourlyRollupParams{
		EnvironmentID:  envID,
		Hour:           pgTimestamp(day.Add(5 * time.Hour)),
		ProbesExpected: 60,
		ProbesRun:      60,
		ProbesOk:       60,
	}))

	require.NoError(t, service.AggregateDay(t.Context(), day))

	rollups, err := q.ListUptimeDailyRollups(t.Context(), queries.ListUptimeDailyRollupsParams{
		EnvironmentID: envID,
		Day:           pgDate(day),
		Day_2:         pgDate(day),
	})
	require.NoError(t, err)
	require.Len(t, rollups, 1)

	row := rollups[0]
	assert.EqualValues(t, 23*3600, row.NodataSeconds)
	assert.EqualValues(t, 3600, row.UpSeconds)
	assert.EqualValues(t, 0, row.DownSeconds)
}

func TestGetViewAvailability(t *testing.T) {
	pool := testdb.Setup(t)
	q := queries.New(pool)
	service := NewService(q)
	userID, envID := seedUptimeFixture(t, pool)

	require.NoError(t, service.UpdateSettings(t.Context(), userID, envID, Settings{
		Enabled: true, IntervalSeconds: 60, FailureThreshold: 3, RecoveryThreshold: 2,
	}))

	// A resolved 1-hour incident fully inside the last 7 days.
	started := time.Now().UTC().Add(-48 * time.Hour)
	_, err := q.InsertUptimeEvent(t.Context(), queries.InsertUptimeEventParams{
		EnvironmentID: envID,
		StartedAt:     pgTimestamp(started),
	})
	require.NoError(t, err)
	err = q.ResolveUptimeEvents(t.Context(), queries.ResolveUptimeEventsParams{
		EnvironmentID: envID,
		ResolvedAt:    pgTimestamp(started.Add(time.Hour)),
	})
	require.NoError(t, err)

	view, err := service.GetView(t.Context(), userID, envID, "7d")
	require.NoError(t, err)

	assert.True(t, view.Settings.Enabled)
	assert.Equal(t, StatusUnknown, view.Settings.Status)
	require.Len(t, view.Incidents, 1)
	assert.Nil(t, view.OpenIncident)
	assert.Len(t, view.Days, 7)
	require.NotNil(t, view.Availability)
	// Only the incident day and today carry data, so the denominator is those
	// two days; one hour of downtime must reduce availability measurably.
	assert.Less(t, *view.Availability, 1.0, "one hour of downtime must reduce availability")
	assert.Greater(t, *view.Availability, 0.9)
}

func TestGetViewReportsOpenIncident(t *testing.T) {
	pool := testdb.Setup(t)
	q := queries.New(pool)
	service := NewService(q)
	userID, envID := seedUptimeFixture(t, pool)

	require.NoError(t, service.UpdateSettings(t.Context(), userID, envID, Settings{
		Enabled: true, IntervalSeconds: 60, FailureThreshold: 3, RecoveryThreshold: 2,
	}))

	_, err := q.InsertUptimeEvent(t.Context(), queries.InsertUptimeEventParams{
		EnvironmentID: envID,
		StartedAt:     pgTimestamp(time.Now().UTC().Add(-30 * time.Minute)),
	})
	require.NoError(t, err)

	view, err := service.GetView(t.Context(), userID, envID, "24h")
	require.NoError(t, err)

	require.NotNil(t, view.OpenIncident)
	assert.Nil(t, view.OpenIncident.ResolvedAt)
	assert.GreaterOrEqual(t, view.OpenIncident.DurationSeconds, int64(30*60))
}

func TestGetViewRejectsUnknownRange(t *testing.T) {
	pool := testdb.Setup(t)
	q := queries.New(pool)
	service := NewService(q)
	userID, envID := seedUptimeFixture(t, pool)

	_, err := service.GetView(t.Context(), userID, envID, "100y")
	assert.Error(t, err)
}

func TestPauseAndResumeMonitor(t *testing.T) {
	pool := testdb.Setup(t)
	q := queries.New(pool)
	service := NewService(q)
	userID, envID := seedUptimeFixture(t, pool)

	require.NoError(t, service.UpdateSettings(t.Context(), userID, envID, Settings{
		Enabled: true, IntervalSeconds: 60, FailureThreshold: 3, RecoveryThreshold: 2,
	}))

	require.NoError(t, service.PauseMonitor(t.Context(), envID))
	state, err := q.GetUptimeState(t.Context(), envID)
	require.NoError(t, err)
	assert.Equal(t, StatusPaused, state.Status)

	require.NoError(t, service.ResumeMonitor(t.Context(), envID))
	state, err = q.GetUptimeState(t.Context(), envID)
	require.NoError(t, err)
	assert.Equal(t, StatusUnknown, state.Status)
}
