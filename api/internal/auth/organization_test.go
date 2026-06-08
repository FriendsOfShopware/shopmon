package auth_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// signUp registers a new user and returns the session token from the JSON response.
func signUp(t *testing.T, serverURL, email, password, name string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"email": email, "password": password, "name": name,
	})
	resp, err := http.Post(serverURL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	token, ok := result["token"].(string)
	require.True(t, ok, "response must contain a token string")
	require.NotEmpty(t, token)
	return token
}

// authRequest performs an authenticated HTTP request with a JSON body.
func authRequest(t *testing.T, method, url string, token string, body interface{}) *http.Response {
	t.Helper()
	var reqBody *bytes.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewReader(jsonBody)
	} else {
		reqBody = bytes.NewReader(nil)
	}
	req, _ := http.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// authPost performs an authenticated POST request with a JSON body.
func authPost(t *testing.T, serverURL, path string, token string, body interface{}) *http.Response {
	t.Helper()
	return authRequest(t, "POST", serverURL+path, token, body)
}

// authGet performs an authenticated GET request.
func authGet(t *testing.T, serverURL, path string, token string) *http.Response {
	t.Helper()
	return authRequest(t, "GET", serverURL+path, token, nil)
}

// authPatch performs an authenticated PATCH request with a JSON body.
func authPatch(t *testing.T, serverURL, path string, token string, body interface{}) *http.Response {
	t.Helper()
	return authRequest(t, "PATCH", serverURL+path, token, body)
}

// authDelete performs an authenticated DELETE request.
func authDelete(t *testing.T, serverURL, path string, token string) *http.Response {
	t.Helper()
	return authRequest(t, "DELETE", serverURL+path, token, nil)
}

// decodeJSON decodes the response body into a generic map.
func decodeJSON(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	return result
}

// createOrg is a helper that creates an organization and returns the parsed response.
func createOrg(t *testing.T, serverURL string, cookie string, name string) map[string]interface{} {
	t.Helper()
	resp := authPost(t, serverURL, "/api/auth/organizations", cookie, map[string]string{"name": name})
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	return decodeJSON(t, resp)
}

// inviteAndAccept invites userB by email and accepts the invitation with userB's token.
// Returns the invitation ID.
func inviteAndAccept(t *testing.T, serverURL string, ownerCookie string, orgID, inviteeEmail string, inviteeCookie string) string {
	t.Helper()

	// Invite
	resp := authPost(t, serverURL, fmt.Sprintf("/api/auth/organizations/%s/invitations", orgID), ownerCookie, map[string]string{
		"email": inviteeEmail,
		"role":  "member",
	})
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	inviteResult := decodeJSON(t, resp)
	invitationID := inviteResult["id"].(string)

	// Accept
	resp2 := authPost(t, serverURL, fmt.Sprintf("/api/auth/invitations/%s/accept", invitationID), inviteeCookie, nil)
	defer func() { _ = resp2.Body.Close() }()
	require.Equal(t, http.StatusOK, resp2.StatusCode)

	return invitationID
}

func TestOrgCreate(t *testing.T) {
	env := testutil.Setup(t)
	cookie := signUp(t, env.Server.URL, "alice@example.com", "password123", "Alice")

	org := createOrg(t, env.Server.URL, cookie, "My Organization")

	assert.NotEmpty(t, org["id"])
	assert.Equal(t, "My Organization", org["name"])
}

func TestOrgCreate_Unauthenticated(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{"name": "Test Org"})
	resp, err := http.Post(env.Server.URL+"/api/auth/organizations", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestOrgUpdate(t *testing.T) {
	env := testutil.Setup(t)
	cookie := signUp(t, env.Server.URL, "alice@example.com", "password123", "Alice")

	org := createOrg(t, env.Server.URL, cookie, "Original Name")
	orgID := org["id"].(string)

	newName := "Updated Name"
	resp := authPatch(t, env.Server.URL, fmt.Sprintf("/api/auth/organizations/%s", orgID), cookie, map[string]interface{}{
		"name": newName,
	})
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	result := decodeJSON(t, resp)
	assert.Equal(t, "ok", result["status"])
}

func TestOrgDelete(t *testing.T) {
	env := testutil.Setup(t)
	cookie := signUp(t, env.Server.URL, "alice@example.com", "password123", "Alice")

	org := createOrg(t, env.Server.URL, cookie, "To Delete")
	orgID := org["id"].(string)

	resp := authDelete(t, env.Server.URL, fmt.Sprintf("/api/auth/organizations/%s", orgID), cookie)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	result := decodeJSON(t, resp)
	assert.Equal(t, "ok", result["status"])
}

func TestOrgDelete_NonOwner(t *testing.T) {
	env := testutil.Setup(t)
	ownerCookie := signUp(t, env.Server.URL, "owner@example.com", "password123", "Owner")
	memberCookie := signUp(t, env.Server.URL, "member@example.com", "password123", "Member")

	org := createOrg(t, env.Server.URL, ownerCookie, "Owner Org")
	orgID := org["id"].(string)

	inviteAndAccept(t, env.Server.URL, ownerCookie, orgID, "member@example.com", memberCookie)

	// Member tries to delete — should be forbidden
	resp := authDelete(t, env.Server.URL, fmt.Sprintf("/api/auth/organizations/%s", orgID), memberCookie)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestOrgInviteAndAccept(t *testing.T) {
	env := testutil.Setup(t)
	ownerCookie := signUp(t, env.Server.URL, "owner@example.com", "password123", "Owner")
	memberCookie := signUp(t, env.Server.URL, "member@example.com", "password123", "Member")

	org := createOrg(t, env.Server.URL, ownerCookie, "Invite Org")
	orgID := org["id"].(string)

	inviteAndAccept(t, env.Server.URL, ownerCookie, orgID, "member@example.com", memberCookie)

	// Verify membership by listing members
	resp := authGet(t, env.Server.URL, fmt.Sprintf("/api/auth/organizations/%s/members", orgID), ownerCookie)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var members []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&members))
	assert.Len(t, members, 2)
}

func TestOrgInviteAndReject(t *testing.T) {
	env := testutil.Setup(t)
	ownerCookie := signUp(t, env.Server.URL, "owner@example.com", "password123", "Owner")
	memberCookie := signUp(t, env.Server.URL, "member@example.com", "password123", "Member")

	org := createOrg(t, env.Server.URL, ownerCookie, "Reject Org")
	orgID := org["id"].(string)

	// Invite
	resp := authPost(t, env.Server.URL, fmt.Sprintf("/api/auth/organizations/%s/invitations", orgID), ownerCookie, map[string]string{
		"email": "member@example.com",
		"role":  "member",
	})
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	inviteResult := decodeJSON(t, resp)
	invitationID := inviteResult["id"].(string)

	// Reject
	resp2 := authPost(t, env.Server.URL, fmt.Sprintf("/api/auth/invitations/%s/reject", invitationID), memberCookie, nil)
	defer func() { _ = resp2.Body.Close() }()
	require.Equal(t, http.StatusOK, resp2.StatusCode)

	// Verify the member is NOT in the org
	resp3 := authGet(t, env.Server.URL, fmt.Sprintf("/api/auth/organizations/%s/members", orgID), ownerCookie)
	defer func() { _ = resp3.Body.Close() }()
	require.Equal(t, http.StatusOK, resp3.StatusCode)

	var members []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp3.Body).Decode(&members))
	assert.Len(t, members, 1) // only the owner
}

func TestOrgRemoveMember(t *testing.T) {
	env := testutil.Setup(t)
	ownerCookie := signUp(t, env.Server.URL, "owner@example.com", "password123", "Owner")
	memberCookie := signUp(t, env.Server.URL, "member@example.com", "password123", "Member")

	org := createOrg(t, env.Server.URL, ownerCookie, "Remove Org")
	orgID := org["id"].(string)

	inviteAndAccept(t, env.Server.URL, ownerCookie, orgID, "member@example.com", memberCookie)

	// Get member's user ID from member list
	resp := authGet(t, env.Server.URL, fmt.Sprintf("/api/auth/organizations/%s/members", orgID), ownerCookie)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var members []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&members))
	require.Len(t, members, 2)

	// Find the member (not the owner)
	var memberUserID string
	for _, m := range members {
		if m["role"] == "member" {
			memberUserID = m["userId"].(string)
			break
		}
	}
	require.NotEmpty(t, memberUserID)

	// Remove the member
	resp2 := authDelete(t, env.Server.URL, fmt.Sprintf("/api/auth/organizations/%s/members/%s", orgID, memberUserID), ownerCookie)
	defer func() { _ = resp2.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	// Verify member is removed
	resp3 := authGet(t, env.Server.URL, fmt.Sprintf("/api/auth/organizations/%s/members", orgID), ownerCookie)
	defer func() { _ = resp3.Body.Close() }()
	require.Equal(t, http.StatusOK, resp3.StatusCode)

	var remaining []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp3.Body).Decode(&remaining))
	assert.Len(t, remaining, 1)
}

func TestOrgLeave(t *testing.T) {
	env := testutil.Setup(t)
	ownerCookie := signUp(t, env.Server.URL, "owner@example.com", "password123", "Owner")
	memberCookie := signUp(t, env.Server.URL, "member@example.com", "password123", "Member")

	org := createOrg(t, env.Server.URL, ownerCookie, "Leave Org")
	orgID := org["id"].(string)

	inviteAndAccept(t, env.Server.URL, ownerCookie, orgID, "member@example.com", memberCookie)

	// Member leaves
	resp := authPost(t, env.Server.URL, fmt.Sprintf("/api/auth/organizations/%s/leave", orgID), memberCookie, nil)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify only owner remains
	resp2 := authGet(t, env.Server.URL, fmt.Sprintf("/api/auth/organizations/%s/members", orgID), ownerCookie)
	defer func() { _ = resp2.Body.Close() }()
	require.Equal(t, http.StatusOK, resp2.StatusCode)

	var members []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&members))
	assert.Len(t, members, 1)
	assert.Equal(t, "owner", members[0]["role"])
}

func TestOrgLeave_OnlyOwner(t *testing.T) {
	env := testutil.Setup(t)
	cookie := signUp(t, env.Server.URL, "owner@example.com", "password123", "Owner")

	org := createOrg(t, env.Server.URL, cookie, "Solo Org")
	orgID := org["id"].(string)

	resp := authPost(t, env.Server.URL, fmt.Sprintf("/api/auth/organizations/%s/leave", orgID), cookie, nil)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestOrgSetRole(t *testing.T) {
	env := testutil.Setup(t)
	ownerCookie := signUp(t, env.Server.URL, "owner@example.com", "password123", "Owner")
	memberCookie := signUp(t, env.Server.URL, "member@example.com", "password123", "Member")

	org := createOrg(t, env.Server.URL, ownerCookie, "Role Org")
	orgID := org["id"].(string)

	inviteAndAccept(t, env.Server.URL, ownerCookie, orgID, "member@example.com", memberCookie)

	// Get member's user ID
	resp := authGet(t, env.Server.URL, fmt.Sprintf("/api/auth/organizations/%s/members", orgID), ownerCookie)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var members []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&members))

	var memberUserID string
	for _, m := range members {
		if m["role"] == "member" {
			memberUserID = m["userId"].(string)
			break
		}
	}
	require.NotEmpty(t, memberUserID)

	// Change role to admin
	resp2 := authPatch(t, env.Server.URL, fmt.Sprintf("/api/auth/organizations/%s/members/%s", orgID, memberUserID), ownerCookie, map[string]string{
		"role": "admin",
	})
	defer func() { _ = resp2.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	// Verify role changed
	resp3 := authGet(t, env.Server.URL, fmt.Sprintf("/api/auth/organizations/%s/members", orgID), ownerCookie)
	defer func() { _ = resp3.Body.Close() }()
	require.Equal(t, http.StatusOK, resp3.StatusCode)

	var updated []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp3.Body).Decode(&updated))

	for _, m := range updated {
		if m["userId"] == memberUserID {
			assert.Equal(t, "admin", m["role"])
		}
	}
}

func TestOrgListMembers(t *testing.T) {
	env := testutil.Setup(t)
	cookie := signUp(t, env.Server.URL, "alice@example.com", "password123", "Alice")

	org := createOrg(t, env.Server.URL, cookie, "List Org")
	orgID := org["id"].(string)

	resp := authGet(t, env.Server.URL, fmt.Sprintf("/api/auth/organizations/%s/members", orgID), cookie)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var members []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&members))
	require.Len(t, members, 1)
	assert.Equal(t, "owner", members[0]["role"])
	assert.Equal(t, "Alice", members[0]["name"])
	assert.Equal(t, "alice@example.com", members[0]["email"])
}
