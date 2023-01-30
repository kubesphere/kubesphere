/*
Copyright 2021 KubeSphere Authors

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

package helm

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"kubesphere.io/kubesphere/pkg/simple/client/gateway"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var testEnv *envtest.Environment

func TestApplicationController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t,
		"Application Controller Test Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(klog.NewKlogr())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:        []string{filepath.Join("..", "..", "..", "config", "ks-core", "crds")},
		AttachControlPlaneOutput: false,
	}
	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

var _ = Context("Helm reconcier", func() {
	Describe("Gateway", func() {
		It("Should setup gateway helm reconcier", func() {
			data := "- group: gateway.kubesphere.io\n  version: v1alpha1\n  kind: Gateway\n  chart: ../../../config/gateway\n"
			f, _ := os.CreateTemp("", "watch")
			os.WriteFile(f.Name(), []byte(data), 0)

			mgr, err := ctrl.NewManager(cfg, ctrl.Options{MetricsBindAddress: "0"})
			Expect(err).NotTo(HaveOccurred(), "failed to create a manager")

			reconciler := &Reconciler{GatewayOptions: &gateway.Options{WatchesPath: f.Name()}}
			err = reconciler.SetupWithManager(mgr)
			Expect(err).NotTo(HaveOccurred(), "failed to setup helm reconciler")

		})
	})
})
