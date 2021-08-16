// Copyright 2020 The Operator-SDK Authors
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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// ResourceCreatedAt creates new prometheus metrics for primary resource,
// with information {"name", "namespace", "group", "version", "kind"}
var ResourceCreatedAt = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "resource_created_at_seconds",
	Help: "Timestamp at which a resource was created",
}, []string{"name", "namespace", "group", "version", "kind"})

func init() {
	metrics.Registry.MustRegister(
		ResourceCreatedAt,
	)
}
