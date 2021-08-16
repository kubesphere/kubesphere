// +kubebuilder:object:generate=true

package panels

// a table panel
type TablePanel struct {
	Sort   *Sort `json:"sort,omitempty"`
	Scroll bool  `json:"scroll,omitempty"`
}

type Sort struct {
	Col  int  `json:"col,omitempty"`
	Desc bool `json:"desc,omitempty"`
}
