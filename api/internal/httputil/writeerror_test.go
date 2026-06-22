package httputil

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteErrorAuto(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "pgx no rows maps to 404",
			err:         pgx.ErrNoRows,
			wantStatus:  http.StatusNotFound,
			wantMessage: "not found",
		},
		{
			name:        "wrapped pgx no rows maps to 404",
			err:         fmt.Errorf("loading org: %w", pgx.ErrNoRows),
			wantStatus:  http.StatusNotFound,
			wantMessage: "not found",
		},
		{
			name:        "ErrNotFound maps to 404",
			err:         ErrNotFound,
			wantStatus:  http.StatusNotFound,
			wantMessage: "not found",
		},
		{
			name:        "wrapped ErrNotFound maps to 404",
			err:         fmt.Errorf("shop %d: %w", 7, ErrNotFound),
			wantStatus:  http.StatusNotFound,
			wantMessage: "not found",
		},
		{
			name:        "ErrForbidden maps to 403",
			err:         ErrForbidden,
			wantStatus:  http.StatusForbidden,
			wantMessage: MsgForbidden,
		},
		{
			name:        "ErrBadRequest maps to 400",
			err:         ErrBadRequest,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "bad request",
		},
		{
			name:        "ErrConflict maps to 409",
			err:         ErrConflict,
			wantStatus:  http.StatusConflict,
			wantMessage: "conflict",
		},
		{
			name:        "unknown error maps to 500 without leaking details",
			err:         errors.New("secret database connection string failed"),
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			WriteErrorAuto(rec, tt.err)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			var body ErrorResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, tt.wantMessage, body.Message)

			// Verify the raw JSON shape is {"message": "..."}.
			var raw map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &raw))
			assert.Len(t, raw, 1)
			_, ok := raw["message"]
			assert.True(t, ok)
		})
	}
}

func TestWriteErrorAutoDoesNotLeakInternalError(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteErrorAuto(rec, errors.New("sensitive internal detail"))

	assert.NotContains(t, rec.Body.String(), "sensitive internal detail")
}
