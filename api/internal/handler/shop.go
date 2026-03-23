package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	goqueue "github.com/shyim/go-queue"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/crypto"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/jobs"
	"github.com/friendsofshopware/shopmon/api/internal/shopware"
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

	shop, err := h.queries.GetShopByID(r.Context(), int32(shopId))
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
		return
	}

	if !h.requireOrgMembership(w, r, user, shop.OrganizationID) {
		return
	}

	extRows, err := h.queries.GetShopExtensions(r.Context(), int32(shopId))
	if err != nil {
		slog.Error("failed to get shop extensions", "error", err)
		extRows = nil
	}
	extensions := make([]api.ShopExtension, 0, len(extRows))
	for _, e := range extRows {
		active := e.Active
		installed := e.Installed
		name := e.Name
		label := e.Label
		version := e.Version

		var ratingAvg *float32
		if e.RatingAverage != nil {
			v := float32(*e.RatingAverage)
			ratingAvg = &v
		}

		var changelogStr *string
		if e.Changelog != nil && len(e.Changelog) > 0 {
			s := string(e.Changelog)
			changelogStr = &s
		}

		var installedAt *time.Time
		if e.InstalledAt != nil {
			if t, parseErr := time.Parse(time.RFC3339, *e.InstalledAt); parseErr == nil {
				installedAt = &t
			}
		}

		latestVersion := ""
		if e.LatestVersion != nil {
			latestVersion = *e.LatestVersion
		}

		extensions = append(extensions, api.ShopExtension{
			Name:          name,
			Label:         label,
			Version:       version,
			LatestVersion: latestVersion,
			Active:        active,
			Installed:     installed,
			StoreLink:     e.StoreLink,
			RatingAverage: ratingAvg,
			Changelog:     changelogStr,
			InstalledAt:   installedAt,
		})
	}

	taskRows, err := h.queries.GetShopScheduledTasks(r.Context(), int32(shopId))
	if err != nil {
		slog.Error("failed to get shop scheduled tasks", "error", err)
		taskRows = nil
	}
	tasks := make([]api.ScheduledTask, 0, len(taskRows))
	for _, t := range taskRows {
		taskID := t.TaskID
		name := t.Name
		status := t.Status
		interval := int(t.Interval)
		overdue := t.Overdue

		var lastExec *time.Time
		if t.LastExecutionTime != nil {
			if parsed, parseErr := time.Parse(time.RFC3339, *t.LastExecutionTime); parseErr == nil {
				lastExec = &parsed
			}
		}
		var nextExec *time.Time
		if t.NextExecutionTime != nil {
			if parsed, parseErr := time.Parse(time.RFC3339, *t.NextExecutionTime); parseErr == nil {
				nextExec = &parsed
			}
		}

		tasks = append(tasks, api.ScheduledTask{
			Id:                taskID,
			Name:              name,
			Status:            status,
			RunInterval:       interval,
			Overdue:           overdue,
			LastExecutionTime: lastExec,
			NextExecutionTime: nextExec,
		})
	}

	queueRows, err := h.queries.GetShopQueues(r.Context(), int32(shopId))
	if err != nil {
		slog.Error("failed to get shop queues", "error", err)
		queueRows = nil
	}
	queues := make([]api.Queue, 0, len(queueRows))
	for _, q := range queueRows {
		name := q.Name
		size := int(q.Size)
		queues = append(queues, api.Queue{
			Name: name,
			Size: size,
		})
	}

	var cacheInfo *api.CacheInfo
	cacheRow, err := h.queries.GetShopCache(r.Context(), int32(shopId))
	if err == nil {
		id := int(cacheRow.ID)
		cacheInfo = &api.CacheInfo{
			Id:           &id,
			Environment:  &cacheRow.Environment,
			HttpCache:    &cacheRow.HttpCache,
			CacheAdapter: &cacheRow.CacheAdapter,
		}
	}

	checkRows, err := h.queries.GetShopChecks(r.Context(), int32(shopId))
	if err != nil {
		slog.Error("failed to get shop checks", "error", err)
		checkRows = nil
	}
	checks := make([]api.ShopCheck, 0, len(checkRows))
	for _, c := range checkRows {
		checkID := c.CheckID
		level := c.Level
		message := c.Message
		checks = append(checks, api.ShopCheck{
			Id:      checkID,
			Level:   level,
			Message: message,
			Link:    c.Link,
		})
	}

	shopIDPtr := int32(shopId)
	sitespeedRows, err := h.queries.GetShopSitespeeds(r.Context(), &shopIDPtr)
	if err != nil {
		slog.Error("failed to get shop sitespeeds", "error", err)
		sitespeedRows = nil
	}
	sitespeeds := make([]api.Sitespeed, 0, len(sitespeedRows))
	for _, s := range sitespeedRows {
		createdAt := pgtimeToTime(s.CreatedAt)

		var ttfb, fullyLoaded, lcp, fcp, transferSize *float32
		var cls *float32
		if s.Ttfb != nil {
			v := float32(*s.Ttfb)
			ttfb = &v
		}
		if s.FullyLoaded != nil {
			v := float32(*s.FullyLoaded)
			fullyLoaded = &v
		}
		if s.LargestContentfulPaint != nil {
			v := float32(*s.LargestContentfulPaint)
			lcp = &v
		}
		if s.FirstContentfulPaint != nil {
			v := float32(*s.FirstContentfulPaint)
			fcp = &v
		}
		if s.CumulativeLayoutShift != nil {
			cls = s.CumulativeLayoutShift
		}
		if s.TransferSize != nil {
			v := float32(*s.TransferSize)
			transferSize = &v
		}

		var deployment *api.SitespeedDeployment
		if s.DeploymentID != nil {
			deployment = &api.SitespeedDeployment{
				Id:   int(*s.DeploymentID),
				Name: fmt.Sprintf("Deployment #%d", *s.DeploymentID),
			}
		}

		sitespeeds = append(sitespeeds, api.Sitespeed{
			CreatedAt:              createdAt,
			Ttfb:                   ttfb,
			FullyLoaded:            fullyLoaded,
			LargestContentfulPaint: lcp,
			FirstContentfulPaint:   fcp,
			CumulativeLayoutShift:  cls,
			TransferSize:           transferSize,
			Deployment:             deployment,
		})
	}

	// Fetch changelogs for this shop
	shopIDForChangelog := int32(shopId)
	changelogRows, err := h.queries.GetShopChangelogs(r.Context(), &shopIDForChangelog)
	if err != nil {
		slog.Error("failed to get shop changelogs", "error", err)
		changelogRows = nil
	}
	changelogs := make([]api.AccountChangelog, 0, len(changelogRows))
	for _, c := range changelogRows {
		clShopID := 0
		if c.ShopID != nil {
			clShopID = int(*c.ShopID)
		}
		var ext []api.ExtensionDiff
		if c.Extensions != nil {
			json.Unmarshal(c.Extensions, &ext)
		}
		if ext == nil {
			ext = []api.ExtensionDiff{}
		}
		changelogs = append(changelogs, api.AccountChangelog{
			Id:                   int(c.ID),
			ShopId:               clShopID,
			ShopName:             shop.Name,
			ShopOrganizationName: shop.OrganizationName,
			ShopOrganizationId:   shop.OrganizationID,
			Extensions:           ext,
			OldShopwareVersion:   c.OldShopwareVersion,
			NewShopwareVersion:   c.NewShopwareVersion,
			Date:                 pgtimeToTime(c.Date),
		})
	}

	// Deployment count
	deployCount, err := h.queries.CountShopDeployments(r.Context(), int32(shopId))
	if err != nil {
		deployCount = 0
	}

	var ignores *[]string
	if shop.Ignores != nil && len(shop.Ignores) > 0 {
		var ign []string
		if err := json.Unmarshal(shop.Ignores, &ign); err == nil {
			ignores = &ign
		}
	}

	var sitespeedUrls *[]string
	if shop.SitespeedUrls != nil && len(shop.SitespeedUrls) > 0 {
		var urls []string
		if err := json.Unmarshal(shop.SitespeedUrls, &urls); err == nil {
			sitespeedUrls = &urls
		}
	}

	var lastChangelog *api.AccountChangelog
	if shop.LastChangelog != nil && len(shop.LastChangelog) > 0 {
		var lc api.AccountChangelog
		if err := json.Unmarshal(shop.LastChangelog, &lc); err == nil && !lc.Date.IsZero() {
			lastChangelog = &lc
		}
	}

	projectID := int(shop.ProjectID)
	var projectIDPtr *int
	if shop.ProjectID > 0 {
		projectIDPtr = &projectID
	}

	detail := api.ShopDetail{
		Id:                 int(shop.ID),
		Name:               shop.Name,
		Url:                shop.Url,
		Favicon:            shop.Favicon,
		Status:             shop.Status,
		ShopwareVersion:    shop.ShopwareVersion,
		LastScrapedAt:      pgtimeToTimePtr(shop.LastScrapedAt),
		LastScrapedError:   shop.LastScrapedError,
		OrganizationId:     shop.OrganizationID,
		OrganizationName:   shop.OrganizationName,
		ProjectId:          projectIDPtr,
		ProjectName:        shop.ProjectName,
		ProjectDescription: shop.ProjectDescription,
		ShopImage:          shop.ShopImage,
		ShopToken:          shop.ShopToken,
		Ignores:            ignores,
		CreatedAt:          pgtimeToTime(shop.CreatedAt),
		SitespeedEnabled:   shop.SitespeedEnabled,
		SitespeedDetailUrl: sitespeedDetailUrl(h.cfg, shop.ID, shop.SitespeedEnabled),
		SitespeedUrls:      sitespeedUrls,
		Extensions:         extensions,
		ScheduledTasks:     tasks,
		Queues:             queues,
		Cache:              cacheInfo,
		Checks:             checks,
		Sitespeeds:         sitespeeds,
		Changelogs:         changelogs,
		DeploymentsCount:   int(deployCount),
		LastChangelog:      lastChangelog,
	}

	httputil.WriteJSON(w, http.StatusOK, detail)
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

	// Encrypt client secret
	encryptedSecret, err := crypto.Encrypt(req.ClientSecret, h.cfg.AppSecret)
	if err != nil {
		slog.Error("failed to encrypt client secret", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create shop")
		return
	}

	shopToken := ""
	if req.ShopToken != nil {
		shopToken = *req.ShopToken
	}

	// Validate shop connectivity before creating
	swClient := shopware.NewClient(req.ShopUrl, req.ClientId, req.ClientSecret, shopToken)
	configData, err := swClient.Get(r.Context(), "/_info/config")
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Cannot reach shop. Check your credentials and shop URL.")
		return
	}

	var shopConfig struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(configData, &shopConfig); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "Invalid response from shop")
		return
	}

	shopID, err := h.queries.CreateShop(r.Context(), queries.CreateShopParams{
		OrganizationID:  projectOrgID,
		ProjectID:       int32(req.ProjectId),
		Name:            req.Name,
		Url:             req.ShopUrl,
		ClientID:        req.ClientId,
		ClientSecret:    encryptedSecret,
		ShopwareVersion: shopConfig.Version,
		ShopToken:       shopToken,
	})
	if err != nil {
		slog.Error("failed to create shop", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create shop")
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

	shop, err := h.queries.GetShopByID(r.Context(), int32(shopId))
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
		return
	}

	if !h.requireOrgMembership(w, r, user, shop.OrganizationID) {
		return
	}

	var req api.UpdateShopRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Merge with existing values
	name := shop.Name
	if req.Name != nil {
		name = *req.Name
	}
	url := shop.Url
	if req.ShopUrl != nil {
		url = *req.ShopUrl
	}
	clientID := shop.ClientID
	if req.ClientId != nil {
		clientID = *req.ClientId
	}
	clientSecret := shop.ClientSecret
	if req.ClientSecret != nil {
		// Validate new credentials before saving
		swClient := shopware.NewClient(url, clientID, *req.ClientSecret, shop.ShopToken)
		if _, err := swClient.Get(r.Context(), "/_info/config"); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "Cannot reach shop with new credentials. Check your credentials and shop URL.")
			return
		}

		encrypted, err := crypto.Encrypt(*req.ClientSecret, h.cfg.AppSecret)
		if err != nil {
			slog.Error("failed to encrypt client secret", "error", err)
			httputil.WriteError(w, http.StatusInternalServerError, "failed to update shop")
			return
		}
		clientSecret = encrypted
	}

	ignores := shop.Ignores
	if req.Ignores != nil {
		ignoresJSON, err := json.Marshal(req.Ignores)
		if err == nil {
			ignores = ignoresJSON
		}
	}

	projectID := req.ProjectId

	if err := h.queries.UpdateShop(r.Context(), queries.UpdateShopParams{
		Name:         name,
		Url:          url,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Ignores:      ignores,
		ProjectID:    int32(projectID),
		ID:           int32(shopId),
	}); err != nil {
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

	orgID, err := h.queries.GetShopOrganizationID(r.Context(), int32(shopId))
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
		return
	}

	if !h.requireOrgMembership(w, r, user, orgID) {
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

	orgID, err := h.queries.GetShopOrganizationID(r.Context(), int32(shopId))
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
		return
	}

	if !h.requireOrgMembership(w, r, user, orgID) {
		return
	}

	var body api.RefreshShopJSONRequestBody
	httputil.DecodeBody(r, &body)

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

	creds, err := h.queries.GetShopCredentials(r.Context(), int32(shopId))
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
		return
	}

	orgID, err := h.queries.GetShopOrganizationID(r.Context(), int32(shopId))
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
		return
	}

	if !h.requireOrgMembership(w, r, user, orgID) {
		return
	}

	// Decrypt credentials
	decryptedSecret, err := crypto.Decrypt(creds.ClientSecret, h.cfg.AppSecret)
	if err != nil {
		slog.Error("failed to decrypt client secret", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to clear cache")
		return
	}

	client := shopware.NewClient(creds.Url, creds.ClientID, decryptedSecret, creds.ShopToken)
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

	creds, err := h.queries.GetShopCredentials(r.Context(), int32(shopId))
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
		return
	}

	orgID, err := h.queries.GetShopOrganizationID(r.Context(), int32(shopId))
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
		return
	}

	if !h.requireOrgMembership(w, r, user, orgID) {
		return
	}

	decryptedSecret, err := crypto.Decrypt(creds.ClientSecret, h.cfg.AppSecret)
	if err != nil {
		slog.Error("failed to decrypt client secret", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to reschedule task")
		return
	}

	client := shopware.NewClient(creds.Url, creds.ClientID, decryptedSecret, creds.ShopToken)

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

	orgID, err := h.queries.GetShopOrganizationID(r.Context(), int32(shopId))
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
		return
	}
	if !h.requireOrgMembership(w, r, user, orgID) {
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

	orgID, err := h.queries.GetShopOrganizationID(r.Context(), int32(shopId))
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
		return
	}
	if !h.requireOrgMembership(w, r, user, orgID) {
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

	orgID, err := h.queries.GetShopOrganizationID(r.Context(), int32(shopId))
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
		return
	}
	if !h.requireOrgMembership(w, r, user, orgID) {
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

	orgID, err := h.queries.GetShopOrganizationID(r.Context(), int32(shopId))
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
		return
	}

	if !h.requireOrgMembership(w, r, user, orgID) {
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
