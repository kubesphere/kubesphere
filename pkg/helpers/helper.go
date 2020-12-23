package helpers

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func MapAsLabelSelector(m map[string]string) labels.Selector {
	ls := metav1.LabelSelector{
		MatchLabels: m,
	}
	selector, _ := metav1.LabelSelectorAsSelector(&ls)
	return selector
}
