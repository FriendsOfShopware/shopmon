package notify

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestTranslatorInterpolatesAndFallsBack(t *testing.T) {
	tr := NewTranslator()

	got := tr.T("en", "notification.statusDegraded.message", map[string]any{"from": "green", "to": "red"})
	if want := "Status changed from green to red"; got != want {
		t.Fatalf("en interpolation = %q, want %q", got, want)
	}

	got = tr.T("de", "notification.statusDegraded.message", map[string]any{"from": "green", "to": "red"})
	if want := "Status von green auf red geändert"; got != want {
		t.Fatalf("de interpolation = %q, want %q", got, want)
	}

	// Unknown locale falls back to the default locale catalog.
	got = tr.T("fr", "notification.authError.title", map[string]any{"name": "Shop"})
	if want := "Environment: Shop could not be updated"; got != want {
		t.Fatalf("fallback locale = %q, want %q", got, want)
	}

	// Unknown key returns the raw key rather than failing.
	if got := tr.T("en", "does.not.exist", nil); got != "does.not.exist" {
		t.Fatalf("missing key = %q, want raw key", got)
	}
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

	if len(inApp.sent) != 2 {
		t.Fatalf("in-app sends = %d, want 2", len(inApp.sent))
	}
	if inApp.sent[0].Body != "Status changed from green to yellow" {
		t.Fatalf("en body = %q", inApp.sent[0].Body)
	}
	if inApp.sent[1].Body != "Status von green auf yellow geändert" {
		t.Fatalf("de body = %q", inApp.sent[1].Body)
	}
}

func TestDispatchEmailDedupGatesEmailOnly(t *testing.T) {
	inApp := &recordingChannel{name: ChannelInApp}
	email := &recordingChannel{name: ChannelEmail}
	locker := &fakeLocker{}
	d := newTestDispatcher(inApp, email, locker)

	recipients := []Recipient{{ID: "u1", Locale: "en"}}

	// First dispatch acquires the lock: both channels fire.
	d.Dispatch(context.Background(), degradeEvent(), recipients)
	// Second dispatch is deduped: in-app re-records, email is suppressed.
	d.Dispatch(context.Background(), degradeEvent(), recipients)

	if len(inApp.sent) != 2 {
		t.Fatalf("in-app sends = %d, want 2 (idempotent re-record)", len(inApp.sent))
	}
	if len(email.sent) != 1 {
		t.Fatalf("email sends = %d, want 1 (deduped)", len(email.sent))
	}
}

func TestDispatchDataFetchErrorSkipsEmail(t *testing.T) {
	inApp := &recordingChannel{name: ChannelInApp}
	email := &recordingChannel{name: ChannelEmail}
	d := newTestDispatcher(inApp, email, &fakeLocker{})

	d.Dispatch(context.Background(), Event{
		Type:       EventDataFetchError,
		Level:      LevelError,
		DedupKey:   "environment.not.updated_1",
		TitleKey:   "notification.dataFetchError.title",
		MessageKey: "notification.dataFetchError.message",
		Params:     map[string]any{"name": "Shop"},
	}, []Recipient{{ID: "u1", Locale: "en"}})

	if len(inApp.sent) != 1 {
		t.Fatalf("in-app sends = %d, want 1", len(inApp.sent))
	}
	if len(email.sent) != 0 {
		t.Fatalf("email sends = %d, want 0 (data-fetch is in-app only)", len(email.sent))
	}
}
