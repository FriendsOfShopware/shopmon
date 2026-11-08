package notify

import (
	"context"
	"sync"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/crypto"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAppSecret = "0123456789abcdef0123456789abcdef"

// fakeWebhookStore returns encrypted webhooks per organization and per
// environment's shop.
type fakeWebhookStore struct {
	preferences     map[string][]queries.DestinationNotificationPreference
	preferenceError error
	webhooks        map[string]string
	shopWebhooks    map[int32]string
	lookups         int
}

func (s *fakeWebhookStore) GetEventNotificationDestinations(_ context.Context, arg queries.GetEventNotificationDestinationsParams) ([]queries.NotificationDestination, error) {
	s.lookups++
	var rows []queries.NotificationDestination
	if dsn, ok := s.webhooks[arg.OrganizationID]; ok {
		rows = append(rows, queries.NotificationDestination{ID: "org", Dsn: dsn})
	}
	if dsn, ok := s.shopWebhooks[arg.EnvironmentID]; ok {
		rows = append(rows, queries.NotificationDestination{ID: "shop", Dsn: dsn})
	}
	return rows, nil
}

type destinationPost struct {
	url  string
	text string
}

type recordingPoster struct {
	mu    sync.Mutex
	posts []destinationPost
	err   error
	// failURL, when set, fails only posts to that webhook.
	failURL string
}

func (p *recordingPoster) Post(_ context.Context, webhookURL, text string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.err != nil || (p.failURL != "" && p.failURL == webhookURL) {
		return assert.AnError
	}
	p.posts = append(p.posts, destinationPost{url: webhookURL, text: text})
	return nil
}

func encryptWebhook(t *testing.T, url string) string {
	t.Helper()
	e, err := crypto.Encrypt(url, testAppSecret)
	require.NoError(t, err)
	return e
}

func newDestinationTestDispatcher(t *testing.T, poster *recordingPoster, webhooks map[string]string) (*Dispatcher, *fakeWebhookStore) {
	t.Helper()
	encrypted := make(map[string]string, len(webhooks))
	for org, url := range webhooks {
		encrypted[org] = encryptWebhook(t, url)
	}
	store := &fakeWebhookStore{webhooks: encrypted, shopWebhooks: map[int32]string{}}

	d := newTestDispatcher(&recordingChannel{name: ChannelInApp}, &recordingChannel{name: ChannelEmail}, &fakeLocker{})
	d.destinations = &destinationChannel{
		store:     store,
		poster:    poster,
		appSecret: testAppSecret,
		baseURL:   "https://shopmon.example.test",
		tr:        d.tr,
	}
	return d, store
}

func orgDegradeEvent() Event {
	ev := degradeEvent()
	ev.OrgID = "org-1"
	ev.ScopeType = ScopeEnvironment
	ev.ScopeID = "7"
	ev.Link = Link{Name: "account.environments.detail", Params: map[string]string{"environmentId": "7"}}
	return ev
}

func TestDestinationPostsOncePerEventRegardlessOfRecipients(t *testing.T) {
	poster := &recordingPoster{}
	d, _ := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": "slack://default?webhook_url=https%3A%2F%2Fhooks.slack.com%2Fservices%2Fa"})

	res := d.Dispatch(context.Background(), orgDegradeEvent(), []Recipient{
		{ID: "u1", Locale: "en"},
		{ID: "u2", Locale: "de"},
	})

	require.Len(t, poster.posts, 1)
	assert.Equal(t, "slack://default?webhook_url=https%3A%2F%2Fhooks.slack.com%2Fservices%2Fa", poster.posts[0].url)
	assert.True(t, res.Delivery())
}

func TestDestinationPostsWithoutSubscribers(t *testing.T) {
	poster := &recordingPoster{}
	d, _ := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": "slack://default?webhook_url=https%3A%2F%2Fhooks.slack.com%2Fservices%2Fa"})

	res := d.Dispatch(context.Background(), orgDegradeEvent(), nil)

	assert.Len(t, poster.posts, 1)
	assert.Equal(t, 1, res.Delivered)
}

func TestDestinationSkippedWhenOrganizationHasNoWebhook(t *testing.T) {
	poster := &recordingPoster{}
	d, _ := newDestinationTestDispatcher(t, poster, nil)

	res := d.Dispatch(context.Background(), orgDegradeEvent(), nil)

	assert.Empty(t, poster.posts)
	assert.Zero(t, res.Failed)
	assert.Zero(t, res.Skipped)
}

func TestDestinationSkippedForEventsWithoutOrganization(t *testing.T) {
	poster := &recordingPoster{}
	d, store := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": "slack://default?webhook_url=https%3A%2F%2Fhooks.slack.com%2Fservices%2Fa"})

	d.Dispatch(context.Background(), degradeEvent(), nil)

	assert.Empty(t, poster.posts)
	assert.Zero(t, store.lookups)
}

func TestDestinationIsDeduped(t *testing.T) {
	poster := &recordingPoster{}
	d, _ := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": "slack://default?webhook_url=https%3A%2F%2Fhooks.slack.com%2Fservices%2Fa"})

	d.Dispatch(context.Background(), orgDegradeEvent(), nil)
	res := d.Dispatch(context.Background(), orgDegradeEvent(), nil)

	assert.Len(t, poster.posts, 1)
	assert.Equal(t, 1, res.Skipped)
}

func TestDestinationFailureIsReported(t *testing.T) {
	poster := &recordingPoster{err: assert.AnError}
	d, _ := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": "slack://default?webhook_url=https%3A%2F%2Fhooks.slack.com%2Fservices%2Fa"})

	res := d.Dispatch(context.Background(), orgDegradeEvent(), []Recipient{{ID: "u1", Locale: "en"}})

	assert.Equal(t, 1, res.Failed)
	assert.False(t, res.Delivery(), "a failed Destination post must not be recorded as sent")
}

// A failed post must not leave the dedup lock behind: the caller retries, and a
// held lock would turn that retry into a "skipped" that is recorded as sent.
func TestDestinationFailureReleasesDedupLockForRetry(t *testing.T) {
	poster := &recordingPoster{err: assert.AnError}
	d, _ := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": "slack://default?webhook_url=https%3A%2F%2Fhooks.slack.com%2Fservices%2Fa"})

	res := d.Dispatch(context.Background(), orgDegradeEvent(), nil)
	require.Equal(t, 1, res.Failed)

	poster.err = nil
	res = d.Dispatch(context.Background(), orgDegradeEvent(), nil)

	assert.Equal(t, 1, res.Delivered)
	assert.Len(t, poster.posts, 1)
}

const (
	orgWebhook  = "slack://default?webhook_url=https%3A%2F%2Fhooks.slack.com%2Fservices%2Forg"
	shopWebhook = "slack://default?webhook_url=https%3A%2F%2Fhooks.slack.com%2Fservices%2Fshop"
)

func TestDestinationPostsToOrganizationAndShopChannels(t *testing.T) {
	poster := &recordingPoster{}
	d, store := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": orgWebhook})
	store.shopWebhooks[7] = encryptWebhook(t, shopWebhook)

	res := d.Dispatch(context.Background(), orgDegradeEvent(), nil)

	require.Len(t, poster.posts, 2)
	assert.Equal(t, orgWebhook, poster.posts[0].url)
	assert.Equal(t, shopWebhook, poster.posts[1].url)
	assert.Equal(t, 2, res.Delivered)
}

func TestDestinationPostsToShopChannelAlone(t *testing.T) {
	poster := &recordingPoster{}
	d, store := newDestinationTestDispatcher(t, poster, nil)
	store.shopWebhooks[7] = encryptWebhook(t, shopWebhook)

	d.Dispatch(context.Background(), orgDegradeEvent(), nil)

	require.Len(t, poster.posts, 1)
	assert.Equal(t, shopWebhook, poster.posts[0].url)
}

func TestDestinationSameWebhookOnBothLevelsPostsOnce(t *testing.T) {
	poster := &recordingPoster{}
	d, store := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": orgWebhook})
	store.shopWebhooks[7] = encryptWebhook(t, orgWebhook)

	d.Dispatch(context.Background(), orgDegradeEvent(), nil)

	assert.Len(t, poster.posts, 1)
}

func TestDestinationShopFailureDoesNotHoldBackOrganization(t *testing.T) {
	poster := &recordingPoster{failURL: shopWebhook}
	d, store := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": orgWebhook})
	store.shopWebhooks[7] = encryptWebhook(t, shopWebhook)

	res := d.Dispatch(context.Background(), orgDegradeEvent(), nil)
	assert.Equal(t, 1, res.Delivered)
	assert.Equal(t, 1, res.Failed)

	// The retry re-posts only to the shop: the organization channel is deduped
	// on its own lock, the shop lock was released on failure.
	poster.failURL = ""
	res = d.Dispatch(context.Background(), orgDegradeEvent(), nil)
	assert.Equal(t, 1, res.Delivered)
	assert.Equal(t, 1, res.Skipped)
	require.Len(t, poster.posts, 2)
	assert.Equal(t, shopWebhook, poster.posts[1].url)
}

// An unknown lock state must count as a failure: a skip would let the caller
// record the alert as delivered although Destination was never called.
func TestDestinationLockErrorIsAFailure(t *testing.T) {
	poster := &recordingPoster{}
	d, _ := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": orgWebhook})
	d.locker = &fakeLocker{err: assert.AnError}

	res := d.Dispatch(context.Background(), orgDegradeEvent(), nil)

	assert.Empty(t, poster.posts)
	assert.Equal(t, 1, res.Failed)
	assert.Zero(t, res.Skipped)
}

// The same holds for email: an unknown lock state must not let the caller
// record the alert as delivered just because the in-app send succeeded.
func TestEmailLockErrorIsAFailure(t *testing.T) {
	inApp := &recordingChannel{name: ChannelInApp}
	email := &recordingChannel{name: ChannelEmail}
	d := newTestDispatcher(inApp, email, &fakeLocker{err: assert.AnError})

	res := d.Dispatch(context.Background(), degradeEvent(), []Recipient{{ID: "u1", Locale: "en"}})

	assert.Len(t, inApp.sent, 1)
	assert.Empty(t, email.sent)
	assert.Equal(t, 1, res.Failed)
	assert.Zero(t, res.Skipped)
	assert.False(t, res.Delivery())
}

// Replacing a webhook inside the dedup window must not dedup the new channel
// against the old one.
func TestDestinationReplacedWebhookIsNotDeduped(t *testing.T) {
	poster := &recordingPoster{}
	d, store := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": orgWebhook})

	d.Dispatch(context.Background(), orgDegradeEvent(), nil)
	store.webhooks["org-1"] = encryptWebhook(t, "slack://default?webhook_url=https%3A%2F%2Fhooks.slack.com%2Fservices%2Freplaced")
	d.Dispatch(context.Background(), orgDegradeEvent(), nil)

	require.Len(t, poster.posts, 2)
	assert.Equal(t, "slack://default?webhook_url=https%3A%2F%2Fhooks.slack.com%2Fservices%2Freplaced", poster.posts[1].url)
}

func TestDestinationRenderEscapesAndLinks(t *testing.T) {
	c := &destinationChannel{tr: NewTranslator(), baseURL: "https://shopmon.example.test/"}
	ev := orgDegradeEvent()
	ev.Params = map[string]any{"name": "Shop <Prod> & Co", "from": "green", "to": "red"}
	ev.Reasons = []StatusReason{{Level: "red", Key: "does.not.exist"}}

	text := c.render(ev)

	assert.Contains(t, text, "⚠️ ")
	assert.Contains(t, text, "Shop <Prod> & Co")
	assert.Contains(t, text, "Status changed from green to red")
	assert.Contains(t, text, "\n• does.not.exist")
	assert.Contains(t, text, "https://shopmon.example.test/app/environments/7")
}

// A corrupt destination must not suppress working destinations at the same scope.
func TestDestinationDecryptionFailureDoesNotBlockOtherTargets(t *testing.T) {
	poster := &recordingPoster{}
	d, store := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": orgWebhook})
	store.shopWebhooks[7] = "corrupt encrypted DSN"
	res := d.Dispatch(t.Context(), orgDegradeEvent(), nil)
	assert.Equal(t, 1, res.Delivered)
	assert.Equal(t, 1, res.Failed)
	assert.False(t, res.Delivery())
}

func TestInvalidEnvironmentIDDoesNotSelectAnotherShop(t *testing.T) {
	poster := &recordingPoster{}
	d, store := newDestinationTestDispatcher(t, poster, nil)
	store.shopWebhooks[2147483647] = encryptWebhook(t, shopWebhook)
	event := orgDegradeEvent()
	event.ScopeID = "999999999999999999"
	res := d.Dispatch(t.Context(), event, nil)
	assert.Zero(t, res.Delivered)
}

func (s *fakeWebhookStore) ListDestinationNotificationPreferences(_ context.Context, id string) ([]queries.DestinationNotificationPreference, error) {
	return s.preferences[id], s.preferenceError
}

func TestDestinationPreferencesResolveDefaultsAndEnvironmentOverrides(t *testing.T) {
	poster := &recordingPoster{}
	d, store := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": orgWebhook})
	store.preferences = map[string][]queries.DestinationNotificationPreference{
		"org": {{ScopeType: "global", Channel: "destination", EventType: string(EventStatusDegraded), Enabled: false}},
	}
	event := orgDegradeEvent()
	res := d.Dispatch(t.Context(), event, nil)
	assert.Zero(t, res.Delivered)
	assert.Zero(t, res.Failed)
	assert.Empty(t, poster.posts)
	// No dedup lock is acquired for an opted-out event. An explicit environment
	// opt-in must therefore deliver immediately, even inside the same window.
	store.preferences["org"] = append(store.preferences["org"], queries.DestinationNotificationPreference{ScopeType: "environment", ScopeID: "7", Channel: "destination", EventType: string(EventStatusDegraded), Enabled: true})
	res = d.Dispatch(t.Context(), event, nil)
	assert.Equal(t, 1, res.Delivered)
	// Another environment continues to inherit the global opt-out.
	event.ScopeID = "8"
	event.DedupKey += "8"
	res = d.Dispatch(t.Context(), event, nil)
	assert.Zero(t, res.Delivered)
	assert.Len(t, poster.posts, 1)
}

func TestDestinationOptOutDoesNotSuppressAnotherEnabledCopy(t *testing.T) {
	poster := &recordingPoster{}
	d, store := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": orgWebhook})
	store.shopWebhooks[7] = encryptWebhook(t, orgWebhook)
	store.preferences = map[string][]queries.DestinationNotificationPreference{"org": {{ScopeType: "global", Channel: "destination", Enabled: false}}}
	res := d.Dispatch(t.Context(), orgDegradeEvent(), nil)
	assert.Equal(t, 1, res.Delivered)
	assert.Len(t, poster.posts, 1)
}

func TestDestinationPreferenceReadFailureIsRetryable(t *testing.T) {
	poster := &recordingPoster{}
	d, store := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": orgWebhook})
	store.preferenceError = assert.AnError
	res := d.Dispatch(t.Context(), orgDegradeEvent(), nil)
	assert.Equal(t, 1, res.Failed)
	assert.Empty(t, poster.posts)
	store.preferenceError = nil
	res = d.Dispatch(t.Context(), orgDegradeEvent(), nil)
	assert.Equal(t, 1, res.Delivered)
}

func TestUnknownEventsNeverReachSharedDestinations(t *testing.T) {
	poster := &recordingPoster{}
	d, store := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": orgWebhook})
	event := orgDegradeEvent()
	event.Type = "unknown.event"
	res := d.Dispatch(t.Context(), event, nil)
	assert.Zero(t, res.Delivered)
	assert.Zero(t, store.lookups)
}

func TestDestinationEquivalentWebhookEncodingsDeduplicate(t *testing.T) {
	poster := &recordingPoster{}
	d, store := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": orgWebhook})
	store.shopWebhooks[7] = encryptWebhook(t, "slack://default?webhook_url=https://hooks.slack.com/services/org")
	res := d.Dispatch(t.Context(), orgDegradeEvent(), nil)
	assert.Equal(t, 1, res.Delivered)
	require.Len(t, poster.posts, 1)
	// A later edit of encoding alone must retain the completed marker.
	store.webhooks["org-1"] = store.shopWebhooks[7]
	res = d.Dispatch(t.Context(), orgDegradeEvent(), nil)
	assert.Equal(t, 1, res.Skipped)
	assert.Len(t, poster.posts, 1)
}

func TestDestinationPendingAttemptMustRetry(t *testing.T) {
	poster := &recordingPoster{}
	d, _ := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": orgWebhook})
	event := orgDegradeEvent()
	key := (destinationTarget{dsn: orgWebhook}).lockKey(event.DedupKey)
	// Models both an active sender and a failed sender whose cleanup failed.
	locker := &fakeLocker{held: map[string]bool{key: true}}
	d.locker = locker
	res := d.Dispatch(t.Context(), event, nil)
	assert.Equal(t, 1, res.Failed)
	assert.Zero(t, res.Skipped)
	assert.Empty(t, poster.posts)
	locker.release(t.Context(), key)
	res = d.Dispatch(t.Context(), event, nil)
	assert.Equal(t, 1, res.Delivered)
	res = d.Dispatch(t.Context(), event, nil)
	assert.Equal(t, 1, res.Skipped)
}

func TestDestinationSeparateBatchesKeepInAppKey(t *testing.T) {
	poster := &recordingPoster{}
	d, _ := newDestinationTestDispatcher(t, poster, map[string]string{"org-1": orgWebhook})
	event := orgDegradeEvent()
	event.Type = EventAdvisoryDetected
	event.DestinationDedupKey = "batch-1"
	assert.Equal(t, 1, d.Dispatch(t.Context(), event, nil).Delivered)
	assert.Equal(t, 1, d.Dispatch(t.Context(), event, nil).Skipped)
	event.DestinationDedupKey = "batch-2"
	assert.Equal(t, 1, d.Dispatch(t.Context(), event, nil).Delivered)
	assert.Len(t, poster.posts, 2)
}
