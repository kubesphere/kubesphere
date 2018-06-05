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
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	stdstrings "strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	restclient "k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/api/testapi"
	apitesting "k8s.io/kubernetes/pkg/api/testing"
	api "k8s.io/kubernetes/pkg/apis/core"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/scheme"
	"k8s.io/kubernetes/pkg/printers"
)

// This init should be removed after switching this command and its tests to user external types.
func init() {
	api.AddToScheme(scheme.Scheme)
}

func initTestErrorHandler(t *testing.T) {
	cmdutil.BehaviorOnFatal(func(str string, code int) {
		t.Errorf("Error running command (exit code %d): %s", code, str)
	})
}

func defaultHeader() http.Header {
	header := http.Header{}
	header.Set("Content-Type", runtime.ContentTypeJSON)
	return header
}

func defaultClientConfig() *restclient.Config {
	return &restclient.Config{
		APIPath: "/api",
		ContentConfig: restclient.ContentConfig{
			NegotiatedSerializer: scheme.Codecs,
			ContentType:          runtime.ContentTypeJSON,
			GroupVersion:         &schema.GroupVersion{Version: "v1"},
		},
	}
}

func testData() (*api.PodList, *api.ServiceList, *api.ReplicationControllerList) {
	pods := &api.PodList{
		ListMeta: metav1.ListMeta{
			ResourceVersion: "15",
		},
		Items: []api.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "test", ResourceVersion: "10"},
				Spec:       apitesting.DeepEqualSafePodSpec(),
			},
			{
				ObjectMeta: metav1.ObjectMeta{Name: "bar", Namespace: "test", ResourceVersion: "11"},
				Spec:       apitesting.DeepEqualSafePodSpec(),
			},
		},
	}
	svc := &api.ServiceList{
		ListMeta: metav1.ListMeta{
			ResourceVersion: "16",
		},
		Items: []api.Service{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "baz", Namespace: "test", ResourceVersion: "12"},
				Spec: api.ServiceSpec{
					SessionAffinity: "None",
					Type:            api.ServiceTypeClusterIP,
				},
			},
		},
	}
	rc := &api.ReplicationControllerList{
		ListMeta: metav1.ListMeta{
			ResourceVersion: "17",
		},
		Items: []api.ReplicationController{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "rc1", Namespace: "test", ResourceVersion: "18"},
				Spec: api.ReplicationControllerSpec{
					Replicas: 1,
				},
			},
		},
	}
	return pods, svc, rc
}

type testDescriber struct {
	Name, Namespace string
	Settings        printers.DescriberSettings
	Output          string
	Err             error
}

func (t *testDescriber) Describe(namespace, name string, describerSettings printers.DescriberSettings) (output string, err error) {
	t.Namespace, t.Name = namespace, name
	t.Settings = describerSettings
	return t.Output, t.Err
}

func objBody(codec runtime.Codec, obj runtime.Object) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader([]byte(runtime.EncodeOrDie(codec, obj))))
}

func policyObjBody(obj runtime.Object) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader([]byte(runtime.EncodeOrDie(testapi.Policy.Codec(), obj))))
}

func bytesBody(bodyBytes []byte) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader(bodyBytes))
}

func stringBody(body string) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader([]byte(body)))
}

func newAllPhasePodList() *api.PodList {
	nodeName := "kubernetes-node-abcd"
	return &api.PodList{
		Items: []api.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test1",
					CreationTimestamp: metav1.Time{Time: time.Now().AddDate(-10, 0, 0)},
				},
				Spec: api.PodSpec{
					Containers: make([]api.Container, 2),
					NodeName:   nodeName,
				},
				Status: api.PodStatus{
					Phase: api.PodPending,
					ContainerStatuses: []api.ContainerStatus{
						{Ready: true, RestartCount: 3, State: api.ContainerState{Running: &api.ContainerStateRunning{}}},
						{RestartCount: 3},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test2",
					CreationTimestamp: metav1.Time{Time: time.Now().AddDate(-10, 0, 0)},
				},
				Spec: api.PodSpec{
					Containers: make([]api.Container, 2),
					NodeName:   nodeName,
				},
				Status: api.PodStatus{
					Phase: api.PodRunning,
					ContainerStatuses: []api.ContainerStatus{
						{Ready: true, RestartCount: 3, State: api.ContainerState{Running: &api.ContainerStateRunning{}}},
						{RestartCount: 3},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test3",
					CreationTimestamp: metav1.Time{Time: time.Now().AddDate(-10, 0, 0)},
				},
				Spec: api.PodSpec{
					Containers: make([]api.Container, 2),
					NodeName:   nodeName,
				},
				Status: api.PodStatus{
					Phase: api.PodSucceeded,
					ContainerStatuses: []api.ContainerStatus{
						{Ready: true, RestartCount: 3, State: api.ContainerState{Running: &api.ContainerStateRunning{}}},
						{RestartCount: 3},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test4",
					CreationTimestamp: metav1.Time{Time: time.Now().AddDate(-10, 0, 0)},
				},
				Spec: api.PodSpec{
					Containers: make([]api.Container, 2),
					NodeName:   nodeName,
				},
				Status: api.PodStatus{
					Phase: api.PodFailed,
					ContainerStatuses: []api.ContainerStatus{
						{Ready: true, RestartCount: 3, State: api.ContainerState{Running: &api.ContainerStateRunning{}}},
						{RestartCount: 3},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test5",
					CreationTimestamp: metav1.Time{Time: time.Now().AddDate(-10, 0, 0)},
				},
				Spec: api.PodSpec{
					Containers: make([]api.Container, 2),
					NodeName:   nodeName,
				},
				Status: api.PodStatus{
					Phase: api.PodUnknown,
					ContainerStatuses: []api.ContainerStatus{
						{Ready: true, RestartCount: 3, State: api.ContainerState{Running: &api.ContainerStateRunning{}}},
						{RestartCount: 3},
					},
				},
			}},
	}
}

func TestNormalizationFuncGlobalExistence(t *testing.T) {
	// This test can be safely deleted when we will not support multiple flag formats
	root := NewKubectlCommand(cmdutil.NewFactory(nil), os.Stdin, os.Stdout, os.Stderr)

	if root.Parent() != nil {
		t.Fatal("We expect the root command to be returned")
	}
	if root.GlobalNormalizationFunc() == nil {
		t.Fatal("We expect that root command has a global normalization function")
	}

	if reflect.ValueOf(root.GlobalNormalizationFunc()).Pointer() != reflect.ValueOf(root.Flags().GetNormalizeFunc()).Pointer() {
		t.Fatal("root command seems to have a wrong normalization function")
	}

	sub := root
	for sub.HasSubCommands() {
		sub = sub.Commands()[0]
	}

	// In case of failure of this test check this PR: spf13/cobra#110
	if reflect.ValueOf(sub.Flags().GetNormalizeFunc()).Pointer() != reflect.ValueOf(root.Flags().GetNormalizeFunc()).Pointer() {
		t.Fatal("child and root commands should have the same normalization functions")
	}
}

func genResponseWithJsonEncodedBody(bodyStruct interface{}) (*http.Response, error) {
	jsonBytes, err := json.Marshal(bodyStruct)
	if err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: 200, Header: defaultHeader(), Body: bytesBody(jsonBytes)}, nil
}

func Test_deprecatedAlias(t *testing.T) {
	var correctCommandCalled bool
	makeCobraCommand := func() *cobra.Command {
		cobraCmd := new(cobra.Command)
		cobraCmd.Use = "print five lines"
		cobraCmd.Run = func(*cobra.Command, []string) {
			correctCommandCalled = true
		}
		return cobraCmd
	}

	original := makeCobraCommand()
	alias := deprecatedAlias("echo", makeCobraCommand())

	if len(alias.Deprecated) == 0 {
		t.Error("deprecatedAlias should always have a non-empty .Deprecated")
	}
	if !stdstrings.Contains(alias.Deprecated, "print") {
		t.Error("deprecatedAlias should give the name of the new function in its .Deprecated field")
	}
	if !alias.Hidden {
		t.Error("deprecatedAlias should never have .Hidden == false (deprecated aliases should be hidden)")
	}

	if alias.Name() != "echo" {
		t.Errorf("deprecatedAlias has name %q, expected %q",
			alias.Name(), "echo")
	}
	if original.Name() != "print" {
		t.Errorf("original command has name %q, expected %q",
			original.Name(), "print")
	}

	buffer := new(bytes.Buffer)
	alias.SetOutput(buffer)
	alias.Execute()
	str := buffer.String()
	if !stdstrings.Contains(str, "deprecated") || !stdstrings.Contains(str, "print") {
		t.Errorf("deprecation warning %q does not include enough information", str)
	}

	// It would be nice to test to see that original.Run == alias.Run
	// Unfortunately Golang does not allow comparing functions. I could do
	// this with reflect, but that's technically invoking undefined
	// behavior. Best we can do is make sure that the function is called.
	if !correctCommandCalled {
		t.Errorf("original function doesn't appear to have been called by alias")
	}
}
