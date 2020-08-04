/*
Copyright 2020 KubeSphere Authors

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

package e2e_test

import (
	"flag"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
	"kubesphere.io/kubesphere/pkg/test"
)

var ctx *test.TestCtx

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Networking E2e Suite")
}

var _ = BeforeSuite(func() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "false")
	flag.Set("v", "4")
	flag.Parse()
	klog.SetOutput(GinkgoWriter)

	ctx = test.NewTestCtx(nil, os.Getenv("TEST_NAMESPACE"))
	Expect(ctx.Setup(os.Getenv("YAML_PATH"), "", v1alpha1.AddToScheme)).ShouldNot(HaveOccurred())
	deployName := os.Getenv("DEPLOY_NAME")
	Expect(test.WaitForController(ctx.Client, ctx.Namespace, deployName, 1, time.Second*5, time.Minute)).ShouldNot(HaveOccurred(), "Controlller failed to start")
	klog.Infoln("Controller is up, begin to test ")
})

var _ = AfterSuite(func() {
	ctx.Cleanup(nil)
})
