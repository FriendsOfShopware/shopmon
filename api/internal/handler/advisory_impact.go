package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	advisoryread "github.com/friendsofshopware/shopmon/api/internal/readmodel/advisory"
)

type adminPreviewAdvisoryImpactInput struct {
	Body api.AdminPreviewAdvisoryImpactRequest
}

type adminPreviewAdvisoryImpactOutput struct {
	Body api.AdminAdvisoryImpact
}

func (h *Handler) AdminPreviewAdvisoryImpact(ctx context.Context, input *adminPreviewAdvisoryImpactInput) (*adminPreviewAdvisoryImpactOutput, error) {
	if _, err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	packages := make([]advisoryread.PackageInput, 0, len(input.Body.Packages))
	for _, pkg := range input.Body.Packages {
		packages = append(packages, advisoryread.PackageInput{PackageName: pkg.PackageName, AffectedVersions: pkg.AffectedVersions})
	}
	result, err := h.advisories.PreviewImpact(ctx, packages)
	if err != nil {
		if errors.Is(err, advisoryread.ErrInvalidInput) {
			return nil, huma.Error400BadRequest(err.Error())
		}
		slog.ErrorContext(ctx, "failed to preview advisory impact", "error", err)
		return nil, huma.Error500InternalServerError("failed to preview advisory impact")
	}
	return &adminPreviewAdvisoryImpactOutput{Body: result}, nil
}
