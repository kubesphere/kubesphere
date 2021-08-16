// +kubebuilder:object:generate=true

package panels

// Graph visualizes range query results into a linear graph
type Graph struct {
	// Name of the graph panel
	//Title string `json:"title,omitempty"`
	// Must be `graph`
	//Type string `json:"type"`
	// Panel ID
	//Id int64 `json:"id,omitempty"`
	// Panel description
	Description string `json:"description,omitempty"`
	// A collection of queries
	//Targets []Target `json:"targets,omitempty"`
	// Display as a bar chart
	Bars bool `json:"bars,omitempty"`
	// Set series color
	Colors []string `json:"colors,omitempty"`
	// Display as a line chart
	Lines bool `json:"lines,omitempty"`
	// Display as a stacked chart
	Stack bool `json:"stack,omitempty"`
	// Y-axis options
	Yaxes []Yaxis `json:"yaxes,omitempty"`
}

type Yaxis struct {
	// Limit the decimal numbers
	Decimals int64 `json:"decimals,omitempty"`
	// Display unit
	Format string `json:"format,omitempty"`
}
