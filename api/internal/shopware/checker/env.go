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
	if !validEnvironments[env] {
		output.Warning("shopware.env",
			"The environment is set to '"+input.CacheInfo.Environment+"'. It should be set to 'production' or 'staging'.",
			"Shopware", "")
	} else {
		output.Success("shopware.env",
			"Environment is correctly set to '"+input.CacheInfo.Environment+"'.",
			"Shopware", "")
	}
}
