package handler

import (
	"context"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/deployment"
)

// HashApiKeyToken is retained for compatibility with integrations that seeded
// API keys through handler tests. New code should use deployment.HashAPIKeyToken.
func HashApiKeyToken(token string) string {
	return deployment.HashAPIKeyToken(token)
}

type getApiKeyScopesOutput struct {
	Body []api.ApiKeyScope
}

type getApiKeysInput struct {
	OrgID  api.OrgId  `path:"orgId"`
	ShopID api.ShopId `path:"shopId"`
}

type getApiKeysOutput struct {
	Body []api.ApiKey
}

type createApiKeyInput struct {
	OrgID  api.OrgId  `path:"orgId"`
	ShopID api.ShopId `path:"shopId"`
	Body   api.CreateApiKeyRequest
}

type createApiKeyOutput struct {
	Status int
	Body   api.CreateApiKeyResponse
}

type deleteApiKeyInput struct {
	OrgID  api.OrgId  `path:"orgId"`
	ShopID api.ShopId `path:"shopId"`
	KeyID  api.KeyId  `path:"keyId"`
}

type deleteApiKeyOutput struct {
	Status int
}

// GetApiKeyScopes returns available API key scopes.
func (h *Handler) GetApiKeyScopes(_ context.Context, _ *struct{}) (*getApiKeyScopesOutput, error) {
	scopes := deployment.AvailableScopes()
	result := make([]api.ApiKeyScope, 0, len(scopes))
	for _, scope := range scopes {
		result = append(result, api.ApiKeyScope{
			Value:       scope.Value,
			Label:       scope.Label,
			Description: scope.Description,
		})
	}
	return &getApiKeyScopesOutput{Body: result}, nil
}

// GetApiKeys lists API keys for a shop.
func (h *Handler) GetApiKeys(ctx context.Context, input *getApiKeysInput) (*getApiKeysOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	keys, err := h.deployments.ListAPIKeys(ctx, deployment.ShopCommand{
		UserID:         user.ID,
		OrganizationID: input.OrgID,
		ShopID:         int32(input.ShopID),
	})
	if err != nil {
		return nil, h.writeDeploymentError(ctx, "list api keys", err)
	}

	result := make([]api.ApiKey, 0, len(keys))
	for _, key := range keys {
		result = append(result, api.ApiKey{
			Id:         key.ID,
			Name:       key.Name,
			Scopes:     key.Scopes,
			CreatedAt:  key.CreatedAt,
			LastUsedAt: key.LastUsedAt,
		})
	}
	return &getApiKeysOutput{Body: result}, nil
}

// CreateApiKey creates a new API key for a shop.
func (h *Handler) CreateApiKey(ctx context.Context, input *createApiKeyInput) (*createApiKeyOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	key, err := h.deployments.CreateAPIKey(ctx, deployment.CreateAPIKeyCommand{
		ShopCommand: deployment.ShopCommand{
			UserID:         user.ID,
			OrganizationID: input.OrgID,
			ShopID:         int32(input.ShopID),
		},
		Name:   input.Body.Name,
		Scopes: input.Body.Scopes,
	})
	if err != nil {
		return nil, h.writeDeploymentError(ctx, "create api key", err)
	}
	return &createApiKeyOutput{
		Status: http.StatusCreated,
		Body: api.CreateApiKeyResponse{
			Id:     key.ID,
			Name:   key.Name,
			Scopes: key.Scopes,
			Token:  key.Token,
		},
	}, nil
}

// DeleteApiKey deletes an API key.
func (h *Handler) DeleteApiKey(ctx context.Context, input *deleteApiKeyInput) (*deleteApiKeyOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.deployments.DeleteAPIKey(ctx, deployment.ShopCommand{
		UserID:         user.ID,
		OrganizationID: input.OrgID,
		ShopID:         int32(input.ShopID),
	}, input.KeyID); err != nil {
		return nil, h.writeDeploymentError(ctx, "delete api key", err)
	}
	return &deleteApiKeyOutput{Status: http.StatusNoContent}, nil
}
