package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNotifications(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedNotification(t, "user-1", "test-key-1", "warning", "Test Alert", "Something happened")
	env.SeedNotification(t, "user-1", "test-key-2", "error", "Critical Alert", "Something broke")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var notifications []api.Notification
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&notifications))
	assert.Len(t, notifications, 2)
}

func TestGetNotifications_Unauthenticated(t *testing.T) {
	env := testutil.Setup(t)

	resp, err := http.Get(env.Server.URL + "/api/notifications")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestDeleteNotification(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	notifID := env.SeedNotification(t, "user-1", "test-key-1", "warning", "Test Alert", "Something happened")

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/notifications/%d", env.Server.URL, notifID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify it's gone
	req, _ = http.NewRequest("GET", env.Server.URL+"/api/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var notifications []api.Notification
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&notifications))
	assert.Empty(t, notifications)
}

func TestDeleteAllNotifications(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedNotification(t, "user-1", "test-key-1", "warning", "Alert 1", "Message 1")
	env.SeedNotification(t, "user-1", "test-key-2", "error", "Alert 2", "Message 2")

	req, _ := http.NewRequest("DELETE", env.Server.URL+"/api/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify all gone
	req, _ = http.NewRequest("GET", env.Server.URL+"/api/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var notifications []api.Notification
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&notifications))
	assert.Empty(t, notifications)
}

func TestMarkNotificationsRead(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedNotification(t, "user-1", "test-key-1", "warning", "Alert 1", "Message 1")

	req, _ := http.NewRequest("POST", env.Server.URL+"/api/notifications/mark-read", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify all are read
	req, _ = http.NewRequest("GET", env.Server.URL+"/api/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var notifications []api.Notification
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&notifications))
	require.Len(t, notifications, 1)
	assert.True(t, notifications[0].Read)
}
