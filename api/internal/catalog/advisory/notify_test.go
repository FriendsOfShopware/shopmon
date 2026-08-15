package advisory

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/friendsofshopware/shopmon/api/internal/notify"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/sbom"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/shyim/go-mailer/mailertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdvisoryPublicIDPrefersCVE(t *testing.T) {
	cve := "CVE-2026-1234"
	ghsa := "GHSA-xxxx"
	assert.Equal(t, "CVE-2026-1234", advisoryPublicID(queries.ComposerAdvisory{
		AdvisoryID: "PKSA-1", Cve: &cve, GhsaID: &ghsa,
	}))
	assert.Equal(t, "GHSA-xxxx", advisoryPublicID(queries.ComposerAdvisory{
		AdvisoryID: "PKSA-1", GhsaID: &ghsa,
	}))
	assert.Equal(t, "PKSA-1", advisoryPublicID(queries.ComposerAdvisory{AdvisoryID: "PKSA-1"}))
}

func TestDisplaySeverityUsesOverride(t *testing.T) {
	high := "high"
	critical := "CRITICAL"
	assert.Equal(t, "Critical", displaySeverity(queries.ComposerAdvisory{
		Severity: &high, SeverityOverride: &critical,
	}))
	assert.Equal(t, "High", displaySeverity(queries.ComposerAdvisory{Severity: &high}))
	assert.Equal(t, "Unknown", displaySeverity(queries.ComposerAdvisory{}))
}

func TestRecommendedFixPrefersAdminUpgrade(t *testing.T) {
	upgrade := "6.7.10.1"
	row := queries.ComposerAdvisory{
		RecommendedUpgrade:   &upgrade,
		FirstPatchedVersions: json.RawMessage(`{"6.6":"6.6.10.18","6.7":"6.7.9.0"}`),
	}
	assert.Equal(t, "6.7.10.1", recommendedFix(row, "6.7.8.0"))

	row.RecommendedUpgrade = nil
	assert.Equal(t, "6.7.9.0", recommendedFix(row, "6.7.8.0"))
	assert.Equal(t, "6.6.10.18", recommendedFix(row, "6.6.1.0"))
}

func TestAdvisoryReasonIncludesPackageAndFix(t *testing.T) {
	cve := "CVE-2026-1234"
	high := "high"
	upgrade := "1.2.4"
	reason := advisoryReason(queries.ComposerAdvisory{
		AdvisoryID:         "CVE-2026-1234",
		Title:              "SSE buffer grows unbounded",
		Cve:                &cve,
		Severity:           &high,
		RecommendedUpgrade: &upgrade,
	}, sbom.Match{
		AdvisoryID:       "CVE-2026-1234",
		PackageName:      "mcp/sdk",
		InstalledVersion: "1.2.3",
	})

	assert.Equal(t, "red", reason.Level)
	assert.Equal(t, "check.security.advisoryAlert", reason.Key)
	assert.Equal(t, "High", reason.Params["severity"])
	assert.Equal(t, "CVE-2026-1234", reason.Params["id"])
	assert.Equal(t, "mcp/sdk", reason.Params["package"])
	assert.Equal(t, "1.2.3", reason.Params["installedVersion"])
	assert.Equal(t, "1.2.4", reason.Params["recommended"])
}

func TestAdvisoryAlertLink(t *testing.T) {
	single := advisoryAlertLink(9, []string{"CVE-2026-1"})
	assert.Equal(t, "account.advisories.detail", single.Name)
	assert.Equal(t, "CVE-2026-1", single.Params["id"])

	burst := advisoryAlertLink(9, []string{"CVE-1", "CVE-2"})
	assert.Equal(t, "account.environments.detail", burst.Name)
	assert.Equal(t, "9", burst.Params["environmentId"])
}

func TestNotifyNewAdvisoriesEmailIncludesDetails(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()

	env.SeedUser(t, "user-1", "Shyim", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Org One", "org-one", "user-1")
	shopID := env.SeedShop(t, "org-1", "Demo")
	envID := int32(env.SeedEnvironment(t, "org-1", shopID, "Production", "https://example.com"))

	require.NoError(t, env.Queries.SubscribeEnvironment(ctx, queries.SubscribeEnvironmentParams{
		UserID:  "user-1",
		ScopeID: strconv.Itoa(int(envID)),
	}))

	_, err := env.Pool.Exec(ctx, `
		INSERT INTO composer_advisory (advisory_id, title, cve, ghsa_id, severity, recommended_upgrade, is_visible)
		VALUES ('CVE-2026-1234',
			'Client HttpTransport SSE buffer grows unbounded when server withholds the event delimiter',
			'CVE-2026-1234', 'GHSA-test-sse', 'high', '1.2.4', true)
	`)
	require.NoError(t, err)

	_, err = env.Pool.Exec(ctx, `
		INSERT INTO environment_sbom_component (environment_id, package_name, version)
		VALUES ($1, 'mcp/sdk', '1.2.3')
	`, envID)
	require.NoError(t, err)

	rec := mailertest.NewRecordingTransport("")
	mailSender, err := mail.NewServiceWithTransport(rec, "noreply@shopmon.test", "", "https://app.shopmon.test")
	require.NoError(t, err)

	svc := &Service{
		pool:     env.Pool,
		queries:  env.Queries,
		notifier: notify.NewDispatcher(env.Queries, mailSender),
	}
	index := sbom.NewIndex([]sbom.AdvisoryPackage{{
		AdvisoryID:       "CVE-2026-1234",
		PackageName:      "mcp/sdk",
		AffectedVersions: ">=1.0.0,<1.2.4",
	}})

	count, err := svc.rematchEnvironment(ctx, envID, index)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	mailertest.AssertEmailCount(t, rec, 1)
	sent, ok := rec.Last()
	require.True(t, ok)
	body := unfoldQuotedPrintable(sent.Bytes())

	assert.Contains(t, body, "1_new_security_advisory_affects")
	assert.Contains(t, body, "1 new advisory affects this environment")
	assert.NotContains(t, body, "1 new advisories")
	assert.Contains(t, body, "High")
	assert.Contains(t, body, "CVE-2026-1234")
	assert.Contains(t, body, "mcp/sdk")
	assert.Contains(t, body, "1.2.3")
	assert.Contains(t, body, "1.2.4")
	assert.Contains(t, body, "View advisory")
	assert.Contains(t, body, "https://app.shopmon.test/app/advisories/CVE-2026-1234")

	notified, err := env.Queries.ListNotifiedAdvisoryIDsForEnvironment(ctx, envID)
	require.NoError(t, err)
	assert.Equal(t, []string{"CVE-2026-1234"}, notified)
}

func TestNotifyNewAdvisoriesEmailPluralLinksToEnvironment(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()

	env.SeedUser(t, "user-1", "Shyim", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Org One", "org-one", "user-1")
	shopID := env.SeedShop(t, "org-1", "Demo")
	envID := int32(env.SeedEnvironment(t, "org-1", shopID, "Production", "https://example.com"))

	require.NoError(t, env.Queries.SubscribeEnvironment(ctx, queries.SubscribeEnvironmentParams{
		UserID:  "user-1",
		ScopeID: strconv.Itoa(int(envID)),
	}))

	_, err := env.Pool.Exec(ctx, `
		INSERT INTO composer_advisory (advisory_id, title, cve, severity, is_visible)
		VALUES
			('CVE-2026-1', 'First issue', 'CVE-2026-1', 'critical', true),
			('CVE-2026-2', 'Second issue', 'CVE-2026-2', 'medium', true)
	`)
	require.NoError(t, err)

	_, err = env.Pool.Exec(ctx, `
		INSERT INTO environment_sbom_component (environment_id, package_name, version)
		VALUES ($1, 'shopware/core', '6.7.8.0')
	`, envID)
	require.NoError(t, err)

	rec := mailertest.NewRecordingTransport("")
	mailSender, err := mail.NewServiceWithTransport(rec, "noreply@shopmon.test", "", "https://app.shopmon.test")
	require.NoError(t, err)

	svc := &Service{
		pool:     env.Pool,
		queries:  env.Queries,
		notifier: notify.NewDispatcher(env.Queries, mailSender),
	}
	index := sbom.NewIndex([]sbom.AdvisoryPackage{
		{AdvisoryID: "CVE-2026-1", PackageName: "shopware/core", AffectedVersions: ">=6.7.0.0,<6.7.10.1"},
		{AdvisoryID: "CVE-2026-2", PackageName: "shopware/core", AffectedVersions: ">=6.7.0.0,<6.7.9.0"},
	})

	count, err := svc.rematchEnvironment(ctx, envID, index)
	require.NoError(t, err)
	require.Equal(t, 2, count)

	mailertest.AssertEmailCount(t, rec, 1)
	sent, ok := rec.Last()
	require.True(t, ok)
	body := unfoldQuotedPrintable(sent.Bytes())

	assert.Contains(t, body, "2_new_security_advisories_affect")
	assert.Contains(t, body, "First issue")
	assert.Contains(t, body, "Second issue")
	assert.Contains(t, body, "Critical")
	assert.Contains(t, body, "Medium")
	assert.Contains(t, body, "View environment")
	assert.Contains(t, body, "https://app.shopmon.test/app/environments/"+strconv.Itoa(int(envID)))
}

// unfoldQuotedPrintable strips MIME quoted-printable soft line breaks so
// assertions can match readable phrases in a captured raw message.
func unfoldQuotedPrintable(raw []byte) string {
	s := string(raw)
	s = strings.ReplaceAll(s, "=\r\n", "")
	s = strings.ReplaceAll(s, "=\n", "")
	return s
}