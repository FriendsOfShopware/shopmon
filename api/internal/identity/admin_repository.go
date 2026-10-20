package identity

import (
	"context"
	"errors"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type adminStore struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
}

func (r *adminStore) ListUsers(ctx context.Context, filter AdminUserFilter) ([]AdminUser, error) {
	rows, err := r.queries.AdminListUsers(ctx, queries.AdminListUsersParams{
		Limit: filter.Limit, Offset: filter.Offset, Search: filter.Search,
		Role: filter.Role, Status: filter.Status, SortBy: filter.SortBy, SortDir: filter.SortDirection,
	})
	if err != nil {
		return nil, fmt.Errorf("query admin users: %w", err)
	}
	users := make([]AdminUser, 0, len(rows))
	for _, row := range rows {
		users = append(users, AdminUser{
			ID: row.ID, Name: row.Name, Email: row.Email, EmailVerified: row.EmailVerified,
			Image: row.Image, Role: row.Role, Banned: row.Banned, BanReason: row.BanReason,
			CreatedAt: row.CreatedAt.Time,
		})
	}
	return users, nil
}

func (r *adminStore) CountUsers(ctx context.Context, filter AdminUserFilter) (int32, error) {
	count, err := r.queries.AdminCountUsers(ctx, queries.AdminCountUsersParams{Search: filter.Search, Role: filter.Role, Status: filter.Status})
	if err != nil {
		return 0, fmt.Errorf("query admin user count: %w", err)
	}
	return count, nil
}

func (r *adminStore) FindUser(ctx context.Context, userID string) (AdminUserDetail, error) {
	row, err := r.queries.AdminGetUserDetail(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdminUserDetail{}, ErrUserNotFound
		}
		return AdminUserDetail{}, fmt.Errorf("query admin user detail: %w", err)
	}
	detail := AdminUserDetail{
		AdminUser: AdminUser{
			ID: row.ID, Name: row.Name, Email: row.Email, EmailVerified: row.EmailVerified,
			Image: row.Image, Role: row.Role, Banned: row.Banned, BanReason: row.BanReason,
			CreatedAt: row.CreatedAt.Time,
		},
		UpdatedAt: row.UpdatedAt.Time,
	}
	if row.BanExpires.Valid {
		value := row.BanExpires.Time
		detail.BanExpires = &value
	}
	return detail, nil
}

func (r *adminStore) ListUserAuthProviders(ctx context.Context, userID string) ([]AdminAuthProvider, error) {
	rows, err := r.queries.AdminListUserAccounts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query user auth providers: %w", err)
	}
	result := make([]AdminAuthProvider, 0, len(rows))
	for _, row := range rows {
		result = append(result, AdminAuthProvider{ID: row.ID, Provider: row.ProviderID, AccountID: row.AccountID, CreatedAt: row.CreatedAt.Time})
	}
	return result, nil
}

func (r *adminStore) ListUserSessions(ctx context.Context, userID string) ([]AdminSession, error) {
	rows, err := r.queries.AdminListUserSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query user sessions: %w", err)
	}
	result := make([]AdminSession, 0, len(rows))
	for _, row := range rows {
		result = append(result, AdminSession{
			ID: row.ID, IPAddress: row.IpAddress, UserAgent: row.UserAgent,
			ImpersonatedBy: row.ImpersonatedBy, CreatedAt: row.CreatedAt.Time, ExpiresAt: row.ExpiresAt.Time,
		})
	}
	return result, nil
}

func (r *adminStore) ListUserMemberships(ctx context.Context, userID string) ([]AdminMembership, error) {
	rows, err := r.queries.AdminListUserMemberships(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query user memberships: %w", err)
	}
	result := make([]AdminMembership, 0, len(rows))
	for _, row := range rows {
		result = append(result, AdminMembership{
			OrganizationID: row.OrganizationID, OrganizationName: row.OrganizationName,
			OrganizationSlug: row.OrganizationSlug, Role: row.Role, CreatedAt: row.CreatedAt.Time,
		})
	}
	return result, nil
}

func (r *adminStore) ListUserAuditLog(ctx context.Context, userID string, limit int32) ([]AdminAuditEntry, error) {
	rows, err := r.queries.AdminListUserAuditLog(ctx, queries.AdminListUserAuditLogParams{ActorUserID: &userID, Limit: limit})
	if err != nil {
		return nil, fmt.Errorf("query user audit log: %w", err)
	}
	result := make([]AdminAuditEntry, 0, len(rows))
	for _, row := range rows {
		result = append(result, AdminAuditEntry{
			ID: row.ID, ActorUserID: row.ActorUserID, ActorName: row.ActorName, ActorEmail: row.ActorEmail,
			Action: row.Action, TargetUserID: row.TargetUserID, TargetName: row.TargetName,
			TargetEmail: row.TargetEmail, Detail: row.Detail, IPAddress: row.IpAddress, CreatedAt: row.CreatedAt.Time,
		})
	}
	return result, nil
}

func (r *adminStore) SetUserRole(ctx context.Context, userID, role string) error {
	if err := r.queries.AdminSetUserRole(ctx, queries.AdminSetUserRoleParams{Role: role, ID: userID}); err != nil {
		return fmt.Errorf("execute user role update: %w", err)
	}
	return nil
}

func (r *adminStore) BanUserAndDeleteSessions(ctx context.Context, userID string, reason *string) error {
	return r.withTx(ctx, func(txq *queries.Queries) error {
		if err := txq.AdminBanUser(ctx, queries.AdminBanUserParams{BanReason: reason, ID: userID}); err != nil {
			return fmt.Errorf("execute user ban: %w", err)
		}
		if err := txq.DeleteUserSessions(ctx, userID); err != nil {
			return fmt.Errorf("delete banned user sessions: %w", err)
		}
		return nil
	})
}

func (r *adminStore) UnbanUser(ctx context.Context, userID string) error {
	if err := r.queries.AdminUnbanUser(ctx, userID); err != nil {
		return fmt.Errorf("execute user unban: %w", err)
	}
	return nil
}

func (r *adminStore) UserExists(ctx context.Context, userID string) (bool, error) {
	if _, err := r.queries.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query user: %w", err)
	}
	return true, nil
}

func (r *adminStore) CreateImpersonationSession(ctx context.Context, session AdminImpersonationSession) error {
	impersonatedBy := session.ImpersonatedBy
	if _, err := r.queries.AdminCreateImpersonationSession(ctx, queries.AdminCreateImpersonationSessionParams{
		ID: session.ID, ExpiresAt: pgtype.Timestamp{Time: session.ExpiresAt, Valid: true}, Token: session.Token,
		IpAddress: session.IPAddress, UserAgent: session.UserAgent, UserID: session.UserID,
		ImpersonatedBy: &impersonatedBy,
	}); err != nil {
		return fmt.Errorf("insert impersonation session: %w", err)
	}
	return nil
}

func (r *adminStore) withTx(ctx context.Context, operation func(*queries.Queries) error) error {
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
