// Copyright (c) 2016 Tigera, Inc. All rights reserved.

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

package backend

import (
	"fmt"

	"github.com/projectcalico/libcalico-go/lib/apiconfig"
	bapi "github.com/projectcalico/libcalico-go/lib/backend/api"
	"github.com/projectcalico/libcalico-go/lib/backend/etcdv3"
	"github.com/projectcalico/libcalico-go/lib/backend/k8s"
	log "github.com/sirupsen/logrus"
)

// NewClient creates a new backend datastore client.
func NewClient(config apiconfig.CalicoAPIConfig) (c bapi.Client, err error) {
	log.Debugf("Using datastore type '%s'", config.Spec.DatastoreType)
	switch config.Spec.DatastoreType {
	case apiconfig.EtcdV3:
		c, err = etcdv3.NewEtcdV3Client(&config.Spec.EtcdConfig)
	case apiconfig.Kubernetes:
		c, err = k8s.NewKubeClient(&config.Spec)
	default:
		err = fmt.Errorf("unknown datastore type: %v",
			config.Spec.DatastoreType)
	}
	return
}
