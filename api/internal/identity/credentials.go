package identity

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrRegistrationDisabled = errors.New("registration is disabled")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrAccountBanned        = errors.New("account is banned")
	ErrInvalidVerification  = errors.New("invalid or expired verification")
)

const (
	maxFailedAttempts       = 5
	loginLockoutDuration    = 15 * time.Minute
	emailVerificationExpiry = 24 * time.Hour
	passwordResetExpiry     = time.Hour
)

type LockoutError struct {
	RetryAfter time.Duration
}

func (e *LockoutError) Error() string {
	return "account temporarily locked"
}

type CredentialUser struct {
	ID            string
	Name          string
	Email         string
	EmailVerified bool
	Image         *string
	Role          string
	Ban           Ban
}

type Verification struct {
	ID         string
	Identifier string
	Token      string
	ExpiresAt  time.Time
}

type CredentialRepository interface {
	EmailExists(ctx context.Context, email string) (bool, error)
	CreateCredentialUser(ctx context.Context, user CredentialUser, accountID, passwordHash string) error
	FindUserByEmail(ctx context.Context, email string) (CredentialUser, error)
	CredentialPassword(ctx context.Context, userID string) (string, error)
	CreateVerification(ctx context.Context, verification Verification) error
	FindVerification(ctx context.Context, token string) (Verification, error)
	VerifyEmail(ctx context.Context, userID, verificationID string) error
	ResetPassword(ctx context.Context, userID, verificationID, passwordHash string) error
}

type CredentialMailer interface {
	SendConfirmation(ctx context.Context, recipient, name, token string) error
	SendPasswordReset(ctx context.Context, recipient, name, token string) error
}

type CredentialConfig struct {
	RegistrationEnabled bool
}

type CredentialService struct {
	repository CredentialRepository
	sessions   *SessionService
	mailer     CredentialMailer
	config     CredentialConfig
	lockout    *loginLockout
	now        func() time.Time
}

func NewCredentialService(repository CredentialRepository, sessions *SessionService, mailer CredentialMailer, config CredentialConfig) *CredentialService {
	return &CredentialService{
		repository: repository,
		sessions:   sessions,
		mailer:     mailer,
		config:     config,
		lockout:    newLoginLockout(),
		now:        time.Now,
	}
}

type SignUpCommand struct {
	Email    string
	Password string
	Name     string
	Session  SessionMetadata
}

type Authentication struct {
	Token string
	User  CredentialUser
}

func (s *CredentialService) SignUp(ctx context.Context, command SignUpCommand) (Authentication, error) {
	if !s.config.RegistrationEnabled {
		return Authentication{}, ErrRegistrationDisabled
	}
	if len(command.Password) < 8 {
		return Authentication{}, ErrPasswordTooShort
	}
	exists, err := s.repository.EmailExists(ctx, command.Email)
	if err != nil {
		return Authentication{}, fmt.Errorf("check signup email: %w", err)
	}
	if exists {
		return Authentication{}, ErrEmailInUse
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(command.Password), 12)
	if err != nil {
		return Authentication{}, fmt.Errorf("hash signup password: %w", err)
	}
	user := CredentialUser{ID: uuid.NewString(), Name: command.Name, Email: command.Email, Role: "user"}
	if err := s.repository.CreateCredentialUser(ctx, user, uuid.NewString(), string(hash)); err != nil {
		return Authentication{}, fmt.Errorf("create credential user: %w", err)
	}

	verificationToken, tokenErr := randomToken()
	if tokenErr != nil {
		slog.ErrorContext(ctx, "failed to generate email verification token", "userID", user.ID, "error", tokenErr)
	} else if err := s.repository.CreateVerification(ctx, Verification{
		ID: uuid.NewString(), Identifier: user.Email, Token: verificationToken,
		ExpiresAt: s.now().Add(emailVerificationExpiry),
	}); err != nil {
		slog.ErrorContext(ctx, "failed to create email verification", "userID", user.ID, "error", err)
	} else if s.mailer != nil {
		if err := s.mailer.SendConfirmation(ctx, user.Email, user.Name, verificationToken); err != nil {
			slog.ErrorContext(ctx, "failed to send confirmation email", "userID", user.ID, "error", err)
		}
	}

	token, err := s.sessions.Issue(ctx, user.ID, command.Session)
	if err != nil {
		return Authentication{}, fmt.Errorf("issue signup session: %w", err)
	}
	return Authentication{Token: token, User: user}, nil
}

type SignInCommand struct {
	Email    string
	Password string
	Session  SessionMetadata
}

func (s *CredentialService) SignIn(ctx context.Context, command SignInCommand) (Authentication, error) {
	if retryAfter, locked := s.lockout.locked(command.Email, s.now()); locked {
		return Authentication{}, &LockoutError{RetryAfter: retryAfter}
	}
	user, err := s.repository.FindUserByEmail(ctx, command.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			s.lockout.recordFailure(command.Email, s.now())
			return Authentication{}, ErrInvalidCredentials
		}
		return Authentication{}, fmt.Errorf("find sign-in user: %w", err)
	}
	if user.Ban.ActiveAt(s.now()) {
		return Authentication{}, ErrAccountBanned
	}
	passwordHash, err := s.repository.CredentialPassword(ctx, user.ID)
	if err != nil {
		if errors.Is(err, ErrPasswordNotSet) {
			s.lockout.recordFailure(command.Email, s.now())
			return Authentication{}, ErrInvalidCredentials
		}
		return Authentication{}, fmt.Errorf("load sign-in credential: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(command.Password)); err != nil {
		s.lockout.recordFailure(command.Email, s.now())
		return Authentication{}, ErrInvalidCredentials
	}
	s.lockout.reset(command.Email)
	token, err := s.sessions.Issue(ctx, user.ID, command.Session)
	if err != nil {
		return Authentication{}, fmt.Errorf("issue sign-in session: %w", err)
	}
	return Authentication{Token: token, User: user}, nil
}

func (s *CredentialService) VerifyEmail(ctx context.Context, token string) error {
	verification, err := s.repository.FindVerification(ctx, token)
	if err != nil {
		return err
	}
	user, err := s.repository.FindUserByEmail(ctx, verification.Identifier)
	if err != nil {
		return fmt.Errorf("find verification user: %w", err)
	}
	if err := s.repository.VerifyEmail(ctx, user.ID, verification.ID); err != nil {
		return fmt.Errorf("verify user email: %w", err)
	}
	return nil
}

func (s *CredentialService) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.repository.FindUserByEmail(ctx, email)
	if errors.Is(err, ErrUserNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("find password reset user: %w", err)
	}
	token, err := randomToken()
	if err != nil {
		return fmt.Errorf("generate password reset token: %w", err)
	}
	if err := s.repository.CreateVerification(ctx, Verification{
		ID: uuid.NewString(), Identifier: user.ID, Token: token,
		ExpiresAt: s.now().Add(passwordResetExpiry),
	}); err != nil {
		return fmt.Errorf("create password reset verification: %w", err)
	}
	if s.mailer != nil {
		if err := s.mailer.SendPasswordReset(ctx, email, user.Name, token); err != nil {
			slog.ErrorContext(ctx, "failed to send password reset email", "userID", user.ID, "error", err)
		}
	}
	return nil
}

func (s *CredentialService) ResetPassword(ctx context.Context, token, password string) (string, error) {
	if len(password) < 8 {
		return "", ErrPasswordTooShort
	}
	verification, err := s.repository.FindVerification(ctx, token)
	if err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("hash reset password: %w", err)
	}
	if err := s.repository.ResetPassword(ctx, verification.Identifier, verification.ID, string(hash)); err != nil {
		return "", fmt.Errorf("reset credential password: %w", err)
	}
	return verification.Identifier, nil
}

type lockoutEntry struct {
	failures int
	lockedAt time.Time
	lastSeen time.Time
}

type loginLockout struct {
	mu      sync.Mutex
	entries map[string]*lockoutEntry
}

func newLoginLockout() *loginLockout {
	return &loginLockout{entries: make(map[string]*lockoutEntry)}
}

func (l *loginLockout) locked(email string, now time.Time) (time.Duration, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	key := credentialEmailKey(email)
	entry, ok := l.entries[key]
	if !ok || entry.failures < maxFailedAttempts {
		return 0, false
	}
	remaining := loginLockoutDuration - now.Sub(entry.lockedAt)
	if remaining <= 0 {
		delete(l.entries, key)
		return 0, false
	}
	return remaining, true
}

func (l *loginLockout) recordFailure(email string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	key := credentialEmailKey(email)
	entry, ok := l.entries[key]
	if !ok {
		entry = &lockoutEntry{}
		l.entries[key] = entry
	}
	entry.failures++
	entry.lastSeen = now
	if entry.failures == maxFailedAttempts {
		entry.lockedAt = now
	}
	for candidate, value := range l.entries {
		if now.Sub(value.lastSeen) > loginLockoutDuration {
			delete(l.entries, candidate)
		}
	}
}

func (l *loginLockout) reset(email string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, credentialEmailKey(email))
}

func credentialEmailKey(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
