package models

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/models/controllers"
)

type ResourceList struct {
	Total int         `json:"total,omitempty"`
	Page  int         `json:"page,omitempty"`
	Limit int         `json:"limit,omitempty"`
	Items interface{} `json:"items,omitempty"`
}

func getController(resource string) (controllers.Controller, error) {
	var ctl controllers.Controller
	attr := controllers.CommonAttribute{DB: client.NewDBClient()}
	switch resource {
	case controllers.Deployments:
		ctl = &controllers.DeploymentCtl{attr}
	case controllers.Statefulsets:
		ctl = &controllers.StatefulsetCtl{attr}
	case controllers.Daemonsets:
		ctl = &controllers.DaemonsetCtl{attr}
	case controllers.Ingresses:
		ctl = &controllers.IngressCtl{attr}
	case controllers.PersistentVolumeClaim:
		ctl = &controllers.PvcCtl{attr}
	case controllers.Roles:
		ctl = &controllers.RoleCtl{attr}
	case controllers.ClusterRoles:
		ctl = &controllers.ClusterRoleCtl{attr}
	case controllers.Services:
		ctl = &controllers.ServiceCtl{attr}
	case controllers.Pods:
		ctl = &controllers.PodCtl{attr}
	case controllers.Namespaces:
		ctl = &controllers.NamespaceCtl{attr}
	case controllers.StorageClasses:
		ctl = &controllers.StorageClassCtl{attr}
	default:
		return nil, errors.New("invalid resource type")
	}
	return ctl, nil

}

func getConditions(str string) (map[string]string, error) {
	dict := make(map[string]string)
	if len(str) == 0 {
		return dict, nil
	}
	list := strings.Split(str, ",")
	for _, item := range list {
		kvs := strings.Split(item, "=")
		if len(kvs) < 2 {
			return nil, errors.New("invalid condition input")
		}
		dict[kvs[0]] = kvs[1]
	}
	return dict, nil
}

func getPaging(str string) (map[string]int, error) {
	paging := make(map[string]int)
	if len(str) == 0 {
		return paging, nil
	}
	list := strings.Split(str, ",")
	for _, item := range list {
		kvs := strings.Split(item, "=")
		if len(kvs) < 2 {
			return nil, errors.New("invalid Paging input")
		}

		value, err := strconv.Atoi(kvs[1])
		if err != nil {
			return nil, err
		}

		paging[kvs[0]] = value
	}
	return paging, nil
}

func ListResource(resourceName, conditonSrt, pagingStr string) (*ResourceList, error) {
	conditions, err := getConditions(conditonSrt)
	if err != nil {
		return nil, err
	}

	pagingMap, err := getPaging(pagingStr)
	if err != nil {
		return nil, err
	}

	conditionStr, paging := generateConditionAndPaging(conditions, pagingMap)

	ctl, err := getController(resourceName)
	if err != nil {
		return nil, err
	}

	total, items, err := ctl.ListWithConditions(conditionStr, paging)
	if err != nil {
		return nil, err
	}

	return &ResourceList{Total: total, Items: items, Page: pagingMap["page"], Limit: pagingMap["limit"]}, nil
}

func generateConditionAndPaging(conditions map[string]string, paging map[string]int) (string, *controllers.Paging) {
	conditionStr := ""

	for k, v := range conditions {
		if len(conditionStr) == 0 {
			conditionStr = fmt.Sprintf("%s = \"%s\" ", k, v)
		} else {
			conditionStr = fmt.Sprintf("%s AND %s = \"%s\" ", conditionStr, k, v)
		}
	}

	if paging["limit"] > 0 && paging["page"] >= 0 {
		offset := (paging["page"] - 1) * paging["limit"]
		return conditionStr, &controllers.Paging{Limit: paging["limit"], Offset: offset}
	}

	return conditionStr, nil
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
		resourceStatus := controllers.Updating
		if resource == controllers.PersistentVolumeClaim {
			resourceStatus = controllers.PvcPending
		}
		if len(namespace) > 0 {
			status, err = ListResource(resource, fmt.Sprintf("status=%s,namespace=%s", resourceStatus, namespace), "")
		} else {
			status, err = ListResource(resource, fmt.Sprintf("status=%s", resourceStatus), "")
		}

		if err != nil {
			return nil, err
		}

		count := status.Total
		//items := status.Items
		res.Count[resource] = count
		//res.Items[resource] = items
	}

	return &res, nil
}

func GetClusterResourceStatus() (*workLoadStatus, error) {

	return GetNamespacesResourceStatus("")
}
