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

package create

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	rbac "k8s.io/api/rbac/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	cmdtesting "k8s.io/kubernetes/pkg/kubectl/cmd/testing"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"
)

func TestCreateClusterRoleBinding(t *testing.T) {
	expectBinding := &rbac.ClusterRoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name: "fake-binding",
		},
		TypeMeta: v1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "ClusterRole",
			Name:     "fake-clusterrole",
		},
		Subjects: []rbac.Subject{
			{
				Kind:     rbac.UserKind,
				APIGroup: "rbac.authorization.k8s.io",
				Name:     "fake-user",
			},
			{
				Kind:     rbac.GroupKind,
				APIGroup: "rbac.authorization.k8s.io",
				Name:     "fake-group",
			},
			{
				Kind:      rbac.ServiceAccountKind,
				Namespace: "fake-namespace",
				Name:      "fake-account",
			},
		},
	}

	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()

	ns := legacyscheme.Codecs

	info, _ := runtime.SerializerInfoForMediaType(ns.SupportedMediaTypes(), runtime.ContentTypeJSON)
	encoder := ns.EncoderForVersion(info.Serializer, groupVersion)
	decoder := ns.DecoderToVersion(info.Serializer, groupVersion)

	tf.Namespace = "test"
	tf.Client = &ClusterRoleBindingRESTClient{
		RESTClient: &fake.RESTClient{
			NegotiatedSerializer: ns,
			Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
				switch p, m := req.URL.Path, req.Method; {
				case p == "/clusterrolebindings" && m == "POST":
					bodyBits, err := ioutil.ReadAll(req.Body)
					if err != nil {
						t.Fatalf("TestCreateClusterRoleBinding error: %v", err)
						return nil, nil
					}

					if obj, _, err := decoder.Decode(bodyBits, nil, &rbac.ClusterRoleBinding{}); err == nil {
						if !reflect.DeepEqual(obj.(*rbac.ClusterRoleBinding), expectBinding) {
							t.Fatalf("TestCreateClusterRoleBinding: expected:\n%#v\nsaw:\n%#v", expectBinding, obj.(*rbac.ClusterRoleBinding))
							return nil, nil
						}
					} else {
						t.Fatalf("TestCreateClusterRoleBinding error, could not decode the request body into rbac.ClusterRoleBinding object: %v", err)
						return nil, nil
					}

					responseBinding := &rbac.ClusterRoleBinding{}
					responseBinding.Name = "fake-binding"
					return &http.Response{StatusCode: 201, Header: defaultHeader(), Body: ioutil.NopCloser(bytes.NewReader([]byte(runtime.EncodeOrDie(encoder, responseBinding))))}, nil
				default:
					t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
					return nil, nil
				}
			}),
		},
	}

	expectedOutput := "clusterrolebinding.rbac.authorization.k8s.io/" + expectBinding.Name + "\n"
	ioStreams, _, buf, _ := genericclioptions.NewTestIOStreams()
	cmd := NewCmdCreateClusterRoleBinding(tf, ioStreams)
	cmd.Flags().Set("clusterrole", "fake-clusterrole")
	cmd.Flags().Set("user", "fake-user")
	cmd.Flags().Set("group", "fake-group")
	cmd.Flags().Set("output", "name")
	cmd.Flags().Set("serviceaccount", "fake-namespace:fake-account")
	cmd.Run(cmd, []string{"fake-binding"})
	if buf.String() != expectedOutput {
		t.Errorf("TestCreateClusterRoleBinding: expected %v\n but got %v\n", expectedOutput, buf.String())
	}
}

type ClusterRoleBindingRESTClient struct {
	*fake.RESTClient
}

func (c *ClusterRoleBindingRESTClient) Post() *restclient.Request {
	config := restclient.ContentConfig{
		ContentType:          runtime.ContentTypeJSON,
		NegotiatedSerializer: c.NegotiatedSerializer,
	}

	info, _ := runtime.SerializerInfoForMediaType(c.NegotiatedSerializer.SupportedMediaTypes(), runtime.ContentTypeJSON)
	serializers := restclient.Serializers{
		Encoder: c.NegotiatedSerializer.EncoderForVersion(info.Serializer, schema.GroupVersion{Group: "rbac.authorization.k8s.io", Version: "v1beta1"}),
		Decoder: c.NegotiatedSerializer.DecoderToVersion(info.Serializer, schema.GroupVersion{Group: "rbac.authorization.k8s.io", Version: "v1beta1"}),
	}
	if info.StreamSerializer != nil {
		serializers.StreamingSerializer = info.StreamSerializer.Serializer
		serializers.Framer = info.StreamSerializer.Framer
	}
	return restclient.NewRequest(c, "POST", &url.URL{Host: "localhost"}, c.VersionedAPIPath, config, serializers, nil, nil, 0)
}

func defaultHeader() http.Header {
	header := http.Header{}
	header.Set("Content-Type", runtime.ContentTypeJSON)
	return header
}
