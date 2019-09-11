package wsnetworkpolicy

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	k8snetwork "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	netv1lister "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
	tenant "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/controller/network/controllerapi"
	controllertesting "kubesphere.io/kubesphere/pkg/controller/network/testing"
)

var (
	fakeControllerBuilder *controllertesting.FakeControllerBuilder
	c                     controllerapi.Controller
	npLister              netv1lister.NetworkPolicyLister
	stopCh                chan struct{}
	deletePolicy          metav1.DeletionPropagation
	testName              string
)

var _ = Describe("Wsnetworkpolicy", func() {
	BeforeEach(func() {
		deletePolicy = metav1.DeletePropagationBackground
		fakeControllerBuilder = controllertesting.NewFakeControllerBuilder()
		informer, k8sinformer := fakeControllerBuilder.NewControllerInformer()
		stopCh = make(chan struct{})
		c = NewController(fakeControllerBuilder.KubeClient, fakeControllerBuilder.KsClient,
			informer.Network().V1alpha1().WorkspaceNetworkPolicies(), k8sinformer.Networking().V1().NetworkPolicies(),
			k8sinformer.Core().V1().Namespaces(), informer.Tenant().V1alpha1().Workspaces())
		originalController := c.(*controller)
		go originalController.wsnpInformer.Informer().Run(stopCh)
		go originalController.networkPolicyInformer.Informer().Run(stopCh)
		go originalController.namespaceInformer.Informer().Run(stopCh)
		go originalController.workspaceInformer.Informer().Run(stopCh)
		originalController.recorder = &record.FakeRecorder{}
		go c.Run(1, stopCh)
		npLister = k8sinformer.Networking().V1().NetworkPolicies().Lister()
		testName = "test"
		ns1 := newWorkspaceNamespaces("ns1", testName)
		ns2 := newWorkspaceNamespaces("ns2", testName)
		_, err := fakeControllerBuilder.KubeClient.CoreV1().Namespaces().Create(ns1)
		Expect(err).ShouldNot(HaveOccurred())
		_, err = fakeControllerBuilder.KubeClient.CoreV1().Namespaces().Create(ns2)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		close(stopCh)
	})

	It("Should proper ingress rule when using workspaceSelector", func() {
		label := map[string]string{"workspace": "test-selector"}
		ws := &tenant.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "test",
				Labels: label,
			},
		}
		_, err := fakeControllerBuilder.KsClient.TenantV1alpha1().Workspaces().Create(ws)
		wsnp := newWorkspaceNP(testName)
		wsnp.Spec.PolicyTypes = []k8snetwork.PolicyType{k8snetwork.PolicyTypeIngress}
		wsnp.Spec.Ingress = []v1alpha1.WorkspaceNetworkPolicyIngressRule{
			{
				From: []v1alpha1.WorkspaceNetworkPolicyPeer{
					{
						WorkspaceSelector: &metav1.LabelSelector{
							MatchLabels: label,
						},
					},
				},
			}}
		_, err = fakeControllerBuilder.KsClient.NetworkV1alpha1().WorkspaceNetworkPolicies().Create(wsnp)
		Expect(err).ShouldNot(HaveOccurred())
		expect1Json := `{
			"apiVersion": "networking.k8s.io/v1",
			"kind": "NetworkPolicy",
			"metadata": {
				"name": "test-np",
				"namespace": "ns1",
				"labels": {
					"networking.kubesphere.io/wsnp": "test"
				}
			},
			"spec": {
				"policyTypes": [
					"Ingress"
				],
				"ingress": [
					{
						"from": [
							{
								"namespaceSelector": { 
									"matchLabels": {
										"kubesphere.io/workspace": "test"
									}
								}
							}
						]
					}
				]
			}
		}`
		expect1 := &k8snetwork.NetworkPolicy{}
		Expect(controllertesting.StringToObject(expect1Json, expect1)).ShouldNot(HaveOccurred())
		nps := []*k8snetwork.NetworkPolicy{}
		Eventually(func() error {
			selector, _ := labels.Parse(workspaceNetworkPolicyLabel + "==test")
			nps, err = npLister.List(selector)
			if err != nil {
				klog.Errorf("Failed to list npmerr:%s", err.Error())
				return err
			}
			if len(nps) != 2 {
				return fmt.Errorf("Length is not right, current length :%d", len(nps))
			}
			return nil
		}, time.Second*5, time.Second).ShouldNot(HaveOccurred())

		for _, np := range nps {
			Expect(np.Labels).To(Equal(expect1.Labels))
			Expect(np.Spec).To(Equal(expect1.Spec))
		}
		// create a new ws will change the `From`
		ws2 := &tenant.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "test2",
				Labels: label,
			},
		}
		_, err = fakeControllerBuilder.KsClient.TenantV1alpha1().Workspaces().Create(ws2)
		Expect(err).ShouldNot(HaveOccurred())
		expect2Json := `{
			"apiVersion": "networking.k8s.io/v1",
			"kind": "NetworkPolicy",
			"metadata": {
				"name": "test-np",
				"namespace": "ns1",
				"labels": {
					"networking.kubesphere.io/wsnp": "test"
				}
			},
			"spec": {
				"policyTypes": [
					"Ingress"
				],
				"ingress": [
					{
						"from": [
							{
								"namespaceSelector": { 
									"matchExpressions": [{
										"key": "kubesphere.io/workspace",
										"operator":"In",
										"values": ["test", "test2"]
									}]
								}
							}
						]
					}
				]
			}
		}`
		expect2 := &k8snetwork.NetworkPolicy{}
		Expect(controllertesting.StringToObject(expect2Json, expect2)).ShouldNot(HaveOccurred())

		id := func(element interface{}) string {
			e := element.(*k8snetwork.NetworkPolicy)
			return e.Namespace
		}
		Eventually(func() []*k8snetwork.NetworkPolicy {
			selector, _ := labels.Parse(workspaceNetworkPolicyLabel + "=test")
			nps, err := npLister.List(selector)
			if err != nil {
				return nil
			}
			if len(nps) != 2 {
				klog.Errorf("Length is not right, current length :%d", len(nps))
				return nil
			}
			return nps
		}, time.Second*5, time.Second).Should(MatchAllElements(id, Elements{
			"ns1": PointTo(MatchFields(IgnoreExtras, Fields{
				"Spec": Equal(expect2.Spec),
			})),
			"ns2": PointTo(MatchFields(IgnoreExtras, Fields{
				"Spec": Equal(expect2.Spec),
			})),
		}))
	})

	It("Should create networkpolicies", func() {
		//create a wsnp
		_, err := fakeControllerBuilder.KsClient.NetworkV1alpha1().WorkspaceNetworkPolicies().Create(newWorkspaceNP(testName))
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(func() error {
			selector, _ := labels.Parse(workspaceNetworkPolicyLabel + "=" + testName)
			nps, err := npLister.List(selector)
			if err != nil {
				return err
			}
			if len(nps) != 2 {
				return fmt.Errorf("Length is not right, current length :%d", len(nps))
			}
			return nil
		}, time.Second*5, time.Second).ShouldNot(HaveOccurred())
		err = fakeControllerBuilder.KsClient.NetworkV1alpha1().WorkspaceNetworkPolicies().Delete(testName, &metav1.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		})
		Expect(err).ShouldNot(HaveOccurred())
	})
})

func newWorkspaceNP(name string) *v1alpha1.WorkspaceNetworkPolicy {
	return &v1alpha1.WorkspaceNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.WorkspaceNetworkPolicySpec{
			Workspace: name,
		},
	}
}

func newWorkspaceNamespaces(ns, ws string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ns,
			Labels: map[string]string{workspaceSelectorLabel: ws},
		},
	}
}
