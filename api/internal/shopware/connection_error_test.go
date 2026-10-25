package shopware

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectionErrorGuidance(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want string
	}{
		{&ApiError{StatusCode: 401, OAuthCode: "invalid_client", Body: "sensitive-fixture-value"}, "integration access keys"},
		{&ApiError{StatusCode: 403, Body: "sensitive-fixture-value"}, "rejected authentication or access"},
		{&ApiError{StatusCode: 301, Body: "sensitive-fixture-value"}, "canonical HTTPS shop URL"},
		{&ApiError{StatusCode: 405, Body: "sensitive-fixture-value"}, "OAuth requires POST"},
		{x509.CertificateInvalidError{Reason: x509.Expired}, "expired or is not yet valid"},
		{x509.UnknownAuthorityError{}, "could not be verified"},
		{context.DeadlineExceeded, "timed out"},
		{fmt.Errorf("sensitive-fixture-value"), "Cannot reach shop"},
	} {
		message := ConnectionErrorMessage(fmt.Errorf("connection: %w", tc.err))
		assert.Contains(t, message, tc.want)
		assert.NotContains(t, message, "sensitive-fixture-value")
	}
}

func TestOAuthRedirectDoesNotSendCredentialsOrChangeMethod(t *testing.T) {
	for _, status := range []int{301, 302, 303, 307, 308} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			targetCalls := 0
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { targetCalls++; w.WriteHeader(405) }))
			defer target.Close()
			source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				http.Redirect(w, r, target.URL, status)
			}))
			defer source.Close()
			err := newTestClient(source.URL).Authenticate(context.Background())
			require.Error(t, err)
			assert.Zero(t, targetCalls)
			assert.Contains(t, ConnectionErrorMessage(err), "canonical HTTPS shop URL")
		})
	}
}

func TestOAuthInvalidClientIsSanitized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`{"error":"invalid_client","error_description":"secret"}`))
	}))
	defer server.Close()
	err := newTestClient(server.URL).Authenticate(context.Background())
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "secret")
	assert.Contains(t, ConnectionErrorMessage(err), "integration access keys")
}

func TestUntrustedTLSRemainsRejected(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("untrusted TLS request reached server") }))
	defer server.Close()
	err := newTestClient(server.URL).Authenticate(context.Background())
	require.Error(t, err)
	assert.Contains(t, ConnectionErrorMessage(err), "could not be verified")
}
