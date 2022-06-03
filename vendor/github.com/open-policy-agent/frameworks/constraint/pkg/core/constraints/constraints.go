package constraints

import (
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// SemanticEqual returns whether the specs of the constraints are equal. It
// ignores status and metadata because neither are relevant as to how a
// constraint is enforced. It is assumed that the author is comparing
// two constraints with the same GVK/namespace/name
func SemanticEqual(c1 *unstructured.Unstructured, c2 *unstructured.Unstructured) bool {
	s1, exists, err := unstructured.NestedMap(c1.Object, "spec")
	if err != nil || !exists {
		return false
	}
	s2, exists, err := unstructured.NestedMap(c2.Object, "spec")
	if err != nil || !exists {
		return false
	}
	return reflect.DeepEqual(s1, s2)
}
