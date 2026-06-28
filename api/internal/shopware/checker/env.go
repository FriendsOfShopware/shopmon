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
	env := strings.ToLower(input.CacheInfo.Environment)
	params := map[string]any{"environment": input.CacheInfo.Environment}
	if !validEnvironments[env] {
		output.Warning("shopware.env", "check.env.invalid", params, "Shopware", "")
	} else {
		output.Success("shopware.env", "check.env.valid", params, "Shopware", "")
	}
}
