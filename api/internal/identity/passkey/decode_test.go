package passkey

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

// Sample ES256 COSE key from go-webauthn's assertion tests.
const cosePublicKeyRawURL = "pQMmIAEhWCAoCF-x0dwEhzQo-ABxHIAgr_5WL6cJceREc81oIwFn7iJYIHEHx8ZhBIE42L26-rSC_3l0ZaWEmsHAKyP9rgslApUdAQI"

func TestDecodeCOSEPublicKey(t *testing.T) {
	t.Parallel()

	raw, err := base64.RawURLEncoding.DecodeString(cosePublicKeyRawURL)
	require.NoError(t, err)

	cases := []struct {
		name  string
		value string
	}{
		{name: "raw-url", value: cosePublicKeyRawURL},
		{name: "standard padded", value: base64.StdEncoding.EncodeToString(raw)},
		{name: "standard raw", value: base64.RawStdEncoding.EncodeToString(raw)},
		{name: "url padded", value: base64.URLEncoding.EncodeToString(raw)},
		{name: "hex", value: hex.EncodeToString(raw)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := decodeCOSEPublicKey(tc.value)
			require.NoError(t, err)
			require.Equal(t, raw, decoded)
		})
	}
}

func TestDecodeCOSEPublicKeyRejectsGarbage(t *testing.T) {
	t.Parallel()

	_, err := decodeCOSEPublicKey("not-a-cose-key")
	require.Error(t, err)
}
