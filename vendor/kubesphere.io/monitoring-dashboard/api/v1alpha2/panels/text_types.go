// +kubebuilder:object:generate=true

package panels

// dashboard text type
// referers to https://pkg.go.dev/github.com/K-Phoen/grabana/decoder#DashboardText
type TextPanel struct {
	Mode    string `json:"mode,omitempty"`
	Content string `json:"content,omitempty"`
}
