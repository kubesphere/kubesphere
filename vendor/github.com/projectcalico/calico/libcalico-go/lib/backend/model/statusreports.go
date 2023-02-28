// Copyright (c) 2016-2018 Tigera, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/projectcalico/calico/libcalico-go/lib/errors"
)

var (
	matchActiveStatusReport = regexp.MustCompile("^/?calico/felix/v2/([^/]+)/host/([^/]+)/status$")
	matchLastStatusReport   = regexp.MustCompile("^/?calico/felix/v2/([^/]+)/host/([^/]+)/last_reported_status")
	typeStatusReport        = reflect.TypeOf(StatusReport{})
)

type ActiveStatusReportKey struct {
	Hostname     string `json:"-" validate:"required,hostname"`
	RegionString string
}

func (key ActiveStatusReportKey) defaultPath() (string, error) {
	return key.defaultDeletePath()
}

func (key ActiveStatusReportKey) defaultDeletePath() (string, error) {
	if key.Hostname == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "hostname"}
	}
	if key.RegionString == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "regionString"}
	}
	if strings.Contains(key.RegionString, "/") {
		return "", ErrorSlashInRegionString(key.RegionString)
	}
	e := fmt.Sprintf("/calico/felix/v2/%s/host/%s/status", key.RegionString, key.Hostname)
	return e, nil
}

func (key ActiveStatusReportKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key ActiveStatusReportKey) valueType() (reflect.Type, error) {
	return typeStatusReport, nil
}

func (key ActiveStatusReportKey) String() string {
	return fmt.Sprintf("StatusReport(hostname=%s)", key.Hostname)
}

type ActiveStatusReportListOptions struct {
	Hostname     string
	RegionString string
}

func (options ActiveStatusReportListOptions) defaultPathRoot() string {
	k := "/calico/felix/v2/"
	if options.RegionString == "" {
		return k
	}
	k = k + options.RegionString + "/host"
	if options.Hostname == "" {
		return k
	}
	k = k + fmt.Sprintf("/%s/status", options.Hostname)
	return k
}

func (options ActiveStatusReportListOptions) KeyFromDefaultPath(ekey string) Key {
	log.Debugf("Get StatusReport key from %s", ekey)
	r := matchActiveStatusReport.FindAllStringSubmatch(ekey, -1)
	if len(r) != 1 {
		log.Debugf("Didn't match regex")
		return nil
	}
	regionString := r[0][1]
	name := r[0][2]
	if options.RegionString != "" && regionString != options.RegionString {
		log.Debugf("Didn't match region %s != %s", options.RegionString, regionString)
		return nil
	}
	if options.Hostname != "" && name != options.Hostname {
		log.Debugf("Didn't match name %s != %s", options.Hostname, name)
		return nil
	}
	return ActiveStatusReportKey{Hostname: name, RegionString: regionString}
}

type LastStatusReportKey struct {
	Hostname     string `json:"-" validate:"required,hostname"`
	RegionString string
}

func (key LastStatusReportKey) defaultPath() (string, error) {
	return key.defaultDeletePath()
}

func (key LastStatusReportKey) defaultDeletePath() (string, error) {
	if key.Hostname == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "hostname"}
	}
	if key.RegionString == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "regionString"}
	}
	if strings.Contains(key.RegionString, "/") {
		return "", ErrorSlashInRegionString(key.RegionString)
	}
	e := fmt.Sprintf("/calico/felix/v2/%s/host/%s/last_reported_status", key.RegionString, key.Hostname)
	return e, nil
}

func (key LastStatusReportKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key LastStatusReportKey) valueType() (reflect.Type, error) {
	return typeStatusReport, nil
}

func (key LastStatusReportKey) String() string {
	return fmt.Sprintf("StatusReport(hostname=%s)", key.Hostname)
}

type LastStatusReportListOptions struct {
	Hostname     string
	RegionString string
}

func (options LastStatusReportListOptions) defaultPathRoot() string {
	k := "/calico/felix/v2/"
	if options.RegionString == "" {
		return k
	}
	k = k + options.RegionString + "/host"
	if options.Hostname == "" {
		return k
	}
	k = k + fmt.Sprintf("/%s/last_reported_status", options.Hostname)
	return k
}

func (options LastStatusReportListOptions) KeyFromDefaultPath(ekey string) Key {
	log.Debugf("Get StatusReport key from %s", ekey)
	r := matchLastStatusReport.FindAllStringSubmatch(ekey, -1)
	if len(r) != 1 {
		log.Debugf("Didn't match regex")
		return nil
	}
	regionString := r[0][1]
	name := r[0][2]
	if options.RegionString != "" && regionString != options.RegionString {
		log.Debugf("Didn't match region %s != %s", options.RegionString, regionString)
		return nil
	}
	if options.Hostname != "" && name != options.Hostname {
		log.Debugf("Didn't match name %s != %s", options.Hostname, name)
		return nil
	}
	return LastStatusReportKey{Hostname: name, RegionString: regionString}
}

type StatusReport struct {
	Timestamp     string  `json:"time"`
	UptimeSeconds float64 `json:"uptime"`
	FirstUpdate   bool    `json:"first_update"`
}
