// Get information on automatic metrics, and manage custom metrics. See also api/custom_measures.
package sonargo

import "net/http"

type MetricsService struct {
	client *Client
}

type Metric struct {
	ID           string `url:"id,omitempty" json:"id,omitempty"`
	Description  string `url:"description,omitempty"` // Description:"Description",ExampleValue:"Size of the team"
	Direction    int    `json:"direction"`
	Domain       string `url:"domain,omitempty"` // Description:"Domain",ExampleValue:"Tests"
	Key          string `url:"key,omitempty"`    // Description:"Key",ExampleValue:"team_size"
	Name         string `url:"name,omitempty"`   // Description:"Name",ExampleValue:"Team Size"
	Type         string `url:"type,omitempty"`   // Description:"Metric type key",ExampleValue:"INT"
	Qualitative  bool   `json:"qualitative,omitempty"`
	Hidden       bool   `json:"hidden,omitempty"`
	Custom       bool   `json:"custom,omitempty"`
	DecimalScale int    `json:"decimalScale,omitempty"`
}

func (s *MetricsService) GetDefaultMetrics() []Metric {
	result := []Metric{}
	for index, key := range metrics {
		result = append(result, Metric{
			Key:         key,
			Name:        metricNames[index],
			Description: metricsDescription[index],
		})
	}
	return result
}

type MetricsDomainsObject struct {
	Domains []string `json:"domains,omitempty"`
}

type MetricsSearchObject struct {
	Metrics []*Metric `json:"metrics,omitempty"`
	P       int64     `json:"p,omitempty"`
	Ps      int64     `json:"ps,omitempty"`
	Total   int64     `json:"total,omitempty"`
}

type MetricsTypesObject struct {
	Types []string `json:"types,omitempty"`
}

type MetricsCreateOption Metric

// Create Create custom metric.<br /> Requires 'Administer System' permission.
func (s *MetricsService) Create(opt *MetricsCreateOption) (v *Metric, resp *http.Response, err error) {
	err = s.ValidateCreateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "metrics/create", opt)
	if err != nil {
		return
	}
	v = new(Metric)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return
	}
	return
}

type MetricsDeleteOption struct {
	Ids  string `url:"ids,omitempty"`  // Description:"Metrics ids to delete.",ExampleValue:"5, 23, 42"
	Keys string `url:"keys,omitempty"` // Description:"Metrics keys to delete",ExampleValue:"team_size, business_value"
}

// Delete Delete metrics and associated measures. Delete only custom metrics.<br />Ids or keys must be provided. <br />Requires 'Administer System' permission.
func (s *MetricsService) Delete(opt *MetricsDeleteOption) (resp *http.Response, err error) {
	err = s.ValidateDeleteOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "metrics/delete", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

// Domains List all custom metric domains.
func (s *MetricsService) Domains() (v *MetricsDomainsObject, resp *http.Response, err error) {
	req, err := s.client.NewRequest("GET", "metrics/domains", nil)
	if err != nil {
		return
	}
	v = new(MetricsDomainsObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type MetricsSearchOption struct {
	F        string `url:"f,omitempty"`        // Description:"Comma-separated list of the fields to be returned in response. All the fields are returned by default.",ExampleValue:""
	IsCustom string `url:"isCustom,omitempty"` // Description:"Choose custom metrics following 3 cases:<ul><li>true: only custom metrics are returned</li><li>false: only non custom metrics are returned</li><li>not specified: all metrics are returned</li></ul>",ExampleValue:"true"
	P        int    `url:"p,omitempty"`        // Description:"1-based page number",ExampleValue:"42"
	Ps       int    `url:"ps,omitempty"`       // Description:"Page size. Must be greater than 0 and less or equal than 500",ExampleValue:"20"
}

// Search Search for metrics
func (s *MetricsService) Search(opt *MetricsSearchOption) (v *MetricsSearchObject, resp *http.Response, err error) {
	err = s.ValidateSearchOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "metrics/search", opt)
	if err != nil {
		return
	}
	v = new(MetricsSearchObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

// Types List all available metric types.
func (s *MetricsService) Types() (v *MetricsTypesObject, resp *http.Response, err error) {
	req, err := s.client.NewRequest("GET", "metrics/types", nil)
	if err != nil {
		return
	}
	v = new(MetricsTypesObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type MetricsUpdateOption Metric

// Update Update a custom metric.<br /> Requires 'Administer System' permission.
func (s *MetricsService) Update(opt *MetricsUpdateOption) (v *Metric, resp *http.Response, err error) {
	err = s.ValidateUpdateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "metrics/update", opt)
	if err != nil {
		return
	}
	v = new(Metric)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return
	}
	return
}
