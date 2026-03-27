package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListProjects(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	env.SeedProject(t, "org-1", "Project A")
	env.SeedProject(t, "org-1", "Project B")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/organizations/org-1/projects", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var projects []api.Project
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&projects))
	assert.Len(t, projects, 2)
}

func TestListProjects_NotMember(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	// Create org without user-1 as member
	_, err := env.Pool.Exec(t.Context(), `INSERT INTO organization (id, name, slug, created_at) VALUES ('org-other', 'Other Org', 'other-org', NOW())`)
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/organizations/org-other/projects", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestCreateProject(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")

	body, _ := json.Marshal(api.CreateProjectRequest{
		Name:        "New Project",
		Description: strPtr("A test project"),
	})

	req, _ := http.NewRequest("POST", env.Server.URL+"/api/organizations/org-1/projects", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotNil(t, result["id"])
}

func TestUpdateProject(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	projectID := env.SeedProject(t, "org-1", "Old Name")

	newName := "Updated Name"
	body, _ := json.Marshal(api.UpdateProjectRequest{
		Name: &newName,
	})

	req, _ := http.NewRequest("PATCH", fmt.Sprintf("%s/api/organizations/org-1/projects/%d", env.Server.URL, projectID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestDeleteProject(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	projectID := env.SeedProject(t, "org-1", "To Delete")

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/organizations/org-1/projects/%d", env.Server.URL, projectID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify it's gone
	req, _ = http.NewRequest("GET", env.Server.URL+"/api/organizations/org-1/projects", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var projects []api.Project
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&projects))
	assert.Empty(t, projects)
}

func TestDeleteProject_WithShops_Fails(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	projectID := env.SeedProject(t, "org-1", "Has Shops")
	env.SeedShop(t, "org-1", projectID, "Shop", "https://shop.example.com")

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/organizations/org-1/projects/%d", env.Server.URL, projectID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func strPtr(s string) *string {
	return &s
}
