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

type Label struct {
	Raw     *LabelResponse
	Jenkins *Jenkins
	Base    string
}

type MODE string

const (
	NORMAL    MODE = "NORMAL"
	EXCLUSIVE      = "EXCLUSIVE"
)

type LabelNode struct {
	NodeName        string `json:"nodeName"`
	NodeDescription string `json:"nodeDescription"`
	NumExecutors    int64  `json:"numExecutors"`
	Mode            string `json:"mode"`
	Class           string `json:"_class"`
}

type LabelResponse struct {
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	Nodes          []LabelNode `json:"nodes"`
	Offline        bool        `json:"offline"`
	IdleExecutors  int64       `json:"idleExecutors"`
	BusyExecutors  int64       `json:"busyExecutors"`
	TotalExecutors int64       `json:"totalExecutors"`
}

func (l *Label) GetName() string {
	return l.Raw.Name
}

func (l *Label) GetNodes() []LabelNode {
	return l.Raw.Nodes
}

func (l *Label) Poll() (int, error) {
	response, err := l.Jenkins.Requester.GetJSON(l.Base, l.Raw, nil)
	if err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}
