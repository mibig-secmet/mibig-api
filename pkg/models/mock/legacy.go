package mock

import (
	"secondarymetabolites.org/mibig-api/pkg/models"
)

type LegacyModel struct {
}

func (m *LegacyModel) CreateSubmission(submission *models.LegacySubmission) error {
	return nil
}

func (m *LegacyModel) CreateGeneSubmission(submission *models.LegacyGeneSubmission) error {
	return nil
}

func (m *LegacyModel) CreateNrpsSubmission(submission *models.LegacyNrpsSubmission) error {
	return nil
}
