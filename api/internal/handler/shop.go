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

// GetOrganizationShops returns all shops in an organization.
func (h *Handler) GetOrganizationShops(w http.ResponseWriter, r *http.Request, orgId api.OrgId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	rows, err := h.queries.ListShopsByOrganization(r.Context(), orgId)
	if err != nil {
		slog.Error("failed to list shops by organization", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get shops")
		return
	}

	result := make([]api.AccountShop, 0, len(rows))
	for _, row := range rows {
		result = append(result, api.AccountShop{
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
			ProjectName:      row.ProjectName,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetShop returns full shop details.
func (h *Handler) GetShop(w http.ResponseWriter, r *http.Request, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	shop, ok := h.loadAuthorizedShop(w, r, user, int32(shopId))
	if !ok {
		return
	}

	aggregate := h.loadShopDetailAggregate(r.Context(), int32(shopId))
	httputil.WriteJSON(w, http.StatusOK, h.buildShopDetail(shop, aggregate))
}

// CreateShop creates a new shop.
func (h *Handler) CreateShop(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var req api.CreateShopRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.ShopUrl == "" || req.ClientId == "" || req.ClientSecret == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name, shopUrl, clientId, and clientSecret are required")
		return
	}

	projectOrgID, err := h.queries.GetProjectOrganizationID(r.Context(), int32(req.ProjectId))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "project not found")
		return
	}

	if !h.requireOrgMembership(w, r, user, projectOrgID) {
		return
	}

	shopToken := ""
	if req.ShopToken != nil {
		shopToken = *req.ShopToken
	}

	shopID, err := h.runCreateShop(r.Context(), projectOrgID, createShopCommand{
		Name:         req.Name,
		ShopURL:      req.ShopUrl,
		ClientID:     req.ClientId,
		ClientSecret: req.ClientSecret,
		ProjectID:    int32(req.ProjectId),
		ShopToken:    shopToken,
	})
	if err != nil {
		slog.Error("failed to create shop", "error", err)
		httputil.WriteError(w, http.StatusBadRequest, "Cannot reach shop. Check your credentials and shop URL.")
		return
	}

	// Enqueue scrape task
	if err := goqueue.Dispatch(r.Context(), h.bus, jobs.ShopScrape{ShopID: shopID}); err != nil {
		slog.Error("failed to enqueue scrape task", "error", err)
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]interface{}{"id": shopID})
}

// UpdateShop updates an existing shop.
func (h *Handler) UpdateShop(w http.ResponseWriter, r *http.Request, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	shop, ok := h.loadAuthorizedShop(w, r, user, int32(shopId))
	if !ok {
		return
	}

	var req api.UpdateShopRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd, err := h.buildUpdateShopCommand(r.Context(), shop, req)
	if err != nil {
		slog.Error("failed to build shop update", "error", err)
		httputil.WriteError(w, http.StatusBadRequest, "Cannot reach shop with new credentials. Check your credentials and shop URL.")
		return
	}

	if err := h.runUpdateShop(r.Context(), int32(shopId), cmd); err != nil {
		slog.Error("failed to update shop", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update shop")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteShop deletes a shop.
func (h *Handler) DeleteShop(w http.ResponseWriter, r *http.Request, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if _, ok := h.loadAuthorizedShop(w, r, user, int32(shopId)); !ok {
		return
	}

	// Clean up deployment outputs from S3
	if h.storage != nil {
		deployments, err := h.queries.ListDeployments(r.Context(), queries.ListDeploymentsParams{
			ShopID: int32(shopId),
			Limit:  1000,
			Offset: 0,
		})
		if err == nil {
			for _, d := range deployments {
				if err := h.storage.DeleteDeploymentOutput(r.Context(), int(d.ID)); err != nil {
					slog.Warn("failed to delete deployment output", "deploymentId", d.ID, "error", err)
				}
			}
		}
	}

	if err := h.queries.DeleteShop(r.Context(), int32(shopId)); err != nil {
		slog.Error("failed to delete shop", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete shop")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RefreshShop triggers a re-scrape of the shop.
func (h *Handler) RefreshShop(w http.ResponseWriter, r *http.Request, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if _, ok := h.loadAuthorizedShop(w, r, user, int32(shopId)); !ok {
		return
	}

	var body api.RefreshShopJSONRequestBody
	if err := httputil.DecodeBody(r, &body); err != nil && err != io.EOF {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := goqueue.Dispatch(r.Context(), h.bus, jobs.ShopScrape{ShopID: int32(shopId)}); err != nil {
		slog.Error("failed to enqueue scrape task", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to enqueue refresh task")
		return
	}

	if body.Sitespeed != nil && *body.Sitespeed {
		if err := goqueue.Dispatch(r.Context(), h.bus, jobs.SitespeedScrape{ShopID: int32(shopId)}); err != nil {
			slog.Warn("failed to enqueue sitespeed task", "error", err)
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

// ClearShopCache clears the Shopware cache for a shop.
func (h *Handler) ClearShopCache(w http.ResponseWriter, r *http.Request, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	creds, ok := h.loadAuthorizedShopCredentials(w, r, user, int32(shopId))
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
		slog.Error("failed to clear shop cache", "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to clear cache on shop")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RescheduleTask reschedules a scheduled task on the shop.
func (h *Handler) RescheduleTask(w http.ResponseWriter, r *http.Request, shopId api.ShopId, taskId api.TaskId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	creds, ok := h.loadAuthorizedShopCredentials(w, r, user, int32(shopId))
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
		httputil.WriteError(w, http.StatusBadGateway, "failed to reschedule task on shop")
		return
	}

	// Also trigger the scheduled task runner
	if _, err := client.Post(r.Context(), "/_action/scheduled-task/run", nil); err != nil {
		slog.Warn("failed to trigger scheduled task runner", "error", err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetShopSubscription checks if the user is subscribed to shop notifications.
func (h *Handler) GetShopSubscription(w http.ResponseWriter, r *http.Request, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if _, ok := h.loadAuthorizedShop(w, r, user, int32(shopId)); !ok {
		return
	}

	key := shopNotificationKey(shopId)
	subscribed := false
	for _, n := range user.Notifications {
		if n == key {
			subscribed = true
			break
		}
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]bool{"subscribed": subscribed})
}

// SubscribeToShop subscribes the user to shop notifications.
func (h *Handler) SubscribeToShop(w http.ResponseWriter, r *http.Request, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if _, ok := h.loadAuthorizedShop(w, r, user, int32(shopId)); !ok {
		return
	}

	key := shopNotificationKey(shopId)

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

// UnsubscribeFromShop unsubscribes the user from shop notifications.
func (h *Handler) UnsubscribeFromShop(w http.ResponseWriter, r *http.Request, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if _, ok := h.loadAuthorizedShop(w, r, user, int32(shopId)); !ok {
		return
	}

	key := shopNotificationKey(shopId)
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

// UpdateSitespeedSettings updates sitespeed settings for a shop.
func (h *Handler) UpdateSitespeedSettings(w http.ResponseWriter, r *http.Request, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if _, ok := h.loadAuthorizedShop(w, r, user, int32(shopId)); !ok {
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

	if err := h.queries.UpdateShopSitespeedSettings(r.Context(), queries.UpdateShopSitespeedSettingsParams{
		SitespeedEnabled: req.Enabled,
		SitespeedUrls:    urlsJSON,
		ID:               int32(shopId),
	}); err != nil {
		slog.Error("failed to update sitespeed settings", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update sitespeed settings")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
