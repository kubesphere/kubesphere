package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"kubesphere.io/client-go/client"
	"kubesphere.io/client-go/client/generic"
)

func TestDelayingDeliverer(t *testing.T) {

	config := &rest.Config{
		Host:     "http://127.0.0.1:9090",
		Username: "tester",
		Password: "P@88w0rd",
	}

	sch := runtime.NewScheme()
	v1.AddToScheme(sch)
	c, _ := generic.New(config, client.Options{Scheme: sch})

	// sar := &authv1.SelfSubjectRulesReview{
	// 	Spec: authv1.SelfSubjectRulesReviewSpec{
	// 		Namespace: "kube-system",
	// 	},
	// }
	// err := c.Create(context.TODO(), sar)

	pods := &v1.PodList{}

	err := c.List(context.TODO(), pods)

	assert.Nil(t, err)
}
