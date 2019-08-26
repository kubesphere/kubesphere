// Copyright (c) 2016-2017 Tigera, Inc. All rights reserved.

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

package apiconfig

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
)

type DatastoreType string

const (
	EtcdV3              DatastoreType = "etcdv3"
	Kubernetes          DatastoreType = "kubernetes"
	KindCalicoAPIConfig               = "CalicoAPIConfig"
)

// CalicoAPIConfig contains the connection information for a Calico CalicoAPIConfig resource
type CalicoAPIConfig struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the BGPConfiguration.
	Spec CalicoAPIConfigSpec `json:"spec,omitempty"`
}

// CalicoAPIConfigSpec contains the specification for a Calico CalicoAPIConfig resource.
type CalicoAPIConfigSpec struct {
	DatastoreType DatastoreType `json:"datastoreType" envconfig:"DATASTORE_TYPE" default:"etcdv3"`
	// Inline the ectd config fields
	EtcdConfig
	// Inline the k8s config fields.
	KubeConfig
}

type EtcdConfig struct {
	EtcdEndpoints    string `json:"etcdEndpoints" envconfig:"ETCD_ENDPOINTS"`
	EtcdDiscoverySrv string `json:"etcdDiscoverySrv" envconfig:"ETCD_DISCOVERY_SRV"`
	EtcdUsername     string `json:"etcdUsername" envconfig:"ETCD_USERNAME"`
	EtcdPassword     string `json:"etcdPassword" envconfig:"ETCD_PASSWORD"`
	EtcdKeyFile      string `json:"etcdKeyFile" envconfig:"ETCD_KEY_FILE"`
	EtcdCertFile     string `json:"etcdCertFile" envconfig:"ETCD_CERT_FILE"`
	EtcdCACertFile   string `json:"etcdCACertFile" envconfig:"ETCD_CA_CERT_FILE"`

	// These config file parameters are to support inline certificates, keys and CA / Trusted certificate.
	// There are no corresponding environment variables to avoid accidental exposure.
	EtcdKey    string `json:"etcdKey" ignored:"true"`
	EtcdCert   string `json:"etcdCert" ignored:"true"`
	EtcdCACert string `json:"etcdCACert" ignored:"true"`
}

type KubeConfig struct {
	Kubeconfig               string `json:"kubeconfig" envconfig:"KUBECONFIG" default:""`
	K8sAPIEndpoint           string `json:"k8sAPIEndpoint" envconfig:"K8S_API_ENDPOINT" default:""`
	K8sKeyFile               string `json:"k8sKeyFile" envconfig:"K8S_KEY_FILE" default:""`
	K8sCertFile              string `json:"k8sCertFile" envconfig:"K8S_CERT_FILE" default:""`
	K8sCAFile                string `json:"k8sCAFile" envconfig:"K8S_CA_FILE" default:""`
	K8sAPIToken              string `json:"k8sAPIToken" ignore:"true"`
	K8sInsecureSkipTLSVerify bool   `json:"k8sInsecureSkipTLSVerify" envconfig:"K8S_INSECURE_SKIP_TLS_VERIFY" default:""`
	K8sDisableNodePoll       bool   `json:"k8sDisableNodePoll" envconfig:"K8S_DISABLE_NODE_POLL" default:""`

	// K8sUsePodCIDR controls whether or not IPAM blocks are generated based on Node.Spec.PodCIDR. Set this
	// to true when using host-local IPAM, and set to false when using calico-ipam.
	K8sUsePodCIDR bool `json:"usePodCIDR" envconfig:"USE_POD_CIDR" default:""`
}

// NewCalicoAPIConfig creates a new (zeroed) CalicoAPIConfig struct with the
// TypeMetadata initialised to the current version.
func NewCalicoAPIConfig() *CalicoAPIConfig {
	return &CalicoAPIConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindCalicoAPIConfig,
			APIVersion: apiv3.GroupVersionCurrent,
		},
	}
}

// IsAlphaFeatureSet checks if the comma separated features have the
// name set in it.
func IsAlphaFeatureSet(features, name string) bool {

	fs := strings.Split(features, ",")
	for _, f := range fs {
		if f == name {
			return true
		}
	}

	return false
}
