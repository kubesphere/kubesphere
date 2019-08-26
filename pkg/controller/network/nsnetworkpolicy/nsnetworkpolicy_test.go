package nsnetworkpolicy

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
	nsnplister "kubesphere.io/kubesphere/pkg/client/listers/network/v1alpha1"
	"kubesphere.io/kubesphere/pkg/controller/network/controllerapi"
	"kubesphere.io/kubesphere/pkg/controller/network/provider"
	controllertesting "kubesphere.io/kubesphere/pkg/controller/network/testing"
)

var (
	fakeControllerBuilder *controllertesting.FakeControllerBuilder
	c                     controllerapi.Controller
	stopCh                chan struct{}
	calicoProvider        *provider.FakeCalicoNetworkProvider
	nsnpLister            nsnplister.NamespaceNetworkPolicyLister
)

var _ = Describe("Nsnetworkpolicy", func() {
	BeforeEach(func() {
		fakeControllerBuilder = controllertesting.NewFakeControllerBuilder()
		stopCh = make(chan struct{})
		informer, _ := fakeControllerBuilder.NewControllerInformer()
		calicoProvider = provider.NewFakeCalicoNetworkProvider()
		c = NewController(fakeControllerBuilder.KubeClient, fakeControllerBuilder.KsClient, informer.Network().V1alpha1().NamespaceNetworkPolicies(), calicoProvider)
		go informer.Network().V1alpha1().NamespaceNetworkPolicies().Informer().Run(stopCh)
		originalController := c.(*controller)
		originalController.recorder = &record.FakeRecorder{}
		go c.Run(1, stopCh)
		nsnpLister = informer.Network().V1alpha1().NamespaceNetworkPolicies().Lister()
	})

	It("Should create a new calico object", func() {
		objSrt := `{
			"apiVersion": "network.kubesphere.io/v1alpha1",
			"kind": "NetworkPolicy",
			"metadata": {
				"name": "allow-tcp-6379",
				"namespace": "production"
			},
			"spec": {
				"selector": "color == 'red'",
				"ingress": [
					{
						"action": "Allow",
						"protocol": "TCP",
						"source": {
							"selector": "color == 'blue'"
						},
						"destination": {
							"ports": [
								6379
							]
						}
					}
				]
			}
		}`
		obj := &v1alpha1.NamespaceNetworkPolicy{}
		Expect(controllertesting.StringToObject(objSrt, obj)).ShouldNot(HaveOccurred())
		_, err := fakeControllerBuilder.KsClient.NetworkV1alpha1().NamespaceNetworkPolicies(obj.Namespace).Create(obj)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(func() bool {
			exist, _ := calicoProvider.CheckExist(obj)
			return exist
		}).Should(BeTrue())
		obj, _ = fakeControllerBuilder.KsClient.NetworkV1alpha1().NamespaceNetworkPolicies(obj.Namespace).Get(obj.Name, metav1.GetOptions{})
		Expect(obj.Finalizers).To(HaveLen(1))
		// TestUpdate
		newStr := "color == 'green'"
		obj.Spec.Selector = newStr
		_, err = fakeControllerBuilder.KsClient.NetworkV1alpha1().NamespaceNetworkPolicies(obj.Namespace).Update(obj)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(func() string {
			o, err := calicoProvider.Get(obj)
			if err != nil {
				return err.Error()
			}
			n := o.(*v1alpha1.NamespaceNetworkPolicy)
			return n.Spec.Selector
		}).Should(Equal(newStr))
		// TestDelete
		Expect(fakeControllerBuilder.KsClient.NetworkV1alpha1().NamespaceNetworkPolicies(obj.Namespace).Delete(obj.Name, &metav1.DeleteOptions{})).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		close(stopCh)
	})
})
