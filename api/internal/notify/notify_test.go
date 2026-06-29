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

// recordingChannel captures every Send for assertions.
type recordingChannel struct {
	name ChannelName
	mu   sync.Mutex
	sent []RenderedMessage
}

func (c *recordingChannel) Name() ChannelName { return c.name }

func (c *recordingChannel) Send(_ context.Context, _ Recipient, _ Event, msg RenderedMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()
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
