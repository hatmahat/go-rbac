package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hatmahat/go-rbac/rbac"
	"github.com/hatmahat/go-rbac/rbacgorm"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// dummy handler demonstrating privilege check
func ProtectedHandler(c echo.Context) error {
	ctx := c.Request().Context()

	privs, ok := rbac.GetPrivilegesFromContext(ctx)
	if !ok || !privs["read:compliance"] {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
	}

	userID, _ := rbac.GetUserIDFromContext(ctx)
	return c.JSON(http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Hello user %s! You have access.", userID),
	})
}

func main() {
	// 1. Setup in-memory SQLite and seed dummy data
	db := initDB()
	seedData(db)

	// ✅ Create the GORM-based privilege repository
	privRepo := rbacgorm.NewGormPrivilegeRepository(db)

	// 2. Initialize RBAC service with 1-minute auto-refresh
	rbacService := rbac.NewRBACService(privRepo, 1*time.Minute, rbac.NewConsoleLogger())

	// 3. Setup Echo
	e := echo.New()

	// 4. Middleware: inject RBAC context manually using generic logic
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			roleID := c.Request().Header.Get("X-Role-ID")
			userID := c.Request().Header.Get("X-User-ID")

			if roleID == "" || userID == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing headers"})
			}

			privileges, err := rbacService.GetRolePrivileges(c.Request().Context(), roleID)
			if err != nil {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "cannot fetch privileges"})
			}

			ctx := rbac.InjectContext(c.Request().Context(), roleID, userID, privileges)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	})

	// 5. Protected route
	e.GET("/compliance", ProtectedHandler)

	// 6. Start server
	log.Println("Server started at :8080")
	e.Start(":8080")
}

// ------- Mock DB setup below ---------

func initDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Exec(`
	CREATE TABLE IF NOT EXISTS privileges (
		id TEXT PRIMARY KEY,
		code TEXT NOT NULL
	);
	`)

	db.Exec(`
	CREATE TABLE IF NOT EXISTS role_privileges (
		id TEXT PRIMARY KEY,
		role_id TEXT NOT NULL,
		privilege_id TEXT NOT NULL
	);
	`)

	return db
}

func seedData(db *gorm.DB) {
	// Seed privilege
	db.Exec(`INSERT INTO privileges (id, code) VALUES ('p1', 'read:compliance')`)

	// Link 'admin' role to the privilege
	db.Exec(`INSERT INTO role_privileges (id, role_id, privilege_id) VALUES ('rp1', 'admin', 'p1')`)

	// 'guest' role has no privileges
}
