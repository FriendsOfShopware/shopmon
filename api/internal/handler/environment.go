package handler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/monitoring"
	environmentread "github.com/friendsofshopware/shopmon/api/internal/readmodel/environment"
)

// GetOrganizationEnvironments returns all environments in an organization.
func (h *Handler) GetOrganizationEnvironments(w http.ResponseWriter, r *http.Request, orgID api.OrgId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	result, err := h.environments.OrganizationEnvironments(r.Context(), user.ID, orgID)
	if err != nil {
		h.writeEnvironmentReadError(w, r, "list organization environments", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetEnvironment returns full environment details.
func (h *Handler) GetEnvironment(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId, params api.GetEnvironmentParams) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var langStr *string
	if params.Language != nil {
		s := string(*params.Language)
		langStr = &s
	}
	subscribed, err := h.monitoring.IsEnvironmentSubscribed(r.Context(), user.ID, int32(environmentId))
	if err != nil {
		slog.Error("failed to check environment subscription", "environmentId", environmentId, "error", err)
		subscribed = false
	}
	detail, err := h.environments.Detail(r.Context(), user.ID, int32(environmentId), langStr, subscribed)
	if err != nil {
		h.writeEnvironmentReadError(w, r, "get environment detail", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, detail)
}

func (h *Handler) writeEnvironmentReadError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	switch {
	case errors.Is(err, environmentread.ErrNotFound):
		httputil.WriteError(w, http.StatusNotFound, "environment not found")
	case errors.Is(err, environmentread.ErrNotAuthorized):
		httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
	default:
		httputil.WriteInternalError(w, r, err, "failed to load environment")
	}
}

// CreateEnvironment creates a new environment.
func (h *Handler) CreateEnvironment(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var req api.CreateEnvironmentRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.ShopUrl == "" || req.ClientId == "" || req.ClientSecret == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name, shopUrl, clientId, and clientSecret are required")
		return
	}

	environmentToken := ""
	if req.EnvironmentToken != nil {
		environmentToken = *req.EnvironmentToken
	}

	environmentID, err := h.monitoring.CreateEnvironment(r.Context(), monitoring.CreateEnvironmentCommand{
		UserID:           user.ID,
		Name:             req.Name,
		URL:              req.ShopUrl,
		ClientID:         req.ClientId,
		ClientSecret:     req.ClientSecret,
		ShopID:           int32(req.ShopId),
		EnvironmentToken: environmentToken,
	})
	switch {
	case errors.Is(err, monitoring.ErrShopNotFound):
		httputil.WriteError(w, http.StatusBadRequest, "shop not found")
		return
	case errors.Is(err, monitoring.ErrNotAuthorized):
		httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
		return
	case errors.Is(err, monitoring.ErrConnectionFailed):
		slog.Error("failed to create environment", "error", err)
		httputil.WriteError(w, http.StatusBadRequest, "Cannot reach shop. Check your credentials and shop URL.")
		return
	case err != nil:
		slog.Error("failed to create environment", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create environment")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]interface{}{"id": environmentID})
}

// UpdateEnvironment updates an existing environment.
func (h *Handler) UpdateEnvironment(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var req api.UpdateEnvironmentRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.monitoring.UpdateEnvironment(r.Context(), monitoring.UpdateEnvironmentCommand{
		UserID:        user.ID,
		EnvironmentID: int32(environmentId),
		ShopID:        int32(req.ShopId),
		Name:          req.Name,
		URL:           req.ShopUrl,
		ClientID:      req.ClientId,
		ClientSecret:  req.ClientSecret,
		Ignores:       req.Ignores,
	})
	switch {
	case errors.Is(err, monitoring.ErrEnvironmentNotFound):
		httputil.WriteError(w, http.StatusNotFound, "environment not found")
		return
	case errors.Is(err, monitoring.ErrNotAuthorized):
		httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
		return
	case errors.Is(err, monitoring.ErrShopNotFound):
		httputil.WriteError(w, http.StatusBadRequest, "target shop not found")
		return
	case errors.Is(err, monitoring.ErrShopOrganizationMismatch):
		httputil.WriteError(w, http.StatusForbidden, "target shop belongs to a different organization")
		return
	case errors.Is(err, monitoring.ErrConnectionFailed):
		slog.Error("failed to build environment update", "error", err)
		httputil.WriteError(w, http.StatusBadRequest, "Cannot reach shop with new credentials. Check your credentials and shop URL.")
		return
	case err != nil:
		slog.Error("failed to update environment", "environmentID", environmentId, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update environment")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteEnvironment deletes an environment.
func (h *Handler) DeleteEnvironment(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	err := h.monitoring.DeleteEnvironment(r.Context(), user.ID, int32(environmentId))
	switch {
	case errors.Is(err, monitoring.ErrEnvironmentNotFound):
		httputil.WriteError(w, http.StatusNotFound, "environment not found")
		return
	case errors.Is(err, monitoring.ErrNotAuthorized):
		httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
		return
	case err != nil:
		slog.Error("failed to delete environment", "environmentID", environmentId, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete environment")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RefreshEnvironment triggers a re-scrape of the environment.
func (h *Handler) RefreshEnvironment(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var body api.RefreshEnvironmentJSONRequestBody
	if err := httputil.DecodeBody(r, &body); err != nil && err != io.EOF {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sitespeed := body.Sitespeed != nil && *body.Sitespeed
	err := h.monitoring.RefreshEnvironment(r.Context(), user.ID, int32(environmentId), sitespeed)
	switch {
	case errors.Is(err, monitoring.ErrEnvironmentNotFound):
		httputil.WriteError(w, http.StatusNotFound, "environment not found")
		return
	case errors.Is(err, monitoring.ErrNotAuthorized):
		httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
		return
	case err != nil:
		slog.Error("failed to enqueue scrape task", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to enqueue refresh task")
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// ClearEnvironmentCache clears the Shopware cache for an environment.
func (h *Handler) ClearEnvironmentCache(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	err := h.monitoring.ClearEnvironmentCache(r.Context(), user.ID, int32(environmentId))
	switch {
	case errors.Is(err, monitoring.ErrEnvironmentNotFound):
		httputil.WriteError(w, http.StatusNotFound, "environment not found")
		return
	case errors.Is(err, monitoring.ErrNotAuthorized):
		httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
		return
	case errors.Is(err, monitoring.ErrCredentialDecryption):
		slog.Error("failed to decrypt client secret", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to clear cache")
		return
	case errors.Is(err, monitoring.ErrRemoteOperation):
		slog.Error("failed to clear environment cache", "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to clear cache on environment")
		return
	case err != nil:
		slog.Error("failed to clear environment cache", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to clear cache")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RescheduleTask reschedules a scheduled task on the environment.
func (h *Handler) RescheduleTask(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId, taskId api.TaskId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	err := h.monitoring.RescheduleTask(r.Context(), user.ID, int32(environmentId), string(taskId))
	switch {
	case errors.Is(err, monitoring.ErrEnvironmentNotFound):
		httputil.WriteError(w, http.StatusNotFound, "environment not found")
		return
	case errors.Is(err, monitoring.ErrNotAuthorized):
		httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
		return
	case errors.Is(err, monitoring.ErrCredentialDecryption):
		slog.Error("failed to decrypt client secret", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to reschedule task")
		return
	case errors.Is(err, monitoring.ErrRemoteOperation):
		slog.Error("failed to reschedule task", "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to reschedule task on environment")
		return
	case err != nil:
		slog.Error("failed to reschedule task", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to reschedule task")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SubscribeToEnvironment subscribes the user to environment notifications.
func (h *Handler) SubscribeToEnvironment(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	err := h.monitoring.SubscribeToEnvironment(r.Context(), user.ID, int32(environmentId), user.Notifications)
	if h.writeEnvironmentSubscriptionError(w, r, "subscribe", err) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UnsubscribeFromEnvironment unsubscribes the user from environment notifications.
func (h *Handler) UnsubscribeFromEnvironment(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	err := h.monitoring.UnsubscribeFromEnvironment(r.Context(), user.ID, int32(environmentId), user.Notifications)
	if h.writeEnvironmentSubscriptionError(w, r, "unsubscribe", err) {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeEnvironmentSubscriptionError(w http.ResponseWriter, r *http.Request, operation string, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, monitoring.ErrEnvironmentNotFound):
		httputil.WriteError(w, http.StatusNotFound, "environment not found")
	case errors.Is(err, monitoring.ErrNotAuthorized):
		httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
	default:
		slog.ErrorContext(r.Context(), "environment subscription failed", "operation", operation, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update environment subscription")
	}
	return true
}

// UpdateSitespeedSettings updates sitespeed settings for an environment.
func (h *Handler) UpdateSitespeedSettings(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var req api.SitespeedSettingsRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var urls []string
	if req.Urls != nil {
		urls = append(urls, (*req.Urls)...)
	}
	err := h.monitoring.UpdateSitespeedSettings(r.Context(), user.ID, int32(environmentId), req.Enabled, urls)
	switch {
	case errors.Is(err, monitoring.ErrEnvironmentNotFound):
		httputil.WriteError(w, http.StatusNotFound, "environment not found")
		return
	case errors.Is(err, monitoring.ErrNotAuthorized):
		httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
		return
	case err != nil:
		slog.Error("failed to update sitespeed settings", "environmentID", environmentId, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update sitespeed settings")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetEnvironmentStatusEvents returns the status change history for an environment.
func (h *Handler) GetEnvironmentStatusEvents(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	events, err := h.environments.StatusEvents(r.Context(), user.ID, int32(environmentId))
	if err != nil {
		h.writeEnvironmentReadError(w, r, "list environment status events", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, events)
}
