package rbac

import (
	"reflect"
	"sync"
	"testing"
)

func TestRolePrivilegesCache_Get(t *testing.T) {
	type fields struct {
		cache map[string]map[string]bool
	}
	type args struct {
		roleID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]bool
		want1  bool
	}{
		{
			name: "existing role",
			fields: fields{
				cache: map[string]map[string]bool{
					"role1": {
						"privilege1": true,
						"privilege2": true,
					},
				},
			},
			args: args{
				roleID: "role1",
			},
			want: map[string]bool{
				"privilege1": true,
				"privilege2": true,
			},
			want1: true,
		},
		{
			name: "non-existing role",
			fields: fields{
				cache: map[string]map[string]bool{
					"role1": {
						"privilege1": true,
						"privilege2": true,
					},
				},
			},
			args: args{
				roleID: "role2",
			},
			want:  nil,
			want1: false,
		},
		{
			name: "empty cache",
			fields: fields{
				cache: map[string]map[string]bool{},
			},
			args: args{
				roleID: "role1",
			},
			want:  nil,
			want1: false,
		},
	}

	for i := range tests {
		tt := tests[i] // Use local variable to avoid copying
		t.Run(tt.name, func(t *testing.T) {
			c := &RolePrivilegesCache{
				mu:    sync.RWMutex{}, // Create a new mutex for each test
				cache: tt.fields.cache,
			}
			got, got1 := c.Get(tt.args.roleID)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RolePrivilegesCache.Get() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("RolePrivilegesCache.Get() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestRolePrivilegesCache_Set(t *testing.T) {
	type fields struct {
		cache map[string]map[string]bool
	}
	type args struct {
		roleID     string
		privileges map[string]bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		check  func(*testing.T, map[string]map[string]bool)
	}{
		{
			name: "add new role",
			fields: fields{
				cache: map[string]map[string]bool{},
			},
			args: args{
				roleID: "role1",
				privileges: map[string]bool{
					"privilege1": true,
					"privilege2": true,
				},
			},
			check: func(t *testing.T, cache map[string]map[string]bool) {
				if len(cache) != 1 {
					t.Errorf("cache should have 1 entry, got %d", len(cache))
				}
				if !reflect.DeepEqual(cache["role1"], map[string]bool{"privilege1": true, "privilege2": true}) {
					t.Errorf("cache['role1'] = %v, want %v", cache["role1"], map[string]bool{"privilege1": true, "privilege2": true})
				}
			},
		},
		{
			name: "update existing role",
			fields: fields{
				cache: map[string]map[string]bool{
					"role1": {
						"privilege1": true,
						"privilege2": true,
					},
				},
			},
			args: args{
				roleID: "role1",
				privileges: map[string]bool{
					"privilege3": true,
					"privilege4": true,
				},
			},
			check: func(t *testing.T, cache map[string]map[string]bool) {
				if len(cache) != 1 {
					t.Errorf("cache should have 1 entry, got %d", len(cache))
				}
				if !reflect.DeepEqual(cache["role1"], map[string]bool{"privilege3": true, "privilege4": true}) {
					t.Errorf("cache['role1'] = %v, want %v", cache["role1"], map[string]bool{"privilege3": true, "privilege4": true})
				}
			},
		},
		{
			name: "add role with empty privileges",
			fields: fields{
				cache: map[string]map[string]bool{},
			},
			args: args{
				roleID:     "role1",
				privileges: map[string]bool{},
			},
			check: func(t *testing.T, cache map[string]map[string]bool) {
				if len(cache) != 1 {
					t.Errorf("cache should have 1 entry, got %d", len(cache))
				}
				if len(cache["role1"]) != 0 {
					t.Errorf("cache['role1'] should be empty, got %v", cache["role1"])
				}
			},
		},
	}

	for i := range tests {
		tt := tests[i] // Use local variable to avoid copying
		t.Run(tt.name, func(t *testing.T) {
			c := &RolePrivilegesCache{
				mu:    sync.RWMutex{}, // Create a new mutex for each test
				cache: tt.fields.cache,
			}
			c.Set(tt.args.roleID, tt.args.privileges)
			tt.check(t, c.cache)
		})
	}
}

func TestRolePrivilegesCache_Delete(t *testing.T) {
	type fields struct {
		cache map[string]map[string]bool
	}
	type args struct {
		roleID string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		check  func(*testing.T, map[string]map[string]bool)
	}{
		{
			name: "delete existing role",
			fields: fields{
				cache: map[string]map[string]bool{
					"role1": {
						"privilege1": true,
						"privilege2": true,
					},
					"role2": {
						"privilege3": true,
					},
				},
			},
			args: args{
				roleID: "role1",
			},
			check: func(t *testing.T, cache map[string]map[string]bool) {
				if len(cache) != 1 {
					t.Errorf("cache should have 1 entry, got %d", len(cache))
				}
				if _, exists := cache["role1"]; exists {
					t.Errorf("cache should not contain 'role1'")
				}
				if _, exists := cache["role2"]; !exists {
					t.Errorf("cache should still contain 'role2'")
				}
			},
		},
		{
			name: "delete non-existing role",
			fields: fields{
				cache: map[string]map[string]bool{
					"role1": {
						"privilege1": true,
						"privilege2": true,
					},
				},
			},
			args: args{
				roleID: "role2",
			},
			check: func(t *testing.T, cache map[string]map[string]bool) {
				if len(cache) != 1 {
					t.Errorf("cache should have 1 entry, got %d", len(cache))
				}
				if _, exists := cache["role1"]; !exists {
					t.Errorf("cache should still contain 'role1'")
				}
			},
		},
		{
			name: "delete from empty cache",
			fields: fields{
				cache: map[string]map[string]bool{},
			},
			args: args{
				roleID: "role1",
			},
			check: func(t *testing.T, cache map[string]map[string]bool) {
				if len(cache) != 0 {
					t.Errorf("cache should be empty, got %d entries", len(cache))
				}
			},
		},
	}

	for i := range tests {
		tt := tests[i] // Use local variable to avoid copying
		t.Run(tt.name, func(t *testing.T) {
			c := &RolePrivilegesCache{
				mu:    sync.RWMutex{}, // Create a new mutex for each test
				cache: tt.fields.cache,
			}
			c.Delete(tt.args.roleID)
			tt.check(t, c.cache)
		})
	}
}

func TestRolePrivilegesCache_ClearCache(t *testing.T) {
	type fields struct {
		cache map[string]map[string]bool
	}
	tests := []struct {
		name   string
		fields fields
		check  func(*testing.T, map[string]map[string]bool)
	}{
		{
			name: "clear non-empty cache",
			fields: fields{
				cache: map[string]map[string]bool{
					"role1": {
						"privilege1": true,
						"privilege2": true,
					},
					"role2": {
						"privilege3": true,
					},
				},
			},
			check: func(t *testing.T, cache map[string]map[string]bool) {
				if len(cache) != 0 {
					t.Errorf("cache should be empty, got %d entries", len(cache))
				}
			},
		},
		{
			name: "clear empty cache",
			fields: fields{
				cache: map[string]map[string]bool{},
			},
			check: func(t *testing.T, cache map[string]map[string]bool) {
				if len(cache) != 0 {
					t.Errorf("cache should be empty, got %d entries", len(cache))
				}
			},
		},
	}

	for i := range tests {
		tt := tests[i] // Use local variable to avoid copying
		t.Run(tt.name, func(t *testing.T) {
			c := &RolePrivilegesCache{
				mu:    sync.RWMutex{}, // Create a new mutex for each test
				cache: tt.fields.cache,
			}
			c.ClearCache()
			tt.check(t, c.cache)
		})
	}
}

func TestRolePrivilegesCache_GetAllKeys(t *testing.T) {
	type fields struct {
		cache map[string]map[string]bool
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "get keys from non-empty cache",
			fields: fields{
				cache: map[string]map[string]bool{
					"role1": {
						"privilege1": true,
						"privilege2": true,
					},
					"role2": {
						"privilege3": true,
					},
				},
			},
			want: []string{"role1", "role2"},
		},
		{
			name: "get keys from empty cache",
			fields: fields{
				cache: map[string]map[string]bool{},
			},
			want: []string{},
		},
	}

	for i := range tests {
		tt := tests[i] // Use local variable to avoid copying
		t.Run(tt.name, func(t *testing.T) {
			c := &RolePrivilegesCache{
				mu:    sync.RWMutex{}, // Create a new mutex for each test
				cache: tt.fields.cache,
			}
			got := c.GetAllKeys()

			// Sort both slices to ensure consistent comparison
			if len(got) != len(tt.want) {
				t.Errorf("RolePrivilegesCache.GetAllKeys() got = %v, want %v", got, tt.want)
				return
			}

			// Create a map to check if all expected keys are present
			wantMap := make(map[string]bool)
			for _, key := range tt.want {
				wantMap[key] = true
			}

			for _, key := range got {
				if !wantMap[key] {
					t.Errorf("RolePrivilegesCache.GetAllKeys() returned unexpected key %s", key)
				}
			}
		})
	}
}
