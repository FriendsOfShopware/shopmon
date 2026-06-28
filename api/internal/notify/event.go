// Package notify is the notification domain: it turns typed Events into
// delivered messages across pluggable Channels, honouring per-recipient
// preferences and rendering text from a shared i18n catalog.
//
// The design separates four concerns that were previously tangled together in
// the scrape job:
//
//   - Event       — what happened (typed, with structured params)
//   - Preferences — who is told and on which channels (PreferenceResolver)
//   - Channel     — how it is delivered (in-app, email, future: webhook/Slack)
//   - Translator  — what it says, rendered per recipient locale
//
// Adding a new alert is "register an EventType + add translation keys". Adding a
// new delivery method is "implement one Channel". Job code does not change.
package notify

// Level is the severity of a notification, matching the persisted user
// notification levels and the frontend alert styling.
type Level string

const (
	LevelInfo    Level = "info"
	LevelWarning Level = "warning"
	LevelError   Level = "error"
)

// ScopeType identifies what a notification is about, so preferences can be
// resolved with an override chain (global -> organization -> environment).
type ScopeType string

const (
	ScopeGlobal       ScopeType = "global"
	ScopeOrganization ScopeType = "organization"
	ScopeEnvironment  ScopeType = "environment"
)

// EventType is the stable identifier for a kind of notification. It keys the
// registry (default level, channels, translation keys) and is the extension
// point for future alerts.
type EventType string

const (
	EventStatusDegraded  EventType = "status_degraded"
	EventStatusRecovered EventType = "status_recovered"
	EventAuthError       EventType = "auth_error"
	EventDataFetchError  EventType = "data_fetch_error"
)

// Link is a frontend route reference attached to a notification.
type Link struct {
	Name   string            `json:"name"`
	Params map[string]string `json:"params"`
}

// StatusReason is a single check that drove a status transition (e.g. a newly
// failing or newly recovered check). It is rendered the same way as a check:
// translate Key with Params. It travels into notification params (for the UI) and
// into the status-event timeline.
type StatusReason struct {
	Level  string         `json:"level"`
	Key    string         `json:"messageKey"`
	Params map[string]any `json:"params,omitempty"`
	Source string         `json:"source"`
}

// Event is a single thing that happened and may warrant notifying users. It
// carries only facts (keys + params); rendering into language happens per
// recipient at delivery time.
type Event struct {
	Type      EventType
	Level     Level
	ScopeType ScopeType
	ScopeID   string

	// OrgID is the organization the scope belongs to, used to resolve
	// organization-level preference overrides. Empty when not applicable.
	OrgID string

	// DedupKey uniquely identifies this notification instance. It is the
	// user_notification upsert key (idempotent re-recording) and the basis for
	// the per-event email dedup lock.
	DedupKey string

	// TitleKey and MessageKey are catalog keys rendered with Params.
	TitleKey   string
	MessageKey string
	Params     map[string]any

	// Reasons are the checks that caused this event (status transitions). They
	// are surfaced in the UI and email body, and stored on the timeline.
	Reasons []StatusReason

	Link Link
}
