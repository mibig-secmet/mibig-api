package postgres

import (
	"database/sql"
	"secondarymetabolites.org/mibig-api/pkg/models"
)

type LegacyModel struct {
	DB *sql.DB
}

func (m *LegacyModel) CreateSubmission(s *models.LegacySubmission) error {
	stmt, err := m.DB.Prepare("INSERT INTO mibig.submissions(submitted, modified, raw, v) VALUES($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(s.Submitted, s.Modified, s.Raw, s.Version); err != nil {
		return err
	}
	return nil
}

func (m *LegacyModel) CreateGeneSubmission(s *models.LegacyGeneSubmission) error {
	stmt, err := m.DB.Prepare("INSERT INTO mibig.gene_submissions(bgc_id, submitted, modified, raw, v) VALUES($1, $2, $3, $4, $5)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(s.BgcId, s.Submitted, s.Modified, s.Raw, s.Version); err != nil {
		return err
	}

	return nil
}

func (m *LegacyModel) CreateNrpsSubmission(s *models.LegacyNrpsSubmission) error {
	stmt, err := m.DB.Prepare("INSERT INTO mibig.nrps_submissions(bgc_id, submitted, modified, raw, v) VALUES($1, $2, $3, $4, $5)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(s.BgcId, s.Submitted, s.Modified, s.Raw, s.Version); err != nil {
		return err
	}

	return nil
}
