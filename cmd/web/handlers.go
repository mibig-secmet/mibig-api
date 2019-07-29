package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"secondarymetabolites.org/mibig-api/pkg/models"
	"secondarymetabolites.org/mibig-api/pkg/queries"
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

type queryContainer struct {
	Query        *queries.Query `json:"query"`
	SearchString string         `json:"search_string"`
	Paginate     int            `json:"paginate"`
	Offset       int            `json:"offset"`
	Verbose      bool           `json:"verbose"`
}

type queryResult struct {
	Total    int                      `json:"total"`
	Clusters []models.RepositoryEntry `json:"clusters"`
	Offset   int                      `json:"offset"`
	Paginate int                      `json:"paginate"`
	Stats    string                   `json:"stats"`
}

type queryError struct {
	Message string `json:"message"`
	Error   bool   `json:"error"`
}

func (app *application) search(c *gin.Context) {
	var qc queryContainer
	err := c.BindJSON(&qc)
	if err != nil {
		app.serverError(c, err)
		return
	}

	if qc.Query == nil && qc.SearchString == "" {
		c.JSON(http.StatusBadRequest, queryError{Message: "Invalid query", Error: true})
		return
	}

	if qc.Query == nil {
		qc.Query, err = queries.NewQueryFromString(qc.SearchString)
		if err != nil {
			app.serverError(c, err)
			return
		}
	}

	var entry_ids []int
	entry_ids, err = app.MibigModel.Search(qc.Query.Terms)
	if err != nil {
		c.JSON(http.StatusBadRequest, queryError{Message: err.Error(), Error: true})
		return
	}

	var clusters []models.RepositoryEntry
	clusters, err = app.MibigModel.Get(entry_ids)
	if err != nil {
		app.serverError(c, err)
		return
	}

	result := queryResult{
		Total:    len(entry_ids),
		Clusters: clusters,
		Offset:   qc.Offset,
		Paginate: qc.Paginate,
		Stats:    "Implement me",
	}

	c.JSON(http.StatusOK, &result)
}

func (app *application) available(c *gin.Context) {
	category := c.Param("category")
	term := c.Param("term")
	available, err := app.MibigModel.Available(category, term)
	if err == models.ErrInvalidCategory {
		c.JSON(http.StatusBadRequest, queryError{Message: err.Error(), Error: true})
		return
	} else if err != nil {
		app.serverError(c, err)
		return
	}

	c.JSON(http.StatusOK, &available)
}
