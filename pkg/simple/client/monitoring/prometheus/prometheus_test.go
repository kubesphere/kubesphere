package prometheus

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/json-iterator/go"
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestGetNamedMetrics(t *testing.T) {
	tests := []struct {
		fakeResp string
		expected string
	}{
		{
			fakeResp: "metrics-vector-type-prom.json",
			expected: "metrics-vector-type-res.json",
		},
		{
			fakeResp: "metrics-error-prom.json",
			expected: "metrics-error-res.json",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			expected := make([]monitoring.Metric, 0)
			err := jsonFromFile(tt.expected, &expected)
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
		fakeResp string
		expected string
	}{
		{
			fakeResp: "metrics-matrix-type-prom.json",
			expected: "metrics-matrix-type-res.json",
		},
		{
			fakeResp: "metrics-error-prom.json",
			expected: "metrics-error-res.json",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			expected := make([]monitoring.Metric, 0)
			err := jsonFromFile(tt.expected, &expected)
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

func TestGetMetadata(t *testing.T) {
	tests := []struct {
		fakeResp string
		expected string
	}{
		{
			fakeResp: "metadata-prom.json",
			expected: "metadata-res.json",
		},
		{
			fakeResp: "metadata-notfound-prom.json",
			expected: "metadata-notfound-res.json",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			expected := make([]monitoring.Metadata, 0)
			err := jsonFromFile(tt.expected, &expected)
			if err != nil {
				t.Fatal(err)
			}
			if len(expected) == 0 {
				expected = nil
			}

			srv := mockPrometheusService("/api/v1/targets/metadata", tt.fakeResp)
			defer srv.Close()

			client, _ := NewPrometheus(&Options{Endpoint: srv.URL})
			result := client.GetMetadata("default")
			if diff := cmp.Diff(result, expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", expected, diff)
			}
		})
	}
}

func TestGetMetricLabels(t *testing.T) {
	tests := []struct {
		fakeResp string
		expected string
	}{
		{
			fakeResp: "labels-prom.json",
			expected: "labels-res.json",
		},
		{
			fakeResp: "labels-error-prom.json",
			expected: "labels-error-res.json",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			var expected monitoring.MetricLabels
			err := jsonFromFile(tt.expected, &expected)
			if err != nil {
				t.Fatal(err)
			}
			if len(expected) == 0 {
				expected = monitoring.MetricLabels{}
			}

			srv := mockPrometheusService("/api/v1/series", tt.fakeResp)
			defer srv.Close()

			client, _ := NewPrometheus(&Options{Endpoint: srv.URL})
			result := client.GetMetricLabels("default", time.Now(), time.Now())
			if ok := reflect.DeepEqual(result, expected); !ok {
				t.Fatalf("expect %v, but got %v", expected, result)
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

func jsonFromFile(expectedFile string, expectedJsonPtr interface{}) error {
	json, err := ioutil.ReadFile(fmt.Sprintf("./testdata/%s", expectedFile))
	if err != nil {
		return err
	}
	err = jsoniter.Unmarshal(json, expectedJsonPtr)
	if err != nil {
		return err
	}

	return nil
}
