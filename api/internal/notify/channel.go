package notify

import "context"

// ChannelName identifies a delivery channel.
type ChannelName string

const (
	ChannelInApp ChannelName = "in_app"
	ChannelEmail ChannelName = "email"
	// ChannelDestination is scope-level: it is not a per-recipient Channel
	// and is dispatched once per event (see Dispatcher.dispatchDestinations).
	ChannelDestination ChannelName = "destination"
)

// Recipient is a user targeted by an event, resolved to the fields delivery and
// rendering need.
type Recipient struct {
	ID     string
	Name   string
	Email  string
	Locale string
}

// RenderedMessage is the event text rendered into a single recipient's locale.
type RenderedMessage struct {
	Title string
	Body  string
}

// Channel delivers a rendered event to one recipient. Implementations must be
// safe for concurrent use.
type Channel interface {
	Name() ChannelName
	Send(ctx context.Context, r Recipient, ev Event, msg RenderedMessage) error
}
