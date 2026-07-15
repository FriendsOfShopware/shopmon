package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/identity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CredentialRepository struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
}

var _ identity.CredentialRepository = (*CredentialRepository)(nil)

func NewCredentialRepository(pool *pgxpool.Pool, q *queries.Queries) *CredentialRepository {
	return &CredentialRepository{pool: pool, queries: q}
}

func (r *CredentialRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	if _, err := r.queries.GetUserByEmail(ctx, email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query user by email: %w", err)
	}
	return true, nil
}

func (r *CredentialRepository) CreateCredentialUser(ctx context.Context, user identity.CredentialUser, accountID, passwordHash string) error {
	return r.withTx(ctx, func(txq *queries.Queries) error {
		if _, err := txq.CreateUser(ctx, queries.CreateUserParams{ID: user.ID, Name: user.Name, Email: user.Email}); err != nil {
			return fmt.Errorf("insert user: %w", err)
		}
		if err := txq.CreateAccount(ctx, queries.CreateAccountParams{
			ID: accountID, AccountID: user.ID, ProviderID: "credential", UserID: user.ID, Password: &passwordHash,
		}); err != nil {
			return fmt.Errorf("insert credential account: %w", err)
		}
		return nil
	})
}

func (r *CredentialRepository) FindUserByEmail(ctx context.Context, email string) (identity.CredentialUser, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.CredentialUser{}, identity.ErrUserNotFound
		}
		return identity.CredentialUser{}, fmt.Errorf("query user by email: %w", err)
	}
	var banExpires *time.Time
	if row.BanExpires.Valid {
		value := row.BanExpires.Time
		banExpires = &value
	}
	return identity.CredentialUser{
		ID: row.ID, Name: row.Name, Email: row.Email, EmailVerified: row.EmailVerified,
		Image: row.Image, Role: row.Role,
		Ban: identity.Ban{Enabled: row.Banned != nil && *row.Banned, ExpiresAt: banExpires},
	}, nil
}

func (r *CredentialRepository) CredentialPassword(ctx context.Context, userID string) (string, error) {
	account, err := r.queries.GetAccountByProviderAndUser(ctx, queries.GetAccountByProviderAndUserParams{ProviderID: "credential", UserID: userID})
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

func (r *CredentialRepository) CreateVerification(ctx context.Context, verification identity.Verification) error {
	if err := r.queries.CreateVerification(ctx, queries.CreateVerificationParams{
		ID: verification.ID, Identifier: verification.Identifier, Value: verification.Token,
		ExpiresAt: pgtype.Timestamp{Time: verification.ExpiresAt, Valid: true},
	}); err != nil {
		return fmt.Errorf("insert verification: %w", err)
	}
	return nil
}

func (r *CredentialRepository) FindVerification(ctx context.Context, token string) (identity.Verification, error) {
	row, err := r.queries.GetVerificationByValue(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.Verification{}, identity.ErrInvalidVerification
		}
		return identity.Verification{}, fmt.Errorf("query verification: %w", err)
	}
	return identity.Verification{ID: row.ID, Identifier: row.Identifier, Token: row.Value, ExpiresAt: row.ExpiresAt.Time}, nil
}

func (r *CredentialRepository) VerifyEmail(ctx context.Context, userID, verificationID string) error {
	return r.withTx(ctx, func(txq *queries.Queries) error {
		if err := txq.UpdateUserEmailVerified(ctx, userID); err != nil {
			return fmt.Errorf("mark user email verified: %w", err)
		}
		if err := txq.DeleteVerification(ctx, verificationID); err != nil {
			return fmt.Errorf("delete email verification: %w", err)
		}
		return nil
	})
}

func (r *CredentialRepository) ResetPassword(ctx context.Context, userID, verificationID, passwordHash string) error {
	return r.withTx(ctx, func(txq *queries.Queries) error {
		if err := txq.UpdateUserPassword(ctx, queries.UpdateUserPasswordParams{Password: &passwordHash, UserID: userID}); err != nil {
			return fmt.Errorf("update credential password: %w", err)
		}
		if err := txq.DeleteUserSessions(ctx, userID); err != nil {
			return fmt.Errorf("delete user sessions: %w", err)
		}
		if err := txq.DeleteVerification(ctx, verificationID); err != nil {
			return fmt.Errorf("delete password verification: %w", err)
		}
		return nil
	})
}

func (r *CredentialRepository) withTx(ctx context.Context, operation func(*queries.Queries) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := operation(r.queries.WithTx(tx)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
