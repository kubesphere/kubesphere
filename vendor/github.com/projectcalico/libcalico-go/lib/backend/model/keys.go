// Copyright (c) 2016-2019 Tigera, Inc. All rights reserved.
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

package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	net2 "net"
	"reflect"
	"strings"
	"time"

	v3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/net"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
)

// RawString is used a value type to indicate that the value is a bare non-JSON string
type rawString string
type rawBool bool
type rawIP net.IP

var rawStringType = reflect.TypeOf(rawString(""))
var rawBoolType = reflect.TypeOf(rawBool(true))
var rawIPType = reflect.TypeOf(rawIP{})

// Key represents a parsed datastore key.
type Key interface {
	// defaultPath() returns a common path representation of the object used by
	// etcdv3 and other datastores.
	defaultPath() (string, error)

	// defaultDeletePath() returns a common path representation used by etcdv3
	// and other datastores to delete the object.
	defaultDeletePath() (string, error)

	// defaultDeleteParentPaths() returns an ordered slice of paths that should
	// be removed after deleting the primary path (given by defaultDeletePath),
	// provided there are no child entries associated with those paths.  This is
	// only used by directory based KV stores (such as etcdv3).  With a directory
	// based KV store, creation of a resource may also create parent directory entries
	// that could be shared by multiple resources, and therefore the parent directories
	// can only be removed when there are no more resources under them.  The list of
	// parent paths is ordered, and directories should be removed in the order supplied
	// in the slice and only if the directory is empty.
	defaultDeleteParentPaths() ([]string, error)

	// valueType returns the object type associated with this key.
	valueType() (reflect.Type, error)

	// String returns a unique string representation of this key.  The string
	// returned by this method must uniquely identify this Key.
	String() string
}

// Interface used to perform datastore lookups.
type ListInterface interface {
	// defaultPathRoot() returns a default stringified root path, i.e. path
	// to the directory containing all the keys to be listed.
	defaultPathRoot() string

	// BUG(smc) I think we should remove this and use the package KeyFromDefaultPath function.
	// KeyFromDefaultPath parses the default path representation of the
	// Key type for this list.  It returns nil if passed a different kind
	// of path.
	KeyFromDefaultPath(key string) Key
}

// KVPair holds a typed key and value object as well as datastore specific
// revision information.
//
// The Value is dependent on the Key, but in general will be on of the following
// types:
// -  A pointer to a struct
// -  A slice or map
// -  A bare string, boolean value or IP address (i.e. without quotes, so not
//    JSON format).
type KVPair struct {
	Key      Key
	Value    interface{}
	Revision string
	UID      *types.UID
	TTL      time.Duration // For writes, if non-zero, key has a TTL.
}

// KVPairList hosts a slice of KVPair structs and a Revision, returned from a Ls
type KVPairList struct {
	KVPairs  []*KVPair
	Revision string
}

// KeyToDefaultPath converts one of the Keys from this package into a unique
// '/'-delimited path, which is suitable for use as the key when storing the
// value in a hierarchical (i.e. one with directories and leaves) key/value
// datastore such as etcd v3.
//
// Each unique key returns a unique path.
//
// Keys with a hierarchical relationship share a common prefix.  However, in
// order to support datastores that do not support storing data at non-leaf
// nodes in the hierarchy (such as etcd v3), the path returned for a "parent"
// key, is not a direct ancestor of its children.
func KeyToDefaultPath(key Key) (string, error) {
	return key.defaultPath()
}

// KeyToDefaultDeletePath converts one of the Keys from this package into a
// unique '/'-delimited path, which is suitable for use as the key when
// (recursively) deleting the value from a hierarchical (i.e. one with
// directories and leaves) key/value datastore such as etcd v3.
//
// KeyToDefaultDeletePath returns a different path to KeyToDefaultPath when
// it is a passed a Key that represents a non-leaf which, for example, has its
// own metadata but also contains other resource types as children.
//
// KeyToDefaultDeletePath returns the common prefix of the non-leaf key and
// its children so that a recursive delete of that key would delete the
// object itself and any children it has.
func KeyToDefaultDeletePath(key Key) (string, error) {
	return key.defaultDeletePath()
}

func KeyToValueType(key Key) (reflect.Type, error) {
	return key.valueType()
}

// KeyToDefaultDeleteParentPaths returns a slice of '/'-delimited
// paths which are used to delete parent entries that may be auto-created
// by directory-based KV stores (e.g. etcd v3).  These paths should also be
// removed provided they have no more child entries.
//
// The list of parent paths is ordered, and directories should be removed
// in the order supplied in the slice and only if the directory is empty.
//
// For example,
// 	KeyToDefaultDeletePaths(WorkloadEndpointKey{
// 		Nodename: "h",
// 		OrchestratorID: "o",
// 		WorkloadID: "w",
// 		EndpointID: "e",
// 	})
// returns
//
// ["/calico/v1/host/h/workload/o/w/endpoint",
//  "/calico/v1/host/h/workload/o/w"]
//
// indicating that these paths should also be deleted when they are empty.
// In this example it is equivalent to deleting the workload when there are
// no more endpoints in the workload.
func KeyToDefaultDeleteParentPaths(key Key) ([]string, error) {
	return key.defaultDeleteParentPaths()
}

// ListOptionsToDefaultPathRoot converts list options struct into a
// common-prefix path suitable for querying a datastore that uses the paths
// returned by KeyToDefaultPath.
func ListOptionsToDefaultPathRoot(listOptions ListInterface) string {
	return listOptions.defaultPathRoot()
}

// ListOptionsIsFullyQualified returns true if the options actually specify a fully
// qualified resource rather than a partial match.
func ListOptionsIsFullyQualified(listOptions ListInterface) bool {
	// Construct the path prefix and then check to see if that actually corresponds to
	// the path of a resource instance.
	return listOptions.KeyFromDefaultPath(listOptions.defaultPathRoot()) != nil
}

// IsListOptionsLastSegmentPrefix returns true if the final segment of the default path
// root is a name prefix rather than the full name.
func IsListOptionsLastSegmentPrefix(listOptions ListInterface) bool {
	// Only supported for ResourceListOptions.
	rl, ok := listOptions.(ResourceListOptions)
	return ok && rl.IsLastSegmentIsPrefix()
}

// KeyFromDefaultPath parses the default path representation of a key into one
// of our <Type>Key structs.  Returns nil if the string doesn't match one of
// our key types.
func KeyFromDefaultPath(path string) Key {
	if m := matchWorkloadEndpoint.FindStringSubmatch(path); m != nil {
		log.Debugf("Path is a workload endpoint: %v", path)
		return WorkloadEndpointKey{
			Hostname:       m[1],
			OrchestratorID: unescapeName(m[2]),
			WorkloadID:     unescapeName(m[3]),
			EndpointID:     unescapeName(m[4]),
		}
	} else if m := matchHostEndpoint.FindStringSubmatch(path); m != nil {
		log.Debugf("Path is a host endpoint: %v", path)
		return HostEndpointKey{
			Hostname:   m[1],
			EndpointID: unescapeName(m[2]),
		}
	} else if m := matchNetworkSet.FindStringSubmatch(path); m != nil {
		log.Debugf("Path is a network set: %v", path)
		return NetworkSetKey{
			Name: unescapeName(m[1]),
		}
	} else if m := matchPolicy.FindStringSubmatch(path); m != nil {
		log.Debugf("Path is a policy: %v", path)
		return PolicyKey{
			Name: unescapeName(m[2]),
		}
	} else if m := matchProfile.FindStringSubmatch(path); m != nil {
		log.Debugf("Path is a profile: %v (%v)", path, m[2])
		pk := ProfileKey{unescapeName(m[1])}
		switch m[2] {
		case "tags":
			log.Debugf("Profile tags")
			return ProfileTagsKey{ProfileKey: pk}
		case "rules":
			log.Debugf("Profile rules")
			return ProfileRulesKey{ProfileKey: pk}
		case "labels":
			log.Debugf("Profile labels")
			return ProfileLabelsKey{ProfileKey: pk}
		}
		return nil
	} else if m := matchHostIp.FindStringSubmatch(path); m != nil {
		log.Debugf("Path is a host ID: %v", path)
		return HostIPKey{Hostname: m[1]}
	} else if m := matchIPPool.FindStringSubmatch(path); m != nil {
		log.Debugf("Path is a pool: %v", path)
		mungedCIDR := m[1]
		cidr := strings.Replace(mungedCIDR, "-", "/", 1)
		_, c, err := net.ParseCIDR(cidr)
		if err != nil {
			log.WithError(err).Warningf("Failed to parse CIDR %s", cidr)
		} else {
			return IPPoolKey{CIDR: *c}
		}
	} else if m := matchGlobalConfig.FindStringSubmatch(path); m != nil {
		log.Debugf("Path is a global felix config: %v", path)
		return GlobalConfigKey{Name: m[1]}
	} else if m := matchHostConfig.FindStringSubmatch(path); m != nil {
		log.Debugf("Path is a host config: %v", path)
		return HostConfigKey{Hostname: m[1], Name: m[2]}
	} else if matchReadyFlag.MatchString(path) {
		log.Debugf("Path is a ready flag: %v", path)
		return ReadyFlagKey{}
	} else if k := (NodeBGPConfigListOptions{}).KeyFromDefaultPath(path); k != nil {
		return k
	} else if k := (GlobalBGPConfigListOptions{}).KeyFromDefaultPath(path); k != nil {
		return k
	} else if k := (BlockAffinityListOptions{}).KeyFromDefaultPath(path); k != nil {
		return k
	} else if k := (BlockListOptions{}).KeyFromDefaultPath(path); k != nil {
		return k
	} else if k := (ResourceListOptions{Kind: v3.KindNode}).KeyFromDefaultPath(path); k != nil {
		return k
	} else if k := (ResourceListOptions{Kind: v3.KindBGPPeer}).KeyFromDefaultPath(path); k != nil {
		return k
	} else if k := (ResourceListOptions{Kind: v3.KindNetworkPolicy}).KeyFromDefaultPath(path); k != nil {
		return k
	} else if k := (ResourceListOptions{Kind: v3.KindIPPool}).KeyFromDefaultPath(path); k != nil {
		return k
	} else if k := (HostEndpointStatusListOptions{}).KeyFromDefaultPath(path); k != nil {
		return k
	} else if k := (WorkloadEndpointStatusListOptions{}).KeyFromDefaultPath(path); k != nil {
		return k
	} else if k := (ActiveStatusReportListOptions{}).KeyFromDefaultPath(path); k != nil {
		return k
	} else if k := (LastStatusReportListOptions{}).KeyFromDefaultPath(path); k != nil {
		return k
	} else {
		log.Debugf("Path is unknown: %v", path)
	}
	// Not a key we know about.
	return nil
}

// ParseValue parses the default JSON representation of our data into one of
// our value structs, according to the type of key.  I.e. if passed a
// PolicyKey as the first parameter, it will try to parse rawData into a
// Policy struct.
func ParseValue(key Key, rawData []byte) (interface{}, error) {
	valueType, err := key.valueType()
	if err != nil {
		return nil, err
	}
	if valueType == rawStringType {
		return string(rawData), nil
	}
	if valueType == rawBoolType {
		return string(rawData) == "true", nil
	}
	if valueType == rawIPType {
		ip := net2.ParseIP(string(rawData))
		if ip == nil {
			return nil, nil
		}
		return &net.IP{ip}, nil
	}
	value := reflect.New(valueType)
	elem := value.Elem()
	if elem.Kind() == reflect.Struct && elem.NumField() > 0 {
		if elem.Field(0).Type() == reflect.ValueOf(key).Type() {
			elem.Field(0).Set(reflect.ValueOf(key))
		}
	}
	iface := value.Interface()
	err = json.Unmarshal(rawData, iface)
	if err != nil {
		// This is a special case to address backwards compatibility from the time when we had no state information as block affinity value.
		// example:
		// Key: "/calico/ipam/v2/host/myhost.io/ipv4/block/172.29.82.0-26"
		// Value: ""
		// In 3.0.7 we added block affinity state as the value, so old "" value is no longer a valid JSON, so for that
		// particular case we replace the "" with a "{}" so it can be parsed and we don't leak blocks after upgrade to Calico 3.0.7
		// See: https://github.com/projectcalico/calico/issues/1956
		if bytes.Equal(rawData, []byte(``)) && valueType == typeBlockAff {
			rawData = []byte(`{}`)
			if err = json.Unmarshal(rawData, iface); err != nil {
				return nil, err
			}
		} else {
			log.Warningf("Failed to unmarshal %#v into value %#v",
				string(rawData), value)
			return nil, err
		}
	}

	if elem.Kind() != reflect.Struct {
		// Pointer to a map or slice, unwrap.
		iface = elem.Interface()
	}
	return iface, nil
}

// Serialize a value in the model to a []byte to stored in the datastore.  This
// performs the opposite processing to ParseValue()
func SerializeValue(d *KVPair) ([]byte, error) {
	valueType, err := d.Key.valueType()
	if err != nil {
		return nil, err
	}
	if d.Value == nil {
		return json.Marshal(nil)
	}
	if valueType == rawStringType {
		return []byte(d.Value.(string)), nil
	}
	if valueType == rawBoolType {
		return []byte(fmt.Sprint(d.Value)), nil
	}
	if valueType == rawIPType {
		return []byte(fmt.Sprint(d.Value)), nil
	}
	return json.Marshal(d.Value)
}
