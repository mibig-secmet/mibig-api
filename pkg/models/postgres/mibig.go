package postgres

import (
	"database/sql"
	"encoding/json"
	"github.com/lib/pq"
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
	statement := `SELECT term, description, COUNT(1) AS entry_count, safe_class FROM mibig.rel_entries_types
		LEFT JOIN mibig.bgc_types USING (bgc_type_id)
		GROUP BY bgc_type_id, term, description, safe_class
		ORDER BY entry_count DESC`

	var clusters []models.StatCluster

	rows, err := m.DB.Query(statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		cluster := models.StatCluster{}
		if err = rows.Scan(&cluster.Type, &cluster.Description, &cluster.Count, &cluster.Class); err != nil {
			return nil, err
		}
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

func (m *MibigModel) Repository() ([]models.RepositoryEntry, error) {
	statement := `SELECT
		a.acc,
		a.data#>>'{cluster, minimal}' AS minimal,
		a.data#>>'{cluster, compounds}' AS compounds,
		array_agg(b.term) AS biosyn_class,
		array_agg(b.safe_class) AS safe_class,
		t.name
	FROM mibig.entries a
	JOIN mibig.rel_entries_types USING(entry_id)
	JOIN mibig.bgc_types b USING (bgc_type_id)
	JOIN mibig.taxa t USING (tax_id)
	GROUP BY acc, data, t.name
	ORDER BY acc`

	var entries []models.RepositoryEntry

	rows, err := m.DB.Query(statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var classes []string
		var css_classes []string
		var compounds models.CompoundList
		var compounds_raw string

		entry := models.RepositoryEntry{}
		if err = rows.Scan(&entry.Accession, &entry.Minimal, &compounds_raw, pq.Array(&classes), pq.Array(&css_classes), &entry.OrganismName); err != nil {
			return nil, err
		}

		if err = json.Unmarshal([]byte(compounds_raw), &compounds); err != nil {
			return nil, err
		}

		for _, compound := range compounds {
			entry.Products = append(entry.Products, compound.Name)
		}

		for i := range classes {
			tag := models.ProductTag{Name: classes[i], Class: css_classes[i]}
			entry.ProductTags = append(entry.ProductTags, tag)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (m *MibigModel) Get(id int) (*models.Entry, error) {
	return nil, nil
}
