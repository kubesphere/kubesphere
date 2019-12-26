package resources

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
	"knative.dev/serving/pkg/apis/serving/v1beta1"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/server/params"
)

type s8sServiceSearcher struct {
}

func (s *s8sServiceSearcher) get(namespace, name string) (interface{}, error) {
	return nil, nil
}

func (s *s8sServiceSearcher) match(match map[string]string, item *v1beta1.Service) bool {
	return true
}

func (s *s8sServiceSearcher) fuzzy(match map[string]string, item *v1beta1.Service) bool {
	return true
}

func (s *s8sServiceSearcher) compare(a, b *v1beta1.Service, orderBy string) bool {
	return true
}

func (s *s8sServiceSearcher) search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error) {
	// TODO: finish this
	result, err := informers.ServerlessInformerFactory().Serving().V1beta1().Services().Lister().Services(namespace).List(labels.Everything())

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	/*
	result := make([]*v1beta1.Service, 0)

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
