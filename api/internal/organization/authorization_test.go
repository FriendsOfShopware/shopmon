package organization

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubAuthorizationRepository struct {
	exists    bool
	existsErr error
	role      Role
	roleErr   error
}

func (r *stubAuthorizationRepository) OrganizationExists(context.Context, string) (bool, error) {
	return r.exists, r.existsErr
}

func (r *stubAuthorizationRepository) FindRole(context.Context, string, string) (Role, error) {
	return r.role, r.roleErr
}

func TestAuthorizerRequireOrganization(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		authorizer := NewAuthorizer(&stubAuthorizationRepository{exists: true})
		require.NoError(t, authorizer.RequireOrganization(t.Context(), "org-1"))
	})

	t.Run("not found", func(t *testing.T) {
		authorizer := NewAuthorizer(&stubAuthorizationRepository{})
		require.ErrorIs(t, authorizer.RequireOrganization(t.Context(), "missing"), ErrOrganizationNotFound)
	})

	t.Run("repository error", func(t *testing.T) {
		repositoryErr := errors.New("database unavailable")
		authorizer := NewAuthorizer(&stubAuthorizationRepository{existsErr: repositoryErr})
		require.ErrorIs(t, authorizer.RequireOrganization(t.Context(), "org-1"), repositoryErr)
	})
}

func TestAuthorizerMembershipAndRole(t *testing.T) {
	t.Run("returns membership role", func(t *testing.T) {
		authorizer := NewAuthorizer(&stubAuthorizationRepository{role: RoleAdmin})
		role, err := authorizer.Membership(t.Context(), "user-1", "org-1")
		require.NoError(t, err)
		assert.Equal(t, RoleAdmin, role)
	})

	t.Run("preserves membership not found", func(t *testing.T) {
		authorizer := NewAuthorizer(&stubAuthorizationRepository{roleErr: ErrMembershipNotFound})
		_, err := authorizer.Membership(t.Context(), "user-1", "org-1")
		require.ErrorIs(t, err, ErrMembershipNotFound)
	})

	t.Run("allows configured role", func(t *testing.T) {
		authorizer := NewAuthorizer(&stubAuthorizationRepository{role: RoleAdmin})
		role, err := authorizer.RequireRole(t.Context(), "user-1", "org-1", RoleOwner, RoleAdmin)
		require.NoError(t, err)
		assert.Equal(t, RoleAdmin, role)
	})

	t.Run("rejects other role", func(t *testing.T) {
		authorizer := NewAuthorizer(&stubAuthorizationRepository{role: RoleMember})
		_, err := authorizer.RequireRole(t.Context(), "user-1", "org-1", RoleOwner, RoleAdmin)
		require.ErrorIs(t, err, ErrRoleNotAllowed)
	})
}

func TestRoleHierarchy(t *testing.T) {
	tests := []struct {
		name      string
		actor     Role
		target    Role
		canGrant  bool
		canRemove bool
	}{
		{name: "owner and owner", actor: RoleOwner, target: RoleOwner, canGrant: true},
		{name: "owner and admin", actor: RoleOwner, target: RoleAdmin, canGrant: true, canRemove: true},
		{name: "admin and admin", actor: RoleAdmin, target: RoleAdmin},
		{name: "admin and member", actor: RoleAdmin, target: RoleMember, canGrant: true, canRemove: true},
		{name: "member and member", actor: RoleMember, target: RoleMember},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.canGrant, tt.actor.CanGrant(tt.target))
			assert.Equal(t, tt.canRemove, tt.actor.CanRemove(tt.target))
		})
	}
}

func TestParseRole(t *testing.T) {
	role, ok := ParseRole("admin")
	assert.True(t, ok)
	assert.Equal(t, RoleAdmin, role)

	role, ok = ParseRole("super-admin")
	assert.False(t, ok)
	assert.Empty(t, role)
}
