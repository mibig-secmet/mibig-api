package postgres

import (
	"database/sql"
	"secondarymetabolites.org/mibig-api/pkg/models"
)

type MibigModel struct {
	DB *sql.DB
}

func (m *MibigModel) Count() (int, error) {
	statement := `SELECT COUNT(entry_id) FROM mibig.entries`
	var count int

	err := m.DB.QueryRow(statement).Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}

func (m *MibigModel) ClusterStats() ([]models.StatCluster, error) {
	statement := `SELECT term, description, COUNT(1) AS entry_count FROM mibig.rel_entries_types
		LEFT JOIN mibig.bgc_types USING (bgc_type_id)
		GROUP BY bgc_type_id, term, description
		ORDER BY entry_count DESC`

	var clusters []models.StatCluster

	rows, err := m.DB.Query(statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		cluster := models.StatCluster{}
		if err = rows.Scan(&cluster.Type, &cluster.Description, &cluster.Count); err != nil {
			return nil, err
		}
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

func (m *MibigModel) Get(id int) (*models.Entry, error) {
	return nil, nil
}
