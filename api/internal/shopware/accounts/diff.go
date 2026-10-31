package accounts

import (
	"sort"
	"strings"
	"time"
)

// Level is the threat rating of a detected change.
//
// It deliberately does not reuse checker.Status: that type has no "info" and
// its green means "checked and fine", which is the wrong thing to say about a
// revoked permission. Info changes are recorded in the history and never
// become a check.
type Level string

const (
	LevelRed    Level = "red"
	LevelYellow Level = "yellow"
	LevelInfo   Level = "info"
)

// ChangeKind names one detected difference between two snapshots. The value is
// stored, drives the translation key and is part of the check id, so these
// strings are a stable contract.
type ChangeKind string

const (
	// Red: someone gained full control of the shop.
	ChangeNewAdminUser        ChangeKind = "new_admin_user"
	ChangeNewAdminIntegration ChangeKind = "new_admin_integration"
	ChangeAdminGranted        ChangeKind = "admin_granted"

	// Yellow: access was created or widened, but not to full admin.
	ChangeNewUser             ChangeKind = "new_user"
	ChangeNewIntegration      ChangeKind = "new_integration"
	ChangeRolesAdded          ChangeKind = "roles_added"
	ChangeRolePrivilegesGrown ChangeKind = "role_privileges_grown"
	// ChangeRolePrivilegesChanged covers an edit that left the count untouched
	// but altered the set. The endpoint exposes no privilege list, so this is
	// as specific as it can get.
	ChangeRolePrivilegesChanged  ChangeKind = "role_privileges_changed"
	ChangeAccessKeysAdded        ChangeKind = "access_keys_added"
	ChangeReactivated            ChangeKind = "reactivated"
	ChangePasswordRecoveryOpened ChangeKind = "password_recovery_opened"

	// Info: history only. Access was narrowed, or the change carries no risk.
	ChangeAdminRevoked           ChangeKind = "admin_revoked"
	ChangeRolesRemoved           ChangeKind = "roles_removed"
	ChangeRolePrivilegesShrunk   ChangeKind = "role_privileges_shrunk"
	ChangeAccessKeysRemoved      ChangeKind = "access_keys_removed"
	ChangeDeactivated            ChangeKind = "deactivated"
	ChangeDeleted                ChangeKind = "deleted"
	ChangeNewManagedIntegration  ChangeKind = "new_managed_integration"
	ChangeNewACLRole             ChangeKind = "new_acl_role"
	ChangePasswordChanged        ChangeKind = "password_changed"
	ChangePasswordRecoveryClosed ChangeKind = "password_recovery_closed"
)

// levels is the single source of truth for how dangerous each change is.
var levels = map[ChangeKind]Level{
	ChangeNewAdminUser:        LevelRed,
	ChangeNewAdminIntegration: LevelRed,
	ChangeAdminGranted:        LevelRed,

	ChangeNewUser:                LevelYellow,
	ChangeNewIntegration:         LevelYellow,
	ChangeRolesAdded:             LevelYellow,
	ChangeRolePrivilegesGrown:    LevelYellow,
	ChangeRolePrivilegesChanged:  LevelYellow,
	ChangeAccessKeysAdded:        LevelYellow,
	ChangeReactivated:            LevelYellow,
	ChangePasswordRecoveryOpened: LevelYellow,

	ChangeAdminRevoked:           LevelInfo,
	ChangeRolesRemoved:           LevelInfo,
	ChangeRolePrivilegesShrunk:   LevelYellow,
	ChangeAccessKeysRemoved:      LevelInfo,
	ChangeDeactivated:            LevelInfo,
	ChangeDeleted:                LevelInfo,
	ChangeNewManagedIntegration:  LevelInfo,
	ChangeNewACLRole:             LevelInfo,
	ChangePasswordChanged:        LevelInfo,
	ChangePasswordRecoveryClosed: LevelInfo,
}

// Level reports the threat rating. An unknown kind falls back to info so a
// future change type can never silently turn a shop red.
func (c ChangeKind) Level() Level {
	if l, ok := levels[c]; ok {
		return l
	}
	return LevelInfo
}

// Change is one detected difference, ready to be stored as an event.
type Change struct {
	Kind      Kind
	AccountID string
	Label     string
	Change    ChangeKind
	Level     Level
	Params    map[string]any
	AdminPath string
}

// Diff reports what changed between a stored baseline and a fresh snapshot.
//
// Callers must not pass an empty snapshot: with nothing to compare against,
// every account reads as deleted. fetchAccounts rejects an empty payload
// before it gets here.
//
// One account can yield several changes in a single run (reactivated *and*
// granted admin); each becomes its own event and its own check.
func Diff(baseline, current []Account) []Change {
	prev := index(baseline)
	changes := make([]Change, 0)

	// Which roles are actually assigned right now. A role nobody holds grants
	// nothing, so growing its privileges is noise rather than a finding.
	assigned := assignedRoles(current)

	for _, acc := range current {
		old, existed := prev[key(acc)]
		if !existed {
			changes = append(changes, newAccountChange(acc))
			continue
		}
		changes = append(changes, compare(old, acc, assigned)...)
	}

	seen := index(current)
	for _, old := range baseline {
		if _, stillThere := seen[key(old)]; stillThere {
			continue
		}
		changes = append(changes, change(old, ChangeDeleted, nil))
	}

	sort.SliceStable(changes, func(a, b int) bool {
		return levelWeight(changes[a].Level) > levelWeight(changes[b].Level)
	})
	return changes
}

func levelWeight(l Level) int {
	switch l {
	case LevelRed:
		return 3
	case LevelYellow:
		return 2
	default:
		return 1
	}
}

// newAccountChange rates an account that was not in the baseline.
func newAccountChange(acc Account) Change {
	switch acc.Kind {
	case KindUser:
		if acc.Admin {
			return change(acc, ChangeNewAdminUser, nil)
		}
		return change(acc, ChangeNewUser, nil)
	case KindIntegration:
		// An app-managed integration appears because an app was installed,
		// which Shopmon already reports as an extension change. Alerting again
		// would make every app install look like a security event.
		if acc.ManagedByApp {
			return change(acc, ChangeNewManagedIntegration, map[string]any{"appName": acc.AppName})
		}
		if acc.Admin {
			return change(acc, ChangeNewAdminIntegration, nil)
		}
		return change(acc, ChangeNewIntegration, nil)
	default:
		return change(acc, ChangeNewACLRole, map[string]any{"count": acc.PrivilegeCount})
	}
}

// compare rates every difference between the stored and the current state of
// one account.
func compare(old, acc Account, assigned map[string]int) []Change {
	var out []Change

	if !old.Admin && acc.Admin {
		out = append(out, change(acc, ChangeAdminGranted, nil))
	}
	if old.Admin && !acc.Admin {
		out = append(out, change(acc, ChangeAdminRevoked, nil))
	}

	if old.RolesHash != acc.RolesHash {
		added, removed := roleDelta(old.RoleIDs, acc.RoleIDs)
		if added > 0 {
			out = append(out, change(acc, ChangeRolesAdded, map[string]any{"count": added}))
		}
		if removed > 0 {
			out = append(out, change(acc, ChangeRolesRemoved, map[string]any{"count": removed}))
		}
	}

	if acc.Kind == KindACLRole {
		out = append(out, rolePrivilegeChanges(old, acc, assigned)...)
	}

	switch {
	case acc.AccessKeyCount > old.AccessKeyCount:
		out = append(out, change(acc, ChangeAccessKeysAdded, countParams(old.AccessKeyCount, acc.AccessKeyCount)))
	case acc.AccessKeyCount < old.AccessKeyCount:
		out = append(out, change(acc, ChangeAccessKeysRemoved, countParams(old.AccessKeyCount, acc.AccessKeyCount)))
	}

	if !old.Active && acc.Active {
		out = append(out, change(acc, ChangeReactivated, nil))
	}
	if old.Active && !acc.Active {
		out = append(out, change(acc, ChangeDeactivated, nil))
	}

	if !old.PasswordRecoveryPending && acc.PasswordRecoveryPending {
		out = append(out, change(acc, ChangePasswordRecoveryOpened, nil))
	}
	if old.PasswordRecoveryPending && !acc.PasswordRecoveryPending {
		out = append(out, change(acc, ChangePasswordRecoveryClosed, nil))
	}

	if !sameTime(old.PasswordChangedAt, acc.PasswordChangedAt) {
		out = append(out, change(acc, ChangePasswordChanged, nil))
	}

	return out
}

// rolePrivilegeChanges rates a change to an ACL role's privilege set.
//
// The endpoint reports only a count and a hash, never the privileges
// themselves, so a shrinking count does NOT prove that the role merely lost
// rights: it can shed five read permissions and gain order:delete in the same
// edit, ending up smaller and far more powerful. Any changed hash on a role
// somebody actually holds is therefore reported as a finding, with the count
// movement carried in the params so a reviewer can see which way it went.
//
// A role nobody is assigned to grants nothing, so a change there cannot widen
// access and stays informational -- app updates rewrite such roles routinely.
func rolePrivilegeChanges(old, acc Account, assigned map[string]int) []Change {
	if old.PrivilegeCount == acc.PrivilegeCount && old.PrivilegesHash == acc.PrivilegesHash {
		return nil
	}

	holders := assigned[acc.ID]
	params := map[string]any{
		"from":    old.PrivilegeCount,
		"to":      acc.PrivilegeCount,
		"holders": holders,
	}

	kind := ChangeRolePrivilegesChanged
	switch {
	case acc.PrivilegeCount > old.PrivilegeCount:
		kind = ChangeRolePrivilegesGrown
	case acc.PrivilegeCount < old.PrivilegeCount:
		kind = ChangeRolePrivilegesShrunk
	}

	c := change(acc, kind, params)
	if holders == 0 {
		c.Level = LevelInfo
	} else {
		// Shrinking is only informational when nobody holds the role; for a
		// held role the set changed and the endpoint cannot prove it is safe.
		c.Level = LevelYellow
	}
	return []Change{c}
}

func change(acc Account, kind ChangeKind, params map[string]any) Change {
	return Change{
		Kind:      acc.Kind,
		AccountID: acc.ID,
		Label:     acc.Label,
		Change:    kind,
		Level:     kind.Level(),
		Params:    params,
		AdminPath: acc.AdminPath,
	}
}

func countParams(from, to int32) map[string]any {
	return map[string]any{"from": from, "to": to}
}

func key(a Account) string { return string(a.Kind) + ":" + a.ID }

func index(in []Account) map[string]Account {
	out := make(map[string]Account, len(in))
	for _, a := range in {
		out[key(a)] = a
	}
	return out
}

// assignedRoles counts how many users and integrations hold each role.
func assignedRoles(in []Account) map[string]int {
	out := make(map[string]int)
	for _, a := range in {
		if a.Kind == KindACLRole {
			continue
		}
		for _, id := range a.RoleIDs {
			out[id]++
		}
	}
	return out
}

func roleDelta(old, current []string) (added, removed int) {
	had := make(map[string]bool, len(old))
	for _, id := range old {
		had[id] = true
	}
	has := make(map[string]bool, len(current))
	for _, id := range current {
		has[id] = true
		if !had[id] {
			added++
		}
	}
	for _, id := range old {
		if !has[id] {
			removed++
		}
	}
	return added, removed
}

func sameTime(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equal(*b)
}

// CheckIDPrefix namespaces every account check id. It must not be a prefix of
// (or prefixed by) any other checker id prefix: checker.UnavailableSource.Owns
// matches with strings.HasPrefix, so an overlapping namespace would let one
// checker's outage claim another's findings.
const CheckIDPrefix = "account."

// CheckID is the stable handle for one finding. It is deterministic so the
// same open finding keeps one check row across scrapes, and so two events
// about the same change collapse onto one check instead of fighting over the
// unique (environment_id, check_id) key.
func CheckID(kind, accountID, change string) string {
	return CheckIDPrefix + kind + "." + accountID + "." + change
}

// MessageKey derives the translation catalog key from a change kind:
// "new_admin_user" -> "check.account.newAdminUser".
func MessageKey(change string) string {
	parts := strings.Split(change, "_")
	var b strings.Builder
	b.WriteString("check.account.")
	for i, p := range parts {
		if p == "" {
			continue
		}
		if i == 0 {
			b.WriteString(p)
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]))
		b.WriteString(p[1:])
	}
	return b.String()
}
