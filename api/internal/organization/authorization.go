package organization

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrMembershipNotFound   = errors.New("organization membership not found")
	ErrRoleNotAllowed       = errors.New("organization role not allowed")
)

// Role is a user's authority within an organization.
type Role string

const (
	RoleMember Role = "member"
	RoleAdmin  Role = "admin"
	RoleOwner  Role = "owner"
)

// ParseRole converts an external role value into a valid organization role.
func ParseRole(value string) (Role, bool) {
	role := Role(value)
	switch role {
	case RoleMember, RoleAdmin, RoleOwner:
		return role, true
	default:
		return "", false
	}
}

// CanGrant reports whether this role may grant the target role. Owners may
// grant any valid role; other roles must be strictly higher than the target.
func (r Role) CanGrant(target Role) bool {
	if !r.valid() || !target.valid() {
		return false
	}
	return r == RoleOwner || r.weight() > target.weight()
}

// CanRemove reports whether this role may remove a member with the target role.
func (r Role) CanRemove(target Role) bool {
	return r.valid() && target.valid() && r.weight() > target.weight()
}

func (r Role) valid() bool {
	_, ok := ParseRole(string(r))
	return ok
}

func (r Role) weight() int {
	switch r {
	case RoleMember:
		return 1
	case RoleAdmin:
		return 2
	case RoleOwner:
		return 3
	default:
		return 0
	}
}

// AuthorizationRepository is the persistence needed for organization access decisions.
type AuthorizationRepository interface {
	OrganizationExists(ctx context.Context, organizationID string) (bool, error)
	FindRole(ctx context.Context, userID, organizationID string) (Role, error)
}

// Authorizer applies organization membership and role policy independently of HTTP and SQL.
type Authorizer struct {
	repository AuthorizationRepository
}

func NewAuthorizer(repository AuthorizationRepository) *Authorizer {
	return &Authorizer{repository: repository}
}

// RequireOrganization verifies that an organization exists.
func (a *Authorizer) RequireOrganization(ctx context.Context, organizationID string) error {
	exists, err := a.repository.OrganizationExists(ctx, organizationID)
	if err != nil {
		return fmt.Errorf("check organization existence: %w", err)
	}
	if !exists {
		return ErrOrganizationNotFound
	}
	return nil
}

// Membership returns the user's role or ErrMembershipNotFound.
func (a *Authorizer) Membership(ctx context.Context, userID, organizationID string) (Role, error) {
	role, err := a.repository.FindRole(ctx, userID, organizationID)
	if err != nil {
		return "", fmt.Errorf("find organization membership: %w", err)
	}
	return role, nil
}

// IsMember reports membership without exposing role details to dependent capabilities.
func (a *Authorizer) IsMember(ctx context.Context, userID, organizationID string) (bool, error) {
	_, err := a.Membership(ctx, userID, organizationID)
	if errors.Is(err, ErrMembershipNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// RequireRole returns the user's role when it is one of the allowed roles.
func (a *Authorizer) RequireRole(ctx context.Context, userID, organizationID string, allowedRoles ...Role) (Role, error) {
	role, err := a.Membership(ctx, userID, organizationID)
	if err != nil {
		return "", err
	}
	for _, allowedRole := range allowedRoles {
		if role == allowedRole {
			return role, nil
		}
	}
	return "", fmt.Errorf("%w: %s", ErrRoleNotAllowed, role)
}
