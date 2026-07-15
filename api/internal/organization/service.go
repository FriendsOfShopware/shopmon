package organization

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvitationNotFound      = errors.New("organization invitation not found")
	ErrInvitationEmailMismatch = errors.New("organization invitation belongs to another email")
	ErrInvalidRole             = errors.New("invalid organization role")
	ErrCannotGrantRole         = errors.New("cannot grant equal or higher organization role")
	ErrCannotRemoveRole        = errors.New("cannot remove equal or higher organization role")
	ErrOwnerRequired           = errors.New("organization owner role required")
	ErrLastOwner               = errors.New("last organization owner cannot leave")
)

const invitationLifetime = 7 * 24 * time.Hour

type Organization struct {
	ID        string
	Name      string
	Slug      string
	Logo      *string
	CreatedAt time.Time
}

type Member struct {
	ID        string
	UserID    string
	Role      Role
	Name      string
	Email     string
	Image     *string
	CreatedAt time.Time
}

type Invitation struct {
	ID             string
	OrganizationID string
	Email          string
	Role           *Role
	Status         string
	ExpiresAt      time.Time
	CreatedAt      time.Time
	InviterID      string
	InviterName    string
}

type Membership struct {
	Organization Organization
	Role         Role
}

type FullOrganization struct {
	Organization Organization
	Members      []Member
	Invitations  []Invitation
}

type Repository interface {
	CreateWithOwner(ctx context.Context, organization Organization, memberID, ownerID, sessionToken string) error
	Find(ctx context.Context, organizationID string) (Organization, error)
	Update(ctx context.Context, organization Organization) error
	Delete(ctx context.Context, organizationID string) error
	CreateInvitation(ctx context.Context, invitation Invitation) error
	FindInvitation(ctx context.Context, invitationID string) (Invitation, error)
	UpdateInvitationStatus(ctx context.Context, invitationID, status string) error
	AcceptInvitation(ctx context.Context, invitationID, memberID, organizationID, userID string, role Role) error
	DeleteMember(ctx context.Context, organizationID, userID string) error
	CountOwners(ctx context.Context, organizationID string) (int32, error)
	UpdateMemberRole(ctx context.Context, organizationID, userID string, role Role) error
	ListMembers(ctx context.Context, organizationID string) ([]Member, error)
	ListInvitations(ctx context.Context, organizationID string) ([]Invitation, error)
	ListForUser(ctx context.Context, userID string) ([]Membership, error)
	DeleteInvitation(ctx context.Context, invitationID string) error
	SetActiveOrganization(ctx context.Context, sessionToken, organizationID string) error
}

type OrganizationAuthorizer interface {
	RequireOrganization(ctx context.Context, organizationID string) error
	Membership(ctx context.Context, userID, organizationID string) (Role, error)
	RequireRole(ctx context.Context, userID, organizationID string, allowedRoles ...Role) (Role, error)
}

type InvitationMailer interface {
	SendInvitation(ctx context.Context, recipient, inviterName, organizationName, invitationID string) error
}

type Service struct {
	repository Repository
	authorizer OrganizationAuthorizer
	mailer     InvitationMailer
	now        func() time.Time
}

func NewService(repository Repository, authorizer OrganizationAuthorizer, mailer InvitationMailer) *Service {
	return &Service{
		repository: repository,
		authorizer: authorizer,
		mailer:     mailer,
		now:        time.Now,
	}
}

type CreateCommand struct {
	UserID       string
	SessionToken string
	Name         string
}

func (s *Service) Create(ctx context.Context, command CreateCommand) (Organization, error) {
	organization := Organization{
		ID:   uuid.NewString(),
		Name: command.Name,
	}
	organization.Slug = organization.ID
	if err := s.repository.CreateWithOwner(ctx, organization, uuid.NewString(), command.UserID, command.SessionToken); err != nil {
		return Organization{}, fmt.Errorf("create organization with owner: %w", err)
	}
	return organization, nil
}

type UpdateCommand struct {
	UserID         string
	OrganizationID string
	Name           *string
	Logo           *string
}

func (s *Service) Update(ctx context.Context, command UpdateCommand) error {
	if _, err := s.authorizer.RequireRole(ctx, command.UserID, command.OrganizationID, RoleOwner, RoleAdmin); err != nil {
		return err
	}

	organization, err := s.repository.Find(ctx, command.OrganizationID)
	if err != nil {
		return fmt.Errorf("find organization for update: %w", err)
	}
	if command.Name != nil {
		organization.Name = *command.Name
	}
	if command.Logo != nil {
		organization.Logo = command.Logo
	}
	if err := s.repository.Update(ctx, organization); err != nil {
		return fmt.Errorf("update organization: %w", err)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, userID, organizationID string) error {
	if _, err := s.authorizer.RequireRole(ctx, userID, organizationID, RoleOwner); err != nil {
		if errors.Is(err, ErrMembershipNotFound) || errors.Is(err, ErrRoleNotAllowed) {
			return ErrOwnerRequired
		}
		return err
	}
	if err := s.repository.Delete(ctx, organizationID); err != nil {
		return fmt.Errorf("delete organization: %w", err)
	}
	return nil
}

type InviteCommand struct {
	OrganizationID string
	InviterID      string
	InviterName    string
	Email          string
	Role           string
}

func (s *Service) Invite(ctx context.Context, command InviteCommand) (string, error) {
	requestedRole, ok := ParseRole(command.Role)
	if !ok {
		return "", ErrInvalidRole
	}

	callerRole, err := s.authorizer.RequireRole(ctx, command.InviterID, command.OrganizationID, RoleOwner, RoleAdmin)
	if err != nil {
		return "", err
	}
	if !callerRole.CanGrant(requestedRole) {
		return "", ErrCannotGrantRole
	}

	organization, err := s.repository.Find(ctx, command.OrganizationID)
	if err != nil {
		return "", fmt.Errorf("find organization for invitation: %w", err)
	}
	invitationID := uuid.NewString()
	role := requestedRole
	if err := s.repository.CreateInvitation(ctx, Invitation{
		ID:             invitationID,
		OrganizationID: command.OrganizationID,
		Email:          command.Email,
		Role:           &role,
		Status:         "pending",
		ExpiresAt:      s.now().Add(invitationLifetime),
		InviterID:      command.InviterID,
	}); err != nil {
		return "", fmt.Errorf("create organization invitation: %w", err)
	}

	if s.mailer != nil {
		if err := s.mailer.SendInvitation(ctx, command.Email, command.InviterName, organization.Name, invitationID); err != nil {
			// Invitation persistence is authoritative. A mail delivery failure is
			// observable but cannot safely roll the row back after it was created.
			slog.ErrorContext(ctx, "failed to send organization invitation", "organizationID", command.OrganizationID, "invitationID", invitationID, "email", command.Email, "error", err)
		}
	}
	return invitationID, nil
}

func (s *Service) AcceptInvitation(ctx context.Context, invitationID, userID, userEmail string) error {
	invitation, err := s.repository.FindInvitation(ctx, invitationID)
	if err != nil {
		return fmt.Errorf("find invitation for acceptance: %w", err)
	}
	if invitation.Email != userEmail {
		return ErrInvitationEmailMismatch
	}

	role := RoleMember
	if invitation.Role != nil {
		role = *invitation.Role
	}
	if err := s.repository.AcceptInvitation(ctx, invitation.ID, uuid.NewString(), invitation.OrganizationID, userID, role); err != nil {
		return fmt.Errorf("accept organization invitation: %w", err)
	}
	return nil
}

func (s *Service) RejectInvitation(ctx context.Context, invitationID, userEmail string) error {
	invitation, err := s.repository.FindInvitation(ctx, invitationID)
	if err != nil {
		return fmt.Errorf("find invitation for rejection: %w", err)
	}
	if invitation.Email != userEmail {
		return ErrInvitationEmailMismatch
	}
	if err := s.repository.UpdateInvitationStatus(ctx, invitationID, "rejected"); err != nil {
		return fmt.Errorf("reject organization invitation: %w", err)
	}
	return nil
}

func (s *Service) RemoveMember(ctx context.Context, actorUserID, organizationID, targetUserID string) error {
	callerRole, err := s.authorizer.RequireRole(ctx, actorUserID, organizationID, RoleOwner, RoleAdmin)
	if err != nil {
		return err
	}
	targetRole, err := s.authorizer.Membership(ctx, targetUserID, organizationID)
	if err != nil && !errors.Is(err, ErrMembershipNotFound) {
		return fmt.Errorf("find target organization role: %w", err)
	}
	if err == nil && !callerRole.CanRemove(targetRole) {
		return ErrCannotRemoveRole
	}
	if err := s.repository.DeleteMember(ctx, organizationID, targetUserID); err != nil {
		return fmt.Errorf("remove organization member: %w", err)
	}
	return nil
}

func (s *Service) Leave(ctx context.Context, userID, organizationID string) error {
	role, err := s.authorizer.Membership(ctx, userID, organizationID)
	if err != nil {
		return err
	}
	if role == RoleOwner {
		ownerCount, err := s.repository.CountOwners(ctx, organizationID)
		if err != nil {
			return fmt.Errorf("count organization owners: %w", err)
		}
		if ownerCount <= 1 {
			return ErrLastOwner
		}
	}
	if err := s.repository.DeleteMember(ctx, organizationID, userID); err != nil {
		return fmt.Errorf("leave organization: %w", err)
	}
	return nil
}

func (s *Service) SetMemberRole(ctx context.Context, actorUserID, organizationID, targetUserID, value string) error {
	role, ok := ParseRole(value)
	if !ok {
		return ErrInvalidRole
	}
	if _, err := s.authorizer.RequireRole(ctx, actorUserID, organizationID, RoleOwner); err != nil {
		if errors.Is(err, ErrMembershipNotFound) || errors.Is(err, ErrRoleNotAllowed) {
			return ErrOwnerRequired
		}
		return err
	}
	if err := s.repository.UpdateMemberRole(ctx, organizationID, targetUserID, role); err != nil {
		return fmt.Errorf("update organization member role: %w", err)
	}
	return nil
}

func (s *Service) ListMembers(ctx context.Context, userID, organizationID string) ([]Member, error) {
	if err := s.requireMembership(ctx, userID, organizationID); err != nil {
		return nil, err
	}
	members, err := s.repository.ListMembers(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list organization members: %w", err)
	}
	return members, nil
}

func (s *Service) ListInvitations(ctx context.Context, userID, organizationID string) ([]Invitation, error) {
	if _, err := s.authorizer.RequireRole(ctx, userID, organizationID, RoleOwner, RoleAdmin); err != nil {
		return nil, err
	}
	invitations, err := s.repository.ListInvitations(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list organization invitations: %w", err)
	}
	return invitations, nil
}

func (s *Service) Full(ctx context.Context, userID, organizationID string) (FullOrganization, error) {
	if err := s.requireMembership(ctx, userID, organizationID); err != nil {
		return FullOrganization{}, err
	}
	record, err := s.repository.Find(ctx, organizationID)
	if err != nil {
		return FullOrganization{}, fmt.Errorf("find full organization: %w", err)
	}
	members, err := s.repository.ListMembers(ctx, organizationID)
	if err != nil {
		return FullOrganization{}, fmt.Errorf("list full organization members: %w", err)
	}
	invitations, err := s.repository.ListInvitations(ctx, organizationID)
	if err != nil {
		return FullOrganization{}, fmt.Errorf("list full organization invitations: %w", err)
	}
	return FullOrganization{Organization: record, Members: members, Invitations: invitations}, nil
}

func (s *Service) ListForUser(ctx context.Context, userID string) ([]Membership, error) {
	memberships, err := s.repository.ListForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list organizations for user: %w", err)
	}
	return memberships, nil
}

func (s *Service) HasAdministrativePermission(ctx context.Context, userID, organizationID string) (bool, error) {
	_, err := s.authorizer.RequireRole(ctx, userID, organizationID, RoleOwner, RoleAdmin)
	if errors.Is(err, ErrMembershipNotFound) || errors.Is(err, ErrRoleNotAllowed) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("authorize organization administration: %w", err)
	}
	return true, nil
}

func (s *Service) CancelInvitation(ctx context.Context, userID, invitationID string) error {
	invitation, err := s.repository.FindInvitation(ctx, invitationID)
	if err != nil {
		return fmt.Errorf("find invitation for cancellation: %w", err)
	}
	if _, err := s.authorizer.RequireRole(ctx, userID, invitation.OrganizationID, RoleOwner, RoleAdmin); err != nil {
		return err
	}
	if err := s.repository.DeleteInvitation(ctx, invitationID); err != nil {
		return fmt.Errorf("cancel organization invitation: %w", err)
	}
	return nil
}

func (s *Service) SetActive(ctx context.Context, userID, sessionToken, organizationID string) error {
	if err := s.requireMembership(ctx, userID, organizationID); err != nil {
		return err
	}
	if err := s.repository.SetActiveOrganization(ctx, sessionToken, organizationID); err != nil {
		return fmt.Errorf("set active organization: %w", err)
	}
	return nil
}

// EnsureActive selects the user's first organization when a session does not
// yet have one. It returns the existing value unchanged when already set.
func (s *Service) EnsureActive(ctx context.Context, userID, sessionToken string, current *string) (*string, error) {
	if current != nil {
		return current, nil
	}
	memberships, err := s.repository.ListForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list organizations for active selection: %w", err)
	}
	if len(memberships) == 0 {
		return nil, nil
	}
	organizationID := memberships[0].Organization.ID
	if err := s.repository.SetActiveOrganization(ctx, sessionToken, organizationID); err != nil {
		return nil, fmt.Errorf("persist selected active organization: %w", err)
	}
	return &organizationID, nil
}

func (s *Service) requireMembership(ctx context.Context, userID, organizationID string) error {
	if err := s.authorizer.RequireOrganization(ctx, organizationID); err != nil {
		return err
	}
	if _, err := s.authorizer.Membership(ctx, userID, organizationID); err != nil {
		return err
	}
	return nil
}
