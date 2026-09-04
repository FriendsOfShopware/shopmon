package checker

import (
	"context"
	"fmt"
	"time"
)

func checkTasks(_ context.Context, input Input, output *Output) {
	// An unfetched task list is indistinguishable from "no tasks" here, and
	// would resolve to the all-running success below.
	if input.Missing.ScheduledTasks {
		output.MarkUnavailable(SourceShopware, prefixScheduledTask)
		return
	}

	hasWarning := false
	for _, task := range input.ScheduledTasks {
		if task.Status == "inactive" {
			continue
		}
		if task.RunInterval <= 3600 {
			continue
		}
		if isTaskOverdue(task, input.TaskGrace) {
			hasWarning = true
			output.Warning(
				fmt.Sprintf("%s%s", prefixScheduledTask, task.Name),
				"check.task.overdue",
				map[string]any{"name": task.Name},
				SourceShopware, "")
		}
	}

	if !hasWarning {
		output.Success(prefixScheduledTask+"all", "check.task.allRunning", nil, SourceShopware, "")
	}
}

// isTaskOverdue reports whether a task is late by more than the grace period.
// Shopware writes next_execution_time when it queues the task, so the worker
// picking it up a moment later is normal operation — only a task still unrun
// after the grace period has actually stalled.
func isTaskOverdue(task ScheduledTask, grace time.Duration) bool {
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
			return time.Now().After(t.Add(grace))
		}
	}
	return false
}
