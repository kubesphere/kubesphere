// +kubebuilder:object:generate=true

package templatings

//referers to https://pkg.go.dev/github.com/K-Phoen/grabana/decoder#DashboardVariable
// type Templatings struct {
// 	Interval   *VariableInterval   `json:"interval,omitempty"`
// 	Custom     *VariableCustom     `json:"custom,omitempty"`
// 	Query      *VariableQuery      `json:"query,omitempty"`
// 	Const      *VariableConst      `json:"const,omitempty"`
// 	Datasource *VariableDatasource `json:"datasource,omitempty"`
// }

// type VariableInterval struct {
// 	Name    string   `json:"name,omitempty"`
// 	Label   string   `json:"label,omitempty"`
// 	Default string   `json:"default,omitempty"`
// 	Values  []string `json:"values,omitempty,flow"`
// }

// type VariableCustom struct {
// 	Name       string            `json:"name,omitempty"`
// 	Label      string            `json:"label,omitempty"`
// 	Default    string            `json:"default,omitempty"`
// 	ValuesMap  map[string]string `json:"values_map,omitempty"`
// 	IncludeAll bool              `json:"include_all,omitempty"`
// 	AllValue   string            `json:"all_value,omitempty"`
// }

// type VariableConst struct {
// 	Name      string            `json:"name,omitempty"`
// 	Label     string            `json:"label,omitempty"`
// 	Default   string            `json:"default,omitempty"`
// 	ValuesMap map[string]string `json:"values_map,omitempty"`
// }

// type VariableQuery struct {
// 	Name       string `json:"name,omitempty"`
// 	Label      string `json:"label,omitempty"`
// 	Datasource string `json:"datasource,omitempty"`
// 	Request    string `json:"request,omitempty"`
// 	Regex      string `json:"regex,omitempty"`
// 	IncludeAll bool   `json:"include_all,omitempty"`
// 	DefaultAll bool   `json:"default_all,omitempty"`
// 	AllValue   string `json:"all_value,omitempty"`
// }

// type VariableDatasource struct {
// 	Name       string `json:"name,omitempty"`
// 	Label      string `json:"label,omitempty"`
// 	Query       string `json:"query,omitempty"`
// 	Regex      string `json:"regex,omitempty"`
// 	IncludeAll bool   `json:"include_all,omitempty"`
// }

// type TemplateVar struct {
// 	// common properties, more than one containing
// 	Name       string            `json:"name,omitempty"`
// 	Type       string            `json:"type,omitempty"`
// 	Auto       bool              `json:"auto,omitempty"`
// 	AutoCount  *int              `json:"auto_count,omitempty"`
// 	Label      string            `json:"label,omitempty"`
// 	Default    string            `json:"default,omitempty"`
// 	ValuesMap  map[string]string `json:"values_map,omitempty"`
// 	IncludeAll bool              `json:"include_all,omitempty"`
// 	AllValue   string            `json:"all_value,omitempty"`
// 	Regex      string            `json:"regex,omitempty"`
// 	// from type VariableInterval
// 	Values []string `json:"values,omitempty,flow"`

// 	// type VariableCustom has no special field
// 	// type VariableConst has no special field

// 	// from type VariableQuery
// 	Datasource string `json:"datasource,omitempty"`
// 	Request    string `json:"request,omitempty"`
// 	DefaultAll bool   `json:"default_all,omitempty"`

// 	// from type VariableDatasource
// 	Query string `json:"query,omitempty"`
// }

// Refers to https://pkg.go.dev/github.com/grafana-tools/sdk#Templating
//TemplateVar combines above types to one struct

type TemplateVar struct {
	Name        string   `json:"name,omitempty"`
	Type        string   `json:"type,omitempty"`
	Auto        bool     `json:"auto,omitempty"`
	AutoCount   *int     `json:"auto_count,omitempty"`
	Datasource  *string  `json:"datasource,omitempty"`
	Options     []Option `json:"options,omitempty"`
	IncludeAll  bool     `json:"includeAll,omitempty"`
	AllFormat   string   `json:"allFormat,omitempty"`
	AllValue    string   `json:"allValue,omitempty"`
	Multi       bool     `json:"multi,omitempty"`
	MultiFormat string   `json:"multiFormat,omitempty"`
	Query       string   `json:"query,omitempty"`
	Regex       string   `json:"regex,omitempty"`
	Label       string   `json:"label,omitempty"`
	Hide        uint8    `json:"hide,omitempty"`
	Sort        int      `json:"sort,omitempty"`
}

type Option struct {
	Text     string `json:"text,omitempty"`
	Value    string `json:"value,omitempty"`
	Selected bool   `json:"selected,omitempty"`
}
