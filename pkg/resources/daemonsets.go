package resources

import (
	"sort"
	"strings"

	"k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func daemonSetQuery(namespace string, conditions *conditions, orderBy string, reverse bool) ([]interface{}, error) {
	daemonsets, err := DaemonSetLister.DaemonSets(namespace).List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*v1.DaemonSet, 0)

	if len(conditions.match) == 0 && len(conditions.fuzzy) == 0 {
		result = daemonsets
	} else {
		for _, item := range daemonsets {
			for k, v := range conditions.match {
				if k == "status" {
					switch v {
					case "running":
						if item.Status.DesiredNumberScheduled == item.Status.NumberAvailable {
							result = append(result, item)
						}
					case "stopped":
					case "updating":
						if item.Status.DesiredNumberScheduled != item.Status.NumberAvailable {
							result = append(result, item)
						}
					}
				}
			}
			for k, v := range conditions.fuzzy {
				if k == "name" && strings.Contains(item.Name, v) {
					result = append(result, item)
				}
				if k == "app" && (strings.Contains(item.Labels["chart"], v) || strings.Contains(item.Labels["release"], v)) {
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
