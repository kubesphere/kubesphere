package secret

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

var testSecret = &v1.Secret{
	TypeMeta: metav1.TypeMeta{
		Kind:       "Secret",
		APIVersion: "v1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:            "prometheus-k8s",
		Namespace:       "kube-system",
		ResourceVersion: "1234567",
		Labels: map[string]string{
			"modifiedAt": "1670227209",
			"name":       "snapshot-controller",
			"owner":      "helm",
			"status":     "superseded",
			"version":    "2",
		},
	},
	Data: map[string][]byte{
		"testdata": []byte("thisisatestsecret"),
	},
	Type: "helm.sh/release.v1",
}

func BenchmarkContains(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if contains(testSecret, "metadata.labels.status!=superseded") {
			b.Error("test failed")
		}
	}

}

func BenchmarkDefaultList(b *testing.B) {
	s := &secretSearcher{}
	secretList := make([]runtime.Object, 0)
	secretList = append(secretList, testSecret)

	for i := 0; i < 20; i++ {
		ttt := testSecret.DeepCopy()
		ttt.ObjectMeta.ResourceVersion = rand.String(10)
		secretList = append(secretList, ttt)
	}
	q := query.New()
	q.Filters[query.ParameterFieldSelector] = "metadata.resourceVersion=1234567"

	b.Run("", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			list := v1alpha3.DefaultList(secretList, q, s.compare, s.filter)
			if list.TotalItems != 1 {
				b.Error("test failed")
			}
		}
	})

}
