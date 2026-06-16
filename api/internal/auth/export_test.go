package auth

// SetGithubBaseURLsForTest overrides the GitHub OAuth and API base URLs and
// returns a function that restores the previous values. It exists only for
// tests (this file is compiled solely for the test binary) so that the external
// auth_test package can point the GitHub flow at an httptest mock server.
func SetGithubBaseURLsForTest(oauthBaseURL, apiBaseURL string) (restore func()) {
	prevOAuth, prevAPI := githubOAuthBaseURL, githubAPIBaseURL
	githubOAuthBaseURL = oauthBaseURL
	githubAPIBaseURL = apiBaseURL
	return func() {
		githubOAuthBaseURL = prevOAuth
		githubAPIBaseURL = prevAPI
	}
}
