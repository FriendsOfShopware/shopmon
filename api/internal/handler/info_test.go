package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckExtensionCompatibility(t *testing.T) {
	// Mock the Shopware API
	mockShopwareAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/swplatform/autoupdate" {
			http.NotFound(w, r)
			return
		}
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Decode and verify the request body
		var req struct {
			ShopwareVersion       string `json:"shopwareVersion"`
			FutureShopwareVersion string `json:"futureShopwareVersion"`
			Plugins               []struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"plugins"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "6.5.0.0", req.ShopwareVersion)
		assert.Equal(t, "6.6.0.0", req.FutureShopwareVersion)
		assert.Len(t, req.Plugins, 1)
		assert.Equal(t, "SwagPayPal", req.Plugins[0].Name)

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]api.ExtensionCompatibilityResult{
			{
				Name:  "SwagPayPal",
				Label: "PayPal",
				Status: struct {
					Label string `json:"label"`
					Name  string `json:"name"`
					Type  string `json:"type"`
				}{
					Name:  "compatible",
					Label: "Compatible",
					Type:  "green",
				},
			},
		})
	}))
	defer mockShopwareAPI.Close()

	env := testutil.Setup(t, func(cfg *config.Config) {
		cfg.ShopwareAPIURL = mockShopwareAPI.URL
	})

	body, _ := json.Marshal(api.ExtensionCompatibilityRequest{
		CurrentVersion: "6.5.0.0",
		FutureVersion:  "6.6.0.0",
		Extensions: []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}{
			{Name: "SwagPayPal", Version: "5.0.0"},
		},
	})

	req, _ := http.NewRequest("POST", env.Server.URL+"/api/info/extension-compatibility", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var results []api.ExtensionCompatibilityResult
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&results))
	require.Len(t, results, 1)
	assert.Equal(t, "SwagPayPal", results[0].Name)
	assert.Equal(t, "PayPal", results[0].Label)
	assert.Equal(t, "compatible", results[0].Status.Name)
	assert.Equal(t, "green", results[0].Status.Type)
}

func TestCheckExtensionCompatibility_APIError(t *testing.T) {
	mockShopwareAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer mockShopwareAPI.Close()

	env := testutil.Setup(t, func(cfg *config.Config) {
		cfg.ShopwareAPIURL = mockShopwareAPI.URL
	})

	body, _ := json.Marshal(api.ExtensionCompatibilityRequest{
		CurrentVersion: "6.5.0.0",
		FutureVersion:  "6.6.0.0",
		Extensions: []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}{
			{Name: "SwagPayPal", Version: "5.0.0"},
		},
	})

	req, _ := http.NewRequest("POST", env.Server.URL+"/api/info/extension-compatibility", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

func TestGetShopwareVersions(t *testing.T) {
	mockVersionsAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"6.6.0.0": []string{"8.2", "8.3"},
			"6.5.8.0": []string{"8.1", "8.2", "8.3"},
		})
	}))
	defer mockVersionsAPI.Close()

	env := testutil.Setup(t, func(cfg *config.Config) {
		cfg.ShopwareVersionsURL = mockVersionsAPI.URL
	})

	resp, err := http.Get(env.Server.URL + "/api/info/shopware-versions")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var versions map[string][]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&versions))
	assert.Contains(t, versions, "6.6.0.0")
	assert.Contains(t, versions, "6.5.8.0")
	assert.Equal(t, []string{"8.2", "8.3"}, versions["6.6.0.0"])
}

func TestCheckExtensionCompatibility_InvalidBody(t *testing.T) {
	env := testutil.Setup(t)

	req, _ := http.NewRequest("POST", env.Server.URL+"/api/info/extension-compatibility", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
