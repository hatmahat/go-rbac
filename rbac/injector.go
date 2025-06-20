package rbac

import (
	"context"
)

// InjectContext attaches roleID, userID, and privileges into the given context
func InjectContext(ctx context.Context, roleID string, userID string, privileges map[string]bool) context.Context {
	ctx = context.WithValue(ctx, RoleIDKey, roleID)
	ctx = context.WithValue(ctx, UserIDKey, userID)
	ctx = context.WithValue(ctx, PrivilegesKey, privileges)
	return ctx
}
