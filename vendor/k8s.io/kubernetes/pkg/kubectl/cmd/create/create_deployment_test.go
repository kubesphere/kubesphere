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

package create

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/kubectl"
	cmdtesting "k8s.io/kubernetes/pkg/kubectl/cmd/testing"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"
)

func Test_generatorFromName(t *testing.T) {
	const (
		nonsenseName         = "not-a-real-generator-name"
		basicName            = cmdutil.DeploymentBasicV1Beta1GeneratorName
		basicAppsV1Beta1Name = cmdutil.DeploymentBasicAppsV1Beta1GeneratorName
		basicAppsV1Name      = cmdutil.DeploymentBasicAppsV1GeneratorName
		deploymentName       = "deployment-name"
	)
	imageNames := []string{"image-1", "image-2"}

	generator, ok := generatorFromName(nonsenseName, imageNames, deploymentName)
	assert.Nil(t, generator)
	assert.False(t, ok)

	generator, ok = generatorFromName(basicName, imageNames, deploymentName)
	assert.True(t, ok)

	{
		expectedGenerator := &kubectl.DeploymentBasicGeneratorV1{
			BaseDeploymentGenerator: kubectl.BaseDeploymentGenerator{
				Name:   deploymentName,
				Images: imageNames,
			},
		}
		assert.Equal(t, expectedGenerator, generator)
	}

	generator, ok = generatorFromName(basicAppsV1Beta1Name, imageNames, deploymentName)
	assert.True(t, ok)

	{
		expectedGenerator := &kubectl.DeploymentBasicAppsGeneratorV1Beta1{
			BaseDeploymentGenerator: kubectl.BaseDeploymentGenerator{
				Name:   deploymentName,
				Images: imageNames,
			},
		}
		assert.Equal(t, expectedGenerator, generator)
	}

	generator, ok = generatorFromName(basicAppsV1Name, imageNames, deploymentName)
	assert.True(t, ok)

	{
		expectedGenerator := &kubectl.DeploymentBasicAppsGeneratorV1{
			BaseDeploymentGenerator: kubectl.BaseDeploymentGenerator{
				Name:   deploymentName,
				Images: imageNames,
			},
		}
		assert.Equal(t, expectedGenerator, generator)
	}
}

func TestCreateDeployment(t *testing.T) {
	depName := "jonny-dep"
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()

	ns := legacyscheme.Codecs
	fakeDiscovery := "{\"kind\":\"APIResourceList\",\"apiVersion\":\"v1\",\"groupVersion\":\"apps/v1\",\"resources\":[{\"name\":\"deployments\",\"singularName\":\"\",\"namespaced\":true,\"kind\":\"Deployment\",\"verbs\":[\"create\",\"delete\",\"deletecollection\",\"get\",\"list\",\"patch\",\"update\",\"watch\"],\"shortNames\":[\"deploy\"],\"categories\":[\"all\"]}]}"
	tf.Client = &fake.RESTClient{
		NegotiatedSerializer: ns,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(fakeDiscovery))),
			}, nil
		}),
	}
	tf.ClientConfigVal = &restclient.Config{}
	tf.Namespace = "test"

	ioStreams, _, buf, _ := genericclioptions.NewTestIOStreams()
	cmd := NewCmdCreateDeployment(tf, ioStreams)
	cmd.Flags().Set("dry-run", "true")
	cmd.Flags().Set("output", "name")
	cmd.Flags().Set("image", "hollywood/jonny.depp:v2")
	cmd.Run(cmd, []string{depName})
	expectedOutput := "deployment.apps/" + depName + "\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output: %s, but got: %s", expectedOutput, buf.String())
	}
}

func TestCreateDeploymentNoImage(t *testing.T) {
	depName := "jonny-dep"
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()

	ns := legacyscheme.Codecs
	fakeDiscovery := "{\"kind\":\"APIResourceList\",\"apiVersion\":\"v1\",\"groupVersion\":\"apps/v1\",\"resources\":[{\"name\":\"deployments\",\"singularName\":\"\",\"namespaced\":true,\"kind\":\"Deployment\",\"verbs\":[\"create\",\"delete\",\"deletecollection\",\"get\",\"list\",\"patch\",\"update\",\"watch\"],\"shortNames\":[\"deploy\"],\"categories\":[\"all\"]}]}"
	tf.Client = &fake.RESTClient{
		NegotiatedSerializer: ns,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(fakeDiscovery))),
			}, nil
		}),
	}
	tf.ClientConfigVal = &restclient.Config{}
	tf.Namespace = "test"

	ioStreams := genericclioptions.NewTestIOStreamsDiscard()
	cmd := NewCmdCreateDeployment(tf, ioStreams)
	cmd.Flags().Set("output", "name")
	options := &DeploymentOpts{
		CreateSubcommandOptions: &CreateSubcommandOptions{
			PrintFlags: NewPrintFlags("created"),
			DryRun:     true,
			IOStreams:  ioStreams,
		},
	}

	err := options.Complete(tf, cmd, []string{depName})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = options.Run(tf)
	assert.Error(t, err, "at least one image must be specified")
}
