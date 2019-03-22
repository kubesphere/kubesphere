/*

 Copyright 2019 The KubeSphere Authors.

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

package log

import (
	fb "kubesphere.io/kubesphere/pkg/simple/client/fluentbit"
	"time"
)

type FluentbitCRDResult struct {
	Status int              `json:"status"`
	CRD    fb.FluentBitSpec `json:"CRD,omitempty"`
}

type FluentbitCRDDeleteResult struct {
	Status int `json:"status"`
}

type FluentbitSettingsResult struct {
	Status int    `json:"status"`
	Enable string `json:"Enable,omitempty"`
}

type FluentbitFilter struct {
	Type       string `json:"type"`
	Field      string `json:"field"`
	Expression string `json:"expression"`
}

type FluentbitFiltersResult struct {
	Status  int               `json:"status"`
	Filters []FluentbitFilter `json:"filters,omitempty"`
}

type FluentbitOutputsResult struct {
	Status  int               `json:"status"`
	Outputs []fb.OutputPlugin `json:"outputs,omitempty"`
}

type OutputDBBinding struct {
	Id         uint   `gorm:"primary_key;auto_increment;unique"`
	Type       string `gorm:"not null"`
	Name       string `gorm:"not null"`
	Parameters string `gorm:"not null"`
	Internal   bool
	Enable     bool      `gorm:"not null"`
	Updatetime time.Time `gorm:"not null"`
}