package federateddeployment

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type fedreatedDeploymentGetter struct {
	informer informers.SharedInformerFactory
}

func New(informer informers.SharedInformerFactory) v1alpha3.Interface {
	return &fedreatedDeploymentGetter{
		informer: informer,
	}
}

func (f *fedreatedDeploymentGetter) Get(namespace, name string) (runtime.Object, error) {
	return f.informer.Types().V1beta1().FederatedDeployments().Lister().FederatedDeployments(namespace).Get(name)
}

func (f *fedreatedDeploymentGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	federatedDeployments, err := f.informer.Types().V1beta1().FederatedDeployments().Lister().FederatedDeployments(namespace).List(query.Selector())

	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, fedDeployment := range federatedDeployments {
		result = append(result, fedDeployment)
	}

	return v1alpha3.DefaultList(result, query, f.compare, f.filter), nil
}

func (f *fedreatedDeploymentGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftFedDeployment, ok := left.(*v1beta1.FederatedDeployment)
	if !ok {
		return false
	}

	rightFedDeployment, ok := right.(*v1beta1.FederatedDeployment)
	if !ok {
		return false
	}

	switch field {
	case query.FieldUpdateTime:
		fallthrough
	case query.FieldLastUpdateTimestamp:
		return lastUpdateTime(leftFedDeployment) > lastUpdateTime(rightFedDeployment)
	default:
		return v1alpha3.DefaultObjectMetaCompare(leftFedDeployment.ObjectMeta, rightFedDeployment.ObjectMeta, field)
	}
}

func (f *fedreatedDeploymentGetter) filter(object runtime.Object, filter query.Filter) bool {
	fedDeployment, ok := object.(*v1beta1.FederatedDeployment)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaFilter(fedDeployment.ObjectMeta, filter)

}

func lastUpdateTime(fedDeployment *v1beta1.FederatedDeployment) string {
	lut := fedDeployment.CreationTimestamp.Time.String()
	for _, condition := range fedDeployment.Status.Conditions {
		if condition.LastUpdateTime > lut {
			lut = condition.LastUpdateTime
		}
	}
	return lut
}
