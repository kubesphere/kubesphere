package resources

import (
	"k8s.io/apimachinery/pkg/labels"
	"knative.dev/serving/pkg/apis/serving/v1beta1"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/server/params"
	"sort"
)

type revisionSearcher struct {
}

func (s *revisionSearcher) get(namespace, name string) (interface{}, error) {
	// Not implemented. WIP
	return nil, nil
}

func (s *revisionSearcher) match(match map[string]string, item *v1beta1.Revision) bool {
	// Left empty to make search work. WIP
	return true
}

func (s *revisionSearcher) fuzzy(match map[string]string, item *v1beta1.Revision) bool {
	// Left empty to make search work. WIP
	return true
}

func (s *revisionSearcher) compare(a, b *v1beta1.Revision, orderBy string) bool {
	// Left empty to make search work. WIP
	return true
}

func (s *revisionSearcher) search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error) {
	revisions, err := informers.ServerlessInformerFactory().Serving().V1beta1().Revisions().Lister().Revisions(namespace).List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*v1beta1.Revision, 0)

	if len(conditions.Match) == 0 && len(conditions.Fuzzy) == 0 {
		result = revisions
	} else {
		for _, item := range revisions {
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

	r := make([]interface{}, 0)
	for _, i := range result {
		r = append(r, i)
	}
	return r, nil
}
