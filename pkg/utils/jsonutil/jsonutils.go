/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package jsonutil

import (
	"encoding/json"
	"strings"

	"k8s.io/klog/v2"
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
