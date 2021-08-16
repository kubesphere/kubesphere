// +kubebuilder:object:generate=true

package time

// Time ranges of the metrics for display
type Time struct {
	// Start time in the format of `^now([+-][0-9]+[smhdwMy])?$`, eg. `now-1M`.
	// It denotes the end time is set to the last month since now.
	From string `json:"from,omitempty" json:"from,omitempty"`
	// End time in the format of `^now([+-][0-9]+[smhdwMy])?$`, eg. `now-1M`.
	// It denotes the start time is set to the last month since now.
	To string `json:"to,omitempty" json:"to,omitempty"`
}
