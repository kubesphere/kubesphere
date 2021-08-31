// +kubebuilder:object:generate=true

package panels

// Query editor options
type CommonPanel struct {
	// Name of the  panel
	Title string `json:"title,omitempty"`
	// Type of the  panel
	Type string `json:"type,omitempty"`
	// Panel ID
	Id int64 `json:"id,omitempty"`
	// Description
	Description *string `json:"description,omitempty"`
	// Datasource
	Datasource *string `json:"datasource,omitempty"`
	// Height
	Height   *string `json:"height,omitempty"`
	Decimals *int64  `json:"decimals,omitempty"`
	// A collection of queries
	Targets []Target `json:"targets,omitempty"`
	// Set series color
	Colors []string `json:"colors,omitempty"`
	// legend
	Legend []string `json:"legend,omitempty"`
	// Display unit
	Format string `json:"format,omitempty"`
}

// +kubebuilder:object:generate=true

// Query editor options
// Referers to https://pkg.go.dev/github.com/grafana-tools/sdk#Target
type Target struct {
	// Reference ID
	RefID int64 `json:"refId,omitempty"`
	// only support prometheus,and the corresponding fields are as follows:
	// Input for fetching metrics.
	Expression string `json:"expr,omitempty"`
	// Legend format for outputs. You can make a dynamic legend with templating variables.
	LegendFormat string `json:"legendFormat,omitempty"`
	// Set series time interval
	Step string `json:"step,omitempty"`
}
