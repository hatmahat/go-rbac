package rbac

import (
	"context"

	"gorm.io/gorm"
)

func FetchRolePrivileges(ctx context.Context, db *gorm.DB, roleID string) (map[string]bool, error) {
	query := `
		SELECT p.code
		FROM privileges p
		JOIN role_privileges rp ON p.id = rp.privilege_id
		WHERE rp.role_id = ?
	`

	rows, err := db.Raw(query, roleID).Rows()
	if err != nil {
		return nil, err
	}

	privileges := make(map[string]bool)
	for rows.Next() {
		var code string
		err = rows.Scan(&code)
		if err != nil {
			return nil, err
		}
		privileges[code] = true
	}

	return privileges, nil
}
