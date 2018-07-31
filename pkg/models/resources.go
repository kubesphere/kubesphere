package models

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/models/controllers"
	"kubesphere.io/kubesphere/pkg/options"
)

const (
	limit = "limit"
	page  = "page"
)

type ResourceList struct {
	Total int         `json:"total,omitempty"`
	Page  int         `json:"page,omitempty"`
	Limit int         `json:"limit,omitempty"`
	Items interface{} `json:"items,omitempty"`
}

func getController(resource string) (controllers.Controller, error) {
	switch resource {
	case controllers.Deployments, controllers.Statefulsets, controllers.Daemonsets, controllers.Ingresses,
		controllers.PersistentVolumeClaim, controllers.Roles, controllers.ClusterRoles, controllers.Services,
		controllers.Pods, controllers.Namespaces, controllers.StorageClasses:

		return controllers.ResourceControllers.Controllers[resource], nil
	default:
		return nil, fmt.Errorf("invalid resource Name '%s'", resource)
	}
	return nil, nil

}

func getConditions(str string) (map[string]string, map[string]string, error) {
	match := make(map[string]string)
	fuzzy := make(map[string]string)
	if len(str) == 0 {
		return nil, nil, nil
	}
	list := strings.Split(str, ",")
	for _, item := range list {
		if strings.Count(item, "=") >= 2 {
			return nil, nil, errors.New("invalid condition input, invalid character \"=\"")
		}

		if strings.Count(item, "~") >= 2 {
			return nil, nil, errors.New("invalid condition input, invalid character \"~\"")
		}

		if strings.Count(item, "=") == 1 {
			kvs := strings.Split(item, "=")
			if len(kvs) < 2 || len(kvs[1]) == 0 {
				return nil, nil, errors.New("invalid condition input")
			}
			match[kvs[0]] = kvs[1]
			continue
		}

		if strings.Count(item, "~") == 1 {
			kvs := strings.Split(item, "~")
			if len(kvs) < 2 || len(kvs[1]) == 0 {
				return nil, nil, errors.New("invalid condition input")
			}
			fuzzy[kvs[0]] = kvs[1]
			continue
		}

		return nil, nil, errors.New("invalid condition input")

	}
	return match, fuzzy, nil
}

func getPaging(resourceName, pagingStr string) (*controllers.Paging, map[string]int, error) {
	defaultPaging := &controllers.Paging{Limit: 10, Offset: 0}
	defautlPagingMap := map[string]int{"page": 1, "limit": 10}
	if resourceName == controllers.Namespaces {
		defaultPaging = nil
		defautlPagingMap = map[string]int{"page": 0, "limit": 0}
	}
	pagingMap := make(map[string]int)

	if len(pagingStr) == 0 {
		return defaultPaging, defautlPagingMap, nil
	}

	list := strings.Split(pagingStr, ",")
	for _, item := range list {
		kvs := strings.Split(item, "=")
		if len(kvs) < 2 {
			return nil, nil, errors.New("invalid Paging input")
		}

		value, err := strconv.Atoi(kvs[1])
		if err != nil {
			return nil, nil, errors.New("invalid Paging input")
		}

		pagingMap[kvs[0]] = value
	}

	if pagingMap[limit] <= 0 || pagingMap[page] <= 0 {
		return nil, nil, errors.New("invalid Paging input")
	}

	if pagingMap[limit] > 0 && pagingMap[page] > 0 {
		offset := (pagingMap[page] - 1) * pagingMap[limit]
		return &controllers.Paging{Limit: pagingMap[limit], Offset: offset}, pagingMap, nil
	}

	return defaultPaging, defautlPagingMap, nil
}

func ListResource(resourceName, conditonSrt, pagingStr string) (*ResourceList, error) {
	match, fuzzy, err := getConditions(conditonSrt)
	if err != nil {
		return nil, err
	}

	paging, pagingMap, err := getPaging(resourceName, pagingStr)
	if err != nil {
		return nil, err
	}

	conditionStr := generateConditionStr(match, fuzzy)

	ctl, err := getController(resourceName)
	if err != nil {
		return nil, err
	}

	total, items, err := ctl.ListWithConditions(conditionStr, paging)
	if err != nil {
		return nil, err
	}

	return &ResourceList{Total: total, Items: items, Page: pagingMap[page], Limit: pagingMap[limit]}, nil
}

func generateConditionStr(match map[string]string, fuzzy map[string]string) string {
	conditionStr := ""

	for k, v := range match {
		if len(conditionStr) == 0 {
			conditionStr = fmt.Sprintf("%s = \"%s\" ", k, v)
		} else {
			conditionStr = fmt.Sprintf("%s AND %s = \"%s\" ", conditionStr, k, v)
		}
	}

	for k, v := range fuzzy {
		if len(conditionStr) == 0 {
			conditionStr = fmt.Sprintf("%s like '%%%s%%' ", k, v)
		} else {
			conditionStr = fmt.Sprintf("%s AND %s like '%%%s%%' ", conditionStr, k, v)
		}
	}

	return conditionStr
}

type workLoadStatus struct {
	NameSpace string                 `json:"namespace"`
	Count     map[string]int         `json:"data"`
	Items     map[string]interface{} `json:"items,omitempty"`
}

func GetNamespacesResourceStatus(namespace string) (*workLoadStatus, error) {
	res := workLoadStatus{Count: make(map[string]int), NameSpace: namespace, Items: make(map[string]interface{})}
	var status *ResourceList
	var err error
	for _, resource := range []string{controllers.Deployments, controllers.Statefulsets, controllers.Daemonsets, controllers.PersistentVolumeClaim} {
		notReadyStatus := controllers.Updating
		if resource == controllers.PersistentVolumeClaim {
			notReadyStatus = controllers.PvcPending
		}
		if len(namespace) > 0 {
			status, err = ListResource(resource, fmt.Sprintf("status=%s,namespace=%s", notReadyStatus, namespace), "")
		} else {
			status, err = ListResource(resource, fmt.Sprintf("status=%s", notReadyStatus), "")
		}

		if err != nil {
			return nil, err
		}

		count := status.Total
		res.Count[resource] = count
	}

	return &res, nil
}

func GetClusterResourceStatus() (*workLoadStatus, error) {

	return GetNamespacesResourceStatus("")
}

func GetApplication(clusterId string) (interface{}, error) {
	ctl := &controllers.ApplicationCtl{OpenpitrixAddr: options.ServerOptions.GetOpAddress()}
	return ctl.GetApp(clusterId)
}

func ListApplication(runtimeId, conditions, pagingStr string) (*ResourceList, error) {
	paging, pagingMap, err := getPaging(controllers.Applications, pagingStr)
	if err != nil {
		return nil, err
	}

	match, fuzzy, err := getConditions(conditions)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	ctl := &controllers.ApplicationCtl{OpenpitrixAddr: options.ServerOptions.GetOpAddress()}
	total, items, err := ctl.ListApplication(runtimeId, match, fuzzy, paging)

	if err != nil {
		glog.Errorf("get application list failed, reason: %s", err)
		return nil, err
	}

	return &ResourceList{Total: total, Items: items, Page: pagingMap[page], Limit: pagingMap[limit]}, nil
}
