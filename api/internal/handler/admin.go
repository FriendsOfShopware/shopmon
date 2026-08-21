package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	adminread "github.com/friendsofshopware/shopmon/api/internal/readmodel/admin"
)

type adminGetStatsOutput struct {
	Body api.AdminStats
}

func (h *Handler) AdminGetStats(ctx context.Context, _ *struct{}) (*adminGetStatsOutput, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := h.admin.Stats(ctx)
	if err != nil {
		return nil, h.writeAdminReadError(ctx, "get stats", err)
	}
	return &adminGetStatsOutput{Body: result}, nil
}

type adminGetOrganizationsInput struct {
	Limit          int    `query:"limit"`
	Offset         int    `query:"offset"`
	SortBy         string `query:"sortBy" enum:"name,createdAt,shopCount,memberCount"`
	SortDirection  string `query:"sortDirection" enum:"asc,desc"`
	SearchField    string `query:"searchField"`
	SearchOperator string `query:"searchOperator"`
	SearchValue    string `query:"searchValue"`
	FilterField    string `query:"filterField"`
	FilterOperator string `query:"filterOperator"`
	FilterValue    string `query:"filterValue"`
}

type adminGetOrganizationsOutput struct {
	Body api.AdminOrganizationsResponse
}

func (h *Handler) AdminGetOrganizations(ctx context.Context, input *adminGetOrganizationsInput) (*adminGetOrganizationsOutput, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := h.admin.Organizations(ctx, api.AdminGetOrganizationsParams{
		Limit:          optionalQuery(input.Limit),
		Offset:         optionalQuery(input.Offset),
		SortBy:         optionalQuery(api.AdminGetOrganizationsParamsSortBy(input.SortBy)),
		SortDirection:  optionalQuery(api.AdminGetOrganizationsParamsSortDirection(input.SortDirection)),
		SearchField:    optionalQuery(input.SearchField),
		SearchOperator: optionalQuery(input.SearchOperator),
		SearchValue:    optionalQuery(input.SearchValue),
		FilterField:    optionalQuery(input.FilterField),
		FilterOperator: optionalQuery(input.FilterOperator),
		FilterValue:    optionalQuery(input.FilterValue),
	})
	if err != nil {
		return nil, h.writeAdminReadError(ctx, "list organizations", err)
	}
	return &adminGetOrganizationsOutput{Body: result}, nil
}

type adminGetEnvironmentsInput struct {
	Limit          int    `query:"limit"`
	Offset         int    `query:"offset"`
	SortBy         string `query:"sortBy"`
	SortDirection  string `query:"sortDirection" enum:"asc,desc"`
	SearchField    string `query:"searchField"`
	SearchOperator string `query:"searchOperator"`
	SearchValue    string `query:"searchValue"`
	FilterField    string `query:"filterField"`
	FilterOperator string `query:"filterOperator"`
	FilterValue    string `query:"filterValue"`
}

type adminGetEnvironmentsOutput struct {
	Body api.AdminEnvironmentsResponse
}

func (h *Handler) AdminGetEnvironments(ctx context.Context, input *adminGetEnvironmentsInput) (*adminGetEnvironmentsOutput, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := h.admin.Environments(ctx, api.AdminGetEnvironmentsParams{
		Limit:          optionalQuery(input.Limit),
		Offset:         optionalQuery(input.Offset),
		SortBy:         optionalQuery(input.SortBy),
		SortDirection:  optionalQuery(api.AdminGetEnvironmentsParamsSortDirection(input.SortDirection)),
		SearchField:    optionalQuery(input.SearchField),
		SearchOperator: optionalQuery(input.SearchOperator),
		SearchValue:    optionalQuery(input.SearchValue),
		FilterField:    optionalQuery(input.FilterField),
		FilterOperator: optionalQuery(input.FilterOperator),
		FilterValue:    optionalQuery(input.FilterValue),
	})
	if err != nil {
		return nil, h.writeAdminReadError(ctx, "list environments", err)
	}
	return &adminGetEnvironmentsOutput{Body: result}, nil
}

type adminGetGrowthOutput struct {
	Body api.AdminGrowth
}

func (h *Handler) AdminGetGrowth(ctx context.Context, _ *struct{}) (*adminGetGrowthOutput, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := h.admin.Growth(ctx)
	if err != nil {
		return nil, h.writeAdminReadError(ctx, "get growth", err)
	}
	return &adminGetGrowthOutput{Body: result}, nil
}

type adminGetRecentActivityOutput struct {
	Body api.AdminRecentActivity
}

func (h *Handler) AdminGetRecentActivity(ctx context.Context, _ *struct{}) (*adminGetRecentActivityOutput, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := h.admin.RecentActivity(ctx)
	if err != nil {
		return nil, h.writeAdminReadError(ctx, "get recent activity", err)
	}
	return &adminGetRecentActivityOutput{Body: result}, nil
}

type adminGetShopwareVersionsOutput struct {
	Body []api.ShopwareVersionCount
}

func (h *Handler) AdminGetShopwareVersions(ctx context.Context, _ *struct{}) (*adminGetShopwareVersionsOutput, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := h.admin.ShopwareVersions(ctx)
	if err != nil {
		return nil, h.writeAdminReadError(ctx, "get Shopware versions", err)
	}
	return &adminGetShopwareVersionsOutput{Body: result}, nil
}

type adminGetOrganizationDetailInput struct {
	OrgID string `path:"orgId"`
}

type adminGetOrganizationDetailOutput struct {
	Body api.AdminOrganizationDetail
}

func (h *Handler) AdminGetOrganizationDetail(ctx context.Context, input *adminGetOrganizationDetailInput) (*adminGetOrganizationDetailOutput, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := h.admin.OrganizationDetail(ctx, input.OrgID)
	if err != nil {
		return nil, h.writeAdminReadError(ctx, "get organization detail", err)
	}
	return &adminGetOrganizationDetailOutput{Body: result}, nil
}

type adminGetEnvironmentDetailInput struct {
	EnvID int `path:"envId"`
}

type adminGetEnvironmentDetailOutput struct {
	Body api.AdminEnvironmentDetail
}

func (h *Handler) AdminGetEnvironmentDetail(ctx context.Context, input *adminGetEnvironmentDetailInput) (*adminGetEnvironmentDetailOutput, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := h.admin.EnvironmentDetail(ctx, input.EnvID)
	if err != nil {
		return nil, h.writeAdminReadError(ctx, "get environment detail", err)
	}
	return &adminGetEnvironmentDetailOutput{Body: result}, nil
}

type adminGetAuditLogInput struct {
	Limit        int    `query:"limit"`
	Offset       int    `query:"offset"`
	Action       string `query:"action"`
	ActorUserId  string `query:"actorUserId"`
	TargetUserId string `query:"targetUserId"`
}

type adminGetAuditLogOutput struct {
	Body api.AdminAuditLogResponse
}

func (h *Handler) AdminGetAuditLog(ctx context.Context, input *adminGetAuditLogInput) (*adminGetAuditLogOutput, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	result, err := h.admin.AuditLog(ctx, api.AdminGetAuditLogParams{
		Limit:        optionalQuery(input.Limit),
		Offset:       optionalQuery(input.Offset),
		Action:       optionalQuery(input.Action),
		ActorUserId:  optionalQuery(input.ActorUserId),
		TargetUserId: optionalQuery(input.TargetUserId),
	})
	if err != nil {
		return nil, h.writeAdminReadError(ctx, "get audit log", err)
	}
	return &adminGetAuditLogOutput{Body: result}, nil
}

func optionalQuery[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}

func (h *Handler) writeAdminReadError(ctx context.Context, operation string, err error) error {
	switch {
	case errors.Is(err, adminread.ErrOrganizationNotFound):
		return huma.Error404NotFound("organization not found")
	case errors.Is(err, adminread.ErrEnvironmentNotFound):
		return huma.Error404NotFound("environment not found")
	default:
		slog.ErrorContext(ctx, "admin read failed", "operation", operation, "error", err)
		return huma.Error500InternalServerError("failed to load admin data")
	}
}
