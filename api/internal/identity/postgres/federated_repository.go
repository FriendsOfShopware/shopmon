package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/identity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FederatedRepository struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
}

var _ identity.FederatedRepository = (*FederatedRepository)(nil)

func NewFederatedRepository(pool *pgxpool.Pool, q *queries.Queries) *FederatedRepository {
	return &FederatedRepository{pool: pool, queries: q}
}

func (r *FederatedRepository) Provision(ctx context.Context, profile identity.FederatedProfile) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	txq := r.queries.WithTx(tx)
	if err := txq.LockFederatedAccount(ctx, queries.LockFederatedAccountParams{ProviderID: profile.Provider, AccountID: profile.AccountID}); err != nil {
		return "", fmt.Errorf("lock federated account: %w", err)
	}

	userID, err := r.resolveUser(ctx, txq, profile)
	if err != nil {
		return "", err
	}
	if profile.ImageURL != "" {
		image := profile.ImageURL
		if err := txq.UpdateUserProfile(ctx, queries.UpdateUserProfileParams{Name: profile.Name, Image: &image, ID: userID}); err != nil {
			return "", fmt.Errorf("update federated user profile: %w", err)
		}
	}
	if profile.OrganizationID != nil && *profile.OrganizationID != "" {
		member, err := txq.IsOrgMember(ctx, queries.IsOrgMemberParams{OrganizationID: *profile.OrganizationID, UserID: userID})
		if err != nil {
			return "", fmt.Errorf("check federated organization membership: %w", err)
		}
		if !member {
			if err := txq.CreateMember(ctx, queries.CreateMemberParams{
				ID: uuid.NewString(), OrganizationID: *profile.OrganizationID, UserID: userID, Role: "member",
			}); err != nil {
				return "", fmt.Errorf("create federated organization membership: %w", err)
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit transaction: %w", err)
	}
	return userID, nil
}

func (r *FederatedRepository) resolveUser(ctx context.Context, q *queries.Queries, profile identity.FederatedProfile) (string, error) {
	account, err := q.GetAccountByProviderAndAccountID(ctx, queries.GetAccountByProviderAndAccountIDParams{
		ProviderID: profile.Provider, AccountID: profile.AccountID,
	})
	if err == nil {
		return account.UserID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("query federated account: %w", err)
	}

	_, err = q.GetUserByEmail(ctx, profile.Email)
	var userID string
	switch {
	case err == nil:
		// An email address is not authorization to link another login method.
		return "", identity.ErrEmailInUse
	case errors.Is(err, pgx.ErrNoRows):
		userID = uuid.NewString()
		if _, err := q.CreateUser(ctx, queries.CreateUserParams{ID: userID, Name: profile.Name, Email: profile.Email}); err != nil {
			return "", fmt.Errorf("insert federated user: %w", err)
		}
		if err := q.UpdateUserEmailVerified(ctx, userID); err != nil {
			return "", fmt.Errorf("mark federated email verified: %w", err)
		}
	default:
		return "", fmt.Errorf("query federated user by email: %w", err)
	}

	if err := q.CreateAccount(ctx, queries.CreateAccountParams{
		ID: uuid.NewString(), AccountID: profile.AccountID, ProviderID: profile.Provider, UserID: userID,
	}); err != nil {
		return "", fmt.Errorf("insert federated account: %w", err)
	}
	return userID, nil
}

func (r *FederatedRepository) Link(ctx context.Context, userID string, profile identity.FederatedProfile) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin link transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.queries.WithTx(tx)
	// Serialize sign-in and explicit linking for the same provider identity.
	if err := q.LockFederatedAccount(ctx, queries.LockFederatedAccountParams{ProviderID: profile.Provider, AccountID: profile.AccountID}); err != nil {
		return fmt.Errorf("lock federated account: %w", err)
	}
	account, err := q.GetAccountByProviderAndAccountID(ctx, queries.GetAccountByProviderAndAccountIDParams{
		ProviderID: profile.Provider, AccountID: profile.AccountID,
	})
	switch {
	case err == nil:
		if account.UserID != userID {
			return identity.ErrFederatedAccountInUse
		}
	case errors.Is(err, pgx.ErrNoRows):
		if err := q.CreateAccount(ctx, queries.CreateAccountParams{
			ID: uuid.NewString(), ProviderID: profile.Provider, AccountID: profile.AccountID, UserID: userID,
		}); err != nil {
			return fmt.Errorf("insert linked account: %w", err)
		}
	default:
		return fmt.Errorf("query linked account: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit link transaction: %w", err)
	}
	return nil
}
