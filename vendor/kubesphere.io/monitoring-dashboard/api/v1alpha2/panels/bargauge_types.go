// +kubebuilder:object:generate=true
package panels

type BarGaugeOptions struct {
	Orientation string `json:"orientation,omitempty"`
	TextMode    string `json:"textMode,omitempty"`
	ColorMode   string `json:"colorMode,omitempty"`
	GraphMode   string `json:"graphMode,omitempty"`
	JustifyMode string `json:"justifyMode,omitempty"`
	DisplayMode string `json:"displayMode,omitempty"`
	Content     string `json:"content,omitempty"`
	Mode        string `json:"mode,omitempty"`
}

// refers to https://pkg.go.dev/github.com/grafana-tools/sdk#BarGaugePanel
type BarGaugePanel struct {
	Options *BarGaugeOptions `json:"options,omitempty"`
	// FieldConfig FieldConfig `json:"fieldConfig,omitempty"`
}
