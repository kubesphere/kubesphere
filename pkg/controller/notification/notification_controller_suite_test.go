/*
Copyright 2020 The KubeSphere Authors.

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

package notification

import (
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"kubesphere.io/kubesphere/pkg/apis"
	v2alpha1 "kubesphere.io/kubesphere/pkg/apis/notification/v2alpha1"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteName := "Cache Suite"
	RunSpecsWithDefaultAndCustomReporters(t, suiteName, []Reporter{printer.NewlineReporter{}, printer.NewProwReporter(suiteName)})
}

var testenv *envtest.Environment
var cfg *rest.Config
var k8sManager ctrl.Manager

var _ = BeforeSuite(func(done Done) {

	testenv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "config", "crds")},
	}

	var err error
	cfg, err = testenv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = v2alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = apis.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: "0",
	})
	Expect(err).ToNot(HaveOccurred())

	r, err := NewController(fake.NewSimpleClientset(), k8sManager.GetClient(), k8sManager.GetCache())
	Expect(err).ToNot(HaveOccurred())
	err = k8sManager.Add(r)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	Expect(testenv.Stop()).To(Succeed())
})
