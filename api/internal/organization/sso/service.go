package sso

import (
	"context"
	"errors"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/organization"
	"github.com/google/uuid"
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

type Repository interface {
	List(ctx context.Context, organizationID string) ([]Provider, error)
	Create(ctx context.Context, id, organizationID, providerID, issuer, domain string, configuration ProviderConfiguration) error
	Update(ctx context.Context, organizationID, providerID, issuer, domain string, configuration ProviderConfiguration) error
	Delete(ctx context.Context, organizationID, providerID string) error
	FindByDomain(ctx context.Context, domain string) (Provider, error)
	FindByID(ctx context.Context, providerID string) (Provider, error)
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
	repository Repository
	authorizer Authorizer
	oidc       OIDCGateway
}

func NewService(repository Repository, authorizer Authorizer, oidc OIDCGateway) *Service {
	return &Service{
		repository: repository,
		authorizer: authorizer,
		oidc:       oidc,
	}
}

func (s *Service) List(ctx context.Context, userID, organizationID string) ([]Provider, error) {
	if _, err := s.authorizer.Membership(ctx, userID, organizationID); err != nil {
		return nil, err
	}

	providers, err := s.repository.List(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list SSO providers: %w", err)
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
	if err := s.repository.Create(ctx, uuid.NewString(), command.OrganizationID, providerID, command.Issuer, command.Domain, command.Configuration); err != nil {
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
	if err := s.repository.Update(ctx, command.OrganizationID, command.ProviderID, command.Issuer, command.Domain, command.Configuration); err != nil {
		return fmt.Errorf("update SSO provider: %w", err)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, userID, organizationID, providerID string) error {
	if _, err := s.authorizer.RequireRole(ctx, userID, organizationID, organization.RoleOwner, organization.RoleAdmin); err != nil {
		return err
	}
	if err := s.repository.Delete(ctx, organizationID, providerID); err != nil {
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
	provider, err := s.repository.FindByDomain(ctx, domain)
	if err != nil {
		return Provider{}, fmt.Errorf("find SSO provider by domain: %w", err)
	}
	return provider, nil
}

func (s *Service) FindByID(ctx context.Context, providerID string) (Provider, error) {
	provider, err := s.repository.FindByID(ctx, providerID)
	if err != nil {
		return Provider{}, fmt.Errorf("find SSO provider by ID: %w", err)
	}
	return provider, nil
}
