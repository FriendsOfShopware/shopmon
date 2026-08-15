package notify

import (
	"context"
	"fmt"
	"strings"

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

	// Subject: event override, then the registry key, then the notification title.
	subject := msg.Title
	subjectKey := ev.EmailSubjectKey
	if subjectKey == "" {
		subjectKey = meta.emailSubjectKey
	}
	if subjectKey != "" {
		subject = c.tr.T(r.Locale, subjectKey, ev.Params)
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

	email := c.mail.BuildAlertEmail(r.Name, subject, intro, message, c.alertAction(r.Locale, ev))

	if err := c.mail.Send(ctx, r.Email, email); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

func (c *emailChannel) alertAction(locale string, ev Event) *mail.AlertAction {
	url := frontendURL(c.mail.ProductURL(), ev.Link)
	if url == "" {
		return nil
	}
	return &mail.AlertAction{
		Text: c.tr.T(locale, actionTextKey(ev.Link.Name), nil),
		URL:  url,
	}
}

func actionTextKey(route string) string {
	switch route {
	case "account.environments.detail":
		return "email.viewEnvironment"
	case "account.advisories.detail":
		return "email.viewAdvisory"
	default:
		return "email.viewDetails"
	}
}

// frontendURL turns a stored Vue route reference into an absolute app URL.
func frontendURL(base string, link Link) string {
	base = strings.TrimRight(base, "/")
	if base == "" || link.Name == "" {
		return ""
	}
	switch link.Name {
	case "account.environments.detail":
		if id := link.Params["environmentId"]; id != "" {
			return base + "/app/environments/" + id
		}
	case "account.advisories.detail":
		if id := link.Params["id"]; id != "" {
			return base + "/app/advisories/" + id
		}
	}
	return ""
}
