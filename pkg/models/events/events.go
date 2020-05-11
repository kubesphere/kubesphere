package events

import (
	corev1 "k8s.io/api/core/v1"
	eventsv1alpha1 "kubesphere.io/kubesphere/pkg/api/events/v1alpha1"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/simple/client/events"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"strings"
	"time"
)

type Interface interface {
	Events(queryParam *eventsv1alpha1.Query, MutateFilterFunc func(*events.Filter)) (*eventsv1alpha1.APIResponse, error)
}

type eventsOperator struct {
	client events.Client
}

func NewEventsOperator(client events.Client) Interface {
	return &eventsOperator{client}
}

func (eo *eventsOperator) Events(queryParam *eventsv1alpha1.Query,
	MutateFilterFunc func(*events.Filter)) (*eventsv1alpha1.APIResponse, error) {
	filter := &events.Filter{
		InvolvedObjectNames:     stringutils.Split(queryParam.InvolvedObjectNameFilter, ","),
		InvolvedObjectNameFuzzy: stringutils.Split(queryParam.InvolvedObjectNameSearch, ","),
		InvolvedObjectkinds:     stringutils.Split(queryParam.InvolvedObjectKindFilter, ","),
		Reasons:                 stringutils.Split(queryParam.ReasonFilter, ","),
		ReasonFuzzy:             stringutils.Split(queryParam.ReasonSearch, ","),
		MessageFuzzy:            stringutils.Split(queryParam.MessageSearch, ","),
		Type:                    queryParam.TypeFilter,
		StartTime:               queryParam.StartTime,
		EndTime:                 queryParam.EndTime,
	}
	if MutateFilterFunc != nil {
		MutateFilterFunc(filter)
	}

	var ar eventsv1alpha1.APIResponse
	var err error
	switch queryParam.Operation {
	case "histogram":
		if len(filter.InvolvedObjectNamespaceMap) == 0 {
			ar.Histogram = &events.Histogram{}
		} else {
			ar.Histogram, err = eo.client.CountOverTime(filter, queryParam.Interval)
		}
	case "statistics":
		if len(filter.InvolvedObjectNamespaceMap) == 0 {
			ar.Statistics = &events.Statistics{}
		} else {
			ar.Statistics, err = eo.client.StatisticsOnResources(filter)
		}
	default:
		if len(filter.InvolvedObjectNamespaceMap) == 0 {
			ar.Events = &events.Events{}
		} else {
			ar.Events, err = eo.client.SearchEvents(filter, queryParam.From, queryParam.Size, queryParam.Sort)
		}
	}
	if err != nil {
		return nil, err
	}
	return &ar, nil
}

func IntersectedNamespaces(queryParam *eventsv1alpha1.Query, workspaces []*tenantv1alpha1.Workspace,
	namespaces []*corev1.Namespace, canListAll bool) map[string]time.Time {
	var (
		wsFilterSet = stringSet(stringutils.Split(queryParam.WorkspaceFilter, ","))
		wsSearchArr = stringutils.Split(queryParam.WorkspaceSearch, ",")
		nsFilterSet = stringSet(stringutils.Split(queryParam.InvolvedObjectNamespaceFilter, ","))
		nsSearchArr = stringutils.Split(queryParam.InvolvedObjectNamespaceSearch, ",")

		wsMap             = make(map[string]*tenantv1alpha1.Workspace)
		nsCreationTimeMap = make(map[string]time.Time)

		showEvtsInNsWithoutWs = len(wsFilterSet) == 0 && len(wsSearchArr) == 0 && canListAll
		showEvtsWithoutNs     = len(wsFilterSet) == 0 && len(wsSearchArr) == 0 &&
			len(nsFilterSet) == 0 && len(nsSearchArr) == 0 && canListAll
	)

	if showEvtsWithoutNs {
		nsCreationTimeMap[""] = time.Time{}
	}

	for _, ws := range workspaces {
		if len(wsFilterSet) > 0 {
			if _, ok := wsFilterSet[ws.Name]; !ok {
				continue
			}
		}
		if len(wsSearchArr) > 0 && !stringContains(ws.Name, wsSearchArr) {
			continue
		}
		wsMap[ws.Name] = ws
	}
	for _, ns := range namespaces {
		if len(nsFilterSet) > 0 {
			if _, ok := nsFilterSet[ns.Name]; !ok {
				continue
			}
		}
		if len(nsSearchArr) > 0 && !stringContains(ns.Name, nsSearchArr) {
			continue
		}
		if ws, ok := ns.Labels[constants.WorkspaceLabelKey]; ok && ws != "" {
			if _, ok := wsMap[ws]; !ok {
				continue
			}
		} else if !showEvtsInNsWithoutWs {
			continue
		}
		nsCreationTimeMap[ns.Name] = ns.CreationTimestamp.Time
	}
	return nsCreationTimeMap
}

func stringSet(strs []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, str := range strs {
		m[str] = struct{}{}
	}
	return m
}

func stringContains(str string, subStrs []string) bool {
	for _, sub := range subStrs {
		if strings.Contains(str, sub) {
			return true
		}
	}
	return false
}
