// +kubebuilder:object:generate=true

package panels

// Graph visualizes range query results into a linear graph
type GraphPanel struct {
	// Display as a bar chart
	Bars bool `json:"bars,omitempty"`
	// Display as a line chart
	Lines bool `json:"lines,omitempty"`
	// Display as a stacked chart
	Stack bool `json:"stack,omitempty"`

	Xaxis Axis `json:"xaxis,omitempty"`
	// Y-axis options
	Yaxes []Axis `json:"yaxes,omitempty"`
}

type Axis struct {
	// Limit the decimal numbers
	Decimals int64 `json:"decimals,omitempty"`
	// Display unit
	Format string `json:"format,omitempty"`
}

// type Legend struct {
// 	AlignAsTable bool  `json:"alignAsTable,omitempty"`
// 	Avg          bool  `json:"avg,omitempty"`
// 	Current      bool  `json:"current,omitempty" `
// 	HideEmpty    bool  `json:"hideEmpty,omitempty"`
// 	HideZero     bool  `json:"hideZero,omitempty"`
// 	Max          bool  `json:"max,omitempty"`
// 	Min          bool  `json:"min,omitempty"`
// 	RightSide    bool  `json:"rightSide,omitempty"`
// 	Show         bool  `json:"show,omitempty"`
// 	SideWidth    *uint `json:"sideWidth,omitempty"`
// 	Total        bool  `json:"total,omitempty"`
// 	Values       bool  `json:"values,omitempty"`
// }
