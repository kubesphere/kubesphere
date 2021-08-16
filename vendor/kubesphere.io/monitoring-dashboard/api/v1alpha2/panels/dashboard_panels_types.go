// +kubebuilder:object:generate=true

package panels

import (
	"encoding/json"
	"errors"
	"reflect"
)

// Supported panel type
// type Panel struct {
// 	// private property of the bargauge panel
// 	Options *BarGaugeOptions `json:"options,omitempty"`

// 	// ****common properties start
// 	// Name pf the panel
// 	Title string `json:"title,omitempty"`
// 	// The type of the panel
// 	Type string `json:"type,omitempty"`
// 	// Panel ID
// 	Id int64 `json:"id,omitempty"`
// 	// Panel description
// 	Description string `json:"description,omitempty"`
// 	// datasource
// 	Datasource string `json:"datasource,omitempty"`
// 	// A collection of queries
// 	Targets []Target `json:"targets,omitempty"`
// 	// Display as a bar chart
// 	Bars bool `json:"bars,omitempty"`
// 	// Set series color
// 	Colors []string `json:"colors,omitempty"`
// 	// color settings
// 	// todo graph
// 	Color []string `json:"color,omitempty"`
// 	// Display as a line chart
// 	Lines bool `json:"lines,omitempty"`
// 	// Display as a stacked chart
// 	Stack bool `json:"stack,omitempty"`
// 	// legend
// 	Legend []string `json:"legend,omitempty,flow"`
// 	// height
// 	Height string `json:"height,omitempty"`
// 	// transparent
// 	Transparent            bool     `json:"transparent,omitempty"`
// 	HiddenColumns          []string `json:"hidden_columns,omitempty,flow"`
// 	TimeSeriesAggregations []string `json:"time_series_aggregations,omitempty"`
// 	// value name
// 	ValueName string `json:"valueName,omitempty"`
// 	// *****common properties end

// 	// private property of graph panel
// 	// Y-axis options
// 	Yaxes []Yaxis `json:"yaxes,omitempty"`

// 	// private properties of singlestat panel
// 	// spark line: full or bottom
// 	SparkLine string `json:"sparkline,omitempty"`
// 	// Limit the decimal numbers
// 	Decimals int64 `json:"decimals,omitempty"`
// 	// Display unit
// 	Format string `json:"format,omitempty"`
// 	// gauge
// 	Gauge *Gauge `json:"gauge,omitempty"`

// 	// table panel has no private property

// 	// private properties of text panel
// 	HTML     string `json:"html,omitempty"`
// 	Markdown string `json:"markdown,omitempty"`
// }

type (
	Panel struct {
		CommonPanel `json:",inline"`
		*GraphPanel `json:",inline"`
		// RowPanel        *RowPanel        `json:",inline"`
		*SinglestatPanel `json:",inline"`
		*TablePanel      `json:",inline"`
		*TextPanel       `json:",inline"`
		*BarGaugePanel   `json:",inline"`
		// *CustomPanel     `json:",inline"`
	}
	probePanel struct {
		CommonPanel
	}

	// CustomPanel map[string]runtime.RawExtension
)

func (p *Panel) UnmarshalJSON(b []byte) (err error) {
	var probe probePanel
	if err = json.Unmarshal(b, &probe); err == nil {
		p.CommonPanel = probe.CommonPanel
		switch probe.Type {
		case "graph":
			var graph GraphPanel
			if err = json.Unmarshal(b, &graph); err == nil {
				if !isZero(reflect.ValueOf(graph)) {
					p.GraphPanel = &graph
				}
			}
		case "table":
			var table TablePanel
			if err = json.Unmarshal(b, &table); err == nil {
				if !isZero(reflect.ValueOf(table)) {
					p.TablePanel = &table
				}
			}
		case "text":
			var text TextPanel
			if err = json.Unmarshal(b, &text); err == nil {
				if !isZero(reflect.ValueOf(text)) {
					p.TextPanel = &text
				}
			}
		case "singlestat":
			var singlestat SinglestatPanel
			if err = json.Unmarshal(b, &singlestat); err == nil {
				if !isZero(reflect.ValueOf(singlestat)) {
					p.SinglestatPanel = &singlestat
				}
			}
		case "bargauge":
			var bargauge BarGaugePanel
			if err = json.Unmarshal(b, &bargauge); err == nil {
				if !isZero(reflect.ValueOf(bargauge)) {
					p.BarGaugePanel = &bargauge
				}
			}
		case "row":
			// var row RowPanel
			// if err = json.Unmarshal(b, &row); err == nil {
			// 	p.RowPanel = &row
			// }
			// default:
			// 	var custom = make(CustomPanel)
			// 	if err = json.Unmarshal(b, &custom); err == nil {
			// 		p.CustomPanel = &custom
			// 	}
		}
	}
	return
}

func (p *Panel) MarshalJSON() ([]byte, error) {
	var outCommon = struct {
		CommonPanel
	}{p.CommonPanel}

	switch p.Type {
	case "graph":
		if p.GraphPanel == nil {
			return json.Marshal(outCommon)
		}
		var outGraph = struct {
			CommonPanel
			GraphPanel
		}{p.CommonPanel, *p.GraphPanel}
		return json.Marshal(outGraph)
	case "table":
		if p.TablePanel == nil {
			return json.Marshal(outCommon)
		}
		var outTable = struct {
			CommonPanel
			TablePanel
		}{p.CommonPanel, *p.TablePanel}
		return json.Marshal(outTable)
	case "text":
		if p.TextPanel == nil {
			return json.Marshal(outCommon)
		}
		var outText = struct {
			CommonPanel
			TextPanel
		}{p.CommonPanel, *p.TextPanel}
		return json.Marshal(outText)
	case "singlestat":
		if p.SinglestatPanel == nil {
			return json.Marshal(outCommon)
		}
		var outSinglestat = struct {
			CommonPanel
			SinglestatPanel
		}{p.CommonPanel, *p.SinglestatPanel}
		return json.Marshal(outSinglestat)

	case "bargauge":
		if p.BarGaugePanel == nil {
			return json.Marshal(outCommon)
		}
		var outBarGauge = struct {
			CommonPanel
			BarGaugePanel
		}{p.CommonPanel, *p.BarGaugePanel}
		return json.Marshal(outBarGauge)
	case "row":
		var outRow = struct {
			CommonPanel
		}{p.CommonPanel}
		return json.Marshal(outRow)
		// default:
		// 	var outCustom = struct {
		// 		CommonPanel
		// 		CustomPanel
		// 	}{p.CommonPanel, *p.CustomPanel}
		// 	return json.Marshal(outCustom)
	}
	return nil, errors.New("can't marshal unknown panel type")
}

func isZero(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	}
	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
}
