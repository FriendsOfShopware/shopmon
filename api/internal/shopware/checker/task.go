package checker

import (
	"context"
	"fmt"
	"time"
)

func checkTasks(_ context.Context, input Input, output *Output) {
	hasWarning := false
	for _, task := range input.ScheduledTasks {
		if task.Status == "inactive" {
			continue
		}
		if task.RunInterval <= 3600 {
			continue
		}
		if isTaskOverdue(task) {
			hasWarning = true
			output.Warning(
				fmt.Sprintf("task.%s", task.Name),
				"check.task.overdue",
				map[string]any{"name": task.Name},
				"Shopware", "")
		}
	}

	if !hasWarning {
		output.Success("task.all", "check.task.allRunning", nil, "Shopware", "")
	}
}

func isTaskOverdue(task ScheduledTask) bool {
	if task.NextExecutionTime == nil {
		return false
	}
	// Try multiple time formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000+00:00",
		"2006-01-02T15:04:05+00:00",
		"2006-01-02 15:04:05",
	}
	for _, format := range formats {
		t, err := time.Parse(format, *task.NextExecutionTime)
		if err == nil {
			return time.Now().After(t)
		}
	}
	return false
}
