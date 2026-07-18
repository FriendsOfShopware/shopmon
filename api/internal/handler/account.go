package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	accountread "github.com/friendsofshopware/shopmon/api/internal/readmodel/account"
)

func (h *Handler) GetAccountMe(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	profile, err := h.account.Profile(r.Context(), user.ID)
	if err != nil {
		h.writeAccountReadError(w, r, "get account profile", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, profile)
}

// supportedLocales bounds the locales a user may select, matching the frontend
// catalog and the server-side email translator.
var supportedLocales = map[string]bool{"en": true, "de": true}

// UpdateAccountMe updates the current user's preferences (currently locale).
func (h *Handler) UpdateAccountMe(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var body api.UpdateAccountMeJSONRequestBody
	if err := httputil.DecodeBody(r, &body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Locale != nil {
		if !supportedLocales[*body.Locale] {
			httputil.WriteError(w, http.StatusBadRequest, "unsupported locale")
			return
		}
		if err := h.account.UpdateLocale(r.Context(), user.ID, *body.Locale); err != nil {
			slog.Error("failed to update user locale", "userId", user.ID, "error", err)
			httputil.WriteError(w, http.StatusInternalServerError, "failed to update preferences")
			return
		}
	}

	profile, err := h.account.Profile(r.Context(), user.ID)
	if err != nil {
		h.writeAccountReadError(w, r, "get account profile", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, profile)
}

func (h *Handler) GetAccountExtensions(w http.ResponseWriter, r *http.Request, params api.GetAccountExtensionsParams) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	var language *string
	if params.Language != nil {
		value := string(*params.Language)
		language = &value
	}
	result, err := h.account.Extensions(r.Context(), user.ID, h.getActiveOrganizationID(r), language)
	if err != nil {
		h.writeAccountReadError(w, r, "get account extensions", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetAccountExtension(w http.ResponseWriter, r *http.Request, name string, params api.GetAccountExtensionParams) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	var language *string
	if params.Language != nil {
		value := string(*params.Language)
		language = &value
	}
	result, err := h.account.Extension(r.Context(), user.ID, h.getActiveOrganizationID(r), name, language)
	if err != nil {
		h.writeAccountReadError(w, r, "get account extension", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetAccountOrganizations(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	result, err := h.account.Organizations(r.Context(), user.ID)
	if err != nil {
		h.writeAccountReadError(w, r, "get account organizations", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetAccountEnvironments(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	result, err := h.account.Environments(r.Context(), user.ID, h.getActiveOrganizationID(r))
	if err != nil {
		h.writeAccountReadError(w, r, "get account environments", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetAccountShops(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	result, err := h.account.Shops(r.Context(), user.ID, h.getActiveOrganizationID(r))
	if err != nil {
		h.writeAccountReadError(w, r, "get account shops", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetAccountChangelogs(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	result, err := h.account.Changelogs(r.Context(), user.ID, h.getActiveOrganizationID(r))
	if err != nil {
		h.writeAccountReadError(w, r, "get account changelogs", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetAccountSubscribedEnvironments(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	httputil.WriteJSON(w, http.StatusOK, h.account.SubscribedEnvironments(r.Context(), user.ID))
}

func (h *Handler) writeAccountReadError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	if errors.Is(err, accountread.ErrExtensionNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "extension not found")
		return
	}
	slog.ErrorContext(r.Context(), "account read failed", "operation", operation, "error", err)
	httputil.WriteError(w, http.StatusInternalServerError, "failed to load account data")
}
