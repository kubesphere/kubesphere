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

package model

import (
	"fmt"
	"regexp"

	"reflect"

	"sort"

	log "github.com/sirupsen/logrus"

	"github.com/projectcalico/calico/libcalico-go/lib/errors"
)

var (
	matchProfile = regexp.MustCompile("^/?calico/v1/policy/profile/([^/]+)/(rules|labels)$")
	typeProfile  = reflect.TypeOf(Profile{})
)

// The profile key actually returns the common parent of the three separate entries.
// It is useful to define this to re-use some of the common machinery, and can be used
// for delete processing since delete needs to remove the common parent.
type ProfileKey struct {
	Name string `json:"-" validate:"required,name"`
}

func (key ProfileKey) defaultPath() (string, error) {
	if key.Name == "" {
		return "", errors.ErrorInsufficientIdentifiers{Name: "name"}
	}
	e := fmt.Sprintf("/calico/v1/policy/profile/%s", escapeName(key.Name))
	return e, nil
}

func (key ProfileKey) defaultDeletePath() (string, error) {
	return key.defaultPath()
}

func (key ProfileKey) defaultDeleteParentPaths() ([]string, error) {
	return nil, nil
}

func (key ProfileKey) valueType() (reflect.Type, error) {
	return typeProfile, nil
}

func (key ProfileKey) String() string {
	return fmt.Sprintf("Profile(name=%s)", key.Name)
}

// ProfileRulesKey implements the KeyInterface for the profile rules
type ProfileRulesKey struct {
	ProfileKey
}

func (key ProfileRulesKey) defaultPath() (string, error) {
	e, err := key.ProfileKey.defaultPath()
	return e + "/rules", err
}

func (key ProfileRulesKey) valueType() (reflect.Type, error) {
	return reflect.TypeOf(ProfileRules{}), nil
}

func (key ProfileRulesKey) String() string {
	return fmt.Sprintf("ProfileRules(name=%s)", key.Name)
}

// ProfileLabelsKey implements the KeyInterface for the profile labels
type ProfileLabelsKey struct {
	ProfileKey
}

func (key ProfileLabelsKey) defaultPath() (string, error) {
	e, err := key.ProfileKey.defaultPath()
	return e + "/labels", err
}

func (key ProfileLabelsKey) valueType() (reflect.Type, error) {
	return reflect.TypeOf(map[string]string{}), nil
}

func (key ProfileLabelsKey) String() string {
	return fmt.Sprintf("ProfileLabels(name=%s)", key.Name)
}

type ProfileListOptions struct {
	Name string
}

func (options ProfileListOptions) defaultPathRoot() string {
	k := "/calico/v1/policy/profile"
	if options.Name == "" {
		return k
	}
	k = k + fmt.Sprintf("/%s", escapeName(options.Name))
	return k
}

func (options ProfileListOptions) KeyFromDefaultPath(path string) Key {
	log.Debugf("Get Profile key from %s", path)
	r := matchProfile.FindAllStringSubmatch(path, -1)
	if len(r) != 1 {
		log.Debugf("Didn't match regex")
		return nil
	}
	name := unescapeName(r[0][1])
	kind := r[0][2]
	if options.Name != "" && name != options.Name {
		log.Debugf("Didn't match name %s != %s", options.Name, name)
		return nil
	}
	pk := ProfileKey{Name: name}
	switch kind {
	case "labels":
		return ProfileLabelsKey{ProfileKey: pk}
	case "rules":
		return ProfileRulesKey{ProfileKey: pk}
	}
	return pk
}

// The profile structure is defined to allow the client to define a conversion interface
// to map between the API and backend profiles.  However, in the actual underlying
// implementation the profile is written as three separate entries - rules, tags and labels.
type Profile struct {
	Rules  ProfileRules
	Tags   []string
	Labels map[string]string
}

type ProfileRules struct {
	InboundRules  []Rule `json:"inbound_rules,omitempty" validate:"omitempty,dive"`
	OutboundRules []Rule `json:"outbound_rules,omitempty" validate:"omitempty,dive"`
}

func (_ *ProfileListOptions) ListConvert(ds []*KVPair) []*KVPair {

	profiles := make(map[string]*KVPair)
	var name string
	for _, d := range ds {
		switch t := d.Key.(type) {
		case ProfileLabelsKey:
			name = t.Name
		case ProfileRulesKey:
			name = t.Name
		default:
			panic(fmt.Errorf("Unexpected key type: %v", t))
		}

		// Get the KVPair for the profile, initialising if just created.
		pd, ok := profiles[name]
		if !ok {
			log.Debugf("Initialise profile %v", name)
			pd = &KVPair{
				Value: &Profile{},
				Key:   ProfileKey{Name: name},
			}
			profiles[name] = pd
		}

		p := pd.Value.(*Profile)
		switch t := d.Value.(type) {
		case map[string]string: // must be labels
			log.Debugf("Store labels %v", t)
			p.Labels = t
		case *ProfileRules: // must be rules
			log.Debugf("Store rules %v", t)
			p.Rules = *t
		default:
			panic(fmt.Errorf("Unexpected type: %v", t))
		}
		pd.Value = p
	}

	log.Debugf("Map of profiles: %v", profiles)

	// To store the keys in slice in sorted order
	var keys []string
	for k := range profiles {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]*KVPair, len(keys))
	for i, k := range keys {
		out[i] = profiles[k]
	}

	log.Debugf("Sorted groups of profiles: %v", out)

	return out
}
