package checker

import (
	"context"
	"strings"
)

var validEnvironments = map[string]bool{
	"production": true,
	"staging":    true,
	"prod":       true,
	"stage":      true,
}

func checkEnv(_ context.Context, input Input, output *Output) {
	// An unfetched cache info would look like an empty (and therefore invalid)
	// environment name, which is a fabricated warning rather than a finding.
	if input.Missing.CacheInfo {
		output.MarkUnavailable(SourceShopware, prefixEnv)
		return
	}

	env := strings.ToLower(input.CacheInfo.Environment)
	params := map[string]any{"environment": input.CacheInfo.Environment}
	if !validEnvironments[env] {
		output.Warning(prefixEnv, "check.env.invalid", params, SourceShopware, "")
	} else {
		output.Success(prefixEnv, "check.env.valid", params, SourceShopware, "")
	}
}
