package converter

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"

	yamlConverter "github.com/ghodss/yaml"
	"github.com/grafana-tools/sdk"
	"github.com/mitchellh/mapstructure"
	v1alpha2 "kubesphere.io/monitoring-dashboard/api/v1alpha2"
	ansModel "kubesphere.io/monitoring-dashboard/api/v1alpha2/annotations"
	panelsModel "kubesphere.io/monitoring-dashboard/api/v1alpha2/panels"
	templatingsModel "kubesphere.io/monitoring-dashboard/api/v1alpha2/templatings"
)

type k8sDashboard struct {
	APIVersion string                  `json:"apiVersion" yaml:"apiVersion"`
	Kind       string                  `json:"kind" yaml:"kind"`
	Metadata   map[string]string       `json:"metadata" yaml:"metadata"`
	Spec       *v1alpha2.DashboardSpec `json:"spec" yaml:"spec"`
}

// Converter struct: this struct has a log property, so other newly added methods can access this log
type Converter struct {
	OutputJson []byte
	OutputYaml []byte
}

// NewConverter: new a Converter struct object with a logger object
func NewConverter() *Converter {
	return &Converter{}
}

// ConvertToDashboard converts the input json content to Dashboard model
func (converter *Converter) ConvertToDashboard(content []byte, isClusterCrd bool, ns string, name string) (*k8sDashboard, error) {
	// convert to a dashboard
	dashboard, err := converter.convert(content, isClusterCrd)
	if err != nil {
		return nil, fmt.Errorf("could parse input: %s", err.Error())
	}

	apiVersion := v1alpha2.GroupVersion.Group + "/" + v1alpha2.GroupVersion.Version

	kind := "Dashboard"
	if isClusterCrd {
		kind = "ClusterDashboard"
	}

	metadata := make(map[string]string)
	if ns == "" {
		ns = "default"
	}
	if !isClusterCrd {
		metadata["namespace"] = ns
	}
	metadata["name"] = name

	return &k8sDashboard{
		APIVersion: apiVersion,
		Kind:       kind,
		Metadata:   metadata,
		Spec:       dashboard,
	}, nil

}

// ConvertDashboardToJson converts the input json content to json bytes content
func (converter *Converter) ConvertDashboardToJson(content []byte, isClusterCrd bool, ns string, name string) error {
	manifest, err := converter.ConvertToDashboard(content, isClusterCrd, ns, name)
	if err != nil {
		return fmt.Errorf("could not convert json content to dashboard: %s", err.Error())
	}
	convertedJson, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("could not marshal dashboard to json: %s", err.Error())
	}

	converter.OutputJson = convertedJson
	return nil
}

// ConvertDashboardToYaml converts the input json content to yaml bytes content
func (converter *Converter) ConvertDashboardToYaml(content []byte, isClusterCrd bool, ns string, name string) error {
	err := converter.ConvertDashboardToJson(content, isClusterCrd, ns, name)
	if err != nil {
		return fmt.Errorf("could not marshal dashboard to json: %s", err.Error())
	}
	convertedYaml, err := yamlConverter.JSONToYAML(converter.OutputJson)
	if err != nil {
		return fmt.Errorf("could not convert json to yaml: %s", err.Error())
	}

	converter.OutputYaml = convertedYaml
	return nil
}

// ConvertFromFile converts the input json file to yaml/json bytes content
func (converter *Converter) ConvertFromFile(input io.Reader, isClusterCrd bool, ns string, name string) error {
	content, err := ioutil.ReadAll(input)
	if err != nil {
		return fmt.Errorf("could not read input: %s", err.Error())
	}

	err = converter.ConvertDashboardToYaml(content, isClusterCrd, ns, name)
	if err != nil {
		return fmt.Errorf("could not convert from input: %s", err.Error())
	}

	return nil
}

// ConvertToKubsphereDashboardManifests converts to a k8s mainfest file
func (converter *Converter) ConvertToKubsphereDashboardManifests(input io.Reader, output io.Writer, isClusterCrd bool, ns string, name string) error {

	err := converter.ConvertFromFile(input, isClusterCrd, ns, name)
	if err != nil {
		return err
	}
	_, err = output.Write([]byte(converter.OutputYaml))
	if err != nil {
		return err
	}
	return nil
}

// convert reads a input Converter file, then extract needed fields to the yaml model
func (converter *Converter) convert(content []byte, isClusterCrd bool) (*v1alpha2.DashboardSpec, error) {

	board := &sdk.Board{}
	if err := json.Unmarshal(content, board); err != nil {
		return nil, fmt.Errorf("could not unmarshall dashboard: %s", err.Error())
	}

	// a yaml model
	dashboard := &v1alpha2.DashboardSpec{}

	// starts to convert general settings
	converter.convertGeneralSettings(board, dashboard)
	// starts to convert templating variables
	converter.convertVariables(board.Templating.List, dashboard)
	// starts to convert annotations
	converter.convertAnnotations(board.Annotations.List, dashboard)
	// starts to convert pannels
	converter.convertPanels(board.Panels, dashboard, isClusterCrd)
	// starts to convert rows
	// some dashboards only include rows
	converter.convertRows(board.Rows, dashboard, isClusterCrd)

	return dashboard, nil
}

// convert GeneralSettings
func (converter *Converter) convertGeneralSettings(board *sdk.Board, dashboard *v1alpha2.DashboardSpec) {
	dashboard.Title = board.Title
	dashboard.Editable = board.Editable
	dashboard.SharedCrosshair = board.SharedCrosshair
	dashboard.Tags = board.Tags
	dashboard.Time.From = board.Time.From
	dashboard.Time.To = board.Time.To
	dashboard.Timezone = board.Timezone

	if board.Refresh != nil {
		dashboard.AutoRefresh = board.Refresh.Value
	}
}

// convert Annotations
func (converter *Converter) convertAnnotations(annotations []sdk.Annotation, dashboard *v1alpha2.DashboardSpec) {
	for _, annotation := range annotations {
		// grafana-sdk doesn't expose the "builtIn" field, so we work around that by skipping
		// the annotation we know to be built-in by its name
		if annotation.Name == "Annotations & Alerts" {
			continue
		}

		if annotation.Type != "tags" {
			continue
		}

		datasource := ""
		if annotation.Datasource != nil {
			datasource = *annotation.Datasource
		}

		dashboard.Annotations = append(dashboard.Annotations, ansModel.Annotation{
			Name:        annotation.Name,
			Datasource:  datasource,
			IconColor:   annotation.IconColor,
			Tags:        annotation.Tags,
			ShowLine:    annotation.ShowLine,
			LineColor:   annotation.LineColor,
			IconSize:    annotation.IconSize,
			Enable:      annotation.Enable,
			Query:       annotation.Query,
			Expr:        annotation.Expr,
			Step:        annotation.Step,
			TextField:   annotation.TextField,
			TextFormat:  annotation.TextFormat,
			TitleFormat: annotation.TitleFormat,
			TagsField:   annotation.TagsField,
			TagKeys:     annotation.TagKeys,
			Type:        annotation.Type,
		})
	}
}

// convert templating variables
func (converter *Converter) convertVariables(variables []sdk.TemplateVar, dashboard *v1alpha2.DashboardSpec) {
	for _, variable := range variables {
		if variable.Query == nil {
			continue
		}
		q, ok := variable.Query.(string)
		if !ok {
			continue
		}
		var options []templatingsModel.Option
		for _, op := range variable.Options {
			options = append(options, templatingsModel.Option{
				Text:     op.Text,
				Value:    op.Value,
				Selected: op.Selected,
			})
		}
		v := templatingsModel.TemplateVar{
			Name:        variable.Name,
			Type:        variable.Type,
			Auto:        variable.Auto,
			AutoCount:   variable.AutoCount,
			Datasource:  variable.Datasource,
			Options:     options,
			Query:       q,
			IncludeAll:  variable.IncludeAll,
			AllFormat:   variable.AllFormat,
			AllValue:    variable.AllValue,
			Multi:       variable.Multi,
			MultiFormat: variable.MultiFormat,
			Regex:       variable.Regex,
			Label:       variable.Label,
			Hide:        variable.Hide,
			Sort:        variable.Sort,
		}
		dashboard.Templatings = append(dashboard.Templatings, v)
	}
}

//convert rows
func (converter *Converter) convertPanels(panels []*sdk.Panel, dashboard *v1alpha2.DashboardSpec, isClusterCrd bool) {

	for _, panel := range panels {
		if panel.Type == "row" {
			for _, rowPanel := range panel.Panels {
				convertedPanel, ok := converter.convertDataPanel(rowPanel, isClusterCrd)
				if ok {
					dashboard.Panels = append(dashboard.Panels, convertedPanel)
				}
			}
		} else {
			convertedPanel, ok := converter.convertDataPanel(*panel, isClusterCrd)
			if ok {
				dashboard.Panels = append(dashboard.Panels, convertedPanel)
			}
		}
	}

}

//convert rows
func (converter *Converter) convertRows(rows []*sdk.Row, dashboard *v1alpha2.DashboardSpec, isClusterCrd bool) {

	for _, row := range rows {
		if row == nil {
			continue
		}
		panels := row.Panels
		if panels == nil || len(rows) == 0 {
			continue
		}
		for _, pl := range panels {
			convertedPanel, ok := converter.convertDataPanel(pl, isClusterCrd)
			if ok {
				dashboard.Panels = append(dashboard.Panels, convertedPanel)
			}
		}
	}
}

// convert different types of the given panel
func (converter *Converter) convertDataPanel(panel sdk.Panel, isClusterCrd bool) (*panelsModel.Panel, bool) {
	switch panel.Type {
	case "graph":
		return converter.convertGraph(panel, isClusterCrd), true
	case "singlestat":
		return converter.convertSingleStat(panel, isClusterCrd), true
	case "bargauge":
		return converter.convertBarGauge(panel, isClusterCrd), true
	case "table":
		return converter.convertTable(panel, isClusterCrd), true
	case "text":
		return converter.convertText(panel), true
	default:
		if panel.OfType == sdk.CustomType {
			return converter.convertCustom(panel, isClusterCrd), true
		}
	}
	return &panelsModel.Panel{}, false
}

// a graph panel
func (converter *Converter) convertGraph(panel sdk.Panel, isClusterCrd bool) *panelsModel.Panel {
	// filled with values of the given fields
	var height *string
	if panel.Height != nil {
		var h = panel.Height.(string)
		height = &h
	}
	graph := &panelsModel.Panel{
		CommonPanel: panelsModel.CommonPanel{
			Title:       panel.Title,
			Id:          int64(panel.ID),
			Type:        panel.Type,
			Description: panel.CommonPanel.Description,
			Height:      height,
			Datasource:  panel.Datasource,
			Colors:      defaultColors(),
		},
	}

	if panel.GraphPanel == nil {
		return graph
	}

	graph.CommonPanel.Decimals = uintpointToInt64point(panel.GraphPanel.Decimals)
	graph.CommonPanel.Legend = converter.convertLegend(panel.GraphPanel.Legend)

	graph.GraphPanel = &panelsModel.GraphPanel{
		Bars:  panel.GraphPanel.Bars,
		Lines: panel.GraphPanel.Lines,
		Stack: panel.GraphPanel.Stack,
		Xaxis: panelsModel.Axis{
			Format:   panel.GraphPanel.Xaxis.Format,
			Decimals: int64(panel.GraphPanel.Xaxis.Decimals),
		},
	}

	// converts target
	if panel.GraphPanel.Targets != nil && len(panel.GraphPanel.Targets) > 0 {
		for index, target := range panel.GraphPanel.Targets {
			graphTarget := converter.convertTarget(target, index)
			if graphTarget == nil {
				continue
			}

			graph.CommonPanel.Targets = append(graph.CommonPanel.Targets, *graphTarget)
		}

	}

	// converts yaxes
	for _, yaxis := range panel.GraphPanel.Yaxes {
		graph.GraphPanel.Yaxes = append(graph.GraphPanel.Yaxes, panelsModel.Axis{
			Format:   handleGraphFormat(yaxis.Format),
			Decimals: int64(yaxis.Decimals),
		})
		break
	}

	return graph
}

func (converter *Converter) convertLegend(sdkLegend sdk.Legend) []string {
	var legend []string

	if !sdkLegend.Show {
		legend = append(legend, "hide")
	}
	if sdkLegend.AlignAsTable {
		legend = append(legend, "as_table")
	}
	if sdkLegend.RightSide {
		legend = append(legend, "to_the_right")
	}
	if sdkLegend.Min {
		legend = append(legend, "min")
	}
	if sdkLegend.Max {
		legend = append(legend, "max")
	}
	if sdkLegend.Avg {
		legend = append(legend, "avg")
	}
	if sdkLegend.Current {
		legend = append(legend, "current")
	}
	if sdkLegend.Total {
		legend = append(legend, "total")
	}
	if sdkLegend.HideEmpty {
		legend = append(legend, "no_null_series")
	}
	if sdkLegend.HideZero {
		legend = append(legend, "no_zero_series")
	}

	return legend
}

// singlestat panel
func (converter *Converter) convertSingleStat(panel sdk.Panel, isClusterCrd bool) *panelsModel.Panel {
	var height *string
	if panel.Height != nil {
		var h = panel.Height.(string)
		height = &h
	}
	singleStat := &panelsModel.Panel{
		CommonPanel: panelsModel.CommonPanel{
			Title:       panel.Title,
			Id:          int64(panel.ID),
			Type:        panel.Type,
			Description: panel.CommonPanel.Description,
			Height:      height,
			Datasource:  panel.Datasource,
		},
	}

	if panel.SinglestatPanel == nil {
		return singleStat
	}

	singleStat.CommonPanel.Format = panel.SinglestatPanel.Format
	singleStat.CommonPanel.Decimals = intToInt64point(panel.SinglestatPanel.Decimals)

	singleStat.SinglestatPanel = &panelsModel.SinglestatPanel{
		ValueName: panel.SinglestatPanel.ValueName,
	}

	if len(panel.SinglestatPanel.Colors) == 3 {
		singleStat.CommonPanel.Colors = []string{
			panel.SinglestatPanel.Colors[0],
			panel.SinglestatPanel.Colors[1],
			panel.SinglestatPanel.Colors[2],
		}
	} else {
		singleStat.CommonPanel.Colors = defaultColors()
	}

	if panel.SinglestatPanel.SparkLine.Show && panel.SinglestatPanel.SparkLine.Full {
		singleStat.SparkLine = "full"
	}
	if panel.SinglestatPanel.SparkLine.Show && !panel.SinglestatPanel.SparkLine.Full {
		singleStat.SparkLine = "bottom"
	}

	// handles targets
	if panel.SinglestatPanel.Targets != nil && len(panel.SinglestatPanel.Targets) > 0 {
		for index, target := range panel.SinglestatPanel.Targets {
			target := converter.convertTarget(target, index)
			if target == nil {
				continue
			}

			singleStat.CommonPanel.Targets = append(singleStat.CommonPanel.Targets, *target)
		}

	}

	// handles gauge
	singleStat.Gauge = panelsModel.Gauge{
		MaxValue:         int64(panel.SinglestatPanel.Gauge.MaxValue),
		MinValue:         int64(panel.SinglestatPanel.Gauge.MinValue),
		Show:             panel.SinglestatPanel.Gauge.Show,
		ThresholdLabels:  panel.SinglestatPanel.Gauge.ThresholdLabels,
		ThresholdMarkers: panel.SinglestatPanel.Gauge.ThresholdMarkers,
	}

	return singleStat
}

// gauge
func (converter *Converter) convertCustom(panel sdk.Panel, isClusterCrd bool) *panelsModel.Panel {
	// set options
	var height *string
	if panel.Height != nil {
		var h = panel.Height.(string)
		height = &h
	}
	customPanel := &panelsModel.Panel{
		CommonPanel: panelsModel.CommonPanel{
			Title:       panel.Title,
			Id:          int64(panel.ID),
			Type:        "singlestat",
			Description: panel.CommonPanel.Description,
			Height:      height,
			Datasource:  panel.Datasource,
		},
	}

	if panel.CustomPanel == nil {
		return customPanel
	}

	var sdkTargets []sdk.Target

	custom := *panel.CustomPanel

	if err := mapstructure.Decode(custom["targets"], &sdkTargets); err != nil {
		return customPanel
	}

	var targets []panelsModel.Target

	for index, target := range sdkTargets {
		t := converter.convertTarget(target, index)
		if t == nil {
			continue
		}
		targets = append(targets, *t)
	}

	customPanel.CommonPanel.Targets = targets

	return customPanel
}

// bar gauge
func (converter *Converter) convertBarGauge(panel sdk.Panel, isClusterCrd bool) *panelsModel.Panel {
	// set options
	var height *string
	if panel.Height != nil {
		var h = panel.Height.(string)
		height = &h
	}
	barGaugePanel := &panelsModel.Panel{
		CommonPanel: panelsModel.CommonPanel{
			Title:       panel.Title,
			Id:          int64(panel.ID),
			Type:        panel.Type,
			Description: panel.CommonPanel.Description,
			Height:      height,
			Datasource:  panel.Datasource,
		},
	}

	if panel.BarGaugePanel == nil {
		return barGaugePanel
	}

	barGaugePanel.BarGaugePanel = &panelsModel.BarGaugePanel{
		Options: &panelsModel.BarGaugeOptions{
			Orientation: panel.BarGaugePanel.Options.Orientation,
			TextMode:    panel.BarGaugePanel.Options.TextMode,
			ColorMode:   panel.BarGaugePanel.Options.ColorMode,
			GraphMode:   panel.BarGaugePanel.Options.GraphMode,
			JustifyMode: panel.BarGaugePanel.Options.JustifyMode,
			DisplayMode: panel.BarGaugePanel.Options.DisplayMode,
			Content:     panel.BarGaugePanel.Options.Content,
			Mode:        panel.BarGaugePanel.Options.Mode,
		},
	}

	// handles targets
	if panel.BarGaugePanel.Targets != nil && len(panel.BarGaugePanel.Targets) > 0 {
		for index, target := range panel.BarGaugePanel.Targets {
			barGaugeTarget := converter.convertTarget(target, index)
			if barGaugeTarget == nil {
				continue
			}

			barGaugePanel.CommonPanel.Targets = append(barGaugePanel.CommonPanel.Targets, *barGaugeTarget)
		}

	}

	return barGaugePanel
}

// converts a table panel
func (converter *Converter) convertTable(panel sdk.Panel, isClusterCrd bool) *panelsModel.Panel {
	var height *string
	if panel.Height != nil {
		var h = panel.Height.(string)
		height = &h
	}
	tablePanel := &panelsModel.Panel{
		CommonPanel: panelsModel.CommonPanel{
			Title:       panel.Title,
			Id:          int64(panel.ID),
			Type:        panel.Type,
			Description: panel.CommonPanel.Description,
			Height:      height,
			Datasource:  panel.Datasource,
		},
	}

	if panel.TablePanel == nil {
		return tablePanel
	}

	tablePanel.TablePanel = &panelsModel.TablePanel{
		Scroll: panel.TablePanel.Scroll,
	}

	if panel.TablePanel.Targets != nil && len(panel.TablePanel.Targets) > 0 {
		for index, target := range panel.TablePanel.Targets {
			graphTarget := converter.convertTarget(target, index)
			if graphTarget == nil {
				continue
			}

			tablePanel.CommonPanel.Targets = append(tablePanel.CommonPanel.Targets, *graphTarget)
		}
	}

	if panel.TablePanel.Sort != nil {
		tablePanel.TablePanel.Sort = &panelsModel.Sort{
			Col:  panel.TablePanel.Sort.Col,
			Desc: panel.TablePanel.Sort.Desc,
		}
	}

	return tablePanel
}

// converts a text panel
func (converter *Converter) convertText(panel sdk.Panel) *panelsModel.Panel {
	var height *string
	if panel.Height != nil {
		var h = panel.Height.(string)
		height = &h
	}

	textPanel := &panelsModel.Panel{
		CommonPanel: panelsModel.CommonPanel{
			Title:       panel.Title,
			Id:          int64(panel.ID),
			Type:        panel.Type,
			Description: panel.CommonPanel.Description,
			Height:      height,
			Datasource:  panel.Datasource,
		},
	}

	if panel.TextPanel == nil {
		return textPanel
	}

	textPanel.TextPanel = &panelsModel.TextPanel{
		Mode:    panel.TextPanel.Mode,
		Content: panel.TextPanel.Content,
	}

	return textPanel
}

func (converter *Converter) convertTarget(target sdk.Target, index int) *panelsModel.Target {
	// looks like a prometheus target
	return converter.convertPrometheusTarget(target, index)
}

func (converter *Converter) convertPrometheusTarget(target sdk.Target, index int) *panelsModel.Target {
	t := &panelsModel.Target{
		// RefID: target.RefID,
		RefID:        int64(index) + 1,
		LegendFormat: handleLegendFormat(target.LegendFormat),
	}

	// adjusts the query expression to adapt to the ks cluster
	converedExpr := convertExpr(target.Expr)
	if converedExpr == "" {
		t.Expression = target.Expr
		return t
	}

	t.Expression = fmt.Sprintf("%s", converedExpr)
	t.Step = toString(target.Step)
	return t

}

func panelSpan(panel sdk.Panel) int64 {
	return int64(panel.ID)
}

func defaultOption(opt sdk.Current) string {
	if opt.Value == nil {
		return ""
	}

	return opt.Value.(string)
}

func handleZimu(z string) int64 {
	n, _ := strconv.Atoi(z)
	return int64(n) + 1
}

func toString(step int) string {
	// number := int(step / 60)
	// if number == 0 {
	// 	return strconv.Itoa(step) + "s"
	// }
	// if number > 60 {
	// 	return strconv.Itoa(number/60) + "h"
	// }
	// return strconv.Itoa(number) + "m"
	return "1m"
}

func handleGraphFormat(f string) string {

	if f == "bytes" || f == "Bps" {
		f = "Byte"
	} else if f == "percent" || f == "percentunit" {
		f = "percent (0.0-1.0)"
	} else {
		f = "none"
	}
	return f
}

func handleLegendFormat(l string) string {
	badPat := regexp.MustCompile(`\{(\s+\w+\s+)\}`)
	if match := badPat.Match([]byte(l)); match {
		f := func(s string) string {
			stripReg := regexp.MustCompile(`\s+`)
			return stripReg.ReplaceAllString(s, "")
		}
		return badPat.ReplaceAllStringFunc(l, f)
	}
	return l
}

func pointToString(des *string) string {
	d := ""
	if des != nil {
		d = *des
	}
	return d
}

func uintpointToInt64point(up *uint) *int64 {
	if up != nil {
		var t = int64(*up)
		return &t
	}
	return nil
}

func intToInt64point(o int) *int64 {
	var c = int64(o)
	return &c
}

func defaultColors() []string {
	return []string{"#60acfc", "#23c2db", "#64d5b2", "#d5ec5a", "#ffb64e", "#fb816d", "#d15c7f"}
}

func convertExpr(expr string) string {

	// free the door if don't match a `[{}]` regex style
	pat := regexp.MustCompile(`[\{\}]`)
	if !pat.Match([]byte(expr)) {
		return ""
	}
	// handles $interval or $__interval
	pat1 := regexp.MustCompile(`\$_{0,2}interval`)
	if pat1.Match([]byte(expr)) {
		expr = pat1.ReplaceAllString(expr, "3m")
	}

	// if contains irate/rate/count func, just removes `\{.*\}`
	pat2 := regexp.MustCompile(`\{.*?\}`)
	if matchCommon := pat2.Match([]byte(expr)); matchCommon {
		expr = pat2.ReplaceAllString(expr, "")
	}

	// if contains count, removes `>\d+`
	pat3 := regexp.MustCompile(`>\d+`)
	if matchCount := pat3.Match([]byte(expr)); matchCount {
		expr = pat3.ReplaceAllString(expr, "")
	}
	return expr

}
