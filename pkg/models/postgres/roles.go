package postgres

import (
	"database/sql"
	"errors"

	"secondarymetabolites.org/mibig-api/pkg/models"
)

type RoleModel struct {
	DB *sql.DB
}

func (m *RoleModel) Ping() error {
	return m.DB.Ping()
}

func (m *RoleModel) List() ([]models.Role, error) {
	var roles []models.Role
	statement := `SELECT role_id, name, description FROM mibig_submitters.roles`
	rows, err := m.DB.Query(statement)
	if err != nil {
		// No roles is not an error in this context
		if errors.Is(err, sql.ErrNoRows) {
			return roles, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var role models.Role
		err = rows.Scan(&role.Id, &role.Name, &role.Description)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (m *RoleModel) Add(name, description string) (int, error) {
	statement := `INSERT INTO mibig_submitters.roles (name, description) VALUES (?, ?)`
	ret, err := m.DB.Exec(statement, name, description)
	if err != nil {
		return 0, err
	}

	roleId, err := ret.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(roleId), nil
}

func (m *RoleModel) UserCount(name string) (int, error) {
	var count int
	statement := `SELECT COUNT(role_id) FROM mibig_submitters.rel_submitters_roles LEFT JOIN mibig_submitters.roles USING (role_id)
	WHERE name = ?`
	row := m.DB.QueryRow(statement, name)
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (m *RoleModel) Delete(name string) error {
	var roleId int

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	row := tx.QueryRow(`SELECT role_id FROM mibig_submitters.roles WHERE name = ?`, name)
	err = row.Scan(&roleId)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`DELETE FROM mibig_submitters.rel_submitters_roles WHERE role_id = ?`, roleId)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`DELETE FROM mibig_submitters.roles WHERE role_id = ?`, roleId)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
