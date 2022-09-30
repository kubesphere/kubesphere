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

package v1alpha2

import (
	"log"
	"os"
	"testing"

	apinetworkingv1alpha3 "istio.io/api/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestStorageStrategy(t *testing.T) {
	err := SchemeBuilder.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatal(err)
	}
	c := fake.NewFakeClientWithScheme(scheme.Scheme)

	key := types.NamespacedName{
		Name:      "foo",
		Namespace: "default",
	}
	created := &Strategy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Strategy",
			APIVersion: "servicemesh.kubesphere.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: StrategySpec{
			Template: VirtualServiceTemplateSpec{
				Spec: apinetworkingv1alpha3.VirtualService{
					Hosts: []string{
						"details",
					},
				},
			},
		},
	}
	g := gomega.NewGomegaWithT(t)

	// Test Create
	fetched := &Strategy{}
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
}

func TestStrategyRead(t *testing.T) {

	//g := gomega.NewGomegaWithT(t)

	var str []byte
	file, err := os.ReadFile("/Users/zry/go/src/kubesphere.io/kubesphere/config/samples/servicemesh_v1alpha2_strategy.yaml")
	if err == nil {
		obj, _, _ := scheme.Codecs.UniversalDeserializer().Decode(file, nil, &Strategy{})
		switch obj.(type) {
		case *Strategy:
			str, err = json.Marshal(obj)
			t.Logf("Read strategy %s", str)
		}
	}
}
