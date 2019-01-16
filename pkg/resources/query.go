package resources

import (
	"regexp"
	"strings"

	"kubesphere.io/kubesphere/pkg/errors"
)

type conditions struct {
	match map[string]string
	fuzzy map[string]string
}

func ListResource(namespace, resource, conditionStr, orderBy string, reverse bool, limit, offset int) (*ResourceList, error) {

	items := make([]interface{}, 0)
	total := 0
	var err error

	conditions, err := parseToConditions(conditionStr)

	if err != nil {
		return nil, err
	}

	var result []interface{}

	switch resource {
	case "deployments":
		result, err = deploymentQuery(namespace, conditions, orderBy, reverse)
	case "statefulsets":
		result, err = statefulSetQuery(namespace, conditions, orderBy, reverse)
	case "daemonsets":
		result, err = daemonSetQuery(namespace, conditions, orderBy, reverse)
	case "jobs":
		result, err = jobQuery(namespace, conditions, orderBy, reverse)
	case "cronjobs":
		result, err = cronJobQuery(namespace, conditions, orderBy, reverse)
	case "services":
		result, err = serviceQuery(namespace, conditions, orderBy, reverse)
	case "ingresses":
		result, err = ingressQuery(namespace, conditions, orderBy, reverse)
	case "persistentvolumeclaims":
		result, err = persistentVolumeClaimQuery(namespace, conditions, orderBy, reverse)
	case "secrets":
		result, err = secretQuery(namespace, conditions, orderBy, reverse)
	case "configmaps":
		result, err = configMapQuery(namespace, conditions, orderBy, reverse)
	default:
		return nil, errors.New(errors.NotFound, "not found")
	}

	if err != nil {
		return nil, errors.New(errors.Internal, err.Error())
	}

	total = len(result)

	for i, d := range result {
		if i >= offset && len(items) < limit {
			items = append(items, d)
		}
	}

	return &ResourceList{TotalCount: total, Items: items}, nil
}

func parseToConditions(str string) (*conditions, error) {
	conditions := &conditions{match: make(map[string]string, 0), fuzzy: make(map[string]string, 0)}

	if str == "" {
		return conditions, nil
	}

	for _, item := range strings.Split(str, ",") {
		if strings.Count(item, "=") > 1 || strings.Count(item, "~") > 1 {
			return nil, errors.New(errors.InvalidArgument, "invalid condition input")
		}
		if groups := regexp.MustCompile(`(\S+)([=~])(\S+)`).FindStringSubmatch(item); len(groups) == 4 {
			if groups[2] == "=" {
				for _, s := range strings.Split(groups[1], "|") {
					conditions.match[s] = groups[3]
				}
			} else {
				for _, s := range strings.Split(groups[1], "|") {
					conditions.fuzzy[s] = groups[3]
				}
			}
		} else {
			return nil, errors.New(errors.InvalidArgument, "invalid condition input")
		}
	}
	return conditions, nil
}

type ResourceList struct {
	TotalCount int           `json:"total_count"`
	Items      []interface{} `json:"items"`
}

func labelSearch(m map[string]string, value string) bool {
	for k, v := range m {
		if strings.Contains(k, value) {
			return true
		} else if strings.Contains(v, value) {
			return true
		}
	}
	return false
}
