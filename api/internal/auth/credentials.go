package auth

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/authapi"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

// loginLockout tracks failed login attempts per email and locks an account
// after maxFailedAttempts failures within the sliding window. State is kept
// in-memory (resets on restart); for multi-instance deployments this should be
// backed by a shared store.
const (
	maxFailedAttempts = 5
	lockoutDuration   = 15 * time.Minute
)

type lockoutEntry struct {
	failures int
	lockedAt time.Time
	lastSeen time.Time
}

type loginLockout struct {
	mu       sync.Mutex
	entries  map[string]*lockoutEntry
	sweepers sync.Once
}

var failedLogins = &loginLockout{entries: make(map[string]*lockoutEntry)}

// startSweeper launches a single background goroutine that periodically evicts
// stale entries so the map cannot grow without bound (e.g. from login attempts
// against many distinct non-existent emails). Started lazily on first write.
func (l *loginLockout) startSweeper() {
	l.sweepers.Do(func() {
		go func() {
			ticker := time.NewTicker(lockoutDuration)
			defer ticker.Stop()
			for range ticker.C {
				l.sweep()
			}
		}()
	})
}

// sweep removes entries that have been idle longer than the lockout window.
func (l *loginLockout) sweep() {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := time.Now().Add(-lockoutDuration)
	for key, e := range l.entries {
		if e.lastSeen.Before(cutoff) {
			delete(l.entries, key)
		}
	}
}

func lockoutKey(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// locked reports whether the email is currently locked out.
func (l *loginLockout) locked(email string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.entries[lockoutKey(email)]
	if !ok {
		return false
	}
	if e.failures >= maxFailedAttempts {
		if time.Since(e.lockedAt) < lockoutDuration {
			return true
		}
		// Lockout window elapsed, reset.
		delete(l.entries, lockoutKey(email))
	}
	return false
}

// recordFailure increments the failure counter and starts the lockout window
// once the threshold is reached.
func (l *loginLockout) recordFailure(email string) {
	l.startSweeper()

	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	key := lockoutKey(email)
	e, ok := l.entries[key]
	if !ok {
		e = &lockoutEntry{}
		l.entries[key] = e
	}
	e.failures++
	e.lastSeen = now
	if e.failures >= maxFailedAttempts {
		e.lockedAt = now
	}
}

// reset clears the failure counter after a successful login.
func (l *loginLockout) reset(email string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, lockoutKey(email))
}

// SignUpEmail handles email+password registration.
func (h *AuthHandler) SignUpEmail(w http.ResponseWriter, r *http.Request) {
	if h.cfg.DisableRegistration {
		httputil.WriteError(w, http.StatusForbidden, "registration is disabled")
		return
	}

	var req signUpEmailRequest
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
	passwordStr := string(hashedPassword)
	err = h.withTx(r.Context(), func(txq *queries.Queries) error {
		if _, err := txq.CreateUser(r.Context(), queries.CreateUserParams{
			ID:    userID,
			Name:  req.Name,
			Email: req.Email,
		}); err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		if err := txq.CreateAccount(r.Context(), queries.CreateAccountParams{
			ID:         uuid.New().String(),
			AccountID:  userID,
			ProviderID: "credential",
			UserID:     userID,
			Password:   &passwordStr,
		}); err != nil {
			return fmt.Errorf("create account: %w", err)
		}

		return nil
	})
	if err != nil {
		slog.Error("failed to create user account", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create user")
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
	if err := h.mail.Send(req.Email, "Confirm your email address",
		mail.BuildConfirmationEmail(req.Name, verifyURL)); err != nil {
		slog.Error("failed to send confirmation email", "error", err, "email", req.Email)
	}

	// Create session
	token, err := h.createSession(r, userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, tokenUserResponse{
		Token: token,
		User: authenticatedUserResponse{
			ID:    userID,
			Name:  req.Name,
			Email: req.Email,
		},
	})
}

// SignInEmail handles email+password login.
func (h *AuthHandler) SignInEmail(w http.ResponseWriter, r *http.Request) {
	var req signInEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if failedLogins.locked(req.Email) {
		w.Header().Set("Retry-After", fmt.Sprintf("%d", int(lockoutDuration.Seconds())))
		httputil.WriteError(w, http.StatusTooManyRequests, "account temporarily locked due to too many failed login attempts, please try again later")
		return
	}

	user, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		failedLogins.recordFailure(req.Email)
		httputil.WriteError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if isBanActive(user.Banned, user.BanExpires) {
		httputil.WriteError(w, http.StatusForbidden, "account is banned")
		return
	}

	account, err := h.queries.GetAccountByProviderAndUser(r.Context(), queries.GetAccountByProviderAndUserParams{
		ProviderID: "credential",
		UserID:     user.ID,
	})
	if err != nil || account.Password == nil {
		failedLogins.recordFailure(req.Email)
		httputil.WriteError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*account.Password), []byte(req.Password)); err != nil {
		failedLogins.recordFailure(req.Email)
		httputil.WriteError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	failedLogins.reset(req.Email)

	token, err := h.createSession(r, user.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, tokenUserResponse{
		Token: token,
		User: authenticatedUserResponse{
			ID:            user.ID,
			Name:          user.Name,
			Email:         user.Email,
			EmailVerified: user.EmailVerified,
			Image:         user.Image,
			Role:          user.Role,
		},
	})
}

// SignOut logs the user out.
func (h *AuthHandler) SignOut(w http.ResponseWriter, r *http.Request) {
	token := httputil.ExtractToken(r)
	if token != "" {
		if err := h.queries.DeleteSession(r.Context(), token); err != nil {
			slog.Error("failed to delete session on sign out", "error", err)
		}
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

	su, err := ValidateSession(r.Context(), h.queries, token)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "invalid session")
		return
	}

	// Auto-select first organization if none is active
	if su.Session.ActiveOrganizationID == nil {
		orgs, err := h.queries.ListUserOrganizations(r.Context(), su.User.ID)
		if err == nil && len(orgs) > 0 {
			orgID := orgs[0].ID
			_ = h.queries.SetActiveOrganization(r.Context(), queries.SetActiveOrganizationParams{
				ActiveOrganizationID: &orgID,
				Token:                token,
			})
			su.Session.ActiveOrganizationID = &orgID
		}
	}

	httputil.WriteJSON(w, http.StatusOK, currentSessionEnvelope{
		User: authenticatedUserResponse{
			ID:            su.User.ID,
			Name:          su.User.Name,
			Email:         su.User.Email,
			Image:         su.User.Image,
			Role:          su.User.Role,
			Notifications: su.User.Notifications,
		},
		Session: currentSessionResponse{
			ID:                   su.Session.ID,
			UserID:               su.Session.UserID,
			ExpiresAt:            su.Session.ExpiresAt,
			ActiveOrganizationID: su.Session.ActiveOrganizationID,
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

	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

// ForgetPassword sends a password reset email.
func (h *AuthHandler) ForgetPassword(w http.ResponseWriter, r *http.Request) {
	var req forgetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Always return success to prevent email enumeration
	user, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
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
	if err := h.mail.Send(req.Email, "Reset your password",
		mail.BuildPasswordResetEmail(user.Name, resetURL)); err != nil {
		slog.Error("failed to send password reset email", "error", err, "email", req.Email)
	}

	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}

// ResetPassword resets a user's password using a token.
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
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
	err = h.withTx(r.Context(), func(txq *queries.Queries) error {
		if err := txq.UpdateUserPassword(r.Context(), queries.UpdateUserPasswordParams{
			Password: &passwordStr,
			UserID:   verification.Identifier,
		}); err != nil {
			return fmt.Errorf("update user password: %w", err)
		}

		if err := txq.DeleteUserSessions(r.Context(), verification.Identifier); err != nil {
			return fmt.Errorf("delete user sessions: %w", err)
		}

		if err := txq.DeleteVerification(r.Context(), verification.ID); err != nil {
			return fmt.Errorf("delete verification: %w", err)
		}

		return nil
	})
	if err != nil {
		slog.Error("failed to reset password", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to reset password")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, newStatusResponse())
}
