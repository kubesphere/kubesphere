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
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"kubesphere.io/api/cluster/v1alpha1"
	"kubesphere.io/api/notification/v2beta2"
	"kubesphere.io/api/types/v1beta1"
	"kubesphere.io/api/types/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"

	"kubesphere.io/kubesphere/pkg/apis"
	"kubesphere.io/kubesphere/pkg/constants"
)

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteName := "Cache Suite"
	RunSpecsWithDefaultAndCustomReporters(t, suiteName, []Reporter{printer.NewlineReporter{}, printer.NewProwReporter(suiteName)})
}

var (
	_ = Describe("Secret", func() {
		_ = v2beta2.AddToScheme(scheme.Scheme)
		_ = apis.AddToScheme(scheme.Scheme)
		_ = v1beta2.AddToScheme(scheme.Scheme)

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret-foo",
				Namespace: constants.NotificationSecretNamespace,
				Labels: map[string]string{
					constants.NotificationManagedLabel: "true",
				},
			},
		}

		config := &v2beta2.Config{
			ObjectMeta: metav1.ObjectMeta{
				Name: "config-foo",
				Labels: map[string]string{
					"type": "global",
				},
			},
		}

		receiver := &v2beta2.Receiver{
			ObjectMeta: metav1.ObjectMeta{
				Name: "receiver-foo",
				Labels: map[string]string{
					"type": "default",
				},
			},
		}

		router := &v2beta2.Router{
			ObjectMeta: metav1.ObjectMeta{
				Name: "router-foo",
			},
		}

		silence := &v2beta2.Silence{
			ObjectMeta: metav1.ObjectMeta{
				Name: "silence-foo",
				Labels: map[string]string{
					"type": "global",
				},
			},
		}

		host := &v1alpha1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: "host",
			},
		}

		var (
			cl               client.Client
			ksCache          cache.Cache
			k8sClient        kubernetes.Interface
			informerCacheCtx context.Context
		)
		BeforeEach(func() {
			k8sClient = fakek8s.NewSimpleClientset()
			//nolint:staticcheck
			cl = fake.NewFakeClientWithScheme(scheme.Scheme)
			informerCacheCtx = context.TODO()
			ksCache = &fakeCache{
				k8sClient,
				cl,
			}
		})

		// Add Tests for OpenAPI validation (or additional CRD features) specified in
		// your API definition.
		// Avoid adding tests for vanilla CRUD operations because they would
		// test Kubernetes API server, which isn't the goal here.
		Context("Notification Controller", func() {
			It("Should create successfully", func() {

				r, err := NewController(k8sClient, cl, ksCache)
				Expect(err).ToNot(HaveOccurred())

				// Create a secret
				Expect(cl.Create(context.Background(), secret)).Should(Succeed())
				Expect(r.reconcile(secret)).Should(Succeed())

				fedSecret := &v1beta1.FederatedSecret{}
				By("Expecting to create federated secret successfully")

				err = ksCache.Get(context.Background(), client.ObjectKey{Name: secret.Name, Namespace: constants.NotificationSecretNamespace}, fedSecret)
				Expect(err).Should(Succeed())
				Expect(fedSecret.Name).Should(Equal(secret.Name))

				// Update a secret
				err = ksCache.Get(context.Background(), client.ObjectKey{Name: secret.Name, Namespace: constants.NotificationSecretNamespace}, secret)
				Expect(err).Should(Succeed())
				secret.StringData = map[string]string{"foo": "bar"}
				Expect(cl.Update(context.Background(), secret)).Should(Succeed())
				Expect(r.reconcile(secret)).Should(Succeed())

				By("Expecting to update federated secret successfully")

				err = ksCache.Get(context.Background(), client.ObjectKey{Name: secret.Name, Namespace: constants.NotificationSecretNamespace}, fedSecret)
				Expect(err).Should(Succeed())
				Expect(fedSecret.Spec.Template.StringData["foo"]).Should(Equal("bar"))

				// Create a Config
				Expect(cl.Create(context.Background(), config)).Should(Succeed())
				Expect(r.reconcile(config)).Should(Succeed())

				fedConfig := &v1beta2.FederatedNotificationConfig{}
				By("Expecting to create federated object successfully")
				err = ksCache.Get(context.Background(), client.ObjectKey{Name: config.Name}, fedConfig)
				Expect(err).Should(Succeed())
				Expect(fedConfig.Name).Should(Equal(config.Name))

				// Update a config
				err = ksCache.Get(context.Background(), client.ObjectKey{Name: config.Name}, config)
				Expect(err).Should(Succeed())
				config.Labels = map[string]string{"foo": "bar"}
				Expect(cl.Update(context.Background(), config)).Should(Succeed())
				Expect(r.reconcile(config)).Should(Succeed())

				By("Expecting to update federated object successfully")
				err = ksCache.Get(context.Background(), client.ObjectKey{Name: config.Name}, fedConfig)
				Expect(err).Should(Succeed())
				Expect(fedConfig.Spec.Template.Labels["foo"]).Should(Equal("bar"))

				// Create a receiver
				Expect(cl.Create(context.Background(), receiver)).Should(Succeed())
				Expect(r.reconcile(receiver)).Should(Succeed())

				fedReceiver := &v1beta2.FederatedNotificationReceiver{}
				By("Expecting to create federated object successfully")
				err = ksCache.Get(context.Background(), client.ObjectKey{Name: receiver.Name}, fedReceiver)
				Expect(err).Should(Succeed())
				Expect(fedReceiver.Name).Should(Equal(receiver.Name))

				// Update a receiver
				err = ksCache.Get(context.Background(), client.ObjectKey{Name: receiver.Name}, receiver)
				Expect(err).Should(Succeed())
				receiver.Labels = map[string]string{"foo": "bar"}
				Expect(cl.Update(context.Background(), receiver)).Should(Succeed())
				Expect(r.reconcile(receiver)).Should(Succeed())

				By("Expecting to update federated object successfully")

				err = ksCache.Get(context.Background(), client.ObjectKey{Name: receiver.Name}, fedReceiver)
				Expect(err).Should(Succeed())
				Expect(fedReceiver.Spec.Template.Labels["foo"]).Should(Equal("bar"))

				// Create a router
				Expect(cl.Create(context.Background(), router)).Should(Succeed())
				Expect(r.reconcile(router)).Should(Succeed())

				fedRouter := &v1beta2.FederatedNotificationRouter{}
				By("Expecting to create federated object successfully")
				err = ksCache.Get(context.Background(), client.ObjectKey{Name: router.Name}, fedRouter)
				Expect(err).Should(Succeed())
				Expect(fedRouter.Name).Should(Equal(router.Name))

				// Update a receiver
				err = ksCache.Get(context.Background(), client.ObjectKey{Name: router.Name}, router)
				Expect(err).Should(Succeed())
				router.Labels = map[string]string{"foo": "bar"}
				Expect(cl.Update(context.Background(), router)).Should(Succeed())
				Expect(r.reconcile(router)).Should(Succeed())

				By("Expecting to update federated object successfully")

				err = ksCache.Get(context.Background(), client.ObjectKey{Name: router.Name}, fedRouter)
				Expect(err).Should(Succeed())
				Expect(fedRouter.Spec.Template.Labels["foo"]).Should(Equal("bar"))

				// Create a receiver
				Expect(cl.Create(context.Background(), silence)).Should(Succeed())
				Expect(r.reconcile(silence)).Should(Succeed())

				fedSilence := &v1beta2.FederatedNotificationSilence{}
				By("Expecting to create federated object successfully")
				err = ksCache.Get(context.Background(), client.ObjectKey{Name: silence.Name}, fedSilence)
				Expect(err).Should(Succeed())
				Expect(fedSilence.Name).Should(Equal(silence.Name))

				// Update a receiver
				err = ksCache.Get(context.Background(), client.ObjectKey{Name: silence.Name}, silence)
				Expect(err).Should(Succeed())
				silence.Labels = map[string]string{"foo": "bar"}
				Expect(cl.Update(context.Background(), silence)).Should(Succeed())
				Expect(r.reconcile(silence)).Should(Succeed())

				By("Expecting to update federated object successfully")

				err = ksCache.Get(context.Background(), client.ObjectKey{Name: silence.Name}, fedSilence)
				Expect(err).Should(Succeed())
				Expect(fedSilence.Spec.Template.Labels["foo"]).Should(Equal("bar"))

				// Add a cluster
				Expect(cl.Create(informerCacheCtx, host)).Should(Succeed())
				Expect(r.reconcile(secret)).Should(Succeed())

				By("Expecting to update federated secret successfully")

				err = ksCache.Get(context.Background(), client.ObjectKey{Name: secret.Name, Namespace: constants.NotificationSecretNamespace}, fedSecret)
				Expect(err).Should(Succeed())

				Expect(fedSecret.Spec.Overrides[0].ClusterName).Should(Equal("host"))

				// Delete a cluster
				Expect(cl.Delete(informerCacheCtx, host)).Should(Succeed())
				Expect(r.reconcile(secret)).Should(Succeed())

				By("Expecting to update federated secret successfully")

				fedSecret = &v1beta1.FederatedSecret{}
				err = ksCache.Get(context.Background(), client.ObjectKey{Name: secret.Name, Namespace: constants.NotificationSecretNamespace}, fedSecret)
				Expect(err).Should(Succeed())
				Expect(fedSecret.Spec.Overrides).Should(BeNil())

			})
		})
	})
)

const defaultResync = 600 * time.Second

type fakeCache struct {
	K8sClient kubernetes.Interface
	client.Reader
}

// GetInformerForKind returns the informer for the GroupVersionKind
func (f *fakeCache) GetInformerForKind(_ context.Context, _ schema.GroupVersionKind) (cache.Informer, error) {
	return nil, nil
}

// GetInformer returns the informer for the obj
func (f *fakeCache) GetInformer(_ context.Context, _ client.Object) (cache.Informer, error) {
	fakeInformerFactory := k8sinformers.NewSharedInformerFactory(f.K8sClient, defaultResync)
	return fakeInformerFactory.Core().V1().Namespaces().Informer(), nil
}

func (f *fakeCache) IndexField(_ context.Context, _ client.Object, _ string, _ client.IndexerFunc) error {
	return nil
}

func (f *fakeCache) Start(_ context.Context) error {
	return nil
}

func (f *fakeCache) WaitForCacheSync(_ context.Context) bool {
	return true
}
