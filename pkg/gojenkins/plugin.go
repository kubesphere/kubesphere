// Copyright 2015 Vadim Kravcenko
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package gojenkins

import (
	"strconv"
)

type Plugins struct {
	Jenkins *Jenkins
	Raw     *PluginResponse
	Base    string
	Depth   int
}

type PluginResponse struct {
	Plugins []Plugin `json:"plugins"`
}

type Plugin struct {
	Active        bool        `json:"active"`
	BackupVersion interface{} `json:"backupVersion"`
	Bundled       bool        `json:"bundled"`
	Deleted       bool        `json:"deleted"`
	Dependencies  []struct {
		Optional  string `json:"optional"`
		ShortName string `json:"shortname"`
		Version   string `json:"version"`
	} `json:"dependencies"`
	Downgradable        bool   `json:"downgradable"`
	Enabled             bool   `json:"enabled"`
	HasUpdate           bool   `json:"hasUpdate"`
	LongName            string `json:"longName"`
	Pinned              bool   `json:"pinned"`
	ShortName           string `json:"shortName"`
	SupportsDynamicLoad string `json:"supportsDynamicLoad"`
	URL                 string `json:"url"`
	Version             string `json:"version"`
}

func (p *Plugins) Count() int {
	return len(p.Raw.Plugins)
}

func (p *Plugins) Contains(name string) *Plugin {
	for _, p := range p.Raw.Plugins {
		if p.LongName == name || p.ShortName == name {
			return &p
		}
	}
	return nil
}

func (p *Plugins) Poll() (int, error) {
	qr := map[string]string{
		"depth": strconv.Itoa(p.Depth),
	}
	response, err := p.Jenkins.Requester.GetJSON(p.Base, p.Raw, qr)
	if err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}
