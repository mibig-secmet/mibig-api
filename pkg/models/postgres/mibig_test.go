package postgres

import (
	"database/sql"
	"github.com/google/go-cmp/cmp"
	_ "github.com/lib/pq"
	"io/ioutil"
	"secondarymetabolites.org/mibig-api/pkg/models"
	"secondarymetabolites.org/mibig-api/pkg/queries"
	"testing"
)

type MibigModelTest struct {
	m        MibigModel
	Teardown func()
}

func newMibigTestDB(t *testing.T) *MibigModelTest {
	mt := MibigModelTest{}

	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=secret dbname=mibig_test sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	script, err := ioutil.ReadFile("./testdata/setup.sql")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}

	mt.m = MibigModel{DB: db}

	mt.Teardown = func() {
		script, err := ioutil.ReadFile("./testdata/teardown.sql")
		if err != nil {
			t.Fatal(err)
		}

		_, err = db.Exec(string(script))
		if err != nil {
			t.Fatal(err)
		}
		db.Close()
	}
	return &mt
}

func TestMibigModel(t *testing.T) {
	if testing.Short() {
		t.Skip("postgres: skipping integration test")
	}

	mt := newMibigTestDB(t)
	defer mt.Teardown()

	t.Run("Counts", mt.MibigModelCounts)
	t.Run("ClusterStats", mt.MibigModelClusterStats)
	t.Run("Repository", mt.MibigModelRepository)
	t.Run("Get", mt.MibigModelGet)
	t.Run("Search", mt.MibigModelSearch)
	t.Run("Available", mt.MibigModelAvailable)

}

func (mt *MibigModelTest) MibigModelCounts(t *testing.T) {
	counts, err := mt.m.Counts()
	if err != nil {
		t.Fatal(err)
	}

	if counts.Total != 2 {
		t.Errorf("want 2, got %d", counts.Total)
	}
}

func (mt *MibigModelTest) MibigModelClusterStats(t *testing.T) {
	expected := []models.StatCluster{
		{Type: "nrps", Description: "Nonribosomal peptide", Count: 1, Class: "nrps"},
		{Type: "pks", Description: "Polyketide", Count: 1, Class: "pks"},
		{Type: "ripp", Description: "Ribosomally synthesized and post-translationally modified peptide", Count: 1, Class: "ripp"},
	}

	stats, err := mt.m.ClusterStats()
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, stats) {
		t.Errorf("ClusterStats unexpected results:\n%s", cmp.Diff(expected, stats))
	}
}

func (mt *MibigModelTest) MibigModelRepository(t *testing.T) {
	expected := []models.RepositoryEntry{
		{Accession: "BGC0000535", Complete: "complete", Minimal: false, Products: []string{"nisin A"}, ProductTags: []models.ProductTag{{Name: "Lanthipeptide", Class: "ripp"}}, OrganismName: "Lactococcus lactis subsp. lactis"},
		{Accession: "BGC0001070", Complete: "complete", Minimal: false, Products: []string{"kirromycin"}, ProductTags: []models.ProductTag{
			{Name: "NRP", Class: "nrps"}, {Name: "Modular type I polyketide", Class: "pks"}, {Name: "Trans-AT type I polyketide", Class: "pks"},
		}, OrganismName: "Streptomyces collinus Tu 365"},
	}

	repo, err := mt.m.Repository()
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, repo) {
		t.Errorf("Repository unexpected results:\n%s", cmp.Diff(expected, repo))
	}
}

func (mt *MibigModelTest) MibigModelGet(t *testing.T) {
	tests := []struct {
		Name           string
		Ids            []int
		ExpectedResult []models.RepositoryEntry
		ExpectedError  error
	}{
		{Name: "One", Ids: []int{535}, ExpectedResult: []models.RepositoryEntry{
			{Accession: "BGC0000535", Complete: "complete", Minimal: false, Products: []string{"nisin A"}, ProductTags: []models.ProductTag{{Name: "Lanthipeptide", Class: "ripp"}}, OrganismName: "Lactococcus lactis subsp. lactis"},
		}, ExpectedError: nil},
		{Name: "Two", Ids: []int{535, 1070}, ExpectedResult: []models.RepositoryEntry{
			{Accession: "BGC0000535", Complete: "complete", Minimal: false, Products: []string{"nisin A"}, ProductTags: []models.ProductTag{{Name: "Lanthipeptide", Class: "ripp"}}, OrganismName: "Lactococcus lactis subsp. lactis"},
			{Accession: "BGC0001070", Complete: "complete", Minimal: false, Products: []string{"kirromycin"}, ProductTags: []models.ProductTag{
				{Name: "NRP", Class: "nrps"}, {Name: "Modular type I polyketide", Class: "pks"}, {Name: "Trans-AT type I polyketide", Class: "pks"},
			}, OrganismName: "Streptomyces collinus Tu 365"},
		}, ExpectedError: nil},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			repo, err := mt.m.Get(tt.Ids)
			if err != tt.ExpectedError {
				t.Fatalf("Get(%v) unexpected error: want %s, got %s", tt.Ids, tt.ExpectedError, err)
			}

			if !cmp.Equal(tt.ExpectedResult, repo) {
				t.Errorf("Get(%v) unexpected results:\n%s", tt.Ids, cmp.Diff(tt.ExpectedResult, repo))
			}
		})
	}
}

func (mt *MibigModelTest) MibigModelSearch(t *testing.T) {
	tests := []struct {
		Name           string
		Query          queries.QueryTerm
		ExpectedResult []int
		ExpectedError  error
	}{
		{Name: "RiPP", Query: &queries.Expression{Category: "type", Term: "ripp"}, ExpectedResult: []int{535}, ExpectedError: nil},
		{Name: "Operation/OR", Query: &queries.Operation{
			Operation: queries.OR,
			Left:      &queries.Expression{Category: "type", Term: "ripp"},
			Right:     &queries.Expression{Category: "type", Term: "nrps"},
		}, ExpectedResult: []int{535, 1070}, ExpectedError: nil},
		{Name: "Guess Category", Query: &queries.Expression{Category: "unknown", Term: "ripp"}, ExpectedResult: []int{535}, ExpectedError: nil},
		{Name: "Guess Invalid Category", Query: &queries.Expression{Category: "unknown", Term: "foobarbaz"}, ExpectedResult: nil, ExpectedError: models.ErrInvalidCategory},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			repo, err := mt.m.Search(tt.Query)
			if err != tt.ExpectedError {
				t.Fatalf("Search(%v) unexpected error: want %v, got %v", tt.Query, tt.ExpectedError, err)
			}

			if !cmp.Equal(tt.ExpectedResult, repo) {
				t.Errorf("Search(%v) unexpected results:\n%s", tt.Query, cmp.Diff(tt.ExpectedResult, repo))
			}
		})
	}
}

func (mt *MibigModelTest) MibigModelAvailable(t *testing.T) {
	tests := []struct {
		Name           string
		Category       string
		Term           string
		ExpectedResult []models.AvailableTerm
		ExpectedError  error
	}{
		{Name: "type", Category: "type", Term: "r", ExpectedResult: []models.AvailableTerm{
			{Val: "ripp", Desc: "Ribosomally synthesized and post-translationally modified peptide"},
		}, ExpectedError: nil},
		{Name: "invalid", Category: "foo", Term: "bar", ExpectedResult: nil, ExpectedError: models.ErrInvalidCategory},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			repo, err := mt.m.Available(tt.Category, tt.Term)
			if err != tt.ExpectedError {
				t.Fatalf("Available(%s, %s) unexpected error: want %v, got %v", tt.Category, tt.Term, tt.ExpectedError, err)
			}

			if !cmp.Equal(tt.ExpectedResult, repo) {
				t.Errorf("Available(%s, %s) unexpected results:\n%s", tt.Category, tt.Term, cmp.Diff(tt.ExpectedResult, repo))
			}
		})
	}
}
