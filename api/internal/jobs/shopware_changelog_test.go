package jobs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShopwareVersionRe(t *testing.T) {
	valid := []string{
		"6.7.11.1",
		"6.6.10.0",
		"6.5.0.0-rc1",
		"6.7.0.0-RC5",
		"6.4.20",
	}
	for _, v := range valid {
		assert.Truef(t, shopwareVersionRe.MatchString(v), "expected %q to be valid", v)
	}

	// Anything that could redirect the request URL must be rejected.
	invalid := []string{
		"../secret",
		"6.7.11.1/../../etc",
		"6.7.11.1?foo=bar",
		"6.7.11.1#frag",
		"latest",
		"",
	}
	for _, v := range invalid {
		assert.Falsef(t, shopwareVersionRe.MatchString(v), "expected %q to be rejected", v)
	}
}
