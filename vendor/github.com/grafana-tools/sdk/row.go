package sdk

/*
   Copyright 2016 Alexander I.Grafov <grafov@gmail.com>
   Copyright 2016-2019 The Grafana SDK authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

	   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

   ॐ तारे तुत्तारे तुरे स्व
*/

// Row represents single row of Grafana dashboard.
type Row struct {
	Title     string  `json:"title"`
	ShowTitle bool    `json:"showTitle"`
	Collapse  bool    `json:"collapse"`
	Editable  bool    `json:"editable"`
	Height    Height  `json:"height"`
	Panels    []Panel `json:"panels"`
	Repeat    *string `json:"repeat"`
}

var lastPanelID uint

func (r *Row) Add(panel *Panel) {
	lastPanelID++
	panel.ID = lastPanelID
	r.Panels = append(r.Panels, *panel)
}

func (r *Row) AddDashlist(data *DashlistPanel) {
	lastPanelID++
	panel := NewDashlist("")
	panel.ID = lastPanelID
	panel.DashlistPanel = data
	r.Panels = append(r.Panels, *panel)
}

func (r *Row) AddGraph(data *GraphPanel) {
	lastPanelID++
	panel := NewGraph("")
	panel.ID = lastPanelID
	panel.GraphPanel = data
	r.Panels = append(r.Panels, *panel)
}

func (r *Row) AddTable(data *TablePanel) {
	lastPanelID++
	panel := NewTable("")
	panel.ID = lastPanelID
	panel.TablePanel = data
	r.Panels = append(r.Panels, *panel)
}

func (r *Row) AddText(data *TextPanel) {
	lastPanelID++
	panel := NewText("")
	panel.ID = lastPanelID
	panel.TextPanel = data
	r.Panels = append(r.Panels, *panel)
}

func (r *Row) AddStat(data *StatPanel) {
	lastPanelID++
	panel := NewStat("")
	panel.ID = lastPanelID
	panel.StatPanel = data
	r.Panels = append(r.Panels, *panel)
}

func (r *Row) AddSinglestat(data *SinglestatPanel) {
	lastPanelID++
	panel := NewSinglestat("")
	panel.ID = lastPanelID
	panel.SinglestatPanel = data
	r.Panels = append(r.Panels, *panel)
}

func (r *Row) AddCustom(data *CustomPanel) {
	lastPanelID++
	panel := NewCustom("")
	panel.ID = lastPanelID
	panel.CustomPanel = data
	r.Panels = append(r.Panels, *panel)
}
