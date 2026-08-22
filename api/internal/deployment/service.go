package deployment

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotAuthorized            = errors.New("not authorized for organization")
	ErrShopNotFound             = errors.New("shop not found")
	ErrShopOrganizationMismatch = errors.New("shop does not belong to organization")
	ErrEnvironmentNotFound      = errors.New("environment not found")
	ErrDeploymentNotFound       = errors.New("deployment not found")
	ErrAPIKeyInvalid            = errors.New("invalid api key")
	ErrDeploymentScopeRequired  = errors.New("api key does not have deployments scope")
	ErrEnvironmentShopMismatch  = errors.New("environment does not belong to api key shop")
	ErrAPIKeyNameRequired       = errors.New("api key name is required")
	ErrEnvironmentIDRequired    = errors.New("environment id is required")
	ErrCommandRequired          = errors.New("deployment command is required")
	ErrOutputNotFound           = errors.New("deployment output not found")
	ErrOutputUnavailable        = errors.New("deployment output unavailable")
)

const (
	ScopeDeployments               = "deployments"
	deploymentOutputCleanupTimeout = 30 * time.Second
)

type Scope struct {
	Value       string
	Label       string
	Description string
}

func AvailableScopes() []Scope {
	return []Scope{{
		Value:       ScopeDeployments,
		Label:       "Deployments",
		Description: "Create and manage deployments",
	}}
}

type Environment struct {
	ID             int32
	OrganizationID string
	ShopID         int32
}

type Deployment struct {
	ID            int32
	EnvironmentID int32
	Name          string
	Command       string
	ReturnCode    int32
	StartDate     time.Time
	EndDate       time.Time
	ExecutionTime float32
	Composer      []byte
	Reference     *string
	CreatedAt     time.Time
	Output        *string
}

type APIKey struct {
	ID             string
	ShopID         int32
	OrganizationID string
	Name           string
	TokenHash      string
	Scopes         []string
	CreatedAt      time.Time
	LastUsedAt     *time.Time
}

type CreateDeploymentRecord struct {
	EnvironmentID int32
	Name          string
	Command       string
	ReturnCode    int32
	StartDate     time.Time
	EndDate       time.Time
	ExecutionTime float32
	Composer      []byte
	Reference     *string
}

type Repository interface {
	FindEnvironment(ctx context.Context, environmentID int32) (Environment, error)
	OrganizationIDForShop(ctx context.Context, shopID int32) (string, error)
	ListDeployments(ctx context.Context, environmentID, limit, offset int32) ([]Deployment, error)
	FindDeployment(ctx context.Context, environmentID, deploymentID int32) (Deployment, error)
	DeleteDeployment(ctx context.Context, environmentID, deploymentID int32) error
	CreateDeployment(ctx context.Context, deployment CreateDeploymentRecord) (int32, error)
	ListAPIKeys(ctx context.Context, shopID int32) ([]APIKey, error)
	CreateAPIKey(ctx context.Context, key APIKey) error
	DeleteAPIKey(ctx context.Context, shopID int32, keyID string) error
	FindAPIKeyByTokenHash(ctx context.Context, tokenHash string) (APIKey, error)
	TouchAPIKey(ctx context.Context, keyID string) error
}

type MembershipAuthorizer interface {
	IsMember(ctx context.Context, userID, organizationID string) (bool, error)
}

type OutputStore interface {
	GetDeploymentOutput(ctx context.Context, deploymentID int) (string, error)
	DeleteDeploymentOutput(ctx context.Context, deploymentID int) error
	PresignUpload(ctx context.Context, deploymentID int) (string, error)
}

type PostDeploymentDispatcher interface {
	DispatchPostDeploymentScrape(ctx context.Context, environmentID int32, delay time.Duration) error
}

type Service struct {
	repository  Repository
	authorizer  MembershipAuthorizer
	outputs     OutputStore
	dispatcher  PostDeploymentDispatcher
	frontendURL string
	scrapeDelay time.Duration
}

func NewService(repository Repository, authorizer MembershipAuthorizer, outputs OutputStore, dispatcher PostDeploymentDispatcher, frontendURL string, scrapeDelay time.Duration) *Service {
	return &Service{
		repository:  repository,
		authorizer:  authorizer,
		outputs:     outputs,
		dispatcher:  dispatcher,
		frontendURL: strings.TrimRight(frontendURL, "/"),
		scrapeDelay: scrapeDelay,
	}
}

func (s *Service) List(ctx context.Context, userID string, environmentID, limit, offset int32) ([]Deployment, error) {
	if _, err := s.authorizedEnvironment(ctx, userID, environmentID); err != nil {
		return nil, err
	}
	deployments, err := s.repository.ListDeployments(ctx, environmentID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list deployments: %w", err)
	}
	return deployments, nil
}

func (s *Service) Get(ctx context.Context, userID string, environmentID, deploymentID int32) (Deployment, error) {
	if _, err := s.authorizedEnvironment(ctx, userID, environmentID); err != nil {
		return Deployment{}, err
	}
	result, err := s.repository.FindDeployment(ctx, environmentID, deploymentID)
	if err != nil {
		return Deployment{}, err
	}
	if s.outputs == nil {
		return result, nil
	}
	output, err := s.outputs.GetDeploymentOutput(ctx, int(deploymentID))
	if errors.Is(err, ErrOutputNotFound) {
		return result, nil
	}
	if err != nil {
		return Deployment{}, fmt.Errorf("%w: get deployment %d output: %w", ErrOutputUnavailable, deploymentID, err)
	}
	result.Output = &output
	return result, nil
}

func (s *Service) Delete(ctx context.Context, userID string, environmentID, deploymentID int32) error {
	if _, err := s.authorizedEnvironment(ctx, userID, environmentID); err != nil {
		return err
	}
	if err := s.repository.DeleteDeployment(ctx, environmentID, deploymentID); err != nil {
		return fmt.Errorf("delete deployment: %w", err)
	}
	if s.outputs != nil {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), deploymentOutputCleanupTimeout)
		defer cancel()
		if err := s.outputs.DeleteDeploymentOutput(cleanupCtx, int(deploymentID)); err != nil {
			slog.WarnContext(cleanupCtx, "failed to delete deployment output", "deploymentId", deploymentID, "error", err)
		}
	}
	return nil
}

type ShopCommand struct {
	UserID         string
	OrganizationID string
	ShopID         int32
}

func (s *Service) ListAPIKeys(ctx context.Context, cmd ShopCommand) ([]APIKey, error) {
	if err := s.authorizeShop(ctx, cmd); err != nil {
		return nil, err
	}
	keys, err := s.repository.ListAPIKeys(ctx, cmd.ShopID)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	return keys, nil
}

type CreateAPIKeyCommand struct {
	ShopCommand
	Name   string
	Scopes []string
}

type CreatedAPIKey struct {
	ID     string
	Name   string
	Scopes []string
	Token  string
}

func (s *Service) CreateAPIKey(ctx context.Context, cmd CreateAPIKeyCommand) (CreatedAPIKey, error) {
	if err := s.authorizeShop(ctx, cmd.ShopCommand); err != nil {
		return CreatedAPIKey{}, err
	}
	if strings.TrimSpace(cmd.Name) == "" {
		return CreatedAPIKey{}, ErrAPIKeyNameRequired
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return CreatedAPIKey{}, fmt.Errorf("generate api key token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)
	key := APIKey{
		ID:        uuid.NewString(),
		ShopID:    cmd.ShopID,
		Name:      cmd.Name,
		TokenHash: HashAPIKeyToken(token),
		Scopes:    append([]string(nil), cmd.Scopes...),
	}
	if err := s.repository.CreateAPIKey(ctx, key); err != nil {
		return CreatedAPIKey{}, fmt.Errorf("create api key: %w", err)
	}
	return CreatedAPIKey{ID: key.ID, Name: key.Name, Scopes: key.Scopes, Token: token}, nil
}

func (s *Service) DeleteAPIKey(ctx context.Context, cmd ShopCommand, keyID string) error {
	if err := s.authorizeShop(ctx, cmd); err != nil {
		return err
	}
	if err := s.repository.DeleteAPIKey(ctx, cmd.ShopID, keyID); err != nil {
		return fmt.Errorf("delete api key: %w", err)
	}
	return nil
}

type CreateCLICommand struct {
	Token         string
	EnvironmentID int32
	Name          *string
	Command       string
	ReturnCode    int32
	StartDate     time.Time
	EndDate       time.Time
	ExecutionTime float32
	Composer      *string
	Reference     *string
}

type CreateCLIResult struct {
	DeploymentID int32
	Name         string
	UploadURL    string
	URL          string
}

func (s *Service) CreateCLI(ctx context.Context, cmd CreateCLICommand) (CreateCLIResult, error) {
	key, err := s.repository.FindAPIKeyByTokenHash(ctx, HashAPIKeyToken(cmd.Token))
	if err != nil {
		return CreateCLIResult{}, err
	}
	if err := s.repository.TouchAPIKey(ctx, key.ID); err != nil {
		slog.WarnContext(ctx, "failed to update api key last-used time", "apiKeyId", key.ID, "error", err)
	}
	if !containsScope(key.Scopes, ScopeDeployments) {
		return CreateCLIResult{}, ErrDeploymentScopeRequired
	}
	if cmd.EnvironmentID <= 0 {
		return CreateCLIResult{}, ErrEnvironmentIDRequired
	}
	if strings.TrimSpace(cmd.Command) == "" {
		return CreateCLIResult{}, ErrCommandRequired
	}
	environment, err := s.repository.FindEnvironment(ctx, cmd.EnvironmentID)
	if err != nil {
		return CreateCLIResult{}, err
	}
	if environment.ShopID != key.ShopID {
		return CreateCLIResult{}, ErrEnvironmentShopMismatch
	}

	name := fmt.Sprintf("CLI Deployment #%d", cmd.EnvironmentID)
	if cmd.Name != nil {
		name = *cmd.Name
	}
	var composer []byte
	if cmd.Composer != nil {
		composer = []byte(*cmd.Composer)
	}
	deploymentID, err := s.repository.CreateDeployment(ctx, CreateDeploymentRecord{
		EnvironmentID: cmd.EnvironmentID,
		Name:          name,
		Command:       cmd.Command,
		ReturnCode:    cmd.ReturnCode,
		StartDate:     cmd.StartDate,
		EndDate:       cmd.EndDate,
		ExecutionTime: cmd.ExecutionTime,
		Composer:      composer,
		Reference:     cmd.Reference,
	})
	if err != nil {
		return CreateCLIResult{}, fmt.Errorf("create deployment: %w", err)
	}

	if s.dispatcher != nil {
		if err := s.dispatcher.DispatchPostDeploymentScrape(ctx, cmd.EnvironmentID, s.scrapeDelay); err != nil {
			slog.ErrorContext(ctx, "failed to enqueue post-deployment scrape", "environmentId", cmd.EnvironmentID, "error", err)
		}
	}

	var uploadURL string
	if s.outputs != nil {
		uploadURL, err = s.outputs.PresignUpload(ctx, int(deploymentID))
		if err != nil {
			return CreateCLIResult{}, fmt.Errorf("%w: create upload URL: %w", ErrOutputUnavailable, err)
		}
	}
	return CreateCLIResult{
		DeploymentID: deploymentID,
		Name:         name,
		UploadURL:    uploadURL,
		URL:          fmt.Sprintf("%s/environments/%d/deployments/%d", s.frontendURL, cmd.EnvironmentID, deploymentID),
	}, nil
}

func (s *Service) authorizedEnvironment(ctx context.Context, userID string, environmentID int32) (Environment, error) {
	environment, err := s.repository.FindEnvironment(ctx, environmentID)
	if err != nil {
		return Environment{}, err
	}
	member, err := s.authorizer.IsMember(ctx, userID, environment.OrganizationID)
	if err != nil {
		return Environment{}, fmt.Errorf("authorize environment organization: %w", err)
	}
	if !member {
		return Environment{}, ErrNotAuthorized
	}
	return environment, nil
}

func (s *Service) authorizeShop(ctx context.Context, cmd ShopCommand) error {
	member, err := s.authorizer.IsMember(ctx, cmd.UserID, cmd.OrganizationID)
	if err != nil {
		return fmt.Errorf("authorize organization membership: %w", err)
	}
	if !member {
		return ErrNotAuthorized
	}
	organizationID, err := s.repository.OrganizationIDForShop(ctx, cmd.ShopID)
	if err != nil {
		return err
	}
	if organizationID != cmd.OrganizationID {
		return ErrShopOrganizationMismatch
	}
	return nil
}

func containsScope(scopes []string, expected string) bool {
	return slices.Contains(scopes, expected)
}

// HashAPIKeyToken returns the SHA-256 hex digest stored for a plaintext API key.
func HashAPIKeyToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
