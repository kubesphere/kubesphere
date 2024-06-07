/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package workspace

import (
	"context"
	"os"
	"testing"
	"time"

	"kubesphere.io/kubesphere/pkg/controller/controllertest"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"kubesphere.io/kubesphere/pkg/scheme"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var k8sClient client.Client
var k8sManager ctrl.Manager
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc

func TestWorkspaceController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workspace Controller Test Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(klog.NewKlogr())

	By("bootstrapping test environment")
	t := true
	if os.Getenv("TEST_USE_EXISTING_CLUSTER") == "true" {
		testEnv = &envtest.Environment{
			UseExistingCluster: &t,
		}
	} else {
		crdDirPaths, err := controllertest.LoadCrdPath()
		Expect(err).ToNot(HaveOccurred())
		testEnv = &envtest.Environment{
			CRDDirectoryPaths:        crdDirPaths,
			AttachControlPlaneOutput: false,
		}
	}

	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&Reconciler{}).SetupWithManager(&kscontroller.Manager{Manager: k8sManager})
	Expect(err).ToNot(HaveOccurred())

	ctx, cancel = context.WithCancel(context.Background())

	go func() {
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())
})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	Eventually(func() error {
		return testEnv.Stop()
	}, 30*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
})
