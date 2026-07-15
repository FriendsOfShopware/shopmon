package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/packagesmirror"
)

// GetPackagesTokenConfiguration returns the packages token configuration.
func (h *Handler) GetPackagesTokenConfiguration(w http.ResponseWriter, _ *http.Request) {
	configuration := h.packages.Configuration()
	httputil.WriteJSON(w, http.StatusOK, api.PackagesTokenConfiguration{
		Configured:  configuration.Configured,
		ComposerUrl: configuration.ComposerURL,
	})
}

// GetPackagesTokens lists packages tokens for a shop.
func (h *Handler) GetPackagesTokens(w http.ResponseWriter, r *http.Request, orgID api.OrgId, shopID api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	tokens, err := h.packages.List(r.Context(), packagesmirror.ShopCommand{
		UserID:         user.ID,
		OrganizationID: orgID,
		ShopID:         int32(shopID),
	})
	if err != nil {
		h.writePackagesError(w, r, "list packages tokens", err)
		return
	}

	response := make([]api.PackagesToken, 0, len(tokens))
	for _, token := range tokens {
		response = append(response, api.PackagesToken{
			Id:           token.ID,
			Source:       token.Source,
			LastSyncedAt: token.LastSyncedAt,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, response)
}

// CreatePackagesToken creates a new packages token.
func (h *Handler) CreatePackagesToken(w http.ResponseWriter, r *http.Request, orgID api.OrgId, shopID api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var request api.CreatePackagesTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "token is required")
		return
	}
	if err := h.packages.Create(r.Context(), packagesmirror.CreateCommand{
		ShopCommand: packagesmirror.ShopCommand{
			UserID:         user.ID,
			OrganizationID: orgID,
			ShopID:         int32(shopID),
		},
		Token: request.Token,
	}); err != nil {
		h.writePackagesError(w, r, "create packages token", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// DeletePackagesToken deletes a packages token.
func (h *Handler) DeletePackagesToken(w http.ResponseWriter, r *http.Request, orgID api.OrgId, shopID api.ShopId, tokenID api.TokenId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if err := h.packages.Delete(r.Context(), packagesmirror.TokenCommand{
		ShopCommand: packagesmirror.ShopCommand{
			UserID:         user.ID,
			OrganizationID: orgID,
			ShopID:         int32(shopID),
		},
		TokenID: int(tokenID),
	}); err != nil {
		h.writePackagesError(w, r, "delete packages token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SyncPackagesToken syncs a packages token.
func (h *Handler) SyncPackagesToken(w http.ResponseWriter, r *http.Request, orgID api.OrgId, shopID api.ShopId, tokenID api.TokenId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if err := h.packages.Sync(r.Context(), packagesmirror.TokenCommand{
		ShopCommand: packagesmirror.ShopCommand{
			UserID:         user.ID,
			OrganizationID: orgID,
			ShopID:         int32(shopID),
		},
		TokenID: int(tokenID),
	}); err != nil {
		h.writePackagesError(w, r, "sync packages token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writePackagesError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	switch {
	case errors.Is(err, packagesmirror.ErrNotConfigured):
		httputil.WriteError(w, http.StatusNotFound, "packages API not configured")
	case errors.Is(err, packagesmirror.ErrNotAuthorized):
		httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
	case errors.Is(err, packagesmirror.ErrShopNotFound):
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
	case errors.Is(err, packagesmirror.ErrShopOrganizationMismatch):
		httputil.WriteError(w, http.StatusForbidden, "shop does not belong to this organization")
	case errors.Is(err, packagesmirror.ErrTokenNotFound):
		httputil.WriteError(w, http.StatusNotFound, "token not found")
	case errors.Is(err, packagesmirror.ErrTokenRequired):
		httputil.WriteError(w, http.StatusBadRequest, "token is required")
	default:
		var remoteError *packagesmirror.RemoteError
		if errors.As(err, &remoteError) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(remoteError.StatusCode)
			_, _ = w.Write(remoteError.Body)
			return
		}
		slog.ErrorContext(r.Context(), "packages operation failed", "operation", operation, "error", err)
		if errors.Is(err, packagesmirror.ErrRemote) {
			httputil.WriteError(w, http.StatusBadGateway, "packages API request failed")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "packages operation failed")
	}
}
