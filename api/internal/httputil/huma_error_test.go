package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewErrorMapsUnprocessableToBadRequest(t *testing.T) {
	err := huma.NewError(http.StatusUnprocessableEntity, "validation failed")
	se, ok := err.(*JSONStatusError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, se.GetStatus())
	assert.Equal(t, "validation failed", se.Error())
}

func TestNewErrorJSONShape(t *testing.T) {
	err := huma.NewError(http.StatusUnauthorized, "unauthorized", &huma.ErrorDetail{
		Message:  "ignored",
		Location: "header.Authorization",
	})

	se, ok := err.(*JSONStatusError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, se.GetStatus())
	assert.Equal(t, "unauthorized", se.Error())
	assert.Equal(t, "application/json", se.ContentType("application/problem+json"))

	body, marshalErr := json.Marshal(se)
	require.NoError(t, marshalErr)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(body, &raw))
	assert.Equal(t, map[string]any{"message": "unauthorized"}, raw)
}

func TestWriteStatusError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteStatusError(rec, huma.Error403Forbidden(MsgAdminRequired))

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var raw map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &raw))
	assert.Equal(t, map[string]any{"message": MsgAdminRequired}, raw)
}
