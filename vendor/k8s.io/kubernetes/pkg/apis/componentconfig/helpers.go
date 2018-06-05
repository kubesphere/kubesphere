/*
Copyright 2016 The Kubernetes Authors.

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

package componentconfig

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilnet "k8s.io/apimachinery/pkg/util/net"
)

// used for validating command line opts
// TODO(mikedanese): remove these when we remove command line flags

type IPVar struct {
	Val *string
}

func (v IPVar) Set(s string) error {
	if len(s) == 0 {
		v.Val = nil
		return nil
	}
	if net.ParseIP(s) == nil {
		return fmt.Errorf("%q is not a valid IP address", s)
	}
	if v.Val == nil {
		// it's okay to panic here since this is programmer error
		panic("the string pointer passed into IPVar should not be nil")
	}
	*v.Val = s
	return nil
}

func (v IPVar) String() string {
	if v.Val == nil {
		return ""
	}
	return *v.Val
}

func (v IPVar) Type() string {
	return "ip"
}

// IPPortVar allows IP or IP:port formats.
type IPPortVar struct {
	Val *string
}

func (v IPPortVar) Set(s string) error {
	if len(s) == 0 {
		v.Val = nil
		return nil
	}

	if v.Val == nil {
		// it's okay to panic here since this is programmer error
		panic("the string pointer passed into IPPortVar should not be nil")
	}

	// Both IP and IP:port are valid.
	// Attempt to parse into IP first.
	if net.ParseIP(s) != nil {
		*v.Val = s
		return nil
	}

	// Can not parse into IP, now assume IP:port.
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return fmt.Errorf("%q is not in a valid format (ip or ip:port): %v", s, err)
	}
	if net.ParseIP(host) == nil {
		return fmt.Errorf("%q is not a valid IP address", host)
	}
	if _, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf("%q is not a valid number", port)
	}
	*v.Val = s
	return nil
}

func (v IPPortVar) String() string {
	if v.Val == nil {
		return ""
	}
	return *v.Val
}

func (v IPPortVar) Type() string {
	return "ipport"
}

type PortRangeVar struct {
	Val *string
}

func (v PortRangeVar) Set(s string) error {
	if _, err := utilnet.ParsePortRange(s); err != nil {
		return fmt.Errorf("%q is not a valid port range: %v", s, err)
	}
	if v.Val == nil {
		// it's okay to panic here since this is programmer error
		panic("the string pointer passed into PortRangeVar should not be nil")
	}
	*v.Val = s
	return nil
}

func (v PortRangeVar) String() string {
	if v.Val == nil {
		return ""
	}
	return *v.Val
}

func (v PortRangeVar) Type() string {
	return "port-range"
}

// ConvertObjToConfigMap converts an object to a ConfigMap.
// This is specifically meant for ComponentConfigs.
func ConvertObjToConfigMap(name string, obj runtime.Object) (*v1.ConfigMap, error) {
	eJSONBytes, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: map[string]string{
			name: string(eJSONBytes[:]),
		},
	}
	return cm, nil
}
