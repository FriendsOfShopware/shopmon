package shopware

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const missingPrivilegeBody = `{"errors":[{"status":"403","code":"FRAMEWORK__MISSING_PRIVILEGE_ERROR","title":"Forbidden","detail":"{\"message\":\"Missing privilege\",\"missingPrivileges\":[\"plugin:read\"]}","meta":{"parameters":{}}}]}`

const missingPrivilegeBodyEscaped = `{"errors":[{"status":"403","code":"FRAMEWORK__MISSING_PRIVILEGE_ERROR","title":"Forbidden","detail":"{\u0022message\u0022:\u0022Missing privilege\u0022,\u0022missingPrivileges\u0022:[\u0022message_queue_stats:read\u0022]}"}]}`

func TestClassifyFailure(t *testing.T) {
	var probe any
	syntaxErr := json.Unmarshal([]byte("<html>"), &probe)
	require.Error(t, syntaxErr)

	for _, tc := range []struct {
		name string
		err  error
		want FailureClass
		det  bool
	}{
		{"missing privilege", &ApiError{StatusCode: 403, Body: missingPrivilegeBody}, FailureMissingPrivilege, true},
		{"missing privilege escaped", &ApiError{StatusCode: 403, Body: missingPrivilegeBodyEscaped}, FailureMissingPrivilege, true},
		{"waf forbidden", &ApiError{StatusCode: 403, Body: "<!DOCTYPE html><title>Sucuri</title>"}, FailureForbidden, true},
		{"invalid client", &ApiError{StatusCode: 401, OAuthCode: "invalid_client"}, FailureInvalidCredentials, true},
		{"unauthorized", &ApiError{StatusCode: 401, Body: `{"errors":[{"code":"4","status":"401","title":"Client authentication failed"}]}`}, FailureInvalidCredentials, true},
		{"redirect token", &ApiError{StatusCode: 302}, FailureRedirect, true},
		{"redirect api", &ApiError{StatusCode: 301, Body: "redirect detected"}, FailureRedirect, true},
		{"not found", &ApiError{StatusCode: 404}, FailureNotFound, true},
		{"method not allowed", &ApiError{StatusCode: 405}, FailureNotFound, true},
		{"shop 500", &ApiError{StatusCode: 500, Body: `{"errors":[{"code":"0","status":"500","title":"Internal Server Error","detail":"boom"}]}`}, FailureShopError, false},
		{"shop 502 html", &ApiError{StatusCode: 502, Body: "<html>502 Bad Gateway</html>"}, FailureShopError, false},
		{"rate limited", &ApiError{StatusCode: 429}, FailureShopError, false},
		{"timeout", fmt.Errorf("token request: %w", context.DeadlineExceeded), FailureTimeout, false},
		{"tls unknown authority", x509.UnknownAuthorityError{}, FailureTLS, true},
		{"tls expired", x509.CertificateInvalidError{Reason: x509.Expired}, FailureTLS, true},
		{"dns", &net.DNSError{IsNotFound: true, Name: "shop.example"}, FailureUnreachable, false},
		{"invalid json", fmt.Errorf("decode token response: %w", syntaxErr), FailureInvalidResponse, false},
		{"unknown", fmt.Errorf("something else"), FailureUnknown, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := ClassifyFailure(fmt.Errorf("wrap: %w", tc.err))
			assert.Equal(t, tc.want, f.Class)
			assert.Equal(t, tc.det, f.Deterministic())
		})
	}
}

func TestMissingPrivileges(t *testing.T) {
	assert.Equal(t, []string{"plugin:read"}, MissingPrivileges(&ApiError{StatusCode: 403, Body: missingPrivilegeBody}))
	assert.Equal(t, []string{"message_queue_stats:read"}, MissingPrivileges(&ApiError{StatusCode: 403, Body: missingPrivilegeBodyEscaped}))
	assert.Empty(t, MissingPrivileges(&ApiError{StatusCode: 403, Body: "<html>blocked</html>"}))
	assert.Empty(t, MissingPrivileges(fmt.Errorf("nope")))

	// Shop-controlled values that are not privilege tokens must be dropped:
	// they end up in notifications and the database.
	body := `{"errors":[{"code":"FRAMEWORK__MISSING_PRIVILEGE_ERROR","detail":"{\"missingPrivileges\":[\"plugin:read\",\"<script>alert(1)</script>\",\"no colon\",\"app:read\"]}"}]}`
	assert.Equal(t, []string{"app:read", "plugin:read"}, MissingPrivileges(&ApiError{StatusCode: 403, Body: body}))
}

func TestConnectionErrorMessageMissingPrivilege(t *testing.T) {
	msg := ConnectionErrorMessage(&ApiError{StatusCode: 403, Body: missingPrivilegeBody})
	assert.Contains(t, msg, "plugin:read")
	assert.Contains(t, msg, "Integrations")
}
