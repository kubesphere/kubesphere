/*
Copyright 2017 The Kubernetes Authors.

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

package checkpoint

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"

	apiv1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeclient "k8s.io/client-go/kubernetes/fake"
	utiltest "k8s.io/kubernetes/pkg/kubelet/kubeletconfig/util/test"
)

func TestNewRemoteConfigSource(t *testing.T) {
	cases := []struct {
		desc   string
		source *apiv1.NodeConfigSource
		expect RemoteConfigSource
		err    string
	}{
		// all NodeConfigSource subfields nil
		{"all NodeConfigSource subfields nil",
			&apiv1.NodeConfigSource{}, nil, "exactly one subfield must be non-nil"},
		{"ConfigMapRef: empty name, namespace, and UID",
			&apiv1.NodeConfigSource{ConfigMapRef: &apiv1.ObjectReference{}}, nil, "invalid ObjectReference"},
		// ConfigMapRef: empty name and namespace
		{"ConfigMapRef: empty name and namespace",
			&apiv1.NodeConfigSource{ConfigMapRef: &apiv1.ObjectReference{UID: "uid"}}, nil, "invalid ObjectReference"},
		// ConfigMapRef: empty name and UID
		{"ConfigMapRef: empty name and UID",
			&apiv1.NodeConfigSource{ConfigMapRef: &apiv1.ObjectReference{Namespace: "namespace"}}, nil, "invalid ObjectReference"},
		// ConfigMapRef: empty namespace and UID
		{"ConfigMapRef: empty namespace and UID",
			&apiv1.NodeConfigSource{ConfigMapRef: &apiv1.ObjectReference{Name: "name"}}, nil, "invalid ObjectReference"},
		// ConfigMapRef: empty UID
		{"ConfigMapRef: empty namespace and UID",
			&apiv1.NodeConfigSource{ConfigMapRef: &apiv1.ObjectReference{Name: "name", Namespace: "namespace"}}, nil, "invalid ObjectReference"},
		// ConfigMapRef: empty namespace
		{"ConfigMapRef: empty namespace and UID",
			&apiv1.NodeConfigSource{ConfigMapRef: &apiv1.ObjectReference{Name: "name", UID: "uid"}}, nil, "invalid ObjectReference"},
		// ConfigMapRef: empty name
		{"ConfigMapRef: empty namespace and UID",
			&apiv1.NodeConfigSource{ConfigMapRef: &apiv1.ObjectReference{Namespace: "namespace", UID: "uid"}}, nil, "invalid ObjectReference"},
		// ConfigMapRef: valid reference
		{"ConfigMapRef: valid reference",
			&apiv1.NodeConfigSource{ConfigMapRef: &apiv1.ObjectReference{Name: "name", Namespace: "namespace", UID: "uid"}},
			&remoteConfigMap{&apiv1.NodeConfigSource{ConfigMapRef: &apiv1.ObjectReference{Name: "name", Namespace: "namespace", UID: "uid"}}}, ""},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			source, _, err := NewRemoteConfigSource(c.source)
			utiltest.ExpectError(t, err, c.err)
			if err != nil {
				return
			}
			// underlying object should match the object passed in
			if !apiequality.Semantic.DeepEqual(c.expect.object(), source.object()) {
				t.Errorf("case %q, expect RemoteConfigSource %s but got %s", c.desc, spew.Sdump(c.expect), spew.Sdump(source))
			}
		})
	}
}

func TestRemoteConfigMapUID(t *testing.T) {
	const expect = "uid"
	source, _, err := NewRemoteConfigSource(&apiv1.NodeConfigSource{ConfigMapRef: &apiv1.ObjectReference{Name: "name", Namespace: "namespace", UID: expect}})
	if err != nil {
		t.Fatalf("error constructing remote config source: %v", err)
	}
	uid := source.UID()
	if expect != uid {
		t.Errorf("expect %q, but got %q", expect, uid)
	}
}

func TestRemoteConfigMapAPIPath(t *testing.T) {
	const namespace = "namespace"
	const name = "name"
	source, _, err := NewRemoteConfigSource(&apiv1.NodeConfigSource{ConfigMapRef: &apiv1.ObjectReference{Name: name, Namespace: namespace, UID: "uid"}})
	if err != nil {
		t.Fatalf("error constructing remote config source: %v", err)
	}
	expect := fmt.Sprintf(configMapAPIPathFmt, namespace, name)
	path := source.APIPath()
	if expect != path {
		t.Errorf("expect %q, but got %q", expect, path)
	}
}

func TestRemoteConfigMapDownload(t *testing.T) {
	cm := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "name",
			Namespace: "namespace",
			UID:       "uid",
		}}
	client := fakeclient.NewSimpleClientset(cm)
	payload, err := NewConfigMapPayload(cm)
	if err != nil {
		t.Fatalf("error constructing payload: %v", err)
	}

	makeSource := func(source *apiv1.NodeConfigSource) RemoteConfigSource {
		s, _, err := NewRemoteConfigSource(source)
		if err != nil {
			t.Fatalf("error constructing remote config source %v", err)
		}
		return s
	}

	cases := []struct {
		desc   string
		source RemoteConfigSource
		expect Payload
		err    string
	}{
		// object doesn't exist
		{"object doesn't exist",
			makeSource(&apiv1.NodeConfigSource{ConfigMapRef: &apiv1.ObjectReference{Name: "bogus", Namespace: "namespace", UID: "bogus"}}),
			nil, "not found"},
		// UID of downloaded object doesn't match UID of referent found via namespace/name
		{"UID is incorrect for namespace/name",
			makeSource(&apiv1.NodeConfigSource{ConfigMapRef: &apiv1.ObjectReference{Name: "name", Namespace: "namespace", UID: "bogus"}}),
			nil, "does not match"},
		// successful download
		{"object exists and reference is correct",
			makeSource(&apiv1.NodeConfigSource{ConfigMapRef: &apiv1.ObjectReference{Name: "name", Namespace: "namespace", UID: "uid"}}),
			payload, ""},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			payload, _, err := c.source.Download(client)
			utiltest.ExpectError(t, err, c.err)
			if err != nil {
				return
			}
			// downloaded object should match the expected
			if !apiequality.Semantic.DeepEqual(c.expect.object(), payload.object()) {
				t.Errorf("case %q, expect Checkpoint %s but got %s", c.desc, spew.Sdump(c.expect), spew.Sdump(payload))
			}
		})
	}
}

func TestEqualRemoteConfigSources(t *testing.T) {
	cases := []struct {
		desc   string
		a      RemoteConfigSource
		b      RemoteConfigSource
		expect bool
	}{
		{"both nil", nil, nil, true},
		{"a nil", nil, &remoteConfigMap{}, false},
		{"b nil", &remoteConfigMap{}, nil, false},
		{"neither nil, equal", &remoteConfigMap{}, &remoteConfigMap{}, true},
		{"neither nil, not equal",
			&remoteConfigMap{&apiv1.NodeConfigSource{ConfigMapRef: &apiv1.ObjectReference{Name: "a"}}},
			&remoteConfigMap{},
			false},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			if EqualRemoteConfigSources(c.a, c.b) != c.expect {
				t.Errorf("expected EqualRemoteConfigSources to return %t, but got %t", c.expect, !c.expect)
			}
		})
	}
}
