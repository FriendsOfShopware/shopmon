package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// GetAccountMe returns the current user's profile.
func (h *Handler) GetAccountMe(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	dbUser, err := h.queries.GetUserByID(r.Context(), user.ID)
	if err != nil {
		slog.Error("failed to get user", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	profile := api.UserProfile{
		Id:          dbUser.ID,
		DisplayName: dbUser.Name,
		Email:       openapi_types.Email(dbUser.Email),
		CreatedAt:   pgtimeToTime(dbUser.CreatedAt),
	}

	httputil.WriteJSON(w, http.StatusOK, profile)
}

// GetAccountExtensions returns aggregated extensions across all shops the user has access to.
func (h *Handler) GetAccountExtensions(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	rows, err := h.queries.GetUserExtensions(r.Context(), user.ID)
	if err != nil {
		slog.Error("failed to get user extensions", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get extensions")
		return
	}

	// Group extensions by name, aggregating shops
	extMap := make(map[string]*api.AccountExtension)
	extOrder := []string{}

	for _, row := range rows {
		key := row.Name

		ext, exists := extMap[key]
		if !exists {
			latestVersion := ""
			if row.LatestVersion != nil {
				latestVersion = *row.LatestVersion
			}

			var ratingAvg *float32
			if row.RatingAverage != nil {
				v := float32(*row.RatingAverage)
				ratingAvg = &v
			}

			var changelogStr *string
			if len(row.Changelog) > 0 {
				s := string(row.Changelog)
				changelogStr = &s
			}

			var installedAt *time.Time
			if row.InstalledAt != nil {
				if t, err := time.Parse(time.RFC3339, *row.InstalledAt); err == nil {
					installedAt = &t
				}
			}

			ext = &api.AccountExtension{
				Name:          row.Name,
				Label:         row.Label,
				Version:       row.Version,
				LatestVersion: latestVersion,
				Active:        row.Active,
				Installed:     row.Installed,
				StoreLink:     row.StoreLink,
				RatingAverage: ratingAvg,
				Changelog:     changelogStr,
				InstalledAt:   installedAt,
				Shops:         []api.AccountExtensionShop{},
			}
			extMap[key] = ext
			extOrder = append(extOrder, key)
		}

		latestVersion := ""
		if row.LatestVersion != nil {
			latestVersion = *row.LatestVersion
		}

		shop := api.AccountExtensionShop{
			ShopId:               int(row.ShopID),
			ShopName:             row.ShopName,
			ShopUrl:              row.ShopUrl,
			ShopOrganizationName: row.ShopOrganizationName,
			ShopOrganizationId:   row.ShopOrganizationID,
			Version:              row.Version,
			LatestVersion:        latestVersion,
			Active:               row.Active,
			Installed:            row.Installed,
		}
		ext.Shops = append(ext.Shops, shop)
	}

	result := make([]api.AccountExtension, 0, len(extOrder))
	for _, key := range extOrder {
		result = append(result, *extMap[key])
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetAccountOrganizations returns the organizations the user belongs to.
func (h *Handler) GetAccountOrganizations(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	rows, err := h.queries.GetUserOrganizations(r.Context(), user.ID)
	if err != nil {
		slog.Error("failed to get user organizations", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get organizations")
		return
	}

	result := make([]api.AccountOrganization, 0, len(rows))
	for _, row := range rows {
		result = append(result, api.AccountOrganization{
			Id:          row.ID,
			Name:        row.Name,
			Logo:        row.Logo,
			CreatedAt:   pgtimeToTime(row.CreatedAt),
			ShopCount:   int(row.ShopCount),
			MemberCount: int(row.MemberCount),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetAccountShops returns all shops accessible to the user.
func (h *Handler) GetAccountShops(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	rows, err := h.queries.GetUserShops(r.Context(), user.ID)
	if err != nil {
		slog.Error("failed to get user shops", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get shops")
		return
	}

	result := make([]api.AccountShop, 0, len(rows))
	for _, row := range rows {
		var projectId *int
		if row.ProjectID != nil {
			v := int(*row.ProjectID)
			projectId = &v
		}
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
			ProjectId:        projectId,
			ProjectName:      row.ProjectName,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetAccountProjects returns all projects accessible to the user.
func (h *Handler) GetAccountProjects(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	rows, err := h.queries.GetUserProjects(r.Context(), user.ID)
	if err != nil {
		slog.Error("failed to get user projects", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get projects")
		return
	}

	result := make([]api.AccountProject, 0, len(rows))
	for _, row := range rows {
		result = append(result, api.AccountProject{
			Id:               int(row.ID),
			Name:             row.Name,
			Description:      row.Description,
			GitUrl:           row.GitUrl,
			OrganizationId:   row.OrganizationID,
			OrganizationName: row.OrganizationName,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetAccountChangelogs returns changelogs across all shops the user has access to.
func (h *Handler) GetAccountChangelogs(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	rows, err := h.queries.GetUserChangelogs(r.Context(), user.ID)
	if err != nil {
		slog.Error("failed to get user changelogs", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get changelogs")
		return
	}

	result := make([]api.AccountChangelog, 0, len(rows))
	for _, row := range rows {
		shopID := 0
		if row.ShopID != nil {
			shopID = int(*row.ShopID)
		}

		var extensions []api.ExtensionDiff
		if row.Extensions != nil {
			if err := json.Unmarshal(row.Extensions, &extensions); err != nil {
				slog.Error("failed to unmarshal changelog extensions", "error", err)
			}
		}
		if extensions == nil {
			extensions = []api.ExtensionDiff{}
		}

		result = append(result, api.AccountChangelog{
			Id:                   int(row.ID),
			ShopId:               shopID,
			ShopName:             row.ShopName,
			ShopOrganizationName: row.ShopOrganizationName,
			ShopOrganizationId:   row.ShopOrganizationID,
			Extensions:           extensions,
			OldShopwareVersion:   row.OldShopwareVersion,
			NewShopwareVersion:   row.NewShopwareVersion,
			Date:                 pgtimeToTime(row.Date),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetAccountSubscribedShops returns shops the user is subscribed to for notifications.
func (h *Handler) GetAccountSubscribedShops(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	// User notifications is a list of strings like "shop-123"
	subscribedShops := []api.SubscribedShop{}

	for _, notification := range user.Notifications {
		if strings.HasPrefix(notification, "shop-") {
			idStr := strings.TrimPrefix(notification, "shop-")
			shopID, err := strconv.Atoi(idStr)
			if err != nil {
				continue
			}

			// Fetch shop name
			shop, err := h.queries.GetShopByID(r.Context(), int32(shopID))
			if err != nil {
				slog.Warn("failed to get subscribed shop", "shopId", shopID, "error", err)
				continue
			}

			subscribedShops = append(subscribedShops, api.SubscribedShop{
				Id:   shopID,
				Name: shop.Name,
			})
		}
	}

	httputil.WriteJSON(w, http.StatusOK, subscribedShops)
}

// helper to check for "shop-{id}" pattern match
func shopNotificationKey(shopID int) string {
	return fmt.Sprintf("shop-%d", shopID)
}
