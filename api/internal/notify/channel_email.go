package notify

import (
	"context"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/mail"
)

// emailChannel renders and sends notification emails in the recipient's locale.
type emailChannel struct {
	mail mail.Sender
	tr   *Translator
}

func (c *emailChannel) Name() ChannelName { return ChannelEmail }

func (c *emailChannel) Send(ctx context.Context, r Recipient, ev Event, msg RenderedMessage) error {
	if c.mail == nil {
		return nil
	}

	meta := metaFor(ev.Type)

	// Subject: dedicated email subject key when present, otherwise the
	// notification title.
	subject := msg.Title
	if meta.emailSubjectKey != "" {
		subject = c.tr.T(r.Locale, meta.emailSubjectKey, ev.Params)
	}

	intro := c.tr.T(r.Locale, "email.alertIntro", ev.Params)

	message := msg.Body
	if len(ev.Reasons) > 0 {
		message += "\n\n"
		for i, reason := range ev.Reasons {
			if i > 0 {
				message += "\n"
			}
			message += "- " + c.tr.RenderCheck(r.Locale, reason.Key, reason.Params)
		}
	}

	email := c.mail.BuildAlertEmail(r.Name, subject, intro, message)

	if err := c.mail.Send(ctx, r.Email, email); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}
