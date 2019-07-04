package mock

import (
	"secondarymetabolites.org/mibig-api/pkg/models"
)

type MibigModel struct {
}

func (m *MibigModel) Count() (int, error) {
	return 23, nil
}

var fakeStats = []models.StatCluster{
	models.StatCluster{
		Type:        "Polyketide",
		Description: "Polyketide",
		Class:       "Polyketide",
		Count:       17,
	},
	models.StatCluster{
		Type:        "NRP",
		Description: "Nonribosomal peptide",
		Class:       "NRP",
		Count:       6,
	},
}

func (m *MibigModel) ClusterStats() ([]models.StatCluster, error) {
	return fakeStats, nil
}

var fakeRepo = []models.RepositoryEntry{
	models.RepositoryEntry{
		Accession:    "BGC1234567",
		Minimal:      false,
		Products:     []string{"testomycin"},
		ProductTags:  []models.ProductTag{models.ProductTag{Name: "NRP", Class: "NRP"}},
		OrganismName: "E. xample",
	},
}

func (m *MibigModel) Repository() ([]models.RepositoryEntry, error) {
	return fakeRepo, nil
}

func (m *MibigModel) Get(id int) (*models.Entry, error) {
	return nil, nil
}
