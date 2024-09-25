package oci

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"oras.land/oras-go/pkg/registry"
)

func TestRegistry_Api(t *testing.T) {
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

		t.Errorf("unexpected access: %s %s", r.Method, r.URL)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()
	uri, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("invalid test http server: %v", err)
	}

	reg, err := NewRegistry(uri.Host,
		WithTimeout(5*time.Second),
		WithBasicAuth("", ""),
		WithInsecureSkipVerifyTLS(true))
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	ctx := context.Background()
	err = reg.Ping(ctx)
	if err != nil {
		t.Fatalf("Registry.Ping() error = %v", err)
	}

	var registryTags []string

	repo, err := reg.Repository(ctx, testRepo)
	if err != nil {
		t.Fatalf("Registry.Repository() error = %v", err)
	}
	registryTags, err = registry.Tags(ctx, repo)
	if err != nil {
		t.Fatalf("Registry.Repository().Tags() error = %v", err)
	}
	t.Log(len(registryTags))

	err = reg.Repositories(ctx, "", func(repos []string) error {
		for _, repo := range repos {
			if subRepo, found := strings.CutPrefix(repo, ""); found {
				t.Log(subRepo)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Registry.Repositories() error = %v", err)
	}

}
