// +kubebuilder:object:generate=true

package panels

// Query editor options
type Target struct {
	// Input for fetching metrics.
	Expression string `json:"expr,omitempty"`
	// Legend format for outputs. You can make a dynamic legend with templating variables.
	LegendFormat string `json:"legendFormat,omitempty"`
	// Reference ID
	RefID int64 `json:"refId,omitempty"`
	// Set series time interval
	Step string `json:"step,omitempty"`
}
