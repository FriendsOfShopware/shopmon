package accounts

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/shopware"
)

type stubFetcher struct {
	body []byte
	err  error
}

func (s stubFetcher) Get(context.Context, string) ([]byte, error) { return s.body, s.err }

func loadFixture(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile("testdata/accounts.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return raw
}

func TestParseRealResponse(t *testing.T) {
	snap, err := Parse(loadFixture(t))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if !snap.HasGenerated {
		t.Fatal("expected generatedAt to be parsed")
	}
	if got := snap.GeneratedAt.Format("2006-01-02T15:04:05Z"); got != "2026-10-29T07:17:37Z" {
		t.Errorf("generatedAt = %q", got)
	}

	var users, integrations, roles int
	for _, a := range snap.Accounts {
		switch a.Kind {
		case KindUser:
			users++
		case KindIntegration:
			integrations++
		case KindACLRole:
			roles++
		}
	}
	if users != 8 || integrations != 7 || roles != 2 {
		t.Fatalf("users=%d integrations=%d roles=%d, want 8/7/2", users, integrations, roles)
	}

	// Accounts arrive sorted by (kind, id) so a caller never sees shop ordering.
	for i := 1; i < len(snap.Accounts); i++ {
		prev, cur := snap.Accounts[i-1], snap.Accounts[i]
		if prev.Kind > cur.Kind || (prev.Kind == cur.Kind && prev.ID >= cur.ID) {
			t.Fatalf("accounts not sorted at %d: %v/%v then %v/%v", i, prev.Kind, prev.ID, cur.Kind, cur.ID)
		}
	}

	admin := find(t, snap, KindUser, "d53511fba8f02d164c2461efba474c96")
	if !admin.Admin || !admin.Active {
		t.Errorf("expected active admin, got admin=%v active=%v", admin.Admin, admin.Active)
	}
	if admin.AccessKeyCount != 2 {
		t.Errorf("accessKeyCount = %d, want 2", admin.AccessKeyCount)
	}
	if admin.Label != "c***@e***.com" {
		t.Errorf("label = %q, want the masked address unchanged", admin.Label)
	}
	if admin.AdminPath != "/admin#/sw/users/permissions/user/detail/d53511fba8f02d164c2461efba474c96" {
		t.Errorf("adminPath = %q", admin.AdminPath)
	}

	editor := find(t, snap, KindUser, "c8f5a30d8830761d14c6082be10f487e")
	if len(editor.RoleIDs) != 1 || editor.RoleIDs[0] != "1a5cc1d6bd912ee8417e271356ee30a6" {
		t.Errorf("roleIDs = %v", editor.RoleIDs)
	}
	if editor.RolesHash == "" {
		t.Error("expected a roles hash for an account with roles")
	}

	app := find(t, snap, KindIntegration, "059ffd06c0ca0e360bc7228f478c75d3")
	if !app.ManagedByApp || app.AppName != "MenuApp" {
		t.Errorf("managedByApp=%v appName=%q", app.ManagedByApp, app.AppName)
	}

	role := find(t, snap, KindACLRole, "1a5cc1d6bd912ee8417e271356ee30a6")
	if role.PrivilegeCount != 286 || role.PrivilegesHash == "" {
		t.Errorf("privilegeCount=%d hash=%q", role.PrivilegeCount, role.PrivilegesHash)
	}
	if role.Label != "Editor" {
		t.Errorf("role label = %q, role names are not personal data and must pass through", role.Label)
	}
}

func find(t *testing.T, snap *Snapshot, kind Kind, id string) Account {
	t.Helper()
	for _, a := range snap.Accounts {
		if a.Kind == kind && a.ID == id {
			return a
		}
	}
	t.Fatalf("account %s/%s not found", kind, id)
	return Account{}
}

// Only a missing route means "this shop cannot serve accounts". A 403 is a
// fixable permission problem that has to stay visible.
func TestFetchTreatsOnly404AsUnsupported(t *testing.T) {
	cases := []struct {
		name        string
		status      int
		body        string
		unsupported bool
	}{
		{name: "not found", status: http.StatusNotFound, unsupported: true},
		{
			name:   "missing privilege",
			status: http.StatusForbidden,
			body:   `{"errors":[{"code":"FRAMEWORK__MISSING_PRIVILEGE_ERROR","detail":"{\"message\":\"Missing privilege\",\"missingPrivileges\":[\"frosh_tools:read\"]}"}]}`,
		},
		{name: "server error", status: http.StatusInternalServerError},
		{name: "unauthorized", status: http.StatusUnauthorized},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Fetch(context.Background(), stubFetcher{
				err: &shopware.ApiError{StatusCode: tc.status, Body: tc.body},
			})
			if err == nil {
				t.Fatal("expected an error")
			}
			var unsupported *ErrUnsupported
			if got := errors.As(err, &unsupported); got != tc.unsupported {
				t.Fatalf("ErrUnsupported = %v, want %v (err: %v)", got, tc.unsupported, err)
			}
		})
	}
}

func TestFetchWithoutClientIsUnsupported(t *testing.T) {
	if _, err := Fetch(context.Background(), nil); err == nil {
		t.Fatal("expected an error")
	} else {
		var unsupported *ErrUnsupported
		if !errors.As(err, &unsupported) {
			t.Fatalf("want ErrUnsupported, got %v", err)
		}
	}
}

func TestParseRejectsOversizedPayload(t *testing.T) {
	if _, err := Parse(make([]byte, maxDocumentBytes+1)); err == nil {
		t.Fatal("expected oversized payload to be rejected")
	}
}

func TestParseRejectsOversizedLists(t *testing.T) {
	var b strings.Builder
	b.WriteString(`{"users":[`)
	for i := 0; i < maxEntries+10; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(`{"id":"user-`)
		b.WriteString(itoa(i))
		b.WriteString(`","maskedEmail":"a***@b***.de"}`)
	}
	b.WriteString(`]}`)

	// Trimming would leave the remaining accounts unwatched, which is
	// indistinguishable from them not existing -- and they would be reported
	// as deleted. Refusing keeps the stored baseline.
	if _, err := Parse([]byte(b.String())); err == nil {
		t.Fatal("expected an oversized list to be rejected rather than trimmed")
	}
}

// A shop always has at least one administrator, so an absent user list is a
// truncated response. Accepting it would prune every stored user.
func TestParseRejectsResponseWithoutUsers(t *testing.T) {
	cases := map[string]string{
		"no users key": `{"integrations":[{"id":"69ba0db6351d47369ea35d6873cd115a","label":"x"}]}`,
		"empty users":  `{"users":[],"aclRoles":[{"id":"d11b7bdac63e4ff688a353ee0e3db49b","name":"r"}]}`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(raw)); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf []byte
	for i > 0 {
		buf = append([]byte{byte('0' + i%10)}, buf...)
		i /= 10
	}
	return string(buf)
}

// Shopmon must never persist a readable address, even if the shop sends one.
func TestMaskLabel(t *testing.T) {
	cases := []struct {
		name string
		kind Kind
		in   string
		want string
	}{
		{name: "already masked stays masked", kind: KindUser, in: "l***@e***.com", want: "l***@e***.com"},
		{
			// A single wildcard must not buy trust: the local part here is the
			// person's full name.
			name: "partially masked address is re-masked",
			kind: KindUser,
			in:   "lisa.mueller@e***.com",
			want: "l***@e***.com",
		},
		{name: "masked local part only", kind: KindUser, in: "l***@vierpunkt.de", want: "l***@v***.de"},
		{name: "plaintext address is masked", kind: KindUser, in: "lisa.mueller@example.com", want: "l***@e***.com"},
		{name: "subdomain is masked too", kind: KindUser, in: "bob@mail.example.com", want: "b***@m***.e***.com"},
		{name: "non address keeps one character", kind: KindUser, in: "administrator", want: "a***"},
		{
			// A single-label domain has no public suffix, so the only label is
			// the organisation's own name.
			name: "single label domain is masked",
			kind: KindUser,
			in:   "jane@acmecorp",
			want: "j***@a***",
		},
		{
			// An asterisk anywhere must not buy trust either.
			name: "asterisk in a plaintext local part",
			kind: KindUser,
			in:   "alice+*@example.com",
			want: "a***@e***.com",
		},
		{name: "integration label passes through", kind: KindIntegration, in: "Shopmon Staging", want: "Shopmon Staging"},
		{name: "role name passes through", kind: KindACLRole, in: "Editor", want: "Editor"},
		{name: "empty stays empty", kind: KindUser, in: "", want: ""},
		{name: "control characters are dropped", kind: KindUser, in: "bad\x00label", want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := MaskLabel(tc.kind, tc.in); got != tc.want {
				t.Errorf("MaskLabel(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// The deep link is shop-controlled and gets rendered as a clickable href.
func TestSanitizeAdminPath(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{in: "/admin#/sw/users/permissions/user/detail/abc", want: "/admin#/sw/users/permissions/user/detail/abc"},
		{in: "/admin#/sw/integration/index", want: "/admin#/sw/integration/index"},
		{in: "//evil.example.com/admin", want: ""},
		{in: "https://evil.example.com/admin", want: ""},
		{in: "javascript:alert(1)", want: ""},
		{in: "admin/relative", want: ""},
		{in: "/admin\\..\\x", want: ""},
		{in: "/admin\x00", want: ""},
		{in: "/admin with space", want: ""},
		{in: "", want: ""},
		{in: "/" + strings.Repeat("a", maxAdminPathLen), want: ""},
	}
	for _, tc := range cases {
		if got := SanitizeAdminPath(tc.in); got != tc.want {
			t.Errorf("SanitizeAdminPath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeID(t *testing.T) {
	const valid = "0123456789abcdef0123456789abcdef"
	if got := normalizeID(valid); got != valid {
		t.Errorf("valid id changed: %q", got)
	}
	if got := normalizeID(strings.ToUpper(valid)); got != valid {
		t.Errorf("uppercase id = %q, want lowercased", got)
	}
	if got := normalizeID(""); got != "" {
		t.Errorf("empty id = %q", got)
	}

	// Anything unexpected is hashed, so shop-controlled text never reaches a
	// check id -- but stays stable across scrapes.
	odd := normalizeID("../../etc/passwd")
	if len(odd) != 32 || !isHex(odd) {
		t.Fatalf("unexpected id form: %q", odd)
	}
	if odd != normalizeID("../../etc/passwd") {
		t.Error("hashed id is not stable")
	}
}

func TestNormalizeRoleIDsIsOrderIndependent(t *testing.T) {
	a, hashA := normalizeRoleIDs([]string{"1a5cc1d6bd912ee8417e271356ee30a6", "f2b9a06d2a0b4f0eaba155efd9854b69"})
	b, hashB := normalizeRoleIDs([]string{"f2b9a06d2a0b4f0eaba155efd9854b69", "1a5cc1d6bd912ee8417e271356ee30a6", "f2b9a06d2a0b4f0eaba155efd9854b69"})
	if hashA != hashB {
		t.Fatalf("hash differs for the same set: %q vs %q", hashA, hashB)
	}
	if len(a) != 2 || len(b) != 2 {
		t.Fatalf("expected two deduplicated roles, got %v and %v", a, b)
	}
	if _, empty := normalizeRoleIDs(nil); empty != "" {
		t.Error("no roles must hash to the empty string")
	}
}

// A soft-deleted integration is dropped at parse time, so it is pruned from the
// baseline once and yields exactly one deletion event instead of one per scrape.
func TestParseDropsSoftDeletedEntries(t *testing.T) {
	raw := `{"users":[
	  {"id":"cee011c787fb441d8751cffe0a591eb1","maskedEmail":"a***@e***.de"}
	],"integrations":[
	  {"id":"b3e4487748bf449b9b34287dc6877082","label":"gone","deletedAt":"2026-01-02T03:04:05+00:00"},
	  {"id":"69ba0db6351d47369ea35d6873cd115a","label":"here"}
	],"aclRoles":[
	  {"id":"1a5cc1d6bd912ee8417e271356ee30a6","name":"gone","deletedAt":"2026-01-02T03:04:05+00:00"}
	]}`
	snap, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	var labels []string
	for _, a := range snap.Accounts {
		labels = append(labels, a.Label)
	}
	if len(snap.Accounts) != 2 {
		t.Fatalf("accounts = %v, want the user and the live integration only", labels)
	}
	for _, a := range snap.Accounts {
		if a.Label == "gone" {
			t.Fatalf("a soft-deleted entry survived: %v", labels)
		}
	}
}

func TestParseNormalizesTimestampsToUTC(t *testing.T) {
	raw := `{"users":[{"id":"cee011c787fb441d8751cffe0a591eb1","maskedEmail":"a***@e***.de"}],` +
		`"integrations":[{"id":"69ba0db6351d47369ea35d6873cd115a","label":"x","lastUsageAt":"2026-10-29T09:00:00+02:00"}]}`
	snap, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	var got *time.Time
	for _, a := range snap.Accounts {
		if a.Kind == KindIntegration {
			got = a.LastUsageAt
		}
	}
	if got == nil || got.Format("2006-01-02T15:04:05Z07:00") != "2026-10-29T07:00:00Z" {
		t.Fatalf("lastUsageAt = %v, want the UTC equivalent", got)
	}
}
