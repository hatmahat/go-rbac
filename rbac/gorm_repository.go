package rbac

import (
	"context"

	"gorm.io/gorm"
)

type GormPrivilegeRepository struct {
	db *gorm.DB
}

func NewGormPrivilegeRepository(db *gorm.DB) *GormPrivilegeRepository {
	return &GormPrivilegeRepository{db: db}
}

func (g *GormPrivilegeRepository) FetchPrivilegesByRoleID(ctx context.Context, roleID string) (map[string]bool, error) {
	query := `
		SELECT p.code
		FROM privileges p
		JOIN role_privileges rp ON p.id = rp.privilege_id
		WHERE rp.role_id = ?
	`

	rows, err := g.db.Raw(query, roleID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]bool)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		result[code] = true
	}

	return result, nil
}
