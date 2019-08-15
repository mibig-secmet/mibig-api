package mock

import (
	"secondarymetabolites.org/mibig-api/pkg/models"
	"secondarymetabolites.org/mibig-api/pkg/queries"
)

type MibigModel struct {
}

func (m *MibigModel) Counts() (*models.StatCounts, error) {
	return &models.StatCounts{Total: 23}, nil
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

func (m *MibigModel) GenusStats() ([]models.TaxonStats, error) {
	return nil, nil
}

var fakeRepo = []models.RepositoryEntry{
	models.RepositoryEntry{
		Accession:    "BGC1234567",
		Minimal:      false,
		Complete:     "complete",
		Products:     []string{"testomycin"},
		ProductTags:  []models.ProductTag{models.ProductTag{Name: "NRP", Class: "nrps"}},
		OrganismName: "E. xample",
	},
}

func (m *MibigModel) Repository() ([]models.RepositoryEntry, error) {
	return fakeRepo, nil
}

var fakeDB = map[int]models.RepositoryEntry{
	1: models.RepositoryEntry{
		Accession:    "BGC0000001",
		Complete:     "incomplete",
		Minimal:      false,
		Products:     []string{"testomycin A"},
		ProductTags:  []models.ProductTag{models.ProductTag{Name: "Lipopeptide", Class: "nrps"}},
		OrganismName: "E. xample",
	},
	23: models.RepositoryEntry{
		Accession:    "BGC0000023",
		Complete:     "complete",
		Minimal:      false,
		Products:     []string{"testomycin B"},
		ProductTags:  []models.ProductTag{models.ProductTag{Name: "Lanthipeptide", Class: "ripp"}},
		OrganismName: "E. xample",
	},
	42: models.RepositoryEntry{
		Accession:    "BGC0000042",
		Complete:     "Unknown",
		Minimal:      false,
		Products:     []string{"testomycin C"},
		ProductTags:  []models.ProductTag{models.ProductTag{Name: "glycopeptide", Class: "nrps"}},
		OrganismName: "E. xample",
	},
}

func (m *MibigModel) Get(ids []int) ([]models.RepositoryEntry, error) {
	var entries []models.RepositoryEntry
	for _, id := range ids {
		if entry, ok := fakeDB[id]; ok {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func (m *MibigModel) Search(t queries.QueryTerm) ([]int, error) {
	return []int{1, 23, 42}, nil
}

func (m *MibigModel) Available(category string, term string) ([]models.AvailableTerm, error) {
	terms := map[string][]models.AvailableTerm{
		"type": []models.AvailableTerm{
			{Val: "glycopeptide", Desc: "Glycopeptide"},
		},
	}
	res, ok := terms[category]
	if !ok {
		return nil, models.ErrInvalidCategory
	}
	return res, nil
}

func (m *MibigModel) ResultStats(ids []int) (*models.ResultStats, error) {
	return nil, nil
}
