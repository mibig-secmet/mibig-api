package web

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/smtp"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"

	"secondarymetabolites.org/mibig-api/pkg/models"
	"secondarymetabolites.org/mibig-api/pkg/models/mock"
	"secondarymetabolites.org/mibig-api/pkg/queries"
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
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard
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

	viper.Set("buildTime", "Fake time")
	viper.Set("gitVer", "deadbeef")

	app := &application{
		logger:      logger,
		Mail:        sender,
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
	_, ts, _ := newTestApp()
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

	if version.GitVersion != viper.GetString("gitVer") {
		t.Errorf("Expected %s, got %s", viper.GetString("gitVer"), version.GitVersion)
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

	if stats.Counts.Total != 23 {
		t.Errorf("Expected %d, got %d", 23, stats.Counts.Total)
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

func TestSearch(t *testing.T) {
	app, ts, _ := newTestApp()
	defer ts.Close()

	fake_clusters, _ := app.MibigModel.Get([]int{1, 23, 42})
	tests := []struct {
		Name             string
		Query            *queries.Query
		SearchString     string
		ExpectedStatus   int
		ExpectedResponse *queryResult
		ExpectedError    *queryError
	}{
		{
			Name:           "search string",
			SearchString:   "nrps OR ripp",
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: &queryResult{
				Total:    3,
				Clusters: fake_clusters,
				Offset:   0,
				Paginate: 0,
				Stats:    nil,
			},
		},
		{
			Name:           "empty string",
			SearchString:   "",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError: &queryError{
				Message: "Invalid query",
				Error:   true,
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.Name, func(t *testing.T) {
			req := queryContainer{
				SearchString: tt.SearchString,
				Query:        tt.Query,
			}

			raw_req, err := json.Marshal(&req)
			req_body := bytes.NewReader(raw_req)

			response, err := ts.Client().Post(ts.URL+"/api/v1/search", "application/json", req_body)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()

			if response.StatusCode != tt.ExpectedStatus {
				t.Errorf("Expected %d, got %d", tt.ExpectedStatus, response.StatusCode)
			}

			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}

			if tt.ExpectedResponse != nil {
				var parsed queryResult
				err = json.Unmarshal(body, &parsed)
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(*tt.ExpectedResponse, parsed) {
					t.Errorf("Unexpected response.\n%s", cmp.Diff(*tt.ExpectedResponse, parsed))
				}
			}

			if tt.ExpectedError != nil {
				var parsed queryError
				err = json.Unmarshal(body, &parsed)
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(*tt.ExpectedError, parsed) {
					t.Errorf("Unexpected response.\n%s", cmp.Diff(*tt.ExpectedError, parsed))
				}
			}
		})
	}
}

func TestAvailable(t *testing.T) {
	_, ts, _ := newTestApp()
	defer ts.Close()

	tests := []struct {
		Name             string
		Category         string
		Term             string
		ExpectedStatus   int
		ExpectedResponse []models.AvailableTerm
		ExpectedError    *queryError
	}{
		{
			Name:           "existing category",
			Category:       "type",
			Term:           "gly",
			ExpectedStatus: http.StatusOK,
			ExpectedResponse: []models.AvailableTerm{
				{Val: "glycopeptide", Desc: "Glycopeptide"},
			},
		},
		{
			Name:           "missing category",
			Category:       "foo",
			Term:           "bar",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedError: &queryError{
				Message: "Invalid search category",
				Error:   true,
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.Name, func(t *testing.T) {
			response, err := ts.Client().Get(ts.URL + "/api/v1/available/" + tt.Category + "/" + tt.Term)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()

			if response.StatusCode != tt.ExpectedStatus {
				t.Errorf("Expected %d, got %d", tt.ExpectedStatus, response.StatusCode)
			}

			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}

			if tt.ExpectedResponse != nil {
				var parsed []models.AvailableTerm
				err = json.Unmarshal(body, &parsed)
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(tt.ExpectedResponse, parsed) {
					t.Errorf("Unexpected response.\n%s", cmp.Diff(tt.ExpectedResponse, parsed))
				}
			}

			if tt.ExpectedError != nil {
				var parsed queryError
				err = json.Unmarshal(body, &parsed)
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(*tt.ExpectedError, parsed) {
					t.Errorf("Unexpected response.\n%s", cmp.Diff(*tt.ExpectedError, parsed))
				}
			}
		})
	}
}
