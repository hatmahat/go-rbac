package rbac

import (
	"context"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// Context keys
const (
	RoleIDKey     contextKey = "roleID"
	PrivilegesKey contextKey = "privileges"
	UserIDKey     contextKey = "userID"
	UserNameKey   contextKey = "userName"
)

// GetRoleIDFromContext retrieves the role ID from the context
func GetRoleIDFromContext(ctx context.Context) (string, bool) {
	roleID, ok := ctx.Value(RoleIDKey).(string)
	return roleID, ok
}

// GetPrivilegesFromContext retrieves the privileges map from the context
func GetPrivilegesFromContext(ctx context.Context) (map[string]bool, bool) {
	privileges, ok := ctx.Value(PrivilegesKey).(map[string]bool)
	return privileges, ok
}

// HasPrivilegeInContext checks if a specific privilege exists in the context
func HasPrivilegeInContext(ctx context.Context, privilegeCode string) bool {
	privileges, ok := GetPrivilegesFromContext(ctx)
	if !ok {
		return false
	}
	return privileges[privilegeCode]
}

// GetUserIDFromContext retrieves the user ID from the context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// GetUserNameFromContext retrieves the user name from the context
func GetUserNameFromContext(ctx context.Context) (string, bool) {
	userName, ok := ctx.Value(UserNameKey).(string)
	return userName, ok
}
