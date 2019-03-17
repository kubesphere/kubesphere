package handlers

import (
	"fmt"
	"net/http"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/kiali/kiali/config"
	"github.com/kiali/kiali/log"
	"github.com/kiali/kiali/prometheus"
)

const (
	defaultPrometheusGlobalScrapeInterval = 15 // seconds
)

// PrometheusConfig holds actual Prometheus configuration that is useful to Kiali.
// All durations are in seconds.
type PrometheusConfig struct {
	GlobalScrapeInterval int64 `json:"globalScrapeInterval,omitempty"`
	StorageTsdbRetention int64 `json:"storageTsdbRetention,omitempty"`
}

// PublicConfig is a subset of Kiali configuration that can be exposed to clients to
// help them interact with the system.
type PublicConfig struct {
	IstioNamespace string             `json:"istioNamespace,omitempty"`
	IstioLabels    config.IstioLabels `json:"istioLabels,omitempty"`
	Prometheus     PrometheusConfig   `json:"prometheus,omitempty"`
}

// Config is a REST http.HandlerFunc serving up the Kiali configuration made public to clients.
func Config(w http.ResponseWriter, r *http.Request) {
	defer handlePanic(w)

	// Note that determine the Prometheus config at request time because it is not
	// guaranteed to remain the same during the Kiali lifespan.
	promConfig := getPrometheusConfig()
	config := config.Get()
	publicConfig := PublicConfig{
		IstioNamespace: config.IstioNamespace,
		IstioLabels:    config.IstioLabels,
		Prometheus: PrometheusConfig{
			GlobalScrapeInterval: promConfig.GlobalScrapeInterval,
			StorageTsdbRetention: promConfig.StorageTsdbRetention,
		},
	}

	RespondWithJSONIndent(w, http.StatusOK, publicConfig)
}

type PrometheusPartialConfig struct {
	Global struct {
		Scrape_interval string
	}
}

func getPrometheusConfig() PrometheusConfig {
	promConfig := PrometheusConfig{
		GlobalScrapeInterval: defaultPrometheusGlobalScrapeInterval,
	}

	client, err := prometheus.NewClient()
	if !checkErr(err, "") {
		log.Error(err)
		return promConfig
	}

	configResult, err := client.GetConfiguration()
	if checkErr(err, "Failed to fetch Prometheus configuration") {
		var config PrometheusPartialConfig
		if checkErr(yaml.Unmarshal([]byte(configResult.YAML), &config), "Failed to unmarshal Prometheus configuration") {
			scrapeIntervalString := config.Global.Scrape_interval
			scrapeInterval, err := time.ParseDuration(scrapeIntervalString)
			if checkErr(err, fmt.Sprintf("Invalid global scrape interval [%s]", scrapeIntervalString)) {
				promConfig.GlobalScrapeInterval = int64(scrapeInterval.Seconds())
			}
		}
	}

	flags, err := client.GetFlags()
	if checkErr(err, "Failed to fetch Prometheus flags") {
		if retentionString, ok := flags["storage.tsdb.retention"]; ok {
			retention, err := time.ParseDuration(retentionString)
			if checkErr(err, fmt.Sprintf("Invalid storage.tsdb.retention [%s]", retentionString)) {
				promConfig.StorageTsdbRetention = int64(retention.Seconds())
			}
		}
	}

	return promConfig
}

func checkErr(err error, message string) bool {
	if err != nil {
		log.Errorf("%s: %v", message, err)
		return false
	}
	return true
}
