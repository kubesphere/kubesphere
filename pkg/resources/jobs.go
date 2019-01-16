package resources

import (
	"sort"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func jobQuery(namespace string, conditions *conditions, orderBy string, reverse bool) ([]interface{}, error) {
	jobs, err := JobLister.Jobs(namespace).List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*batchv1.Job, 0)

	if len(conditions.match) == 0 && len(conditions.fuzzy) == 0 {
		result = jobs
	} else {
		for _, item := range jobs {
			status := ""
			updateTime := item.CreationTimestamp.Time
			for _, condition := range item.Status.Conditions {
				if condition.Type == "Failed" && condition.Status == "True" {
					status = "failed"
				}
				if condition.Type == "Complete" && condition.Status == "True" {
					status = "complete"
				}

				if updateTime.Before(condition.LastProbeTime.Time) {
					updateTime = condition.LastProbeTime.Time
				}

				if updateTime.Before(condition.LastTransitionTime.Time) {
					updateTime = condition.LastTransitionTime.Time
				}
			}
			for k, v := range conditions.match {
				if k == "status" {
					switch v {
					case "running":
						if *item.Spec.Completions > item.Status.Succeeded && status == "" {
							result = append(result, item)
						}
					case "failed":
						if status == "failed" {
							result = append(result, item)
						}
					case "complete":
						if status == "complete" {
							result = append(result, item)
						}
					}
				}
				if k == "updateTime" {

				}
			}
			for k, v := range conditions.fuzzy {
				if k == "name" && strings.Contains(item.Name, v) {
					result = append(result, item)
				}
				if k == "label" && labelSearch(item.Labels, v) {
					result = append(result, item)
				}
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if reverse {
			tmp := i
			i = j
			j = tmp
		}
		switch orderBy {
		case "name":
			fallthrough
		default:
			return strings.Compare(result[i].Name, result[j].Name) <= 0
		}
	})

	r := make([]interface{}, 0)
	for _, i := range result {
		r = append(r, i)
	}
	return r, nil
}
