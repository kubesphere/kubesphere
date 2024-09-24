package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	appv2 "kubesphere.io/api/application/v2"
)

func TestLoadRepoIndexFromOci(t *testing.T) {
	testRepos := []string{"helmcharts/nginx", "helmcharts/test-api", "helmcharts/test-ui", "helmcharts/demo-app"}
	testTags := []string{"1.0.0", "1.2.0", "1.0.3"}
	testRepo := testRepos[1]
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && (r.URL.Path == "/v2" || r.URL.Path == "/v2/") {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodGet && r.URL.Path == "/v2/_catalog" {
			result := struct {
				Repositories []string `json:"repositories"`
			}{
				Repositories: testRepos,
			}
			if err := json.NewEncoder(w).Encode(result); err != nil {
				t.Errorf("failed to write response: %v", err)
			}
			return
		}
		if r.Method == http.MethodGet && r.URL.Path == fmt.Sprintf("/v2/%s/tags/list", testRepo) {
			result := struct {
				Tags []string `json:"tags"`
			}{
				Tags: testTags,
			}
			if err := json.NewEncoder(w).Encode(result); err != nil {
				t.Errorf("failed to write response: %v", err)
			}
			return
		}

		t.Logf("unexpected access: %s %s", r.Method, r.URL)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()
	uri, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("invalid test http server: %v", err)
	}

	url := fmt.Sprintf("oci://%s/helmcharts", uri.Host)
	cred := appv2.RepoCredential{
		Username: "",
		Password: "",
	}
	index, err := LoadRepoIndexFromOci(url, cred)
	if err != nil {
		t.Errorf("LoadRepoIndexFromOci() error: %s", err)
	}
	t.Log(len(index.Entries))
}
