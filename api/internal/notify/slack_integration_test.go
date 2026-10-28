package notify_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/crypto"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/notify"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The webhook lookups are SQL joins (environment -> shop, organization), so
// they are exercised against a real database rather than a fake store.
func TestSlackDispatchResolvesOrganizationAndShopWebhooks(t *testing.T) {
	env := testutil.Setup(t)
	env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopA := env.SeedShop(t, "org-1", "Shop A")
	shopB := env.SeedShop(t, "org-1", "Shop B")
	envA := env.SeedEnvironment(t, "org-1", shopA, "Production", "https://a.example.test")
	envB := env.SeedEnvironment(t, "org-1", shopB, "Production", "https://b.example.test")

	encrypt := func(url string) string {
		e, err := crypto.Encrypt(url, env.Cfg.AppSecret)
		require.NoError(t, err)
		return e
	}
	ctx := context.Background()
	require.NoError(t, env.Queries.UpsertOrganizationSlackWebhook(ctx, queries.UpsertOrganizationSlackWebhookParams{
		OrganizationID: "org-1", WebhookUrl: encrypt("https://hooks.slack.com/services/org"),
	}))
	require.NoError(t, env.Queries.UpsertShopSlackWebhook(ctx, queries.UpsertShopSlackWebhookParams{
		ShopID: int32(shopA), WebhookUrl: encrypt("https://hooks.slack.com/services/shop-a"),
	}))

	poster := &testutil.SlackRecorder{}
	d := notify.NewDispatcher(env.Queries, nil, notify.SlackConfig{
		AppSecret: env.Cfg.AppSecret, FrontendURL: "https://shopmon.example.test", Poster: poster,
	})

	event := func(envID int) notify.Event {
		return notify.Event{
			Type:       notify.EventStatusDegraded,
			Level:      notify.LevelWarning,
			ScopeType:  notify.ScopeEnvironment,
			ScopeID:    strconv.Itoa(envID),
			OrgID:      "org-1",
			DedupKey:   "environment.change-status." + strconv.Itoa(envID),
			TitleKey:   "notification.statusDegraded.title",
			MessageKey: "notification.statusDegraded.message",
			Params:     map[string]any{"name": "Shop", "from": "green", "to": "red"},
		}
	}

	res := d.Dispatch(ctx, event(envA), nil)
	assert.Equal(t, 2, res.Delivered)
	require.Len(t, poster.Posts, 2)
	assert.Equal(t, "https://hooks.slack.com/services/org", poster.Posts[0].URL)
	assert.Equal(t, "https://hooks.slack.com/services/shop-a", poster.Posts[1].URL)

	// Shop B has no webhook of its own: only the organization channel.
	res = d.Dispatch(ctx, event(envB), nil)
	assert.Equal(t, 1, res.Delivered)
	require.Len(t, poster.Posts, 3)
	assert.Equal(t, "https://hooks.slack.com/services/org", poster.Posts[2].URL)
}

// The failed post must release its row in the lock table, so a retry within
// the dedup window still reaches Slack.
func TestSlackFailedPostReleasesLockInDatabase(t *testing.T) {
	env := testutil.Setup(t)
	env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")

	ctx := context.Background()
	encrypted, err := crypto.Encrypt("https://hooks.slack.com/services/org", env.Cfg.AppSecret)
	require.NoError(t, err)
	require.NoError(t, env.Queries.UpsertOrganizationSlackWebhook(ctx, queries.UpsertOrganizationSlackWebhookParams{
		OrganizationID: "org-1", WebhookUrl: encrypted,
	}))

	poster := &testutil.SlackRecorder{Err: assert.AnError}
	d := notify.NewDispatcher(env.Queries, nil, notify.SlackConfig{AppSecret: env.Cfg.AppSecret, Poster: poster})
	ev := notify.Event{
		Type:       notify.EventAuthError,
		Level:      notify.LevelError,
		OrgID:      "org-1",
		DedupKey:   "environment.update-auth-error.1",
		TitleKey:   "notification.authError.title",
		MessageKey: "notification.authError.message",
	}

	res := d.Dispatch(ctx, ev, nil)
	require.Equal(t, 1, res.Failed)

	poster.Err = nil
	res = d.Dispatch(ctx, ev, nil)
	assert.Equal(t, 1, res.Delivered)
	assert.Len(t, poster.Posts, 1)
}
