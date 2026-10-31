package accounts

import (
	"testing"
	"time"
)

func user(id string, opts ...func(*Account)) Account {
	a := Account{Kind: KindUser, ID: id, Label: "u***@e***.de", Active: true}
	for _, o := range opts {
		o(&a)
	}
	return a
}

func integration(id string, opts ...func(*Account)) Account {
	a := Account{Kind: KindIntegration, ID: id, Label: "Some Integration", Active: true}
	for _, o := range opts {
		o(&a)
	}
	return a
}

func role(id string, count int32, hash string, opts ...func(*Account)) Account {
	a := Account{Kind: KindACLRole, ID: id, Label: "Editor", Active: true, PrivilegeCount: count, PrivilegesHash: hash}
	for _, o := range opts {
		o(&a)
	}
	return a
}

func admin(a *Account)        { a.Admin = true }
func inactive(a *Account)     { a.Active = false }
func managedByApp(a *Account) { a.ManagedByApp = true; a.AppName = "SwagApp" }

func withRoles(ids ...string) func(*Account) {
	return func(a *Account) { a.RoleIDs, a.RolesHash = normalizeRoleIDs(ids) }
}

func withKeys(n int32) func(*Account) { return func(a *Account) { a.AccessKeyCount = n } }

func recoveryPending(a *Account) { a.PasswordRecoveryPending = true }

func withPasswordChangedAt(s string) func(*Account) {
	return func(a *Account) {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			panic(err)
		}
		utc := t.UTC()
		a.PasswordChangedAt = &utc
	}
}

const (
	idA    = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	idB    = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	roleID = "cccccccccccccccccccccccccccccccc"
)

// want is one expected change, reduced to the fields the matrix is about.
type want struct {
	change ChangeKind
	level  Level
}

func TestDiffChangeMatrix(t *testing.T) {
	cases := []struct {
		name     string
		baseline []Account
		current  []Account
		want     []want
	}{
		{
			name:     "new admin user is red",
			baseline: []Account{},
			current:  []Account{user(idA, admin)},
			want:     []want{{ChangeNewAdminUser, LevelRed}},
		},
		{
			name:     "new non-admin user is yellow",
			baseline: []Account{},
			current:  []Account{user(idA)},
			want:     []want{{ChangeNewUser, LevelYellow}},
		},
		{
			name:     "new admin integration is red",
			baseline: []Account{},
			current:  []Account{integration(idA, admin)},
			want:     []want{{ChangeNewAdminIntegration, LevelRed}},
		},
		{
			name:     "new non-admin integration is yellow",
			baseline: []Account{},
			current:  []Account{integration(idA)},
			want:     []want{{ChangeNewIntegration, LevelYellow}},
		},
		{
			name:     "app managed integration is only informational",
			baseline: []Account{},
			current:  []Account{integration(idA, admin, managedByApp)},
			want:     []want{{ChangeNewManagedIntegration, LevelInfo}},
		},
		{
			name:     "new acl role is only informational",
			baseline: []Account{},
			current:  []Account{role(roleID, 10, "h1")},
			want:     []want{{ChangeNewACLRole, LevelInfo}},
		},
		{
			name:     "granting admin is red",
			baseline: []Account{user(idA)},
			current:  []Account{user(idA, admin)},
			want:     []want{{ChangeAdminGranted, LevelRed}},
		},
		{
			name:     "granting admin to an app managed integration is still red",
			baseline: []Account{integration(idA, managedByApp)},
			current:  []Account{integration(idA, managedByApp, admin)},
			want:     []want{{ChangeAdminGranted, LevelRed}},
		},
		{
			name:     "revoking admin is only informational",
			baseline: []Account{user(idA, admin)},
			current:  []Account{user(idA)},
			want:     []want{{ChangeAdminRevoked, LevelInfo}},
		},
		{
			name:     "assigning a role is yellow",
			baseline: []Account{user(idA)},
			current:  []Account{user(idA, withRoles(roleID))},
			want:     []want{{ChangeRolesAdded, LevelYellow}},
		},
		{
			name:     "removing a role is only informational",
			baseline: []Account{user(idA, withRoles(roleID))},
			current:  []Account{user(idA)},
			want:     []want{{ChangeRolesRemoved, LevelInfo}},
		},
		{
			name:     "swapping a role reports both directions",
			baseline: []Account{user(idA, withRoles(roleID))},
			current:  []Account{user(idA, withRoles(idB))},
			want:     []want{{ChangeRolesAdded, LevelYellow}, {ChangeRolesRemoved, LevelInfo}},
		},
		{
			name:     "a held role gaining privileges is yellow",
			baseline: []Account{role(roleID, 10, "h1"), user(idA, withRoles(roleID))},
			current:  []Account{role(roleID, 20, "h2"), user(idA, withRoles(roleID))},
			want:     []want{{ChangeRolePrivilegesGrown, LevelYellow}},
		},
		{
			name:     "a held role changing privileges without growing is reported",
			baseline: []Account{role(roleID, 10, "h1"), user(idA, withRoles(roleID))},
			current:  []Account{role(roleID, 10, "h2"), user(idA, withRoles(roleID))},
			want:     []want{{ChangeRolePrivilegesChanged, LevelYellow}},
		},
		{
			name:     "an unassigned role gaining privileges cannot widen access",
			baseline: []Account{role(roleID, 10, "h1")},
			current:  []Account{role(roleID, 20, "h2")},
			want:     []want{{ChangeRolePrivilegesGrown, LevelInfo}},
		},
		{
			// The endpoint exposes only a count and a hash, so a shrinking set
			// does not prove the role merely lost rights: it can shed five
			// read permissions and gain order:delete in the same edit.
			name:     "a held role losing privileges is still reported",
			baseline: []Account{role(roleID, 20, "h2"), user(idA, withRoles(roleID))},
			current:  []Account{role(roleID, 10, "h1"), user(idA, withRoles(roleID))},
			want:     []want{{ChangeRolePrivilegesShrunk, LevelYellow}},
		},
		{
			name:     "an unassigned role losing privileges grants nothing",
			baseline: []Account{role(roleID, 20, "h2")},
			current:  []Account{role(roleID, 10, "h1")},
			want:     []want{{ChangeRolePrivilegesShrunk, LevelInfo}},
		},
		{
			name:     "an unassigned role changing privileges grants nothing",
			baseline: []Account{role(roleID, 10, "h1")},
			current:  []Account{role(roleID, 10, "h2")},
			want:     []want{{ChangeRolePrivilegesChanged, LevelInfo}},
		},
		{
			name:     "new api keys are yellow",
			baseline: []Account{user(idA, withKeys(0))},
			current:  []Account{user(idA, withKeys(2))},
			want:     []want{{ChangeAccessKeysAdded, LevelYellow}},
		},
		{
			name:     "removed api keys are only informational",
			baseline: []Account{user(idA, withKeys(2))},
			current:  []Account{user(idA, withKeys(0))},
			want:     []want{{ChangeAccessKeysRemoved, LevelInfo}},
		},
		{
			name:     "reactivation is yellow",
			baseline: []Account{user(idA, inactive)},
			current:  []Account{user(idA)},
			want:     []want{{ChangeReactivated, LevelYellow}},
		},
		{
			name:     "deactivation is only informational",
			baseline: []Account{user(idA)},
			current:  []Account{user(idA, inactive)},
			want:     []want{{ChangeDeactivated, LevelInfo}},
		},
		{
			name:     "an opened password reset is yellow",
			baseline: []Account{user(idA)},
			current:  []Account{user(idA, recoveryPending)},
			want:     []want{{ChangePasswordRecoveryOpened, LevelYellow}},
		},
		{
			name:     "a resolved password reset is only informational",
			baseline: []Account{user(idA, recoveryPending)},
			current:  []Account{user(idA)},
			want:     []want{{ChangePasswordRecoveryClosed, LevelInfo}},
		},
		{
			name:     "a password change is only informational",
			baseline: []Account{user(idA, withPasswordChangedAt("2026-01-01T00:00:00Z"))},
			current:  []Account{user(idA, withPasswordChangedAt("2026-02-01T00:00:00Z"))},
			want:     []want{{ChangePasswordChanged, LevelInfo}},
		},
		{
			name:     "a vanished account is only informational",
			baseline: []Account{user(idA, admin)},
			current:  []Account{user(idB)},
			want:     []want{{ChangeNewUser, LevelYellow}, {ChangeDeleted, LevelInfo}},
		},
		{
			name:     "an unchanged snapshot reports nothing",
			baseline: []Account{user(idA, admin, withRoles(roleID), withKeys(2)), role(roleID, 10, "h1")},
			current:  []Account{user(idA, admin, withRoles(roleID), withKeys(2)), role(roleID, 10, "h1")},
			want:     nil,
		},
		{
			name:     "one account can change in several ways at once",
			baseline: []Account{user(idA, inactive)},
			current:  []Account{user(idA, admin, withKeys(1))},
			want: []want{
				{ChangeAdminGranted, LevelRed},
				{ChangeAccessKeysAdded, LevelYellow},
				{ChangeReactivated, LevelYellow},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Diff(tc.baseline, tc.current)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d changes %v, want %d %v", len(got), kinds(got), len(tc.want), tc.want)
			}
			for _, w := range tc.want {
				if !hasChange(got, w) {
					t.Errorf("missing %s at level %s; got %v", w.change, w.level, kinds(got))
				}
			}
		})
	}
}

func hasChange(got []Change, w want) bool {
	for _, c := range got {
		if c.Change == w.change && c.Level == w.level {
			return true
		}
	}
	return false
}

func kinds(in []Change) []string {
	out := make([]string, 0, len(in))
	for _, c := range in {
		out = append(out, string(c.Change)+"/"+string(c.Level))
	}
	return out
}

// The most dangerous change has to be first so a batched notification leads
// with it and a truncated list never drops the red one.
func TestDiffSortsBySeverity(t *testing.T) {
	got := Diff(
		[]Account{user(idA, inactive)},
		[]Account{user(idA, admin, withKeys(1))},
	)
	if len(got) == 0 || got[0].Level != LevelRed {
		t.Fatalf("expected the red change first, got %v", kinds(got))
	}
}

func TestDiffCarriesLabelAndAdminPath(t *testing.T) {
	acc := user(idA, admin)
	acc.AdminPath = "/admin#/sw/users/permissions/user/detail/" + idA
	got := Diff(nil, []Account{acc})
	if len(got) != 1 {
		t.Fatalf("got %d changes", len(got))
	}
	if got[0].Label != acc.Label || got[0].AdminPath != acc.AdminPath {
		t.Errorf("change = %+v, want label and admin path carried over", got[0])
	}
}

// Every declared change kind needs a level, otherwise it silently degrades to
// info and a real finding would never escalate.
func TestEveryChangeKindHasALevel(t *testing.T) {
	all := []ChangeKind{
		ChangeNewAdminUser, ChangeNewAdminIntegration, ChangeAdminGranted,
		ChangeNewUser, ChangeNewIntegration, ChangeRolesAdded, ChangeRolePrivilegesGrown,
		ChangeAccessKeysAdded, ChangeReactivated, ChangePasswordRecoveryOpened,
		ChangeAdminRevoked, ChangeRolesRemoved, ChangeRolePrivilegesShrunk,
		ChangeRolePrivilegesChanged,
		ChangeAccessKeysRemoved, ChangeDeactivated, ChangeDeleted,
		ChangeNewManagedIntegration, ChangeNewACLRole, ChangePasswordChanged,
		ChangePasswordRecoveryClosed,
	}
	for _, k := range all {
		if _, ok := levels[k]; !ok {
			t.Errorf("change kind %q has no level", k)
		}
	}
	if len(levels) != len(all) {
		t.Errorf("levels has %d entries, the test knows %d -- keep them in sync", len(levels), len(all))
	}
	if ChangeKind("something_new").Level() != LevelInfo {
		t.Error("an unknown change kind must default to info, never to an alert")
	}
}

func TestMessageKey(t *testing.T) {
	cases := map[ChangeKind]string{
		ChangeNewAdminUser:           "check.account.newAdminUser",
		ChangeAdminGranted:           "check.account.adminGranted",
		ChangeRolePrivilegesGrown:    "check.account.rolePrivilegesGrown",
		ChangePasswordRecoveryOpened: "check.account.passwordRecoveryOpened",
		ChangeDeleted:                "check.account.deleted",
	}
	for change, want := range cases {
		if got := MessageKey(string(change)); got != want {
			t.Errorf("MessageKey(%q) = %q, want %q", change, got, want)
		}
	}
}

func TestCheckID(t *testing.T) {
	got := CheckID(string(KindUser), idA, string(ChangeNewAdminUser))
	want := "account.user." + idA + ".new_admin_user"
	if got != want {
		t.Errorf("CheckID = %q, want %q", got, want)
	}
}
