> **Disclaimer:**  
> I originally wrote this library for a specific internal project.  
> Later, I refactored and packaged it as a standalone module with the help of LLMs
> to improve the structure and documentation for reuse and open source sharing.

# go-rbac
[![Go Reference](https://pkg.go.dev/badge/github.com/hatmahat/go-rbac.svg)](https://pkg.go.dev/github.com/hatmahat/go-rbac)

A lightweight, framework-agnostic **Role-Based Access Control (RBAC)** library for Go, built with:

- Context-based privilege injection
- In-memory caching for fast access
- Optional auto-refresh of role privileges
- Minimal dependency (works with any HTTP framework or DB driver)

---

## Features

- No framework lock-in (works with Echo, Gin, Chi, Fiber, etc.)
- Decoupled data layer — bring your own database via PrivilegeRepository 
- Simple API: `HasPrivilege`, `GetRolePrivileges`, `InjectContext`
- Optional GORM-based implementation provided
- Includes built-in RBACService with cache and auto-refresh support

---

## Installation

```bash
go get github.com/hatmahat/go-rbac/rbac
```

## Folder Structure
```
go-rbac/
├── example/                    # Minimal usage example using Echo
│   └── main.go
├── rbac/                       # Core RBAC logic (framework-agnostic)
│   ├── cache.go                # In-memory cache for role privileges
│   ├── context.go              # Context keys and access helpers
│   ├── injector.go             # Inject privileges into context
│   ├── logger.go               # Optional logger (Console or Null)
│   ├── privilege_repository.go # Interface for custom DB repositories 
│   └── service.go              # Main RBAC service logic
├── rbacgorm/                   # Optional GORM-based implementation
│   └── gorm_repository.go
```
## RBAC Model: Privileges, Roles, and Users

This library uses a minimal and flexible RBAC (Role-Based Access Control) model based on three key entities:

| Type      | Description                                      | Example           |
|-----------|--------------------------------------------------|-------------------|
| Privilege | A string that defines a specific permission code | `read:users`      |
| Role      | A group of privileges assigned to a category     | `admin`           |
| User      | Assigned one or more roles to determine access   | user `123` with role `viewer` |

---

### How it works

- A **Privilege** is a string like `read:compliance`, `delete:report`, etc.
- A **Role** (e.g., `admin`, `viewer`) contains a list of such privilege codes.
- A **User** is associated with a role — usually passed in JWT claims or request headers like `X-Role-ID`.
- At runtime, the role’s privileges are injected into the request `context.Context`.
- You can check access easily with helpers like:

```go
if !rbac.HasPrivilegeInContext(ctx, "read:compliance") {
    return errors.New("forbidden")
}
```
You can use any privilege naming convention (e.g., read:users, manage:projects, export:data).
The system treats them as simple string lookups for fast in-memory evaluation.


## Quick Start 
### Step 1: Implement your own PrivilegeRepository
#### Option A: Use the built-in GORM implementation 
```go
import (
	"time"

	"github.com/hatmahat/go-rbac/rbac"
)

repo := rbac.NewGormPrivilegeRepository(db)
rbacService := rbac.NewRBACService(repo, 5*time.Minute, rbac.NewConsoleLogger()) // optional logger
```
#### Option B: Create your own repository (e.g. using database/sql)
```go
package myrepo

import (
	"context"
	"database/sql"
	"fmt"
)

type SQLPrivilegeRepository struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *SQLPrivilegeRepository {
	return &SQLPrivilegeRepository{db: db}
}

func (r *SQLPrivilegeRepository) FetchPrivilegesByRoleID(ctx context.Context, roleID string) (map[string]bool, error) {
	const query = `
		SELECT p.code
		FROM privileges p
		JOIN role_privileges rp ON p.id = rp.privilege_id
		WHERE rp.role_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	privMap := make(map[string]bool)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		privMap[code] = true
	}

	return privMap, nil
}
```
Using your repo in main.go
```go
package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/hatmahat/go-rbac/rbac"
	"your_project/myrepo"
)

func main() {
	db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	repo := myrepo.NewRepo(db)

	// Optional: use NewConsoleLogger() for dev or NewNullLogger() for silence
	rbacService := rbac.NewRBACService(repo, 5*time.Minute, rbac.NewConsoleLogger())

	// Use rbacService in your middleware, handlers, etc.
}
```
### Step 2: Inject into context 
#### Examples

#### 1. Echo
```go
e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        roleID := c.Request().Header.Get("X-Role-ID")
        userID := c.Request().Header.Get("X-User-ID")

        privileges, err := rbacService.GetRolePrivileges(c.Request().Context(), roleID)
        if err != nil {
            return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
        }

        ctx := rbac.InjectContext(c.Request().Context(), roleID, userID, privileges)
        c.SetRequest(c.Request().WithContext(ctx))
        return next(c)
    }
})
```

#### 2. Gin
```go
r.Use(func(c *gin.Context) {
    roleID := c.GetHeader("X-Role-ID")
    userID := c.GetHeader("X-User-ID")

    privileges, err := rbacService.GetRolePrivileges(c.Request.Context(), roleID)
    if err != nil {
        c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
        return
    }

    ctx := rbac.InjectContext(c.Request.Context(), roleID, userID, privileges)
    c.Request = c.Request.WithContext(ctx)
    c.Next()
})
```

#### 3. Chi
```go
r.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        roleID := r.Header.Get("X-Role-ID")
        userID := r.Header.Get("X-User-ID")

        privileges, err := rbacService.GetRolePrivileges(r.Context(), roleID)
        if err != nil {
            http.Error(w, "Forbidden", http.StatusForbidden)
            return
        }

        ctx := rbac.InjectContext(r.Context(), roleID, userID, privileges)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
})
```

#### 4. Fiber
```go
app.Use(func(c *fiber.Ctx) error {
    roleID := c.Get("X-Role-ID")
    userID := c.Get("X-User-ID")

    privileges, err := rbacService.GetRolePrivileges(c.UserContext(), roleID)
    if err != nil {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
    }

    ctx := rbac.InjectContext(c.UserContext(), roleID, userID, privileges)
    c.SetUserContext(ctx)
    return c.Next()
})
```
### About RBACService

The core `RBACService` handles:
- In-memory caching of privileges per role
- Auto-refreshing cache (if interval > 0)
- Fast lookups with `HasPrivilege`, `HasAnyPrivilege`, etc.

You don't need to manage caching or database access manually — just implement `PrivilegeRepository` and call `NewRBACService(...)`.
### RBACService Interface

The `RBACService` is the core entry point for working with roles and privileges. It provides cache-aware methods for checking access and managing privilege data:

| Function | Purpose |
|----------|-------------|
| `GetRolePrivileges(ctx, roleID)` | Returns all privilege codes assigned to a given role. Uses in-memory cache if available. |
| `HasPrivilege(ctx, roleID, privilege)` | Checks whether the given role has a specific privilege. Returns a boolean. |
| `HasAnyPrivilege(ctx, roleID, codes...)` | Returns `true` if the role has **any** of the specified privilege codes. Useful for OR-checks. |
| `SetNewRolePrivileges(ctx, roleID, privileges)` | Sets/overrides the cached privileges for a role (used during setup/testing). Does **not** persist to DB. |
| `DeleteRolePrivileges(ctx, roleID)` | Deletes the privilege cache for a role. Will force a refresh from your DB on next access. |

> All methods auto-refresh from DB if privileges are missing from cache.

## Checking Privileges in Your Handlers
Once you’ve injected RBAC context using InjectContext, you can retrieve and use the privileges easily:
```go
ctx := c.Request().Context()

privs, ok := rbac.GetPrivilegesFromContext(ctx)
if !ok || !privs["read:compliance"] {
    return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
}

userID, _ := rbac.GetUserIDFromContext(ctx)
return c.JSON(http.StatusOK, map[string]string{
    "message": fmt.Sprintf("Hello user %s! You have access.", userID),
})
```
### Example in Business Logic Layer (Service)
```go
func (s *YourService) GetData(ctx context.Context) error {
    if !rbac.HasPrivilegeInContext(ctx, "read:data") {
        return fmt.Errorf("forbidden")
    }

    userID, _ := rbac.GetUserIDFromContext(ctx)
    fmt.Println("Fetching data for user:", userID)

    // continue processing...
}
```


### Built-in Context Helpers:

| Function                                              | Purpose                                                          |
|-------------------------------------------------------|------------------------------------------------------------------|
| `rbac.GetPrivilegesFromContext(ctx)`                  | Returns map of granted privileges from context                  |
| `rbac.HasPrivilegeInContext(ctx, code)`               | Shorthand to check if a specific privilege exists in context    |
| `rbac.GetUserIDFromContext(ctx)`                      | Retrieves user ID from context (if injected earlier)            |
| `rbac.GetRoleIDFromContext(ctx)`                      | Retrieves role ID from context (if injected earlier)            |
| `rbac.InjectContext(ctx, roleID, userID, privileges)` | Injects role ID, user ID, and privileges into request context   |

## Example: Run Locally
### Step 1: Clone and run the example
```bash
git clone https://github.com/hatmahat/go-rbac.git
cd go-rbac/example
go run main.go
```
### Step 2: Test access
#### Role with Access
```bash
curl -H "X-Role-ID: admin" -H "X-User-ID: 123" http://localhost:8080/compliance
```
#### Response:
```json
{
  "message": "Hello user 123! You have access."
}
```
#### Role without Access
```bash
curl -H "X-Role-ID: guest" -H "X-User-ID: 456" http://localhost:8080/compliance
```
#### Response:
```json
{
  "error": "forbidden"
}
```

## Configuring Privileges
You can use any data source. If you’re using SQL, this is the expected schema for the GORM example:
```sql
CREATE TABLE privileges (
  id TEXT PRIMARY KEY,
  code TEXT NOT NULL
);

CREATE TABLE role_privileges (
  id TEXT PRIMARY KEY,
  role_id TEXT NOT NULL,
  privilege_id TEXT NOT NULL
);
```
Or define your own structure by implementing PrivilegeRepository.
