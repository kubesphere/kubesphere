/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	"fmt"

	"kubesphere.io/monitoring-dashboard/api/v1alpha1"
	v1alpha1panels "kubesphere.io/monitoring-dashboard/api/v1alpha1/panels"
	v1alpha2panels "kubesphere.io/monitoring-dashboard/api/v1alpha2/panels"
	v1alpha2templatings "kubesphere.io/monitoring-dashboard/api/v1alpha2/templatings"
	"kubesphere.io/monitoring-dashboard/api/v1alpha2/time"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (_ *Dashboard) Hub() {}

func (src *Dashboard) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha1.Dashboard)
	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.Title = src.Spec.Title
	dst.Spec.Description = src.Spec.Description
	dst.Spec.Time = v1alpha1.Time{
		From: src.Spec.Time.From,
		To:   src.Spec.Time.To,
	}
	pls := []v1alpha1.Panel{}

	for _, panel := range src.Spec.Panels {

		if panel == nil {
			continue
		}

		t := v1alpha1.PanelType(panel.Type)
		if dst.Spec.DataSource == "" && panel.CommonPanel.Datasource != nil {
			dst.Spec.DataSource = point2string(panel.CommonPanel.Datasource)
		}

		dstPanel := v1alpha1.Panel{
			PanelMeta: v1alpha1.PanelMeta{
				Title: panel.Title,
				Id:    panel.Id,
				Type:  t,
			},
		}

		for _, target := range panel.CommonPanel.Targets {
			dstPanel.Targets = append(dstPanel.Targets, v1alpha1panels.Target{
				Expression:   target.Expression,
				LegendFormat: target.LegendFormat,
				RefID:        target.RefID,
				Step:         target.Step,
			})
		}

		switch t {
		case "graph":

			if len(panel.CommonPanel.Colors) > 0 {
				dstPanel.Graph.Colors = panel.CommonPanel.Colors
			}

			graph := panel.GraphPanel

			if graph != nil {

				yaxes := make([]v1alpha1panels.Yaxis, 0)
				for _, yaxis := range graph.Yaxes {
					yaxes = append(yaxes, v1alpha1panels.Yaxis{
						Decimals: yaxis.Decimals,
						Format:   yaxis.Format,
					})
				}
				if len(yaxes) > 0 {
					dstPanel.Graph.Yaxes = yaxes
				}
				if panel.CommonPanel.Description != nil {
					dstPanel.Graph.Description = point2string(panel.CommonPanel.Description)
				}
				if graph.Bars {
					dstPanel.Graph.Bars = graph.Bars
				}
				if graph.Lines {
					dstPanel.Graph.Lines = graph.Lines
				}
				if graph.Stack {
					dstPanel.Graph.Stack = graph.Stack
				}

			}

		case "singlestat":

			if panel.CommonPanel.Decimals != nil {
				dstPanel.SingleStat.Decimals = panel.CommonPanel.Decimals
			}

			if panel.CommonPanel.Format != "" {
				dstPanel.SingleStat.Format = panel.CommonPanel.Format
			}

		case "row":
			// var r v1alpha1panels.Row
			// dstPanel.Row = &r
		default:
			fmt.Println("unhandled panel type: skipped，type:", t)
		}

		pls = append(pls, dstPanel)
	}

	dst.Spec.Panels = pls

	for _, temp := range src.Spec.Templatings {
		dst.Spec.Templatings = append(dst.Spec.Templatings, v1alpha1.Templating{
			Name:  temp.Name,
			Query: temp.Query,
		})
	}

	return nil

}

func (dst *Dashboard) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha1.Dashboard)
	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.Title = src.Spec.Title
	dst.Spec.Description = src.Spec.Description
	dst.Spec.Time = time.Time{
		From: src.Spec.Time.From,
		To:   src.Spec.Time.To,
	}
	pls := []*v1alpha2panels.Panel{}

	for _, panel := range src.Spec.Panels {

		t := v1alpha1.PanelType(panel.Type)

		dstPanel := v1alpha2panels.Panel{
			CommonPanel: v1alpha2panels.CommonPanel{
				Title:      panel.PanelMeta.Title,
				Id:         panel.PanelMeta.Id,
				Type:       string(panel.PanelMeta.Type),
				Datasource: &src.Spec.DataSource,
			},
		}

		for _, target := range panel.Targets {
			dstPanel.CommonPanel.Targets = append(dstPanel.CommonPanel.Targets, v1alpha2panels.Target{
				Expression:   target.Expression,
				LegendFormat: target.LegendFormat,
				RefID:        target.RefID,
				Step:         target.Step,
			})
		}

		switch t {
		case "graph":
			graph := panel.Graph
			if graph != nil {
				dstPanel.CommonPanel.Description = &graph.Description
				yaxes := make([]v1alpha2panels.Axis, 0)
				for _, yaxis := range graph.Yaxes {
					yaxes = append(yaxes, v1alpha2panels.Axis{
						Decimals: yaxis.Decimals,
						Format:   yaxis.Format,
					})
				}
				dstPanel.GraphPanel.Bars = graph.Bars
				dstPanel.GraphPanel.Lines = graph.Lines
				dstPanel.GraphPanel.Stack = graph.Stack
				dstPanel.GraphPanel.Yaxes = yaxes

			}
		case "singlestat":
			singlestat := panel.SingleStat
			if singlestat != nil {
				dstPanel.CommonPanel.Decimals = singlestat.Decimals
				dstPanel.CommonPanel.Format = singlestat.Format
			}

		case "row":
			// var r v1alpha2panels.RowPanel
			// dstPanel.RowPanel = &r
		default:
			fmt.Println("unhandled panel type: skipped，type:", t)
		}

		pls = append(pls, &dstPanel)
	}

	dst.Spec.Panels = pls

	for _, temp := range src.Spec.Templatings {
		dst.Spec.Templatings = append(dst.Spec.Templatings, v1alpha2templatings.TemplateVar{
			Name:  temp.Name,
			Query: temp.Query,
		})
	}

	return nil

}

func point2string(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
