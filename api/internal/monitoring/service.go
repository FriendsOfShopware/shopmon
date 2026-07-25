package monitoring

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

var (
	ErrNotAuthorized              = errors.New("not authorized for organization")
	ErrShopNotFound               = errors.New("shop not found")
	ErrEnvironmentNotFound        = errors.New("environment not found")
	ErrShopOrganizationMismatch   = errors.New("shop does not belong to organization")
	ErrDefaultEnvironmentNotFound = errors.New("default environment not found")
	ErrDefaultEnvironmentMismatch = errors.New("default environment does not belong to shop")
	ErrShopHasEnvironments        = errors.New("shop has existing environments")
	ErrConnectionFailed           = errors.New("shop connection failed")
	ErrCredentialDecryption       = errors.New("credential decryption failed")
	ErrRemoteOperation            = errors.New("remote environment operation failed")
)

const deploymentOutputCleanupTimeout = 30 * time.Second

type Shop struct {
	ID                   int32
	OrganizationID       string
	Name                 string
	Description          *string
	GitURL               *string
	DefaultEnvironmentID *int32
}

type Environment struct {
	ID                    int32
	OrganizationID        string
	ShopID                int32
	Name                  string
	URL                   string
	ClientID              string
	EncryptedClientSecret string
	ShopwareVersion       string
	EnvironmentToken      string
	Ignores               []string
}

type CreateEnvironmentRecord struct {
	OrganizationID        string
	ShopID                int32
	Name                  string
	URL                   string
	ClientID              string
	EncryptedClientSecret string
	ShopwareVersion       string
	EnvironmentToken      string
}

type CreateShopRecord struct {
	OrganizationID string
	Name           string
	Description    *string
	GitURL         *string
}

type Repository interface {
	FindShop(ctx context.Context, shopID int32) (Shop, error)
	FindEnvironment(ctx context.Context, environmentID int32) (Environment, error)
	CreateShopWithEnvironment(ctx context.Context, shop CreateShopRecord, environment CreateEnvironmentRecord) (shopID, environmentID int32, err error)
	CreateEnvironment(ctx context.Context, environment CreateEnvironmentRecord) (int32, error)
	UpdateShop(ctx context.Context, shop Shop, defaultEnvironmentID *int32) error
	CountShopEnvironments(ctx context.Context, shopID int32) (int32, error)
	DeleteShop(ctx context.Context, shopID int32, organizationID string) error
	UpdateEnvironment(ctx context.Context, environment Environment) error
	DeleteEnvironment(ctx context.Context, environmentID int32) error
	ListDeploymentIDs(ctx context.Context, environmentID int32) ([]int32, error)
	UpdateSitespeedSettings(ctx context.Context, environmentID int32, enabled bool, urls []string) error
	SubscribeEnvironment(ctx context.Context, userID string, environmentID int32) error
	UnsubscribeEnvironment(ctx context.Context, userID string, environmentID int32) error
	IsEnvironmentSubscribed(ctx context.Context, userID string, environmentID int32) (bool, error)
}

type MembershipAuthorizer interface {
	IsMember(ctx context.Context, userID, organizationID string) (bool, error)
}

type ConnectionCredentials struct {
	URL              string
	ClientID         string
	ClientSecret     string
	EnvironmentToken string
}

type StoredCredentials struct {
	URL                   string
	ClientID              string
	EncryptedClientSecret string
	EnvironmentToken      string
}

type ShopInfo struct {
	Version string `json:"version"`
}

type EnvironmentGateway interface {
	ValidateConnection(ctx context.Context, credentials ConnectionCredentials) (ShopInfo, error)
	EncryptSecret(secret string) (string, error)
	ClearCache(ctx context.Context, credentials StoredCredentials) error
	RescheduleTask(ctx context.Context, credentials StoredCredentials, taskID string, nextExecution time.Time) error
}

type JobDispatcher interface {
	DispatchEnvironmentScrape(ctx context.Context, environmentID int32) error
	DispatchSitespeedScrape(ctx context.Context, environmentID int32) error
}

type DeploymentOutputStore interface {
	DeleteDeploymentOutput(ctx context.Context, deploymentID int) error
}

type Service struct {
	repository   Repository
	authorizer   MembershipAuthorizer
	environments EnvironmentGateway
	jobs         JobDispatcher
	outputs      DeploymentOutputStore
}

func NewService(repository Repository, authorizer MembershipAuthorizer, environments EnvironmentGateway, jobs JobDispatcher, outputs DeploymentOutputStore) *Service {
	return &Service{
		repository:   repository,
		authorizer:   authorizer,
		environments: environments,
		jobs:         jobs,
		outputs:      outputs,
	}
}

type CreateShopCommand struct {
	UserID           string
	OrganizationID   string
	Name             string
	Description      *string
	GitURL           *string
	EnvironmentName  string
	EnvironmentURL   string
	ClientID         string
	ClientSecret     string
	EnvironmentToken string
}

type CreateShopResult struct {
	ShopID        int32
	EnvironmentID int32
}

func (s *Service) CreateShop(ctx context.Context, cmd CreateShopCommand) (CreateShopResult, error) {
	if err := s.authorize(ctx, cmd.UserID, cmd.OrganizationID); err != nil {
		return CreateShopResult{}, err
	}

	environmentToken, err := resolveEnvironmentToken(cmd.EnvironmentToken)
	if err != nil {
		return CreateShopResult{}, err
	}

	shopInfo, err := s.environments.ValidateConnection(ctx, ConnectionCredentials{
		URL:              cmd.EnvironmentURL,
		ClientID:         cmd.ClientID,
		ClientSecret:     cmd.ClientSecret,
		EnvironmentToken: environmentToken,
	})
	if err != nil {
		return CreateShopResult{}, fmt.Errorf("%w: %w", ErrConnectionFailed, err)
	}

	encryptedSecret, err := s.environments.EncryptSecret(cmd.ClientSecret)
	if err != nil {
		return CreateShopResult{}, fmt.Errorf("encrypt environment secret: %w", err)
	}

	shopID, environmentID, err := s.repository.CreateShopWithEnvironment(ctx, CreateShopRecord{
		OrganizationID: cmd.OrganizationID,
		Name:           cmd.Name,
		Description:    cmd.Description,
		GitURL:         cmd.GitURL,
	}, CreateEnvironmentRecord{
		OrganizationID:        cmd.OrganizationID,
		Name:                  cmd.EnvironmentName,
		URL:                   cmd.EnvironmentURL,
		ClientID:              cmd.ClientID,
		EncryptedClientSecret: encryptedSecret,
		ShopwareVersion:       shopInfo.Version,
		EnvironmentToken:      environmentToken,
	})
	if err != nil {
		return CreateShopResult{}, fmt.Errorf("create shop with environment: %w", err)
	}

	s.dispatchInitialScrape(ctx, environmentID)
	return CreateShopResult{ShopID: shopID, EnvironmentID: environmentID}, nil
}

type CreateEnvironmentCommand struct {
	UserID           string
	ShopID           int32
	Name             string
	URL              string
	ClientID         string
	ClientSecret     string
	EnvironmentToken string
}

func (s *Service) CreateEnvironment(ctx context.Context, cmd CreateEnvironmentCommand) (int32, error) {
	shop, err := s.repository.FindShop(ctx, cmd.ShopID)
	if err != nil {
		return 0, err
	}
	if err := s.authorize(ctx, cmd.UserID, shop.OrganizationID); err != nil {
		return 0, err
	}

	environmentToken, err := resolveEnvironmentToken(cmd.EnvironmentToken)
	if err != nil {
		return 0, err
	}

	shopInfo, err := s.environments.ValidateConnection(ctx, ConnectionCredentials{
		URL:              cmd.URL,
		ClientID:         cmd.ClientID,
		ClientSecret:     cmd.ClientSecret,
		EnvironmentToken: environmentToken,
	})
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrConnectionFailed, err)
	}
	encryptedSecret, err := s.environments.EncryptSecret(cmd.ClientSecret)
	if err != nil {
		return 0, fmt.Errorf("encrypt environment secret: %w", err)
	}

	environmentID, err := s.repository.CreateEnvironment(ctx, CreateEnvironmentRecord{
		OrganizationID:        shop.OrganizationID,
		ShopID:                cmd.ShopID,
		Name:                  cmd.Name,
		URL:                   cmd.URL,
		ClientID:              cmd.ClientID,
		EncryptedClientSecret: encryptedSecret,
		ShopwareVersion:       shopInfo.Version,
		EnvironmentToken:      environmentToken,
	})
	if err != nil {
		return 0, fmt.Errorf("create environment: %w", err)
	}

	s.dispatchInitialScrape(ctx, environmentID)
	return environmentID, nil
}

// resolveEnvironmentToken returns the provided token, or generates a 32-byte hex
// token when the client omits one (e.g. CreateShop or a misnamed request field).
func resolveEnvironmentToken(token string) (string, error) {
	if token != "" {
		return token, nil
	}
	generated, err := generateEnvironmentToken()
	if err != nil {
		return "", fmt.Errorf("generate environment token: %w", err)
	}
	return generated, nil
}

func generateEnvironmentToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type UpdateShopCommand struct {
	UserID               string
	OrganizationID       string
	ShopID               int32
	Name                 *string
	Description          *string
	GitURL               *string
	DefaultEnvironmentID *int32
}

func (s *Service) UpdateShop(ctx context.Context, cmd UpdateShopCommand) error {
	if err := s.authorize(ctx, cmd.UserID, cmd.OrganizationID); err != nil {
		return err
	}
	shop, err := s.repository.FindShop(ctx, cmd.ShopID)
	if err != nil {
		return err
	}
	if shop.OrganizationID != cmd.OrganizationID {
		return ErrShopOrganizationMismatch
	}
	if cmd.Name != nil {
		shop.Name = *cmd.Name
	}
	if cmd.Description != nil {
		shop.Description = cmd.Description
	}
	if cmd.GitURL != nil {
		shop.GitURL = cmd.GitURL
	}

	if cmd.DefaultEnvironmentID != nil {
		environment, err := s.repository.FindEnvironment(ctx, *cmd.DefaultEnvironmentID)
		if errors.Is(err, ErrEnvironmentNotFound) {
			return ErrDefaultEnvironmentNotFound
		}
		if err != nil {
			return fmt.Errorf("find default environment: %w", err)
		}
		if environment.ShopID != cmd.ShopID || environment.OrganizationID != cmd.OrganizationID {
			return ErrDefaultEnvironmentMismatch
		}
	}

	if err := s.repository.UpdateShop(ctx, shop, cmd.DefaultEnvironmentID); err != nil {
		return fmt.Errorf("update shop: %w", err)
	}
	return nil
}

func (s *Service) DeleteShop(ctx context.Context, userID, organizationID string, shopID int32) error {
	if err := s.authorize(ctx, userID, organizationID); err != nil {
		return err
	}
	shop, err := s.repository.FindShop(ctx, shopID)
	if err != nil {
		return err
	}
	if shop.OrganizationID != organizationID {
		return ErrShopOrganizationMismatch
	}

	count, err := s.repository.CountShopEnvironments(ctx, shopID)
	if err != nil {
		return fmt.Errorf("count shop environments: %w", err)
	}
	if count > 0 {
		return ErrShopHasEnvironments
	}
	if err := s.repository.DeleteShop(ctx, shopID, organizationID); err != nil {
		return fmt.Errorf("delete shop: %w", err)
	}
	return nil
}

type UpdateEnvironmentCommand struct {
	UserID        string
	EnvironmentID int32
	ShopID        int32
	Name          *string
	URL           *string
	ClientID      *string
	ClientSecret  *string
	Ignores       *[]string
}

func (s *Service) UpdateEnvironment(ctx context.Context, cmd UpdateEnvironmentCommand) error {
	environment, err := s.authorizedEnvironment(ctx, cmd.UserID, cmd.EnvironmentID)
	if err != nil {
		return err
	}
	if cmd.ShopID != environment.ShopID {
		shop, err := s.repository.FindShop(ctx, cmd.ShopID)
		if err != nil {
			return fmt.Errorf("find target shop: %w", err)
		}
		if shop.OrganizationID != environment.OrganizationID {
			return ErrShopOrganizationMismatch
		}
		environment.ShopID = cmd.ShopID
	}
	if cmd.Name != nil {
		environment.Name = *cmd.Name
	}
	if cmd.URL != nil {
		environment.URL = *cmd.URL
	}
	if cmd.ClientID != nil {
		environment.ClientID = *cmd.ClientID
	}
	if cmd.ClientSecret != nil {
		if _, err := s.environments.ValidateConnection(ctx, ConnectionCredentials{
			URL:              environment.URL,
			ClientID:         environment.ClientID,
			ClientSecret:     *cmd.ClientSecret,
			EnvironmentToken: environment.EnvironmentToken,
		}); err != nil {
			return fmt.Errorf("%w: %w", ErrConnectionFailed, err)
		}
		encryptedSecret, err := s.environments.EncryptSecret(*cmd.ClientSecret)
		if err != nil {
			return fmt.Errorf("encrypt environment secret: %w", err)
		}
		environment.EncryptedClientSecret = encryptedSecret
	}
	if cmd.Ignores != nil {
		environment.Ignores = append([]string(nil), (*cmd.Ignores)...)
	}

	if err := s.repository.UpdateEnvironment(ctx, environment); err != nil {
		return fmt.Errorf("update environment: %w", err)
	}
	return nil
}

func (s *Service) DeleteEnvironment(ctx context.Context, userID string, environmentID int32) error {
	if _, err := s.authorizedEnvironment(ctx, userID, environmentID); err != nil {
		return err
	}

	if s.outputs != nil {
		deploymentIDs, err := s.repository.ListDeploymentIDs(ctx, environmentID)
		if err != nil {
			slog.WarnContext(ctx, "failed to list deployment outputs for environment cleanup", "environmentID", environmentID, "error", err)
		} else {
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), deploymentOutputCleanupTimeout)
			defer cancel()
			for _, deploymentID := range deploymentIDs {
				if err := cleanupCtx.Err(); err != nil {
					slog.WarnContext(ctx, "stopped deleting deployment outputs", "environmentID", environmentID, "error", err)
					break
				}
				if err := s.outputs.DeleteDeploymentOutput(cleanupCtx, int(deploymentID)); err != nil {
					slog.WarnContext(ctx, "failed to delete deployment output", "deploymentID", deploymentID, "error", err)
				}
			}
		}
	}

	if err := s.repository.DeleteEnvironment(ctx, environmentID); err != nil {
		return fmt.Errorf("delete environment: %w", err)
	}
	return nil
}

func (s *Service) RefreshEnvironment(ctx context.Context, userID string, environmentID int32, sitespeed bool) error {
	if _, err := s.authorizedEnvironment(ctx, userID, environmentID); err != nil {
		return err
	}
	if err := s.jobs.DispatchEnvironmentScrape(ctx, environmentID); err != nil {
		return fmt.Errorf("dispatch environment scrape: %w", err)
	}
	if sitespeed {
		if err := s.jobs.DispatchSitespeedScrape(ctx, environmentID); err != nil {
			slog.WarnContext(ctx, "failed to enqueue sitespeed task", "environmentID", environmentID, "error", err)
		}
	}
	return nil
}

func (s *Service) ClearEnvironmentCache(ctx context.Context, userID string, environmentID int32) error {
	environment, err := s.authorizedEnvironment(ctx, userID, environmentID)
	if err != nil {
		return err
	}
	return s.environments.ClearCache(ctx, storedCredentials(environment))
}

func (s *Service) RescheduleTask(ctx context.Context, userID string, environmentID int32, taskID string) error {
	environment, err := s.authorizedEnvironment(ctx, userID, environmentID)
	if err != nil {
		return err
	}
	return s.environments.RescheduleTask(ctx, storedCredentials(environment), taskID, time.Now().UTC())
}

func (s *Service) UpdateSitespeedSettings(ctx context.Context, userID string, environmentID int32, enabled bool, urls []string) error {
	if _, err := s.authorizedEnvironment(ctx, userID, environmentID); err != nil {
		return err
	}
	if err := s.repository.UpdateSitespeedSettings(ctx, environmentID, enabled, urls); err != nil {
		return fmt.Errorf("update sitespeed settings: %w", err)
	}
	return nil
}

func (s *Service) SubscribeToEnvironment(ctx context.Context, userID string, environmentID int32, _ []string) error {
	if _, err := s.authorizedEnvironment(ctx, userID, environmentID); err != nil {
		return err
	}
	if err := s.repository.SubscribeEnvironment(ctx, userID, environmentID); err != nil {
		return fmt.Errorf("subscribe to environment: %w", err)
	}
	return nil
}

func (s *Service) UnsubscribeFromEnvironment(ctx context.Context, userID string, environmentID int32, _ []string) error {
	if _, err := s.authorizedEnvironment(ctx, userID, environmentID); err != nil {
		return err
	}
	if err := s.repository.UnsubscribeEnvironment(ctx, userID, environmentID); err != nil {
		return fmt.Errorf("unsubscribe from environment: %w", err)
	}
	return nil
}

func (s *Service) IsEnvironmentSubscribed(ctx context.Context, userID string, environmentID int32) (bool, error) {
	return s.repository.IsEnvironmentSubscribed(ctx, userID, environmentID)
}

func (s *Service) authorizedEnvironment(ctx context.Context, userID string, environmentID int32) (Environment, error) {
	environment, err := s.repository.FindEnvironment(ctx, environmentID)
	if err != nil {
		return Environment{}, err
	}
	if err := s.authorize(ctx, userID, environment.OrganizationID); err != nil {
		return Environment{}, err
	}
	return environment, nil
}

func (s *Service) authorize(ctx context.Context, userID, organizationID string) error {
	member, err := s.authorizer.IsMember(ctx, userID, organizationID)
	if err != nil {
		return fmt.Errorf("authorize organization membership: %w", err)
	}
	if !member {
		return ErrNotAuthorized
	}
	return nil
}

func (s *Service) dispatchInitialScrape(ctx context.Context, environmentID int32) {
	if err := s.jobs.DispatchEnvironmentScrape(ctx, environmentID); err != nil {
		slog.ErrorContext(ctx, "failed to enqueue initial scrape", "environmentID", environmentID, "error", err)
	}
}

func storedCredentials(environment Environment) StoredCredentials {
	return StoredCredentials{
		URL:                   environment.URL,
		ClientID:              environment.ClientID,
		EncryptedClientSecret: environment.EncryptedClientSecret,
		EnvironmentToken:      environment.EnvironmentToken,
	}
}
