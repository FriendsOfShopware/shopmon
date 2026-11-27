package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testKey = "0123456789abcdef0123456789abcdef"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	encrypted, err := Encrypt("slack://default?webhook_url=x", testKey)
	require.NoError(t, err)

	plaintext, err := Decrypt(encrypted, testKey)
	require.NoError(t, err)
	assert.Equal(t, "slack://default?webhook_url=x", plaintext)
}

func TestDecryptRejectsWrongIVLength(t *testing.T) {
	assert.NotPanics(t, func() {
		_, err := Decrypt("AA==:AA==", testKey)
		require.Error(t, err)
	})
}
