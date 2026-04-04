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

// GetAccountExtensions returns aggregated extensions across all environments the user has access to.
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

	// Group extensions by name, aggregating environments
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
				Environments:  []api.AccountExtensionEnvironment{},
			}
			extMap[key] = ext
			extOrder = append(extOrder, key)
		}

		latestVersion := ""
		if row.LatestVersion != nil {
			latestVersion = *row.LatestVersion
		}

		env := api.AccountExtensionEnvironment{
			EnvironmentId:               int(row.EnvironmentID),
			EnvironmentName:             row.EnvironmentName,
			EnvironmentUrl:              row.EnvironmentUrl,
			EnvironmentOrganizationName: row.EnvironmentOrganizationName,
			EnvironmentOrganizationId:   row.EnvironmentOrganizationID,
			Version:                     row.Version,
			LatestVersion:               latestVersion,
			Active:                      row.Active,
			Installed:                   row.Installed,
		}
		ext.Environments = append(ext.Environments, env)
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
			Id:               row.ID,
			Name:             row.Name,
			Logo:             row.Logo,
			CreatedAt:        pgtimeToTime(row.CreatedAt),
			EnvironmentCount: int(row.EnvironmentCount),
			MemberCount:      int(row.MemberCount),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetAccountEnvironments returns all environments accessible to the user.
func (h *Handler) GetAccountEnvironments(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	rows, err := h.queries.GetUserEnvironments(r.Context(), user.ID)
	if err != nil {
		slog.Error("failed to get user environments", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get environments")
		return
	}

	result := make([]api.AccountEnvironment, 0, len(rows))
	for _, row := range rows {
		var shopId *int
		if row.ShopID != nil {
			v := int(*row.ShopID)
			shopId = &v
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
			ShopId:           shopId,
			ShopName:         row.ShopName,
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
		result = append(result, api.AccountShop{
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

// GetAccountChangelogs returns changelogs across all environments the user has access to.
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
		environmentID := 0
		if row.EnvironmentID != nil {
			environmentID = int(*row.EnvironmentID)
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
			Id:                          int(row.ID),
			EnvironmentId:               environmentID,
			EnvironmentName:             row.EnvironmentName,
			EnvironmentOrganizationName: row.EnvironmentOrganizationName,
			EnvironmentOrganizationId:   row.EnvironmentOrganizationID,
			Extensions:                  extensions,
			OldShopwareVersion:          row.OldShopwareVersion,
			NewShopwareVersion:          row.NewShopwareVersion,
			Date:                        pgtimeToTime(row.Date),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetAccountSubscribedEnvironments returns environments the user is subscribed to for notifications.
func (h *Handler) GetAccountSubscribedEnvironments(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	// User notifications is a list of strings like "environment-123"
	subscribedEnvironments := []api.SubscribedEnvironment{}

	for _, notification := range user.Notifications {
		if strings.HasPrefix(notification, "environment-") {
			idStr := strings.TrimPrefix(notification, "environment-")
			environmentID, err := strconv.Atoi(idStr)
			if err != nil {
				continue
			}

			// Fetch environment name
			environment, err := h.queries.GetEnvironmentByID(r.Context(), int32(environmentID))
			if err != nil {
				slog.Warn("failed to get subscribed environment", "environmentId", environmentID, "error", err)
				continue
			}

			subscribedEnvironments = append(subscribedEnvironments, api.SubscribedEnvironment{
				Id:   environmentID,
				Name: environment.Name,
			})
		}
	}

	httputil.WriteJSON(w, http.StatusOK, subscribedEnvironments)
}

// helper to check for "environment-{id}" pattern match
func environmentNotificationKey(environmentID int) string {
	return fmt.Sprintf("environment-%d", environmentID)
}
