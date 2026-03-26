package checker

import (
	"context"
	"fmt"
	"time"
)

func checkTasks(_ context.Context, input Input, output *Output) {
	for _, task := range input.ScheduledTasks {
		if task.Status == "inactive" {
			continue
		}
		if task.RunInterval <= 3600 {
			continue
		}
		if isTaskOverdue(task) {
			output.Warning(
				fmt.Sprintf("task.%s", task.Name),
				fmt.Sprintf("Scheduled task '%s' is overdue", task.Name),
				"Shopware", "")
		}
	}

	output.Success("task.all", "All scheduled tasks are running correctly.", "Shopware", "")
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
