/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"
)

func TestServiceAddUpdateApiService(t *testing.T) {
	uri := "http://172.31.188.161:8080"
	apiServer := extensionsv1alpha1.APIService{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "v1alpha1.local.kubesphere.io",
		},
		Spec: extensionsv1alpha1.APIServiceSpec{
			Group:   "local.kubesphere.io",
			Version: "v1alpha1",
			Endpoint: extensionsv1alpha1.Endpoint{
				URL:                &uri,
				Service:            nil,
				CABundle:           nil,
				InsecureSkipVerify: false,
			},
		},
		Status: extensionsv1alpha1.APIServiceStatus{},
	}

	openApiV2Services := NewOpenApiV2Services()
	err := openApiV2Services.AddUpdateApiService(&apiServer)
	assert.Equal(t, err, nil)
	val, err := openApiV2Services.MergeSpecCache()
	assert.Equal(t, err, nil)
	t.Log(val)
}
