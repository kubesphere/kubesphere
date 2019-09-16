/*
 *
 * Copyright 2019 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */
package resources

import (
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/labels"
)

type hpaSearcher struct {
}

func (*hpaSearcher) get(namespace, name string) (interface{}, error) {
	return informers.SharedInformerFactory().Autoscaling().V2beta2().HorizontalPodAutoscalers().Lister().HorizontalPodAutoscalers(namespace).Get(name)
}

func hpaTargetMatch(item *autoscalingv2beta2.HorizontalPodAutoscaler, kind, name string) bool {
	return item.Spec.ScaleTargetRef.Kind == kind && item.Spec.ScaleTargetRef.Name == name
}

// exactly Match
func (*hpaSearcher) match(match map[string]string, item *autoscalingv2beta2.HorizontalPodAutoscaler) bool {
	for k, v := range match {
		switch k {
		case TargetKind:
			fallthrough
		case TargetName:
			kind := match[TargetKind]
			name := match[TargetName]
			if !hpaTargetMatch(item, kind, name) {
				return false
			}
		case Name:
			names := strings.Split(v, "|")
			if !sliceutil.HasString(names, item.Name) {
				return false
			}
		case Keyword:
			if !strings.Contains(item.Name, v) && !searchFuzzy(item.Labels, "", v) && !searchFuzzy(item.Annotations, "", v) {
				return false
			}
		default:
			if item.Labels[k] != v {
				return false
			}
		}
	}
	return true
}

// Fuzzy searchInNamespace
func (*hpaSearcher) fuzzy(fuzzy map[string]string, item *autoscalingv2beta2.HorizontalPodAutoscaler) bool {
	for k, v := range fuzzy {
		switch k {
		case Name:
			if !strings.Contains(item.Name, v) && !strings.Contains(item.Annotations[constants.DisplayNameAnnotationKey], v) {
				return false
			}
		case Label:
			if !searchFuzzy(item.Labels, "", v) {
				return false
			}
		case annotation:
			if !searchFuzzy(item.Annotations, "", v) {
				return false
			}
			return false
		case app:
			if !strings.Contains(item.Labels[chart], v) && !strings.Contains(item.Labels[release], v) {
				return false
			}
		default:
			if !searchFuzzy(item.Labels, k, v) && !searchFuzzy(item.Annotations, k, v) {
				return false
			}
		}
	}
	return true
}

func (*hpaSearcher) compare(a, b *autoscalingv2beta2.HorizontalPodAutoscaler, orderBy string) bool {
	switch orderBy {
	case CreateTime:
		return a.CreationTimestamp.Time.Before(b.CreationTimestamp.Time)
	case Name:
		fallthrough
	default:
		return strings.Compare(a.Name, b.Name) <= 0
	}
}

func (s *hpaSearcher) search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error) {

	horizontalPodAutoscalers, err := informers.SharedInformerFactory().Autoscaling().V2beta2().HorizontalPodAutoscalers().Lister().HorizontalPodAutoscalers(namespace).List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*autoscalingv2beta2.HorizontalPodAutoscaler, 0)

	if len(conditions.Match) == 0 && len(conditions.Fuzzy) == 0 {
		result = horizontalPodAutoscalers
	} else {
		for _, item := range horizontalPodAutoscalers {
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
