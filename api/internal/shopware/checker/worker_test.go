package checker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckWorker(t *testing.T) {
	tests := []struct {
		name      string
		enabled   bool
		wantLevel Status
		wantKey   string
	}{
		{
			name:      "admin worker enabled is warning",
			enabled:   true,
			wantLevel: StatusYellow,
			wantKey:   "check.worker.enabled",
		},
		{
			name:      "admin worker disabled is success",
			enabled:   false,
			wantLevel: StatusGreen,
			wantKey:   "check.worker.disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config ShopConfig
			config.AdminWorker.EnableAdminWorker = tt.enabled

			output := NewOutput(nil)
			checkWorker(context.Background(), Input{Config: config}, output)

			checks := output.Result().Checks
			require.Len(t, checks, 1)

			check := checks[0]
			assert.Equal(t, "admin.worker", check.ID)
			assert.Equal(t, tt.wantLevel, check.Level)
			assert.Equal(t, tt.wantKey, check.MessageKey)
			assert.Equal(t, "Shopware", check.Source)
		})
	}
}
