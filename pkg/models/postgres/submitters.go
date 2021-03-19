package postgres

import (
	"database/sql"
	"errors"
	"log"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"secondarymetabolites.org/mibig-api/pkg/models"
	"secondarymetabolites.org/mibig-api/pkg/utils"
)

type SubmitterModel struct {
	DB            *sql.DB
	roleIdCache   map[int64]*models.Role
	roleNameCache map[string]*models.Role
}

func NewSubmitterModel(DB *sql.DB) *SubmitterModel {
	return &SubmitterModel{
		DB:            DB,
		roleIdCache:   make(map[int64]*models.Role, 5),
		roleNameCache: make(map[string]*models.Role, 5),
	}
}

func (m *SubmitterModel) Ping() error {
	return m.DB.Ping()
}

func (m *SubmitterModel) Insert(submitter *models.Submitter, password string) error {
	var err error
	submitter.PasswordHash, err = utils.GeneratePassword(password)
	if err != nil {
		return err
	}

	if submitter.Id == "" {
		submitter.Id, err = utils.GenerateUid(15)
		if err != nil {
			return err
		}
	}

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	statement := `INSERT INTO mibig_submitters.submitters
(user_id, email, name, call_name, institution, password_hash, is_public, gdpr_consent, active)
VALUES
($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err = tx.Exec(statement, submitter.Id, submitter.Email, submitter.Name, submitter.CallName,
		submitter.Institution, submitter.PasswordHash, submitter.Public, submitter.GDPRConsent, submitter.Active)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, role := range submitter.Roles {
		_, err = tx.Exec("INSERT INTO mibig_submitters.rel_submitters_roles VALUES($1, $2)", submitter.Id, role.Id)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (m *SubmitterModel) GetRolesById(role_ids []int64) ([]models.Role, error) {
	var roles []models.Role

	for _, id := range role_ids {
		role, ok := m.roleIdCache[id]
		if !ok {
			statement := `SELECT role_id, name, description FROM mibig_submitters.roles WHERE role_id = $1`
			row := m.DB.QueryRow(statement, id)
			role = &models.Role{}
			err := row.Scan(&role.Id, &role.Name, &role.Description)
			if err != nil {
				return nil, err
			}
			m.roleIdCache[id] = role
		}
		roles = append(roles, *role)
	}

	return roles, nil
}

func (m *SubmitterModel) GetRolesByName(role_names []string) ([]models.Role, error) {
	var roles []models.Role

	for _, name := range role_names {
		role, ok := m.roleNameCache[name]
		if !ok {
			statement := `SELECT role_id, description FROM mibig_submitters.roles WHERE name = $1`
			row := m.DB.QueryRow(statement, name)
			role = &models.Role{Name: name}
			err := row.Scan(&role.Id, &role.Description)
			if err != nil {
				return nil, err
			}
			m.roleNameCache[name] = role
		}
		roles = append(roles, *role)
	}

	return roles, nil
}

func (m *SubmitterModel) Get(email string, active_only bool) (*models.Submitter, error) {
	var submitter models.Submitter
	statement := `SELECT u.user_id, u.email, u.name, u.call_name, u.institution, u.password_hash, u.is_public, u.gdpr_consent, u.active, array_agg(role_id) AS role_ids 
FROM mibig_submitters.submitters AS u
LEFT JOIN mibig_submitters.rel_submitters_roles USING (user_id)
WHERE u.email = $1`
	if active_only {
		statement += " AND active = TRUE"
	}
	statement += ` GROUP BY user_id;`

	var role_ids []int64

	row := m.DB.QueryRow(statement, email)
	err := row.Scan(&submitter.Id, &submitter.Email, &submitter.Name, &submitter.CallName, &submitter.Institution,
		&submitter.PasswordHash, &submitter.Public, &submitter.GDPRConsent, &submitter.Active, pq.Array(&role_ids))
	if err != nil {
		return nil, err
	}

	submitter.Roles, err = m.GetRolesById(role_ids)
	if err != nil {
		return nil, err
	}

	return &submitter, nil
}

func (m *SubmitterModel) Authenticate(email, password string) (*models.Submitter, error) {

	submitter, err := m.Get(email, true)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrInvalidCredentials
		} else {
			return nil, err
		}
	}

	err = bcrypt.CompareHashAndPassword(submitter.PasswordHash, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, models.ErrInvalidCredentials
		} else {
			return nil, err
		}
	}

	return submitter, nil
}

func (m *SubmitterModel) ChangePassword(userId string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	_, err = m.DB.Exec(`UPDATE mibig_submitters.submitters SET password_hash = $1 WHERE user_id = $2`, hashedPassword, userId)
	if err != nil {
		return err
	}

	return nil
}

func (m *SubmitterModel) Update(submitter *models.Submitter, password string) error {
	tx, err := m.DB.Begin()
	if err != nil {
		log.Println("Error starting TX", err.Error())
		return err
	}

	if password == "" {
		row := tx.QueryRow(`SELECT password_hash FROM mibig_submitters.submitters WHERE user_id = $1`, submitter.Id)
		err = row.Scan(&submitter.PasswordHash)
		if err != nil {
			log.Println("Error getting hashed password", err.Error())
			return err
		}
	} else {
		submitter.PasswordHash, err = utils.GeneratePassword(password)
		if err != nil {
			return err
		}
	}

	statement := `UPDATE mibig_submitters.submitters SET
email = $1, name = $2, call_name = $3, password_hash = $4, institution = $5, is_public = $6, gdpr_consent = $7, active = $8
WHERE user_id = $9`
	_, err = tx.Exec(statement, submitter.Email, submitter.Name, submitter.CallName, submitter.PasswordHash, submitter.Institution,
		submitter.Public, submitter.GDPRConsent, submitter.Active, submitter.Id)
	if err != nil {
		tx.Rollback()
		log.Println("Error updating user", submitter.Id, err.Error())
		return err
	}

	existing_roles, err := getExistingRoles(tx, submitter.Id)
	if err != nil {
		tx.Rollback()
		return err
	}

	wanted_roles, err := getWantedRoles(tx, submitter.Roles)
	if err != nil {
		tx.Rollback()
		return err
	}

	to_delete := utils.DifferenceInt(existing_roles, wanted_roles)
	to_add := utils.DifferenceInt(wanted_roles, existing_roles)

	for _, roleId := range to_delete {
		_, err = tx.Exec("DELETE FROM mibig_submitters.rel_submitters_roles WHERE user_id = $1 AND role_id = $2", submitter.Id, roleId)
		if err != nil {
			tx.Rollback()
			log.Println("Error deleting roles", err.Error())
			return err
		}
	}

	for _, roleId := range to_add {
		_, err = tx.Exec("INSERT INTO mibig_submitters.rel_submitters_roles VALUES($1, $2)", submitter.Id, roleId)
		if err != nil {
			tx.Rollback()
			log.Println("Error adding roles", err.Error())
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func getExistingRoles(tx *sql.Tx, userId string) ([]int, error) {
	existing_roles := make([]int, 0, 5)
	rows, err := tx.Query("SELECT role_id FROM mibig_submitters.rel_submitters_roles WHERE user_id = $1", userId)
	if err != nil {
		tx.Rollback()
		log.Println("Error getting existing roles for user", userId, err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var roleId int
		err = rows.Scan(&roleId)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		existing_roles = append(existing_roles, roleId)
	}
	return existing_roles, nil
}

func getWantedRoles(tx *sql.Tx, roles []models.Role) ([]int, error) {
	wanted_roles := make([]int, 0, 5)

	for _, role := range roles {
		wanted_roles = append(wanted_roles, role.Id)
	}

	return wanted_roles, nil
}

func (m *SubmitterModel) List() ([]models.Submitter, error) {
	var submitters []models.Submitter
	statement := `SELECT u.user_id, u.email, u.name, u.call_name, u.institution, u.password_hash, u.is_public, u.gdpr_consent, u.active, array_agg(role_id) AS role_ids 
FROM mibig_submitters.submitters AS u
LEFT JOIN mibig_submitters.rel_submitters_roles USING (user_id)
GROUP BY user_id`
	rows, err := m.DB.Query(statement)
	if err != nil {
		// No users is not an error in this context
		if errors.Is(err, sql.ErrNoRows) {
			return submitters, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var submitter models.Submitter
		var role_ids []int64

		err := rows.Scan(&submitter.Id, &submitter.Email, &submitter.Name, &submitter.CallName, &submitter.Institution,
			&submitter.PasswordHash, &submitter.Public, &submitter.GDPRConsent, &submitter.Active, pq.Array(&role_ids))
		if err != nil {
			return nil, err
		}

		submitter.Roles, err = m.GetRolesById(role_ids)
		if err != nil {
			return nil, err
		}

		submitters = append(submitters, submitter)
	}

	return submitters, nil
}

func (m *SubmitterModel) Delete(email string) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}

	var userId string

	row := tx.QueryRow("SELECT user_id FROM mibig_submitters.submitters WHERE email = $1", email)
	err = row.Scan(&userId)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM mibig_submitters.rel_submitters_roles WHERE user_id = $1", userId)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM mibig_submitters.submitters WHERE user_id = $1", userId)
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
