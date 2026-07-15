package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	organizationsso "github.com/friendsofshopware/shopmon/api/internal/organization/sso"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	queries *queries.Queries
}

func NewRepository(q *queries.Queries) *Repository {
	return &Repository{queries: q}
}

func (r *Repository) List(ctx context.Context, organizationID string) ([]organizationsso.Provider, error) {
	rows, err := r.queries.ListSSOProviders(ctx, &organizationID)
	if err != nil {
		return nil, fmt.Errorf("query SSO providers: %w", err)
	}

	providers := make([]organizationsso.Provider, 0, len(rows))
	for _, row := range rows {
		provider, err := decodeProvider(row)
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}
	return providers, nil
}

func (r *Repository) FindByDomain(ctx context.Context, domain string) (organizationsso.Provider, error) {
	row, err := r.queries.GetSSOProviderByDomain(ctx, domain)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return organizationsso.Provider{}, organizationsso.ErrProviderNotFound
		}
		return organizationsso.Provider{}, fmt.Errorf("query SSO provider by domain: %w", err)
	}
	return decodeProvider(row)
}

func (r *Repository) FindByID(ctx context.Context, providerID string) (organizationsso.Provider, error) {
	row, err := r.queries.GetSSOProviderByProviderID(ctx, providerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return organizationsso.Provider{}, organizationsso.ErrProviderNotFound
		}
		return organizationsso.Provider{}, fmt.Errorf("query SSO provider by ID: %w", err)
	}
	return decodeProvider(row)
}

func (r *Repository) Create(ctx context.Context, id, organizationID, providerID, issuer, domain string, configuration organizationsso.ProviderConfiguration) error {
	encoded, err := json.Marshal(configuration)
	if err != nil {
		return fmt.Errorf("encode OIDC configuration: %w", err)
	}
	encodedString := string(encoded)
	if err := r.queries.CreateSSOProvider(ctx, queries.CreateSSOProviderParams{
		ID:             id,
		Issuer:         issuer,
		OidcConfig:     &encodedString,
		ProviderID:     providerID,
		OrganizationID: &organizationID,
		Domain:         domain,
	}); err != nil {
		return fmt.Errorf("execute SSO provider insert: %w", err)
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, organizationID, providerID, issuer, domain string, configuration organizationsso.ProviderConfiguration) error {
	encoded, err := json.Marshal(configuration)
	if err != nil {
		return fmt.Errorf("encode OIDC configuration: %w", err)
	}
	encodedString := string(encoded)
	if err := r.queries.UpdateSSOProvider(ctx, queries.UpdateSSOProviderParams{
		Issuer:         issuer,
		OidcConfig:     &encodedString,
		Domain:         domain,
		ProviderID:     providerID,
		OrganizationID: &organizationID,
	}); err != nil {
		return fmt.Errorf("execute SSO provider update: %w", err)
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, organizationID, providerID string) error {
	if err := r.queries.DeleteSSOProvider(ctx, queries.DeleteSSOProviderParams{
		ProviderID:     providerID,
		OrganizationID: &organizationID,
	}); err != nil {
		return fmt.Errorf("execute SSO provider delete: %w", err)
	}
	return nil
}

func decodeProvider(row queries.SsoProvider) (organizationsso.Provider, error) {
	provider := organizationsso.Provider{
		ID: row.ProviderID, Issuer: row.Issuer, Domain: row.Domain, OrganizationID: row.OrganizationID,
	}
	if row.OidcConfig != nil {
		if err := json.Unmarshal([]byte(*row.OidcConfig), &provider.Configuration); err != nil {
			return organizationsso.Provider{}, fmt.Errorf("decode OIDC configuration for provider %q: %w", row.ProviderID, err)
		}
	}
	return provider, nil
}
