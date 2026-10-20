package sso

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/organization"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrInvalidConfiguration = errors.New("invalid SSO provider configuration")
	ErrIssuerRequired       = errors.New("issuer is required")
	ErrInvalidIssuer        = errors.New("invalid issuer URL")
	ErrDiscoveryFailed      = errors.New("OIDC discovery failed")
	ErrProviderNotFound     = errors.New("SSO provider not found")
)

// ValidationError identifies a tenant-supplied OIDC endpoint that cannot be
// stored safely. Cause is suitable for returning in a validation response.
type ValidationError struct {
	Cause error
}

func (e *ValidationError) Error() string {
	return "invalid OIDC endpoint: " + e.Cause.Error()
}

func (e *ValidationError) Unwrap() error {
	return ErrInvalidConfiguration
}

type ProviderConfiguration struct {
	AuthorizationEndpoint string  `json:"authorizationEndpoint"`
	TokenEndpoint         string  `json:"tokenEndpoint"`
	JWKSEndpoint          string  `json:"jwksEndpoint"`
	ClientID              string  `json:"clientId"`
	ClientSecret          *string `json:"clientSecret,omitempty"`
}

type Provider struct {
	ID             string
	Issuer         string
	Domain         string
	OrganizationID *string
	Configuration  ProviderConfiguration
}

type Discovery struct {
	Issuer                string
	AuthorizationEndpoint string
	TokenEndpoint         string
	JWKSEndpoint          string
	UserInfoEndpoint      string
	Scopes                []string
}

type Authorizer interface {
	Membership(ctx context.Context, userID, organizationID string) (organization.Role, error)
	RequireRole(ctx context.Context, userID, organizationID string, allowedRoles ...organization.Role) (organization.Role, error)
}

type OIDCGateway interface {
	ValidateConfiguration(configuration ProviderConfiguration, issuer string) error
	Discover(ctx context.Context, issuer string) (Discovery, error)
}

type Service struct {
	queries    *queries.Queries
	authorizer Authorizer
	oidc       OIDCGateway
}

func NewService(q *queries.Queries, authorizer Authorizer, oidc OIDCGateway) *Service {
	return &Service{
		queries:    q,
		authorizer: authorizer,
		oidc:       oidc,
	}
}

func (s *Service) List(ctx context.Context, userID, organizationID string) ([]Provider, error) {
	if _, err := s.authorizer.Membership(ctx, userID, organizationID); err != nil {
		return nil, err
	}

	rows, err := s.queries.ListSSOProviders(ctx, &organizationID)
	if err != nil {
		return nil, fmt.Errorf("list SSO providers: %w", err)
	}

	providers := make([]Provider, 0, len(rows))
	for _, row := range rows {
		provider, err := decodeProvider(row)
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}
	return providers, nil
}

type UpdateCommand struct {
	UserID         string
	OrganizationID string
	ProviderID     string
	Issuer         string
	Domain         string
	Configuration  ProviderConfiguration
}

type CreateCommand struct {
	UserID         string
	OrganizationID string
	Issuer         string
	Domain         string
	Configuration  ProviderConfiguration
}

func (s *Service) Create(ctx context.Context, command CreateCommand) error {
	if _, err := s.authorizer.RequireRole(ctx, command.UserID, command.OrganizationID, organization.RoleOwner, organization.RoleAdmin); err != nil {
		return err
	}
	if err := s.oidc.ValidateConfiguration(command.Configuration, command.Issuer); err != nil {
		return fmt.Errorf("validate SSO provider: %w", err)
	}
	providerID := "sso-" + command.Domain
	encoded, err := encodeConfiguration(command.Configuration)
	if err != nil {
		return err
	}
	if err := s.queries.CreateSSOProvider(ctx, queries.CreateSSOProviderParams{
		ID:             uuid.NewString(),
		Issuer:         command.Issuer,
		OidcConfig:     &encoded,
		ProviderID:     providerID,
		OrganizationID: &command.OrganizationID,
		Domain:         command.Domain,
	}); err != nil {
		return fmt.Errorf("create SSO provider: %w", err)
	}
	return nil
}

func (s *Service) Update(ctx context.Context, command UpdateCommand) error {
	if _, err := s.authorizer.RequireRole(ctx, command.UserID, command.OrganizationID, organization.RoleOwner, organization.RoleAdmin); err != nil {
		return err
	}
	if err := s.oidc.ValidateConfiguration(command.Configuration, command.Issuer); err != nil {
		return fmt.Errorf("validate SSO provider: %w", err)
	}
	encoded, err := encodeConfiguration(command.Configuration)
	if err != nil {
		return err
	}
	if err := s.queries.UpdateSSOProvider(ctx, queries.UpdateSSOProviderParams{
		Issuer:         command.Issuer,
		OidcConfig:     &encoded,
		Domain:         command.Domain,
		ProviderID:     command.ProviderID,
		OrganizationID: &command.OrganizationID,
	}); err != nil {
		return fmt.Errorf("update SSO provider: %w", err)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, userID, organizationID, providerID string) error {
	if _, err := s.authorizer.RequireRole(ctx, userID, organizationID, organization.RoleOwner, organization.RoleAdmin); err != nil {
		return err
	}
	if err := s.queries.DeleteSSOProvider(ctx, queries.DeleteSSOProviderParams{
		ProviderID:     providerID,
		OrganizationID: &organizationID,
	}); err != nil {
		return fmt.Errorf("delete SSO provider: %w", err)
	}
	return nil
}

func (s *Service) Discover(ctx context.Context, issuer string) (Discovery, error) {
	if issuer == "" {
		return Discovery{}, ErrIssuerRequired
	}
	discovery, err := s.oidc.Discover(ctx, issuer)
	if err != nil {
		return Discovery{}, fmt.Errorf("discover SSO provider: %w", err)
	}
	if discovery.Scopes == nil {
		discovery.Scopes = []string{}
	}
	return discovery, nil
}

func (s *Service) FindByDomain(ctx context.Context, domain string) (Provider, error) {
	row, err := s.queries.GetSSOProviderByDomain(ctx, domain)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Provider{}, ErrProviderNotFound
		}
		return Provider{}, fmt.Errorf("find SSO provider by domain: %w", err)
	}
	return decodeProvider(row)
}

func (s *Service) FindByID(ctx context.Context, providerID string) (Provider, error) {
	row, err := s.queries.GetSSOProviderByProviderID(ctx, providerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Provider{}, ErrProviderNotFound
		}
		return Provider{}, fmt.Errorf("find SSO provider by ID: %w", err)
	}
	return decodeProvider(row)
}

func encodeConfiguration(configuration ProviderConfiguration) (string, error) {
	encoded, err := json.Marshal(configuration)
	if err != nil {
		return "", fmt.Errorf("encode OIDC configuration: %w", err)
	}
	return string(encoded), nil
}

func decodeProvider(row queries.SsoProvider) (Provider, error) {
	provider := Provider{
		ID: row.ProviderID, Issuer: row.Issuer, Domain: row.Domain, OrganizationID: row.OrganizationID,
	}
	if row.OidcConfig != nil {
		if err := json.Unmarshal([]byte(*row.OidcConfig), &provider.Configuration); err != nil {
			return Provider{}, fmt.Errorf("decode OIDC configuration for provider %q: %w", row.ProviderID, err)
		}
	}
	return provider, nil
}
