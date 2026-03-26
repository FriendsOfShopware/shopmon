package checker

import (
	"context"
	"encoding/json"
	"log/slog"
)

type froshToolsCheck struct {
	ID          string  `json:"id"`
	State       string  `json:"state"`
	Snippet     string  `json:"snippet"`
	Current     *string `json:"current,omitempty"`
	Recommended *string `json:"recommended,omitempty"`
	URL         string  `json:"url,omitempty"`
}

// Checks that should not escalate shop status
var ignoredFroshChecks = map[string]bool{
	"scheduledTaskGood":    true,
	"scheduledTaskWarning": true,
	"queuesGood":           true,
	"queuesWarning":        true,
	"prodGood":             true,
	"not-prod":             true,
	"adminWorkerGood":      true,
	"adminWorkerWarning":   true,
}

func checkFroshTools(ctx context.Context, input Input, output *Output) {
	var found bool
	for _, ext := range input.Extensions {
		if ext.Name == "FroshTools" && ext.Active && ext.Installed {
			found = true
			break
		}
	}
	if !found {
		return
	}

	if input.Client == nil {
		return
	}

	// Fetch health status
	healthData, err := input.Client.Get(ctx, "/_action/frosh-tools/health/status")
	if err != nil {
		slog.Warn("failed to fetch FroshTools health status", "error", err)
		return
	}

	var healthChecks []froshToolsCheck
	if err := json.Unmarshal(healthData, &healthChecks); err != nil {
		slog.Warn("failed to parse FroshTools health data", "error", err)
		return
	}

	mapFroshChecks(healthChecks, output)

	// Fetch performance status
	perfData, err := input.Client.Get(ctx, "/_action/frosh-tools/performance/status")
	if err != nil {
		slog.Warn("failed to fetch FroshTools performance status", "error", err)
		return
	}

	var perfChecks []froshToolsCheck
	if err := json.Unmarshal(perfData, &perfChecks); err != nil {
		slog.Warn("failed to parse FroshTools performance data", "error", err)
		return
	}

	mapFroshChecks(perfChecks, output)
}

func mapFroshChecks(checks []froshToolsCheck, output *Output) {
	for _, c := range checks {
		if c.Snippet == "" && c.ID == "" {
			continue
		}

		id := "frosh." + c.Snippet
		if c.Snippet == "" {
			id = "frosh." + c.ID
		}

		message := getFroshMessage(c.Snippet)
		if message == "" {
			message = c.Snippet
		}

		if c.Current != nil && c.Recommended != nil {
			message += " (Current: " + *c.Current + ", Recommended: " + *c.Recommended + ")"
		}

		// Skip checks that shouldn't escalate status
		if ignoredFroshChecks[c.Snippet] {
			switch c.State {
			case "STATE_OK":
				output.Success(id, message, "FroshTools", c.URL)
			default:
				// Add the check but don't let it escalate
				output.Success(id, message, "FroshTools", c.URL)
			}
			continue
		}

		switch c.State {
		case "STATE_OK":
			output.Success(id, message, "FroshTools", c.URL)
		case "STATE_WARNING":
			output.Warning(id, message, "FroshTools", c.URL)
		case "STATE_ERROR":
			output.Error(id, message, "FroshTools", c.URL)
		}
	}
}

func getFroshMessage(snippet string) string {
	messages := map[string]string{
		"phpGood":                  "PHP version is up to date",
		"phpOutdated":              "PHP version is outdated",
		"mysqlGood":                "MySQL version is up to date",
		"mysqlOutdated":            "MySQL version is outdated",
		"opcacheGood":              "OPcache is enabled",
		"opcacheDisabled":          "OPcache is disabled",
		"opcacheNoJit":             "OPcache JIT is not enabled",
		"scheduledTaskGood":        "All scheduled tasks are running",
		"scheduledTaskWarning":     "Some scheduled tasks are overdue",
		"queuesGood":               "Message queue is empty",
		"queuesWarning":            "Message queue has pending messages",
		"prodGood":                 "Environment is set to production",
		"not-prod":                 "Environment is not set to production",
		"adminWorkerGood":          "Admin worker is disabled",
		"adminWorkerWarning":       "Admin worker is enabled",
		"publicFilesystemGood":     "Public filesystem is using a proper adapter",
		"publicFilesystemWarning":  "Public filesystem is using local adapter",
		"privateFilesystemGood":    "Private filesystem is using a proper adapter",
		"privateFilesystemWarning": "Private filesystem is using local adapter",
		"queueConnectionGood":      "Queue connection is properly configured",
		"queueConnectionWarning":   "Queue connection is using sync adapter",
		"sessionConnectionGood":    "Session storage is properly configured",
		"sessionConnectionWarning": "Session storage is using default file adapter",
		"mailerGood":               "Mailer is properly configured",
		"mailerWarning":            "Mailer is using null transport",
		"incrementGood":            "Increment storage is properly configured",
		"incrementWarning":         "Increment storage is using default adapter",
		"numberRangeGood":          "Number range storage is properly configured",
		"numberRangeWarning":       "Number range storage is using default adapter",
		"lockGood":                 "Lock storage is properly configured",
		"lockWarning":              "Lock storage is using default flock adapter",
		"cacheGood":                "Cache adapter is properly configured",
		"cacheWarning":             "Cache is using filesystem adapter",
		"puppeteerGood":            "Puppeteer/Chrome is available",
		"puppeteerWarning":         "Puppeteer/Chrome is not available",
	}
	if msg, ok := messages[snippet]; ok {
		return msg
	}
	return ""
}
