package shopware

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

// FailureClass is the owner-facing category of a failed shop request. The
// scrape job maps classes to translation keys and notification policy, so the
// set stays small and every value must have a user-actionable (or at least
// user-explainable) meaning.
type FailureClass string

const (
	FailureMissingPrivilege   FailureClass = "missing_privilege"
	FailureInvalidCredentials FailureClass = "invalid_credentials"
	FailureForbidden          FailureClass = "forbidden"
	FailureNotFound           FailureClass = "not_found"
	FailureRedirect           FailureClass = "redirect"
	FailureTimeout            FailureClass = "timeout"
	FailureTLS                FailureClass = "tls"
	FailureUnreachable        FailureClass = "unreachable"
	FailureShopError          FailureClass = "shop_error"
	FailureInvalidResponse    FailureClass = "invalid_response"
	FailureUnknown            FailureClass = "unknown"
)

// Failure is the classified result of one failed shop request.
type Failure struct {
	Class             FailureClass
	StatusCode        int
	MissingPrivileges []string
}

// Deterministic reports whether the failure cannot heal on its own: it needs
// a change in the shop (permissions, credentials, URL, certificate, firewall)
// rather than time. Deterministic failures notify on first sight; transient
// ones only after repeated consecutive scrapes.
func (f Failure) Deterministic() bool {
	switch f.Class {
	case FailureMissingPrivilege, FailureInvalidCredentials, FailureForbidden,
		FailureNotFound, FailureRedirect, FailureTLS:
		return true
	default:
		return false
	}
}

// ClassifyFailure reduces an arbitrary shop request error to a Failure. It is
// the one place that interprets upstream status codes and bodies; callers must
// not surface raw errors to users (bodies are shop-controlled).
func ClassifyFailure(err error) Failure {
	if err == nil {
		return Failure{Class: FailureUnknown}
	}

	var certificate x509.CertificateInvalidError
	var unknownAuthority x509.UnknownAuthorityError
	var hostname x509.HostnameError
	if errors.As(err, &certificate) || errors.As(err, &unknownAuthority) || errors.As(err, &hostname) {
		return Failure{Class: FailureTLS}
	}

	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		return classifyAPIError(apiErr)
	}

	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &syntaxErr) || errors.As(err, &typeErr) {
		return Failure{Class: FailureInvalidResponse}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return Failure{Class: FailureTimeout}
	}
	var network net.Error
	if errors.As(err, &network) && network.Timeout() {
		return Failure{Class: FailureTimeout}
	}
	var dnsErr *net.DNSError
	var opErr *net.OpError
	if errors.As(err, &dnsErr) || errors.As(err, &opErr) {
		return Failure{Class: FailureUnreachable}
	}

	return Failure{Class: FailureUnknown}
}

func classifyAPIError(apiErr *ApiError) Failure {
	f := Failure{StatusCode: apiErr.StatusCode}
	switch {
	case apiErr.StatusCode >= 300 && apiErr.StatusCode < 400:
		f.Class = FailureRedirect
	case apiErr.OAuthCode == "invalid_client" || apiErr.OAuthCode == "access_denied":
		f.Class = FailureInvalidCredentials
	case apiErr.StatusCode == http.StatusUnauthorized:
		f.Class = FailureInvalidCredentials
	case apiErr.StatusCode == http.StatusForbidden:
		f.MissingPrivileges = parseMissingPrivileges(apiErr.Body)
		if len(f.MissingPrivileges) > 0 {
			f.Class = FailureMissingPrivilege
		} else {
			f.Class = FailureForbidden
		}
	case apiErr.StatusCode == http.StatusNotFound || apiErr.StatusCode == http.StatusMethodNotAllowed:
		f.Class = FailureNotFound
	case apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode >= 500:
		f.Class = FailureShopError
	default:
		f.Class = FailureUnknown
	}
	return f
}

// privilegeRe bounds the shop-controlled privilege strings that may reach
// notifications and the database to actual Shopware ACL token shapes.
var privilegeRe = regexp.MustCompile(`^[a-z0-9_]+(:[a-z0-9_]+)+$`)

const maxMissingPrivileges = 20

// MissingPrivileges extracts the ACL privileges a 403 response reports as
// missing, sorted and deduplicated. It returns nil for any other error.
func MissingPrivileges(err error) []string {
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		return nil
	}
	return parseMissingPrivileges(apiErr.Body)
}

// parseMissingPrivileges reads Shopware's FRAMEWORK__MISSING_PRIVILEGE_ERROR
// payload: errors[].detail is itself a JSON string carrying
// {"message":"Missing privilege","missingPrivileges":[...]}. Some versions
// embed the object directly; both shapes are accepted.
func parseMissingPrivileges(body string) []string {
	if body == "" || !strings.Contains(body, "missingPrivileges") {
		return nil
	}
	var payload struct {
		Errors []struct {
			Detail json.RawMessage `json:"detail"`
		} `json:"errors"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return nil
	}

	seen := map[string]bool{}
	var out []string
	for _, e := range payload.Errors {
		var detail struct {
			MissingPrivileges []string `json:"missingPrivileges"`
		}
		raw := e.Detail
		var asString string
		if err := json.Unmarshal(raw, &asString); err == nil {
			raw = []byte(asString)
		}
		if err := json.Unmarshal(raw, &detail); err != nil {
			continue
		}
		for _, p := range detail.MissingPrivileges {
			p = strings.ToLower(strings.TrimSpace(p))
			if len(p) > 128 || !privilegeRe.MatchString(p) || seen[p] {
				continue
			}
			seen[p] = true
			out = append(out, p)
			if len(out) >= maxMissingPrivileges {
				sort.Strings(out)
				return out
			}
		}
	}
	sort.Strings(out)
	return out
}
