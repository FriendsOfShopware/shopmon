package identity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrLastAuthenticationMethod = errors.New("cannot remove last authentication method")
	ErrPasswordNotSet           = errors.New("credential password is not set")
	ErrIncorrectPassword        = errors.New("current password is incorrect")
	ErrEmailInUse               = errors.New("email is already in use")
	ErrPasswordTooShort         = errors.New("password is too short")
)

type UserSession struct {
	ID             string
	ExpiresAt      time.Time
	CreatedAt      time.Time
	IPAddress      *string
	UserAgent      *string
	ImpersonatedBy *string
}

type LinkedAccount struct {
	ID        string
	Provider  string
	AccountID string
	CreatedAt time.Time
}

type Passkey struct {
	ID         string
	Name       *string
	DeviceType string
	BackedUp   bool
	CreatedAt  time.Time
}

type AccountRepository interface {
	ListSessions(ctx context.Context, userID string) ([]UserSession, error)
	RevokeSession(ctx context.Context, userID, sessionID string) error
	DeleteSession(ctx context.Context, token string) error
	ListLinkedAccounts(ctx context.Context, userID string) ([]LinkedAccount, error)
	UnlinkAccount(ctx context.Context, userID, providerID string) error
	ListPasskeys(ctx context.Context, userID string) ([]Passkey, error)
	DeletePasskey(ctx context.Context, userID, passkeyID string) error
	CredentialPassword(ctx context.Context, userID string) (string, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	UpdateEmail(ctx context.Context, userID, email string) error
	UpdateName(ctx context.Context, userID, name string) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	DeleteUser(ctx context.Context, userID string) error
}

type AccountService struct {
	repository AccountRepository
}

func NewAccountService(repository AccountRepository) *AccountService {
	return &AccountService{repository: repository}
}

func (s *AccountService) ListSessions(ctx context.Context, userID string) ([]UserSession, error) {
	sessions, err := s.repository.ListSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list user sessions: %w", err)
	}
	return sessions, nil
}

func (s *AccountService) RevokeSession(ctx context.Context, userID, sessionID string) error {
	if err := s.repository.RevokeSession(ctx, userID, sessionID); err != nil {
		return fmt.Errorf("revoke user session: %w", err)
	}
	return nil
}

func (s *AccountService) ListLinkedAccounts(ctx context.Context, userID string) ([]LinkedAccount, error) {
	accounts, err := s.repository.ListLinkedAccounts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list linked accounts: %w", err)
	}
	return accounts, nil
}

func (s *AccountService) UnlinkAccount(ctx context.Context, userID, providerID string) error {
	accounts, err := s.repository.ListLinkedAccounts(ctx, userID)
	if err != nil {
		return fmt.Errorf("list linked accounts before unlink: %w", err)
	}
	passkeys, err := s.repository.ListPasskeys(ctx, userID)
	if err != nil {
		return fmt.Errorf("list passkeys before unlink: %w", err)
	}
	if len(accounts)+len(passkeys) <= 1 {
		return ErrLastAuthenticationMethod
	}
	if err := s.repository.UnlinkAccount(ctx, userID, providerID); err != nil {
		return fmt.Errorf("unlink account provider: %w", err)
	}
	return nil
}

type ChangeEmailCommand struct {
	UserID          string
	Email           string
	CurrentPassword string
}

func (s *AccountService) ChangeEmail(ctx context.Context, command ChangeEmailCommand) error {
	passwordHash, err := s.repository.CredentialPassword(ctx, command.UserID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(command.CurrentPassword)); err != nil {
		return ErrIncorrectPassword
	}
	exists, err := s.repository.EmailExists(ctx, command.Email)
	if err != nil {
		return fmt.Errorf("check email availability: %w", err)
	}
	if exists {
		return ErrEmailInUse
	}
	if err := s.repository.UpdateEmail(ctx, command.UserID, command.Email); err != nil {
		return fmt.Errorf("update user email: %w", err)
	}
	return nil
}

func (s *AccountService) UpdateName(ctx context.Context, userID, name string) error {
	if name == "" {
		return nil
	}
	if err := s.repository.UpdateName(ctx, userID, name); err != nil {
		return fmt.Errorf("update user name: %w", err)
	}
	return nil
}

type ChangePasswordCommand struct {
	UserID          string
	CurrentPassword string
	NewPassword     string
}

func (s *AccountService) ChangePassword(ctx context.Context, command ChangePasswordCommand) error {
	if len(command.NewPassword) < 8 {
		return ErrPasswordTooShort
	}
	passwordHash, err := s.repository.CredentialPassword(ctx, command.UserID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(command.CurrentPassword)); err != nil {
		return ErrIncorrectPassword
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(command.NewPassword), 12)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}
	if err := s.repository.UpdatePassword(ctx, command.UserID, string(newHash)); err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	return nil
}

func (s *AccountService) DeleteUser(ctx context.Context, userID string) error {
	if err := s.repository.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func (s *AccountService) ListPasskeys(ctx context.Context, userID string) ([]Passkey, error) {
	passkeys, err := s.repository.ListPasskeys(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list user passkeys: %w", err)
	}
	return passkeys, nil
}

func (s *AccountService) DeletePasskey(ctx context.Context, userID, passkeyID string) error {
	if err := s.repository.DeletePasskey(ctx, userID, passkeyID); err != nil {
		return fmt.Errorf("delete user passkey: %w", err)
	}
	return nil
}

func (s *AccountService) StopImpersonating(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	if err := s.repository.DeleteSession(ctx, token); err != nil {
		return fmt.Errorf("delete impersonation session: %w", err)
	}
	return nil
}
