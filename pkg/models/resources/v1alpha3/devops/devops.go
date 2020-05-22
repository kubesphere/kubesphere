package devops

import (
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/kubesphere/pkg/api"
	devopsv1alpha3 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

type devopsGetter struct {
	informers ksinformers.SharedInformerFactory
}

func New(ksinformer ksinformers.SharedInformerFactory) v1alpha3.Interface {
	return &devopsGetter{informers: ksinformer}
}

func (n devopsGetter) Get(_, name string) (runtime.Object, error) {
	return n.informers.Devops().V1alpha3().DevOpsProjects().Lister().Get(name)
}

func (n devopsGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	projects, err := n.informers.Devops().V1alpha3().DevOpsProjects().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, project := range projects {
		result = append(result, project)
	}

	return v1alpha3.DefaultList(result, query, n.compare, n.filter), nil
}

func (n devopsGetter) filter(item runtime.Object, filter query.Filter) bool {
	devOpsProject, ok := item.(*devopsv1alpha3.DevOpsProject)
	if !ok {
		return false
	}
	return v1alpha3.DefaultObjectMetaFilter(devOpsProject.ObjectMeta, filter)
}

func (n devopsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftProject, ok := left.(*devopsv1alpha3.DevOpsProject)
	if !ok {
		return false
	}

	rightProject, ok := right.(*devopsv1alpha3.DevOpsProject)
	if !ok {
		return true
	}
	return v1alpha3.DefaultObjectMetaCompare(leftProject.ObjectMeta, rightProject.ObjectMeta, field)
}
