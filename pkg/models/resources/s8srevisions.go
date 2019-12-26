package resources

import (
	"k8s.io/apimachinery/pkg/labels"
	"knative.dev/serving/pkg/apis/serving/v1beta1"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/server/params"
)

type revisionSearcher struct {
}

func (s *revisionSearcher) get(namespace, name string) (interface{}, error) {
	return nil, nil
}

func (s *revisionSearcher) match(match map[string]string, item *v1beta1.Revision) bool {
	return true
}

func (s *revisionSearcher) fuzzy(match map[string]string, item *v1beta1.Revision) bool {
	return true
}

func (s *revisionSearcher) compare(a, b *v1beta1.Revision, orderBy string) bool {
	return true
}

func (s *revisionSearcher) search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error) {
	result, err := informers.ServerlessInformerFactory().Serving().V1beta1().Revisions().Lister().Revisions(namespace).List(labels.Everything())

	if err != nil {
		return nil, err
	}

	/*
	result := make([]*v1beta1beta1.Service, 0)

	if len(conditions.Match) == 0 && len(conditions.Fuzzy) == 0 {
		result = services
	} else {
		for _, item := range services {
			if s.match(conditions.Match, item) && s.fuzzy(conditions.Fuzzy, item) {
				result = append(result, item)
			}
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if reverse {
			tmp := i
			i = j
			j = tmp
		}
		return s.compare(result[i], result[j], orderBy)
	})
	*/

	r := make([]interface{}, 0)
	for _, i := range result {
		r = append(r, i)
	}
	return r, nil
}
