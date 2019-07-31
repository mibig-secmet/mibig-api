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
	LEFT JOIN mibig.bgc_types t ON (counter.biosyn_class = t.name)
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
		a.data#>>'{cluster, loci, completeness}' AS complete,
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
		var maybe_completeness sql.NullString

		entry := models.RepositoryEntry{}
		if err := rows.Scan(&entry.Accession, &entry.Minimal, &maybe_completeness, &compounds_raw,
			pq.Array(&classes), pq.Array(&css_classes), &entry.OrganismName); err != nil {
			return nil, err
		}

		if maybe_completeness.Valid {
			entry.Complete = maybe_completeness.String
		} else {
			entry.Complete = "Unknown"
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
		a.data#>>'{cluster, loci, completeness}' AS complete,
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

func (m *MibigModel) GuessCategory(term string) (string, error) {

	for _, category := range []string{"type", "acc", "compound", "genus", "species"} {
		statement := categoryDetector[category]
		var count int
		if err := m.DB.QueryRow(statement, term).Scan(&count); err != nil {
			return "", err
		}
		if count > 0 {
			return category, nil
		}
	}
	return "", models.ErrInvalidCategory
}

var statementByCategory = map[string]string{
	"type": `SELECT entry_id FROM mibig.entries e LEFT JOIN mibig.rel_entries_types ret USING (entry_id) WHERE bgc_type_id IN (
	WITH RECURSIVE all_subtypes AS (
		SELECT bgc_type_id, parent_id FROM mibig.bgc_types WHERE term = $1
	UNION
		SELECT r.bgc_type_id, r.parent_id FROM mibig.bgc_types r INNER JOIN all_subtypes s ON s.bgc_type_id = r.parent_id)
	SELECT bgc_type_id FROM all_subtypes)`,
	"compound":     `SELECT entry_id FROM mibig.compounds WHERE name ILIKE $1`,
	"acc":          `SELECT entry_id FROM mibig.entries WHERE acc ILIKE $1`,
	"superkingdom": `SELECT entry_id FROM mibig.entries LEFT JOIN mibig.taxa USING (tax_id) WHERE superkingdom ILIKE $1`,
	"phylum":       `SELECT entry_id FROM mibig.entries LEFT JOIN mibig.taxa USING (tax_id) WHERE phylum ILIKE $1`,
	"class":        `SELECT entry_id FROM mibig.entries LEFT JOIN mibig.taxa USING (tax_id) WHERE class ILIKE $1`,
	"order":        `SELECT entry_id FROM mibig.entries LEFT JOIN mibig.taxa USING (tax_id) WHERE taxonomic_order ILIKE $1`,
	"family":       `SELECT entry_id FROM mibig.entries LEFT JOIN mibig.taxa USING (tax_id) WHERE family ILIKE $1`,
	"genus":        `SELECT entry_id FROM mibig.entries LEFT JOIN mibig.taxa USING (tax_id) WHERE genus ILIKE $1`,
	"species":      `SELECT entry_id FROM mibig.entries LEFT JOIN mibig.taxa USING (tax_id) WHERE species ILIKE $1`,
	"completeness": `SELECT entry_id FROM mibig.entries WHERE data#>>'{cluster, loci, completeness}' ILIKE $1`,
}

func (m *MibigModel) Search(t queries.QueryTerm) ([]int, error) {
	var entry_ids []int
	switch v := t.(type) {
	case *queries.Expression:
		if v.Category == "unknown" {
			cat, err := m.GuessCategory(v.Term)
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

var availableByCategory = map[string]string{
	"type":         `SELECT DISTINCT(term), description FROM mibig.bgc_types WHERE term ILIKE concat($1::text, '%') OR description ILIKE concat($1::text, '%') ORDER BY term`,
	"compound":     `SELECT DISTINCT(name), name FROM mibig.compounds WHERE name ILIKE concat($1::text, '%')`,
	"acc":          `SELECT DISTINCT(acc), acc FROM mibig.entries WHERE acc ILIKE concat('%', $1::text, '%')`,
	"superkingdom": `SELECT DISTINCT(superkingdom), superkingdom FROM mibig.taxa WHERE superkingdom ILIKE concat('%', $1::text, '%')`,
	"phylum":       `SELECT DISTINCT(phylum), phylum FROM mibig.taxa WHERE phylum ILIKE concat('%', $1::text, '%')`,
	"class":        `SELECT DISTINCT(class), class FROM mibig.taxa WHERE class ILIKE concat('%', $1::text, '%')`,
	"order":        `SELECT DISTINCT(taxonomic_order), taxonomic_order FROM mibig.taxa WHERE taxonomic_order ILIKE concat('%', $1::text, '%')`,
	"family":       `SELECT DISTINCT(family), family FROM mibig.taxa WHERE family ILIKE concat('%', $1::text, '%')`,
	"genus":        `SELECT DISTINCT(genus), genus FROM mibig.taxa WHERE genus ILIKE concat('%', $1::text, '%')`,
	"species":      `SELECT DISTINCT(species), species FROM mibig.taxa WHERE species ILIKE concat('%', $1::text, '%')`,
	"completeness": `SELECT DISTINCT(data#>>'{cluster, loci, completeness}'), data#>>'{cluster, loci, completeness}' FROM mibig.entries WHERE data#>>'{cluster, loci, completeness}' ILIKE concat($1::text, '%')`,
}

func (m *MibigModel) Available(category string, term string) ([]models.AvailableTerm, error) {
	var (
		available []models.AvailableTerm
		statement string
		ok        bool
	)

	if statement, ok = availableByCategory[category]; !ok {
		return nil, models.ErrInvalidCategory
	}
	rows, err := m.DB.Query(statement, term)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var av models.AvailableTerm
		err = rows.Scan(&av.Val, &av.Desc)
		if err != nil {
			return nil, err
		}
		available = append(available, av)
	}
	return available, nil
}
