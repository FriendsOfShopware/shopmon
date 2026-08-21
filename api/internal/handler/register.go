package handler

import "github.com/danielgtaylor/huma/v2"

// Register attaches all API operations to the Huma API.
func Register(api huma.API, h *Handler) {
	registerHealth(api, h)
	registerInfo(api, h)
	registerAccount(api, h)
	registerNotifications(api, h)
	registerShops(api, h)
	registerEnvironments(api, h)
	registerAPIKeys(api, h)
	registerPackagesTokens(api, h)
	registerDeployments(api, h)
	registerSSO(api, h)
	registerAdvisories(api, h)
	registerAdmin(api, h)
}
