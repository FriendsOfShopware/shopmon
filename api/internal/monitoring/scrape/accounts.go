package scrape

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/otel/attribute"

	"github.com/friendsofshopware/shopmon/api/internal/database"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/notify"
	"github.com/friendsofshopware/shopmon/api/internal/ptr"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/accounts"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/checker"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/sanitize"
)

// errEmptyAccountPayload marks a response that parsed but carried no accounts
// at all. A Shopware shop always has at least one administrator, so this is a
// broken proxy or a partial deployment — never a shop that genuinely lost
// every account. Treating it as data would diff every account as deleted, wipe
// the baseline, and make every admin look brand new (and the shop turn red) on
// the next successful run.
const errEmptyAccountPayload = "accounts payload contained no entries"

// accountResult carries what a scrape learned about an environment's accounts.
//
// checks is populated on every path, including the ones that never reach the
// shop: findings are derived from the stored events, so a 404, a revoked ACL
// or a deinstalled plugin must not silently resolve an open security finding.
type accountResult struct {
	supported   bool
	fetched     bool
	baselineRun bool
	accounts    []accounts.Account
	changes     []accounts.Change
	checks      []checker.Check
	generatedAt pgtype.Timestamp
	err         string
	// fetchErr is the typed upstream error behind a failed fetch, kept so the
	// scrape can classify it for the owner (missing privilege, timeout, …)
	// instead of only logging err.
	fetchErr error
}

// fetchAccounts reads the shop's accounts, diffs them against the stored
// baseline and derives the checks for every finding that is still open.
//
// Every failure mode is non-fatal: an environment whose accounts cannot be read
// still completes its scrape.
func (h *Service) fetchAccounts(ctx context.Context, env queries.GetAllEnvironmentsRow, client accounts.Fetcher, extensions []extensionEntry) accountResult {
	if !env.AccountMonitoringEnabled {
		// Switching the feature off acknowledges the open findings, so there is
		// nothing left to re-emit here.
		return accountResult{}
	}

	ctx, span := tracer.Start(ctx, "environment.scrape.accounts")
	defer span.End()

	res := accountResult{checks: h.openAccountChecks(ctx, env)}

	if !hasActiveFroshTools(extensions) || client == nil {
		return res
	}

	snapshot, err := accounts.Fetch(ctx, client)
	if err != nil {
		if unsupported, ok := errors.AsType[*accounts.ErrUnsupported](err); ok {
			// Expected for FroshTools releases without the account route.
			slog.DebugContext(ctx, "accounts endpoint unavailable",
				"environmentId", env.ID, "reason", unsupported.Reason)
			return res
		}
		// Everything else — a 403 from a missing frosh_tools:read ACL above
		// all — is a fixable configuration problem that has to stay visible.
		slog.WarnContext(ctx, "failed to fetch accounts", "environmentId", env.ID, "error", err)
		span.RecordError(err)
		res.supported = true
		res.err = err.Error()
		res.fetchErr = err
		return res
	}

	res.supported = true

	if len(snapshot.Accounts) == 0 {
		slog.WarnContext(ctx, "accounts payload parsed to zero entries, keeping previous baseline",
			"environmentId", env.ID)
		res.err = errEmptyAccountPayload
		return res
	}

	res.fetched = true
	res.accounts = snapshot.Accounts
	if snapshot.HasGenerated {
		res.generatedAt = pgtype.Timestamp{Time: snapshot.GeneratedAt, Valid: true}
	}

	state, err := h.queries.GetEnvironmentAccountState(ctx, env.ID)
	switch {
	case err != nil && !errors.Is(err, pgx.ErrNoRows):
		// Without the state row we cannot tell a first run from a real diff.
		// Guessing either way is harmful — seeding would swallow a genuine new
		// admin, diffing would alarm on every existing one — so skip this run
		// and keep the stored baseline untouched.
		slog.WarnContext(ctx, "failed to load account state, skipping account diff",
			"environmentId", env.ID, "error", err)
		span.RecordError(err)
		res.fetched = false
		res.accounts = nil
		res.err = "account state unavailable: " + err.Error()
		return res
	case err != nil || !state.BaselineAt.Valid:
		// First run for this environment: seed only. Without a prior state
		// every account reads as new, which would turn every existing shop red.
		res.baselineRun = true
		span.SetAttributes(attribute.Bool("accounts.baseline", true))
		return res
	}

	baseline, err := h.queries.ListEnvironmentAccounts(ctx, env.ID)
	if err != nil {
		slog.WarnContext(ctx, "failed to load account baseline, skipping account diff",
			"environmentId", env.ID, "error", err)
		span.RecordError(err)
		res.fetched = false
		res.accounts = nil
		res.err = "account baseline unavailable: " + err.Error()
		return res
	}

	res.changes = accounts.Diff(toAccountModels(baseline), snapshot.Accounts)
	res.checks = append(res.checks, accountChecks(env.Url, res.changes)...)

	span.SetAttributes(
		attribute.Int("accounts.count", len(snapshot.Accounts)),
		attribute.Int("accounts.changes", len(res.changes)),
		attribute.Int("accounts.open_findings", len(res.checks)),
	)
	return res
}

// openAccountChecks rebuilds the checks for findings that are still awaiting
// review. They come from the database rather than from this run, so a shop
// that stops answering cannot resolve a finding nobody acknowledged.
func (h *Service) openAccountChecks(ctx context.Context, env queries.GetAllEnvironmentsRow) []checker.Check {
	open, err := h.queries.ListOpenEnvironmentAccountEvents(ctx, env.ID)
	if err != nil {
		slog.WarnContext(ctx, "failed to load open account findings",
			"environmentId", env.ID, "error", err)
		return nil
	}

	out := make([]checker.Check, 0, len(open))
	for _, e := range open {
		var params map[string]any
		if len(e.Params) > 0 {
			_ = json.Unmarshal(e.Params, &params)
		}
		out = append(out, accountCheck(env.Url, accounts.Kind(e.AccountType), e.AccountID, e.Label,
			accounts.ChangeKind(e.Change), accounts.Level(e.Level), params, ptr.Deref(e.AdminPath)))
	}
	return out
}

// accountChecks turns the changes detected in this run into checks. Only
// findings escalate; informational history never becomes a check, because a
// revoked permission is not something an operator has to acknowledge.
func accountChecks(envURL string, changes []accounts.Change) []checker.Check {
	out := make([]checker.Check, 0, len(changes))
	for _, c := range changes {
		if c.Level == accounts.LevelInfo {
			continue
		}
		out = append(out, accountCheck(envURL, c.Kind, c.AccountID, c.Label, c.Change, c.Level, c.Params, c.AdminPath))
	}
	return out
}

func accountCheck(envURL string, kind accounts.Kind, accountID, label string,
	change accounts.ChangeKind, level accounts.Level, params map[string]any, adminPath string,
) checker.Check {
	merged := make(map[string]any, len(params)+1)
	for k, v := range params {
		merged[k] = v
	}
	merged["label"] = label

	return checker.Check{
		ID:            accounts.CheckID(string(kind), accountID, string(change)),
		Level:         accountCheckLevel(level),
		MessageKey:    accounts.MessageKey(string(change)),
		MessageParams: merged,
		Source:        checker.SourceAccounts,
		Link:          accountAdminURL(envURL, adminPath),
	}
}

// accountCheckLevel maps a threat rating onto a check status. Info never
// reaches here; it would have to become green, which would read as "checked
// and fine" and pollute the checks tab.
func accountCheckLevel(l accounts.Level) checker.Status {
	if l == accounts.LevelRed {
		return checker.StatusRed
	}
	return checker.StatusYellow
}

// accountAdminURL composes the absolute deep link. The path is stored relative
// so a changed shop URL does not leave dead links behind, and the check link is
// rendered as a raw href — a relative path would resolve against Shopmon.
func accountAdminURL(envURL, adminPath string) string {
	if adminPath == "" || envURL == "" {
		return ""
	}
	return strings.TrimRight(envURL, "/") + adminPath
}

// toAccountModels converts stored baseline rows back into the diff's input
// shape.
func toAccountModels(rows []queries.EnvironmentAccount) []accounts.Account {
	out := make([]accounts.Account, 0, len(rows))
	for _, r := range rows {
		var roleIDs []string
		if len(r.RoleIds) > 0 {
			_ = json.Unmarshal(r.RoleIds, &roleIDs)
		}
		out = append(out, accounts.Account{
			Kind:                    accounts.Kind(r.AccountType),
			ID:                      r.AccountID,
			Label:                   r.Label,
			Active:                  r.Active,
			Admin:                   r.Admin,
			RoleIDs:                 roleIDs,
			RolesHash:               r.RolesHash,
			AccessKeyCount:          r.AccessKeyCount,
			PrivilegeCount:          r.PrivilegeCount,
			PrivilegesHash:          r.PrivilegesHash,
			ManagedByApp:            r.ManagedByApp,
			AppName:                 ptr.Deref(r.AppName),
			PasswordRecoveryPending: r.PasswordRecoveryPending,
			LastUsageAt:             database.TimePtr(r.LastUsageAt),
			PasswordChangedAt:       database.TimePtr(r.PasswordChangedAt),
			RemoteCreatedAt:         database.TimePtr(r.RemoteCreatedAt),
			RemoteUpdatedAt:         database.TimePtr(r.RemoteUpdatedAt),
			AdminPath:               ptr.Deref(r.AdminPath),
		})
	}
	return out
}

// persistAccounts writes the baseline, the detected events and the capability
// state. It runs inside the scrape transaction.
func persistAccounts(ctx context.Context, q *queries.Queries, environmentID int32, res accountResult) error {
	if !res.supported {
		// Record that this shop cannot serve accounts, but keep the baseline:
		// the plugin may just be temporarily inactive, and discarding it would
		// make every account new again once it returns.
		return q.UpsertEnvironmentAccountStateStatus(ctx, queries.UpsertEnvironmentAccountStateStatusParams{
			EnvironmentID: environmentID,
			Supported:     false,
			LastError:     nil,
		})
	}

	// A fetch that produced nothing must never reach the pruning below:
	// DeleteEnvironmentAccountsNotIn with an empty keep-list deletes the whole
	// baseline. fetchAccounts already downgrades that case, but the guard
	// belongs on the write path too.
	if !res.fetched || len(res.accounts) == 0 {
		errText := res.err
		if errText == "" {
			errText = errEmptyAccountPayload
		}
		return q.UpsertEnvironmentAccountStateStatus(ctx, queries.UpsertEnvironmentAccountStateStatusParams{
			EnvironmentID: environmentID,
			Supported:     true,
			LastError:     new(sanitize.Error(errText)),
		})
	}

	keys := make([]string, 0, len(res.accounts))
	for _, a := range res.accounts {
		roleIDs, err := json.Marshal(a.RoleIDs)
		if err != nil {
			return fmt.Errorf("marshal role ids for %s: %w", a.ID, err)
		}
		if err := q.UpsertEnvironmentAccount(ctx, queries.UpsertEnvironmentAccountParams{
			EnvironmentID:           environmentID,
			AccountType:             string(a.Kind),
			AccountID:               a.ID,
			Label:                   a.Label,
			Active:                  a.Active,
			Admin:                   a.Admin,
			RoleIds:                 roleIDs,
			RolesHash:               a.RolesHash,
			AccessKeyCount:          a.AccessKeyCount,
			PrivilegeCount:          a.PrivilegeCount,
			PrivilegesHash:          a.PrivilegesHash,
			ManagedByApp:            a.ManagedByApp,
			AppName:                 emptyToNil(a.AppName),
			PasswordRecoveryPending: a.PasswordRecoveryPending,
			LastUsageAt:             database.TimestampFromPtr(a.LastUsageAt),
			PasswordChangedAt:       database.TimestampFromPtr(a.PasswordChangedAt),
			RemoteCreatedAt:         database.TimestampFromPtr(a.RemoteCreatedAt),
			RemoteUpdatedAt:         database.TimestampFromPtr(a.RemoteUpdatedAt),
			AdminPath:               emptyToNil(a.AdminPath),
		}); err != nil {
			return fmt.Errorf("upsert environment account %s: %w", a.ID, err)
		}
		keys = append(keys, string(a.Kind)+":"+a.ID)
	}

	if err := q.DeleteEnvironmentAccountsNotIn(ctx, queries.DeleteEnvironmentAccountsNotInParams{
		EnvironmentID: environmentID,
		Keys:          keys,
	}); err != nil {
		return fmt.Errorf("delete stale environment accounts: %w", err)
	}

	// On the seeding run the diff is meaningless (everything is "new"), so the
	// changes are deliberately dropped rather than stored.
	if !res.baselineRun {
		for _, c := range res.changes {
			params := []byte("{}")
			if len(c.Params) > 0 {
				if b, err := json.Marshal(c.Params); err == nil {
					params = b
				}
			}
			if err := q.InsertEnvironmentAccountEvent(ctx, queries.InsertEnvironmentAccountEventParams{
				EnvironmentID: environmentID,
				AccountType:   string(c.Kind),
				AccountID:     c.AccountID,
				Label:         c.Label,
				Change:        string(c.Change),
				Level:         string(c.Level),
				Params:        params,
				AdminPath:     emptyToNil(c.AdminPath),
			}); err != nil {
				return fmt.Errorf("insert account event %s/%s: %w", c.AccountID, c.Change, err)
			}
		}
	}

	return q.UpsertEnvironmentAccountState(ctx, queries.UpsertEnvironmentAccountStateParams{
		EnvironmentID: environmentID,
		AccountCount:  int32(len(res.accounts)),
		GeneratedAt:   res.generatedAt,
	})
}

// notifyAccountChanges tells the environment's subscribers about findings they
// have not been alerted to yet.
//
// It reads the pending events from the database rather than from this scrape's
// result, so a dispatch that failed last run is retried instead of lost. It
// must run after the scrape transaction commits: marking events notified
// inside it would lose both the marker and the baseline on a rollback.
func (h *Service) notifyAccountChanges(ctx context.Context, env queries.GetAllEnvironmentsRow) {
	if !env.AccountMonitoringEnabled {
		return
	}

	pending, err := h.queries.ListUnnotifiedEnvironmentAccountEvents(ctx, env.ID)
	if err != nil {
		slog.WarnContext(ctx, "failed to load unnotified account events",
			"environmentId", env.ID, "error", err)
		return
	}
	if len(pending) == 0 {
		return
	}

	recipients := h.environmentRecipients(ctx, env)
	if len(recipients) == 0 {
		// Nobody subscribes, so there is nothing to retry either. Marking the
		// events keeps them from being reconsidered every hour.
		h.markAccountEventsNotified(ctx, env.ID, pending)
		return
	}

	reasons := make([]notify.StatusReason, 0, len(pending))
	ids := make([]int64, 0, len(pending))
	level := notify.LevelWarning
	for _, e := range pending {
		ids = append(ids, e.ID)
		if e.Level == string(accounts.LevelRed) {
			level = notify.LevelError
		}
		var params map[string]any
		if len(e.Params) > 0 {
			_ = json.Unmarshal(e.Params, &params)
		}
		if params == nil {
			params = map[string]any{}
		}
		params["label"] = e.Label
		reasons = append(reasons, notify.StatusReason{
			Level:  e.Level,
			Key:    accounts.MessageKey(e.Change),
			Params: params,
			Source: checker.SourceAccounts,
		})
	}
	titleKey := "notification.accountChanged.title"
	messageKey := "notification.accountChanged.message"
	subjectKey := "email.accountChanged.subject"
	if len(pending) == 1 {
		titleKey = "notification.accountChanged.titleOne"
		messageKey = "notification.accountChanged.messageOne"
		subjectKey = "email.accountChanged.subjectOne"
	}

	res := h.notifier.Dispatch(ctx, notify.Event{
		Type:      notify.EventAccountChanged,
		Level:     level,
		ScopeType: notify.ScopeEnvironment,
		ScopeID:   strconv.Itoa(int(env.ID)),
		OrgID:     env.OrganizationID,
		// Keyed by the newest event rather than by the environment alone: a
		// per-environment key would let the hourly email dedup lock swallow a
		// second wave of changes, and the in-app entry for the first wave
		// would be overwritten before anyone had reviewed it.
		DedupKey:        fmt.Sprintf("account_changed:%d:%d", env.ID, ids[len(ids)-1]),
		TitleKey:        titleKey,
		MessageKey:      messageKey,
		EmailSubjectKey: subjectKey,
		Params: map[string]any{
			"name":  environmentDisplayName(env.Name, env.ShopName),
			"count": len(pending),
		},
		Reasons: reasons,
		Link:    accountsLink(env.ID),
	}, recipients)

	// Only mark when nothing failed: a channel error means these subscribers
	// were not told, and marking anyway would drop the events from every
	// future run so the alert would never be retried. A dispatch that
	// delivered nothing but failed nothing (every channel skipped by the dedup
	// lock or by recipient preference) counts as handled.
	if res.Failed > 0 {
		slog.WarnContext(ctx, "account alert partially failed; leaving events unnotified for retry",
			"environmentId", env.ID, "delivered", res.Delivered, "failed", res.Failed, "events", len(pending))
		return
	}

	h.markAccountEventsNotified(ctx, env.ID, pending)
}

func (h *Service) markAccountEventsNotified(ctx context.Context, environmentID int32, events []queries.EnvironmentAccountEvent) {
	ids := make([]int64, 0, len(events))
	for _, e := range events {
		ids = append(ids, e.ID)
	}
	if err := h.queries.MarkEnvironmentAccountEventsNotified(ctx, queries.MarkEnvironmentAccountEventsNotifiedParams{
		EnvironmentID: environmentID,
		Ids:           ids,
	}); err != nil {
		slog.WarnContext(ctx, "failed to mark account events notified",
			"environmentId", environmentID, "error", err)
	}
}

// accountsLink points at the environment's account monitoring tab, where the
// findings can be reviewed and acknowledged.
func accountsLink(envID int32) notify.Link {
	return notify.Link{
		Name: "account.environments.detail.accounts",
		Params: map[string]string{
			"environmentId": strconv.Itoa(int(envID)),
		},
	}
}
