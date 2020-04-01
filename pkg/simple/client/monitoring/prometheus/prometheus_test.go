package prometheus

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/json-iterator/go"
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetNamedMetrics(t *testing.T) {
	tests := []struct {
		name     string
		fakeResp string
		expected string
	}{
		{"prom returns good values", "metrics-vector-type-prom.json", "metrics-vector-type-res.json"},
		{"prom returns error", "metrics-error-prom.json", "metrics-error-res.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected, err := jsonFromFile(tt.expected)
			if err != nil {
				t.Fatal(err)
			}

			srv := mockPrometheusService("/api/v1/query", tt.fakeResp)
			defer srv.Close()

			client, _ := NewPrometheus(&Options{Endpoint: srv.URL})
			result := client.GetNamedMetrics([]string{"cluster_cpu_utilisation"}, time.Now(), monitoring.ClusterOption{})
			if diff := cmp.Diff(result, expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", expected, diff)
			}
		})
	}
}

func TestGetNamedMetricsOverTime(t *testing.T) {
	tests := []struct {
		name     string
		fakeResp string
		expected string
	}{
		{"prom returns good values", "metrics-matrix-type-prom.json", "metrics-matrix-type-res.json"},
		{"prom returns error", "metrics-error-prom.json", "metrics-error-res.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected, err := jsonFromFile(tt.expected)
			if err != nil {
				t.Fatal(err)
			}

			srv := mockPrometheusService("/api/v1/query_range", tt.fakeResp)
			defer srv.Close()

			client, _ := NewPrometheus(&Options{Endpoint: srv.URL})
			result := client.GetNamedMetricsOverTime([]string{"cluster_cpu_utilisation"}, time.Now().Add(-time.Minute*3), time.Now(), time.Minute, monitoring.ClusterOption{})
			if diff := cmp.Diff(result, expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", expected, diff)
			}
		})
	}
}

func mockPrometheusService(pattern, fakeResp string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc(pattern, func(res http.ResponseWriter, req *http.Request) {
		b, _ := ioutil.ReadFile(fmt.Sprintf("./testdata/%s", fakeResp))
		res.Write(b)
	})
	return httptest.NewServer(mux)
}

func jsonFromFile(expectedFile string) ([]monitoring.Metric, error) {
	expectedJson := []monitoring.Metric{}

	json, err := ioutil.ReadFile(fmt.Sprintf("./testdata/%s", expectedFile))
	if err != nil {
		return expectedJson, err
	}
	err = jsoniter.Unmarshal(json, &expectedJson)
	if err != nil {
		return expectedJson, err
	}

	return expectedJson, nil
}
