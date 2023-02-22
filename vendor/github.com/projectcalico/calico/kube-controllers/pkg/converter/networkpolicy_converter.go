// Copyright (c) 2017-2021 Tigera, Inc. All rights reserved.
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

package converter

import (
	"errors"
	"fmt"
	"strings"

	api "github.com/projectcalico/api/pkg/apis/projectcalico/v3"

	"github.com/projectcalico/calico/libcalico-go/lib/backend/k8s/conversion"
	cerrors "github.com/projectcalico/calico/libcalico-go/lib/errors"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

type policyConverter struct {
}

// NewPolicyConverter Constructor for policyConverter
func NewPolicyConverter() Converter {
	return &policyConverter{}
}

// Convert takes a Kubernetes NetworkPolicy and returns a Calico api.NetworkPolicy representation.
func (p *policyConverter) Convert(k8sObj interface{}) (interface{}, error) {
	np, ok := k8sObj.(*networkingv1.NetworkPolicy)

	if !ok {
		tombstone, ok := k8sObj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return nil, fmt.Errorf("couldn't get object from tombstone %+v", k8sObj)
		}
		np, ok = tombstone.Obj.(*networkingv1.NetworkPolicy)
		if !ok {
			return nil, fmt.Errorf("tombstone contained object that is not a NetworkPolicy %+v", k8sObj)
		}
	}

	c := conversion.NewConverter()
	kvp, err := c.K8sNetworkPolicyToCalico(np)
	// Silently ignore rule conversion errors. We don't expect any conversion errors
	// since the data given to us here is validated by the Kubernetes API. The conversion
	// code ignores any rules that it cannot parse, and we will pass the valid ones to Felix.
	var e *cerrors.ErrorPolicyConversion
	if err != nil && !errors.As(err, &e) {
		return nil, err
	}
	cnp := kvp.Value.(*api.NetworkPolicy)

	// Isolate the metadata fields that we care about. ResourceVersion, CreationTimeStamp, etc are
	// not relevant so we ignore them. This prevents unnecessary updates.
	cnp.ObjectMeta = metav1.ObjectMeta{Name: cnp.Name, Namespace: cnp.Namespace}

	return *cnp, err
}

// GetKey returns the 'namespace/name' for the given Calico NetworkPolicy as its key.
func (p *policyConverter) GetKey(obj interface{}) string {
	policy := obj.(api.NetworkPolicy)
	return fmt.Sprintf("%s/%s", policy.Namespace, policy.Name)
}

func (p *policyConverter) DeleteArgsFromKey(key string) (string, string) {
	splits := strings.SplitN(key, "/", 2)
	return splits[0], splits[1]
}
