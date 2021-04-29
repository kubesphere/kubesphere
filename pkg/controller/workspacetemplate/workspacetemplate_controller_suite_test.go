/*
Copyright 2019 The KubeSphere Authors.

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

package workspacetemplate

import (
	"testing"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/klogr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	helmappscheme "kubesphere.io/api/application/v1alpha1"
	typesv1beta1 "kubesphere.io/api/types/v1beta1"

	"kubesphere.io/kubesphere/pkg/apis"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
)

func TestWorkspaceTemplateController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t,
		"WorkspaceTemplate Controller Test Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(klogr.New())

	err := helmappscheme.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = apis.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = typesv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	close(done)
}, 60)
