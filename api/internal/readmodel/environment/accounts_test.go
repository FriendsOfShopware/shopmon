package environment

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
)

func ts(t time.Time) pgtype.Timestamp { return pgtype.Timestamp{Time: t, Valid: true} }

func TestPostureHints(t *testing.T) {
	now := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)
	old := now.Add(-200 * 24 * time.Hour)
	recent := now.Add(-2 * 24 * time.Hour)

	integration := func(label string, mutate func(*queries.EnvironmentAccount)) queries.EnvironmentAccount {
		row := queries.EnvironmentAccount{
			AccountType:     "integration",
			AccountID:       label,
			Label:           label,
			Admin:           true,
			RemoteCreatedAt: ts(old),
		}
		mutate(&row)
		return row
	}

	cases := []struct {
		name string
		rows []queries.EnvironmentAccount
		want []string
	}{
		{
			name: "an old administrator integration that was never used",
			rows: []queries.EnvironmentAccount{integration("Erp Connector", func(*queries.EnvironmentAccount) {})},
			want: []string{hintUnusedAdminIntegration},
		},
		{
			name: "an integration used recently is fine",
			rows: []queries.EnvironmentAccount{integration("Shopmon", func(r *queries.EnvironmentAccount) {
				r.LastUsageAt = ts(recent)
			})},
			want: nil,
		},
		{
			// It has simply not been used yet; flagging it would make every new
			// integration look like a problem.
			name: "a freshly created integration is not dormant",
			rows: []queries.EnvironmentAccount{integration("New", func(r *queries.EnvironmentAccount) {
				r.RemoteCreatedAt = ts(recent)
			})},
			want: nil,
		},
		{
			// Not every shop reports a creation date; without a fallback such
			// an integration would never qualify for the hint.
			name: "an integration without a creation date falls back to first seen",
			rows: []queries.EnvironmentAccount{integration("Legacy", func(r *queries.EnvironmentAccount) {
				r.RemoteCreatedAt = pgtype.Timestamp{}
				r.FirstSeenAt = ts(old)
			})},
			want: []string{hintUnusedAdminIntegration},
		},
		{
			name: "an integration first seen moments ago is not dormant",
			rows: []queries.EnvironmentAccount{integration("Fresh", func(r *queries.EnvironmentAccount) {
				r.RemoteCreatedAt = pgtype.Timestamp{}
				r.FirstSeenAt = ts(recent)
			})},
			want: nil,
		},
		{
			name: "a non-admin integration is not worth pointing at",
			rows: []queries.EnvironmentAccount{integration("Reader", func(r *queries.EnvironmentAccount) {
				r.Admin = false
			})},
			want: nil,
		},
		{
			// Its lifecycle belongs to the app, and "last used" says nothing
			// about whether the app itself is in use.
			name: "an app-managed integration is left to its app",
			rows: []queries.EnvironmentAccount{integration("SwagApp", func(r *queries.EnvironmentAccount) {
				r.ManagedByApp = true
			})},
			want: nil,
		},
		{
			name: "a user with its own API keys",
			rows: []queries.EnvironmentAccount{{
				AccountType: "user", AccountID: "u1", Label: "a***@e.de", AccessKeyCount: 2,
			}},
			want: []string{hintUserWithApiKeys},
		},
		{
			name: "a user with a pending password reset",
			rows: []queries.EnvironmentAccount{{
				AccountType: "user", AccountID: "u1", Label: "a***@e.de", PasswordRecoveryPending: true,
			}},
			want: []string{hintPasswordRecoveryPending},
		},
		{
			name: "both user hints can apply at once",
			rows: []queries.EnvironmentAccount{{
				AccountType: "user", AccountID: "u1", Label: "a***@e.de",
				AccessKeyCount: 1, PasswordRecoveryPending: true,
			}},
			want: []string{hintUserWithApiKeys, hintPasswordRecoveryPending},
		},
		{
			name: "an ordinary user produces no hint",
			rows: []queries.EnvironmentAccount{{AccountType: "user", AccountID: "u1", Label: "a***@e.de"}},
			want: nil,
		},
		{
			// Roles are rated through the changes made to them, not here.
			name: "permission roles are not rated",
			rows: []queries.EnvironmentAccount{{AccountType: "acl_role", AccountID: "r1", Label: "Editor"}},
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := postureHints("https://shop.example.com", tc.rows, now)

			hints := make([]string, 0, len(got))
			for _, h := range got {
				hints = append(hints, h.Hint)
			}
			assert.ElementsMatch(t, tc.want, hints)
		})
	}
}

func TestPostureHintsCarryLabelAndDeepLink(t *testing.T) {
	now := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)
	path := "/admin#/sw/integration/index"

	got := postureHints("https://shop.example.com/", []queries.EnvironmentAccount{{
		AccountType:     "integration",
		AccountID:       "i1",
		Label:           "Erp Connector",
		Admin:           true,
		RemoteCreatedAt: ts(now.Add(-200 * 24 * time.Hour)),
		AdminPath:       &path,
	}}, now)

	require.Len(t, got, 1)
	assert.Equal(t, "Erp Connector", got[0].Label)
	// The supporting facts travel as data; the sentence is rendered per viewer,
	// so a server-formatted "never used" would appear in English next to German.
	assert.Nil(t, got[0].LastUsageAt, "never used means no timestamp, not a string")
	require.NotNil(t, got[0].CreatedAt)
	require.NotNil(t, got[0].AdminUrl)
	assert.Equal(t, "https://shop.example.com/admin#/sw/integration/index", *got[0].AdminUrl)
}

func TestPostureHintReportsLastUsage(t *testing.T) {
	now := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)
	lastUsed := time.Date(2025, 3, 4, 0, 0, 0, 0, time.UTC)

	got := postureHints("https://shop.example.com", []queries.EnvironmentAccount{{
		AccountType:     "integration",
		AccountID:       "i1",
		Label:           "Erp Connector",
		Admin:           true,
		RemoteCreatedAt: ts(now.Add(-400 * 24 * time.Hour)),
		LastUsageAt:     ts(lastUsed),
	}}, now)

	require.Len(t, got, 1)
	require.NotNil(t, got[0].LastUsageAt)
	assert.True(t, lastUsed.Equal(*got[0].LastUsageAt))
}

func TestPostureHintReportsAccessKeyCount(t *testing.T) {
	got := postureHints("https://shop.example.com", []queries.EnvironmentAccount{{
		AccountType: "user", AccountID: "u1", Label: "a***@e.de", AccessKeyCount: 3,
	}}, time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC))

	require.Len(t, got, 1)
	require.NotNil(t, got[0].AccessKeyCount)
	assert.Equal(t, 3, *got[0].AccessKeyCount)
}

// A finding has to say what the account looks like now, so a reviewer can
// judge it without following the deep link into the shop.
func TestIndexSnapshots(t *testing.T) {
	rows := []queries.EnvironmentAccount{{
		AccountType:    "user",
		AccountID:      "u1",
		Label:          "a***@e.de",
		Active:         true,
		Admin:          true,
		RoleIds:        []byte(`["r1","r2"]`),
		AccessKeyCount: 2,
		FirstSeenAt:    ts(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
	}}

	got := indexSnapshots(rows)
	snap := got["user:u1"]
	require.NotNil(t, snap)
	assert.True(t, snap.Admin)
	assert.Equal(t, 2, snap.RoleCount, "role ids are counted, not exposed")
	assert.Equal(t, 2, snap.AccessKeyCount)
	assert.Nil(t, got["user:gone"], "a deleted account simply has no snapshot")
}

// The deep link is composed server-side because the stored path is relative and
// a raw href would otherwise resolve against Shopmon's own origin.
func TestAdminURL(t *testing.T) {
	path := "/admin#/x"
	empty := ""

	require.NotNil(t, adminURL("https://shop.example.com/", &path))
	assert.Equal(t, "https://shop.example.com/admin#/x", *adminURL("https://shop.example.com/", &path))
	assert.Nil(t, adminURL("https://shop.example.com", nil))
	assert.Nil(t, adminURL("https://shop.example.com", &empty))
	assert.Nil(t, adminURL("", &path))
}

// The message templates interpolate {label}, but the label lives in its own
// column. Without the merge every finding renders with an empty subject, which
// is the one thing a reviewer actually needs to see.
func TestRenderParamsAlwaysCarriesTheLabel(t *testing.T) {
	got := renderParams(nil, "m***@e***.com")
	require.NotNil(t, got)
	assert.Equal(t, "m***@e***.com", (*got)["label"])

	got = renderParams([]byte(`{}`), "a***@e.de")
	require.NotNil(t, got)
	assert.Equal(t, "a***@e.de", (*got)["label"])

	// Existing params survive alongside it.
	got = renderParams([]byte(`{"from":10,"to":20}`), "Editor")
	require.NotNil(t, got)
	assert.Equal(t, "Editor", (*got)["label"])
	assert.EqualValues(t, 10, (*got)["from"])
	assert.EqualValues(t, 20, (*got)["to"])

	// A malformed params column must not lose the label.
	got = renderParams([]byte(`not json`), "Fallback")
	require.NotNil(t, got)
	assert.Equal(t, "Fallback", (*got)["label"])
}

// Zero open findings means something different in each state, so the overview
// has to be able to tell them apart.
func TestAccountMonitoringStateFoldsInTheOptOut(t *testing.T) {
	assert.Equal(t, AccountStateOff, accountMonitoringState(false, AccountStateOK),
		"a disabled environment reports off whatever the stored state says")
	assert.Equal(t, AccountStateOff, accountMonitoringState(false, ""))

	assert.Equal(t, AccountStateOK, accountMonitoringState(true, AccountStateOK))
	assert.Equal(t, AccountStateError, accountMonitoringState(true, AccountStateError))
	assert.Equal(t, AccountStateUnsupported, accountMonitoringState(true, AccountStateUnsupported))

	// No stored row yet: the environment has not been scraped for accounts.
	assert.Equal(t, AccountStatePending, accountMonitoringState(true, ""))
}
