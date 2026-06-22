// Package httputil provides shared HTTP helpers used across handler, auth, and middleware packages.
package httputil

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
)

// ErrorResponse is the standard error response shape for all endpoints.
type ErrorResponse struct {
	Message string `json:"message"`
}

// Sentinel errors that handlers can return (optionally wrapped via fmt.Errorf
// with %w) so that WriteErrorAuto can map them to the appropriate HTTP status
// code centrally.
var (
	// ErrNotFound maps to 404 Not Found.
	ErrNotFound = errors.New("not found")
	// ErrForbidden maps to 403 Forbidden.
	ErrForbidden = errors.New("forbidden")
	// ErrBadRequest maps to 400 Bad Request.
	ErrBadRequest = errors.New("bad request")
	// ErrConflict maps to 409 Conflict.
	ErrConflict = errors.New("conflict")
)

// Standard 403 messages. Use these for generic authorization failures so the
// API returns consistent wording; reserve bespoke messages for cases that
// convey a distinct, actionable reason (e.g. role-hierarchy or tenancy rules).
const (
	// MsgForbidden is the generic insufficient-permissions message.
	MsgForbidden = "insufficient permissions"
	// MsgAdminRequired is returned when an action requires the admin role.
	MsgAdminRequired = "admin access required"
)

// WriteForbidden writes a 403 with the standard insufficient-permissions message.
func WriteForbidden(w http.ResponseWriter) {
	WriteError(w, http.StatusForbidden, MsgForbidden)
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

// WriteError writes a standardized JSON error response.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, ErrorResponse{Message: message})
}

// WriteErrorAuto inspects err and writes an appropriate HTTP error response.
//
// It maps well-known sentinel errors (and pgx.ErrNoRows) to specific status
// codes. For any unrecognized error it logs the real error via slog and returns
// a generic 500 response so that internal details are never leaked to the
// client.
func WriteErrorAuto(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, pgx.ErrNoRows), errors.Is(err, ErrNotFound):
		WriteError(w, http.StatusNotFound, "not found")
	case errors.Is(err, ErrForbidden):
		WriteError(w, http.StatusForbidden, MsgForbidden)
	case errors.Is(err, ErrBadRequest):
		WriteError(w, http.StatusBadRequest, "bad request")
	case errors.Is(err, ErrConflict):
		WriteError(w, http.StatusConflict, "conflict")
	default:
		slog.Error("unhandled error", "error", err)
		WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}

// DecodeBody decodes a JSON request body into the given value.
func DecodeBody(r *http.Request, v interface{}) error {
	defer func() { _ = r.Body.Close() }()
	return json.NewDecoder(r.Body).Decode(v)
}

// ExtractToken extracts the session token from the Authorization: Bearer header.
func ExtractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}
