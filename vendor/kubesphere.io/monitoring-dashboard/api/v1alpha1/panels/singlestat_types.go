// +kubebuilder:object:generate=true

package panels

// SingleStat shows instant query result
type SingleStat struct {
	// Name of the signlestat panel
	//Title string `json:"title,omitempty"`
	// Must be `singlestat`
	//Type string `json:"type"`
	// Panel ID
	//Id int64 `json:"id,omitempty"`
	// A collection of queries
	//Targets []Target `json:"targets,omitempty"`
	// Limit the decimal numbers
	Decimals *int64 `json:"decimals,omitempty"`
	// Display unit
	Format string `json:"format,omitempty"`
}
