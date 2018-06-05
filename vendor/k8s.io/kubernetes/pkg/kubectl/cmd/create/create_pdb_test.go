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
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	cmdtesting "k8s.io/kubernetes/pkg/kubectl/cmd/testing"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions"
)

func TestCreatePdb(t *testing.T) {
	pdbName := "my-pdb"
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()

	ns := legacyscheme.Codecs

	tf.Client = &fake.RESTClient{
		GroupVersion:         schema.GroupVersion{Group: "policy", Version: "v1beta1"},
		NegotiatedSerializer: ns,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(&bytes.Buffer{}),
			}, nil
		}),
	}
	tf.ClientConfigVal = &restclient.Config{}
	tf.Namespace = "test"

	outputFormat := "name"

	ioStreams, _, buf, _ := genericclioptions.NewTestIOStreams()
	cmd := NewCmdCreatePodDisruptionBudget(tf, ioStreams)
	cmd.Flags().Set("min-available", "1")
	cmd.Flags().Set("selector", "app=rails")
	cmd.Flags().Set("dry-run", "true")
	cmd.Flags().Set("output", outputFormat)

	printFlags := NewPrintFlags("created")
	printFlags.OutputFormat = &outputFormat

	options := &PodDisruptionBudgetOpts{
		CreateSubcommandOptions: &CreateSubcommandOptions{
			PrintFlags: printFlags,
			Name:       pdbName,
			IOStreams:  ioStreams,
		},
	}
	err := options.Complete(cmd, []string{pdbName})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = options.Run(tf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedOutput := "poddisruptionbudget.policy/" + pdbName + "\n"
	if buf.String() != expectedOutput {
		t.Errorf("expected output: %s, but got: %s", expectedOutput, buf.String())
	}
}
