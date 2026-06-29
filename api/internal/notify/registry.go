package notify

import "time"

// eventMeta is the per-event-type configuration: default delivery channels, the
// email subject key, and the email dedup window. It is the single source of
// truth for how an event behaves.
type eventMeta struct {
	// defaultChannels are delivered when a recipient has no overriding
	// preference (Phase 1: always used).
	defaultChannels []ChannelName

	// emailSubjectKey is the catalog key for the email subject. Empty means the
	// notification title is reused as the subject.
	emailSubjectKey string

	// emailDedupTTL gates email sends behind a lock for this duration to prevent
	// spam. Zero means no email-level dedup (the in-app upsert is idempotent on
	// its own).
	emailDedupTTL time.Duration
}

// registry maps every EventType to its behaviour. Adding an alert type is a
// matter of adding an entry here plus the matching translation keys.
var registry = map[EventType]eventMeta{
	EventStatusDegraded: {
		defaultChannels: []ChannelName{ChannelInApp, ChannelEmail},
		emailSubjectKey: "email.statusDegraded.subject",
		emailDedupTTL:   time.Hour,
	},
	EventStatusRecovered: {
		defaultChannels: []ChannelName{ChannelInApp, ChannelEmail},
		emailSubjectKey: "email.statusRecovered.subject",
		emailDedupTTL:   time.Hour,
	},
	EventAuthError: {
		defaultChannels: []ChannelName{ChannelInApp, ChannelEmail},
		emailSubjectKey: "email.authError.subject",
		emailDedupTTL:   time.Hour,
	},
	EventDataFetchError: {
		// Data-fetch errors are re-recorded every scrape, so the email is gated
		// behind a one-hour dedup lock to avoid spamming while a fetch keeps
		// failing.
		defaultChannels: []ChannelName{ChannelInApp, ChannelEmail},
		emailSubjectKey: "email.dataFetchError.subject",
		emailDedupTTL:   time.Hour,
	},
}

// metaFor returns the registry entry for an event type, falling back to a safe
// in-app-only default for unregistered types.
func metaFor(t EventType) eventMeta {
	if m, ok := registry[t]; ok {
		return m
	}
	return eventMeta{defaultChannels: []ChannelName{ChannelInApp}}
}

// orderedEventTypes is the stable display order for ListEventTypes (registry is
// a map and therefore unordered).
var orderedEventTypes = []EventType{
	EventStatusDegraded,
	EventStatusRecovered,
	EventAuthError,
	EventDataFetchError,
}

// EventTypeInfo describes a notifiable event type for the preferences UI.
type EventTypeInfo struct {
	Type            string
	DefaultChannels []string
}

// ListEventTypes returns the notifiable event types and their default delivery
// channels, in a stable order. It is the single source of truth for the
// preferences UI, so new event types appear automatically.
func ListEventTypes() []EventTypeInfo {
	out := make([]EventTypeInfo, 0, len(orderedEventTypes))
	for _, t := range orderedEventTypes {
		meta := metaFor(t)
		channels := make([]string, 0, len(meta.defaultChannels))
		for _, c := range meta.defaultChannels {
			channels = append(channels, string(c))
		}
		out = append(out, EventTypeInfo{Type: string(t), DefaultChannels: channels})
	}
	return out
}
