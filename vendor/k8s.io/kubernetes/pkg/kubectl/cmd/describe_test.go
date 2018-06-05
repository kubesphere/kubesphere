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
	"fmt"
	"net/http"
	"strings"
	"testing"

	"k8s.io/client-go/rest/fake"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	cmdtesting "k8s.io/kubernetes/pkg/kubectl/cmd/testing"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"
	"k8s.io/kubernetes/pkg/kubectl/scheme"
)

// Verifies that schemas that are not in the master tree of Kubernetes can be retrieved via Get.
func TestDescribeUnknownSchemaObject(t *testing.T) {
	d := &testDescriber{Output: "test output"}
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()
	_, _, codec := cmdtesting.NewExternalScheme()
	tf.DescriberVal = d
	tf.UnstructuredClient = &fake.RESTClient{
		NegotiatedSerializer: unstructuredSerializer,
		Resp:                 &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(codec, cmdtesting.NewInternalType("", "", "foo"))},
	}

	streams, _, buf, _ := genericclioptions.NewTestIOStreams()

	tf.Namespace = "non-default"
	cmd := NewCmdDescribe("kubectl", tf, streams)
	cmd.Run(cmd, []string{"type", "foo"})

	if d.Name != "foo" || d.Namespace != "" {
		t.Errorf("unexpected describer: %#v", d)
	}

	if buf.String() != fmt.Sprintf("%s", d.Output) {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

// Verifies that schemas that are not in the master tree of Kubernetes can be retrieved via Get.
func TestDescribeUnknownNamespacedSchemaObject(t *testing.T) {
	d := &testDescriber{Output: "test output"}
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()
	_, _, codec := cmdtesting.NewExternalScheme()

	tf.DescriberVal = d
	tf.UnstructuredClient = &fake.RESTClient{
		NegotiatedSerializer: unstructuredSerializer,
		Resp:                 &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(codec, cmdtesting.NewInternalNamespacedType("", "", "foo", "non-default"))},
	}
	tf.Namespace = "non-default"

	streams, _, buf, _ := genericclioptions.NewTestIOStreams()

	cmd := NewCmdDescribe("kubectl", tf, streams)
	cmd.Run(cmd, []string{"namespacedtype", "foo"})

	if d.Name != "foo" || d.Namespace != "non-default" {
		t.Errorf("unexpected describer: %#v", d)
	}

	if buf.String() != fmt.Sprintf("%s", d.Output) {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestDescribeObject(t *testing.T) {
	_, _, rc := testData()
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()
	codec := legacyscheme.Codecs.LegacyCodec(scheme.Versions...)

	d := &testDescriber{Output: "test output"}
	tf.DescriberVal = d
	tf.UnstructuredClient = &fake.RESTClient{
		NegotiatedSerializer: unstructuredSerializer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch p, m := req.URL.Path, req.Method; {
			case p == "/namespaces/test/replicationcontrollers/redis-master" && m == "GET":
				return &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(codec, &rc.Items[0])}, nil
			default:
				t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
				return nil, nil
			}
		}),
	}
	tf.Namespace = "test"

	streams, _, buf, _ := genericclioptions.NewTestIOStreams()

	cmd := NewCmdDescribe("kubectl", tf, streams)
	cmd.Flags().Set("filename", "../../../test/e2e/testing-manifests/guestbook/legacy/redis-master-controller.yaml")
	cmd.Run(cmd, []string{})

	if d.Name != "redis-master" || d.Namespace != "test" {
		t.Errorf("unexpected describer: %#v", d)
	}

	if buf.String() != fmt.Sprintf("%s", d.Output) {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestDescribeListObjects(t *testing.T) {
	pods, _, _ := testData()
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()
	codec := legacyscheme.Codecs.LegacyCodec(scheme.Versions...)

	d := &testDescriber{Output: "test output"}
	tf.DescriberVal = d
	tf.UnstructuredClient = &fake.RESTClient{
		NegotiatedSerializer: unstructuredSerializer,
		Resp:                 &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(codec, pods)},
	}

	streams, _, buf, _ := genericclioptions.NewTestIOStreams()

	tf.Namespace = "test"
	cmd := NewCmdDescribe("kubectl", tf, streams)
	cmd.Run(cmd, []string{"pods"})
	if buf.String() != fmt.Sprintf("%s\n\n%s", d.Output, d.Output) {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestDescribeObjectShowEvents(t *testing.T) {
	pods, _, _ := testData()
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()
	codec := legacyscheme.Codecs.LegacyCodec(scheme.Versions...)

	d := &testDescriber{Output: "test output"}
	tf.DescriberVal = d
	tf.UnstructuredClient = &fake.RESTClient{
		NegotiatedSerializer: unstructuredSerializer,
		Resp:                 &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(codec, pods)},
	}

	tf.Namespace = "test"
	cmd := NewCmdDescribe("kubectl", tf, genericclioptions.NewTestIOStreamsDiscard())
	cmd.Flags().Set("show-events", "true")
	cmd.Run(cmd, []string{"pods"})
	if d.Settings.ShowEvents != true {
		t.Errorf("ShowEvents = true expected, got ShowEvents = %v", d.Settings.ShowEvents)
	}
}

func TestDescribeObjectSkipEvents(t *testing.T) {
	pods, _, _ := testData()
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()
	codec := legacyscheme.Codecs.LegacyCodec(scheme.Versions...)

	d := &testDescriber{Output: "test output"}
	tf.DescriberVal = d
	tf.UnstructuredClient = &fake.RESTClient{
		NegotiatedSerializer: unstructuredSerializer,
		Resp:                 &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(codec, pods)},
	}

	tf.Namespace = "test"
	cmd := NewCmdDescribe("kubectl", tf, genericclioptions.NewTestIOStreamsDiscard())
	cmd.Flags().Set("show-events", "false")
	cmd.Run(cmd, []string{"pods"})
	if d.Settings.ShowEvents != false {
		t.Errorf("ShowEvents = false expected, got ShowEvents = %v", d.Settings.ShowEvents)
	}
}

func TestDescribeHelpMessage(t *testing.T) {
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()

	streams, _, buf, _ := genericclioptions.NewTestIOStreams()

	cmd := NewCmdDescribe("kubectl", tf, streams)
	cmd.SetArgs([]string{"-h"})
	cmd.SetOutput(buf)
	_, err := cmd.ExecuteC()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	got := buf.String()

	expected := `describe (-f FILENAME | TYPE [NAME_PREFIX | -l label] | TYPE/NAME)`
	if !strings.Contains(got, expected) {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", expected, got)
	}

	unexpected := `describe (-f FILENAME | TYPE [NAME_PREFIX | -l label] | TYPE/NAME) [flags]`
	if strings.Contains(got, unexpected) {
		t.Errorf("Expected not to contain: \n %v\nGot:\n %v\n", unexpected, got)
	}
}
