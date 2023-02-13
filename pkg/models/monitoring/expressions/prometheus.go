/*
Copyright 2020 KubeSphere Authors

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

package expressions

import (
	"github.com/prometheus-community/prom-label-proxy/injectproxy"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

func init() {
	register("prometheus", labelReplace)
}

func labelReplace(input, ns string) (string, error) {
	root, err := parser.ParseExpr(input)
	if err != nil {
		return "", err
	}

	err = injectproxy.NewEnforcer(false, &labels.Matcher{
		Type:  labels.MatchEqual,
		Name:  "namespace",
		Value: ns,
	}).EnforceNode(root)
	if err != nil {
		return "", err
	}

	return root.String(), nil
}
