package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"secondarymetabolites.org/mibig-api/pkg/models"
	"secondarymetabolites.org/mibig-api/pkg/queries"
	"secondarymetabolites.org/mibig-api/pkg/utils"
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
	statement := `SELECT term, description, entry_count, safe_class FROM
	(
		SELECT jsonb_array_elements_text(data#>'{cluster, biosyn_class}') AS biosyn_class,
			   COUNT(1) AS entry_count FROM mibig.entries GROUP BY biosyn_class
	) counter
	LEFT JOIN mibig.bgc_types t ON (counter.biosyn_class = t.term)
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
		array_agg(b.name) AS biosyn_class,
		array_agg(b.safe_class) AS safe_class,
		t.name
	FROM mibig.entries a
	JOIN mibig.rel_entries_types USING(entry_id)
	JOIN mibig.bgc_types b USING (bgc_type_id)
	JOIN mibig.taxa t USING (tax_id)
	GROUP BY acc, data, t.name
	ORDER BY acc`

	rows, err := m.DB.Query(statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return parseRepositoryEntriesFromDB(rows)
}

func parseRepositoryEntriesFromDB(rows *sql.Rows) ([]models.RepositoryEntry, error) {
	var entries []models.RepositoryEntry

	for rows.Next() {
		var classes []string
		var css_classes []string
		var compounds models.CompoundList
		var compounds_raw string

		entry := models.RepositoryEntry{}
		if err := rows.Scan(&entry.Accession, &entry.Minimal, &compounds_raw, pq.Array(&classes), pq.Array(&css_classes), &entry.OrganismName); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(compounds_raw), &compounds); err != nil {
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

func (m *MibigModel) Get(ids []int) ([]models.RepositoryEntry, error) {
	statement := `SELECT
		a.acc,
		a.data#>>'{cluster, minimal}' AS minimal,
		a.data#>>'{cluster, compounds}' AS compounds,
		array_agg(b.name) AS biosyn_class,
		array_agg(b.safe_class) AS safe_class,
		t.name
	FROM ( SELECT * FROM unnest($1::int[]) AS entry_id) vals
	JOIN mibig.entries a USING (entry_id)
	JOIN mibig.rel_entries_types USING (entry_id)
	JOIN mibig.bgc_types b USING (bgc_type_id)
	JOIN mibig.taxa t USING (tax_id)
	GROUP BY acc, data, t.name
	ORDER BY acc`

	rows, err := m.DB.Query(statement, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return parseRepositoryEntriesFromDB(rows)
}

var categoryDetector = map[string]string{
	"type":     `SELECT COUNT(bgc_type_id) FROM mibig.bgc_types WHERE term ILIKE $1`,
	"acc":      `SELECT COUNT(entry_id) FROM mibig.entries WHERE acc ILIKE $1`,
	"compound": `SELECT COUNT(entry_id) FROM mibig.compounds WHERE name ILIKE $1`,
	"genus":    `SELECT COUNT(tax_id) FROM mibig.taxa WHERE genus ILIKE $1`,
	"species":  `SELECT COUNT(tax_id) FROM mibig.taxa WHERE species ILIKE $1`,
}

func (m *MibigModel) GuessCategory(expression *queries.Expression) (string, error) {

	for _, category := range []string{"type", "acc", "compound", "genus", "species"} {
		statement := categoryDetector[category]
		var count int
		if err := m.DB.QueryRow(statement, expression.Term).Scan(&count); err != nil {
			return expression.Category, err
		}
		if count > 0 {
			return category, nil
		}
	}
	return expression.Category, nil
}

var statementByCategory = map[string]string{
	"type": `SELECT entry_id FROM mibig.entries e LEFT JOIN mibig.rel_entries_types ret USING (entry_id) WHERE bgc_type_id IN (
	WITH RECURSIVE all_subtypes AS (
		SELECT bgc_type_id, parent_id FROM mibig.bgc_types WHERE term = $1
	UNION
		SELECT r.bgc_type_id, r.parent_id FROM mibig.bgc_types r INNER JOIN all_subtypes s ON s.bgc_type_id = r.parent_id)
	SELECT bgc_type_id FROM all_subtypes)`,
	"compound": `SELECT entry_id FROM mibig.compounds WHERE name ILIKE $1`,
}

func (m *MibigModel) Search(t queries.QueryTerm) ([]int, error) {
	var entry_ids []int
	switch v := t.(type) {
	case *queries.Expression:
		if v.Category == "unknown" {
			cat, err := m.GuessCategory(v)
			if err != nil {
				return nil, err
			}
			v.Category = cat
		}
		statement, ok := statementByCategory[v.Category]
		if !ok {
			return []int{}, nil
		}

		rows, err := m.DB.Query(statement, v.Term)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var entry_id int
			rows.Scan(&entry_id)
			entry_ids = append(entry_ids, entry_id)
		}

		return entry_ids, nil

	case *queries.Operation:
		var (
			err   error
			left  []int
			right []int
		)
		left, err = m.Search(v.Left)
		if err != nil {
			return nil, err
		}
		right, err = m.Search(v.Right)
		if err != nil {
			return nil, err
		}
		switch v.Operation {
		case queries.AND:
			return utils.Intersect(left, right), nil
		case queries.OR:
			return utils.Union(left, right), nil
		case queries.EXCEPT:
			return utils.Difference(left, right), nil
		default:
			return nil, fmt.Errorf("Invalid operation: %s", v.Op())
		}
	}
	// Should never get here
	return entry_ids, nil
}
