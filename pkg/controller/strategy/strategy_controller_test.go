/*
Copyright 2019 The KubeSphere authors.

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

package strategy

import (
	"github.com/knative/pkg/apis/istio/common/v1alpha1"
	"github.com/knative/pkg/apis/istio/v1alpha3"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"testing"
	"time"

	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	servicemeshv1alpha2 "kubesphere.io/kubesphere/pkg/apis/servicemesh/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo", Namespace: "default"}}
var depKey = types.NamespacedName{Name: "details", Namespace: "default"}

const timeout = time.Second * 5

var labels = map[string]string{
	"app.kubernetes.io/name":            "details",
	"app.kubernetes.io/version":         "v1",
	"app":                               "details",
	"servicemesh.kubesphere.io/enabled": "",
}

var svc = v1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "details",
		Namespace: "default",
		Labels:    labels,
	},
	Spec: v1.ServiceSpec{
		Ports: []v1.ServicePort{
			{
				Name:     "http",
				Port:     8080,
				Protocol: v1.ProtocolTCP,
			},
		},
		Selector: labels,
	},
}

func TestReconcile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	instance := &servicemeshv1alpha2.Strategy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
			Labels:    labels,
		},
		Spec: servicemeshv1alpha2.StrategySpec{
			Type: servicemeshv1alpha2.CanaryType,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: servicemeshv1alpha2.VirtualServiceTemplateSpec{
				Spec: v1alpha3.VirtualServiceSpec{
					Hosts:    []string{"details"},
					Gateways: []string{"default"},
					Http: []v1alpha3.HTTPRoute{
						{
							Match: []v1alpha3.HTTPMatchRequest{
								{
									Method: &v1alpha1.StringMatch{
										Exact: "POST",
									},
								},
							},
							Route: []v1alpha3.DestinationWeight{
								{
									Destination: v1alpha3.Destination{
										Host:   "details",
										Subset: "v1",
									},
									Weight: 60,
								},
							},
						},
						{
							Route: []v1alpha3.DestinationWeight{
								{
									Destination: v1alpha3.Destination{
										Host:   "details",
										Subset: "v2",
									},
									Weight: 40,
								},
							},
						},
					},
				},
			},
		},
	}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mgr.GetClient()

	recFn, requests := SetupTestReconcile(newReconciler(mgr))
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	err = c.Create(context.TODO(), &svc)
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create service, %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	//defer c.Delete(context.TODO(), &svc)

	// Create the Strategy object and expect the Reconcile and Deployment to be created
	err = c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}

	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	vs := &v1alpha3.VirtualService{}
	g.Eventually(func() error { return c.Get(context.TODO(), depKey, vs) }, timeout).
		Should(gomega.Succeed())

	if str, err := json.Marshal(vs); err == nil {
		t.Logf("Created virtual service %s\n", str)
	}

	// Delete the Deployment and expect Reconcile to be called for Deployment deletion
	g.Expect(c.Delete(context.TODO(), vs)).NotTo(gomega.HaveOccurred())
	//g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	//g.Eventually(func() error { return c.Get(context.TODO(), depKey, vs) }, timeout).Should(gomega.Succeed())

	// Manually delete Deployment since GC isn't enabled in the test control plane
	g.Eventually(func() error { return c.Delete(context.TODO(), vs) }, timeout).
		Should(gomega.MatchError("virtualservices.networking.istio.io \"details\" not found"))

}
