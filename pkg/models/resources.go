/*
Copyright 2018 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

type searchConditions struct {
	match   map[string]string
	fuzzy   map[string]string
	matchOr map[string]string
	fuzzyOr map[string]string
}

func getController(resource string) (controllers.Controller, error) {
	switch resource {
	case controllers.Deployments, controllers.Statefulsets, controllers.Daemonsets, controllers.Ingresses,
		controllers.PersistentVolumeClaim, controllers.Roles, controllers.ClusterRoles, controllers.Services,
		controllers.Pods, controllers.Namespaces, controllers.StorageClasses, controllers.Jobs, controllers.Cronjobs,
		controllers.Nodes, controllers.Secrets, controllers.ConfigMaps:

		return controllers.ResourceControllers.Controllers[resource], nil
	default:
		return nil, fmt.Errorf("invalid resource Name '%s'", resource)
	}
	return nil, nil

}

func getConditions(str string) (*searchConditions, map[string]string, error) {
	match := make(map[string]string)
	fuzzy := make(map[string]string)
	matchOr := make(map[string]string)
	fuzzyOr := make(map[string]string)
	orderField := make(map[string]string)

	if len(str) == 0 {
		return nil, nil, nil
	}

	conditions := strings.Split(str, ",")
	for _, item := range conditions {
		if strings.Count(item, "=") >= 2 || strings.Count(item, "~") >= 2 {
			return nil, nil, errors.New("invalid condition input")
		}

		if strings.Count(item, "=") == 1 {
			kvs := strings.Split(item, "=")
			if len(kvs) < 2 || len(kvs[1]) == 0 {
				return nil, nil, errors.New("invalid condition input")
			}

			if !strings.Contains(kvs[0], "|") {
				match[kvs[0]] = kvs[1]
			} else {
				multiFields := strings.Split(kvs[0], "|")
				for _, filed := range multiFields {
					if len(filed) > 0 {
						matchOr[filed] = kvs[1]
					}
				}
			}
			continue
		}

		if strings.Count(item, "~") == 1 {
			kvs := strings.Split(item, "~")
			if len(kvs) < 2 || len(kvs[1]) == 0 {
				return nil, nil, errors.New("invalid condition input")
			}
			if !strings.Contains(kvs[0], "|") {
				fuzzy[kvs[0]] = kvs[1]
			} else {
				multiFields := strings.Split(kvs[0], "|")
				if len(multiFields) > 1 && len(multiFields[1]) > 0 {
					orderField[multiFields[0]] = kvs[1]
				}
				for _, filed := range multiFields {
					if len(filed) > 0 {
						fuzzyOr[filed] = kvs[1]
					}
				}
			}
			continue
		}

		return nil, nil, errors.New("invalid condition input")
	}

	return &searchConditions{match: match, fuzzyOr: fuzzyOr, matchOr: matchOr, fuzzy: fuzzy}, orderField, nil
}

func getPaging(resourceName, pagingStr string) (*controllers.Paging, error) {
	defaultPaging := &controllers.Paging{Limit: 10, Offset: 0, Page: 1}
	paging := controllers.Paging{}

	if resourceName == controllers.Namespaces {
		defaultPaging = nil
	}

	if len(pagingStr) == 0 {
		return defaultPaging, nil
	}

	list := strings.Split(pagingStr, ",")
	for _, item := range list {
		kvs := strings.Split(item, "=")
		if len(kvs) < 2 {
			return nil, errors.New("invalid Paging input")
		}

		value, err := strconv.Atoi(kvs[1])
		if err != nil || value <= 0 {
			return nil, errors.New("invalid Paging input")
		}

		if kvs[0] == limit {
			paging.Limit = value
		}

		if kvs[0] == page {
			paging.Page = value
		}
	}

	if paging.Limit > 0 && paging.Page > 0 {
		paging.Offset = (paging.Page - 1) * paging.Limit
		return &paging, nil
	}

	return defaultPaging, nil
}

func generateOrder(orderField map[string]string, order string) string {
	if len(orderField) == 0 {
		return order
	}

	var str string
	for k, v := range orderField {
		if len(str) > 0 {
			str = fmt.Sprintf("%s, (%s like '%%%s%%')", str, k, v)
		} else {
			str = fmt.Sprintf("(%s like '%%%s%%')", k, v)
		}

	}

	if len(order) == 0 {
		return fmt.Sprintf("%s desc", str)
	} else {
		return fmt.Sprintf("%s, %s", str, order)
	}
}

func ListResource(resourceName, conditonSrt, pagingStr, order string) (*ResourceList, error) {
	conditions, OrderFields, err := getConditions(conditonSrt)
	if err != nil {
		return nil, err
	}

	order = generateOrder(OrderFields, order)
	conditionStr := generateConditionStr(conditions)

	paging, err := getPaging(resourceName, pagingStr)
	if err != nil {
		return nil, err
	}

	ctl, err := getController(resourceName)
	if err != nil {
		return nil, err
	}

	total, items, err := ctl.ListWithConditions(conditionStr, paging, order)
	if err != nil {
		return nil, err
	}

	if paging != nil {
		return &ResourceList{Total: total, Items: items, Page: paging.Page, Limit: paging.Limit}, nil
	} else {
		return &ResourceList{Total: total, Items: items}, nil
	}
}

func generateConditionStr(conditions *searchConditions) string {
	conditionStr := ""

	if conditions == nil {
		return conditionStr
	}

	for k, v := range conditions.match {
		if len(conditionStr) == 0 {
			conditionStr = fmt.Sprintf("%s = \"%s\" ", k, v)
		} else {
			conditionStr = fmt.Sprintf("%s AND %s = \"%s\" ", conditionStr, k, v)
		}
	}

	for k, v := range conditions.fuzzy {
		if len(conditionStr) == 0 {
			conditionStr = fmt.Sprintf("%s like '%%%s%%' ", k, v)
		} else {
			conditionStr = fmt.Sprintf("%s AND %s like '%%%s%%' ", conditionStr, k, v)
		}
	}

	for k, v := range conditions.matchOr {
		if len(conditionStr) == 0 {
			conditionStr = fmt.Sprintf("%s = \"%s\" ", k, v)
		} else {
			conditionStr = fmt.Sprintf("%s OR %s = \"%s\" ", conditionStr, k, v)
		}
	}

	for k, v := range conditions.fuzzyOr {
		if len(conditionStr) == 0 {
			conditionStr = fmt.Sprintf("%s like '%%%s%%' ", k, v)
		} else {
			conditionStr = fmt.Sprintf("%s OR %s like '%%%s%%' ", conditionStr, k, v)
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
			status, err = ListResource(resource, fmt.Sprintf("status=%s,namespace=%s", notReadyStatus, namespace), "", "")
		} else {
			status, err = ListResource(resource, fmt.Sprintf("status=%s", notReadyStatus), "", "")
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

func ListApplication(runtimeId, conditionStr, pagingStr string) (*ResourceList, error) {
	paging, err := getPaging(controllers.Applications, pagingStr)
	if err != nil {
		return nil, err
	}

	conditions, _, err := getConditions(conditionStr)
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	if conditions == nil {
		conditions = &searchConditions{}
	}

	ctl := &controllers.ApplicationCtl{OpenpitrixAddr: options.ServerOptions.GetOpAddress()}

	total, items, err := ctl.ListApplication(runtimeId, conditions.match, conditions.fuzzy, paging)

	if err != nil {
		glog.Errorf("get application list failed, reason: %s", err)
		return nil, err
	}

	return &ResourceList{Total: total, Items: items, Page: paging.Page, Limit: paging.Limit}, nil
}
