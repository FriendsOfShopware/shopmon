package checker

import (
	"context"
	"testing"
	"time"
)

func TestCheckTasks(t *testing.T) {
	overdue := time.Now().Add(-time.Hour).Format(time.RFC3339)
	future := time.Now().Add(time.Hour).Format(time.RFC3339)

	tests := []struct {
		name          string
		tasks         []ScheduledTask
		wantSuccess   bool
		wantWarningID string
	}{
		{
			name:        "no tasks emits success",
			tasks:       nil,
			wantSuccess: true,
		},
		{
			name: "overdue active task emits warning without success",
			tasks: []ScheduledTask{
				{
					Name:              "shopware.invalidate_cache",
					Status:            "active",
					RunInterval:       7200,
					NextExecutionTime: &overdue,
				},
			},
			wantSuccess:   false,
			wantWarningID: "task.shopware.invalidate_cache",
		},
		{
			name: "inactive task ignored",
			tasks: []ScheduledTask{
				{
					Name:              "inactive_task",
					Status:            "inactive",
					RunInterval:       7200,
					NextExecutionTime: &overdue,
				},
			},
			wantSuccess: true,
		},
		{
			name: "short interval task ignored",
			tasks: []ScheduledTask{
				{
					Name:              "short_task",
					Status:            "active",
					RunInterval:       3600,
					NextExecutionTime: &overdue,
				},
			},
			wantSuccess: true,
		},
		{
			name: "not overdue task emits success",
			tasks: []ScheduledTask{
				{
					Name:              "future_task",
					Status:            "active",
					RunInterval:       7200,
					NextExecutionTime: &future,
				},
			},
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewOutput(nil)
			checkTasks(context.Background(), Input{ScheduledTasks: tt.tasks}, output)

			var hasSuccess, hasWarning bool
			for _, check := range output.Result().Checks {
				if check.ID == "task.all" && check.Level == StatusGreen {
					hasSuccess = true
				}
				if tt.wantWarningID != "" && check.ID == tt.wantWarningID && check.Level == StatusYellow {
					hasWarning = true
				}
			}

			if hasSuccess != tt.wantSuccess {
				t.Errorf("success emitted = %v, want %v", hasSuccess, tt.wantSuccess)
			}
			if tt.wantWarningID != "" && !hasWarning {
				t.Errorf("expected warning %q not emitted", tt.wantWarningID)
			}
		})
	}
}
