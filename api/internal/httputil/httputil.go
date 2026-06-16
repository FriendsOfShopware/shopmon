// Package httputil provides shared HTTP helpers used across handler, auth, and middleware packages.
package httputil

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

// ErrorResponse is the standard error response shape for all endpoints.
type ErrorResponse struct {
	Message string `json:"message"`
}

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
