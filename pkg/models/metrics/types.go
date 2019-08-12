package metrics

type ServiceMonitorConfig struct {
	Service       string                   `json:"service" description:"service name"`
	Interval      string                   `json:"interval" description:"interval at which metrics should be scraped"`
	ScrapeTimeout string                   `json:"scrapeTimeout" description:"timeout after which the scrape is ended"`
	Endpoints     []ServiceMonitorEndpoint `json:"endpoints" description:"scrapeable endpoint serving Prometheus metrics"`
	Enable        bool                     `json:"enable" description:"set to true to start scraping"`
}

type ServiceMonitorEndpoint struct {
	Port string `json:"port" description:"name of the service port this endpoint refers to"`
	Path string `json:"path" description:"HTTP path to scrape for metrics"`
}
