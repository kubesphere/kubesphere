// +kubebuilder:object:generate=true

package panels

// SingleStat shows instant query result
type SinglestatPanel struct {
	// spark line: full or bottom
	SparkLine string `json:"sparkline,omitempty"`
	// gauge
	Gauge Gauge `json:"gauge,omitempty"`
	// value name
	ValueName string `json:"valueName,omitempty"`
}

// Gauge for a stat
type Gauge struct {
	MaxValue         int64 `json:"maxValue,omitempty"`
	MinValue         int64 `json:"minValue,omitempty"`
	Show             bool  `json:"show,omitempty"`
	ThresholdLabels  bool  `json:"thresholdLabels,omitempty"`
	ThresholdMarkers bool  `json:"thresholdMarkers,omitempty"`
}

// type SparkLine struct {
// 	FillColor *string  `json:"fillColor,omitempty"`
// 	Full      bool     `json:"full,omitempty"`
// 	LineColor *string  `json:"lineColor,omitempty"`
// 	Show      bool     `json:"show,omitempty"`
// 	YMin      *float64 `json:"ymin,omitempty"`
// 	YMax      *float64 `json:"ymax,omitempty"`
// }
