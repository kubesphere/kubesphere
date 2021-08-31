package sdk

/*
   Copyright 2016 Alexander I.Grafov <grafov@gmail.com>
   Copyright 2016-2019 The Grafana SDK authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

	   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

   ॐ तारे तुत्तारे तुरे स्व
*/

// Datasource as described in the doc
// http://docs.grafana.org/reference/http_api/#get-all-datasources
type Datasource struct {
	ID                uint        `json:"id"`
	OrgID             uint        `json:"orgId"`
	Name              string      `json:"name"`
	Type              string      `json:"type"`
	Access            string      `json:"access"` // direct or proxy
	URL               string      `json:"url"`
	Password          *string     `json:"password,omitempty"`
	User              *string     `json:"user,omitempty"`
	Database          *string     `json:"database,omitempty"`
	BasicAuth         *bool       `json:"basicAuth,omitempty"`
	BasicAuthUser     *string     `json:"basicAuthUser,omitempty"`
	BasicAuthPassword *string     `json:"basicAuthPassword,omitempty"`
	IsDefault         bool        `json:"isDefault"`
	JSONData          interface{} `json:"jsonData"`
	SecureJSONData    interface{} `json:"secureJsonData"`
}

// Datasource type as described in
// http://docs.grafana.org/reference/http_api/#available-data-source-types
type DatasourceType struct {
	Metrics  bool   `json:"metrics"`
	Module   string `json:"module"`
	Name     string `json:"name"`
	Partials struct {
		Query string `json:"query"`
	} `json:"datasource"`
	PluginType  string `json:"pluginType"`
	ServiceName string `json:"serviceName"`
	Type        string `json:"type"`
}
