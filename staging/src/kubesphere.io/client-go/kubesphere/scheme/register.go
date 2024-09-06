package scheme

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	appv2 "kubesphere.io/api/application/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"
	gatewayv1alpha2 "kubesphere.io/api/gateway/v1alpha2"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	quotav1alpha2 "kubesphere.io/api/quota/v1alpha2"
	storagev1alpha1 "kubesphere.io/api/storage/v1alpha1"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
	appv1beta1 "sigs.k8s.io/application/api/v1beta1"
)

var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme)
var ParameterCodec = runtime.NewParameterCodec(Scheme)
var localSchemeBuilder = runtime.SchemeBuilder{
	appv2.AddToScheme,
	clusterv1alpha1.AddToScheme,
	corev1alpha1.AddToScheme,
	extensionsv1alpha1.AddToScheme,
	iamv1beta1.AddToScheme,
	quotav1alpha2.AddToScheme,
	storagev1alpha1.AddToScheme,
	tenantv1beta1.AddToScheme,
	gatewayv1alpha2.AddToScheme,
	appv1beta1.AddToScheme,
}

var AddToScheme = localSchemeBuilder.AddToScheme

func init() {
	v1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})
	utilruntime.Must(AddToScheme(Scheme))
	utilruntime.Must(k8sscheme.AddToScheme(Scheme))
}
