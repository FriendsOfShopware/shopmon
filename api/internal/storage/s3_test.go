package storage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteDeploymentOutputDeletesExpectedKey(t *testing.T) {
	var gotMethod string
	var gotPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	storage, err := NewS3Storage(server.URL, "access-key", "secret-key", "shopmon", "us-east-1")
	require.NoError(t, err)

	err = storage.DeleteDeploymentOutput(t.Context(), 8)
	require.NoError(t, err)

	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/shopmon/deployments/8/output.zst", gotPath)
}

func TestDeleteDeploymentOutputTimesOut(t *testing.T) {
	requestStarted := make(chan struct{})
	var markStarted sync.Once

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		markStarted.Do(func() { close(requestStarted) })
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)

	storage, err := NewS3Storage(server.URL, "access-key", "secret-key", "shopmon", "us-east-1")
	require.NoError(t, err)
	storage.deleteTimeout = 100 * time.Millisecond

	start := time.Now()
	err = storage.DeleteDeploymentOutput(context.Background(), 8)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, elapsed, 2*time.Second)

	var requestReceived bool
	select {
	case <-requestStarted:
		requestReceived = true
	default:
	}
	assert.True(t, requestReceived, "S3 server did not receive delete request")
}
