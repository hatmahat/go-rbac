package rbac

import (
	"context"
	"time"

	logger "github.com/hatmahat/go-rbac/logger"
)

type RBACService interface {
	GetRolePrivileges(ctx context.Context, roleID string) (map[string]bool, error)
	HasPrivilege(ctx context.Context, roleID string, privilege string) (bool, error)
	HasAnyPrivilege(ctx context.Context, roleID string, privilegeCodes ...string) (bool, error)
	SetNewRolePrivileges(ctx context.Context, roleID string, privileges []string) error
	DeleteRolePrivileges(ctx context.Context, roleID string) error
}

type rbacService struct {
	repo  PrivilegeRepository // decoupled abstraction
	cache *RolePrivilegesCache
}

// NewRBACService creates a new RBAC service
func NewRBACService(repo PrivilegeRepository, refreshInterval time.Duration) RBACService {
	service := &rbacService{
		repo:  repo,
		cache: NewRolePrivilegesCache(),
	}

	// Start periodic refresh if interval is greater than 0
	if refreshInterval > 0 {
		go service.startPeriodicRefresh(refreshInterval)
	}

	return service
}

// loadRolePrivileges loads the privileges for a given role ID from the database
// and caches them
func (s *rbacService) loadRolePrivileges(ctx context.Context, roleID string) (map[string]bool, error) {

	privileges, err := s.repo.FetchPrivilegesByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	s.cache.Set(roleID, privileges)

	return privileges, nil
}

// startPeriodicRefresh is a private method that refreshes role privileges at regular intervals
func (s *rbacService) startPeriodicRefresh(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		// Get all role IDs from cache
		roleIDs := s.cache.GetAllKeys()

		// Refresh each role's privileges
		for _, roleID := range roleIDs {
			ctx := context.Background()
			_, err := s.loadRolePrivileges(ctx, roleID)
			if err != nil {
				// Log error but continue with other roles
				logger.Errorw("Error refreshing privileges for role",
					"error", err.Error(),
					"roleID", roleID,
				)
			}
		}
	}
}

// GetRolePrivileges returns the privileges for a given role ID
// It first checks the cache, if not found, it loads the privileges from the database
// and then caches them
func (s *rbacService) GetRolePrivileges(ctx context.Context, roleID string) (map[string]bool, error) {

	privileges, exist := s.cache.Get(roleID)
	if !exist {
		var err error
		privileges, err = s.loadRolePrivileges(ctx, roleID)
		if err != nil {
			return nil, err
		}
		return privileges, nil
	}

	return privileges, nil
}

// HasPrivilege checks if a given role has a specific privilege
func (s *rbacService) HasPrivilege(ctx context.Context, roleID string, privilege string) (bool, error) {

	privileges, err := s.GetRolePrivileges(ctx, roleID)
	if err != nil {
		return false, err
	}

	return privileges[privilege], nil
}

// HasAnyPrivilege checks if a given role has any of the specified privileges
func (s *rbacService) HasAnyPrivilege(ctx context.Context, roleID string, privilegeCodes ...string) (bool, error) {

	privileges, err := s.GetRolePrivileges(ctx, roleID)
	if err != nil {
		return false, err
	}

	for _, code := range privilegeCodes {
		if privileges[code] {
			return true, nil
		}
	}

	return false, nil
}

// SetNewRolePrivileges sets the privileges for a new role
func (s *rbacService) SetNewRolePrivileges(ctx context.Context, roleID string, privileges []string) error {

	privilegesMap := make(map[string]bool)
	for _, privilege := range privileges {
		privilegesMap[privilege] = true
	}

	s.cache.Set(roleID, privilegesMap)

	return nil
}

// DeleteRolePrivileges removes a role's privileges from the cache
func (s *rbacService) DeleteRolePrivileges(ctx context.Context, roleID string) error {
	s.cache.Delete(roleID)
	return nil
}
