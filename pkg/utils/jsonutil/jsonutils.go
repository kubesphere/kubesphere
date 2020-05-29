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

package jsonutil

import (
	"encoding/json"
	"k8s.io/klog"
	"strings"
)

type JsonRawMessage []byte

func (m JsonRawMessage) Find(key string) JsonRawMessage {
	var objmap map[string]json.RawMessage
	err := json.Unmarshal(m, &objmap)
	if err != nil {
		klog.Errorf("Resolve JSON Key failed, find key =%s, err=%s",
			key, err)
		return nil
	}
	return JsonRawMessage(objmap[key])
}

func (m JsonRawMessage) ToList() []JsonRawMessage {
	var lists []json.RawMessage
	err := json.Unmarshal(m, &lists)
	if err != nil {
		klog.Errorf("Resolve JSON List failed, err=%s",
			err)
		return nil
	}
	var res []JsonRawMessage
	for _, v := range lists {
		res = append(res, JsonRawMessage(v))
	}
	return res
}

func (m JsonRawMessage) ToString() string {
	res := strings.Replace(string(m[:]), "\"", "", -1)
	return res
}
