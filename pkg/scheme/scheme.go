/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package scheme

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	kubespherescheme "kubesphere.io/client-go/kubesphere/scheme"
)

// Scheme contains all types of custom Scheme and kubernetes client-go Scheme.
var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme)

func init() {
	// register common meta types into schemas.
	metav1.AddToGroupVersion(Scheme, metav1.SchemeGroupVersion)

	_ = clientgoscheme.AddToScheme(Scheme)
	_ = apiextensionsv1.AddToScheme(Scheme)
	_ = kubespherescheme.AddToScheme(Scheme)
}
