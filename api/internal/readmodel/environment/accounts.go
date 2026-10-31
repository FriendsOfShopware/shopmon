package environment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/accounts"
)

// accountHistoryLimit bounds the timeline. Older entries are trimmed by the
// cleanup job; open findings are never trimmed.
const accountHistoryLimit = 100

// dormantAfter is how long an administrator integration may go unused before
// it is worth pointing at. Credentials nobody uses are also credentials nobody
// rotates.
const dormantAfter = 90 * 24 * time.Hour

// Posture hint identifiers. These rate the current state rather than a change,
// so they are advisory only: never a check, never an effect on shop status.
const (
	hintUnusedAdminIntegration  = "unusedAdminIntegration"
	hintUserWithApiKeys         = "userWithApiKeys"
	hintPasswordRecoveryPending = "passwordRecoveryPending"
)

// Accounts returns the account-monitoring view of one environment: the
// findings awaiting review, the change history and the posture hints.
func (s *Service) Accounts(ctx context.Context, userID string, environmentID int32) (api.EnvironmentAccounts, error) {
	environment, err := s.queries.GetEnvironmentByID(ctx, environmentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.EnvironmentAccounts{}, ErrNotFound
		}
		return api.EnvironmentAccounts{}, fmt.Errorf("get environment: %w", err)
	}
	member, err := s.authorizer.IsMember(ctx, userID, environment.OrganizationID)
	if err != nil {
		return api.EnvironmentAccounts{}, fmt.Errorf("authorize organization membership: %w", err)
	}
	if !member {
		return api.EnvironmentAccounts{}, ErrNotAuthorized
	}

	out := api.EnvironmentAccounts{
		Enabled:      environment.AccountMonitoringEnabled,
		OpenFindings: []api.AccountFinding{},
		Events:       []api.AccountEvent{},
		PostureHints: []api.AccountPostureHint{},
	}

	// A missing state row means the environment was never read; that is not an
	// error, it is the state before the first scrape.
	state, err := s.queries.GetEnvironmentAccountState(ctx, environmentID)
	switch {
	case err == nil:
		out.Supported = state.Supported
		out.BaselineAt = database.TimePtr(state.BaselineAt)
		out.LastSyncedAt = database.TimePtr(state.LastSyncedAt)
		out.LastError = state.LastError
		out.AccountCount = int(state.AccountCount)
	case errors.Is(err, pgx.ErrNoRows):
		out.Supported = true
	default:
		return api.EnvironmentAccounts{}, fmt.Errorf("get account state: %w", err)
	}

	baseline, err := s.queries.ListEnvironmentAccounts(ctx, environmentID)
	if err != nil {
		return api.EnvironmentAccounts{}, fmt.Errorf("list accounts: %w", err)
	}
	snapshots := indexSnapshots(baseline)

	open, err := s.queries.ListOpenEnvironmentAccountEvents(ctx, environmentID)
	if err != nil {
		return api.EnvironmentAccounts{}, fmt.Errorf("list open account findings: %w", err)
	}
	for _, e := range open {
		out.OpenFindings = append(out.OpenFindings, api.AccountFinding{
			Id:         e.ID,
			Level:      e.Level,
			Kind:       e.AccountType,
			Change:     e.Change,
			Label:      e.Label,
			MessageKey: accounts.MessageKey(e.Change),
			Params:     renderParams(e.Params, e.Label),
			AdminUrl:   adminURL(environment.Url, e.AdminPath),
			DetectedAt: database.Time(e.DetectedAt),
			CheckId:    accounts.CheckID(e.AccountType, e.AccountID, e.Change),
			Account:    snapshots[e.AccountType+":"+e.AccountID],
		})
	}

	events, err := s.queries.ListEnvironmentAccountEvents(ctx, queries.ListEnvironmentAccountEventsParams{
		EnvironmentID: environmentID,
		Limit:         accountHistoryLimit,
	})
	if err != nil {
		return api.EnvironmentAccounts{}, fmt.Errorf("list account events: %w", err)
	}
	for _, e := range events {
		out.Events = append(out.Events, api.AccountEvent{
			Id:              e.ID,
			Level:           e.Level,
			Kind:            e.AccountType,
			Change:          e.Change,
			Label:           e.Label,
			MessageKey:      accounts.MessageKey(e.Change),
			Params:          renderParams(e.Params, e.Label),
			AdminUrl:        adminURL(environment.Url, e.AdminPath),
			DetectedAt:      database.Time(e.DetectedAt),
			AcknowledgedAt:  database.TimePtr(e.AcknowledgedAt),
			AcknowledgedBy:  e.AcknowledgedByName,
			AcknowledgeNote: e.AcknowledgeNote,
			Account:         snapshots[e.AccountType+":"+e.AccountID],
		})
	}

	out.PostureHints = postureHints(environment.Url, baseline, time.Now())

	return out, nil
}

// postureHints rates the accounts that currently exist. Only accounts worth
// pointing at are returned — the result is a list of concerns, not a roster.
func postureHints(envURL string, rows []queries.EnvironmentAccount, now time.Time) []api.AccountPostureHint {
	out := make([]api.AccountPostureHint, 0)
	cutoff := now.Add(-dormantAfter)

	for _, r := range rows {
		switch accounts.Kind(r.AccountType) {
		case accounts.KindIntegration:
			// App-managed integrations are maintained by their app, and their
			// "last usage" says nothing about whether the app is in use.
			if !r.Admin || r.ManagedByApp {
				continue
			}
			// An integration created moments ago has simply not been used yet.
			// Shops do not always report a creation date, so fall back to when
			// Shopmon first saw the account -- otherwise such an integration
			// would never qualify for the hint at all. The hint's CreatedAt
			// stays on the shop's own value, since the UI labels it as the
			// creation date in the shop.
			since := r.RemoteCreatedAt
			if !since.Valid {
				since = r.FirstSeenAt
			}
			if !since.Valid || since.Time.After(cutoff) {
				continue
			}
			if r.LastUsageAt.Valid && r.LastUsageAt.Time.After(cutoff) {
				continue
			}
			out = append(out, api.AccountPostureHint{
				Hint:        hintUnusedAdminIntegration,
				Kind:        r.AccountType,
				Label:       r.Label,
				AdminUrl:    adminURL(envURL, r.AdminPath),
				LastUsageAt: database.TimePtr(r.LastUsageAt),
				CreatedAt:   database.TimePtr(r.RemoteCreatedAt),
			})

		case accounts.KindUser:
			if r.AccessKeyCount > 0 {
				count := int(r.AccessKeyCount)
				out = append(out, api.AccountPostureHint{
					Hint:           hintUserWithApiKeys,
					Kind:           r.AccountType,
					Label:          r.Label,
					AdminUrl:       adminURL(envURL, r.AdminPath),
					AccessKeyCount: &count,
					CreatedAt:      database.TimePtr(r.RemoteCreatedAt),
				})
			}
			if r.PasswordRecoveryPending {
				out = append(out, api.AccountPostureHint{
					Hint:      hintPasswordRecoveryPending,
					Kind:      r.AccountType,
					Label:     r.Label,
					AdminUrl:  adminURL(envURL, r.AdminPath),
					CreatedAt: database.TimePtr(r.RemoteCreatedAt),
				})
			}
		}
	}
	return out
}

// adminURL composes the absolute deep link. The stored path is relative so a
// changed shop URL leaves no dead links; a relative href would resolve against
// Shopmon's own origin.
func adminURL(envURL string, adminPath *string) *string {
	if adminPath == nil || *adminPath == "" || envURL == "" {
		return nil
	}
	return new(strings.TrimRight(envURL, "/") + *adminPath)
}

// renderParams merges the account label into the stored change params.
//
// The message templates interpolate {label}, but the label lives in its own
// column rather than in params (it is needed for sorting and for accounts that
// no longer exist). Without this merge every message would render with an
// empty subject — "New administrator account  was created" — which is exactly
// the information a reviewer needs.
func renderParams(raw []byte, label string) *map[string]any {
	params := map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &params); err != nil {
			params = map[string]any{}
		}
	}
	params["label"] = label
	return &params
}

// indexSnapshots maps "type:id" to the account's current state, so a finding
// can show what the account looks like now. Accounts that no longer exist are
// simply absent.
func indexSnapshots(rows []queries.EnvironmentAccount) map[string]*api.AccountSnapshot {
	out := make(map[string]*api.AccountSnapshot, len(rows))
	for _, r := range rows {
		roleCount := 0
		if len(r.RoleIds) > 0 {
			var ids []string
			if err := json.Unmarshal(r.RoleIds, &ids); err == nil {
				roleCount = len(ids)
			}
		}
		out[r.AccountType+":"+r.AccountID] = &api.AccountSnapshot{
			Active:                  r.Active,
			Admin:                   r.Admin,
			RoleCount:               roleCount,
			AccessKeyCount:          int(r.AccessKeyCount),
			PrivilegeCount:          int(r.PrivilegeCount),
			ManagedByApp:            r.ManagedByApp,
			AppName:                 r.AppName,
			PasswordRecoveryPending: r.PasswordRecoveryPending,
			LastUsageAt:             database.TimePtr(r.LastUsageAt),
			PasswordChangedAt:       database.TimePtr(r.PasswordChangedAt),
			CreatedAt:               database.TimePtr(r.RemoteCreatedAt),
			FirstSeenAt:             database.Time(r.FirstSeenAt),
		}
	}
	return out
}
