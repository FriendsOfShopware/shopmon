package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/identity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminRepository struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
}

var _ identity.AdminRepository = (*AdminRepository)(nil)

func NewAdminRepository(pool *pgxpool.Pool, q *queries.Queries) *AdminRepository {
	return &AdminRepository{pool: pool, queries: q}
}

func (r *AdminRepository) ListUsers(ctx context.Context, filter identity.AdminUserFilter) ([]identity.AdminUser, error) {
	rows, err := r.queries.AdminListUsers(ctx, queries.AdminListUsersParams{
		Limit: filter.Limit, Offset: filter.Offset, Search: filter.Search,
		Role: filter.Role, Status: filter.Status, SortBy: filter.SortBy, SortDir: filter.SortDirection,
	})
	if err != nil {
		return nil, fmt.Errorf("query admin users: %w", err)
	}
	users := make([]identity.AdminUser, 0, len(rows))
	for _, row := range rows {
		users = append(users, identity.AdminUser{
			ID: row.ID, Name: row.Name, Email: row.Email, EmailVerified: row.EmailVerified,
			Image: row.Image, Role: row.Role, Banned: row.Banned, BanReason: row.BanReason,
			CreatedAt: row.CreatedAt.Time,
		})
	}
	return users, nil
}

func (r *AdminRepository) CountUsers(ctx context.Context, filter identity.AdminUserFilter) (int32, error) {
	count, err := r.queries.AdminCountUsers(ctx, queries.AdminCountUsersParams{Search: filter.Search, Role: filter.Role, Status: filter.Status})
	if err != nil {
		return 0, fmt.Errorf("query admin user count: %w", err)
	}
	return count, nil
}

func (r *AdminRepository) FindUser(ctx context.Context, userID string) (identity.AdminUserDetail, error) {
	row, err := r.queries.AdminGetUserDetail(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return identity.AdminUserDetail{}, identity.ErrUserNotFound
		}
		return identity.AdminUserDetail{}, fmt.Errorf("query admin user detail: %w", err)
	}
	detail := identity.AdminUserDetail{
		AdminUser: identity.AdminUser{
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

func (r *AdminRepository) ListUserAuthProviders(ctx context.Context, userID string) ([]identity.AdminAuthProvider, error) {
	rows, err := r.queries.AdminListUserAccounts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query user auth providers: %w", err)
	}
	result := make([]identity.AdminAuthProvider, 0, len(rows))
	for _, row := range rows {
		result = append(result, identity.AdminAuthProvider{ID: row.ID, Provider: row.ProviderID, AccountID: row.AccountID, CreatedAt: row.CreatedAt.Time})
	}
	return result, nil
}

func (r *AdminRepository) ListUserSessions(ctx context.Context, userID string) ([]identity.AdminSession, error) {
	rows, err := r.queries.AdminListUserSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query user sessions: %w", err)
	}
	result := make([]identity.AdminSession, 0, len(rows))
	for _, row := range rows {
		result = append(result, identity.AdminSession{
			ID: row.ID, IPAddress: row.IpAddress, UserAgent: row.UserAgent,
			ImpersonatedBy: row.ImpersonatedBy, CreatedAt: row.CreatedAt.Time, ExpiresAt: row.ExpiresAt.Time,
		})
	}
	return result, nil
}

func (r *AdminRepository) ListUserMemberships(ctx context.Context, userID string) ([]identity.AdminMembership, error) {
	rows, err := r.queries.AdminListUserMemberships(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query user memberships: %w", err)
	}
	result := make([]identity.AdminMembership, 0, len(rows))
	for _, row := range rows {
		result = append(result, identity.AdminMembership{
			OrganizationID: row.OrganizationID, OrganizationName: row.OrganizationName,
			OrganizationSlug: row.OrganizationSlug, Role: row.Role, CreatedAt: row.CreatedAt.Time,
		})
	}
	return result, nil
}

func (r *AdminRepository) ListUserAuditLog(ctx context.Context, userID string, limit int32) ([]identity.AdminAuditEntry, error) {
	rows, err := r.queries.AdminListUserAuditLog(ctx, queries.AdminListUserAuditLogParams{ActorUserID: &userID, Limit: limit})
	if err != nil {
		return nil, fmt.Errorf("query user audit log: %w", err)
	}
	result := make([]identity.AdminAuditEntry, 0, len(rows))
	for _, row := range rows {
		result = append(result, identity.AdminAuditEntry{
			ID: row.ID, ActorUserID: row.ActorUserID, ActorName: row.ActorName, ActorEmail: row.ActorEmail,
			Action: row.Action, TargetUserID: row.TargetUserID, TargetName: row.TargetName,
			TargetEmail: row.TargetEmail, Detail: row.Detail, IPAddress: row.IpAddress, CreatedAt: row.CreatedAt.Time,
		})
	}
	return result, nil
}

func (r *AdminRepository) SetUserRole(ctx context.Context, userID, role string) error {
	if err := r.queries.AdminSetUserRole(ctx, queries.AdminSetUserRoleParams{Role: role, ID: userID}); err != nil {
		return fmt.Errorf("execute user role update: %w", err)
	}
	return nil
}

func (r *AdminRepository) BanUserAndDeleteSessions(ctx context.Context, userID string, reason *string) error {
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

func (r *AdminRepository) UnbanUser(ctx context.Context, userID string) error {
	if err := r.queries.AdminUnbanUser(ctx, userID); err != nil {
		return fmt.Errorf("execute user unban: %w", err)
	}
	return nil
}

func (r *AdminRepository) UserExists(ctx context.Context, userID string) (bool, error) {
	if _, err := r.queries.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query user: %w", err)
	}
	return true, nil
}

func (r *AdminRepository) CreateImpersonationSession(ctx context.Context, session identity.AdminImpersonationSession) error {
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

func (r *AdminRepository) withTx(ctx context.Context, operation func(*queries.Queries) error) error {
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
