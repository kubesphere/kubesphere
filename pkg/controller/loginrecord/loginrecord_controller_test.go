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

package loginrecord

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/klogr"
	"kubesphere.io/kubesphere/pkg/apis"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
	"time"
)

var testEnv *envtest.Environment
var k8sManager ctrl.Manager

func TestLoginRecordController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t,
		"LoginRecord Controller Test Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(klogr.New())

	By("bootstrapping test environment")
	t := true
	if os.Getenv("TEST_USE_EXISTING_CLUSTER") == "true" {
		testEnv = &envtest.Environment{
			UseExistingCluster: &t,
		}
	} else {
		testEnv = &envtest.Environment{
			CRDDirectoryPaths:        []string{filepath.Join("..", "..", "..", "config", "crds")},
			AttachControlPlaneOutput: false,
		}
	}

	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = apis.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: "0",
	})
	Expect(err).ToNot(HaveOccurred())

	k8sClient, err := kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())

	ksClient, err := kubesphere.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())

	ksInformers := externalversions.NewSharedInformerFactory(ksClient, time.Second*30)
	Expect(err).NotTo(HaveOccurred())

	loginRecordInformer := ksInformers.Iam().V1alpha2().LoginRecords()
	userInformer := ksInformers.Iam().V1alpha2().Users()

	loginRecordController := NewLoginRecordController(k8sClient, ksClient, loginRecordInformer, userInformer, time.Hour, 1)
	err = k8sManager.Add(loginRecordController)
	Expect(err).NotTo(HaveOccurred())

	go func() {
		stopChan := ctrl.SetupSignalHandler()
		ksInformers.Start(stopChan)
		err = k8sManager.Start(stopChan)
		Expect(err).ToNot(HaveOccurred())
	}()

	close(done)
}, 60)

var _ = Describe("LoginRecord", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1

	BeforeEach(func() {
		admin := &iamv1alpha2.User{
			ObjectMeta: metav1.ObjectMeta{Name: "admin"},
		}
		Expect(k8sManager.GetClient().Create(context.Background(), admin, &client.CreateOptions{})).Should(Succeed())
	})

	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("LoginRecord Controller", func() {
		It("Should create successfully", func() {
			ctx := context.Background()
			username := "admin"
			loginRecord := &iamv1alpha2.LoginRecord{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-1", username),
					Labels: map[string]string{
						iamv1alpha2.UserReferenceLabel: username,
					},
				},
				Spec: iamv1alpha2.LoginRecordSpec{
					Type:      iamv1alpha2.Token,
					Provider:  "",
					Success:   true,
					Reason:    iamv1alpha2.AuthenticatedSuccessfully,
					SourceIP:  "",
					UserAgent: "",
				},
			}

			By("Expecting to create login record successfully")
			Expect(k8sManager.GetClient().Create(ctx, loginRecord, &client.CreateOptions{})).Should(Succeed())

			expected := &iamv1alpha2.LoginRecord{}
			Eventually(func() bool {
				err := k8sManager.GetClient().Get(ctx, types.NamespacedName{Name: loginRecord.Name}, expected)
				fmt.Print(err)
				return !expected.CreationTimestamp.IsZero()
			}, timeout, interval).Should(BeTrue())

			loginRecord.Name = fmt.Sprintf("%s-2", username)
			loginRecord.ResourceVersion = ""
			By("Expecting to create login record successfully")
			Expect(k8sManager.GetClient().Create(ctx, loginRecord, &client.CreateOptions{})).Should(Succeed())

			Eventually(func() bool {
				k8sManager.GetClient().Get(ctx, types.NamespacedName{Name: loginRecord.Name}, expected)
				return !expected.CreationTimestamp.IsZero()
			}, timeout, interval).Should(BeTrue())

			By("Expecting to limit login record successfully")
			Eventually(func() bool {
				loginRecordList := &iamv1alpha2.LoginRecordList{}
				selector := labels.SelectorFromSet(labels.Set{iamv1alpha2.UserReferenceLabel: username})
				k8sManager.GetClient().List(ctx, loginRecordList, &client.ListOptions{LabelSelector: selector})
				return len(loginRecordList.Items) == 1
			}, timeout, interval).Should(BeTrue())
		})
	})
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	gexec.KillAndWait(5 * time.Second)
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
