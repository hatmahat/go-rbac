# go-rbac

A lightweight, framework-agnostic **Role-Based Access Control (RBAC)** library for Go, built with:

- ✅ Context-based privilege injection
- ✅ In-memory caching for fast access
- ✅ Optional auto-refresh of role privileges
- ✅ Minimal dependency (works with any HTTP framework)

---

## ✨ Features

- No framework lock-in (works with Echo, Gin, Chi, Fiber, etc.)
- Simple API: `HasPrivilege`, `GetRolePrivileges`, `InjectContext`
- Built-in GORM query layer
- Middleware helpers available for Echo or custom HTTP setups

---

## 📦 Installation

```bash
go get github.com/hatmahat/go-rbac
```

## 🧱 Folder Structure
```
go-rbac/
├── rbac/               # Core RBAC logic (framework-agnostic)
│   ├── cache.go
│   ├── context.go
│   ├── context_injector.go
│   ├── query.go
│   ├── service.go
├── example/            # Minimal example using Echo
│   └── main.go
├── go.mod
└── README.md
```

## 🚀 Quick Start 
### Step 1: Initialize RBAC service
```go
import "github.com/hatmahat/go-rbac/rbac"

rbacService := rbac.NewRBACService(db, 1*time.Minute)
```
### Step 2: Inject into context 
#### Examples

#### ✅ 1. Echo
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

#### ✅ 2. Gin
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

#### ✅ 3. Chi
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

#### ✅ 4. Fiber
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

## 🧪 Example: Run Locally
### Step 1: Clone and run the example
```bash
git clone https://github.com/hatmahat/go-rbac.git
cd go-rbac/example
go run main.go
```
### Step 2: Test access
#### ✅ Role with Access
```bash
curl -H "X-Role-ID: admin" -H "X-User-ID: 123" http://localhost:8080/compliance
```
#### Response:
```json
{
  "message": "Hello user 123! You have access."
}
```
#### ❌ Role without Access
```bash
curl -H "X-Role-ID: guest" -H "X-User-ID: 456" http://localhost:8080/compliance
```
#### Response:
```json
{
  "error": "forbidden"
}
```

## 🔧 Configuring Privileges
This package expects the following schema (you can adjust as needed):
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
You may modify FetchRolePrivileges() in query.go if your schema differs.

## 🧰 Built-in Helpers
- InjectContext(ctx, roleID, userID, privileges)
- GetPrivilegesFromContext(ctx)
- HasPrivilegeInContext(ctx, "priv-code")