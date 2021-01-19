/*
Copyright 2019 The Kubesphere Authors.

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

package v1alpha2

import (
	"bytes"
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
	api "github.com/osrg/gobgp/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeConfStatus struct {
	RouterId string `json:"routerId,omitempty"`
	As       uint32 `json:"as,omitempty"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BgpConfStatus defines the observed state of BgpConf
type BgpConfStatus struct {
	NodesConfStatus map[string]NodeConfStatus `json:"nodesConfStatus,omitempty"`
}

// +kubebuilder:rbac:groups=network.kubesphere.io,resources=bgpconfs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=network.kubesphere.io,resources=bgpconfs/status,verbs=get;update;patch

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope=Cluster

// BgpConf is the Schema for the bgpconfs API
type BgpConf struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BgpConfSpec   `json:"spec,omitempty"`
	Status BgpConfStatus `json:"status,omitempty"`
}

func (c BgpConfSpec) ConverToGoBgpGlabalConf() (*api.Global, error) {
	c.AsPerRack = nil

	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	var result api.Global
	m := jsonpb.Unmarshaler{}
	return &result, m.Unmarshal(bytes.NewReader(jsonBytes), &result)
}

// Configuration parameters relating to the global BGP router.
type BgpConfSpec struct {
	As               uint32            `json:"as,omitempty"`
	AsPerRack        map[string]uint32 `json:"asPerRack,omitempty"`
	RouterId         string            `json:"routerId,omitempty"`
	ListenPort       int32             `json:"listenPort,omitempty"`
	ListenAddresses  []string          `json:"listenAddresses,omitempty"`
	Families         []uint32          `json:"families,omitempty"`
	UseMultiplePaths bool              `json:"useMultiplePaths,omitempty"`
	GracefulRestart  *GracefulRestart  `json:"gracefulRestart,omitempty"`
}

type GracefulRestart struct {
	Enabled             bool   `json:"enabled,omitempty"`
	RestartTime         uint32 `json:"restartTime,omitempty"`
	HelperOnly          bool   `json:"helperOnly,omitempty"`
	DeferralTime        uint32 `json:"deferralTime,omitempty"`
	NotificationEnabled bool   `json:"notificationEnabled,omitempty"`
	LonglivedEnabled    bool   `json:"longlivedEnabled,omitempty"`
	StaleRoutesTime     uint32 `json:"staleRoutesTime,omitempty"`
	PeerRestartTime     uint32 `json:"peerRestartTime,omitempty"`
	PeerRestarting      bool   `json:"peerRestarting,omitempty"`
	LocalRestarting     bool   `json:"localRestarting,omitempty"`
	Mode                string `json:"mode,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BgpConfList contains a list of BgpConf
type BgpConfList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BgpConf `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BgpConf{}, &BgpConfList{})
}
