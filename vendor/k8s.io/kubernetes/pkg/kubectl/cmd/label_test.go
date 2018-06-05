/*
Copyright 2014 The Kubernetes Authors.

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

package cmd

import (
	"bytes"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest/fake"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	cmdtesting "k8s.io/kubernetes/pkg/kubectl/cmd/testing"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"
	"k8s.io/kubernetes/pkg/kubectl/scheme"
)

func TestValidateLabels(t *testing.T) {
	tests := []struct {
		meta      *metav1.ObjectMeta
		labels    map[string]string
		expectErr bool
		test      string
	}{
		{
			meta: &metav1.ObjectMeta{
				Labels: map[string]string{
					"a": "b",
					"c": "d",
				},
			},
			labels: map[string]string{
				"a": "c",
				"d": "b",
			},
			test:      "one shared",
			expectErr: true,
		},
		{
			meta: &metav1.ObjectMeta{
				Labels: map[string]string{
					"a": "b",
					"c": "d",
				},
			},
			labels: map[string]string{
				"b": "d",
				"c": "a",
			},
			test:      "second shared",
			expectErr: true,
		},
		{
			meta: &metav1.ObjectMeta{
				Labels: map[string]string{
					"a": "b",
					"c": "d",
				},
			},
			labels: map[string]string{
				"b": "a",
				"d": "c",
			},
			test: "no overlap",
		},
		{
			meta: &metav1.ObjectMeta{},
			labels: map[string]string{
				"b": "a",
				"d": "c",
			},
			test: "no labels",
		},
	}
	for _, test := range tests {
		err := validateNoOverwrites(test.meta, test.labels)
		if test.expectErr && err == nil {
			t.Errorf("%s: unexpected non-error", test.test)
		}
		if !test.expectErr && err != nil {
			t.Errorf("%s: unexpected error: %v", test.test, err)
		}
	}
}

func TestParseLabels(t *testing.T) {
	tests := []struct {
		labels         []string
		expected       map[string]string
		expectedRemove []string
		expectErr      bool
	}{
		{
			labels:   []string{"a=b", "c=d"},
			expected: map[string]string{"a": "b", "c": "d"},
		},
		{
			labels:   []string{},
			expected: map[string]string{},
		},
		{
			labels:         []string{"a=b", "c=d", "e-"},
			expected:       map[string]string{"a": "b", "c": "d"},
			expectedRemove: []string{"e"},
		},
		{
			labels:    []string{"ab", "c=d"},
			expectErr: true,
		},
		{
			labels:    []string{"a=b", "c=d", "a-"},
			expectErr: true,
		},
		{
			labels:   []string{"a="},
			expected: map[string]string{"a": ""},
		},
		{
			labels:    []string{"a=%^$"},
			expectErr: true,
		},
	}
	for _, test := range tests {
		labels, remove, err := parseLabels(test.labels)
		if test.expectErr && err == nil {
			t.Errorf("unexpected non-error: %v", test)
		}
		if !test.expectErr && err != nil {
			t.Errorf("unexpected error: %v %v", err, test)
		}
		if !reflect.DeepEqual(labels, test.expected) {
			t.Errorf("expected: %v, got %v", test.expected, labels)
		}
		if !reflect.DeepEqual(remove, test.expectedRemove) {
			t.Errorf("expected: %v, got %v", test.expectedRemove, remove)
		}
	}
}

func TestLabelFunc(t *testing.T) {
	tests := []struct {
		obj       runtime.Object
		overwrite bool
		version   string
		labels    map[string]string
		remove    []string
		expected  runtime.Object
		expectErr bool
	}{
		{
			obj: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"a": "b"},
				},
			},
			labels:    map[string]string{"a": "b"},
			expectErr: true,
		},
		{
			obj: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"a": "b"},
				},
			},
			labels:    map[string]string{"a": "c"},
			overwrite: true,
			expected: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"a": "c"},
				},
			},
		},
		{
			obj: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"a": "b"},
				},
			},
			labels: map[string]string{"c": "d"},
			expected: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"a": "b", "c": "d"},
				},
			},
		},
		{
			obj: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"a": "b"},
				},
			},
			labels:  map[string]string{"c": "d"},
			version: "2",
			expected: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels:          map[string]string{"a": "b", "c": "d"},
					ResourceVersion: "2",
				},
			},
		},
		{
			obj: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"a": "b"},
				},
			},
			labels: map[string]string{},
			remove: []string{"a"},
			expected: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
			},
		},
		{
			obj: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"a": "b", "c": "d"},
				},
			},
			labels: map[string]string{"e": "f"},
			remove: []string{"a"},
			expected: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"c": "d",
						"e": "f",
					},
				},
			},
		},
		{
			obj: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{},
			},
			labels: map[string]string{"a": "b"},
			expected: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"a": "b"},
				},
			},
		},
	}
	for _, test := range tests {
		err := labelFunc(test.obj, test.overwrite, test.version, test.labels, test.remove)
		if test.expectErr {
			if err == nil {
				t.Errorf("unexpected non-error: %v", test)
			}
			continue
		}
		if !test.expectErr && err != nil {
			t.Errorf("unexpected error: %v %v", err, test)
		}
		if !reflect.DeepEqual(test.obj, test.expected) {
			t.Errorf("expected: %v, got %v", test.expected, test.obj)
		}
	}
}

func TestLabelErrors(t *testing.T) {
	testCases := map[string]struct {
		args  []string
		flags map[string]string
		errFn func(error) bool
	}{
		"no args": {
			args:  []string{},
			errFn: func(err error) bool { return strings.Contains(err.Error(), "one or more resources must be specified") },
		},
		"not enough labels": {
			args:  []string{"pods"},
			errFn: func(err error) bool { return strings.Contains(err.Error(), "at least one label update is required") },
		},
		"wrong labels": {
			args:  []string{"pods", "-"},
			errFn: func(err error) bool { return strings.Contains(err.Error(), "at least one label update is required") },
		},
		"wrong labels 2": {
			args:  []string{"pods", "=bar"},
			errFn: func(err error) bool { return strings.Contains(err.Error(), "at least one label update is required") },
		},
		"no resources": {
			args:  []string{"pods-"},
			errFn: func(err error) bool { return strings.Contains(err.Error(), "one or more resources must be specified") },
		},
		"no resources 2": {
			args:  []string{"pods=bar"},
			errFn: func(err error) bool { return strings.Contains(err.Error(), "one or more resources must be specified") },
		},
		"resources but no selectors": {
			args: []string{"pods", "app=bar"},
			errFn: func(err error) bool {
				return strings.Contains(err.Error(), "resource(s) were provided, but no name, label selector, or --all flag specified")
			},
		},
		"multiple resources but no selectors": {
			args: []string{"pods,deployments", "app=bar"},
			errFn: func(err error) bool {
				return strings.Contains(err.Error(), "resource(s) were provided, but no name, label selector, or --all flag specified")
			},
		},
	}

	for k, testCase := range testCases {
		t.Run(k, func(t *testing.T) {
			tf := cmdtesting.NewTestFactory()
			defer tf.Cleanup()

			tf.Namespace = "test"
			tf.ClientConfigVal = defaultClientConfig()

			ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()
			buf := bytes.NewBuffer([]byte{})
			cmd := NewCmdLabel(tf, ioStreams)
			cmd.SetOutput(buf)

			for k, v := range testCase.flags {
				cmd.Flags().Set(k, v)
			}
			opts := NewLabelOptions(ioStreams)
			err := opts.Complete(tf, cmd, testCase.args)
			if err == nil {
				err = opts.Validate()
			}
			if err == nil {
				err = opts.RunLabel()
			}
			if !testCase.errFn(err) {
				t.Errorf("%s: unexpected error: %v", k, err)
				return
			}
			if buf.Len() > 0 {
				t.Errorf("buffer should be empty: %s", string(buf.Bytes()))
			}
		})
	}
}

func TestLabelForResourceFromFile(t *testing.T) {
	pods, _, _ := testData()
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()

	codec := legacyscheme.Codecs.LegacyCodec(scheme.Versions...)

	tf.UnstructuredClient = &fake.RESTClient{
		NegotiatedSerializer: unstructuredSerializer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch req.Method {
			case "GET":
				switch req.URL.Path {
				case "/namespaces/test/replicationcontrollers/cassandra":
					return &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(codec, &pods.Items[0])}, nil
				default:
					t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
					return nil, nil
				}
			case "PATCH":
				switch req.URL.Path {
				case "/namespaces/test/replicationcontrollers/cassandra":
					return &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(codec, &pods.Items[0])}, nil
				default:
					t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
					return nil, nil
				}
			default:
				t.Fatalf("unexpected request: %s %#v\n%#v", req.Method, req.URL, req)
				return nil, nil
			}
		}),
	}
	tf.Namespace = "test"
	tf.ClientConfigVal = defaultClientConfig()

	ioStreams, _, buf, _ := genericclioptions.NewTestIOStreams()
	cmd := NewCmdLabel(tf, ioStreams)
	opts := NewLabelOptions(ioStreams)
	opts.Filenames = []string{"../../../test/e2e/testing-manifests/statefulset/cassandra/controller.yaml"}
	err := opts.Complete(tf, cmd, []string{"a=b"})
	if err == nil {
		err = opts.Validate()
	}
	if err == nil {
		err = opts.RunLabel()
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "labeled") {
		t.Errorf("did not set labels: %s", buf.String())
	}
}

func TestLabelLocal(t *testing.T) {
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()

	tf.UnstructuredClient = &fake.RESTClient{
		NegotiatedSerializer: unstructuredSerializer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			t.Fatalf("unexpected request: %s %#v\n%#v", req.Method, req.URL, req)
			return nil, nil
		}),
	}
	tf.Namespace = "test"
	tf.ClientConfigVal = defaultClientConfig()

	ioStreams, _, buf, _ := genericclioptions.NewTestIOStreams()
	cmd := NewCmdLabel(tf, ioStreams)
	opts := NewLabelOptions(ioStreams)
	opts.Filenames = []string{"../../../test/e2e/testing-manifests/statefulset/cassandra/controller.yaml"}
	opts.local = true
	err := opts.Complete(tf, cmd, []string{"a=b"})
	if err == nil {
		err = opts.Validate()
	}
	if err == nil {
		err = opts.RunLabel()
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "labeled") {
		t.Errorf("did not set labels: %s", buf.String())
	}
}

func TestLabelMultipleObjects(t *testing.T) {
	pods, _, _ := testData()
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()

	codec := legacyscheme.Codecs.LegacyCodec(scheme.Versions...)

	tf.UnstructuredClient = &fake.RESTClient{
		NegotiatedSerializer: unstructuredSerializer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch req.Method {
			case "GET":
				switch req.URL.Path {
				case "/namespaces/test/pods":
					return &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(codec, pods)}, nil
				default:
					t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
					return nil, nil
				}
			case "PATCH":
				switch req.URL.Path {
				case "/namespaces/test/pods/foo":
					return &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(codec, &pods.Items[0])}, nil
				case "/namespaces/test/pods/bar":
					return &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(codec, &pods.Items[1])}, nil
				default:
					t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
					return nil, nil
				}
			default:
				t.Fatalf("unexpected request: %s %#v\n%#v", req.Method, req.URL, req)
				return nil, nil
			}
		}),
	}
	tf.Namespace = "test"
	tf.ClientConfigVal = defaultClientConfig()

	ioStreams, _, buf, _ := genericclioptions.NewTestIOStreams()
	opts := NewLabelOptions(ioStreams)
	opts.all = true
	cmd := NewCmdLabel(tf, ioStreams)
	err := opts.Complete(tf, cmd, []string{"pods", "a=b"})
	if err == nil {
		err = opts.Validate()
	}
	if err == nil {
		err = opts.RunLabel()
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Count(buf.String(), "labeled") != len(pods.Items) {
		t.Errorf("not all labels are set: %s", buf.String())
	}
}
