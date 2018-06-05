/*
Copyright 2016 The Kubernetes Authors.

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

package set

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	"k8s.io/kubernetes/pkg/api/testapi"
	cmdtesting "k8s.io/kubernetes/pkg/kubectl/cmd/testing"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/kubectl/scheme"
	"k8s.io/kubernetes/pkg/printers"
)

func TestImageLocal(t *testing.T) {
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()

	ns := serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	tf.Client = &fake.RESTClient{
		GroupVersion:         schema.GroupVersion{Version: ""},
		NegotiatedSerializer: ns,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			t.Fatalf("unexpected request: %s %#v\n%#v", req.Method, req.URL, req)
			return nil, nil
		}),
	}
	tf.Namespace = "test"
	tf.ClientConfigVal = &restclient.Config{ContentConfig: restclient.ContentConfig{GroupVersion: &schema.GroupVersion{Version: ""}}}

	outputFormat := "name"

	streams, _, buf, _ := genericclioptions.NewTestIOStreams()
	cmd := NewCmdImage(tf, streams)
	cmd.SetOutput(buf)
	cmd.Flags().Set("output", outputFormat)
	cmd.Flags().Set("local", "true")

	opts := SetImageOptions{
		PrintFlags: &printers.PrintFlags{
			JSONYamlPrintFlags: printers.NewJSONYamlPrintFlags(),
			NamePrintFlags:     printers.NewNamePrintFlags(""),

			OutputFormat: &outputFormat,
		},
		FilenameOptions: resource.FilenameOptions{
			Filenames: []string{"../../../../test/e2e/testing-manifests/statefulset/cassandra/controller.yaml"}},
		Local:     true,
		IOStreams: streams,
	}
	err := opts.Complete(tf, cmd, []string{"cassandra=thingy"})
	if err == nil {
		err = opts.Validate()
	}
	if err == nil {
		err = opts.Run()
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "replicationcontroller/cassandra") {
		t.Errorf("did not set image: %s", buf.String())
	}
}

func TestSetImageValidation(t *testing.T) {
	printFlags := &printers.PrintFlags{
		JSONYamlPrintFlags: printers.NewJSONYamlPrintFlags(),
		NamePrintFlags:     printers.NewNamePrintFlags(""),
	}

	testCases := []struct {
		name         string
		imageOptions *SetImageOptions
		expectErr    string
	}{
		{
			name:         "test resource < 1 and filenames empty",
			imageOptions: &SetImageOptions{PrintFlags: printFlags},
			expectErr:    "[one or more resources must be specified as <resource> <name> or <resource>/<name>, at least one image update is required]",
		},
		{
			name: "test containerImages < 1",
			imageOptions: &SetImageOptions{
				PrintFlags: printFlags,
				Resources:  []string{"a", "b", "c"},

				FilenameOptions: resource.FilenameOptions{
					Filenames: []string{"testFile"},
				},
			},
			expectErr: "at least one image update is required",
		},
		{
			name: "test containerImages > 1 and all containers are already specified by *",
			imageOptions: &SetImageOptions{
				PrintFlags: printFlags,
				Resources:  []string{"a", "b", "c"},
				FilenameOptions: resource.FilenameOptions{
					Filenames: []string{"testFile"},
				},
				ContainerImages: map[string]string{
					"test": "test",
					"*":    "test",
				},
			},
			expectErr: "all containers are already specified by *, but saw more than one container_name=container_image pairs",
		},
		{
			name: "success case",
			imageOptions: &SetImageOptions{
				PrintFlags: printFlags,
				Resources:  []string{"a", "b", "c"},
				FilenameOptions: resource.FilenameOptions{
					Filenames: []string{"testFile"},
				},
				ContainerImages: map[string]string{
					"test": "test",
				},
			},
			expectErr: "",
		},
	}
	for _, testCase := range testCases {
		err := testCase.imageOptions.Validate()
		if err != nil {
			if err.Error() != testCase.expectErr {
				t.Errorf("[%s]:expect err:%s got err:%s", testCase.name, testCase.expectErr, err.Error())
			}
		}
		if err == nil && (testCase.expectErr != "") {
			t.Errorf("[%s]:expect err:%s got err:%v", testCase.name, testCase.expectErr, err)
		}
	}
}

func TestSetMultiResourcesImageLocal(t *testing.T) {
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()

	ns := serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	tf.Client = &fake.RESTClient{
		GroupVersion:         schema.GroupVersion{Version: ""},
		NegotiatedSerializer: ns,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			t.Fatalf("unexpected request: %s %#v\n%#v", req.Method, req.URL, req)
			return nil, nil
		}),
	}
	tf.Namespace = "test"
	tf.ClientConfigVal = &restclient.Config{ContentConfig: restclient.ContentConfig{GroupVersion: &schema.GroupVersion{Version: ""}}}

	outputFormat := "name"

	streams, _, buf, _ := genericclioptions.NewTestIOStreams()
	cmd := NewCmdImage(tf, streams)
	cmd.SetOutput(buf)
	cmd.Flags().Set("output", outputFormat)
	cmd.Flags().Set("local", "true")

	opts := SetImageOptions{
		PrintFlags: &printers.PrintFlags{
			JSONYamlPrintFlags: printers.NewJSONYamlPrintFlags(),
			NamePrintFlags:     printers.NewNamePrintFlags(""),

			OutputFormat: &outputFormat,
		},
		FilenameOptions: resource.FilenameOptions{
			Filenames: []string{"../../../../test/fixtures/pkg/kubectl/cmd/set/multi-resource-yaml.yaml"}},
		Local:     true,
		IOStreams: streams,
	}
	err := opts.Complete(tf, cmd, []string{"*=thingy"})
	if err == nil {
		err = opts.Validate()
	}
	if err == nil {
		err = opts.Run()
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedOut := "replicationcontroller/first-rc\nreplicationcontroller/second-rc\n"
	if buf.String() != expectedOut {
		t.Errorf("expected out:\n%s\nbut got:\n%s", expectedOut, buf.String())
	}
}

func TestSetImageRemote(t *testing.T) {
	inputs := []struct {
		name                            string
		object                          runtime.Object
		apiPrefix, apiGroup, apiVersion string
		testAPIGroup                    string
		args                            []string
	}{
		{
			name: "set image extensionsv1beta1 ReplicaSet",
			object: &extensionsv1beta1.ReplicaSet{
				ObjectMeta: metav1.ObjectMeta{Name: "nginx"},
				Spec: extensionsv1beta1.ReplicaSetSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "nginx",
									Image: "nginx",
								},
							},
						},
					},
				},
			},
			testAPIGroup: "extensions",
			apiPrefix:    "/apis", apiGroup: "extensions", apiVersion: "v1beta1",
			args: []string{"replicaset", "nginx", "*=thingy"},
		},
		{
			name: "set image appsv1beta2 ReplicaSet",
			object: &appsv1beta2.ReplicaSet{
				ObjectMeta: metav1.ObjectMeta{Name: "nginx"},
				Spec: appsv1beta2.ReplicaSetSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "nginx",
									Image: "nginx",
								},
							},
						},
					},
				},
			},
			testAPIGroup: "extensions",
			apiPrefix:    "/apis", apiGroup: "apps", apiVersion: "v1beta2",
			args: []string{"replicaset", "nginx", "*=thingy"},
		},
		{
			name: "set image appsv1 ReplicaSet",
			object: &appsv1.ReplicaSet{
				ObjectMeta: metav1.ObjectMeta{Name: "nginx"},
				Spec: appsv1.ReplicaSetSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "nginx",
									Image: "nginx",
								},
							},
						},
					},
				},
			},
			testAPIGroup: "extensions",
			apiPrefix:    "/apis", apiGroup: "apps", apiVersion: "v1",
			args: []string{"replicaset", "nginx", "*=thingy"},
		},
		{
			name: "set image extensionsv1beta1 DaemonSet",
			object: &extensionsv1beta1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{Name: "nginx"},
				Spec: extensionsv1beta1.DaemonSetSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "nginx",
									Image: "nginx",
								},
							},
						},
					},
				},
			},
			testAPIGroup: "extensions",
			apiPrefix:    "/apis", apiGroup: "extensions", apiVersion: "v1beta1",
			args: []string{"daemonset", "nginx", "*=thingy"},
		},
		{
			name: "set image appsv1beta2 DaemonSet",
			object: &appsv1beta2.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{Name: "nginx"},
				Spec: appsv1beta2.DaemonSetSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "nginx",
									Image: "nginx",
								},
							},
						},
					},
				},
			},
			testAPIGroup: "extensions",
			apiPrefix:    "/apis", apiGroup: "apps", apiVersion: "v1beta2",
			args: []string{"daemonset", "nginx", "*=thingy"},
		},
		{
			name: "set image appsv1 DaemonSet",
			object: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{Name: "nginx"},
				Spec: appsv1.DaemonSetSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "nginx",
									Image: "nginx",
								},
							},
						},
					},
				},
			},
			testAPIGroup: "extensions",
			apiPrefix:    "/apis", apiGroup: "apps", apiVersion: "v1",
			args: []string{"daemonset", "nginx", "*=thingy"},
		},
		{
			name: "set image extensionsv1beta1 Deployment",
			object: &extensionsv1beta1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "nginx"},
				Spec: extensionsv1beta1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "nginx",
									Image: "nginx",
								},
							},
						},
					},
				},
			},
			testAPIGroup: "extensions",
			apiPrefix:    "/apis", apiGroup: "extensions", apiVersion: "v1beta1",
			args: []string{"deployment", "nginx", "*=thingy"},
		},
		{
			name: "set image appsv1beta1 Deployment",
			object: &appsv1beta1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "nginx"},
				Spec: appsv1beta1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "nginx",
									Image: "nginx",
								},
							},
						},
					},
				},
			},
			testAPIGroup: "extensions",
			apiPrefix:    "/apis", apiGroup: "apps", apiVersion: "v1beta1",
			args: []string{"deployment", "nginx", "*=thingy"},
		},
		{
			name: "set image appsv1beta2 Deployment",
			object: &appsv1beta2.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "nginx"},
				Spec: appsv1beta2.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "nginx",
									Image: "nginx",
								},
							},
						},
					},
				},
			},
			testAPIGroup: "extensions",
			apiPrefix:    "/apis", apiGroup: "apps", apiVersion: "v1beta2",
			args: []string{"deployment", "nginx", "*=thingy"},
		},
		{
			name: "set image appsv1 Deployment",
			object: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "nginx"},
				Spec: appsv1.DeploymentSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "nginx",
									Image: "nginx",
								},
							},
						},
					},
				},
			},
			testAPIGroup: "extensions",
			apiPrefix:    "/apis", apiGroup: "apps", apiVersion: "v1",
			args: []string{"deployment", "nginx", "*=thingy"},
		},
		{
			name: "set image appsv1beta1 StatefulSet",
			object: &appsv1beta1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Name: "nginx"},
				Spec: appsv1beta1.StatefulSetSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "nginx",
									Image: "nginx",
								},
							},
						},
					},
				},
			},
			testAPIGroup: "apps",
			apiPrefix:    "/apis", apiGroup: "apps", apiVersion: "v1beta1",
			args: []string{"statefulset", "nginx", "*=thingy"},
		},
		{
			name: "set image appsv1beta2 StatefulSet",
			object: &appsv1beta2.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Name: "nginx"},
				Spec: appsv1beta2.StatefulSetSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "nginx",
									Image: "nginx",
								},
							},
						},
					},
				},
			},
			testAPIGroup: "apps",
			apiPrefix:    "/apis", apiGroup: "apps", apiVersion: "v1beta2",
			args: []string{"statefulset", "nginx", "*=thingy"},
		},
		{
			name: "set image appsv1 StatefulSet",
			object: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Name: "nginx"},
				Spec: appsv1.StatefulSetSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "nginx",
									Image: "nginx",
								},
							},
						},
					},
				},
			},
			testAPIGroup: "apps",
			apiPrefix:    "/apis", apiGroup: "apps", apiVersion: "v1",
			args: []string{"statefulset", "nginx", "*=thingy"},
		},
		{
			name: "set image batchv1 Job",
			object: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{Name: "nginx"},
				Spec: batchv1.JobSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "nginx",
									Image: "nginx",
								},
							},
						},
					},
				},
			},
			testAPIGroup: "batch",
			apiPrefix:    "/apis", apiGroup: "batch", apiVersion: "v1",
			args: []string{"job", "nginx", "*=thingy"},
		},
		{
			name: "set image v1.ReplicationController",
			object: &v1.ReplicationController{
				ObjectMeta: metav1.ObjectMeta{Name: "nginx"},
				Spec: v1.ReplicationControllerSpec{
					Template: &v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "nginx",
									Image: "nginx",
								},
							},
						},
					},
				},
			},
			testAPIGroup: "",
			apiPrefix:    "/api", apiGroup: "", apiVersion: "v1",
			args: []string{"replicationcontroller", "nginx", "*=thingy"},
		},
	}
	for _, input := range inputs {
		t.Run(input.name, func(t *testing.T) {
			groupVersion := schema.GroupVersion{Group: input.apiGroup, Version: input.apiVersion}
			testapi.Default = testapi.Groups[input.testAPIGroup]
			tf := cmdtesting.NewTestFactory()
			defer tf.Cleanup()

			codec := scheme.Codecs.CodecForVersions(scheme.Codecs.LegacyCodec(groupVersion), scheme.Codecs.UniversalDecoder(groupVersion), groupVersion, groupVersion)
			ns := serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}
			tf.Namespace = "test"
			tf.Client = &fake.RESTClient{
				GroupVersion:         groupVersion,
				NegotiatedSerializer: ns,
				Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
					resourcePath := testapi.Default.ResourcePath(input.args[0]+"s", tf.Namespace, input.args[1])
					switch p, m := req.URL.Path, req.Method; {
					case p == resourcePath && m == http.MethodGet:
						return &http.Response{StatusCode: http.StatusOK, Header: defaultHeader(), Body: objBody(codec, input.object)}, nil
					case p == resourcePath && m == http.MethodPatch:
						stream, err := req.GetBody()
						if err != nil {
							return nil, err
						}
						bytes, err := ioutil.ReadAll(stream)
						if err != nil {
							return nil, err
						}
						assert.Contains(t, string(bytes), `"image":`+`"`+"thingy"+`"`, fmt.Sprintf("image not updated for %#v", input.object))
						return &http.Response{StatusCode: http.StatusOK, Header: defaultHeader(), Body: objBody(codec, input.object)}, nil
					default:
						t.Errorf("%s: unexpected request: %s %#v\n%#v", "image", req.Method, req.URL, req)
						return nil, fmt.Errorf("unexpected request")
					}
				}),
				VersionedAPIPath: path.Join(input.apiPrefix, testapi.Default.GroupVersion().String()),
			}

			outputFormat := "yaml"

			streams := genericclioptions.NewTestIOStreamsDiscard()
			cmd := NewCmdImage(tf, streams)
			cmd.Flags().Set("output", outputFormat)
			opts := SetImageOptions{
				PrintFlags: &printers.PrintFlags{
					JSONYamlPrintFlags: printers.NewJSONYamlPrintFlags(),
					NamePrintFlags:     printers.NewNamePrintFlags(""),

					OutputFormat: &outputFormat,
				},

				Local:     false,
				IOStreams: streams,
			}
			err := opts.Complete(tf, cmd, input.args)
			assert.NoError(t, err)
			err = opts.Run()
			assert.NoError(t, err)
		})
	}
}
