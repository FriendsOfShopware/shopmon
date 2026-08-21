package auth

import "github.com/danielgtaylor/huma/v2"

// Register attaches all auth operations to the Huma API.
func Register(api huma.API, h *AuthHandler) {
	registerCredentials(api, h)
	registerGithub(api, h)
	registerPasskeys(api, h)
	registerSSOAuth(api, h)
	registerCodes(api, h)
	registerOrganizations(api, h)
	registerAccount(api, h)
	registerAdmin(api, h)
}
