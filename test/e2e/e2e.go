package e2e

import (
	"testing"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/test/e2e/framework/ginkgowrapper"
)

// RunE2ETests checks configuration parameters (specified through flags) and then runs
// E2E tests using the Ginkgo runner.
// This function is called on each Ginkgo node in parallel mode.
func RunE2ETests(t *testing.T) {
	gomega.RegisterFailHandler(ginkgowrapper.Fail)
	klog.Infof("Starting e2e run on Ginkgo node %d", config.GinkgoConfig.ParallelNode)
	ginkgo.RunSpecs(t, "KubeSphere e2e suite")
}
