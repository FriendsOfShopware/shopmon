package handler

import (
	"context"
	"log/slog"
	"net/url"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
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

type composerAdvisoriesInput struct {
	RawBody []byte `contentType:"application/x-www-form-urlencoded"`
}

func (h *Handler) ComposerAdvisoriesForPackages(ctx context.Context, input *composerAdvisoriesInput) (*composerAdvisoriesOutput, error) {
	form, err := url.ParseQuery(string(input.RawBody))
	if err != nil {
		slog.ErrorContext(ctx, "failed to parse Composer advisory request", "error", err)
		return nil, huma.Error400BadRequest("invalid form body")
	}
	requested := make(map[string]struct{})
	for _, name := range httputil.PHPFormArray(form, "packages") {
		requested[strings.ToLower(name)] = struct{}{}
	}
	if len(requested) == 0 {
		return &composerAdvisoriesOutput{Body: api.ComposerAdvisoryResponse{Advisories: map[string][]api.ComposerAdvisory{}}}, nil
	}
	output, err := h.ComposerAdvisories(ctx, &struct{}{})
	if err != nil {
		return nil, err
	}
	for name := range output.Body.Advisories {
		if _, ok := requested[strings.ToLower(name)]; !ok {
			delete(output.Body.Advisories, name)
		}
	}
	return output, nil
}
