package notify

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslatorInterpolatesAndFallsBack(t *testing.T) {
	tr := NewTranslator()

	assert.Equal(t,
		"Status changed from green to red",
		tr.T("en", "notification.statusDegraded.message", map[string]any{"from": "green", "to": "red"}))

	assert.Equal(t,
		"Status von green auf red geändert",
		tr.T("de", "notification.statusDegraded.message", map[string]any{"from": "green", "to": "red"}))

	// Unknown locale falls back to the default locale catalog.
	assert.Equal(t,
		"Environment: Shop could not be updated",
		tr.T("fr", "notification.authError.title", map[string]any{"name": "Shop"}))

	// Unknown key returns the raw key rather than failing.
	assert.Equal(t, "does.not.exist", tr.T("en", "does.not.exist", nil))
}

func TestTranslatorAdvisoryDetectedSingularAndPlural(t *testing.T) {
	tr := NewTranslator()

	assert.Equal(t, "1 new advisory affects this environment",
		tr.T("en", "notification.advisoryDetected.messageOne", map[string]any{"count": 1}))
	assert.Equal(t, "3 new advisories affect this environment",
		tr.T("en", "notification.advisoryDetected.message", map[string]any{"count": 3}))
	assert.Equal(t, "1 new security advisory affects Demo · Production",
		tr.T("en", "email.advisoryDetected.subjectOne", map[string]any{"name": "Demo · Production"}))
	assert.Equal(t, "2 new security advisories affect Demo · Production",
		tr.T("en", "email.advisoryDetected.subject", map[string]any{"count": 2, "name": "Demo · Production"}))

	assert.Equal(t, "1 neue Meldung betrifft diese Umgebung",
		tr.T("de", "notification.advisoryDetected.messageOne", nil))
	assert.Equal(t, "High: SSE buffer (CVE-2026-1) — mcp/sdk 1.2.3",
		tr.T("en", "check.security.advisoryAlert", map[string]any{
			"severity": "High", "title": "SSE buffer", "id": "CVE-2026-1",
			"package": "mcp/sdk", "installedVersion": "1.2.3",
		}))
}

func TestFrontendURLResolvesKnownRoutes(t *testing.T) {
	assert.Equal(t, "https://app.example.test/app/environments/42",
		frontendURL("https://app.example.test/", Link{
			Name:   "account.environments.detail",
			Params: map[string]string{"environmentId": "42"},
		}))
	assert.Equal(t, "https://app.example.test/app/advisories/CVE-2026-1",
		frontendURL("https://app.example.test", Link{
			Name:   "account.advisories.detail",
			Params: map[string]string{"id": "CVE-2026-1"},
		}))
	assert.Empty(t, frontendURL("https://app.example.test", Link{Name: "account.dashboard"}))
	assert.Empty(t, frontendURL("", Link{Name: "account.advisories.detail", Params: map[string]string{"id": "x"}}))
}

func TestActionTextKey(t *testing.T) {
	assert.Equal(t, "email.viewEnvironment", actionTextKey("account.environments.detail"))
	assert.Equal(t, "email.viewAdvisory", actionTextKey("account.advisories.detail"))
	assert.Equal(t, "email.viewDetails", actionTextKey("account.dashboard"))
}

// recordingChannel captures every Send for assertions.
type recordingChannel struct {
	name ChannelName
	mu   sync.Mutex
	sent []RenderedMessage
	// err, when set, makes every send fail, standing in for an unreachable
	// SMTP server or a failing in-app insert.
	err error
}

func (c *recordingChannel) Name() ChannelName { return c.name }

func (c *recordingChannel) Send(_ context.Context, _ Recipient, _ Event, msg RenderedMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.err != nil {
		return c.err
	}
	c.sent = append(c.sent, msg)
	return nil
}

// fakeLocker grants the first acquire of a key and denies subsequent ones,
// mimicking the dedup lock.
type fakeLocker struct {
	mu   sync.Mutex
	held map[string]bool
}

func (l *fakeLocker) acquire(_ context.Context, key string, _ time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.held == nil {
		l.held = map[string]bool{}
	}
	if l.held[key] {
		return false
	}
	l.held[key] = true
	return true
}

func newTestDispatcher(inApp, email *recordingChannel, locker lockAcquirer) *Dispatcher {
	return &Dispatcher{
		tr:     NewTranslator(),
		prefs:  defaultResolver{},
		locker: locker,
		channels: map[ChannelName]Channel{
			ChannelInApp: inApp,
			ChannelEmail: email,
		},
	}
}

func degradeEvent() Event {
	return Event{
		Type:       EventStatusDegraded,
		Level:      LevelWarning,
		DedupKey:   "environment.change-status.1",
		TitleKey:   "notification.statusDegraded.title",
		MessageKey: "notification.statusDegraded.message",
		Params:     map[string]any{"name": "Shop", "from": "green", "to": "yellow"},
	}
}

func TestDispatchRendersPerRecipientLocale(t *testing.T) {
	inApp := &recordingChannel{name: ChannelInApp}
	email := &recordingChannel{name: ChannelEmail}
	d := newTestDispatcher(inApp, email, &fakeLocker{})

	d.Dispatch(context.Background(), degradeEvent(), []Recipient{
		{ID: "u1", Locale: "en"},
		{ID: "u2", Locale: "de"},
	})

	require.Len(t, inApp.sent, 2)
	assert.Equal(t, "Status changed from green to yellow", inApp.sent[0].Body)
	assert.Equal(t, "Status von green auf yellow geändert", inApp.sent[1].Body)
}

func TestDispatchEmailDedupGatesEmailOnly(t *testing.T) {
	inApp := &recordingChannel{name: ChannelInApp}
	email := &recordingChannel{name: ChannelEmail}
	d := newTestDispatcher(inApp, email, &fakeLocker{})

	recipients := []Recipient{{ID: "u1", Locale: "en"}}

	// First dispatch acquires the lock: both channels fire.
	d.Dispatch(context.Background(), degradeEvent(), recipients)
	// Second dispatch is deduped: in-app re-records, email is suppressed.
	d.Dispatch(context.Background(), degradeEvent(), recipients)

	assert.Len(t, inApp.sent, 2, "in-app re-records idempotently")
	assert.Len(t, email.sent, 1, "email is deduped")
}

func TestDispatchDataFetchErrorEmails(t *testing.T) {
	inApp := &recordingChannel{name: ChannelInApp}
	email := &recordingChannel{name: ChannelEmail}
	d := newTestDispatcher(inApp, email, &fakeLocker{})

	dataFetch := Event{
		Type:       EventDataFetchError,
		Level:      LevelError,
		DedupKey:   "environment.not.updated_1",
		TitleKey:   "notification.dataFetchError.title",
		MessageKey: "notification.dataFetchError.message",
		Params:     map[string]any{"name": "Shop"},
	}
	recipients := []Recipient{{ID: "u1", Locale: "en"}}

	// Data-fetch errors now deliver on both channels, with the email deduped.
	d.Dispatch(context.Background(), dataFetch, recipients)
	d.Dispatch(context.Background(), dataFetch, recipients)

	assert.Len(t, inApp.sent, 2, "in-app re-records idempotently")
	assert.Len(t, email.sent, 1, "email fires once then is deduped")
}

// DispatchResult is what lets a caller tell a delivered alert from a lost one.
// Callers that persist an "already told them" marker consult it before writing,
// so a channel failure is retried instead of being recorded as sent.
func TestDispatchResultReportsDeliveryOutcome(t *testing.T) {
	recipients := []Recipient{{ID: "u1", Locale: "en"}}

	t.Run("all channels deliver", func(t *testing.T) {
		inApp := &recordingChannel{name: ChannelInApp}
		email := &recordingChannel{name: ChannelEmail}
		d := newTestDispatcher(inApp, email, &fakeLocker{})

		res := d.Dispatch(context.Background(), degradeEvent(), recipients)

		assert.Zero(t, res.Failed)
		assert.Positive(t, res.Delivered)
		assert.True(t, res.Delivery(), "a clean dispatch is safe to record as sent")
	})

	t.Run("a failing channel is reported", func(t *testing.T) {
		inApp := &recordingChannel{name: ChannelInApp, err: assert.AnError}
		email := &recordingChannel{name: ChannelEmail}
		d := newTestDispatcher(inApp, email, &fakeLocker{})

		res := d.Dispatch(context.Background(), degradeEvent(), recipients)

		assert.Equal(t, 1, res.Failed)
		assert.False(t, res.Delivery(), "a partial failure must not be recorded as sent")
	})

	t.Run("every channel failing is reported", func(t *testing.T) {
		inApp := &recordingChannel{name: ChannelInApp, err: assert.AnError}
		email := &recordingChannel{name: ChannelEmail, err: assert.AnError}
		d := newTestDispatcher(inApp, email, &fakeLocker{})

		res := d.Dispatch(context.Background(), degradeEvent(), recipients)

		assert.Zero(t, res.Delivered)
		assert.Positive(t, res.Failed)
		assert.False(t, res.Delivery())
	})

	t.Run("no recipients is not a failure", func(t *testing.T) {
		inApp := &recordingChannel{name: ChannelInApp}
		email := &recordingChannel{name: ChannelEmail}
		d := newTestDispatcher(inApp, email, &fakeLocker{})

		res := d.Dispatch(context.Background(), degradeEvent(), nil)

		assert.Zero(t, res.Failed)
		assert.Zero(t, res.Delivered)
		assert.False(t, res.Delivery(), "nothing was delivered, but nothing needs retrying either")
	})
}
