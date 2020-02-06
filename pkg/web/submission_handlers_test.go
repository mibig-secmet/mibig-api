package web

import (
	"bytes"
	"encoding/json"
	"github.com/andreyvit/diff"
	"io/ioutil"
	"net/http"
	"net/url"
	"secondarymetabolites.org/mibig-api/pkg/models"
	"strings"
	"testing"
)

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
