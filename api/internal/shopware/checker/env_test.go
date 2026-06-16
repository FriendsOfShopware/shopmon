package checker

import (
	"context"
	"testing"
)

func TestCheckEnv(t *testing.T) {
	tests := []struct {
		name      string
		env       string
		wantLevel Status
	}{
		{name: "production is success", env: "production", wantLevel: StatusGreen},
		{name: "staging is success", env: "staging", wantLevel: StatusGreen},
		{name: "prod is success", env: "prod", wantLevel: StatusGreen},
		{name: "stage is success", env: "stage", wantLevel: StatusGreen},
		{name: "uppercase production is success", env: "PRODUCTION", wantLevel: StatusGreen},
		{name: "mixed case Production is success", env: "Production", wantLevel: StatusGreen},
		{name: "dev is warning", env: "dev", wantLevel: StatusYellow},
		{name: "empty is warning", env: "", wantLevel: StatusYellow},
		{name: "test is warning", env: "test", wantLevel: StatusYellow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewOutput(nil)
			checkEnv(context.Background(), Input{CacheInfo: CacheInfo{Environment: tt.env}}, output)

			checks := output.Result().Checks
			if len(checks) != 1 {
				t.Fatalf("expected 1 check, got %d", len(checks))
			}

			check := checks[0]
			if check.ID != "shopware.env" {
				t.Errorf("ID = %q, want %q", check.ID, "shopware.env")
			}
			if check.Level != tt.wantLevel {
				t.Errorf("Level = %q, want %q", check.Level, tt.wantLevel)
			}
			if check.Source != "Shopware" {
				t.Errorf("Source = %q, want %q", check.Source, "Shopware")
			}
		})
	}
}
