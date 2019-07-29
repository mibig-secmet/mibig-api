package postgres

import (
	"database/sql"
	"github.com/google/go-cmp/cmp"
	_ "github.com/lib/pq"
	"io/ioutil"
	"secondarymetabolites.org/mibig-api/pkg/models"
	"testing"
)

func newTestDB(t *testing.T) (*sql.DB, func()) {
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

	return db, func() {
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
}

func TestMibigModelCount(t *testing.T) {
	if testing.Short() {
		t.Skip("postgres: skipping integration test")
	}

	db, teardown := newTestDB(t)
	defer teardown()

	m := MibigModel{DB: db}

	count, err := m.Count()
	if err != nil {
		t.Fatal(err)
	}

	if count != 2 {
		t.Errorf("want 2, got %d", count)
	}
}

func TestMibigModelClusterStats(t *testing.T) {
	if testing.Short() {
		t.Skip("postgres: skipping integration test")
	}

	expected := []models.StatCluster{
		{Type: "nrps", Description: "Nonribosomal peptide", Count: 1, Class: "nrps"},
		{Type: "pks", Description: "Polyketide", Count: 1, Class: "pks"},
		{Type: "ripp", Description: "Ribosomally synthesized and post-translationally modified peptide", Count: 1, Class: "ripp"},
	}

	db, teardown := newTestDB(t)
	defer teardown()

	m := MibigModel{DB: db}

	stats, err := m.ClusterStats()
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, stats) {
		t.Errorf("ClusterStats unexpected results:\n%s", cmp.Diff(expected, stats))
	}
}

func TestMibigModelRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("postgres: skipping integration test")
	}

	expected := []models.RepositoryEntry{
		{Accession: "BGC0000535", Minimal: false, Products: []string{"nisin A"}, ProductTags: []models.ProductTag{{Name: "Lanthipeptide", Class: "ripp"}}, OrganismName: "Lactococcus lactis subsp. lactis"},
		{Accession: "BGC0001070", Minimal: false, Products: []string{"kirromycin"}, ProductTags: []models.ProductTag{
			{Name: "NRP", Class: "nrps"}, {Name: "Modular type I polyketide", Class: "pks"}, {Name: "Trans-AT type I polyketide", Class: "pks"},
		}, OrganismName: "Streptomyces collinus Tu 365"},
	}

	db, teardown := newTestDB(t)
	defer teardown()

	m := MibigModel{DB: db}

	repo, err := m.Repository()
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(expected, repo) {
		t.Errorf("Repository unexpected results:\n%s", cmp.Diff(expected, repo))
	}
}
