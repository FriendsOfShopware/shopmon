package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/authapi"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

// SignUpEmail handles email+password registration.
func (h *AuthHandler) SignUpEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "email, password, and name are required")
		return
	}

	if len(req.Password) < 8 {
		httputil.WriteError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	_, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err == nil {
		httputil.WriteError(w, http.StatusConflict, "user with this email already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	userID := uuid.New().String()
	_, err = h.queries.CreateUser(r.Context(), queries.CreateUserParams{
		ID:    userID,
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		slog.Error("failed to create user", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Create credential account
	passwordStr := string(hashedPassword)
	err = h.queries.CreateAccount(r.Context(), queries.CreateAccountParams{
		ID:         uuid.New().String(),
		AccountID:  userID,
		ProviderID: "credential",
		UserID:     userID,
		Password:   &passwordStr,
	})
	if err != nil {
		slog.Error("failed to create account", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create account")
		return
	}

	// Send verification email
	verificationToken := generateToken()
	if err := h.queries.CreateVerification(r.Context(), queries.CreateVerificationParams{
		ID:         uuid.New().String(),
		Identifier: req.Email,
		Value:      verificationToken,
		ExpiresAt:  pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
	}); err != nil {
		slog.Warn("failed to create email verification", "error", err, "email", req.Email)
	}

	verifyURL := h.cfg.FrontendURL + "/account/confirm/" + verificationToken
	h.mail.Send(req.Email, "Confirm your email address",
		mail.BuildConfirmationEmail(req.Name, verifyURL))

	// Create session
	token, err := h.createSession(r, userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":    userID,
			"name":  req.Name,
			"email": req.Email,
		},
	})
}

// SignInEmail handles email+password login.
func (h *AuthHandler) SignInEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if user.Banned != nil && *user.Banned {
		httputil.WriteError(w, http.StatusForbidden, "account is banned")
		return
	}

	account, err := h.queries.GetAccountByProviderAndUser(r.Context(), queries.GetAccountByProviderAndUserParams{
		ProviderID: "credential",
		UserID:     user.ID,
	})
	if err != nil || account.Password == nil {
		httputil.WriteError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*account.Password), []byte(req.Password)); err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := h.createSession(r, user.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":            user.ID,
			"name":          user.Name,
			"email":         user.Email,
			"emailVerified": user.EmailVerified,
			"image":         user.Image,
			"role":          user.Role,
		},
	})
}

// SignOut logs the user out.
func (h *AuthHandler) SignOut(w http.ResponseWriter, r *http.Request) {
	token := httputil.ExtractToken(r)
	if token != "" {
		h.queries.DeleteSession(r.Context(), token)
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GetSession returns the current session and user.
func (h *AuthHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	token := httputil.ExtractToken(r)
	if token == "" {
		httputil.WriteError(w, http.StatusUnauthorized, "no session")
		return
	}

	su, err := ValidateSession(r.Context(), h.pool, token)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "invalid session")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":            su.User.ID,
			"name":          su.User.Name,
			"email":         su.User.Email,
			"image":         su.User.Image,
			"role":          su.User.Role,
			"notifications": su.User.Notifications,
		},
		"session": map[string]interface{}{
			"id":                   su.Session.ID,
			"userId":               su.Session.UserID,
			"expiresAt":            su.Session.ExpiresAt,
			"activeOrganizationId": su.Session.ActiveOrganizationID,
		},
	})
}

// VerifyEmail verifies a user's email address.
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request, params authapi.VerifyEmailParams) {
	token := params.Token
	if token == "" {
		httputil.WriteError(w, http.StatusBadRequest, "token is required")
		return
	}

	verification, err := h.queries.GetVerificationByValue(r.Context(), token)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid or expired token")
		return
	}

	user, err := h.queries.GetUserByEmail(r.Context(), verification.Identifier)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "user not found")
		return
	}

	if err = h.queries.UpdateUserEmailVerified(r.Context(), user.ID); err != nil {
		slog.Warn("failed to update email verified status", "error", err, "userID", user.ID)
	}
	if err = h.queries.DeleteVerification(r.Context(), verification.ID); err != nil {
		slog.Warn("failed to delete verification", "error", err, "verificationID", verification.ID)
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ForgetPassword sends a password reset email.
func (h *AuthHandler) ForgetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Always return success to prevent email enumeration
	user, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}

	resetToken := generateToken()
	if err := h.queries.CreateVerification(r.Context(), queries.CreateVerificationParams{
		ID:         uuid.New().String(),
		Identifier: user.ID,
		Value:      resetToken,
		ExpiresAt:  pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour), Valid: true},
	}); err != nil {
		slog.Warn("failed to create password reset verification", "error", err, "userID", user.ID)
	}

	resetURL := h.cfg.FrontendURL + "/account/forgot-password/" + resetToken
	h.mail.Send(req.Email, "Reset your password",
		mail.BuildPasswordResetEmail(user.Name, resetURL))

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ResetPassword resets a user's password using a token.
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.NewPassword) < 8 {
		httputil.WriteError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	verification, err := h.queries.GetVerificationByValue(r.Context(), req.Token)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid or expired token")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	passwordStr := string(hashedPassword)
	if err := h.queries.UpdateUserPassword(r.Context(), queries.UpdateUserPasswordParams{
		Password: &passwordStr,
		UserID:   verification.Identifier,
	}); err != nil {
		slog.Error("failed to update user password", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to reset password")
		return
	}

	h.queries.DeleteUserSessions(r.Context(), verification.Identifier)
	h.queries.DeleteVerification(r.Context(), verification.ID)

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
