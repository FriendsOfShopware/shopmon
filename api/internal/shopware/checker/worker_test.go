package checker

import (
	"context"
	"testing"
)

func TestCheckWorker(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		wantLevel   Status
		wantMessage string
	}{
		{
			name:        "admin worker enabled is warning",
			enabled:     true,
			wantLevel:   StatusYellow,
			wantMessage: "The admin worker is enabled. This can cause performance issues. Consider using the CLI worker instead.",
		},
		{
			name:        "admin worker disabled is success",
			enabled:     false,
			wantLevel:   StatusGreen,
			wantMessage: "Admin worker is disabled. CLI worker is being used.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config ShopConfig
			config.AdminWorker.EnableAdminWorker = tt.enabled

			output := NewOutput(nil)
			checkWorker(context.Background(), Input{Config: config}, output)

			checks := output.Result().Checks
			if len(checks) != 1 {
				t.Fatalf("expected 1 check, got %d", len(checks))
			}

			check := checks[0]
			if check.ID != "admin.worker" {
				t.Errorf("ID = %q, want %q", check.ID, "admin.worker")
			}
			if check.Level != tt.wantLevel {
				t.Errorf("Level = %q, want %q", check.Level, tt.wantLevel)
			}
			if check.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", check.Message, tt.wantMessage)
			}
			if check.Source != "Shopware" {
				t.Errorf("Source = %q, want %q", check.Source, "Shopware")
			}
		})
	}
}
