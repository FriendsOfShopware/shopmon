// Package accounts reads the admin users, integrations and ACL roles a shop
// exposes through the FroshTools security endpoint.
//
// Shopmon does not keep a readable directory of who has access to a shop. The
// snapshot this package produces exists to be diffed against a stored baseline
// (see diff.go), so Shopmon can report *that* something changed and how
// dangerous that change is. Everything here therefore errs toward storing less:
// labels are re-masked locally, deep links are validated, and the payload is
// bounded before it is parsed.
package accounts

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/friendsofshopware/shopmon/api/internal/shopware"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/sanitize"
)

// Path is the FroshTools account endpoint. It is served by newer FroshTools
// releases only; older plugins answer 404 and callers must treat that as
// "unsupported", not as a scrape failure.
const Path = "/_action/frosh-tools/security/accounts"

const (
	// maxDocumentBytes caps the payload. The document is shop-controlled and
	// the admin client reads the body unbounded, so a hostile or runaway
	// response could otherwise allocate freely here and again while flattening.
	// A real response is a few KB.
	maxDocumentBytes = 4 << 20
	// maxEntries caps each of the three lists. A shop with more accounts than
	// this is not a plausible response, and the list drives per-row writes.
	maxEntries = 5000
	// maxLabelLen bounds the stored label. Masked emails and integration names
	// are short.
	maxLabelLen = 190
	// maxAdminPathLen bounds the admin deep link.
	maxAdminPathLen = 512
)

// Fetcher is the subset of the Shopware admin client this package needs.
type Fetcher interface {
	Get(ctx context.Context, path string) ([]byte, error)
}

// Kind distinguishes the three account types the endpoint reports. They share
// one Go type (and one database table) because they differ only in which
// fields are meaningful.
type Kind string

const (
	KindUser        Kind = "user"
	KindIntegration Kind = "integration"
	KindACLRole     Kind = "acl_role"
)

// Account is one monitored principal. Which fields carry meaning depends on
// Kind: roles have no access keys, users have no app name.
type Account struct {
	Kind  Kind
	ID    string
	Label string

	Active bool
	Admin  bool

	// RoleIDs is sorted and deduplicated so RolesHash is stable regardless of
	// the order the shop returns.
	RoleIDs   []string
	RolesHash string

	AccessKeyCount int32

	// PrivilegeCount and PrivilegesHash are reported for ACL roles only. The
	// hash is what detects a role being granted additional privileges without
	// its count changing.
	PrivilegeCount int32
	PrivilegesHash string

	ManagedByApp bool
	AppName      string

	PasswordRecoveryPending bool

	LastUsageAt       *time.Time
	PasswordChangedAt *time.Time
	RemoteCreatedAt   *time.Time
	RemoteUpdatedAt   *time.Time

	// AdminPath is a shop-relative deep link, already validated. Empty when the
	// shop supplied nothing usable.
	AdminPath string
}

// Snapshot is one reading of a shop's accounts.
type Snapshot struct {
	// GeneratedAt is the shop's own timestamp. It is untrusted (clock skew,
	// forgeable) and used for display only — detection always stamps server
	// time.
	GeneratedAt  time.Time
	HasGenerated bool

	// Accounts holds all three kinds, sorted by (Kind, ID) so callers get a
	// stable order.
	Accounts []Account
}

// ErrUnsupported reports that the shop's FroshTools does not expose the account
// endpoint. Callers record this and skip ingestion for the shop rather than
// failing the whole scrape.
type ErrUnsupported struct {
	Reason string
}

func (e *ErrUnsupported) Error() string {
	return fmt.Sprintf("accounts endpoint unsupported: %s", e.Reason)
}

// Fetch retrieves and parses the account snapshot for one shop.
func Fetch(ctx context.Context, client Fetcher) (*Snapshot, error) {
	if client == nil {
		return nil, &ErrUnsupported{Reason: "no client"}
	}

	raw, err := client.Get(ctx, Path)
	if err != nil {
		// Only a missing route means "this shop cannot serve accounts". A 403
		// means the integration lacks the frosh_tools:read ACL — a fixable
		// configuration problem that must surface as an error with last_error,
		// not be filed away as a permanent capability gap nobody is told about.
		if shopware.ClassifyFailure(err).Class == shopware.FailureNotFound {
			return nil, &ErrUnsupported{Reason: "endpoint not found"}
		}
		return nil, fmt.Errorf("fetch accounts: %w", err)
	}

	return Parse(raw)
}

// wire mirrors the endpoint payload. Only the fields Shopmon consumes are
// modelled; unknown fields are ignored so a newer plugin stays compatible.
type wire struct {
	GeneratedAt  string            `json:"generatedAt"`
	Users        []wireUser        `json:"users"`
	Integrations []wireIntegration `json:"integrations"`
	AclRoles     []wireACLRole     `json:"aclRoles"`
}

type wireUser struct {
	ID                      string   `json:"id"`
	MaskedEmail             string   `json:"maskedEmail"`
	Active                  bool     `json:"active"`
	Admin                   bool     `json:"admin"`
	AclRoleIDs              []string `json:"aclRoleIds"`
	AccessKeyCount          int      `json:"accessKeyCount"`
	AccessKeyLastUsageAt    *string  `json:"accessKeyLastUsageAt"`
	PasswordRecoveryPending bool     `json:"passwordRecoveryPending"`
	LastUpdatedPasswordAt   *string  `json:"lastUpdatedPasswordAt"`
	CreatedAt               *string  `json:"createdAt"`
	UpdatedAt               *string  `json:"updatedAt"`
	AdminPath               string   `json:"adminPath"`
}

type wireIntegration struct {
	ID           string   `json:"id"`
	Label        string   `json:"label"`
	Admin        bool     `json:"admin"`
	AclRoleIDs   []string `json:"aclRoleIds"`
	ManagedByApp bool     `json:"managedByApp"`
	AppName      *string  `json:"appName"`
	LastUsageAt  *string  `json:"lastUsageAt"`
	DeletedAt    *string  `json:"deletedAt"`
	CreatedAt    *string  `json:"createdAt"`
	UpdatedAt    *string  `json:"updatedAt"`
	AdminPath    string   `json:"adminPath"`
}

type wireACLRole struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	PrivilegeCount int     `json:"privilegeCount"`
	PrivilegesHash string  `json:"privilegesHash"`
	DeletedAt      *string `json:"deletedAt"`
	CreatedAt      *string `json:"createdAt"`
	UpdatedAt      *string `json:"updatedAt"`
}

// Parse decodes and normalizes an account payload.
func Parse(raw []byte) (*Snapshot, error) {
	if len(raw) > maxDocumentBytes {
		return nil, fmt.Errorf("parse accounts: payload too large (%d bytes)", len(raw))
	}

	var doc wire
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse accounts: %w", err)
	}

	// Silently keeping the first N would leave the rest unwatched, which is
	// indistinguishable from "those accounts do not exist" and would report
	// them as deleted. Refusing keeps the previous baseline instead.
	if err := checkLimits(doc); err != nil {
		return nil, err
	}

	snap := &Snapshot{}
	if ts, ok := parseTime(&doc.GeneratedAt); ok {
		snap.GeneratedAt, snap.HasGenerated = *ts, true
	}

	out := make([]Account, 0, len(doc.Users)+len(doc.Integrations)+len(doc.AclRoles))

	for _, u := range doc.Users {
		id := normalizeID(u.ID)
		if id == "" {
			continue
		}
		roles, hash := normalizeRoleIDs(u.AclRoleIDs)
		lastUsage, _ := parseTime(u.AccessKeyLastUsageAt)
		pwChanged, _ := parseTime(u.LastUpdatedPasswordAt)
		created, _ := parseTime(u.CreatedAt)
		updated, _ := parseTime(u.UpdatedAt)
		out = append(out, Account{
			Kind:                    KindUser,
			ID:                      id,
			Label:                   MaskLabel(KindUser, u.MaskedEmail),
			Active:                  u.Active,
			Admin:                   u.Admin,
			RoleIDs:                 roles,
			RolesHash:               hash,
			AccessKeyCount:          clampCount(u.AccessKeyCount),
			PasswordRecoveryPending: u.PasswordRecoveryPending,
			LastUsageAt:             lastUsage,
			PasswordChangedAt:       pwChanged,
			RemoteCreatedAt:         created,
			RemoteUpdatedAt:         updated,
			AdminPath:               SanitizeAdminPath(u.AdminPath),
		})
	}

	for _, i := range doc.Integrations {
		id := normalizeID(i.ID)
		if id == "" {
			continue
		}
		roles, hash := normalizeRoleIDs(i.AclRoleIDs)
		lastUsage, _ := parseTime(i.LastUsageAt)
		// A soft-deleted integration is gone as far as access is concerned.
		// Dropping it here prunes it from the baseline, which yields exactly
		// one "deleted" event -- keeping it would re-emit that event on every
		// scrape for as long as the shop keeps returning the row.
		if _, deleted := parseTime(i.DeletedAt); deleted {
			continue
		}
		created, _ := parseTime(i.CreatedAt)
		updated, _ := parseTime(i.UpdatedAt)
		appName := ""
		if i.AppName != nil {
			appName = sanitize.Meta(*i.AppName)
		}
		out = append(out, Account{
			Kind:            KindIntegration,
			ID:              id,
			Label:           MaskLabel(KindIntegration, i.Label),
			Active:          true,
			Admin:           i.Admin,
			RoleIDs:         roles,
			RolesHash:       hash,
			ManagedByApp:    i.ManagedByApp,
			AppName:         appName,
			LastUsageAt:     lastUsage,
			RemoteCreatedAt: created,
			RemoteUpdatedAt: updated,
			AdminPath:       SanitizeAdminPath(i.AdminPath),
		})
	}

	for _, r := range doc.AclRoles {
		id := normalizeID(r.ID)
		if id == "" {
			continue
		}
		if _, deleted := parseTime(r.DeletedAt); deleted {
			continue
		}
		created, _ := parseTime(r.CreatedAt)
		updated, _ := parseTime(r.UpdatedAt)
		out = append(out, Account{
			Kind:            KindACLRole,
			ID:              id,
			Label:           MaskLabel(KindACLRole, r.Name),
			Active:          true,
			PrivilegeCount:  clampCount(r.PrivilegeCount),
			PrivilegesHash:  sanitize.Meta(r.PrivilegesHash),
			RemoteCreatedAt: created,
			RemoteUpdatedAt: updated,
		})
	}

	// A Shopware shop always has at least one administrator, so a response that
	// yields no usable user is truncated or filtered. Counted after parsing,
	// not on the wire entries: rows with an unusable id are dropped below, and
	// a response of nothing but those would otherwise pass the check and then
	// prune every stored user from the baseline — after which the next
	// successful run reports each of them as brand new.
	if !hasKind(out, KindUser) {
		return nil, fmt.Errorf("parse accounts: response contained no usable users")
	}

	sort.Slice(out, func(a, b int) bool {
		if out[a].Kind != out[b].Kind {
			return out[a].Kind < out[b].Kind
		}
		return out[a].ID < out[b].ID
	})
	snap.Accounts = out
	return snap, nil
}

func hasKind(in []Account, kind Kind) bool {
	for _, a := range in {
		if a.Kind == kind {
			return true
		}
	}
	return false
}

// checkLimits refuses an implausibly large response rather than trimming it.
func checkLimits(doc wire) error {
	for _, l := range []struct {
		name string
		n    int
	}{
		{"users", len(doc.Users)},
		{"integrations", len(doc.Integrations)},
		{"aclRoles", len(doc.AclRoles)},
	} {
		if l.n > maxEntries {
			return fmt.Errorf("parse accounts: %s list too large (%d entries, limit %d)", l.name, l.n, maxEntries)
		}
	}

	// Per-account role lists get the same treatment. Trimming one would leave
	// the dropped roles out of the hash and the holder count, so granting such
	// a role would raise no finding at all.
	for _, u := range doc.Users {
		if len(u.AclRoleIDs) > maxEntries {
			return fmt.Errorf("parse accounts: user role list too large (%d entries, limit %d)", len(u.AclRoleIDs), maxEntries)
		}
	}
	for _, i := range doc.Integrations {
		if len(i.AclRoleIDs) > maxEntries {
			return fmt.Errorf("parse accounts: integration role list too large (%d entries, limit %d)", len(i.AclRoleIDs), maxEntries)
		}
	}
	return nil
}

// clampCount keeps a shop-supplied counter inside int32 and non-negative, so a
// nonsense value cannot overflow the column or invert a "grew" comparison.
func clampCount(n int) int32 {
	switch {
	case n < 0:
		return 0
	case n > 1_000_000:
		return 1_000_000
	default:
		return int32(n)
	}
}

// normalizeID returns a stable, filesystem- and check-id-safe handle for an
// account. Shopware ids are 32 hex characters; anything else is hashed so the
// account stays trackable without letting shop-controlled text into a check id.
func normalizeID(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	if s == "" {
		return ""
	}
	if len(s) == 32 && isHex(s) {
		return s
	}
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:32]
}

func isHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// normalizeRoleIDs sorts and deduplicates the assigned role ids and returns a
// hash over them. The diff compares the hash, so the shop's ordering cannot
// produce a phantom "roles changed".
func normalizeRoleIDs(in []string) ([]string, string) {
	if len(in) == 0 {
		return nil, ""
	}
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, raw := range in {
		id := normalizeID(raw)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	if len(out) == 0 {
		return nil, ""
	}
	sort.Strings(out)
	sum := sha256.Sum256([]byte(strings.Join(out, ",")))
	return out, hex.EncodeToString(sum[:])
}

// MaskLabel bounds and, for user accounts, masks the label before it is stored.
//
// FroshTools already masks user emails, but Shopmon must not depend on that:
// an older plugin, a patched build or a proxy could return a plaintext address,
// and persisting other people's email addresses is exactly what this feature
// avoids. Integration and role names are not personal data and pass through
// unmasked.
//
// User labels are masked unconditionally rather than only when they look
// unmasked. Accepting anything that merely contains a '*' would let a partially
// masked value such as "lisa.mueller@e***.com" through with the person's full
// name intact -- into the baseline, the events and the notification text.
// Re-masking an already-masked value is idempotent, so there is nothing to gain
// from trusting the input.
func MaskLabel(kind Kind, raw string) string {
	s := sanitize.Meta(raw)
	if s == "" {
		return ""
	}
	if utf8.RuneCountInString(s) > maxLabelLen {
		s = string([]rune(s)[:maxLabelLen])
	}
	if kind != KindUser {
		return s
	}
	return maskEmail(s)
}

// maskEmail reduces an address to its first local character and the first
// character of each domain segment: "lisa@vier.de" -> "l***@v***.de". Applying
// it to an already masked value yields the same string, which is what makes
// masking unconditionally safe.
func maskEmail(s string) string {
	at := strings.LastIndex(s, "@")
	if at <= 0 || at == len(s)-1 {
		// Not an address — keep only the first character rather than guessing.
		return maskSegment(s)
	}
	local, domain := s[:at], s[at+1:]
	parts := strings.Split(domain, ".")
	for i, p := range parts {
		// The last label is a public suffix and carries no identity, so it
		// stays readable -- unless it is the only label, in which case it is
		// the organisation's own name ("jane@acmecorp") and must be masked.
		if p == "" || (i == len(parts)-1 && len(parts) > 1) {
			continue
		}
		parts[i] = maskSegment(p)
	}
	return maskSegment(local) + "@" + strings.Join(parts, ".")
}

// maskSegment keeps the first character and replaces the rest with a fixed
// run, so the original length is not leaked either.
func maskSegment(s string) string {
	first := firstRune(s)
	if first == "" {
		return ""
	}
	return first + "***"
}

func firstRune(s string) string {
	for _, r := range s {
		return string(r)
	}
	return ""
}

// SanitizeAdminPath accepts only shop-relative admin paths.
//
// The value is shop-controlled and Shopmon renders it as a link the operator is
// invited to click. A protocol-relative "//evil.example.com" or a "javascript:"
// value would turn a security finding into an attack. Anything that is not a
// single-slash-rooted relative path is dropped; the finding then shows without
// a link.
func SanitizeAdminPath(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" || len(s) > maxAdminPathLen || !utf8.ValidString(s) {
		return ""
	}
	if !strings.HasPrefix(s, "/") || strings.HasPrefix(s, "//") {
		return ""
	}
	if strings.ContainsAny(s, "\\ \t\r\n") {
		return ""
	}
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return ""
		}
	}
	return s
}

// parseTime reads an RFC3339 timestamp and normalizes it to UTC. Shop
// timestamps carry an offset; the database columns are naive.
func parseTime(raw *string) (*time.Time, bool) {
	if raw == nil {
		return nil, false
	}
	s := strings.TrimSpace(*raw)
	if s == "" {
		return nil, false
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, false
	}
	utc := t.UTC()
	return &utc, true
}
