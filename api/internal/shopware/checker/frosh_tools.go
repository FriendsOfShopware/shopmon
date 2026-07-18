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

		key := c.Snippet
		if key == "" {
			key = c.ID
		}

		id := "frosh." + key
		// messageKey resolves to a catalog entry per snippet. The snippet is also
		// passed as a param so an unknown snippet degrades to its raw name rather
		// than a bare key. current/recommended (when present) drive the generic
		// recommendation suffix appended at render time.
		messageKey := "check.frosh." + key
		params := map[string]any{"snippet": key}
		if c.Current != nil {
			params["current"] = *c.Current
		}
		if c.Recommended != nil {
			params["recommended"] = *c.Recommended
		}

		if ignoredFroshChecks[c.Snippet] {
			output.Success(id, messageKey, params, "FroshTools", c.URL)
			continue
		}

		switch c.State {
		case "STATE_OK":
			output.Success(id, messageKey, params, "FroshTools", c.URL)
		case "STATE_WARNING":
			output.Warning(id, messageKey, params, "FroshTools", c.URL)
		case "STATE_ERROR":
			output.Error(id, messageKey, params, "FroshTools", c.URL)
		}
	}
}
