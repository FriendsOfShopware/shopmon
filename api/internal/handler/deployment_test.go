package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDeployments_Empty(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	projectID := env.SeedProject(t, "org-1", "Test Project")
	shopID := env.SeedShop(t, "org-1", projectID, "My Shop", "https://shop.example.com")

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/shops/%d/deployments", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var deployments []api.Deployment
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&deployments))
	assert.Empty(t, deployments)
}

func TestGetDeployments_WithData(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	projectID := env.SeedProject(t, "org-1", "Test Project")
	shopID := env.SeedShop(t, "org-1", projectID, "My Shop", "https://shop.example.com")

	// Seed a deployment
	now := time.Now()
	_, err := env.Pool.Exec(t.Context(), `
		INSERT INTO deployment (shop_id, name, command, return_code, start_date, end_date, execution_time, created_at)
		VALUES ($1, 'Deploy v1', 'bin/console deploy', 0, $2, $2, 12.5, $2)
	`, shopID, now)
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/shops/%d/deployments", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var deployments []api.Deployment
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&deployments))
	require.Len(t, deployments, 1)
	assert.Equal(t, "bin/console deploy", deployments[0].Command)
	assert.Equal(t, 0, deployments[0].ReturnCode)
}

func TestGetDeployment(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	projectID := env.SeedProject(t, "org-1", "Test Project")
	shopID := env.SeedShop(t, "org-1", projectID, "My Shop", "https://shop.example.com")

	now := time.Now()
	var deploymentID int
	err := env.Pool.QueryRow(t.Context(), `
		INSERT INTO deployment (shop_id, name, command, return_code, start_date, end_date, execution_time, created_at)
		VALUES ($1, 'Deploy v2', 'bin/console deploy', 0, $2, $2, 5.0, $2)
		RETURNING id
	`, shopID, now).Scan(&deploymentID)
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/shops/%d/deployments/%d", env.Server.URL, shopID, deploymentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var detail api.DeploymentDetail
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&detail))
	assert.Equal(t, deploymentID, detail.Id)
	assert.Equal(t, "bin/console deploy", detail.Command)
	// Output should be nil since we have no S3
	assert.Nil(t, detail.Output)
}

func TestDeleteDeployment(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	projectID := env.SeedProject(t, "org-1", "Test Project")
	shopID := env.SeedShop(t, "org-1", projectID, "My Shop", "https://shop.example.com")

	now := time.Now()
	var deploymentID int
	err := env.Pool.QueryRow(t.Context(), `
		INSERT INTO deployment (shop_id, name, command, return_code, start_date, end_date, execution_time, created_at)
		VALUES ($1, 'Deploy v3', 'bin/console deploy', 1, $2, $2, 3.0, $2)
		RETURNING id
	`, shopID, now).Scan(&deploymentID)
	require.NoError(t, err)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/shops/%d/deployments/%d", env.Server.URL, shopID, deploymentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify deleted
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/shops/%d/deployments/%d", env.Server.URL, shopID, deploymentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCreateCliDeployment(t *testing.T) {
	env := testutil.Setup(t)
	// We need a user to own the org, but CLI uses API key auth
	env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	projectID := env.SeedProject(t, "org-1", "Test Project")
	shopID := env.SeedShop(t, "org-1", projectID, "My Shop", "https://shop.example.com")

	// Create API key directly in DB
	apiKeyToken := "shm_test_api_key_token_1234"
	_, err := env.Pool.Exec(t.Context(), `
		INSERT INTO project_api_key (id, project_id, name, token, scopes, created_at)
		VALUES ('apikey-1', $1, 'CI Key', $2, '["deployments"]'::jsonb, NOW())
	`, projectID, apiKeyToken)
	require.NoError(t, err)

	now := time.Now()
	name := "Production Deploy"
	body, _ := json.Marshal(api.CreateCliDeploymentRequest{
		ShopId:        shopID,
		Command:       "bin/console deploy:run",
		ReturnCode:    0,
		StartDate:     now.Add(-time.Minute),
		EndDate:       now,
		ExecutionTime: 60.0,
		Name:          &name,
	})

	req, _ := http.NewRequest("POST", env.Server.URL+"/api/cli/deployments", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+apiKeyToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result api.CreateCliDeploymentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.True(t, result.Success)
	assert.Equal(t, "Production Deploy", result.Name)
	assert.Greater(t, result.DeploymentId, 0)
}

func TestCreateCliDeployment_InvalidToken(t *testing.T) {
	env := testutil.Setup(t)

	now := time.Now()
	body, _ := json.Marshal(api.CreateCliDeploymentRequest{
		ShopId:        1,
		Command:       "deploy",
		ReturnCode:    0,
		StartDate:     now,
		EndDate:       now,
		ExecutionTime: 1.0,
	})

	req, _ := http.NewRequest("POST", env.Server.URL+"/api/cli/deployments", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer invalid-token")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCreateCliDeployment_WrongProject(t *testing.T) {
	env := testutil.Setup(t)
	env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	projectID1 := env.SeedProject(t, "org-1", "Project 1")
	projectID2 := env.SeedProject(t, "org-1", "Project 2")
	// Shop belongs to project 2
	shopID := env.SeedShop(t, "org-1", projectID2, "My Shop", "https://shop.example.com")

	// API key belongs to project 1
	apiKeyToken := "shm_wrong_project_key"
	_, err := env.Pool.Exec(t.Context(), `
		INSERT INTO project_api_key (id, project_id, name, token, scopes, created_at)
		VALUES ('apikey-wp', $1, 'Wrong Project Key', $2, '["deployments"]'::jsonb, NOW())
	`, projectID1, apiKeyToken)
	require.NoError(t, err)

	now := time.Now()
	body, _ := json.Marshal(api.CreateCliDeploymentRequest{
		ShopId:        shopID,
		Command:       "deploy",
		ReturnCode:    0,
		StartDate:     now,
		EndDate:       now,
		ExecutionTime: 1.0,
	})

	req, _ := http.NewRequest("POST", env.Server.URL+"/api/cli/deployments", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+apiKeyToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetDeployments_NotMember(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedUser(t, "user-2", "Other User", "other@example.com", "user")
	env.SeedOrganization(t, "org-2", "Other Org", "other-org", "user-2")
	projectID := env.SeedProject(t, "org-2", "Other Project")
	shopID := env.SeedShop(t, "org-2", projectID, "Other Shop", "https://other.example.com")

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/shops/%d/deployments", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
