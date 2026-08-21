package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/access"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	accountread "github.com/friendsofshopware/shopmon/api/internal/readmodel/account"
)

type getAccountMeOutput struct {
	Body api.UserProfile
}

type updateAccountMeInput struct {
	Body api.UpdateAccountMeJSONBody
}

type getAccountExtensionsInput struct {
	Language string `query:"language"`
}

type getAccountExtensionsOutput struct {
	Body []api.AccountExtension
}

type getAccountExtensionInput struct {
	Name     string `path:"name"`
	Language string `query:"language"`
}

type getAccountExtensionOutput struct {
	Body api.AccountExtension
}

type getAccountOrganizationsOutput struct {
	Body []api.AccountOrganization
}

type getAccountEnvironmentsOutput struct {
	Body []api.AccountEnvironment
}

type getAccountShopsOutput struct {
	Body []api.AccountShop
}

type getAccountChangelogsOutput struct {
	Body []api.AccountChangelog
}

type getAccountSubscribedEnvironmentsOutput struct {
	Body []api.SubscribedEnvironment
}

func (h *Handler) GetAccountMe(ctx context.Context, _ *struct{}) (*getAccountMeOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	profile, err := h.account.Profile(ctx, user.ID)
	if err != nil {
		return nil, h.writeAccountReadError(ctx, "get account profile", err)
	}
	return &getAccountMeOutput{Body: profile}, nil
}

// supportedLocales bounds the locales a user may select, matching the frontend
// catalog and the server-side email translator.
var supportedLocales = map[string]bool{"en": true, "de": true}

func (h *Handler) UpdateAccountMe(ctx context.Context, input *updateAccountMeInput) (*getAccountMeOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	if input.Body.Locale != nil {
		if !supportedLocales[*input.Body.Locale] {
			return nil, huma.Error400BadRequest("unsupported locale")
		}
		if err := h.account.UpdateLocale(ctx, user.ID, *input.Body.Locale); err != nil {
			slog.ErrorContext(ctx, "failed to update user locale", "userId", user.ID, "error", err)
			return nil, huma.Error500InternalServerError("failed to update preferences")
		}
	}

	profile, err := h.account.Profile(ctx, user.ID)
	if err != nil {
		return nil, h.writeAccountReadError(ctx, "get account profile", err)
	}
	return &getAccountMeOutput{Body: profile}, nil
}

func (h *Handler) GetAccountExtensions(ctx context.Context, input *getAccountExtensionsInput) (*getAccountExtensionsOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	result, err := h.account.Extensions(ctx, user.ID, activeOrganizationID(ctx), optionalStringQuery(input.Language))
	if err != nil {
		return nil, h.writeAccountReadError(ctx, "get account extensions", err)
	}
	return &getAccountExtensionsOutput{Body: result}, nil
}

func (h *Handler) GetAccountExtension(ctx context.Context, input *getAccountExtensionInput) (*getAccountExtensionOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	result, err := h.account.Extension(ctx, user.ID, activeOrganizationID(ctx), input.Name, optionalStringQuery(input.Language))
	if err != nil {
		return nil, h.writeAccountReadError(ctx, "get account extension", err)
	}
	return &getAccountExtensionOutput{Body: result}, nil
}

func (h *Handler) GetAccountOrganizations(ctx context.Context, _ *struct{}) (*getAccountOrganizationsOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	result, err := h.account.Organizations(ctx, user.ID)
	if err != nil {
		return nil, h.writeAccountReadError(ctx, "get account organizations", err)
	}
	return &getAccountOrganizationsOutput{Body: result}, nil
}

func (h *Handler) GetAccountEnvironments(ctx context.Context, _ *struct{}) (*getAccountEnvironmentsOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	result, err := h.account.Environments(ctx, user.ID, activeOrganizationID(ctx))
	if err != nil {
		return nil, h.writeAccountReadError(ctx, "get account environments", err)
	}
	return &getAccountEnvironmentsOutput{Body: result}, nil
}

func (h *Handler) GetAccountShops(ctx context.Context, _ *struct{}) (*getAccountShopsOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	result, err := h.account.Shops(ctx, user.ID, activeOrganizationID(ctx))
	if err != nil {
		return nil, h.writeAccountReadError(ctx, "get account shops", err)
	}
	return &getAccountShopsOutput{Body: result}, nil
}

func (h *Handler) GetAccountChangelogs(ctx context.Context, _ *struct{}) (*getAccountChangelogsOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	result, err := h.account.Changelogs(ctx, user.ID, activeOrganizationID(ctx))
	if err != nil {
		return nil, h.writeAccountReadError(ctx, "get account changelogs", err)
	}
	return &getAccountChangelogsOutput{Body: result}, nil
}

func (h *Handler) GetAccountSubscribedEnvironments(ctx context.Context, _ *struct{}) (*getAccountSubscribedEnvironmentsOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	return &getAccountSubscribedEnvironmentsOutput{Body: h.account.SubscribedEnvironments(ctx, user.ID)}, nil
}

func (h *Handler) writeAccountReadError(ctx context.Context, operation string, err error) error {
	if errors.Is(err, accountread.ErrExtensionNotFound) {
		return huma.Error404NotFound("extension not found")
	}
	slog.ErrorContext(ctx, "account read failed", "operation", operation, "error", err)
	return huma.Error500InternalServerError("failed to load account data")
}

func activeOrganizationID(ctx context.Context) *string {
	session := access.SessionFromContext(ctx)
	if session == nil {
		return nil
	}
	return session.ActiveOrganizationID
}

func optionalStringQuery(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
