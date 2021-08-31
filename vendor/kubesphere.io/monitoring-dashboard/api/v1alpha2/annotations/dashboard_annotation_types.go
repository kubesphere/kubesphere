// +kubebuilder:object:generate=true

package anotations

// "github.com/grafana-tools/sdk,omitempty"
// refers to https://pkg.go.dev/github.com/grafana-tools/sdk#Annotation

type Annotation struct {
	Name        string   `json:"name,omitempty" yaml:"name,omitempty"`
	Datasource  string   `json:"datasource,omitempty" yaml:"datasource,omitempty"`
	ShowLine    bool     `json:"showLine,omitempty" yaml:"showLine,omitempty"`
	IconColor   string   `json:"iconColor,omitempty" yaml:"iconColor,omitempty"`
	LineColor   string   `json:"lineColor,omitempty" yaml:"lineColor,omitempty"`
	IconSize    uint     `json:"iconSize,omitempty" yaml:"iconSize,omitempty"`
	Enable      bool     `json:"enable,omitempty" yaml:"enable,omitempty"`
	Query       string   `json:"query,omitempty" yaml:"query,omitempty"`
	Expr        string   `json:"expr,omitempty" yaml:"expr,omitempty"`
	Step        string   `json:"step,omitempty" yaml:"step,omitempty"`
	TextField   string   `json:"textField,omitempty" yaml:"textField,omitempty"`
	TextFormat  string   `json:"textFormat,omitempty" yaml:"textFormat,omitempty"`
	TitleFormat string   `json:"titleFormat,omitempty" yaml:"titleFormat,omitempty"`
	TagsField   string   `json:"tagsField,omitempty" yaml:"tagsField,omitempty"`
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	TagKeys     string   `json:"tagKeys,omitempty" yaml:"tagKeys,omitempty"`
	Type        string   `json:"type,omitempty" yaml:"type,omitempty"`
}
