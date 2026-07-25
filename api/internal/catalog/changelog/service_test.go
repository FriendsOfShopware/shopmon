package changelog

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

func TestShouldFetchChangelog(t *testing.T) {
	// At or above the 6.5.0.0 floor, including pre-releases of the floor.
	want := []string{
		"6.5.0.0",
		"6.5.0.0-rc1",
		"6.5.0.0-RC5",
		"6.5.1.0",
		"6.6.10.0",
		"6.7.11.1",
	}
	for _, v := range want {
		assert.Truef(t, shouldFetchChangelog(v), "expected %q to be fetched", v)
	}

	// Below the floor: the legacy 6.4 / 6.3 / 6.2 / 6.1 lines that 403 upstream.
	skip := []string{
		"6.4.20",
		"6.4.0.0-RC1",
		"6.3.5.3",
		"6.2.3",
		"6.1.0",
		"6.1.0-rc4",
	}
	for _, v := range skip {
		assert.Falsef(t, shouldFetchChangelog(v), "expected %q to be skipped", v)
	}

	// Malformed versions are never fetched.
	for _, v := range []string{"../secret", "latest", ""} {
		assert.Falsef(t, shouldFetchChangelog(v), "expected malformed %q to be skipped", v)
	}
}
