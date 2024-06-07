package v1alpha2

import (
	"log"
	"testing"

	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestStorageResourceQuota(t *testing.T) {
	scheme := runtime.NewScheme()
	err := SchemeBuilder.AddToScheme(scheme)
	if err != nil {
		log.Fatal(err)
	}
	c := fake.NewClientBuilder().WithScheme(scheme).Build()

	key := types.NamespacedName{
		Name: "foo",
	}
	created := &ResourceQuota{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ResourceQuota",
			APIVersion: "quota.kubesphere.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Spec: ResourceQuotaSpec{
			LabelSelector: map[string]string{},
		},
	}
	g := gomega.NewGomegaWithT(t)

	// Test Create
	fetched := &ResourceQuota{
		Spec: ResourceQuotaSpec{
			LabelSelector: map[string]string{},
		},
	}
	g.Expect(c.Create(context.TODO(), created)).To(gomega.Succeed())

	g.Expect(c.Get(context.TODO(), key, fetched)).To(gomega.Succeed())
	g.Expect(fetched).To(gomega.Equal(created))

	// Test Updating the Labels
	updated := fetched.DeepCopy()
	updated.Labels = map[string]string{"hello": "world"}
	g.Expect(c.Update(context.TODO(), updated)).To(gomega.Succeed())

	g.Expect(c.Get(context.TODO(), key, fetched)).To(gomega.Succeed())
	g.Expect(fetched).To(gomega.Equal(updated))

	// Test Delete
	g.Expect(c.Delete(context.TODO(), fetched)).To(gomega.Succeed())
	g.Expect(c.Get(context.TODO(), key, fetched)).ToNot(gomega.Succeed())
}
