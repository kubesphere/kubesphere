// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package v1beta1

import (
	"testing"

	"context"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestStorageApplication(t *testing.T) {
	key := types.NamespacedName{
		Name:      "foo",
		Namespace: "default",
	}
	created := &Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		}}
	g := gomega.NewGomegaWithT(t)

	// Test Create
	fetched := &Application{}
	g.Expect(c.Create(context.TODO(), created)).NotTo(gomega.HaveOccurred())

	g.Expect(c.Get(context.TODO(), key, fetched)).NotTo(gomega.HaveOccurred())
	g.Expect(fetched).To(gomega.Equal(created))

	// Test Updating the Labels
	updated := fetched.DeepCopy()
	updated.Labels = map[string]string{"hello": "world"}
	g.Expect(c.Update(context.TODO(), updated)).NotTo(gomega.HaveOccurred())

	g.Expect(c.Get(context.TODO(), key, fetched)).NotTo(gomega.HaveOccurred())
	g.Expect(fetched).To(gomega.Equal(updated))

	// Test Delete
	g.Expect(c.Delete(context.TODO(), fetched)).NotTo(gomega.HaveOccurred())
	g.Expect(c.Get(context.TODO(), key, fetched)).To(gomega.HaveOccurred())

	// Test stripVersion()
	g.Expect(StripVersion("")).To(gomega.Equal(""))
	g.Expect(StripVersion("v1beta1")).To(gomega.Equal(""))
	g.Expect(StripVersion("apps/v1")).To(gomega.Equal("apps"))
	g.Expect(StripVersion("apps/v1alpha2")).To(gomega.Equal("apps"))
}
