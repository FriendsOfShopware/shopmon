package notify

import (
	"context"
	"sync"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/crypto"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAppSecret = "0123456789abcdef0123456789abcdef"

// fakeWebhookStore returns encrypted webhooks per organization and per
// environment's shop.
type fakeWebhookStore struct {
	webhooks     map[string]string
	shopWebhooks map[int32]string
	lookups      int
}

func (s *fakeWebhookStore) GetOrganizationSlackWebhook(_ context.Context, orgID string) (string, error) {
	s.lookups++
	encrypted, ok := s.webhooks[orgID]
	if !ok {
		return "", pgx.ErrNoRows
	}
	return encrypted, nil
}

func (s *fakeWebhookStore) GetEnvironmentShopSlackWebhook(_ context.Context, environmentID int32) (string, error) {
	s.lookups++
	encrypted, ok := s.shopWebhooks[environmentID]
	if !ok {
		return "", pgx.ErrNoRows
	}
	return encrypted, nil
}

type slackPost struct {
	url  string
	text string
}

type recordingPoster struct {
	mu    sync.Mutex
	posts []slackPost
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
	p.posts = append(p.posts, slackPost{url: webhookURL, text: text})
	return nil
}

func encryptWebhook(t *testing.T, url string) string {
	t.Helper()
	e, err := crypto.Encrypt(url, testAppSecret)
	require.NoError(t, err)
	return e
}

func newSlackTestDispatcher(t *testing.T, poster *recordingPoster, webhooks map[string]string) (*Dispatcher, *fakeWebhookStore) {
	t.Helper()
	encrypted := make(map[string]string, len(webhooks))
	for org, url := range webhooks {
		encrypted[org] = encryptWebhook(t, url)
	}
	store := &fakeWebhookStore{webhooks: encrypted, shopWebhooks: map[int32]string{}}

	d := newTestDispatcher(&recordingChannel{name: ChannelInApp}, &recordingChannel{name: ChannelEmail}, &fakeLocker{})
	d.slack = &slackChannel{
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

func TestSlackPostsOncePerEventRegardlessOfRecipients(t *testing.T) {
	poster := &recordingPoster{}
	d, _ := newSlackTestDispatcher(t, poster, map[string]string{"org-1": "https://hooks.slack.com/services/a"})

	res := d.Dispatch(context.Background(), orgDegradeEvent(), []Recipient{
		{ID: "u1", Locale: "en"},
		{ID: "u2", Locale: "de"},
	})

	require.Len(t, poster.posts, 1)
	assert.Equal(t, "https://hooks.slack.com/services/a", poster.posts[0].url)
	assert.True(t, res.Delivery())
}

func TestSlackPostsWithoutSubscribers(t *testing.T) {
	poster := &recordingPoster{}
	d, _ := newSlackTestDispatcher(t, poster, map[string]string{"org-1": "https://hooks.slack.com/services/a"})

	res := d.Dispatch(context.Background(), orgDegradeEvent(), nil)

	assert.Len(t, poster.posts, 1)
	assert.Equal(t, 1, res.Delivered)
}

func TestSlackSkippedWhenOrganizationHasNoWebhook(t *testing.T) {
	poster := &recordingPoster{}
	d, _ := newSlackTestDispatcher(t, poster, nil)

	res := d.Dispatch(context.Background(), orgDegradeEvent(), nil)

	assert.Empty(t, poster.posts)
	assert.Zero(t, res.Failed)
	assert.Zero(t, res.Skipped)
}

func TestSlackSkippedForEventsWithoutOrganization(t *testing.T) {
	poster := &recordingPoster{}
	d, store := newSlackTestDispatcher(t, poster, map[string]string{"org-1": "https://hooks.slack.com/services/a"})

	d.Dispatch(context.Background(), degradeEvent(), nil)

	assert.Empty(t, poster.posts)
	assert.Zero(t, store.lookups)
}

func TestSlackIsDeduped(t *testing.T) {
	poster := &recordingPoster{}
	d, _ := newSlackTestDispatcher(t, poster, map[string]string{"org-1": "https://hooks.slack.com/services/a"})

	d.Dispatch(context.Background(), orgDegradeEvent(), nil)
	res := d.Dispatch(context.Background(), orgDegradeEvent(), nil)

	assert.Len(t, poster.posts, 1)
	assert.Equal(t, 1, res.Skipped)
}

func TestSlackFailureIsReported(t *testing.T) {
	poster := &recordingPoster{err: assert.AnError}
	d, _ := newSlackTestDispatcher(t, poster, map[string]string{"org-1": "https://hooks.slack.com/services/a"})

	res := d.Dispatch(context.Background(), orgDegradeEvent(), []Recipient{{ID: "u1", Locale: "en"}})

	assert.Equal(t, 1, res.Failed)
	assert.False(t, res.Delivery(), "a failed Slack post must not be recorded as sent")
}

// A failed post must not leave the dedup lock behind: the caller retries, and a
// held lock would turn that retry into a "skipped" that is recorded as sent.
func TestSlackFailureReleasesDedupLockForRetry(t *testing.T) {
	poster := &recordingPoster{err: assert.AnError}
	d, _ := newSlackTestDispatcher(t, poster, map[string]string{"org-1": "https://hooks.slack.com/services/a"})

	res := d.Dispatch(context.Background(), orgDegradeEvent(), nil)
	require.Equal(t, 1, res.Failed)

	poster.err = nil
	res = d.Dispatch(context.Background(), orgDegradeEvent(), nil)

	assert.Equal(t, 1, res.Delivered)
	assert.Len(t, poster.posts, 1)
}

const (
	orgWebhook  = "https://hooks.slack.com/services/org"
	shopWebhook = "https://hooks.slack.com/services/shop"
)

func TestSlackPostsToOrganizationAndShopChannels(t *testing.T) {
	poster := &recordingPoster{}
	d, store := newSlackTestDispatcher(t, poster, map[string]string{"org-1": orgWebhook})
	store.shopWebhooks[7] = encryptWebhook(t, shopWebhook)

	res := d.Dispatch(context.Background(), orgDegradeEvent(), nil)

	require.Len(t, poster.posts, 2)
	assert.Equal(t, orgWebhook, poster.posts[0].url)
	assert.Equal(t, shopWebhook, poster.posts[1].url)
	assert.Equal(t, 2, res.Delivered)
}

func TestSlackPostsToShopChannelAlone(t *testing.T) {
	poster := &recordingPoster{}
	d, store := newSlackTestDispatcher(t, poster, nil)
	store.shopWebhooks[7] = encryptWebhook(t, shopWebhook)

	d.Dispatch(context.Background(), orgDegradeEvent(), nil)

	require.Len(t, poster.posts, 1)
	assert.Equal(t, shopWebhook, poster.posts[0].url)
}

func TestSlackSameWebhookOnBothLevelsPostsOnce(t *testing.T) {
	poster := &recordingPoster{}
	d, store := newSlackTestDispatcher(t, poster, map[string]string{"org-1": orgWebhook})
	store.shopWebhooks[7] = encryptWebhook(t, orgWebhook)

	d.Dispatch(context.Background(), orgDegradeEvent(), nil)

	assert.Len(t, poster.posts, 1)
}

func TestSlackShopFailureDoesNotHoldBackOrganization(t *testing.T) {
	poster := &recordingPoster{failURL: shopWebhook}
	d, store := newSlackTestDispatcher(t, poster, map[string]string{"org-1": orgWebhook})
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

func TestSlackRenderEscapesAndLinks(t *testing.T) {
	c := &slackChannel{tr: NewTranslator(), baseURL: "https://shopmon.example.test/"}
	ev := orgDegradeEvent()
	ev.Params = map[string]any{"name": "Shop <Prod> & Co", "from": "green", "to": "red"}
	ev.Reasons = []StatusReason{{Level: "red", Key: "does.not.exist"}}

	text := c.render(ev)

	assert.Contains(t, text, ":warning: *")
	assert.Contains(t, text, "Shop &lt;Prod&gt; &amp; Co")
	assert.Contains(t, text, "Status changed from green to red")
	assert.Contains(t, text, "\n• does.not.exist")
	assert.Contains(t, text, "<https://shopmon.example.test/app/environments/7|")
}
