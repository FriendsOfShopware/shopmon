package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/identity"
	"github.com/friendsofshopware/shopmon/api/internal/identity/passkey"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	queries *queries.Queries
}

var _ passkey.Repository = (*Repository)(nil)

func NewRepository(q *queries.Queries) *Repository {
	return &Repository{queries: q}
}

func (r *Repository) FindUser(ctx context.Context, userID string) (passkey.User, error) {
	row, err := r.queries.GetUserByID(ctx, userID)
	if err != nil {
		return passkey.User{}, fmt.Errorf("query passkey user: %w", err)
	}
	var banExpires *time.Time
	if row.BanExpires.Valid {
		value := row.BanExpires.Time
		banExpires = &value
	}
	return passkey.User{
		ID: row.ID, Name: row.Name, Email: row.Email,
		Ban: identity.Ban{Enabled: row.Banned != nil && *row.Banned, ExpiresAt: banExpires},
	}, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]passkey.StoredPasskey, error) {
	rows, err := r.queries.GetPasskeysByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query user passkeys: %w", err)
	}
	result := make([]passkey.StoredPasskey, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapPasskey(row))
	}
	return result, nil
}

func (r *Repository) Create(ctx context.Context, stored passkey.StoredPasskey) error {
	if err := r.queries.CreatePasskey(ctx, queries.CreatePasskeyParams{
		ID: stored.ID, Name: stored.Name, PublicKey: stored.PublicKey, UserID: stored.UserID,
		CredentialID: stored.CredentialID, Counter: stored.Counter, DeviceType: stored.DeviceType,
		BackedUp: stored.BackedUp, Transports: stored.Transports, Aaguid: stored.AAGUID,
	}); err != nil {
		return fmt.Errorf("insert passkey: %w", err)
	}
	return nil
}

func (r *Repository) FindByCredentialID(ctx context.Context, credentialID string) (passkey.StoredPasskey, error) {
	row, err := r.queries.GetPasskeyByCredentialID(ctx, credentialID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return passkey.StoredPasskey{}, passkey.ErrNotFound
		}
		return passkey.StoredPasskey{}, fmt.Errorf("query passkey by credential ID: %w", err)
	}
	return passkey.StoredPasskey{
		ID: row.ID, Name: row.Name, PublicKey: row.PublicKey, UserID: row.UserID,
		CredentialID: row.CredentialID, Counter: row.Counter, DeviceType: row.DeviceType,
		BackedUp: row.BackedUp, Transports: row.Transports, AAGUID: row.Aaguid,
	}, nil
}

func (r *Repository) UpdateCounter(ctx context.Context, passkeyID string, counter int32) error {
	if err := r.queries.UpdatePasskeyCounter(ctx, queries.UpdatePasskeyCounterParams{Counter: counter, ID: passkeyID}); err != nil {
		return fmt.Errorf("execute passkey counter update: %w", err)
	}
	return nil
}

func mapPasskey(row queries.Passkey) passkey.StoredPasskey {
	return passkey.StoredPasskey{
		ID: row.ID, Name: row.Name, PublicKey: row.PublicKey, UserID: row.UserID,
		CredentialID: row.CredentialID, Counter: row.Counter, DeviceType: row.DeviceType,
		BackedUp: row.BackedUp, Transports: row.Transports, AAGUID: row.Aaguid,
	}
}
