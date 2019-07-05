package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"secondarymetabolites.org/mibig-api/pkg/models"
	"strings"
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

func (app *application) submit(c *gin.Context) {
	var req models.AccessionRequest
	c.BindJSON(&req)

	if err := app.Mail.Send(req.Email, generateRequestMailBody(&req, app.Mail.Config().Recipient)); err != nil {
		app.serverError(c, err)
		return
	}

	c.String(http.StatusAccepted, "")
}

func generateRequestMailBody(req *models.AccessionRequest, recipient string) []byte {
	compound := strings.Join(req.Compounds, ", ")
	var loci_parts []string
	for _, locus := range req.Loci {
		loci_parts = append(loci_parts, fmt.Sprintf("  %s (%d - %d)", locus.GenBankAccession, locus.Start, locus.End))
	}
	loci := strings.Join(loci_parts, "\n")

	return []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: MIBiG update / request\r\n\r\nName: %s\nEmail: %s\nCompound: %s\nLoci:\n%s",
		req.Email, recipient, req.Name, req.Email, compound, loci))
}
