package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"secondarymetabolites.org/mibig-api/pkg/models"
	"strings"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	w.Write([]byte("Hello from the MIBiG API"))
}

type VersionInfo struct {
	Api        string `json:"api"`
	BuildTime  string `json:"build_time"`
	GitVersion string `json:"git_version"`
}

func (app *application) version(w http.ResponseWriter, r *http.Request) {
	version_info := VersionInfo{
		Api:        "3.0",
		BuildTime:  app.BuildTime,
		GitVersion: app.GitVersion,
	}
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

func (app *application) repository(w http.ResponseWriter, r *http.Request) {
	repository_entries, err := app.MibigModel.Repository()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.returnJson(repository_entries, w)
}

func (app *application) submit(w http.ResponseWriter, r *http.Request) {
	var req models.AccessionRequest

	host_port := fmt.Sprintf("%s:%d", app.Mail.host, app.Mail.port)

	auth := smtp.PlainAuth(
		"",
		app.Mail.username,
		app.Mail.password,
		app.Mail.host,
	)

	if r.Body == nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := smtp.SendMail(
		host_port,
		auth,
		req.Email,
		[]string{app.Mail.recipient},
		generateRequestMailBody(&req, &app.Mail),
	); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	return
}

func generateRequestMailBody(req *models.AccessionRequest, m *mail) []byte {
	compound := strings.Join(req.Compounds, ", ")
	var loci_parts []string
	for _, locus := range req.Loci {
		loci_parts = append(loci_parts, fmt.Sprintf("  %s (%d - %d)", locus.GenBankAccession, locus.Start, locus.End))
	}
	loci := strings.Join(loci_parts, "\n")

	return []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: MIBiG update / request\r\n\r\nName: %s,\nEmail: %s,\nCompound: %s,\nLoci:\n%s",
		req.Email, m.recipient, req.Name, req.Email, compound, loci))
}
