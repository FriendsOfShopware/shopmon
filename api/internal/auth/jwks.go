package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
)

// ssoClaims holds the OIDC ID token claims we read beyond the standard ones
// (sub/iss/aud/exp/iat) that go-oidc validates and exposes on *oidc.IDToken.
type ssoClaims struct {
	Email         string      `json:"email"`
	EmailVerified interface{} `json:"email_verified"`
	Name          string      `json:"name"`
	Picture       string      `json:"picture"`
	Nonce         string      `json:"nonce"`
}

// verifyIDToken verifies an OIDC ID token's RS256 signature against the
// provider's JWKS (fetched from jwksURL) and validates the issuer, audience
// (client id), and expiry. It additionally enforces that the OIDC-required sub
// and iat claims are present. The supplied client (SSRF-safe in production) is
// used for the JWKS fetch and its keys are cached across calls by go-oidc.
func verifyIDToken(ctx context.Context, client *http.Client, idToken, jwksURL, expectedIssuer, expectedAudience string) (*oidc.IDToken, ssoClaims, error) {
	var claims ssoClaims

	// Route go-oidc's JWKS fetch through our chosen HTTP client.
	ctx = oidc.ClientContext(ctx, client)

	// Use the operator-configured JWKS endpoint directly instead of discovery,
	// so the stored jwksEndpoint stays authoritative and we avoid an extra round
	// trip. The key set caches keys and refetches on an unknown kid.
	keySet := oidc.NewRemoteKeySet(ctx, jwksURL)
	verifier := oidc.NewVerifier(expectedIssuer, keySet, &oidc.Config{
		ClientID:             expectedAudience,
		SupportedSigningAlgs: []string{oidc.RS256},
	})

	token, err := verifier.Verify(ctx, idToken)
	if err != nil {
		return nil, claims, fmt.Errorf("validate ID token: %w", err)
	}

	// go-oidc validates sig/iss/aud/exp but does not require sub or iat to be
	// present. sub is the stable subject we key the account on, and iat is
	// OIDC-required, so enforce both.
	if token.Subject == "" {
		return nil, claims, fmt.Errorf("ID token missing required sub claim")
	}
	if token.IssuedAt.IsZero() {
		return nil, claims, fmt.Errorf("ID token missing required iat claim")
	}

	if err := token.Claims(&claims); err != nil {
		return nil, claims, fmt.Errorf("parse ID token claims: %w", err)
	}

	return token, claims, nil
}
