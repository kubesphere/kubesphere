// Copyright (c) 2018-2020 Tigera, Inc. All rights reserved.
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
	"fmt"

	api "github.com/projectcalico/api/pkg/apis/projectcalico/v3"

	"github.com/projectcalico/calico/libcalico-go/lib/backend/k8s/conversion"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

type serviceAccountConverter struct {
}

// NewServiceaccountConverter Constructor to convert ServiceAccount to Profile
func NewServiceAccountConverter() Converter {
	return &serviceAccountConverter{}
}

func (nc *serviceAccountConverter) Convert(k8sObj interface{}) (interface{}, error) {
	c := conversion.NewConverter()
	serviceAccount, ok := k8sObj.(*v1.ServiceAccount)
	if !ok {
		tombstone, ok := k8sObj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return nil, fmt.Errorf("couldn't get object from tombstone %+v", k8sObj)
		}
		serviceAccount, ok = tombstone.Obj.(*v1.ServiceAccount)
		if !ok {
			return nil, fmt.Errorf("tombstone contained object that is not a Serviceaccount %+v", k8sObj)
		}
	}
	kvp, err := c.ServiceAccountToProfile(serviceAccount)
	if err != nil {
		return nil, err
	}
	profile := kvp.Value.(*api.Profile)

	// Isolate the metadata fields that we care about. ResourceVersion, CreationTimeStamp, etc are
	// not relevant so we ignore them. This prevents unnecessary updates.
	profile.ObjectMeta = metav1.ObjectMeta{Name: profile.Name}

	return *profile, nil
}

// GetKey returns name of the Profile as its key.  For Profiles
// backed by Kubernetes serviceaccounts and managed by this controller, the name
// is of format `ksa.namespace.name`.
func (nc *serviceAccountConverter) GetKey(obj interface{}) string {
	profile := obj.(api.Profile)
	return profile.Name
}

func (nc *serviceAccountConverter) DeleteArgsFromKey(key string) (string, string) {
	// Not serviceaccount, so just return the key, which is the profile name.
	return "", key
}
