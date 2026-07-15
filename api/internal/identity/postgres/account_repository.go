package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/identity"
	"github.com/jackc/pgx/v5"
)

type AccountRepository struct {
	queries *queries.Queries
}

var _ identity.AccountRepository = (*AccountRepository)(nil)

func NewAccountRepository(q *queries.Queries) *AccountRepository {
	return &AccountRepository{queries: q}
}

func (r *AccountRepository) ListSessions(ctx context.Context, userID string) ([]identity.UserSession, error) {
	rows, err := r.queries.ListUserSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query user sessions: %w", err)
	}
	sessions := make([]identity.UserSession, 0, len(rows))
	for _, row := range rows {
		sessions = append(sessions, identity.UserSession{
			ID: row.ID, ExpiresAt: row.ExpiresAt.Time, CreatedAt: row.CreatedAt.Time,
			IPAddress: row.IpAddress, UserAgent: row.UserAgent, ImpersonatedBy: row.ImpersonatedBy,
		})
	}
	return sessions, nil
}

func (r *AccountRepository) RevokeSession(ctx context.Context, userID, sessionID string) error {
	if err := r.queries.DeleteSessionByID(ctx, queries.DeleteSessionByIDParams{ID: sessionID, UserID: userID}); err != nil {
		return fmt.Errorf("execute session revoke: %w", err)
	}
	return nil
}

func (r *AccountRepository) DeleteSession(ctx context.Context, token string) error {
	if err := r.queries.DeleteSession(ctx, token); err != nil {
		return fmt.Errorf("execute session delete: %w", err)
	}
	return nil
}

func (r *AccountRepository) ListLinkedAccounts(ctx context.Context, userID string) ([]identity.LinkedAccount, error) {
	rows, err := r.queries.ListUserAccounts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query linked accounts: %w", err)
	}
	accounts := make([]identity.LinkedAccount, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, identity.LinkedAccount{
			ID: row.ID, Provider: row.ProviderID, AccountID: row.AccountID, CreatedAt: row.CreatedAt.Time,
		})
	}
	return accounts, nil
}

func (r *AccountRepository) UnlinkAccount(ctx context.Context, userID, providerID string) error {
	if err := r.queries.DeleteAccountByProviderAndUser(ctx, queries.DeleteAccountByProviderAndUserParams{ProviderID: providerID, UserID: userID}); err != nil {
		return fmt.Errorf("execute account unlink: %w", err)
	}
	return nil
}

func (r *AccountRepository) ListPasskeys(ctx context.Context, userID string) ([]identity.Passkey, error) {
	rows, err := r.queries.ListUserPasskeys(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query user passkeys: %w", err)
	}
	passkeys := make([]identity.Passkey, 0, len(rows))
	for _, row := range rows {
		passkeys = append(passkeys, identity.Passkey{
			ID: row.ID, Name: row.Name, DeviceType: row.DeviceType,
			BackedUp: row.BackedUp, CreatedAt: row.CreatedAt.Time,
		})
	}
	return passkeys, nil
}

func (r *AccountRepository) DeletePasskey(ctx context.Context, userID, passkeyID string) error {
	if err := r.queries.DeletePasskey(ctx, queries.DeletePasskeyParams{ID: passkeyID, UserID: userID}); err != nil {
		return fmt.Errorf("execute passkey delete: %w", err)
	}
	return nil
}

func (r *AccountRepository) CredentialPassword(ctx context.Context, userID string) (string, error) {
	account, err := r.queries.GetAccountByProviderAndUser(ctx, queries.GetAccountByProviderAndUserParams{
		ProviderID: "credential",
		UserID:     userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", identity.ErrPasswordNotSet
		}
		return "", fmt.Errorf("query credential account: %w", err)
	}
	if account.Password == nil {
		return "", identity.ErrPasswordNotSet
	}
	return *account.Password, nil
}

func (r *AccountRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	if _, err := r.queries.GetUserByEmail(ctx, email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query user by email: %w", err)
	}
	return true, nil
}

func (r *AccountRepository) UpdateEmail(ctx context.Context, userID, email string) error {
	if err := r.queries.UpdateUserEmail(ctx, queries.UpdateUserEmailParams{Email: email, ID: userID}); err != nil {
		return fmt.Errorf("execute user email update: %w", err)
	}
	return nil
}

func (r *AccountRepository) UpdateName(ctx context.Context, userID, name string) error {
	if err := r.queries.UpdateUserName(ctx, queries.UpdateUserNameParams{Name: name, ID: userID}); err != nil {
		return fmt.Errorf("execute user name update: %w", err)
	}
	return nil
}

func (r *AccountRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	if err := r.queries.UpdateUserPassword(ctx, queries.UpdateUserPasswordParams{Password: &passwordHash, UserID: userID}); err != nil {
		return fmt.Errorf("execute user password update: %w", err)
	}
	return nil
}

func (r *AccountRepository) DeleteUser(ctx context.Context, userID string) error {
	if err := r.queries.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("execute user delete: %w", err)
	}
	return nil
}
