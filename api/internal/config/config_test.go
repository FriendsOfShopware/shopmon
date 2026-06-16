package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeploymentScrapeDelay(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want time.Duration
	}{
		{name: "default when unset", env: "", want: 5 * time.Minute},
		{name: "custom duration", env: "2m", want: 2 * time.Minute},
		{name: "zero disables delay", env: "0s", want: 0},
		{name: "invalid falls back to default", env: "not-a-duration", want: 5 * time.Minute},
		{name: "negative falls back to default", env: "-1m", want: 5 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.env == "" {
				t.Setenv("DEPLOYMENT_SCRAPE_DELAY", "")
			} else {
				t.Setenv("DEPLOYMENT_SCRAPE_DELAY", tt.env)
			}

			cfg := Load()
			assert.Equal(t, tt.want, cfg.DeploymentScrapeDelay)
		})
	}
}
