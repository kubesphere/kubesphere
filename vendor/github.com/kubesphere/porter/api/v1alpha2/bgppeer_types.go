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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Message struct {
	Notification   string `json:"notification,omitempty"`
	Update         string `json:"update,omitempty"`
	Open           string `json:"open,omitempty"`
	Keepalive      string `json:"keepalive,omitempty"`
	Refresh        string `json:"refresh,omitempty"`
	Discarded      string `json:"discarded,omitempty"`
	Total          string `json:"total,omitempty"`
	WithdrawUpdate string `json:"withdrawUpdate,omitempty"`
	WithdrawPrefix string `json:"withdrawPrefix,omitempty"`
}

type Messages struct {
	Received *Message `json:"received,omitempty"`
	Sent     *Message `json:"sent,omitempty"`
}

type Queues struct {
	Input  uint32 `json:"input,omitempty"`
	Output uint32 `json:"output,omitempty"`
}

type PeerState struct {
	AuthPassword     string    `json:"authPassword,omitempty"`
	Description      string    `json:"description,omitempty"`
	LocalAs          uint32    `json:"localAs,omitempty"`
	Messages         *Messages `json:"messages,omitempty"`
	NeighborAddress  string    `json:"neighborAddress,omitempty"`
	PeerAs           uint32    `json:"peerAs,omitempty"`
	PeerGroup        string    `json:"peerGroup,omitempty"`
	PeerType         uint32    `json:"peerType,omitempty"`
	Queues           *Queues   `json:"queues,omitempty"`
	RemovePrivateAs  uint32    `json:"removePrivateAs,omitempty"`
	RouteFlapDamping bool      `json:"routeFlapDamping,omitempty"`
	SendCommunity    uint32    `json:"sendCommunity,omitempty"`
	SessionState     string    `json:"sessionState,omitempty"`
	AdminState       string    `json:"adminState,omitempty"`
	OutQ             uint32    `json:"outQ,omitempty"`
	Flops            uint32    `json:"flops,omitempty"`
	RouterId         string    `json:"routerId,omitempty"`
}

type TimersState struct {
	ConnectRetry                 string `json:"connectRetry,omitempty"`
	HoldTime                     string `json:"holdTime,omitempty"`
	KeepaliveInterval            string `json:"keepaliveInterval,omitempty"`
	MinimumAdvertisementInterval string `json:"minimumAdvertisementInterval,omitempty"`
	NegotiatedHoldTime           string `json:"negotiatedHoldTime,omitempty"`
	Uptime                       string `json:"uptime,omitempty"`
	Downtime                     string `json:"downtime,omitempty"`
}

type NodePeerStatus struct {
	PeerState   PeerState   `json:"peerState,omitempty"`
	TimersState TimersState `json:"timersState,omitempty"`
}

// BgpPeerStatus defines the observed state of BgpPeer
type BgpPeerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	NodesPeerStatus map[string]NodePeerStatus `json:"nodesPeerStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope=Cluster

// BgpPeer is the Schema for the bgppeers API
type BgpPeer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BgpPeerSpec   `json:"spec,omitempty"`
	Status BgpPeerStatus `json:"status,omitempty"`
}

type Timers struct {
	Config *TimersConfig `json:"config,omitempty"`
}

// https://stackoverflow.com/questions/21151765/cannot-unmarshal-string-into-go-value-of-type-int64
type TimersConfig struct {
	ConnectRetry                 string `json:"connectRetry,omitempty"`
	HoldTime                     string `json:"holdTime,omitempty"`
	KeepaliveInterval            string `json:"keepaliveInterval,omitempty"`
	MinimumAdvertisementInterval string `json:"minimumAdvertisementInterval,omitempty"`
}

type PeerConf struct {
	AuthPassword      string `json:"authPassword,omitempty"`
	Description       string `json:"description,omitempty"`
	LocalAs           uint32 `json:"localAs,omitempty"`
	NeighborAddress   string `json:"neighborAddress,omitempty"`
	PeerAs            uint32 `json:"peerAs,omitempty"`
	PeerGroup         string `json:"peerGroup,omitempty"`
	PeerType          uint32 `json:"peerType,omitempty"`
	RemovePrivateAs   string `json:"removePrivateAs,omitempty"`
	RouteFlapDamping  bool   `json:"routeFlapDamping,omitempty"`
	SendCommunity     uint32 `json:"sendCommunity,omitempty"`
	NeighborInterface string `json:"neighborInterface,omitempty"`
	Vrf               string `json:"vrf,omitempty"`
	AllowOwnAs        uint32 `json:"allowOwnAs,omitempty"`
	ReplacePeerAs     bool   `json:"replacePeerAs,omitempty"`
	AdminDown         bool   `json:"adminDown,omitempty"`
}

type Transport struct {
	MtuDiscovery  bool   `json:"mtuDiscovery,omitempty"`
	PassiveMode   bool   `json:"passiveMode,omitempty"`
	RemoteAddress string `json:"remoteAddress,omitempty"`
	RemotePort    uint32 `json:"remotePort,omitempty"`
	TcpMss        uint32 `json:"tcpMss,omitempty"`
}

type MpGracefulRestartConfig struct {
	Enabled bool `json:"enabled,omitempty"`
}

type MpGracefulRestart struct {
	Config *MpGracefulRestartConfig `json:"config,omitempty"`
}

type Family struct {
	Afi  string `json:"afi,omitempty"`
	Safi string `json:"safi,omitempty"`
}

type AfiSafiConfig struct {
	Family  *Family `json:"family,omitempty"`
	Enabled bool    `json:"enabled,omitempty"`
}

type AddPathsConfig struct {
	Receive bool   `json:"receive,omitempty"`
	SendMax uint32 `json:"sendMax,omitempty"`
}

type AddPaths struct {
	Config *AddPathsConfig `json:"config,omitempty"`
}

type AfiSafi struct {
	MpGracefulRestart *MpGracefulRestart `json:"mpGracefulRestart,omitempty"`
	Config            *AfiSafiConfig     `json:"config,omitempty"`
	AddPaths          *AddPaths          `json:"addPaths,omitempty"`
}

func (c BgpPeerSpec) ConverToGoBgpPeer() (*api.Peer, error) {
	c.NodeSelector = nil

	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	var result api.Peer
	m := jsonpb.Unmarshaler{}
	return &result, m.Unmarshal(bytes.NewReader(jsonBytes), &result)
}

func ConverStatusFromGoBgpPeer(peer *api.Peer) (NodePeerStatus, error) {
	var (
		nodePeerStatus NodePeerStatus
	)

	m := jsonpb.Marshaler{}
	jsonStr, err := m.MarshalToString(peer.State)
	if err != nil {
		return nodePeerStatus, err
	}

	err = json.Unmarshal([]byte(jsonStr), &nodePeerStatus.PeerState)
	if err != nil {
		return nodePeerStatus, err
	}

	jsonStr, err = m.MarshalToString(peer.Timers.State)
	if err != nil {
		return nodePeerStatus, err
	}

	err = json.Unmarshal([]byte(jsonStr), &nodePeerStatus.TimersState)

	return nodePeerStatus, err
}

type EbgpMultihop struct {
	Enabled     bool   `json:"enabled,omitempty"`
	MultihopTtl uint32 `json:"multihopTtl,omitempty"`
}

type BgpPeerSpec struct {
	Conf            *PeerConf        `json:"conf,omitempty"`
	EbgpMultihop    *EbgpMultihop    `json:"ebgpMultihop,omitempty"`
	Timers          *Timers          `json:"timers,omitempty"`
	Transport       *Transport       `json:"transport,omitempty"`
	GracefulRestart *GracefulRestart `json:"gracefulRestart,omitempty"`
	AfiSafis        []*AfiSafi       `json:"afiSafis,omitempty"`

	NodeSelector *metav1.LabelSelector `json:"nodeSelector,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BgpPeerList contains a list of BgpPeer
type BgpPeerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BgpPeer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BgpPeer{}, &BgpPeerList{})
}
