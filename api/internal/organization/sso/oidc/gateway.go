package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	organizationsso "github.com/friendsofshopware/shopmon/api/internal/organization/sso"
)

const maxDiscoveryDocumentSize = 1 << 20

type Gateway struct {
	timeout time.Duration
}

func NewGateway(timeout time.Duration) *Gateway {
	return &Gateway{timeout: timeout}
}

func (g *Gateway) ValidateConfiguration(configuration organizationsso.ProviderConfiguration, issuer string) error {
	endpoints := []string{
		issuer,
		configuration.AuthorizationEndpoint,
		configuration.TokenEndpoint,
		configuration.JWKSEndpoint,
	}
	for _, endpoint := range endpoints {
		if err := httputil.ValidateHTTPSEndpoint(endpoint); err != nil {
			return &organizationsso.ValidationError{Cause: err}
		}
	}
	return nil
}

func (g *Gateway) Discover(ctx context.Context, issuer string) (organizationsso.Discovery, error) {
	issuerURL, err := url.Parse(issuer)
	if err != nil || issuerURL.Host == "" {
		return organizationsso.Discovery{}, organizationsso.ErrInvalidIssuer
	}
	if issuerURL.User != nil || issuerURL.RawQuery != "" || issuerURL.Fragment != "" {
		return organizationsso.Discovery{}, organizationsso.ErrInvalidIssuer
	}

	isLoopback := httputil.IsLoopbackURL(issuer)
	isHTTPS := issuerURL.Scheme == "https"
	isLoopbackHTTP := isLoopback && issuerURL.Scheme == "http"
	if !isHTTPS && !isLoopbackHTTP {
		return organizationsso.Discovery{}, organizationsso.ErrInvalidIssuer
	}

	discoveryURL := *issuerURL
	discoveryURL.Path = strings.TrimSuffix(discoveryURL.Path, "/") + "/.well-known/openid-configuration"

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL.String(), nil)
	if err != nil {
		return organizationsso.Discovery{}, fmt.Errorf("%w: create request: %v", organizationsso.ErrDiscoveryFailed, err)
	}

	response, err := httputil.ClientForURL(discoveryURL.String(), g.timeout).Do(request)
	if err != nil {
		return organizationsso.Discovery{}, fmt.Errorf("%w: fetch discovery document: %v", organizationsso.ErrDiscoveryFailed, err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return organizationsso.Discovery{}, fmt.Errorf("%w: endpoint returned status %d", organizationsso.ErrDiscoveryFailed, response.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxDiscoveryDocumentSize))
	if err != nil {
		return organizationsso.Discovery{}, fmt.Errorf("%w: read discovery document: %v", organizationsso.ErrDiscoveryFailed, err)
	}

	var document struct {
		Issuer                string   `json:"issuer"`
		AuthorizationEndpoint string   `json:"authorization_endpoint"`
		TokenEndpoint         string   `json:"token_endpoint"`
		JWKSURI               string   `json:"jwks_uri"`
		UserInfoEndpoint      string   `json:"userinfo_endpoint"`
		ScopesSupported       []string `json:"scopes_supported"`
	}
	if err := json.Unmarshal(body, &document); err != nil {
		return organizationsso.Discovery{}, fmt.Errorf("%w: decode discovery document: %v", organizationsso.ErrDiscoveryFailed, err)
	}

	return organizationsso.Discovery{
		Issuer:                document.Issuer,
		AuthorizationEndpoint: document.AuthorizationEndpoint,
		TokenEndpoint:         document.TokenEndpoint,
		JWKSEndpoint:          document.JWKSURI,
		UserInfoEndpoint:      document.UserInfoEndpoint,
		Scopes:                document.ScopesSupported,
	}, nil
}
