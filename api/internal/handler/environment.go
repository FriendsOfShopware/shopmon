package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/jobs"

	goqueue "github.com/shyim/go-queue"
)

// GetOrganizationEnvironments returns all environments in the active organization.
func (h *Handler) GetOrganizationEnvironments(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	orgId := h.requireActiveOrganization(w, r)
	if orgId == "" {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	rows, err := h.queries.ListEnvironmentsByOrganization(r.Context(), orgId)
	if err != nil {
		slog.Error("failed to list environments by organization", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get environments")
		return
	}

	result := make([]api.AccountEnvironment, 0, len(rows))
	for _, row := range rows {
		var shopID *int
		if row.ShopID > 0 {
			v := int(row.ShopID)
			shopID = &v
		}
		result = append(result, api.AccountEnvironment{
			Id:               int(row.ID),
			Name:             row.Name,
			Url:              row.Url,
			Favicon:          row.Favicon,
			Status:           row.Status,
			ShopwareVersion:  row.ShopwareVersion,
			LastScrapedAt:    pgtimeToTimePtr(row.LastScrapedAt),
			LastScrapedError: row.LastScrapedError,
			OrganizationId:   row.OrganizationID,
			OrganizationName: row.OrganizationName,
			ShopId:           shopID,
			ShopName:         row.ShopName,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetEnvironment returns full environment details.
func (h *Handler) GetEnvironment(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	environment, ok := h.loadAuthorizedEnvironment(w, r, user, int32(environmentId))
	if !ok {
		return
	}

	aggregate := h.loadEnvironmentDetailAggregate(r.Context(), int32(environmentId))
	httputil.WriteJSON(w, http.StatusOK, h.buildEnvironmentDetail(environment, aggregate))
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

	shopOrgID, err := h.queries.GetShopOrganizationID(r.Context(), int32(req.ShopId))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "shop not found")
		return
	}

	if !h.requireOrgMembership(w, r, user, shopOrgID) {
		return
	}

	environmentToken := ""
	if req.EnvironmentToken != nil {
		environmentToken = *req.EnvironmentToken
	}

	environmentID, err := h.runCreateEnvironment(r.Context(), shopOrgID, createEnvironmentCommand{
		Name:             req.Name,
		ShopURL:          req.ShopUrl,
		ClientID:         req.ClientId,
		ClientSecret:     req.ClientSecret,
		ShopID:           int32(req.ShopId),
		EnvironmentToken: environmentToken,
	})
	if err != nil {
		slog.Error("failed to create environment", "error", err)
		httputil.WriteError(w, http.StatusBadRequest, "Cannot reach shop. Check your credentials and shop URL.")
		return
	}

	// Enqueue scrape task
	if err := goqueue.Dispatch(r.Context(), h.bus, jobs.EnvironmentScrape{EnvironmentID: environmentID}); err != nil {
		slog.Error("failed to enqueue scrape task", "error", err)
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]interface{}{"id": environmentID})
}

// UpdateEnvironment updates an existing environment.
func (h *Handler) UpdateEnvironment(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	environment, ok := h.loadAuthorizedEnvironment(w, r, user, int32(environmentId))
	if !ok {
		return
	}

	var req api.UpdateEnvironmentRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd, err := h.buildUpdateEnvironmentCommand(r.Context(), environment, req)
	if err != nil {
		slog.Error("failed to build environment update", "error", err)
		httputil.WriteError(w, http.StatusBadRequest, "Cannot reach shop with new credentials. Check your credentials and shop URL.")
		return
	}

	if err := h.runUpdateEnvironment(r.Context(), int32(environmentId), cmd); err != nil {
		slog.Error("failed to update environment", "error", err)
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

	if _, ok := h.loadAuthorizedEnvironment(w, r, user, int32(environmentId)); !ok {
		return
	}

	// Clean up deployment outputs from S3
	if h.storage != nil {
		deployments, err := h.queries.ListDeployments(r.Context(), queries.ListDeploymentsParams{
			EnvironmentID: int32(environmentId),
			Limit:         1000,
			Offset:        0,
		})
		if err == nil {
			for _, d := range deployments {
				if err := h.storage.DeleteDeploymentOutput(r.Context(), int(d.ID)); err != nil {
					slog.Warn("failed to delete deployment output", "deploymentId", d.ID, "error", err)
				}
			}
		}
	}

	if err := h.queries.DeleteEnvironment(r.Context(), int32(environmentId)); err != nil {
		slog.Error("failed to delete environment", "error", err)
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

	if _, ok := h.loadAuthorizedEnvironment(w, r, user, int32(environmentId)); !ok {
		return
	}

	var body api.RefreshEnvironmentJSONRequestBody
	if err := httputil.DecodeBody(r, &body); err != nil && err != io.EOF {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := goqueue.Dispatch(r.Context(), h.bus, jobs.EnvironmentScrape{EnvironmentID: int32(environmentId)}); err != nil {
		slog.Error("failed to enqueue scrape task", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to enqueue refresh task")
		return
	}

	if body.Sitespeed != nil && *body.Sitespeed {
		if err := goqueue.Dispatch(r.Context(), h.bus, jobs.SitespeedScrape{EnvironmentID: int32(environmentId)}); err != nil {
			slog.Warn("failed to enqueue sitespeed task", "error", err)
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

// ClearEnvironmentCache clears the Shopware cache for an environment.
func (h *Handler) ClearEnvironmentCache(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	creds, ok := h.loadAuthorizedEnvironmentCredentials(w, r, user, int32(environmentId))
	if !ok {
		return
	}

	client, err := h.newShopwareClientFromCredentials(creds)
	if err != nil {
		slog.Error("failed to decrypt client secret", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to clear cache")
		return
	}

	if _, err := client.Delete(r.Context(), "/_action/cache", nil); err != nil {
		slog.Error("failed to clear environment cache", "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to clear cache on environment")
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

	creds, ok := h.loadAuthorizedEnvironmentCredentials(w, r, user, int32(environmentId))
	if !ok {
		return
	}

	client, err := h.newShopwareClientFromCredentials(creds)
	if err != nil {
		slog.Error("failed to decrypt client secret", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to reschedule task")
		return
	}

	body := map[string]interface{}{
		"status":            "scheduled",
		"nextExecutionTime": time.Now().UTC().Format("2006-01-02T15:04:05+00:00"),
	}
	if _, err := client.Patch(r.Context(), "/scheduled-task/"+taskId, body); err != nil {
		slog.Error("failed to reschedule task", "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to reschedule task on environment")
		return
	}

	// Also trigger the scheduled task runner
	if _, err := client.Post(r.Context(), "/_action/scheduled-task/run", nil); err != nil {
		slog.Warn("failed to trigger scheduled task runner", "error", err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetEnvironmentSubscription checks if the user is subscribed to environment notifications.
func (h *Handler) GetEnvironmentSubscription(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if _, ok := h.loadAuthorizedEnvironment(w, r, user, int32(environmentId)); !ok {
		return
	}

	key := environmentNotificationKey(environmentId)
	subscribed := false
	for _, n := range user.Notifications {
		if n == key {
			subscribed = true
			break
		}
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]bool{"subscribed": subscribed})
}

// SubscribeToEnvironment subscribes the user to environment notifications.
func (h *Handler) SubscribeToEnvironment(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if _, ok := h.loadAuthorizedEnvironment(w, r, user, int32(environmentId)); !ok {
		return
	}

	key := environmentNotificationKey(environmentId)

	for _, n := range user.Notifications {
		if n == key {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	notifications := append(user.Notifications, key)
	notificationsJSON, err := json.Marshal(notifications)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to subscribe")
		return
	}

	if err := h.queries.UpdateUserNotifications(r.Context(), queries.UpdateUserNotificationsParams{
		Notifications: notificationsJSON,
		ID:            user.ID,
	}); err != nil {
		slog.Error("failed to update user notifications", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to subscribe")
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

	if _, ok := h.loadAuthorizedEnvironment(w, r, user, int32(environmentId)); !ok {
		return
	}

	key := environmentNotificationKey(environmentId)
	newNotifications := make([]string, 0, len(user.Notifications))
	for _, n := range user.Notifications {
		if n != key {
			newNotifications = append(newNotifications, n)
		}
	}

	notificationsJSON, err := json.Marshal(newNotifications)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to unsubscribe")
		return
	}

	if err := h.queries.UpdateUserNotifications(r.Context(), queries.UpdateUserNotificationsParams{
		Notifications: notificationsJSON,
		ID:            user.ID,
	}); err != nil {
		slog.Error("failed to update user notifications", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to unsubscribe")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateSitespeedSettings updates sitespeed settings for an environment.
func (h *Handler) UpdateSitespeedSettings(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if _, ok := h.loadAuthorizedEnvironment(w, r, user, int32(environmentId)); !ok {
		return
	}

	var req api.SitespeedSettingsRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var urlsJSON json.RawMessage
	if req.Urls != nil {
		urlsJSON, _ = json.Marshal(req.Urls)
	} else {
		urlsJSON = []byte("[]")
	}

	if err := h.queries.UpdateEnvironmentSitespeedSettings(r.Context(), queries.UpdateEnvironmentSitespeedSettingsParams{
		SitespeedEnabled: req.Enabled,
		SitespeedUrls:    urlsJSON,
		ID:               int32(environmentId),
	}); err != nil {
		slog.Error("failed to update sitespeed settings", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update sitespeed settings")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
