package identity

import (
	"context"
	"errors"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrFederatedEmailRequired   = errors.New("federated identity did not provide an email")
	ErrFederatedEmailUnverified = errors.New("federated identity email is not verified")
)

type FederatedProfile struct {
	Provider       string
	AccountID      string
	Email          string
	EmailVerified  bool
	Name           string
	ImageURL       string
	OrganizationID *string
}

type FederatedService struct {
	repository *federatedStore
}

func NewFederatedService(pool *pgxpool.Pool, q *queries.Queries) *FederatedService {
	return &FederatedService{repository: &federatedStore{pool: pool, queries: q}}
}

func (s *FederatedService) Provision(ctx context.Context, profile FederatedProfile) (string, error) {
	if profile.Email == "" {
		return "", ErrFederatedEmailRequired
	}
	if !profile.EmailVerified {
		return "", ErrFederatedEmailUnverified
	}
	if profile.Name == "" {
		profile.Name = profile.Email
	}
	userID, err := s.repository.Provision(ctx, profile)
	if err != nil {
		return "", fmt.Errorf("provision federated identity: %w", err)
	}
	return userID, nil
}
