package rbac

import "context"

// PrivilegeRepository abstracts data fetching so you can use GORM, pgx, raw SQL, etc.
type PrivilegeRepository interface {
	FetchPrivilegesByRoleID(ctx context.Context, roleID string) (map[string]bool, error)
}
