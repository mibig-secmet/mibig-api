package web

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"secondarymetabolites.org/mibig-api/pkg/models"
	"strconv"
	"strings"
	"time"
)

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

func (app *application) LegacyStoreSubmission(c *gin.Context) {
	mibigJson := c.PostForm("json")
	versionString := c.PostForm("version")

	if mibigJson == "" {
		app.logger.Infow("no JSON provided")
		c.JSON(400, gin.H{"error": true, "message": "json not provided"})
		return
	}

	if !json.Valid([]byte(mibigJson)) {
		app.logger.Infow("invalid json string", "json", mibigJson)
		c.JSON(400, gin.H{"error": true, "message": "Invalid JSON input"})
		return
	}

	if versionString == "" {
		app.logger.Infow("no version provided")
		c.JSON(400, gin.H{"error": true, "message": "Version parameter not provided. Need a version parameter greater than 0"})
		return
	}

	version, err := strconv.ParseInt(versionString, 10, 32)
	if err != nil {
		app.logger.Infow("invalid version string", "version_string", versionString)
		c.JSON(400, gin.H{"error": true, "message": "Version parameter not a valid number"})
		return
	}

	if version <= 0 {
		app.logger.Infow("version too small", "version", version)
		c.JSON(400, gin.H{"error": true, "message": "Need a version parameter greater than 0"})
		return
	}

	submission := models.LegacySubmission{
		Submitted: time.Now().UTC(),
		Modified:  time.Now().UTC(),
		Raw:       mibigJson,
		Version:   int(version),
	}

	if err := app.LegacyModel.CreateSubmission(&submission); err != nil {
		app.serverError(c, err)
		return
	}

	c.Redirect(http.StatusSeeOther, "/static/genes_form.html")
}

func (app *application) LegacyStoreBgcDetailSubmission(c *gin.Context) {
	data := c.PostForm("data")
	target := c.PostForm("target")
	versionString := c.DefaultPostForm("version", "1")
	bgc_id := c.DefaultPostForm("bgc_id", "BGC00000")

	if data == "" {
		app.logger.Infow("no JSON provided")
		c.JSON(400, gin.H{"error": true, "message": "json not provided"})
		return
	}

	if !json.Valid([]byte(data)) {
		app.logger.Infow("invalid json string", "json", data)
		c.JSON(400, gin.H{"error": true, "message": "Invalid JSON input"})
		return
	}

	if target == "" {
		app.logger.Infow("no target provided")
		c.JSON(400, gin.H{"error": true, "message": "target not provided"})
		return
	}

	if versionString == "" {
		app.logger.Infow("no version provided")
		c.JSON(400, gin.H{"error": true, "message": "Version parameter not provided. Need a version parameter greater than 0"})
		return
	}

	version, err := strconv.ParseInt(versionString, 10, 32)
	if err != nil {
		app.logger.Infow("invalid version string", "version_string", versionString)
		c.JSON(400, gin.H{"error": true, "message": "Version parameter not a valid number"})
		return
	}

	if version <= 0 {
		app.logger.Infow("version too small", "version", version)
		c.JSON(400, gin.H{"error": true, "message": "Need a version parameter greater than 0"})
		return
	}

	if target == "gene_info" {
		submission := models.LegacyGeneSubmission{
			BgcId:     bgc_id,
			Submitted: time.Now().UTC(),
			Modified:  time.Now().UTC(),
			Raw:       data,
			Version:   int(version),
		}

		err = app.LegacyModel.CreateGeneSubmission(&submission)
	} else if target == "nrps_info" {
		submission := models.LegacyNrpsSubmission{
			BgcId:     bgc_id,
			Submitted: time.Now().UTC(),
			Modified:  time.Now().UTC(),
			Raw:       data,
			Version:   int(version),
		}

		err = app.LegacyModel.CreateNrpsSubmission(&submission)
	} else {
		app.logger.Infow("Invalid details submission target", "target", target)
		c.JSON(400, gin.H{"error": true, "message": "target parameter not matching. Must be one of 'gene_info' or 'nrps_info'"})
		return
	}

	if err != nil {
		app.serverError(c, err)
		return
	}

	c.AbortWithStatus(204)
}
