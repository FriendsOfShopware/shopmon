package identity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/jackc/pgx/v5"
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

type AccountService struct {
	queries *queries.Queries
}

func NewAccountService(q *queries.Queries) *AccountService {
	return &AccountService{queries: q}
}

func (s *AccountService) ListSessions(ctx context.Context, userID string) ([]UserSession, error) {
	rows, err := s.queries.ListUserSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list user sessions: %w", err)
	}
	sessions := make([]UserSession, 0, len(rows))
	for _, row := range rows {
		sessions = append(sessions, UserSession{
			ID: row.ID, ExpiresAt: row.ExpiresAt.Time, CreatedAt: row.CreatedAt.Time,
			IPAddress: row.IpAddress, UserAgent: row.UserAgent, ImpersonatedBy: row.ImpersonatedBy,
		})
	}
	return sessions, nil
}

func (s *AccountService) RevokeSession(ctx context.Context, userID, sessionID string) error {
	if err := s.queries.DeleteSessionByID(ctx, queries.DeleteSessionByIDParams{ID: sessionID, UserID: userID}); err != nil {
		return fmt.Errorf("revoke user session: %w", err)
	}
	return nil
}

func (s *AccountService) ListLinkedAccounts(ctx context.Context, userID string) ([]LinkedAccount, error) {
	rows, err := s.queries.ListUserAccounts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list linked accounts: %w", err)
	}
	accounts := make([]LinkedAccount, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, LinkedAccount{
			ID: row.ID, Provider: row.ProviderID, AccountID: row.AccountID, CreatedAt: row.CreatedAt.Time,
		})
	}
	return accounts, nil
}

func (s *AccountService) UnlinkAccount(ctx context.Context, userID, providerID string) error {
	accounts, err := s.ListLinkedAccounts(ctx, userID)
	if err != nil {
		return fmt.Errorf("list linked accounts before unlink: %w", err)
	}
	passkeys, err := s.ListPasskeys(ctx, userID)
	if err != nil {
		return fmt.Errorf("list passkeys before unlink: %w", err)
	}
	if len(accounts)+len(passkeys) <= 1 {
		return ErrLastAuthenticationMethod
	}
	if err := s.queries.DeleteAccountByProviderAndUser(ctx, queries.DeleteAccountByProviderAndUserParams{ProviderID: providerID, UserID: userID}); err != nil {
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
	passwordHash, err := s.credentialPassword(ctx, command.UserID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(command.CurrentPassword)); err != nil {
		return ErrIncorrectPassword
	}
	exists, err := s.emailExists(ctx, command.Email)
	if err != nil {
		return fmt.Errorf("check email availability: %w", err)
	}
	if exists {
		return ErrEmailInUse
	}
	if err := s.queries.UpdateUserEmail(ctx, queries.UpdateUserEmailParams{Email: command.Email, ID: command.UserID}); err != nil {
		return fmt.Errorf("update user email: %w", err)
	}
	return nil
}

func (s *AccountService) UpdateName(ctx context.Context, userID, name string) error {
	if name == "" {
		return nil
	}
	if err := s.queries.UpdateUserName(ctx, queries.UpdateUserNameParams{Name: name, ID: userID}); err != nil {
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
	passwordHash, err := s.credentialPassword(ctx, command.UserID)
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
	hash := string(newHash)
	if err := s.queries.UpdateUserPassword(ctx, queries.UpdateUserPasswordParams{Password: &hash, UserID: command.UserID}); err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	return nil
}

func (s *AccountService) DeleteUser(ctx context.Context, userID string) error {
	if err := s.queries.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func (s *AccountService) ListPasskeys(ctx context.Context, userID string) ([]Passkey, error) {
	rows, err := s.queries.ListUserPasskeys(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list user passkeys: %w", err)
	}
	passkeys := make([]Passkey, 0, len(rows))
	for _, row := range rows {
		passkeys = append(passkeys, Passkey{
			ID: row.ID, Name: row.Name, DeviceType: row.DeviceType,
			BackedUp: row.BackedUp, CreatedAt: row.CreatedAt.Time,
		})
	}
	return passkeys, nil
}

func (s *AccountService) DeletePasskey(ctx context.Context, userID, passkeyID string) error {
	if err := s.queries.DeletePasskey(ctx, queries.DeletePasskeyParams{ID: passkeyID, UserID: userID}); err != nil {
		return fmt.Errorf("delete user passkey: %w", err)
	}
	return nil
}

func (s *AccountService) StopImpersonating(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	if err := s.queries.DeleteSession(ctx, token); err != nil {
		return fmt.Errorf("delete impersonation session: %w", err)
	}
	return nil
}

func (s *AccountService) credentialPassword(ctx context.Context, userID string) (string, error) {
	account, err := s.queries.GetAccountByProviderAndUser(ctx, queries.GetAccountByProviderAndUserParams{
		ProviderID: "credential",
		UserID:     userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrPasswordNotSet
		}
		return "", fmt.Errorf("query credential account: %w", err)
	}
	if account.Password == nil {
		return "", ErrPasswordNotSet
	}
	return *account.Password, nil
}

func (s *AccountService) emailExists(ctx context.Context, email string) (bool, error) {
	if _, err := s.queries.GetUserByEmail(ctx, email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query user by email: %w", err)
	}
	return true, nil
}
