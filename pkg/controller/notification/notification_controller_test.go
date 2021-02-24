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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v2 "kubesphere.io/kubesphere/pkg/apis/notification/v2"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/constants"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var (
	_ = Describe("Secret", func() {

		const timeout = time.Second * 30
		const interval = time.Second * 1

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: constants.NotificationSecretNamespace,
			},
		}

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: constants.NotificationSecretNamespace,
			},
		}

		var (
			cl                  client.Client
			ksCache             cache.Cache
			informerCacheCtx    context.Context
			informerCacheCancel context.CancelFunc
		)

		BeforeEach(func() {
			var err error
			cl, err = client.New(cfg, client.Options{})
			Expect(err).NotTo(HaveOccurred())

			ksCache, err = cache.New(cfg, cache.Options{})
			Expect(err).NotTo(HaveOccurred())

			informerCacheCtx, informerCacheCancel = context.WithCancel(context.Background())
			go func(ctx context.Context) {
				defer GinkgoRecover()
				Expect(ksCache.Start(ctx.Done())).To(Succeed())
			}(informerCacheCtx)
			Expect(ksCache.WaitForCacheSync(informerCacheCtx.Done())).To(BeTrue())

			Eventually(func() bool {
				err = cl.Create(informerCacheCtx, namespace)
				if err == nil || errors.IsAlreadyExists(err) {
					return true
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})

		AfterEach(func() {
			By("cleaning up")
			informerCacheCancel()
		})

		// Add Tests for OpenAPI validation (or additonal CRD features) specified in
		// your API definition.
		// Avoid adding tests for vanilla CRUD operations because they would
		// test Kubernetes API server, which isn't the goal here.
		Context("Notification Controller", func() {
			It("Should create successfully", func() {

				// Create a secret
				Expect(cl.Create(context.Background(), secret)).Should(Succeed())
				time.Sleep(time.Second)

				fedSecret := &v1beta1.FederatedSecret{}
				By("Expecting to create federated secret successfully")
				Eventually(func() bool {
					err := ksCache.Get(context.Background(), client.ObjectKey{Name: secret.Name, Namespace: constants.NotificationSecretNamespace}, fedSecret)
					Expect(err).Should(Succeed())
					return !fedSecret.CreationTimestamp.IsZero()
				}, timeout, interval).Should(BeTrue())

				err := ksCache.Get(context.Background(), client.ObjectKey{Name: secret.Name, Namespace: constants.NotificationSecretNamespace}, secret)
				Expect(err).Should(Succeed())
				secret.StringData = map[string]string{"foo": "bar"}
				Expect(cl.Update(context.Background(), secret)).Should(Succeed())
				time.Sleep(time.Second)

				By("Expecting to update federated secret successfully")
				Eventually(func() bool {
					err := ksCache.Get(context.Background(), client.ObjectKey{Name: secret.Name, Namespace: constants.NotificationSecretNamespace}, fedSecret)
					Expect(err).Should(Succeed())
					return string(fedSecret.Spec.Template.Data["foo"]) == "bar"
				}, timeout, interval).Should(BeTrue())
			})
		})
	})

	_ = Describe("Notification", func() {

		const timeout = time.Second * 30
		const interval = time.Second * 1

		obj := &v2.DingTalkConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: constants.NotificationSecretNamespace,
				Labels: map[string]string{
					"type": "default",
				},
			},
		}

		var (
			cl                  client.Client
			ksCache             cache.Cache
			informerCacheCtx    context.Context
			informerCacheCancel context.CancelFunc
		)

		BeforeEach(func() {
			var err error
			cl, err = client.New(cfg, client.Options{})
			Expect(err).NotTo(HaveOccurred())

			ksCache, err = cache.New(cfg, cache.Options{})
			Expect(err).NotTo(HaveOccurred())

			informerCacheCtx, informerCacheCancel = context.WithCancel(context.Background())
			go func(ctx context.Context) {
				defer GinkgoRecover()
				Expect(ksCache.Start(ctx.Done())).To(Succeed())
			}(informerCacheCtx)
			Expect(ksCache.WaitForCacheSync(informerCacheCtx.Done())).To(BeTrue())
		})

		AfterEach(func() {
			By("cleaning up")
			informerCacheCancel()
		})

		// Add Tests for OpenAPI validation (or additonal CRD features) specified in
		// your API definition.
		// Avoid adding tests for vanilla CRUD operations because they would
		// test Kubernetes API server, which isn't the goal here.
		Context("Notification Controller", func() {
			It("Should create successfully", func() {

				// Create a bject
				Expect(cl.Create(context.Background(), obj)).Should(Succeed())
				time.Sleep(time.Second)

				fedObj := &v1beta1.FederatedDingTalkConfig{}
				By("Expecting to create federated object successfully")
				Eventually(func() bool {
					err := ksCache.Get(context.Background(), client.ObjectKey{Name: obj.Name}, fedObj)
					Expect(err).Should(Succeed())
					return !fedObj.CreationTimestamp.IsZero()
				}, timeout, interval).Should(BeTrue())

				err := ksCache.Get(context.Background(), client.ObjectKey{Name: obj.Name}, obj)
				Expect(err).Should(Succeed())
				obj.Labels = map[string]string{"foo": "bar"}
				Expect(cl.Update(context.Background(), obj)).Should(Succeed())
				time.Sleep(time.Second)

				By("Expecting to update federated object successfully")
				Eventually(func() bool {
					err := ksCache.Get(context.Background(), client.ObjectKey{Name: obj.Name}, fedObj)
					Expect(err).Should(Succeed())
					return fedObj.Spec.Template.Labels["foo"] == "bar"
				}, timeout, interval).Should(BeTrue())
			})
		})
	})
)
