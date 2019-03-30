package destinationrule

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO(jeff): add test cases

var namespace = "default"
var lbs = map[string]string{
	"app.kubernetes.io/name":            "bookinfo",
	"servicemesh.kubesphere.io/enabled": "",
	"app":                               "reviews",
}

var service = corev1.Service{}

var deployments = []appsv1.Deployment{
	{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "deploy-v1",
			Labels: map[string]string{
				"app.kubernetes.io/name":            "bookinfo",
				"servicemesh.kubesphere.io/enabled": "",
				"app":                               "reviews",
				"version":                           "v1",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name":            "bookinfo",
					"servicemesh.kubesphere.io/enabled": "",
					"app":                               "reviews",
					"version":                           "v1",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name":            "bookinfo",
						"servicemesh.kubesphere.io/enabled": "",
						"app":                               "reviews",
						"version":                           "v1",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{},
					},
				},
			},
		},
	},
}
