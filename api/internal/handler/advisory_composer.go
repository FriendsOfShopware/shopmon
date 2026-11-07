package handler

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/api"
)

type composerAdvisoriesOutput struct {
	Body api.ComposerAdvisoryResponse
}

func (h *Handler) ComposerAdvisories(ctx context.Context, _ *struct{}) (*composerAdvisoriesOutput, error) {
	body, err := h.advisories.ComposerFeed(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load Composer advisory feed", "error", err)
		return nil, huma.Error500InternalServerError("failed to load Composer advisory feed")
	}
	return &composerAdvisoriesOutput{Body: body}, nil
}
