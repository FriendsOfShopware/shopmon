package packagesmirror

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrNotConfigured            = errors.New("packages API not configured")
	ErrNotAuthorized            = errors.New("not authorized for organization")
	ErrShopNotFound             = errors.New("shop not found")
	ErrShopOrganizationMismatch = errors.New("shop does not belong to organization")
	ErrTokenNotFound            = errors.New("packages token not found")
	ErrTokenRequired            = errors.New("packages token is required")
	ErrRemote                   = errors.New("packages API request failed")
)

type Config struct {
	BaseURL string
	Token   string
}

type Configuration struct {
	Configured  bool
	ComposerURL *string
}

type Token struct {
	ID           int
	Source       string
	LastSyncedAt *int64
}

// RemoteError preserves a response returned by the packages API. HTTP adapters
// may pass it through without coupling the application service to net/http.
type RemoteError struct {
	StatusCode int
	Body       []byte
}

func (e *RemoteError) Error() string {
	return fmt.Sprintf("%s with status %d", ErrRemote, e.StatusCode)
}

func (e *RemoteError) Unwrap() error {
	return ErrRemote
}

type ShopRepository interface {
	OrganizationIDForShop(ctx context.Context, shopID int32) (string, error)
}

type MembershipAuthorizer interface {
	IsMember(ctx context.Context, userID, organizationID string) (bool, error)
}

type Gateway interface {
	List(ctx context.Context, source string) ([]Token, error)
	Create(ctx context.Context, token, source string) error
	Delete(ctx context.Context, tokenID int) error
	Sync(ctx context.Context, tokenID int) error
}

type Service struct {
	shops      ShopRepository
	authorizer MembershipAuthorizer
	gateway    Gateway
	config     Config
}

func NewService(shops ShopRepository, authorizer MembershipAuthorizer, gateway Gateway, config Config) *Service {
	return &Service{shops: shops, authorizer: authorizer, gateway: gateway, config: config}
}

func (s *Service) Configuration() Configuration {
	configured := s.config.BaseURL != "" && s.config.Token != ""
	var composerURL *string
	if configured {
		value := s.config.BaseURL + "/composer"
		composerURL = &value
	}
	return Configuration{Configured: configured, ComposerURL: composerURL}
}

type ShopCommand struct {
	UserID         string
	OrganizationID string
	ShopID         int32
}

func (s *Service) List(ctx context.Context, cmd ShopCommand) ([]Token, error) {
	if err := s.authorizeShop(ctx, cmd); err != nil {
		return nil, err
	}
	tokens, err := s.gateway.List(ctx, source(cmd.ShopID))
	if err != nil {
		return nil, fmt.Errorf("list packages tokens: %w", err)
	}
	return tokens, nil
}

type CreateCommand struct {
	ShopCommand
	Token string
}

func (s *Service) Create(ctx context.Context, cmd CreateCommand) error {
	if err := s.authorizeShop(ctx, cmd.ShopCommand); err != nil {
		return err
	}
	if cmd.Token == "" {
		return ErrTokenRequired
	}
	if err := s.gateway.Create(ctx, cmd.Token, source(cmd.ShopID)); err != nil {
		return fmt.Errorf("create packages token: %w", err)
	}
	return nil
}

type TokenCommand struct {
	ShopCommand
	TokenID int
}

func (s *Service) Delete(ctx context.Context, cmd TokenCommand) error {
	if err := s.authorizeToken(ctx, cmd); err != nil {
		return err
	}
	if err := s.gateway.Delete(ctx, cmd.TokenID); err != nil {
		return fmt.Errorf("delete packages token: %w", err)
	}
	return nil
}

func (s *Service) Sync(ctx context.Context, cmd TokenCommand) error {
	if err := s.authorizeToken(ctx, cmd); err != nil {
		return err
	}
	if err := s.gateway.Sync(ctx, cmd.TokenID); err != nil {
		return fmt.Errorf("sync packages token: %w", err)
	}
	return nil
}

func (s *Service) authorizeToken(ctx context.Context, cmd TokenCommand) error {
	if err := s.authorizeShop(ctx, cmd.ShopCommand); err != nil {
		return err
	}
	tokens, err := s.gateway.List(ctx, source(cmd.ShopID))
	if err != nil {
		return fmt.Errorf("verify packages token ownership: %w", err)
	}
	for _, token := range tokens {
		if token.ID == cmd.TokenID {
			return nil
		}
	}
	return ErrTokenNotFound
}

func (s *Service) authorizeShop(ctx context.Context, cmd ShopCommand) error {
	if s.config.BaseURL == "" {
		return ErrNotConfigured
	}
	member, err := s.authorizer.IsMember(ctx, cmd.UserID, cmd.OrganizationID)
	if err != nil {
		return fmt.Errorf("authorize organization membership: %w", err)
	}
	if !member {
		return ErrNotAuthorized
	}
	organizationID, err := s.shops.OrganizationIDForShop(ctx, cmd.ShopID)
	if err != nil {
		return err
	}
	if organizationID != cmd.OrganizationID {
		return ErrShopOrganizationMismatch
	}
	return nil
}

// source retains compatibility with tokens created by the previous handler implementation.
func source(shopID int32) string {
	return fmt.Sprintf("shopmon-project-%d", shopID)
}
