package rbac

import (
	"sync"
)

type RolePrivilegesCache struct {
	mu    sync.RWMutex
	cache map[string]map[string]bool
}

// NewRolePrivilegesCache creates a new RolePrivilegesCache
func NewRolePrivilegesCache() *RolePrivilegesCache {
	return &RolePrivilegesCache{
		cache: make(map[string]map[string]bool),
	}
}

// Get retrieves the privileges for a given role ID from the cache
func (c *RolePrivilegesCache) Get(roleID string) (map[string]bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	privileges, exist := c.cache[roleID]
	if !exist {
		return nil, false
	}

	return privileges, true
}

// Set sets the privileges for a given role ID in the cache
func (c *RolePrivilegesCache) Set(roleID string, privileges map[string]bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[roleID] = privileges
}

// Delete deletes the privileges for a given role ID from the cache
func (c *RolePrivilegesCache) Delete(roleID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, roleID)
}

// ClearCache clears the cache
func (c *RolePrivilegesCache) ClearCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]map[string]bool)
}

// GetAllKeys returns all role IDs in the cache
func (c *RolePrivilegesCache) GetAllKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.cache))
	for k := range c.cache {
		keys = append(keys, k)
	}
	return keys
}
