package mail

import (
	"errors"
	"net"
	"syscall"
	"testing"

	gomailer "github.com/shyim/go-mailer"
	"github.com/stretchr/testify/assert"
)

func TestSMTPCodeExpected(t *testing.T) {
	for _, code := range []int{421, 450, 451, 452} {
		assert.True(t, SMTPCodeExpected(code), "code %d", code)
	}
	for _, code := range []int{0, 250, 550, 554, 400} {
		assert.False(t, SMTPCodeExpected(code), "code %d", code)
	}
}

func TestIsExpectedSMTPError(t *testing.T) {
	assert.False(t, IsExpectedSMTPError(nil))

	soft := gomailer.NewTransportError("timeout")
	soft.Code = 451
	assert.True(t, IsExpectedSMTPError(soft))
	assert.True(t, IsExpectedSMTPError(errors.Join(errors.New("wrap"), soft)))

	hard := gomailer.NewTransportError("mailbox missing")
	hard.Code = 550
	assert.False(t, IsExpectedSMTPError(hard))

	assert.True(t, IsExpectedSMTPError(&net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: syscall.ECONNREFUSED,
	}))
}
