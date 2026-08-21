package main

import (
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/apirouter"
	"github.com/friendsofshopware/shopmon/api/internal/auth"
	"github.com/friendsofshopware/shopmon/api/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"
)

func openapiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "openapi",
		Short: "Print the OpenAPI 3.0.3 specification as YAML",
		RunE: func(cmd *cobra.Command, args []string) error {
			router := chi.NewRouter()
			var specAPI = apirouter.Mount(router, &handler.Handler{}, &auth.AuthHandler{}, apirouter.Options{})
			body, err := specAPI.OpenAPI().DowngradeYAML()
			if err != nil {
				return fmt.Errorf("encode openapi: %w", err)
			}
			_, err = cmd.OutOrStdout().Write(append(body, '\n'))
			return err
		},
	}
}
