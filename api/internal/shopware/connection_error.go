package shopware

import (
	"crypto/x509"
	"errors"
	"net"
	"net/http"
)

// ConnectionErrorMessage returns only fixed guidance, never URLs, credentials,
// upstream response bodies or arbitrary transport error strings.
func ConnectionErrorMessage(err error) string {
	var certificate x509.CertificateInvalidError
	if errors.As(err, &certificate) && certificate.Reason == x509.Expired {
		return "The shop TLS certificate has expired or is not yet valid. Renew the certificate and check the server clock."
	}
	var unknownAuthority x509.UnknownAuthorityError
	var hostname x509.HostnameError
	if errors.As(err, &certificate) || errors.As(err, &unknownAuthority) || errors.As(err, &hostname) {
		return "The shop TLS certificate could not be verified. Check its hostname, certificate chain and validity."
	}
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode >= 300 && apiErr.StatusCode < 400 {
			return "The shop redirects API requests. Use the canonical HTTPS shop URL; OAuth POST requests must not be redirected to GET."
		}
		if apiErr.OAuthCode == "invalid_client" {
			return "The shop rejected the integration access keys. Check the integration access key ID and secret, and that the integration is active."
		}
		if apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden || apiErr.OAuthCode == "access_denied" {
			return "The shop rejected authentication or access. Check the integration permissions and any HTTP authentication or firewall protecting the shop API."
		}
		if apiErr.StatusCode == http.StatusMethodNotAllowed {
			return "The shop API rejected the HTTP method. Check the canonical shop URL and redirects; OAuth requires POST."
		}
	}
	var network net.Error
	if errors.As(err, &network) && network.Timeout() {
		return "The shop connection timed out. Check shop availability, firewall rules and upstream response times, then retry."
	}
	return "Cannot reach shop. Check the shop URL, network availability and integration credentials."
}
