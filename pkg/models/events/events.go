package events

import (
	eventsv1alpha1 "kubesphere.io/kubesphere/pkg/api/events/v1alpha1"
	"kubesphere.io/kubesphere/pkg/simple/client/events"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
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
