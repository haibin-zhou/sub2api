package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeUserRoleAcceptsOperator(t *testing.T) {
	for _, role := range []string{RoleAdmin, RoleOperator, RoleUser} {
		got, err := normalizeUserRole(role, RoleUser)
		require.NoError(t, err)
		require.Equal(t, role, got)
	}
	got, err := normalizeUserRole("", RoleUser)
	require.NoError(t, err)
	require.Equal(t, RoleUser, got)

	_, err = normalizeUserRole("superadmin", RoleUser)
	require.Error(t, err)
}
