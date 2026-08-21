package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/access"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/identity"
)

// SignUpEmail handles email+password registration.
func (h *AuthHandler) SignUpEmail(ctx context.Context, input *signUpEmailInput) (*tokenUserOutput, error) {
	request := input.Body
	if request.Email == "" || request.Password == "" || request.Name == "" {
		return nil, huma.Error400BadRequest("email, password, and name are required")
	}

	authentication, err := h.credentials.SignUp(ctx, identity.SignUpCommand{
		Email: request.Email, Password: request.Password, Name: request.Name,
		Session: sessionMetadata(requestFromContext(ctx)),
	})
	if err != nil {
		return nil, h.credentialError(ctx, "sign up", err)
	}
	return &tokenUserOutput{Body: tokenUserResponse{
		Token: authentication.Token,
		User: authenticatedUserResponse{
			ID: authentication.User.ID, Name: authentication.User.Name, Email: authentication.User.Email,
		},
	}}, nil
}

// SignInEmail handles email+password login.
func (h *AuthHandler) SignInEmail(ctx context.Context, input *signInEmailInput) (*tokenUserOutput, error) {
	request := input.Body
	authentication, err := h.credentials.SignIn(ctx, identity.SignInCommand{
		Email: request.Email, Password: request.Password, Session: sessionMetadata(requestFromContext(ctx)),
	})
	if err != nil {
		return nil, h.credentialError(ctx, "sign in", err)
	}
	user := authentication.User
	return &tokenUserOutput{Body: tokenUserResponse{
		Token: authentication.Token,
		User: authenticatedUserResponse{
			ID: user.ID, Name: user.Name, Email: user.Email, EmailVerified: user.EmailVerified,
			Image: user.Image, Role: user.Role,
		},
	}}, nil
}

// SignOut logs the user out.
func (h *AuthHandler) SignOut(ctx context.Context, _ *struct{}) (*statusOutput, error) {
	if err := h.sessions.Delete(ctx, httputil.ExtractToken(requestFromContext(ctx))); err != nil {
		slog.ErrorContext(ctx, "failed to delete session on sign out", "error", err)
	}
	return statusOK(), nil
}

// GetSession returns the current session and user.
func (h *AuthHandler) GetSession(ctx context.Context, _ *struct{}) (*sessionOutput, error) {
	token := httputil.ExtractToken(requestFromContext(ctx))
	if token == "" {
		return nil, huma.Error401Unauthorized("no session")
	}
	principal := access.PrincipalFromContext(ctx)
	if principal == nil {
		return nil, huma.Error401Unauthorized("invalid session")
	}
	activeOrganizationID, err := h.organizations.EnsureActive(ctx, principal.User.ID, token, principal.Session.ActiveOrganizationID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to ensure active organization", "userID", principal.User.ID, "error", err)
	} else {
		principal.Session.ActiveOrganizationID = activeOrganizationID
	}

	return &sessionOutput{Body: currentSessionEnvelope{
		User: authenticatedUserResponse{
			ID: principal.User.ID, Name: principal.User.Name, Email: principal.User.Email,
			Image: principal.User.Image, Role: principal.User.Role, Notifications: principal.User.Notifications,
		},
		Session: currentSessionResponse{
			ID: principal.Session.ID, UserID: principal.Session.UserID, ExpiresAt: principal.Session.ExpiresAt,
			ActiveOrganizationID: principal.Session.ActiveOrganizationID,
			ImpersonatedBy:       principal.Session.ImpersonatedBy,
		},
	}}, nil
}

// VerifyEmail verifies a user's email address.
func (h *AuthHandler) VerifyEmail(ctx context.Context, input *verifyEmailInput) (*statusOutput, error) {
	if input.Token == "" {
		return nil, huma.Error400BadRequest("token is required")
	}
	if err := h.credentials.VerifyEmail(ctx, input.Token); err != nil {
		return nil, h.credentialError(ctx, "verify email", err)
	}
	return statusOK(), nil
}

// ForgetPassword sends a password reset email. The response never reveals
// whether the supplied email belongs to an account.
func (h *AuthHandler) ForgetPassword(ctx context.Context, input *forgetPasswordInput) (*statusOutput, error) {
	if err := h.credentials.RequestPasswordReset(ctx, input.Body.Email); err != nil {
		slog.ErrorContext(ctx, "password reset request failed", "error", err)
	}
	return statusOK(), nil
}

// ResetPassword resets a user's password using a one-time token.
func (h *AuthHandler) ResetPassword(ctx context.Context, input *resetPasswordInput) (*statusOutput, error) {
	userID, err := h.credentials.ResetPassword(ctx, input.Body.Token, input.Body.NewPassword)
	if err != nil {
		return nil, h.credentialError(ctx, "reset password", err)
	}
	h.recordAudit(requestFromContext(ctx), userID, AuditActionPasswordReset, userID, "")
	return statusOK(), nil
}

func (h *AuthHandler) credentialError(ctx context.Context, operation string, err error) error {
	var lockout *identity.LockoutError
	switch {
	case errors.As(err, &lockout):
		return huma.ErrorWithHeaders(
			huma.Error429TooManyRequests("account temporarily locked due to too many failed login attempts, please try again later"),
			http.Header{"Retry-After": {fmt.Sprintf("%d", int(lockout.RetryAfter.Seconds()))}},
		)
	case errors.Is(err, identity.ErrRegistrationDisabled):
		return huma.Error403Forbidden("registration is disabled")
	case errors.Is(err, identity.ErrPasswordTooShort):
		return huma.Error400BadRequest("password must be at least 8 characters")
	case errors.Is(err, identity.ErrEmailInUse):
		return huma.Error409Conflict("user with this email already exists")
	case errors.Is(err, identity.ErrInvalidCredentials):
		return huma.Error401Unauthorized("invalid email or password")
	case errors.Is(err, identity.ErrAccountBanned):
		return huma.Error403Forbidden("account is banned")
	case errors.Is(err, identity.ErrInvalidVerification):
		return huma.Error400BadRequest("invalid or expired token")
	case errors.Is(err, identity.ErrUserNotFound):
		return huma.Error400BadRequest("user not found")
	default:
		slog.ErrorContext(ctx, "credential operation failed", "operation", operation, "error", err)
		return huma.Error500InternalServerError("credential operation failed")
	}
}
