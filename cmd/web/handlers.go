package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"secondarymetabolites.org/mibig-api/pkg/models"
)

type VersionInfo struct {
	Api        string `json:"api"`
	BuildTime  string `json:"build_time"`
	GitVersion string `json:"git_version"`
}

func (app *application) version(c *gin.Context) {
	version_info := VersionInfo{
		Api:        "3.0",
		BuildTime:  app.BuildTime,
		GitVersion: app.GitVersion,
	}
	c.JSON(http.StatusOK, &version_info)
}

type Stats struct {
	NumRecords int                  `json:"num_records"`
	Clusters   []models.StatCluster `json:"clusters"`
}

func (app *application) stats(c *gin.Context) {
	count, err := app.MibigModel.Count()
	if err != nil {
		app.serverError(c, err)
		return
	}

	clusters, err := app.MibigModel.ClusterStats()
	if err != nil {
		app.serverError(c, err)
		return
	}

	stat_info := Stats{
		NumRecords: count,
		Clusters:   clusters,
	}

	c.JSON(http.StatusOK, &stat_info)
}

func (app *application) repository(c *gin.Context) {
	repository_entries, err := app.MibigModel.Repository()
	if err != nil {
		app.serverError(c, err)
		return
	}

	c.JSON(http.StatusOK, repository_entries)
}
