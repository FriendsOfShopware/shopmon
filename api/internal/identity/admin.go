package identity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrCannotChangeSelf  = errors.New("cannot change own role")
	ErrInvalidSystemRole = errors.New("invalid system role")
)

const impersonationLifetime = time.Hour

type AdminUser struct {
	ID            string
	Name          string
	Email         string
	EmailVerified bool
	Image         *string
	Role          string
	Banned        *bool
	BanReason     *string
	CreatedAt     time.Time
}

type AdminUserFilter struct {
	Limit         int32
	Offset        int32
	Search        *string
	Role          *string
	Status        *string
	SortBy        string
	SortDirection string
}

type AdminUserPage struct {
	Users []AdminUser
	Total int32
}

type AdminAuthProvider struct {
	ID        string
	Provider  string
	AccountID string
	CreatedAt time.Time
}

type AdminSession struct {
	ID             string
	IPAddress      *string
	UserAgent      *string
	ImpersonatedBy *string
	CreatedAt      time.Time
	ExpiresAt      time.Time
}

type AdminMembership struct {
	OrganizationID   string
	OrganizationName string
	OrganizationSlug string
	Role             string
	CreatedAt        time.Time
}

type AdminAuditEntry struct {
	ID           int64
	ActorUserID  *string
	ActorName    *string
	ActorEmail   *string
	Action       string
	TargetUserID *string
	TargetName   *string
	TargetEmail  *string
	Detail       *string
	IPAddress    *string
	CreatedAt    time.Time
}

type AdminUserDetail struct {
	AdminUser
	BanExpires    *time.Time
	UpdatedAt     time.Time
	AuthProviders []AdminAuthProvider
	Sessions      []AdminSession
	Memberships   []AdminMembership
	AuditLog      []AdminAuditEntry
}

type AdminRepository interface {
	ListUsers(ctx context.Context, filter AdminUserFilter) ([]AdminUser, error)
	CountUsers(ctx context.Context, filter AdminUserFilter) (int32, error)
	FindUser(ctx context.Context, userID string) (AdminUserDetail, error)
	ListUserAuthProviders(ctx context.Context, userID string) ([]AdminAuthProvider, error)
	ListUserSessions(ctx context.Context, userID string) ([]AdminSession, error)
	ListUserMemberships(ctx context.Context, userID string) ([]AdminMembership, error)
	ListUserAuditLog(ctx context.Context, userID string, limit int32) ([]AdminAuditEntry, error)
	SetUserRole(ctx context.Context, userID, role string) error
	BanUserAndDeleteSessions(ctx context.Context, userID string, reason *string) error
	UnbanUser(ctx context.Context, userID string) error
	UserExists(ctx context.Context, userID string) (bool, error)
	CreateImpersonationSession(ctx context.Context, session AdminImpersonationSession) error
}

type AdminImpersonationSession struct {
	ID             string
	Token          string
	ExpiresAt      time.Time
	IPAddress      *string
	UserAgent      *string
	UserID         string
	ImpersonatedBy string
}

type AdminService struct {
	repository AdminRepository
	now        func() time.Time
}

func NewAdminService(repository AdminRepository) *AdminService {
	return &AdminService{repository: repository, now: time.Now}
}

func (s *AdminService) ListUsers(ctx context.Context, filter AdminUserFilter) (AdminUserPage, error) {
	var users []AdminUser
	var total int32
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		var err error
		users, err = s.repository.ListUsers(groupCtx, filter)
		if err != nil {
			return fmt.Errorf("list admin users: %w", err)
		}
		return nil
	})
	group.Go(func() error {
		var err error
		total, err = s.repository.CountUsers(groupCtx, filter)
		if err != nil {
			return fmt.Errorf("count admin users: %w", err)
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		return AdminUserPage{}, err
	}
	return AdminUserPage{Users: users, Total: total}, nil
}

func (s *AdminService) UserDetail(ctx context.Context, userID string) (AdminUserDetail, error) {
	detail, err := s.repository.FindUser(ctx, userID)
	if err != nil {
		return AdminUserDetail{}, fmt.Errorf("find admin user detail: %w", err)
	}

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		var err error
		detail.AuthProviders, err = s.repository.ListUserAuthProviders(groupCtx, userID)
		return wrapAdminReadError("list user auth providers", err)
	})
	group.Go(func() error {
		var err error
		detail.Sessions, err = s.repository.ListUserSessions(groupCtx, userID)
		return wrapAdminReadError("list user sessions", err)
	})
	group.Go(func() error {
		var err error
		detail.Memberships, err = s.repository.ListUserMemberships(groupCtx, userID)
		return wrapAdminReadError("list user memberships", err)
	})
	group.Go(func() error {
		var err error
		detail.AuditLog, err = s.repository.ListUserAuditLog(groupCtx, userID, 50)
		return wrapAdminReadError("list user audit log", err)
	})
	if err := group.Wait(); err != nil {
		return AdminUserDetail{}, err
	}
	return detail, nil
}

func (s *AdminService) SetUserRole(ctx context.Context, actorUserID, targetUserID, role string) error {
	if actorUserID == targetUserID {
		return ErrCannotChangeSelf
	}
	if role != "user" && role != "admin" {
		return ErrInvalidSystemRole
	}
	if err := s.repository.SetUserRole(ctx, targetUserID, role); err != nil {
		return fmt.Errorf("set user role: %w", err)
	}
	return nil
}

func (s *AdminService) BanUser(ctx context.Context, userID string, reason *string) error {
	if err := s.repository.BanUserAndDeleteSessions(ctx, userID, reason); err != nil {
		return fmt.Errorf("ban user and invalidate sessions: %w", err)
	}
	return nil
}

func (s *AdminService) UnbanUser(ctx context.Context, userID string) error {
	if err := s.repository.UnbanUser(ctx, userID); err != nil {
		return fmt.Errorf("unban user: %w", err)
	}
	return nil
}

type ImpersonateCommand struct {
	ActorUserID  string
	TargetUserID string
	IPAddress    string
	UserAgent    string
}

func (s *AdminService) Impersonate(ctx context.Context, command ImpersonateCommand) (string, error) {
	exists, err := s.repository.UserExists(ctx, command.TargetUserID)
	if err != nil {
		return "", fmt.Errorf("check impersonation target: %w", err)
	}
	if !exists {
		return "", ErrUserNotFound
	}
	token, err := randomToken()
	if err != nil {
		return "", fmt.Errorf("generate impersonation token: %w", err)
	}
	var ipAddress, userAgent *string
	if command.IPAddress != "" {
		ipAddress = &command.IPAddress
	}
	if command.UserAgent != "" {
		userAgent = &command.UserAgent
	}
	if err := s.repository.CreateImpersonationSession(ctx, AdminImpersonationSession{
		ID:             "imp-" + uuid.NewString(),
		Token:          token,
		ExpiresAt:      s.now().Add(impersonationLifetime),
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		UserID:         command.TargetUserID,
		ImpersonatedBy: command.ActorUserID,
	}); err != nil {
		return "", fmt.Errorf("create impersonation session: %w", err)
	}
	return token, nil
}

func wrapAdminReadError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func randomToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}
