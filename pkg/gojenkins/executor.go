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

type Executor struct {
	Raw     *ExecutorResponse
	Jenkins *Jenkins
}
type ViewData struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
type ExecutorResponse struct {
	AssignedLabels  []struct{}  `json:"assignedLabels"`
	Description     interface{} `json:"description"`
	Jobs            []InnerJob  `json:"jobs"`
	Mode            string      `json:"mode"`
	NodeDescription string      `json:"nodeDescription"`
	NodeName        string      `json:"nodeName"`
	NumExecutors    int64       `json:"numExecutors"`
	OverallLoad     struct{}    `json:"overallLoad"`
	PrimaryView     struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"primaryView"`
	QuietingDown   bool       `json:"quietingDown"`
	SlaveAgentPort int64      `json:"slaveAgentPort"`
	UnlabeledLoad  struct{}   `json:"unlabeledLoad"`
	UseCrumbs      bool       `json:"useCrumbs"`
	UseSecurity    bool       `json:"useSecurity"`
	Views          []ViewData `json:"views"`
}
