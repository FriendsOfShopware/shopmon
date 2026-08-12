package mail

import (
	"errors"

	"github.com/friendsofshopware/shopmon/api/internal/otelx"
	gomailer "github.com/shyim/go-mailer"
)

// SMTPCodeExpected reports whether an SMTP response code is a soft/transient
// failure (SES 451 timeouts, greylisting, mailbox busy, etc.).
//
// 421/450/451/452 are widely treated as retryable; permanent 5xx rejects and
// other 4xx (e.g. 550) are hard failures.
func SMTPCodeExpected(code int) bool {
	switch code {
	case 421, 450, 451, 452:
		return true
	default:
		return false
	}
}

// IsExpectedSMTPError reports whether err is (or wraps) a retryable SMTP
// transport failure or a retryable network blip talking to the relay.
func IsExpectedSMTPError(err error) bool {
	if err == nil {
		return false
	}
	var te *gomailer.TransportError
	if errors.As(err, &te) && SMTPCodeExpected(te.Code) {
		return true
	}
	return otelx.IsRetryableNetError(err)
}
