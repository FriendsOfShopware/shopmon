package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type healthStatus struct {
	Status string `json:"status"`
}

type getHealthOutput struct {
	Body healthStatus
}

func registerHealth(api huma.API, h *Handler) {
	huma.Register(api, huma.Operation{
		OperationID: "getHealth",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health check",
		Tags:        []string{"Health"},
		Security:    []map[string][]string{},
	}, func(ctx context.Context, _ *struct{}) (*getHealthOutput, error) {
		if err := h.database.Ping(ctx); err != nil {
			return nil, huma.Error503ServiceUnavailable("database connection failed")
		}
		return &getHealthOutput{Body: healthStatus{Status: "ok"}}, nil
	})
}
