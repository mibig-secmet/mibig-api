package main

import (
	"net/http"
	"secondarymetabolites.org/mibig-api/pkg/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	w.Write([]byte("Hello from the MIBiG API"))
}

type VersionInfo struct {
	Api string `json:"api"`
}

func (app *application) version(w http.ResponseWriter, r *http.Request) {
	version_info := VersionInfo{"3.0"}
	app.returnJson(version_info, w)
}

type Stats struct {
	NumRecords int                  `json:"num_records"`
	Clusters   []models.StatCluster `json:"clusters"`
}

func (app *application) stats(w http.ResponseWriter, r *http.Request) {
	count, err := app.MibigModel.Count()
	if err != nil {
		app.serverError(w, err)
		return
	}

	clusters, err := app.MibigModel.ClusterStats()
	if err != nil {
		app.serverError(w, err)
		return
	}

	stat_info := Stats{
		NumRecords: count,
		Clusters:   clusters,
	}

	app.returnJson(stat_info, w)
}
