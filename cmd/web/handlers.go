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
	Counts     *models.StatCounts   `json:"counts"`
	Clusters   []models.StatCluster `json:"clusters"`
	TaxonStats []models.TaxonStats  `json:"taxon_stats"`
}

func (app *application) stats(c *gin.Context) {
	counts, err := app.MibigModel.Counts()
	if err != nil {
		app.serverError(c, err)
		return
	}

	clusters, err := app.MibigModel.ClusterStats()
	if err != nil {
		app.serverError(c, err)
		return
	}

	taxon_stats, err := app.MibigModel.GenusStats()
	if err != nil {
		app.serverError(c, err)
		return
	}

	stat_info := Stats{
		Counts:     counts,
		Clusters:   clusters,
		TaxonStats: taxon_stats,
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
	Stats    *models.ResultStats      `json:"stats"`
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

	stats, err := app.MibigModel.ResultStats(entry_ids)
	if err != nil {
		app.serverError(c, err)
		return
	}

	result := queryResult{
		Total:    len(entry_ids),
		Clusters: clusters,
		Offset:   qc.Offset,
		Paginate: qc.Paginate,
		Stats:    stats,
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

func (app *application) Convert(c *gin.Context) {
	var req struct {
		Search string `form:"search_string"`
	}
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, queryError{Message: err.Error(), Error: true})
		return
	}

	query, err := queries.NewQueryFromString(req.Search)
	if err != nil {
		app.serverError(c, err)
		return
	}

	err = app.MibigModel.GuessCategories(query)
	if err == models.ErrInvalidCategory {
		c.JSON(http.StatusBadRequest, queryError{Message: err.Error(), Error: true})
		return
	} else if err != nil {
		app.serverError(c, err)
		return
	}

	c.JSON(http.StatusOK, query)
}

func (app *application) Contributors(c *gin.Context) {
	var req struct {
		Ids []string `form:"ids[]"`
	}
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, queryError{Message: err.Error(), Error: true})
		return
	}

	contributors, err := app.MibigModel.LookupContributors(req.Ids)
	if err != nil {
		app.serverError(c, err)
		return
	}

	c.JSON(220, contributors)
}
