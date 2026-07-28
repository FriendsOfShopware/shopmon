package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/access"
	"github.com/friendsofshopware/shopmon/api/internal/authapi"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/identity"
)

// SignUpEmail handles email+password registration.
func (h *AuthHandler) SignUpEmail(w http.ResponseWriter, r *http.Request) {
	var request signUpEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if request.Email == "" || request.Password == "" || request.Name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "email, password, and name are required")
		return
	}

	authentication, err := h.credentials.SignUp(r.Context(), identity.SignUpCommand{
		Email: request.Email, Password: request.Password, Name: request.Name,
		Session: sessionMetadata(r),
	})
	if err != nil {
		h.writeCredentialError(w, r, "sign up", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, tokenUserResponse{
		Token: authentication.Token,
		User: authenticatedUserResponse{
			ID: authentication.User.ID, Name: authentication.User.Name, Email: authentication.User.Email,
		},
	})
}

// SignInEmail handles email+password login.
func (h *AuthHandler) SignInEmail(w http.ResponseWriter, r *http.Request) {
	var request signInEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	authentication, err := h.credentials.SignIn(r.Context(), identity.SignInCommand{
		Email: request.Email, Password: request.Password, Session: sessionMetadata(r),
	})
	if err != nil {
		h.writeCredentialError(w, r, "sign in", err)
		return
	}
	user := authentication.User
	httputil.WriteJSON(w, http.StatusOK, tokenUserResponse{
		Token: authentication.Token,
		User: authenticatedUserResponse{
			ID: user.ID, Name: user.Name, Email: user.Email, EmailVerified: user.EmailVerified,
			Image: user.Image, Role: user.Role,
		},
	})
}

// SignOut logs the user out.
func (h *AuthHandler) SignOut(w http.ResponseWriter, r *http.Request) {
	if err := h.sessions.Delete(r.Context(), httputil.ExtractToken(r)); err != nil {
		slog.ErrorContext(r.Context(), "failed to delete session on sign out", "error", err)
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

// GetSession returns the current session and user.
func (h *AuthHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	token := httputil.ExtractToken(r)
	if token == "" {
		httputil.WriteError(w, http.StatusUnauthorized, "no session")
		return
	}
	principal := access.PrincipalFromContext(r.Context())
	if principal == nil {
		httputil.WriteError(w, http.StatusUnauthorized, "invalid session")
		return
	}
	activeOrganizationID, err := h.organizations.EnsureActive(r.Context(), principal.User.ID, token, principal.Session.ActiveOrganizationID)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to ensure active organization", "userID", principal.User.ID, "error", err)
	} else {
		principal.Session.ActiveOrganizationID = activeOrganizationID
	}

	httputil.WriteJSON(w, http.StatusOK, currentSessionEnvelope{
		User: authenticatedUserResponse{
			ID: principal.User.ID, Name: principal.User.Name, Email: principal.User.Email,
			Image: principal.User.Image, Role: principal.User.Role, Notifications: principal.User.Notifications,
		},
		Session: currentSessionResponse{
			ID: principal.Session.ID, UserID: principal.Session.UserID, ExpiresAt: principal.Session.ExpiresAt,
			ActiveOrganizationID: principal.Session.ActiveOrganizationID,
			ImpersonatedBy:       principal.Session.ImpersonatedBy,
		},
	})
}

// VerifyEmail verifies a user's email address.
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request, params authapi.VerifyEmailParams) {
	if params.Token == "" {
		httputil.WriteError(w, http.StatusBadRequest, "token is required")
		return
	}
	if err := h.credentials.VerifyEmail(r.Context(), params.Token); err != nil {
		h.writeCredentialError(w, r, "verify email", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

// ForgetPassword sends a password reset email. The response never reveals
// whether the supplied email belongs to an account.
func (h *AuthHandler) ForgetPassword(w http.ResponseWriter, r *http.Request) {
	var request forgetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.credentials.RequestPasswordReset(r.Context(), request.Email); err != nil {
		slog.ErrorContext(r.Context(), "password reset request failed", "error", err)
	}
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

// ResetPassword resets a user's password using a one-time token.
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var request resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userID, err := h.credentials.ResetPassword(r.Context(), request.Token, request.NewPassword)
	if err != nil {
		h.writeCredentialError(w, r, "reset password", err)
		return
	}
	h.recordAudit(r, userID, AuditActionPasswordReset, userID, "")
	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

func (h *AuthHandler) writeCredentialError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	var lockout *identity.LockoutError
	switch {
	case errors.As(err, &lockout):
		w.Header().Set("Retry-After", fmt.Sprintf("%d", int(lockout.RetryAfter.Seconds())))
		httputil.WriteError(w, http.StatusTooManyRequests, "account temporarily locked due to too many failed login attempts, please try again later")
	case errors.Is(err, identity.ErrRegistrationDisabled):
		httputil.WriteError(w, http.StatusForbidden, "registration is disabled")
	case errors.Is(err, identity.ErrPasswordTooShort):
		httputil.WriteError(w, http.StatusBadRequest, "password must be at least 8 characters")
	case errors.Is(err, identity.ErrEmailInUse):
		httputil.WriteError(w, http.StatusConflict, "user with this email already exists")
	case errors.Is(err, identity.ErrInvalidCredentials):
		httputil.WriteError(w, http.StatusUnauthorized, "invalid email or password")
	case errors.Is(err, identity.ErrAccountBanned):
		httputil.WriteError(w, http.StatusForbidden, "account is banned")
	case errors.Is(err, identity.ErrInvalidVerification):
		httputil.WriteError(w, http.StatusBadRequest, "invalid or expired token")
	case errors.Is(err, identity.ErrUserNotFound):
		httputil.WriteError(w, http.StatusBadRequest, "user not found")
	default:
		slog.ErrorContext(r.Context(), "credential operation failed", "operation", operation, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "credential operation failed")
	}
}
