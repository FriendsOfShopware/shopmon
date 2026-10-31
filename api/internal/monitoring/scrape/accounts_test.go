package scrape

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	catalogsync "github.com/friendsofshopware/shopmon/api/internal/catalog/sync"
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/accounts"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/checker"
	"github.com/friendsofshopware/shopmon/api/internal/testutil/testdb"
)

func adminUser(id, label string) accounts.Account {
	return accounts.Account{
		Kind: accounts.KindUser, ID: id, Label: label,
		Active: true, Admin: true,
		AdminPath: "/admin#/sw/users/permissions/user/detail/" + id,
	}
}

func seededResult(acc ...accounts.Account) accountResult {
	return accountResult{supported: true, fetched: true, baselineRun: true, accounts: acc}
}

func diffedResult(baseline, current []accounts.Account) accountResult {
	return accountResult{
		supported: true, fetched: true,
		accounts: current,
		changes:  accounts.Diff(baseline, current),
	}
}

func accountRows(t *testing.T, h *Service, envID int32) []queries.EnvironmentAccount {
	t.Helper()
	rows, err := h.queries.ListEnvironmentAccounts(context.Background(), envID)
	require.NoError(t, err)
	return rows
}

func eventRows(t *testing.T, pool *pgxpool.Pool, envID int32) []struct {
	Change string
	Level  string
	Label  string
} {
	t.Helper()
	rows, err := pool.Query(context.Background(),
		`SELECT change, level, label FROM environment_account_event WHERE environment_id = $1 ORDER BY id`, envID)
	require.NoError(t, err)
	defer rows.Close()

	var out []struct {
		Change string
		Level  string
		Label  string
	}
	for rows.Next() {
		var e struct {
			Change string
			Level  string
			Label  string
		}
		require.NoError(t, rows.Scan(&e.Change, &e.Level, &e.Label))
		out = append(out, e)
	}
	require.NoError(t, rows.Err())
	return out
}

// The very first run must stay silent. Without a baseline every account reads
// as new, so alerting here would turn every existing shop red on rollout.
func TestPersistAccountsSeedsBaselineWithoutEvents(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()

	res := seededResult(adminUser("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "a***@e***.de"))
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, res))

	assert.Len(t, accountRows(t, h, env.ID), 1, "baseline written")
	assert.Empty(t, eventRows(t, pool, env.ID), "the seeding run must not report anything")

	state, err := h.queries.GetEnvironmentAccountState(ctx, env.ID)
	require.NoError(t, err)
	assert.True(t, state.Supported)
	assert.True(t, state.BaselineAt.Valid, "baseline_at marks the environment as initialised")
	assert.EqualValues(t, 1, state.AccountCount)
}

func TestPersistAccountsReportsNewAdminOnSecondRun(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()

	first := adminUser("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "a***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, seededResult(first)))

	second := adminUser("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "b***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{first}, []accounts.Account{first, second})))

	events := eventRows(t, pool, env.ID)
	require.Len(t, events, 1)
	assert.Equal(t, string(accounts.ChangeNewAdminUser), events[0].Change)
	assert.Equal(t, string(accounts.LevelRed), events[0].Level)
	assert.Equal(t, "b***@e***.de", events[0].Label, "the masked label travels with the finding")

	assert.Len(t, accountRows(t, h, env.ID), 2, "baseline advanced")

	open, err := h.queries.ListOpenEnvironmentAccountEvents(ctx, env.ID)
	require.NoError(t, err)
	assert.Len(t, open, 1, "the finding stays open until someone acknowledges it")
}

// Informational history must never show up in the checks tab or escalate the
// shop: a revoked permission is not something anyone has to confirm.
func TestPersistAccountsKeepsInfoEventsOutOfFindings(t *testing.T) {
	h, _, env := setupPersistTest(t)
	ctx := context.Background()

	before := adminUser("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "a***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, seededResult(before)))

	after := before
	after.Admin = false
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{before}, []accounts.Account{after})))

	open, err := h.queries.ListOpenEnvironmentAccountEvents(ctx, env.ID)
	require.NoError(t, err)
	assert.Empty(t, open, "revoking admin is history, not a finding")

	all, err := h.queries.ListEnvironmentAccountEvents(ctx, queries.ListEnvironmentAccountEventsParams{
		EnvironmentID: env.ID, Limit: 100,
	})
	require.NoError(t, err)
	require.Len(t, all, 1)
	assert.Equal(t, string(accounts.ChangeAdminRevoked), all[0].Change)
}

// An unreachable endpoint says nothing about whether the accounts changed.
// Dropping the baseline would make every account look new once it returns.
func TestPersistAccountsUnsupportedKeepsBaseline(t *testing.T) {
	h, _, env := setupPersistTest(t)
	ctx := context.Background()

	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		seededResult(adminUser("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "a***@e***.de"))))

	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, accountResult{supported: false}))

	assert.Len(t, accountRows(t, h, env.ID), 1, "baseline survives an unsupported run")
	state, err := h.queries.GetEnvironmentAccountState(ctx, env.ID)
	require.NoError(t, err)
	assert.False(t, state.Supported)
	assert.True(t, state.BaselineAt.Valid, "the environment stays initialised")
}

// A shop always has at least one administrator. An empty payload is a broken
// response, and treating it as data would wipe the baseline and then report
// every admin as new.
func TestPersistAccountsEmptyPayloadKeepsBaseline(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()

	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		seededResult(adminUser("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "a***@e***.de"))))

	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, accountResult{
		supported: true, fetched: false, err: errEmptyAccountPayload,
	}))

	assert.Len(t, accountRows(t, h, env.ID), 1, "baseline survives an empty payload")
	assert.Empty(t, eventRows(t, pool, env.ID), "nothing was deleted, so nothing is reported")

	state, err := h.queries.GetEnvironmentAccountState(ctx, env.ID)
	require.NoError(t, err)
	require.NotNil(t, state.LastError)
	assert.Contains(t, *state.LastError, "no entries")
}

func TestPersistAccountsPrunesVanishedAccount(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()

	stays := adminUser("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "a***@e***.de")
	goes := adminUser("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "b***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, seededResult(stays, goes)))

	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{stays, goes}, []accounts.Account{stays})))

	rows := accountRows(t, h, env.ID)
	require.Len(t, rows, 1)
	assert.Equal(t, stays.ID, rows[0].AccountID)

	events := eventRows(t, pool, env.ID)
	require.Len(t, events, 1)
	assert.Equal(t, string(accounts.ChangeDeleted), events[0].Change)
	assert.Equal(t, "b***@e***.de", events[0].Label, "the label outlives the account it describes")
}

// The hourly scrape must not rewrite unchanged rows, or every environment
// produces WAL every hour for nothing.
func TestPersistAccountsIdenticalRunRewritesNothing(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()

	acc := adminUser("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "a***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, seededResult(acc)))

	before := snapshotRowVersions(t, pool)
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{acc}, []accounts.Account{acc})))
	assert.Equal(t, before, snapshotRowVersions(t, pool), "unchanged accounts must not be rewritten")
}

// An open finding is derived from the database, so it has to survive a run in
// which the shop could not be reached at all. Otherwise restarting FroshTools
// would quietly clear a red finding nobody reviewed.
func TestOpenAccountChecksSurviveAFailedFetch(t *testing.T) {
	h, _, env := setupPersistTest(t)
	ctx := context.Background()

	first := adminUser("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "a***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, seededResult(first)))
	second := adminUser("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "b***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{first}, []accounts.Account{first, second})))

	// No client at all: the plugin was removed, or the route answered 404.
	res := h.fetchAccounts(ctx, env, nil, nil)
	require.Len(t, res.checks, 1, "the open finding is still reported")
	assert.Equal(t, checker.StatusRed, res.checks[0].Level)
	assert.Equal(t, checker.SourceAccounts, res.checks[0].Source)
	assert.Equal(t,
		"account.user.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.new_admin_user",
		res.checks[0].ID)
	assert.Equal(t,
		"https://shop.example.com/admin#/sw/users/permissions/user/detail/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		res.checks[0].Link, "the deep link is absolute, the stored path is not")
	assert.Equal(t, "b***@e***.de", res.checks[0].MessageParams["label"])
}

// Turning the feature off must not keep reporting findings nobody can reach a
// UI to clear.
func TestFetchAccountsDisabledReportsNothing(t *testing.T) {
	h, _, env := setupPersistTest(t)
	ctx := context.Background()

	first := adminUser("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "a***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, seededResult(first)))
	second := adminUser("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "b***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{first}, []accounts.Account{first, second})))

	env.AccountMonitoringEnabled = false
	res := h.fetchAccounts(ctx, env, nil, nil)
	assert.Empty(t, res.checks)
	assert.False(t, res.supported)
}

func TestAccountAdminURL(t *testing.T) {
	assert.Equal(t, "https://shop.example.com/admin#/x", accountAdminURL("https://shop.example.com/", "/admin#/x"))
	assert.Equal(t, "", accountAdminURL("https://shop.example.com", ""), "no path means no link")
	assert.Equal(t, "", accountAdminURL("", "/admin#/x"))
}

func TestMarkAccountEventsNotified(t *testing.T) {
	h, _, env := setupPersistTest(t)
	ctx := context.Background()

	first := adminUser("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "a***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, seededResult(first)))
	second := adminUser("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "b***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{first}, []accounts.Account{first, second})))

	pending, err := h.queries.ListUnnotifiedEnvironmentAccountEvents(ctx, env.ID)
	require.NoError(t, err)
	require.Len(t, pending, 1)

	h.markAccountEventsNotified(ctx, env.ID, pending)

	pending, err = h.queries.ListUnnotifiedEnvironmentAccountEvents(ctx, env.ID)
	require.NoError(t, err)
	assert.Empty(t, pending, "a notified event is not reconsidered next hour")
}

// failingMailer makes every email delivery fail, so the dispatch reports a
// failure the notifier has to react to.
type failingMailer struct{ err error }

func (m failingMailer) Send(context.Context, string, mail.Email) error { return m.err }
func (failingMailer) BuildConfirmationEmail(string, string) mail.Email { return mail.Email{} }
func (failingMailer) BuildPasswordResetEmail(string, string) mail.Email {
	return mail.Email{}
}
func (failingMailer) BuildOrgInviteEmail(string, string, string, string) mail.Email {
	return mail.Email{}
}
func (failingMailer) BuildStatusChangeEmail(string, string, string, string) mail.Email {
	return mail.Email{}
}
func (failingMailer) BuildConnectionFailedEmail(string, string, string) mail.Email {
	return mail.Email{}
}
func (failingMailer) BuildAlertEmail(string, string, string, string, *mail.AlertAction) mail.Email {
	return mail.Email{}
}
func (failingMailer) ProductURL() string { return "https://shopmon.example.com" }

// subscribe makes a user a notification recipient for an environment, which is
// the only way notifyAccountChanges has anyone to tell.
func subscribe(t *testing.T, pool *pgxpool.Pool, envID int32) {
	t.Helper()
	ctx := context.Background()
	_, err := pool.Exec(ctx, `
		INSERT INTO "user" (id, name, email, email_verified, created_at, updated_at)
		VALUES ('watcher', 'Watcher', 'watcher@example.com', true, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO member (id, organization_id, user_id, role, created_at)
		VALUES ('member-watcher', 'org-1', 'watcher', 'owner', NOW())
		ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO notification_preference (user_id, scope_type, scope_id, event_type, channel, enabled)
		VALUES ('watcher', 'environment', $1, '', '', true)
		ON CONFLICT DO NOTHING`, strconv.Itoa(int(envID)))
	require.NoError(t, err)
}

func seedOneFinding(t *testing.T, h *Service, envID int32) {
	t.Helper()
	ctx := context.Background()
	first := adminUser("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "a***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, envID, seededResult(first)))
	second := adminUser("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "b***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, envID,
		diffedResult([]accounts.Account{first}, []accounts.Account{first, second})))
}

func unnotifiedCount(t *testing.T, h *Service, envID int32) int {
	t.Helper()
	pending, err := h.queries.ListUnnotifiedEnvironmentAccountEvents(context.Background(), envID)
	require.NoError(t, err)
	return len(pending)
}

// A delivery failure must leave the event unnotified so the next scrape
// retries. Marking regardless would drop the alert for good.
func TestNotifyAccountChangesKeepsEventsForRetryOnFailure(t *testing.T) {
	pool := testdb.Setup(t)
	q := queries.New(pool)
	cfg := &config.Config{}
	h := NewService(pool, q, cfg, failingMailer{err: errors.New("smtp unreachable")},
		catalogsync.NewService(pool, q, cfg), nil)

	envID := seedEnvironment(t, pool, "Production", "6.5.0.0")
	env := reloadEnv(t, h, envID)
	subscribe(t, pool, envID)
	seedOneFinding(t, h, envID)
	require.Equal(t, 1, unnotifiedCount(t, h, envID))

	h.notifyAccountChanges(context.Background(), env)

	assert.Equal(t, 1, unnotifiedCount(t, h, envID),
		"a failed delivery must leave the event pending for the next scrape")
}

// Informational history is not worth an email; only findings are.
func TestNotifyAccountChangesIgnoresInfoEvents(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()
	subscribe(t, pool, env.ID)

	before := adminUser("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "a***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, seededResult(before)))
	after := before
	after.Admin = false
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{before}, []accounts.Account{after})))

	assert.Equal(t, 0, unnotifiedCount(t, h, env.ID),
		"revoking admin is history, not something to mail about")
}

// With monitoring switched off nothing is sent, and the pending events are not
// silently consumed either.
func TestNotifyAccountChangesSkippedWhenDisabled(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	subscribe(t, pool, env.ID)
	seedOneFinding(t, h, env.ID)

	env.AccountMonitoringEnabled = false
	h.notifyAccountChanges(context.Background(), env)

	assert.Equal(t, 1, unnotifiedCount(t, h, env.ID))
}

// The counterpart to the failure case: a dispatch that delivers marks the
// event, so the next scrape does not alert about it again.
func TestNotifyAccountChangesMarksOnSuccess(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	subscribe(t, pool, env.ID)
	seedOneFinding(t, h, env.ID)
	require.Equal(t, 1, unnotifiedCount(t, h, env.ID))

	h.notifyAccountChanges(context.Background(), env)

	assert.Equal(t, 0, unnotifiedCount(t, h, env.ID))

	// And the recipient actually has something in their inbox.
	var inbox int
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM user_notification WHERE user_id = 'watcher'`).Scan(&inbox))
	assert.Equal(t, 1, inbox)

	// A second run has nothing left to send.
	h.notifyAccountChanges(context.Background(), env)
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM user_notification WHERE user_id = 'watcher'`).Scan(&inbox))
	assert.Equal(t, 1, inbox, "an already notified finding must not be sent again")
}

// Nothing serialises scrapes of one environment: the hourly run and a manual
// refresh can overlap, diff the same baseline and reach for the same event.
// Recording it twice would show the finding twice and keep the environment red
// until each copy was acknowledged separately.
func TestPersistAccountsRecordsOneOpenEventPerChange(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()

	first := adminUser("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "a***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, seededResult(first)))

	second := adminUser("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "b***@e***.de")
	overlapping := diffedResult([]accounts.Account{first}, []accounts.Account{first, second})

	// Both runs see the pre-change baseline, exactly as two concurrent scrapes
	// would.
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, overlapping))
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, overlapping))

	assert.Len(t, eventRows(t, pool, env.ID), 1, "the same change must be recorded once")

	open, err := h.queries.ListOpenEnvironmentAccountEvents(ctx, env.ID)
	require.NoError(t, err)
	assert.Len(t, open, 1, "one finding to acknowledge, not two")
}

// Acknowledging closes the finding, so the same change happening again later is
// a new event rather than being swallowed by the guard above.
func TestPersistAccountsRecordsRecurrenceAfterAcknowledgement(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()

	before := adminUser("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "a***@e***.de")
	before.Admin = false
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, seededResult(before)))

	granted := before
	granted.Admin = true
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{before}, []accounts.Account{granted})))

	open, err := h.queries.ListOpenEnvironmentAccountEvents(ctx, env.ID)
	require.NoError(t, err)
	require.Len(t, open, 1)

	rows, err := h.queries.AcknowledgeEnvironmentAccountEvent(ctx, queries.AcknowledgeEnvironmentAccountEventParams{
		ID:            open[0].ID,
		EnvironmentID: env.ID,
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, rows)

	// Rights revoked, then granted again.
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{granted}, []accounts.Account{before})))
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{before}, []accounts.Account{granted})))

	open, err = h.queries.ListOpenEnvironmentAccountEvents(ctx, env.ID)
	require.NoError(t, err)
	require.Len(t, open, 1, "the recurrence opens a fresh finding")

	var granteds int
	for _, e := range eventRows(t, pool, env.ID) {
		if e.Change == string(accounts.ChangeAdminGranted) {
			granteds++
		}
	}
	assert.Equal(t, 2, granteds, "both grants stay in the history")
}

// An unreadable findings table says nothing about whether the findings were
// resolved. Treating it as "none" would prune the stored account checks and
// mail out a recovery for findings that are still open.
func TestOpenAccountChecksReportsUnevaluatedOnReadFailure(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()

	first := adminUser("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "a***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, seededResult(first)))
	second := adminUser("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "b***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{first}, []accounts.Account{first, second})))

	checks, unevaluated := h.openAccountChecks(ctx, env)
	require.False(t, unevaluated)
	require.Len(t, checks, 1)

	// Break the read the way a migration mishap or a permissions problem would.
	_, err := pool.Exec(ctx, `ALTER TABLE environment_account_event RENAME TO environment_account_event_moved`)
	require.NoError(t, err)
	defer func() {
		_, err := pool.Exec(context.Background(),
			`ALTER TABLE environment_account_event_moved RENAME TO environment_account_event`)
		require.NoError(t, err)
	}()

	checks, unevaluated = h.openAccountChecks(ctx, env)
	assert.True(t, unevaluated, "a failed read must be reported, not silently become zero findings")
	assert.Empty(t, checks)

	// fetchAccounts must pass the flag on so the scrape carries the stored
	// checks forward.
	res := h.fetchAccounts(ctx, env, nil, nil)
	assert.True(t, res.unevaluated)
}

// accountsUnavailable has to name the group precisely, or carryOverChecks will
// either miss the account checks or claim checks belonging to another source.
func TestAccountsUnavailableOwnsOnlyAccountChecks(t *testing.T) {
	u := accountsUnavailable()
	require.Len(t, u, 1)

	assert.True(t, u[0].Owns(checker.SourceAccounts, "account.user.abc.new_admin_user"))
	assert.False(t, u[0].Owns(checker.SourceSecurity, "security.CVE-2026-1"))
	assert.False(t, u[0].Owns(checker.SourceFroshTools, "frosh.cacheWarning"))
}

// A transient subscriber lookup failure must not mark the events notified: that
// would drop these alerts permanently instead of retrying next scrape.
func TestNotifyAccountChangesKeepsEventsWhenSubscriberLookupFails(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()
	subscribe(t, pool, env.ID)
	seedOneFinding(t, h, env.ID)
	require.Equal(t, 1, unnotifiedCount(t, h, env.ID))

	_, err := pool.Exec(ctx, `ALTER TABLE notification_preference RENAME TO notification_preference_moved`)
	require.NoError(t, err)
	defer func() {
		_, err := pool.Exec(context.Background(),
			`ALTER TABLE notification_preference_moved RENAME TO notification_preference`)
		require.NoError(t, err)
	}()

	h.notifyAccountChanges(ctx, env)

	assert.Equal(t, 1, unnotifiedCount(t, h, env.ID),
		"a lookup failure must leave the alert pending, not consume it")
}

// An informational row must not occupy the one-open-finding slot. Nothing ever
// acknowledges informational rows, so one would hold the key forever -- and the
// same change arriving later as a real finding, once the role was actually
// assigned to somebody, would be dropped without a check or an alert.
func TestPersistAccountsInfoEventDoesNotBlockLaterFinding(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()

	const roleID = "cccccccccccccccccccccccccccccccc"
	role := func(count int32, hash string) accounts.Account {
		return accounts.Account{
			Kind: accounts.KindACLRole, ID: roleID, Label: "Editor",
			Active: true, PrivilegeCount: count, PrivilegesHash: hash,
		}
	}
	holder := func(roles ...string) accounts.Account {
		a := accounts.Account{Kind: accounts.KindUser, ID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Label: "a***@e.de", Active: true}
		a.RoleIDs = roles
		return a
	}

	// Nobody holds the role yet, so growing it grants nothing: informational.
	unheld, unheldGrown := role(10, "h1"), role(20, "h2")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, seededResult(unheld, holder())))
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{unheld, holder()}, []accounts.Account{unheldGrown, holder()})))

	open, err := h.queries.ListOpenEnvironmentAccountEvents(ctx, env.ID)
	require.NoError(t, err)
	require.Empty(t, open, "an unassigned role grants nothing")

	// The role is assigned and grows again -- now it does widen access.
	held, heldGrown := holder(roleID), holder(roleID)
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{unheldGrown, holder()}, []accounts.Account{unheldGrown, held})))
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{unheldGrown, heldGrown}, []accounts.Account{role(30, "h3"), heldGrown})))

	open, err = h.queries.ListOpenEnvironmentAccountEvents(ctx, env.ID)
	require.NoError(t, err)

	var grown bool
	for _, e := range open {
		if e.Change == string(accounts.ChangeRolePrivilegesGrown) {
			grown = true
			assert.Equal(t, string(accounts.LevelYellow), e.Level)
		}
	}
	assert.True(t, grown, "the earlier informational row must not have swallowed this finding")

	var infos int
	for _, e := range eventRows(t, pool, env.ID) {
		if e.Level == string(accounts.LevelInfo) && e.Change == string(accounts.ChangeRolePrivilegesGrown) {
			infos++
		}
	}
	assert.Equal(t, 1, infos, "the informational row is still in the history")
}

// A failed read of the stored findings does not stop this run from detecting
// new ones. Dropping the freshly detected checks would persist the event and
// send its alert while leaving the environment status unchanged.
func TestFetchAccountsKeepsFreshChecksWhenStoredOnesAreUnreadable(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()

	first := adminUser("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "a***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, seededResult(first)))

	second := adminUser("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "b***@e***.de")
	fetched := &accounts.Snapshot{Accounts: []accounts.Account{first, second}}

	// The open-findings read fails while the shop still answers.
	_, err := pool.Exec(ctx, `ALTER TABLE environment_account_event RENAME TO environment_account_event_moved`)
	require.NoError(t, err)
	defer func() {
		_, err := pool.Exec(context.Background(),
			`ALTER TABLE environment_account_event_moved RENAME TO environment_account_event`)
		require.NoError(t, err)
	}()

	res := h.fetchAccounts(ctx, env, snapshotFetcher{snapshot: fetched}, froshToolsExtensions())
	require.True(t, res.unevaluated, "the stored findings could not be read")
	require.NotEmpty(t, res.checks, "the newly detected admin must still produce a check")

	var ids []string
	for _, c := range res.checks {
		ids = append(ids, c.ID)
	}
	assert.Contains(t, ids,
		"account.user.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.new_admin_user")
}

// snapshotFetcher serves a prepared snapshot in place of the shop.
type snapshotFetcher struct{ snapshot *accounts.Snapshot }

func (f snapshotFetcher) Get(context.Context, string) ([]byte, error) {
	users := make([]map[string]any, 0, len(f.snapshot.Accounts))
	for _, a := range f.snapshot.Accounts {
		if a.Kind != accounts.KindUser {
			continue
		}
		users = append(users, map[string]any{
			"id": a.ID, "maskedEmail": a.Label, "active": a.Active, "admin": a.Admin,
		})
	}
	return json.Marshal(map[string]any{"users": users})
}

func froshToolsExtensions() []extensionEntry {
	return []extensionEntry{{Name: froshToolsExtension, Active: true, Installed: true}}
}

// A failed alert leaves its events pending for the next scrape. If a reviewer
// confirms one in the meantime, re-alerting about it is pure noise.
func TestNotifyAccountChangesSkipsAcknowledgedEvents(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()
	subscribe(t, pool, env.ID)
	seedOneFinding(t, h, env.ID)

	pending, err := h.queries.ListUnnotifiedEnvironmentAccountEvents(ctx, env.ID)
	require.NoError(t, err)
	require.Len(t, pending, 1)

	rows, err := h.queries.AcknowledgeEnvironmentAccountEvent(ctx, queries.AcknowledgeEnvironmentAccountEventParams{
		ID:            pending[0].ID,
		EnvironmentID: env.ID,
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, rows)

	pending, err = h.queries.ListUnnotifiedEnvironmentAccountEvents(ctx, env.ID)
	require.NoError(t, err)
	assert.Empty(t, pending, "a confirmed finding is not worth alerting about")
}

// The same change can genuinely happen twice before anyone reviews it. Keeping
// the first row untouched would leave the open finding describing a state the
// shop has already moved past, with no alert for the newer change.
func TestPersistAccountsRefreshesAnUnreviewedFinding(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()

	const roleID = "cccccccccccccccccccccccccccccccc"
	role := func(count int32, hash string) accounts.Account {
		return accounts.Account{
			Kind: accounts.KindACLRole, ID: roleID, Label: "Editor",
			Active: true, PrivilegeCount: count, PrivilegesHash: hash,
		}
	}
	holder := accounts.Account{
		Kind: accounts.KindUser, ID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Label: "a***@e.de", Active: true,
	}
	holder.RoleIDs = []string{roleID}

	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, seededResult(role(10, "h1"), holder)))

	// Grows once: a finding opens and is marked notified.
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{role(10, "h1"), holder}, []accounts.Account{role(20, "h2"), holder})))
	pending, err := h.queries.ListUnnotifiedEnvironmentAccountEvents(ctx, env.ID)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	h.markAccountEventsNotified(ctx, env.ID, pending)

	// Grows again before anyone reviewed it.
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{role(20, "h2"), holder}, []accounts.Account{role(30, "h3"), holder})))

	open, err := h.queries.ListOpenEnvironmentAccountEvents(ctx, env.ID)
	require.NoError(t, err)
	require.Len(t, open, 1, "still one finding to review, not two")

	var params map[string]any
	require.NoError(t, json.Unmarshal(open[0].Params, &params))
	assert.EqualValues(t, 30, params["to"], "the finding describes the current state, not the stale one")

	pending, err = h.queries.ListUnnotifiedEnvironmentAccountEvents(ctx, env.ID)
	require.NoError(t, err)
	assert.Len(t, pending, 1, "the newer growth is alerted again")

	// An identical re-detection, as two overlapping scrapes produce, must stay
	// a no-op so the same finding is not mailed twice.
	h.markAccountEventsNotified(ctx, env.ID, pending)
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{role(20, "h2"), holder}, []accounts.Account{role(30, "h3"), holder})))
	pending, err = h.queries.ListUnnotifiedEnvironmentAccountEvents(ctx, env.ID)
	require.NoError(t, err)
	assert.Empty(t, pending, "an identical re-detection must not re-alert")
	assert.Len(t, eventRows(t, pool, env.ID), 1)
}

// Switching monitoring off acknowledges the findings that exist at that moment.
// A scrape already in flight must not write new unacknowledged ones afterwards:
// the next scrape no longer emits their checks, so they would sit there with
// nothing to clear them.
func TestPersistAccountsSkipsEventsWhenMonitoringWasDisabledMidScrape(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()

	first := adminUser("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "a***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID, seededResult(first)))

	// The toggle lands while the scrape is between fetch and persist.
	_, err := pool.Exec(ctx,
		`UPDATE environment SET account_monitoring_enabled = false WHERE id = $1`, env.ID)
	require.NoError(t, err)

	second := adminUser("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "b***@e***.de")
	require.NoError(t, persistAccounts(ctx, h.queries, env.ID,
		diffedResult([]accounts.Account{first}, []accounts.Account{first, second})))

	assert.Empty(t, eventRows(t, pool, env.ID), "no stranded finding")
	// The baseline still advances, so re-enabling later does not replay it.
	assert.Len(t, accountRows(t, h, env.ID), 2)
}

// Refreshing an unreviewed finding keeps its row id, so an id-only dedup key
// would collide with the delivery an hour earlier: the dispatcher would skip
// the email as deduped and the newer change would be marked notified without
// anyone being told.
func TestAccountAlertDedupKeyFollowsTheRevision(t *testing.T) {
	first := []queries.EnvironmentAccountEvent{{ID: 7, Revision: 1}}
	refreshed := []queries.EnvironmentAccountEvent{{ID: 7, Revision: 2}}

	assert.NotEqual(t, accountAlertDedupKey(1, first), accountAlertDedupKey(1, refreshed),
		"a refreshed finding must not reuse the previous delivery's key")

	// An identical set keys the same, which is what keeps a retry from mailing
	// the same findings twice.
	assert.Equal(t, accountAlertDedupKey(1, first), accountAlertDedupKey(1, first))

	// A change to any member counts, not just the newest row.
	pair := []queries.EnvironmentAccountEvent{{ID: 7, Revision: 1}, {ID: 9, Revision: 1}}
	olderBumped := []queries.EnvironmentAccountEvent{{ID: 7, Revision: 2}, {ID: 9, Revision: 1}}
	assert.NotEqual(t, accountAlertDedupKey(1, pair), accountAlertDedupKey(1, olderBumped))

	// Environments do not share a key.
	assert.NotEqual(t, accountAlertDedupKey(1, first), accountAlertDedupKey(2, first))
}

// A refresh that lands after the events were read carries a higher revision;
// marking the older delivery must not bury it.
func TestMarkAccountEventsNotifiedIsGuardedByRevision(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()
	seedOneFinding(t, h, env.ID)

	pending, err := h.queries.ListUnnotifiedEnvironmentAccountEvents(ctx, env.ID)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.EqualValues(t, 1, pending[0].Revision)

	// The finding is refreshed while the notification is in flight.
	_, err = pool.Exec(ctx,
		`UPDATE environment_account_event SET revision = revision + 1 WHERE id = $1`, pending[0].ID)
	require.NoError(t, err)

	h.markAccountEventsNotified(ctx, env.ID, pending)

	still, err := h.queries.ListUnnotifiedEnvironmentAccountEvents(ctx, env.ID)
	require.NoError(t, err)
	assert.Len(t, still, 1, "the newer revision stays pending so it is still alerted")
}
