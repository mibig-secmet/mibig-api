package main

import (
	"bytes"
	"encoding/json"
	"github.com/andreyvit/diff"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/smtp"
	"net/url"
	"secondarymetabolites.org/mibig-api/pkg/models"
	"secondarymetabolites.org/mibig-api/pkg/models/mock"
	"strings"
	"testing"
)

type emailRecorder struct {
	addr string
	auth smtp.Auth
	from string
	to   []string
	msg  []byte
}

func mockSend(expectedError error) (func(string, smtp.Auth, string, []string, []byte) error, *emailRecorder) {
	rec := new(emailRecorder)
	return func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
		*rec = emailRecorder{addr, auth, from, to, msg}
		return expectedError
	}, rec
}

func newTestApp() (*application, *httptest.Server, *emailRecorder) {
	logger := setupLogging(true)
	conf := models.MailConfig{
		Username:  "alice",
		Password:  "secret",
		Host:      "mail.example.com",
		Port:      25,
		Recipient: "alice@example.com",
	}
	mail_func, mail_rec := mockSend(nil)
	sender := models.NewSender(conf, mail_func)
	mux := setupMux(true, logger.Desugar())

	app := &application{
		logger:      logger,
		Mail:        sender,
		BuildTime:   "Fake time",
		GitVersion:  "deadbeef",
		MibigModel:  &mock.MibigModel{},
		LegacyModel: &mock.LegacyModel{},
		Mux:         mux,
	}
	mux = app.routes()
	mux.GET("/static/genes_form.html", func(c *gin.Context) {
		c.String(http.StatusOK, "Nothing to see here")
	})

	ts := httptest.NewServer(mux)

	return app, ts, mail_rec
}

func TestVersion(t *testing.T) {
	app, ts, _ := newTestApp()
	defer ts.Close()

	response, err := ts.Client().Get(ts.URL + "/api/v1/version")
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, response.StatusCode)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	var version VersionInfo
	if err := json.Unmarshal(body, &version); err != nil {
		t.Fatal(err)
	}

	if version.GitVersion != app.GitVersion {
		t.Errorf("Expected %s, got %s", app.GitVersion, version.GitVersion)
	}
}

func TestStats(t *testing.T) {
	_, ts, _ := newTestApp()
	defer ts.Close()

	response, err := ts.Client().Get(ts.URL + "/api/v1/stats")
	if err != nil {
		t.Fatal(err)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, response.StatusCode)
	}

	var stats Stats
	if err := json.Unmarshal(body, &stats); err != nil {
		t.Fatal(err)
	}

	if stats.NumRecords != 23 {
		t.Errorf("Expected %d, got %d", 23, stats.NumRecords)
	}

	if len(stats.Clusters) != 2 {
		t.Errorf("Expected %d cluster entries, got %d", 2, len(stats.Clusters))
	}
}

func TestRepository(t *testing.T) {
	_, ts, _ := newTestApp()
	defer ts.Close()

	response, err := ts.Client().Get(ts.URL + "/api/v1/repository")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, got %d", http.StatusOK, response.StatusCode)
	}

	var repo []models.RepositoryEntry
	if err := json.Unmarshal(body, &repo); err != nil {
		t.Fatal(err)
	}

	if len(repo) != 1 {
		t.Errorf("Expected repository of length %d, got %d", 1, len(repo))
	}
}

func TestSubmit(t *testing.T) {
	_, ts, mail_rec := newTestApp()
	defer ts.Close()

	expected_body := strings.Join([]string{
		"From: alice@example.com\r",
		"To: alice@example.com\r",
		"Subject: MIBiG update / request\r",
		"\r",
		"Name: Alice",
		"Email: alice@example.com",
		"Compound: testomycin",
		"Loci:",
		"  ABC12345 (23 - 42)",
	}, "\n")

	req := models.AccessionRequest{
		Name:      "Alice",
		Email:     "alice@example.com",
		Compounds: []string{"testomycin"},
		Loci: []models.AccessionRequestLocus{
			models.AccessionRequestLocus{
				Start:            23,
				End:              42,
				GenBankAccession: "ABC12345",
			},
		},
	}

	raw_req, err := json.Marshal(&req)
	req_body := bytes.NewReader(raw_req)

	response, err := ts.Client().Post(ts.URL+"/api/v1/submit", "application/json", req_body)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusAccepted {
		t.Errorf("Expected %d, got %d (%s)", http.StatusAccepted, response.StatusCode, string(body))
	}

	if actual, expected := strings.TrimSpace(string(mail_rec.msg)), strings.TrimSpace(expected_body); actual != expected {
		t.Errorf("Unexpected email body:\n%v", diff.LineDiff(expected, actual))
	}

}

func TestLegacySubmission(t *testing.T) {
	_, ts, _ := newTestApp()
	defer ts.Close()

	form := url.Values{}
	form.Set("json", `{"foo": "bar"}`)
	form.Set("version", "1")

	response, err := ts.Client().Post(ts.URL+"/api/v1/bgc-registration", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected %d, got %d (%s)", http.StatusOK, response.StatusCode, string(body))
	}
}

func TestLegacyGeneSubmission(t *testing.T) {
	_, ts, _ := newTestApp()
	defer ts.Close()

	form := url.Values{}
	form.Set("data", `{"foo": "bar"}`)
	form.Set("version", "1")
	form.Set("bgc_id", "BGC1234567")
	form.Set("target", "gene_info")

	response, err := ts.Client().Post(ts.URL+"/api/v1/bgc-detail-registration", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusNoContent {
		t.Errorf("Expected %d, got %d (%s)", http.StatusNoContent, response.StatusCode, string(body))
	}
}

func TestLegacyNrpsSubmission(t *testing.T) {
	_, ts, _ := newTestApp()
	defer ts.Close()

	form := url.Values{}
	form.Set("data", `{"foo": "bar"}`)
	form.Set("version", "1")
	form.Set("bgc_id", "BGC1234567")
	form.Set("target", "nrps_info")

	response, err := ts.Client().Post(ts.URL+"/api/v1/bgc-detail-registration", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}

	if response.StatusCode != http.StatusNoContent {
		t.Errorf("Expected %d, got %d (%s)", http.StatusNoContent, response.StatusCode, string(body))
	}
}
