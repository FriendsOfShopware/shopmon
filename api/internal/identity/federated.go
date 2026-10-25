package identity

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrFederatedAccountInUse    = errors.New("federated account is already linked to another user")
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

type FederatedRepository interface {
	Provision(ctx context.Context, profile FederatedProfile) (string, error)
	Link(ctx context.Context, userID string, profile FederatedProfile) error
}

type FederatedService struct {
	repository FederatedRepository
}

func NewFederatedService(repository FederatedRepository) *FederatedService {
	return &FederatedService{repository: repository}
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

// Link attaches an explicitly approved provider identity to the authenticated user.
// Callers must bind approval to the session that initiated the provider flow.
func (s *FederatedService) Link(ctx context.Context, userID string, profile FederatedProfile) error {
	if userID == "" {
		return ErrUserNotFound
	}
	if profile.Email == "" {
		return ErrFederatedEmailRequired
	}
	if !profile.EmailVerified {
		return ErrFederatedEmailUnverified
	}
	if err := s.repository.Link(ctx, userID, profile); err != nil {
		return fmt.Errorf("link federated identity: %w", err)
	}
	return nil
}
