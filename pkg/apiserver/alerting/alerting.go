/*

 Copyright 2019 The KubeSphere Authors.

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

package alerting

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models/alert"
	alclient "kubesphere.io/kubesphere/pkg/simple/client/alert"
	"kubesphere.io/kubesphere/pkg/simple/client/alert/pb"
	k8sclient "kubesphere.io/kubesphere/pkg/simple/client/kubernetes"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	grpcTimeoutSeconds = 60
)

func parseBool(input string) bool {
	if input == "true" {
		return true
	} else {
		return false
	}
}

func parseBools(inputs []string) []bool {
	vs := []bool{}

	for _, input := range inputs {
		if input == "true" {
			vs = append(vs, true)
		} else {
			vs = append(vs, false)
		}
	}

	return vs
}

func parseUint32(s string) (uint32, error) {
	if v, err := strconv.ParseUint(s, 10, 32); err != nil {
		return 0, err
	} else {
		return uint32(v), nil
	}
}

func parseUint32s(ss []string) []uint32 {
	vs := []uint32{}

	for _, s := range ss {
		if v, err := strconv.ParseUint(s, 10, 32); err != nil {
			continue
		} else {
			vs = append(vs, uint32(v))
		}
	}

	return vs
}

func getKV(kv map[string]string) (string, string) {
	key := ""
	value := ""
	for k, v := range kv {
		key = k
		value = v
	}

	return key, value
}

func parseLabelSelector(resourceSelector []map[string]string) string {
	labelSelector := ""
	for _, kv := range resourceSelector {
		k, v := getKV(kv)
		labelSelector = labelSelector + fmt.Sprintf("%s=%s,", k, v)
	}
	labelSelector = strings.TrimSuffix(labelSelector, ",")

	return labelSelector
}

func CreateResourceType(request *restful.Request, response *restful.Response) {
	resourceType := new(models.ResourceType)

	err := request.ReadEntity(&resourceType)
	if err != nil {
		glog.Errorf("CreateResourceType request data error %+v.", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.CreateResourceTypeRequest{
		RsTypeName:  resourceType.RsTypeName,
		RsTypeParam: resourceType.RsTypeParam,
	}

	resp, err := client.CreateResourceType(ctx, req)
	if err != nil {
		glog.Errorf("CreateResourceType failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("CreateResourceType success: %+v", resp)

	response.WriteAsJson(resp)
}

func DescribeResourceTypes(request *restful.Request, response *restful.Response) {
	rsTypeIds := strings.Split(request.QueryParameter("rs_type_ids"), ",")
	rsTypeNames := strings.Split(request.QueryParameter("rs_type_names"), ",")

	sortKey := request.QueryParameter("sort_key")
	reverse := parseBool(request.QueryParameter("reverse"))
	offset, _ := parseUint32(request.QueryParameter("offset"))
	limit, _ := parseUint32(request.QueryParameter("limit"))

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DescribeResourceTypesRequest{
		RsTypeId:   rsTypeIds,
		RsTypeName: rsTypeNames,
		SortKey:    sortKey,
		Reverse:    reverse,
		Offset:     offset,
		Limit:      limit,
	}

	resp, err := client.DescribeResourceTypes(ctx, req)
	if err != nil {
		glog.Errorf("DescribeResourceTypes failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("DescribeResourceTypes success: %+v", resp)

	response.WriteAsJson(resp)
}

func ModifyResourceType(request *restful.Request, response *restful.Response) {
	resourceType := new(models.ResourceType)

	err := request.ReadEntity(&resourceType)
	if err != nil {
		glog.Errorf("ModifyResourceType request data error %+v.", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.ModifyResourceTypeRequest{
		RsTypeId:    resourceType.RsTypeId,
		RsTypeName:  resourceType.RsTypeName,
		RsTypeParam: resourceType.RsTypeParam,
	}

	resp, err := client.ModifyResourceType(ctx, req)
	if err != nil {
		glog.Errorf("ModifyResourceType failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Errorf("ModifyResourceType success: %+v", resp)

	response.WriteAsJson(resp)
}

func DeleteResourceTypes(request *restful.Request, response *restful.Response) {
	rsTypeIds := strings.Split(request.QueryParameter("rs_type_ids"), ",")

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DeleteResourceTypesRequest{
		RsTypeId: rsTypeIds,
	}

	resp, err := client.DeleteResourceTypes(ctx, req)
	if err != nil {
		glog.Errorf("DeleteResourceTypes failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("DeleteResourceTypes success: %+v", resp)

	response.WriteAsJson(resp)
}

func CreateResourceFilter(request *restful.Request, response *restful.Response) {
	rsFilter := new(models.ResourceFilter)

	err := request.ReadEntity(&rsFilter)
	if err != nil {
		glog.Errorf("CreateResourceFilter request data error %+v.", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.CreateResourceFilterRequest{
		RsFilterName:  rsFilter.RsFilterName,
		RsFilterParam: rsFilter.RsFilterParam,
		Status:        rsFilter.Status,
		RsTypeId:      rsFilter.RsTypeId,
	}

	resp, err := client.CreateResourceFilter(ctx, req)
	if err != nil {
		glog.Errorf("CreateResourceFilter failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("CreateResourceFilter success: %+v", resp)

	response.WriteAsJson(resp)
}

func DescribeResourceFilters(request *restful.Request, response *restful.Response) {
	rsFilterIds := strings.Split(request.QueryParameter("rs_filter_ids"), ",")
	rsFilterNames := strings.Split(request.QueryParameter("rs_filter_names"), ",")
	status := strings.Split(request.QueryParameter("status"), ",")
	rsTypeIds := strings.Split(request.QueryParameter("rs_type_ids"), ",")

	sortKey := request.QueryParameter("sort_key")
	reverse := parseBool(request.QueryParameter("reverse"))
	offset, _ := parseUint32(request.QueryParameter("offset"))
	limit, _ := parseUint32(request.QueryParameter("limit"))

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DescribeResourceFiltersRequest{
		RsFilterId:   rsFilterIds,
		RsFilterName: rsFilterNames,
		Status:       status,
		RsTypeId:     rsTypeIds,
		SortKey:      sortKey,
		Reverse:      reverse,
		Offset:       offset,
		Limit:        limit,
	}

	resp, err := client.DescribeResourceFilters(ctx, req)
	if err != nil {
		glog.Errorf("DescribeResourceFilters failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("DescribeResourceFilters success: %+v", resp)

	response.WriteAsJson(resp)
}

func ModifyResourceFilter(request *restful.Request, response *restful.Response) {
	rsFilter := new(models.ResourceFilter)

	err := request.ReadEntity(&rsFilter)
	if err != nil {
		glog.Errorf("ModifyResourceFilter request data error %+v.", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.ModifyResourceFilterRequest{
		RsFilterId:    rsFilter.RsFilterId,
		RsFilterName:  rsFilter.RsFilterName,
		RsFilterParam: rsFilter.RsFilterParam,
		Status:        rsFilter.Status,
		RsTypeId:      rsFilter.RsTypeId,
	}

	resp, err := client.ModifyResourceFilter(ctx, req)
	if err != nil {
		glog.Errorf("ModifyResourceFilter failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("ModifyResourceFilter success: %+v", resp)

	response.WriteAsJson(resp)
}

func DeleteResourceFilters(request *restful.Request, response *restful.Response) {
	rsFilterIds := strings.Split(request.QueryParameter("rs_filter_ids"), ",")

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DeleteResourceFiltersRequest{
		RsFilterId: rsFilterIds,
	}

	resp, err := client.DeleteResourceFilters(ctx, req)
	if err != nil {
		glog.Errorf("DeleteResourceFilters failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("DeleteResourceFilters success: %+v", resp)

	response.WriteAsJson(resp)
}

func CreateMetric(request *restful.Request, response *restful.Response) {
	metric := new(models.Metric)

	err := request.ReadEntity(&metric)
	if err != nil {
		glog.Errorf("CreateMetric request data error %+v.", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.CreateMetricRequest{
		MetricName:  metric.MetricName,
		MetricParam: metric.MetricParam,
		Status:      metric.Status,
		RsTypeId:    metric.RsTypeId,
	}

	resp, err := client.CreateMetric(ctx, req)
	if err != nil {
		glog.Errorf("CreateMetric failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("CreateMetric success: %+v", resp)

	response.WriteAsJson(resp)
}

func DescribeMetrics(request *restful.Request, response *restful.Response) {
	metricIds := strings.Split(request.QueryParameter("metric_ids"), ",")
	metricNames := strings.Split(request.QueryParameter("metric_names"), ",")
	status := strings.Split(request.QueryParameter("status"), ",")
	rsTypeIds := strings.Split(request.QueryParameter("rs_type_ids"), ",")

	sortKey := request.QueryParameter("sort_key")
	reverse := parseBool(request.QueryParameter("reverse"))
	offset, _ := parseUint32(request.QueryParameter("offset"))
	limit, _ := parseUint32(request.QueryParameter("limit"))

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DescribeMetricsRequest{
		MetricId:   metricIds,
		MetricName: metricNames,
		Status:     status,
		RsTypeId:   rsTypeIds,
		SortKey:    sortKey,
		Reverse:    reverse,
		Offset:     offset,
		Limit:      limit,
	}

	resp, err := client.DescribeMetrics(ctx, req)
	if err != nil {
		glog.Errorf("DescribeMetrics failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("DescribeMetrics success: %+v", resp)

	response.WriteAsJson(resp)
}

func ModifyMetric(request *restful.Request, response *restful.Response) {
	metric := new(models.Metric)

	err := request.ReadEntity(&metric)
	if err != nil {
		glog.Errorf("ModifyMetric request data error %+v.", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.ModifyMetricRequest{
		MetricId:    metric.MetricId,
		MetricName:  metric.MetricName,
		MetricParam: metric.MetricParam,
		Status:      metric.Status,
		RsTypeId:    metric.RsTypeId,
	}

	resp, err := client.ModifyMetric(ctx, req)
	if err != nil {
		glog.Errorf("ModifyMetric failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("ModifyMetric success: %+v", resp)

	response.WriteAsJson(resp)
}

func DeleteMetrics(request *restful.Request, response *restful.Response) {
	metricIds := strings.Split(request.QueryParameter("metric_ids"), ",")

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DeleteMetricsRequest{
		MetricId: metricIds,
	}

	resp, err := client.DeleteMetrics(ctx, req)
	if err != nil {
		glog.Errorf("DeleteMetrics failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("DeleteMetrics success: %+v", resp)

	response.WriteAsJson(resp)
}

func CreatePolicy(request *restful.Request, response *restful.Response) {
	policy := new(models.Policy)

	err := request.ReadEntity(&policy)
	if err != nil {
		glog.Errorf("CreatePolicy request data error %+v.", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.CreatePolicyRequest{
		PolicyName:         policy.PolicyName,
		PolicyDescription:  policy.PolicyDescription,
		PolicyConfig:       policy.PolicyConfig,
		AvailableStartTime: policy.AvailableStartTime,
		AvailableEndTime:   policy.AvailableEndTime,
		RsTypeId:           policy.RsTypeId,
	}

	resp, err := client.CreatePolicy(ctx, req)
	if err != nil {
		glog.Errorf("CreatePolicy failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("CreatePolicy success: %+v", resp)

	response.WriteAsJson(resp)
}

func DescribePolicies(request *restful.Request, response *restful.Response) {
	policyIds := strings.Split(request.QueryParameter("policy_ids"), ",")
	policyNames := strings.Split(request.QueryParameter("policy_names"), ",")
	policyDescriptions := strings.Split(request.QueryParameter("policy_descriptions"), ",")
	creators := strings.Split(request.QueryParameter("creators"), ",")
	rsTypeIds := strings.Split(request.QueryParameter("rs_type_ids"), ",")

	sortKey := request.QueryParameter("sort_key")
	reverse := parseBool(request.QueryParameter("reverse"))
	offset, _ := parseUint32(request.QueryParameter("offset"))
	limit, _ := parseUint32(request.QueryParameter("limit"))

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DescribePoliciesRequest{
		PolicyId:          policyIds,
		PolicyName:        policyNames,
		PolicyDescription: policyDescriptions,
		Creator:           creators,
		RsTypeId:          rsTypeIds,
		SortKey:           sortKey,
		Reverse:           reverse,
		Offset:            offset,
		Limit:             limit,
	}

	resp, err := client.DescribePolicies(ctx, req)
	if err != nil {
		glog.Errorf("DescribePolicies failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("DescribePolicies success: %+v", resp)

	response.WriteAsJson(resp)
}

func ModifyPolicy(request *restful.Request, response *restful.Response) {
	policy := new(models.Policy)

	err := request.ReadEntity(&policy)
	if err != nil {
		glog.Errorf("ModifyPolicy request data error %+v.", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.ModifyPolicyRequest{
		PolicyId:           policy.PolicyId,
		PolicyName:         policy.PolicyName,
		PolicyDescription:  policy.PolicyDescription,
		PolicyConfig:       policy.PolicyConfig,
		Creator:            policy.Creator,
		AvailableStartTime: policy.AvailableStartTime,
		AvailableEndTime:   policy.AvailableEndTime,
		RsTypeId:           policy.RsTypeId,
	}

	resp, err := client.ModifyPolicy(ctx, req)
	if err != nil {
		glog.Errorf("ModifyPolicy failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("ModifyPolicy success: %+v", resp)

	response.WriteAsJson(resp)
}

func DeletePolicies(request *restful.Request, response *restful.Response) {
	policyIds := strings.Split(request.QueryParameter("policy_ids"), ",")

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DeletePoliciesRequest{
		PolicyId: policyIds,
	}

	resp, err := client.DeletePolicies(ctx, req)
	if err != nil {
		glog.Errorf("DeletePolicies failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("DeletePolicies success: %+v", resp)

	response.WriteAsJson(resp)
}

type PolicyByAlert struct {
	AlertName          string    `json:"alert_name"`
	PolicyName         string    `json:"policy_name"`
	PolicyDescription  string    `json:"policy_description"`
	PolicyConfig       string    `json:"policy_config"`
	Creator            string    `json:"creator"`
	AvailableStartTime string    `json:"available_start_time"`
	AvailableEndTime   string    `json:"available_end_time"`
	CreateTime         time.Time `json:"create_time"`
	UpdateTime         time.Time `json:"update_time"`
	RsTypeId           string    `json:"rs_type_id"`
}

type ModifyPolicyByAlertResponse struct {
	AlertName string `json:"alert_name"`
}

func modifyPolicyByAlert(resourceMap map[string]string, request *restful.Request, response *restful.Response) {
	resourceSearch, _ := json.Marshal(resourceMap)
	resp := ModifyPolicyByAlertResponse{}

	policyByAlert := new(PolicyByAlert)

	err := request.ReadEntity(&policyByAlert)
	if err != nil {
		glog.Errorf("ModifyPolicyByAlert request data error %+v.", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	alertNames := stringutils.SimplifyStringList(strings.Split(policyByAlert.AlertName, ","))
	if len(alertNames) == 0 {
		glog.Errorf("ModifyPolicyByAlert has no alert name specified.")
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	clientCustom, err := alclient.NewCustomClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	reqAlerts := &pb.DescribeAlertsWithResourceRequest{
		ResourceSearch: string(resourceSearch),
		AlertName:      alertNames,
	}

	respAlerts, err := clientCustom.DescribeAlertsWithResource(ctx, reqAlerts)

	if respAlerts.Total != 1 {
		glog.Infof("ModifyPolicyByAlert get no match alert name or duplicate names.")
		response.WriteAsJson(resp)
		return
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	var req = &pb.ModifyPolicyRequest{
		PolicyId:           respAlerts.AlertSet[0].PolicyId,
		PolicyName:         policyByAlert.PolicyName,
		PolicyDescription:  policyByAlert.PolicyDescription,
		PolicyConfig:       policyByAlert.PolicyConfig,
		Creator:            policyByAlert.Creator,
		AvailableStartTime: policyByAlert.AvailableStartTime,
		AvailableEndTime:   policyByAlert.AvailableEndTime,
		RsTypeId:           policyByAlert.RsTypeId,
	}

	respModify, err := client.ModifyPolicy(ctx, req)
	if err != nil {
		glog.Errorf("ModifyPolicyByAlert failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	if respModify.PolicyId != respAlerts.AlertSet[0].PolicyId {
		glog.Errorf("ModifyPolicyByAlert failed, PolicyId request[%+v] response[%+v] mismatch", respAlerts.AlertSet[0].PolicyId, respModify.PolicyId)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	resp = ModifyPolicyByAlertResponse{
		AlertName: policyByAlert.AlertName,
	}
	glog.Infof("ModifyPolicyByAlert success: %+v", resp)
	response.WriteAsJson(resp)
}

func ModifyPolicyByAlertCluster(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "cluster"

	modifyPolicyByAlert(resourceMap, request, response)
}

func ModifyPolicyByAlertNode(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "node"

	modifyPolicyByAlert(resourceMap, request, response)
}

func ModifyPolicyByAlertWorkspace(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "workspace"
	resourceMap["ws_name"] = request.PathParameter("ws_name")

	modifyPolicyByAlert(resourceMap, request, response)
}

func ModifyPolicyByAlertNamespace(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "namespace"
	resourceMap["ns_name"] = request.PathParameter("ns_name")

	modifyPolicyByAlert(resourceMap, request, response)
}

func ModifyPolicyByAlertWorkload(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "workload"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")

	modifyPolicyByAlert(resourceMap, request, response)
}

func ModifyPolicyByAlertPod(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "pod"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")

	modifyPolicyByAlert(resourceMap, request, response)
}

func ModifyPolicyByAlertContainer(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "container"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")
	resourceMap["pod_name"] = request.PathParameter("pod_name")

	modifyPolicyByAlert(resourceMap, request, response)
}

func CreateRule(request *restful.Request, response *restful.Response) {
	rule := new(models.Rule)

	err := request.ReadEntity(&rule)
	if err != nil {
		glog.Errorf("CreateRule request data error %+v.", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.CreateRuleRequest{
		RuleName:         rule.RuleName,
		Disabled:         rule.Disabled,
		MonitorPeriods:   rule.MonitorPeriods,
		Severity:         rule.Severity,
		MetricsType:      rule.MetricsType,
		ConditionType:    rule.ConditionType,
		Thresholds:       rule.Thresholds,
		Unit:             rule.Unit,
		ConsecutiveCount: rule.ConsecutiveCount,
		Inhibit:          rule.Inhibit,
		PolicyId:         rule.PolicyId,
		MetricId:         rule.MetricId,
	}

	resp, err := client.CreateRule(ctx, req)
	if err != nil {
		glog.Errorf("CreateRule failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("CreateRule success: %+v", resp)

	response.WriteAsJson(resp)
}

func DescribeRules(request *restful.Request, response *restful.Response) {
	ruleIds := strings.Split(request.QueryParameter("rule_ids"), ",")
	ruleNames := strings.Split(request.QueryParameter("rule_names"), ",")
	disables := parseBools(strings.Split(request.QueryParameter("disables"), ","))
	monitorPeriods := parseUint32s(strings.Split(request.QueryParameter("monitor_periods"), ","))
	severities := strings.Split(request.QueryParameter("severities"), ",")
	metricsTypes := strings.Split(request.QueryParameter("metrics_types"), ",")
	conditionTypes := strings.Split(request.QueryParameter("condition_types"), ",")
	thresholds := strings.Split(request.QueryParameter("thresholds"), ",")
	uints := strings.Split(request.QueryParameter("uints"), ",")
	consecutiveCounts := parseUint32s(strings.Split(request.QueryParameter("consecutive_counts"), ","))
	inhibits := parseBools(strings.Split(request.QueryParameter("inhibits"), ","))
	policyIds := strings.Split(request.QueryParameter("policy_ids"), ",")
	metricIds := strings.Split(request.QueryParameter("metric_ids"), ",")

	sortKey := request.QueryParameter("sort_key")
	reverse := parseBool(request.QueryParameter("reverse"))
	offset, _ := parseUint32(request.QueryParameter("offset"))
	limit, _ := parseUint32(request.QueryParameter("limit"))

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DescribeRulesRequest{
		RuleId:           ruleIds,
		RuleName:         ruleNames,
		Disabled:         disables,
		MonitorPeriods:   monitorPeriods,
		Severity:         severities,
		MetricsType:      metricsTypes,
		ConditionType:    conditionTypes,
		Thresholds:       thresholds,
		Unit:             uints,
		ConsecutiveCount: consecutiveCounts,
		Inhibit:          inhibits,
		PolicyId:         policyIds,
		MetricId:         metricIds,
		SortKey:          sortKey,
		Reverse:          reverse,
		Offset:           offset,
		Limit:            limit,
	}

	resp, err := client.DescribeRules(ctx, req)
	if err != nil {
		glog.Errorf("DescribeRules failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("DescribeRules success: %+v", resp)

	response.WriteAsJson(resp)
}

func ModifyRule(request *restful.Request, response *restful.Response) {
	rule := new(models.Rule)

	err := request.ReadEntity(&rule)
	if err != nil {
		glog.Errorf("ModifyRule request data error %+v.", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.ModifyRuleRequest{
		RuleId:           rule.RuleId,
		RuleName:         rule.RuleName,
		Disabled:         rule.Disabled,
		MonitorPeriods:   rule.MonitorPeriods,
		Severity:         rule.Severity,
		MetricsType:      rule.MetricsType,
		ConditionType:    rule.ConditionType,
		Thresholds:       rule.Thresholds,
		Unit:             rule.Unit,
		ConsecutiveCount: rule.ConsecutiveCount,
		Inhibit:          rule.Inhibit,
	}

	resp, err := client.ModifyRule(ctx, req)
	if err != nil {
		glog.Errorf("ModifyRule failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("ModifyRule success: %+v", resp)

	response.WriteAsJson(resp)
}

func DeleteRules(request *restful.Request, response *restful.Response) {
	ruleIds := strings.Split(request.QueryParameter("rule_ids"), ",")

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DeleteRulesRequest{
		RuleId: ruleIds,
	}

	resp, err := client.DeleteRules(ctx, req)
	if err != nil {
		glog.Errorf("DeleteRules failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("DeleteRules success: %+v", resp)

	response.WriteAsJson(resp)
}

func CreateAlert(request *restful.Request, response *restful.Response) {
	alert := new(models.Alert)

	err := request.ReadEntity(&alert)
	if err != nil {
		glog.Errorf("CreateAlert request data error %+v.", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.CreateAlertRequest{
		AlertName:  alert.AlertName,
		PolicyId:   alert.PolicyId,
		RsFilterId: alert.RsFilterId,
	}

	resp, err := client.CreateAlert(ctx, req)
	if err != nil {
		glog.Errorf("CreateAlert failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("CreateAlert success: %+v", resp)

	response.WriteAsJson(resp)
}

func DescribeAlerts(request *restful.Request, response *restful.Response) {
	alertIds := strings.Split(request.QueryParameter("alert_ids"), ",")
	alertNames := strings.Split(request.QueryParameter("alert_names"), ",")
	disables := parseBools(strings.Split(request.QueryParameter("disabled"), ","))
	runningStatus := strings.Split(request.QueryParameter("running_status"), ",")
	policyIds := strings.Split(request.QueryParameter("policy_id"), ",")
	rsFilterIds := strings.Split(request.QueryParameter("rs_filter_id"), ",")
	executorIds := strings.Split(request.QueryParameter("executor_id"), ",")

	sortKey := request.QueryParameter("sort_key")
	reverse := parseBool(request.QueryParameter("reverse"))
	offset, _ := parseUint32(request.QueryParameter("offset"))
	limit, _ := parseUint32(request.QueryParameter("limit"))

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DescribeAlertsRequest{
		AlertId:       alertIds,
		AlertName:     alertNames,
		Disabled:      disables,
		RunningStatus: runningStatus,
		PolicyId:      policyIds,
		RsFilterId:    rsFilterIds,
		ExecutorId:    executorIds,
		SortKey:       sortKey,
		Reverse:       reverse,
		Offset:        offset,
		Limit:         limit,
	}

	resp, err := client.DescribeAlerts(ctx, req)
	if err != nil {
		glog.Errorf("DescribeAlerts failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("DescribeAlerts success: %+v", resp)

	response.WriteAsJson(resp)
}

func ModifyAlert(request *restful.Request, response *restful.Response) {
	alert := new(models.Alert)

	err := request.ReadEntity(&alert)
	if err != nil {
		glog.Errorf("ModifyAlert request data error %+v.", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.ModifyAlertRequest{
		AlertId:    alert.AlertId,
		AlertName:  alert.AlertName,
		Disabled:   alert.Disabled,
		PolicyId:   alert.PolicyId,
		RsFilterId: alert.RsFilterId,
	}

	resp, err := client.ModifyAlert(ctx, req)
	if err != nil {
		glog.Errorf("ModifyAlert failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("ModifyAlert success: %+v", resp)

	response.WriteAsJson(resp)
}

func DeleteAlerts(request *restful.Request, response *restful.Response) {
	alertIds := strings.Split(request.QueryParameter("alert_ids"), ",")

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DeleteAlertsRequest{
		AlertId: alertIds,
	}

	resp, err := client.DeleteAlerts(ctx, req)
	if err != nil {
		glog.Errorf("DeleteAlerts failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("DeleteAlerts success: %+v", resp)

	response.WriteAsJson(resp)
}

func removeResourceFilter(client pb.AlertManagerClient, ctx context.Context, rsFilterId string) {
	var reqDeleteResourceFilter = &pb.DeleteResourceFiltersRequest{
		RsFilterId: strings.Split(rsFilterId, ","),
	}

	client.DeleteResourceFilters(ctx, reqDeleteResourceFilter)
}

func removePolicy(client pb.AlertManagerClient, ctx context.Context, policyId string) {
	var reqDeletePolicies = &pb.DeletePoliciesRequest{
		PolicyId: strings.Split(policyId, ","),
	}

	client.DeletePolicies(ctx, reqDeletePolicies)
}

func createAlertInfo(resourceMap map[string]string, request *restful.Request, response *restful.Response) {
	alertInfo := new(models.AlertInfo)

	err := request.ReadEntity(&alertInfo)
	if err != nil {
		glog.Errorf("createAlertInfo request data error %+v.", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	clientCustom, err := alclient.NewCustomClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	//1. Check Rules Length
	if len(alertInfo.Rules) == 0 {
		glog.Errorf("CreateAlertInfo Rules Length error")
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	//2. Check if resource_type match
	rsTypeIds := strings.Split(alertInfo.RsFilter.RsTypeId, ",")

	var reqRsType = &pb.DescribeResourceTypesRequest{
		RsTypeId: rsTypeIds,
	}

	respRsType, err := client.DescribeResourceTypes(ctx, reqRsType)
	if err != nil {
		glog.Errorf("CreateAlertInfo DescribeResourceTypes failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	if respRsType.Total != 1 {
		glog.Errorf("CreateAlertInfo resource type error")
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	if respRsType.ResourceTypeSet[0].RsTypeName != resourceMap["rs_type_name"] {
		glog.Errorf("CreateAlertInfo resource type mismatch")
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	//3. Check workspace namespace matches
	rsFilterURI := make(map[string]string)
	err = json.Unmarshal([]byte(alertInfo.RsFilter.RsFilterParam), &rsFilterURI)
	if err != nil {
		glog.Errorf("CreateAlertInfo Unmarshal rsFilterURI Error: %+v", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	uriCorrect := true
	switch resourceMap["rs_type_name"] {
	case "cluster":
		break
	case "node":
		break
	case "workspace":
		if resourceMap["ws_name"] != rsFilterURI["ws_name"] {
			uriCorrect = false
		}
	case "namespace":
		if resourceMap["ns_name"] != rsFilterURI["ns_name"] {
			uriCorrect = false
		}
	case "workload":
		if resourceMap["ns_name"] != rsFilterURI["ns_name"] {
			uriCorrect = false
		}
	case "pod":
		if resourceMap["ns_name"] != rsFilterURI["ns_name"] {
			uriCorrect = false
		}
	case "container":
		if resourceMap["ns_name"] != rsFilterURI["ns_name"] {
			uriCorrect = false
		}
	}

	if !uriCorrect {
		glog.Errorf("CreateAlertInfo uri mismatch")
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	//4. Check if same alert_name exists
	resourceSearch, _ := json.Marshal(resourceMap)
	alertNames := strings.Split(alertInfo.Alert.AlertName, ",")
	var reqCheck = &pb.DescribeAlertsWithResourceRequest{
		ResourceSearch: string(resourceSearch),
		AlertName:      alertNames,
	}

	respCheck, err := clientCustom.DescribeAlertsWithResource(ctx, reqCheck)
	if err != nil {
		glog.Errorf("CreateAlertInfo check alert name failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	if respCheck.Total != 0 {
		glog.Errorf("CreateAlertInfo alert name already exists")
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	//5. Create Resource Filter
	var reqRsFilter = &pb.CreateResourceFilterRequest{
		RsFilterName:  alertInfo.RsFilter.RsFilterName,
		RsFilterParam: alertInfo.RsFilter.RsFilterParam,
		Status:        alertInfo.RsFilter.Status,
		RsTypeId:      alertInfo.RsFilter.RsTypeId,
	}

	respRsFilter, err := client.CreateResourceFilter(ctx, reqRsFilter)
	if err != nil {
		glog.Errorf("CreateAlertInfo Resource Filter failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("CreateAlertInfo Resource Filter success: %+v", respRsFilter)
	rsFilterId := respRsFilter.RsFilterId

	//6. Create Policy
	var reqPolicy = &pb.CreatePolicyRequest{
		PolicyName:         alertInfo.Policy.PolicyName,
		PolicyDescription:  alertInfo.Policy.PolicyDescription,
		PolicyConfig:       alertInfo.Policy.PolicyConfig,
		Creator:            alertInfo.Policy.Creator,
		AvailableStartTime: alertInfo.Policy.AvailableStartTime,
		AvailableEndTime:   alertInfo.Policy.AvailableEndTime,
		RsTypeId:           alertInfo.RsFilter.RsTypeId,
	}

	respPolicy, err := client.CreatePolicy(ctx, reqPolicy)
	if err != nil {
		glog.Errorf("CreateAlertInfo Policy failed: %+v", err)

		removeResourceFilter(client, ctx, rsFilterId)

		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("CreateAlertInfo Policy success: %+v", respPolicy)
	policyId := respPolicy.PolicyId

	//7. Create Action
	var reqAction = &pb.CreateActionRequest{
		ActionName:      alertInfo.Action.ActionName,
		PolicyId:        policyId,
		NfAddressListId: alertInfo.Action.NfAddressListId,
	}

	respAction, err := client.CreateAction(ctx, reqAction)
	if err != nil {
		glog.Errorf("CreateAlertInfo Action failed: %+v", err)

		removeResourceFilter(client, ctx, rsFilterId)
		removePolicy(client, ctx, policyId)

		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("CreateAlertInfo Action success: %+v", respAction)

	//8. Create Rules
	createRulesSuccess := true
	for _, rule := range alertInfo.Rules {
		var reqRule = &pb.CreateRuleRequest{
			RuleName:         rule.RuleName,
			Disabled:         rule.Disabled,
			MonitorPeriods:   rule.MonitorPeriods,
			Severity:         rule.Severity,
			MetricsType:      rule.MetricsType,
			ConditionType:    rule.ConditionType,
			Thresholds:       rule.Thresholds,
			Unit:             rule.Unit,
			ConsecutiveCount: rule.ConsecutiveCount,
			Inhibit:          rule.Inhibit,
			PolicyId:         policyId,
			MetricId:         rule.MetricId,
		}

		_, err := client.CreateRule(ctx, reqRule)
		if err != nil {
			createRulesSuccess = false
			break
		}
	}

	if !createRulesSuccess {
		glog.Errorf("CreateAlertInfo Rules failed")

		removeResourceFilter(client, ctx, rsFilterId)
		removePolicy(client, ctx, policyId)

		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("CreateAlertInfo Rules success")

	//9. Create Alert
	var reqAlert = &pb.CreateAlertRequest{
		AlertName:  alertInfo.Alert.AlertName,
		PolicyId:   policyId,
		RsFilterId: rsFilterId,
	}

	respAlert, err := client.CreateAlert(ctx, reqAlert)
	if err != nil {
		glog.Errorf("CreateAlertInfo Alert failed: %+v", err)

		removeResourceFilter(client, ctx, rsFilterId)
		removePolicy(client, ctx, policyId)

		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("CreateAlertInfo Alert success: %+v", respAlert)

	response.WriteAsJson(respAlert)
}

func CreateAlertCluster(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "cluster"

	createAlertInfo(resourceMap, request, response)
}

func CreateAlertNode(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "node"

	createAlertInfo(resourceMap, request, response)
}

func CreateAlertWorkspace(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "workspace"
	resourceMap["ws_name"] = request.PathParameter("ws_name")

	createAlertInfo(resourceMap, request, response)
}

func CreateAlertNamespace(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "namespace"
	resourceMap["ns_name"] = request.PathParameter("ns_name")

	createAlertInfo(resourceMap, request, response)
}

func CreateAlertWorkload(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "workload"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")

	createAlertInfo(resourceMap, request, response)
}

func CreateAlertPod(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "pod"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")

	createAlertInfo(resourceMap, request, response)
}

func CreateAlertContainer(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "container"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")
	resourceMap["pod_name"] = request.PathParameter("pod_name")

	createAlertInfo(resourceMap, request, response)
}

type ModifyAlertByNameResponse struct {
	AlertName string `json:"alert_name"`
}

func modifyAlertByName(resourceMap map[string]string, request *restful.Request, response *restful.Response) {
	alert := new(models.Alert)

	err := request.ReadEntity(&alert)
	if err != nil {
		glog.Errorf("ModifyAlertByName request data error %+v.", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	alertNames := stringutils.SimplifyStringList(strings.Split(alert.AlertName, ","))
	if len(alertNames) == 0 {
		glog.Errorf("ModifyAlertByName has no alert name specified.")
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	clientCustom, err := alclient.NewCustomClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	resourceSearch, _ := json.Marshal(resourceMap)
	var reqCheck = &pb.DescribeAlertsWithResourceRequest{
		ResourceSearch: string(resourceSearch),
		AlertName:      alertNames,
	}

	respAlerts, err := clientCustom.DescribeAlertsWithResource(ctx, reqCheck)
	if err != nil {
		glog.Errorf("ModifyAlertByName check alert name failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	if respAlerts.Total != 1 {
		glog.Infof("ModifyAlertByName get no match alert name or duplicate names.")
		response.WriteAsJson(&ModifyAlertByNameResponse{})
		return
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	var req = &pb.ModifyAlertRequest{
		AlertId:    respAlerts.AlertSet[0].AlertId,
		Disabled:   alert.Disabled,
		PolicyId:   alert.PolicyId,
		RsFilterId: alert.RsFilterId,
	}

	respModify, err := client.ModifyAlert(ctx, req)
	if err != nil {
		glog.Errorf("ModifyAlertByName failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	if respModify.AlertId != respAlerts.AlertSet[0].AlertId {
		glog.Errorf("ModifyAlertByName failed, AlertId request[%+v] response[%+v] mismatch", respAlerts.AlertSet[0].AlertId, respModify.AlertId)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	resp := ModifyAlertByNameResponse{
		AlertName: alert.AlertName,
	}
	glog.Infof("ModifyAlertByName success: %+v", resp)
	response.WriteAsJson(resp)
}

func ModifyAlertByNameCluster(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "cluster"

	modifyAlertByName(resourceMap, request, response)
}

func ModifyAlertByNameNode(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "node"

	modifyAlertByName(resourceMap, request, response)
}

func ModifyAlertByNameWorkspace(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "workspace"
	resourceMap["ws_name"] = request.PathParameter("ws_name")

	modifyAlertByName(resourceMap, request, response)
}

func ModifyAlertByNameNamespace(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "namespace"
	resourceMap["ns_name"] = request.PathParameter("ns_name")

	modifyAlertByName(resourceMap, request, response)
}

func ModifyAlertByNameWorkload(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "workload"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")

	modifyAlertByName(resourceMap, request, response)
}

func ModifyAlertByNamePod(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "pod"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")

	modifyAlertByName(resourceMap, request, response)
}

func ModifyAlertByNameContainer(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "container"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")
	resourceMap["pod_name"] = request.PathParameter("pod_name")

	modifyAlertByName(resourceMap, request, response)
}

type DeleteAlertsByNameResponse struct {
	AlertName []string `json:"alert_name"`
}

func deleteAlertsByName(resourceMap map[string]string, request *restful.Request, response *restful.Response) {
	alertNames := stringutils.SimplifyStringList(strings.Split(request.QueryParameter("alert_names"), ","))
	if len(alertNames) == 0 {
		glog.Errorf("DeleteAlertsByName has no alert name specified.")
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.New("DeleteAlertsByName has no alert name specified."))
		return
	}

	clientCustom, err := alclient.NewCustomClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	resourceSearch, _ := json.Marshal(resourceMap)
	var reqCheck = &pb.DescribeAlertsWithResourceRequest{
		ResourceSearch: string(resourceSearch),
		AlertName:      alertNames,
	}

	respAlerts, err := clientCustom.DescribeAlertsWithResource(ctx, reqCheck)
	if err != nil {
		glog.Errorf("DeleteAlertsByName check alert name failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	if respAlerts.Total == 0 {
		glog.Info("DeleteAlertsByName get no match alert name.")
		response.WriteAsJson(&DeleteAlertsByNameResponse{})
		return
	}

	alertIdName := map[string]string{}

	alertIds := []string{}
	for _, alert := range respAlerts.AlertSet {
		alertIds = append(alertIds, alert.AlertId)
		alertIdName[alert.AlertId] = alert.AlertName
	}

	var req = &pb.DeleteAlertsRequest{
		AlertId: alertIds,
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	respDelete, err := client.DeleteAlerts(ctx, req)
	if err != nil {
		glog.Errorf("DeleteAlertsByName failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	alertNamesSuccess := []string{}
	for _, alertDelete := range respDelete.AlertId {
		alertNamesSuccess = append(alertNamesSuccess, alertIdName[alertDelete])
	}

	resp := DeleteAlertsByNameResponse{
		AlertName: alertNamesSuccess,
	}

	glog.Infof("DeleteAlertsByName success: %+v", resp)
	response.WriteAsJson(resp)
}

func DeleteAlertsByNameCluster(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "cluster"

	deleteAlertsByName(resourceMap, request, response)
}

func DeleteAlertsByNameNode(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "node"

	deleteAlertsByName(resourceMap, request, response)
}

func DeleteAlertsByNameWorkspace(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "workspace"
	resourceMap["ws_name"] = request.PathParameter("ws_name")

	deleteAlertsByName(resourceMap, request, response)
}

func DeleteAlertsByNameNamespace(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "namespace"
	resourceMap["ns_name"] = request.PathParameter("ns_name")

	deleteAlertsByName(resourceMap, request, response)
}

func DeleteAlertsByNameWorkload(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "workload"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")

	deleteAlertsByName(resourceMap, request, response)
}

func DeleteAlertsByNamePod(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "pod"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")

	deleteAlertsByName(resourceMap, request, response)
}

func DeleteAlertsByNameContainer(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "container"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")
	resourceMap["pod_name"] = request.PathParameter("pod_name")

	deleteAlertsByName(resourceMap, request, response)
}

func describeAlertDetails(resourceMap map[string]string, request *restful.Request, response *restful.Response) {
	resourceSearch, _ := json.Marshal(resourceMap)
	alertIds := strings.Split(request.QueryParameter("alert_ids"), ",")
	alertNames := strings.Split(request.QueryParameter("alert_names"), ",")
	disables := parseBools(strings.Split(request.QueryParameter("disabled"), ","))
	runningStatus := strings.Split(request.QueryParameter("running_status"), ",")
	policyIds := strings.Split(request.QueryParameter("policy_ids"), ",")
	creators := strings.Split(request.QueryParameter("creators"), ",")
	rsFilterIds := strings.Split(request.QueryParameter("rs_filter_ids"), ",")
	executorIds := strings.Split(request.QueryParameter("executor_ids"), ",")

	sortKey := request.QueryParameter("sort_key")
	reverse := parseBool(request.QueryParameter("reverse"))
	offset, _ := parseUint32(request.QueryParameter("offset"))
	limit, _ := parseUint32(request.QueryParameter("limit"))

	clientCustom, err := alclient.NewCustomClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DescribeAlertDetailsRequest{
		ResourceSearch: string(resourceSearch),
		SearchWord:     request.QueryParameter("search_word"),
		AlertId:        alertIds,
		AlertName:      alertNames,
		Disabled:       disables,
		RunningStatus:  runningStatus,
		PolicyId:       policyIds,
		Creator:        creators,
		RsFilterId:     rsFilterIds,
		ExecutorId:     executorIds,
		SortKey:        sortKey,
		Reverse:        reverse,
		Offset:         offset,
		Limit:          limit,
	}

	resp, err := clientCustom.DescribeAlertDetails(ctx, req)
	if err != nil {
		glog.Errorf("DescribeAlertDetails failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("DescribeAlertDetails success: %+v", resp)

	response.WriteAsJson(resp)
}

func DescribeAlertDetailsCluster(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "cluster"

	describeAlertDetails(resourceMap, request, response)
}

func DescribeAlertDetailsNode(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "node"

	describeAlertDetails(resourceMap, request, response)
}

func DescribeAlertDetailsWorkspace(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "workspace"
	resourceMap["ws_name"] = request.PathParameter("ws_name")

	describeAlertDetails(resourceMap, request, response)
}

func DescribeAlertDetailsNamespace(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "namespace"
	resourceMap["ns_name"] = request.PathParameter("ns_name")

	describeAlertDetails(resourceMap, request, response)
}

func DescribeAlertDetailsWorkload(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "workload"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")

	describeAlertDetails(resourceMap, request, response)
}

func DescribeAlertDetailsPod(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "pod"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")

	describeAlertDetails(resourceMap, request, response)
}

func DescribeAlertDetailsContainer(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "container"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")
	resourceMap["pod_name"] = request.PathParameter("pod_name")

	describeAlertDetails(resourceMap, request, response)
}

func describeAlertStatus(resourceMap map[string]string, request *restful.Request, response *restful.Response) {
	resourceSearch, _ := json.Marshal(resourceMap)
	alertIds := strings.Split(request.QueryParameter("alert_ids"), ",")
	alertNames := strings.Split(request.QueryParameter("alert_names"), ",")
	disables := parseBools(strings.Split(request.QueryParameter("disabled"), ","))
	runningStatus := strings.Split(request.QueryParameter("running_status"), ",")
	policyIds := strings.Split(request.QueryParameter("policy_ids"), ",")
	creators := strings.Split(request.QueryParameter("creators"), ",")
	rsFilterIds := strings.Split(request.QueryParameter("rs_filter_ids"), ",")
	executorIds := strings.Split(request.QueryParameter("executor_ids"), ",")
	ruleIds := strings.Split(request.QueryParameter("rule_ids"), ",")

	sortKey := request.QueryParameter("sort_key")
	reverse := parseBool(request.QueryParameter("reverse"))
	offset, _ := parseUint32(request.QueryParameter("offset"))
	limit, _ := parseUint32(request.QueryParameter("limit"))

	clientCustom, err := alclient.NewCustomClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DescribeAlertStatusRequest{
		ResourceSearch: string(resourceSearch),
		AlertId:        alertIds,
		AlertName:      alertNames,
		Disabled:       disables,
		RunningStatus:  runningStatus,
		PolicyId:       policyIds,
		Creator:        creators,
		RsFilterId:     rsFilterIds,
		ExecutorId:     executorIds,
		RuleId:         ruleIds,
		SortKey:        sortKey,
		Reverse:        reverse,
		Offset:         offset,
		Limit:          limit,
	}

	resp, err := clientCustom.DescribeAlertStatus(ctx, req)
	if err != nil {
		glog.Errorf("DescribeAlertStatus failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
	}

	glog.Infof("DescribeAlertStatus success: %+v", resp)

	response.WriteAsJson(resp)
}

func DescribeAlertStatusCluster(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "cluster"

	describeAlertStatus(resourceMap, request, response)
}

func DescribeAlertStatusNode(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "node"

	describeAlertStatus(resourceMap, request, response)
}

func DescribeAlertStatusWorkspace(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "workspace"
	resourceMap["ws_name"] = request.PathParameter("ws_name")

	describeAlertStatus(resourceMap, request, response)
}

func DescribeAlertStatusNamespace(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "namespace"
	resourceMap["ns_name"] = request.PathParameter("ns_name")

	describeAlertStatus(resourceMap, request, response)
}

func DescribeAlertStatusWorkload(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "workload"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")

	describeAlertStatus(resourceMap, request, response)
}

func DescribeAlertStatusPod(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "pod"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")

	describeAlertStatus(resourceMap, request, response)
}

func DescribeAlertStatusContainer(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "container"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")
	resourceMap["pod_name"] = request.PathParameter("pod_name")

	describeAlertStatus(resourceMap, request, response)
}

func DescribeHistories(request *restful.Request, response *restful.Response) {
	historyIds := strings.Split(request.QueryParameter("history_ids"), ",")
	historyNames := strings.Split(request.QueryParameter("history_names"), ",")
	events := strings.Split(request.QueryParameter("events"), ",")
	contents := strings.Split(request.QueryParameter("contents"), ",")
	alertIds := strings.Split(request.QueryParameter("alert_ids"), ",")

	sortKey := request.QueryParameter("sort_key")
	reverse := parseBool(request.QueryParameter("reverse"))
	offset, _ := parseUint32(request.QueryParameter("offset"))
	limit, _ := parseUint32(request.QueryParameter("limit"))

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DescribeHistoriesRequest{
		HistoryId:   historyIds,
		HistoryName: historyNames,
		Event:       events,
		Content:     contents,
		AlertId:     alertIds,
		SortKey:     sortKey,
		Reverse:     reverse,
		Offset:      offset,
		Limit:       limit,
	}

	resp, err := client.DescribeHistories(ctx, req)
	if err != nil {
		glog.Errorf("DescribeHistories failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("DescribeHistories success: %+v", resp)

	response.WriteAsJson(resp)
}

func describeHistoryDetail(resourceMap map[string]string, request *restful.Request, response *restful.Response) {
	resourceSearch, _ := json.Marshal(resourceMap)
	historyIds := strings.Split(request.QueryParameter("history_ids"), ",")
	historyNames := strings.Split(request.QueryParameter("history_names"), ",")
	alertNames := strings.Split(request.QueryParameter("alert_names"), ",")
	ruleNames := strings.Split(request.QueryParameter("rule_names"), ",")
	events := strings.Split(request.QueryParameter("events"), ",")
	ruleIds := strings.Split(request.QueryParameter("rule_ids"), ",")
	resourceNames := strings.Split(request.QueryParameter("resource_names"), ",")
	recent := parseBool(request.QueryParameter("recent"))

	sortKey := request.QueryParameter("sort_key")
	reverse := parseBool(request.QueryParameter("reverse"))
	offset, _ := parseUint32(request.QueryParameter("offset"))
	limit, _ := parseUint32(request.QueryParameter("limit"))

	clientCustom, err := alclient.NewCustomClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DescribeHistoryDetailRequest{
		ResourceSearch: string(resourceSearch),
		SearchWord:     request.QueryParameter("search_word"),
		HistoryId:      historyIds,
		HistoryName:    historyNames,
		AlertName:      alertNames,
		RuleName:       ruleNames,
		Event:          events,
		RuleId:         ruleIds,
		ResourceName:   resourceNames,
		Recent:         recent,
		SortKey:        sortKey,
		Reverse:        reverse,
		Offset:         offset,
		Limit:          limit,
	}

	resp, err := clientCustom.DescribeHistoryDetail(ctx, req)
	if err != nil {
		glog.Errorf("DescribeHistoryDetail failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("DescribeHistoryDetail success: %+v", resp)

	response.WriteAsJson(resp)
}

func DescribeHistoryDetailCluster(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "cluster"

	describeHistoryDetail(resourceMap, request, response)
}

func DescribeHistoryDetailNode(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "node"

	describeHistoryDetail(resourceMap, request, response)
}

func DescribeHistoryDetailWorkspace(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "workspace"
	resourceMap["ws_name"] = request.PathParameter("ws_name")

	describeHistoryDetail(resourceMap, request, response)
}

func DescribeHistoryDetailNamespace(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "namespace"
	resourceMap["ns_name"] = request.PathParameter("ns_name")

	describeHistoryDetail(resourceMap, request, response)
}

func DescribeHistoryDetailWorkload(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "workload"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")

	describeHistoryDetail(resourceMap, request, response)
}

func DescribeHistoryDetailPod(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "pod"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")

	describeHistoryDetail(resourceMap, request, response)
}

func DescribeHistoryDetailContainer(request *restful.Request, response *restful.Response) {
	resourceMap := map[string]string{}
	resourceMap["rs_type_name"] = "container"
	resourceMap["ns_name"] = request.PathParameter("ns_name")
	resourceMap["node_id"] = request.PathParameter("node_id")
	resourceMap["pod_name"] = request.PathParameter("pod_name")

	describeHistoryDetail(resourceMap, request, response)
}

func CreateComment(request *restful.Request, response *restful.Response) {
	comment := new(models.Comment)

	err := request.ReadEntity(&comment)
	if err != nil {
		glog.Errorf("CreateComment request data error %+v.", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create comment grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.CreateCommentRequest{
		Addresser: comment.Addresser,
		Content:   comment.Content,
		HistoryId: comment.HistoryId,
	}

	resp, err := client.CreateComment(ctx, req)
	if err != nil {
		glog.Errorf("CreateComment failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("CreateComment success: %+v", resp)

	response.WriteAsJson(resp)
}

func DescribeComments(request *restful.Request, response *restful.Response) {
	commentIds := strings.Split(request.QueryParameter("comment_ids"), ",")
	addressers := strings.Split(request.QueryParameter("addressers"), ",")
	contents := strings.Split(request.QueryParameter("contents"), ",")
	historyIds := strings.Split(request.QueryParameter("history_ids"), ",")

	sortKey := request.QueryParameter("sort_key")
	reverse := parseBool(request.QueryParameter("reverse"))
	offset, _ := parseUint32(request.QueryParameter("offset"))
	limit, _ := parseUint32(request.QueryParameter("limit"))

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DescribeCommentsRequest{
		CommentId: commentIds,
		Addresser: addressers,
		Content:   contents,
		HistoryId: historyIds,
		SortKey:   sortKey,
		Reverse:   reverse,
		Offset:    offset,
		Limit:     limit,
	}

	resp, err := client.DescribeComments(ctx, req)
	if err != nil {
		glog.Errorf("DescribeComments failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("DescribeComments success: %+v", resp)

	response.WriteAsJson(resp)
}

func CreateAction(request *restful.Request, response *restful.Response) {
	action := new(models.Action)

	err := request.ReadEntity(&action)
	if err != nil {
		glog.Errorf("CreateAction request data error %+v.", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.CreateActionRequest{
		ActionName:      action.ActionName,
		TriggerStatus:   action.TriggerStatus,
		TriggerAction:   action.TriggerAction,
		PolicyId:        action.PolicyId,
		NfAddressListId: action.NfAddressListId,
	}

	resp, err := client.CreateAction(ctx, req)
	if err != nil {
		glog.Errorf("CreateAction failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("CreateAction success: %+v", resp)

	response.WriteAsJson(resp)
}

func DescribeActions(request *restful.Request, response *restful.Response) {
	actionIds := strings.Split(request.QueryParameter("action_ids"), ",")
	actionNames := strings.Split(request.QueryParameter("action_names"), ",")
	triggerStatus := strings.Split(request.QueryParameter("trigger_status"), ",")
	triggerActions := strings.Split(request.QueryParameter("trigger_actions"), ",")
	policyIds := strings.Split(request.QueryParameter("policy_ids"), ",")
	nfAddressListIds := strings.Split(request.QueryParameter("nf_address_list_ids"), ",")

	sortKey := request.QueryParameter("sort_key")
	reverse := parseBool(request.QueryParameter("reverse"))
	offset, _ := parseUint32(request.QueryParameter("offset"))
	limit, _ := parseUint32(request.QueryParameter("limit"))

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DescribeActionsRequest{
		ActionId:        actionIds,
		ActionName:      actionNames,
		TriggerStatus:   triggerStatus,
		TriggerAction:   triggerActions,
		PolicyId:        policyIds,
		NfAddressListId: nfAddressListIds,
		SortKey:         sortKey,
		Reverse:         reverse,
		Offset:          offset,
		Limit:           limit,
	}

	resp, err := client.DescribeActions(ctx, req)
	if err != nil {
		glog.Errorf("DescribeActions failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("DescribeActions success: %+v", resp)

	response.WriteAsJson(resp)
}

func ModifyAction(request *restful.Request, response *restful.Response) {
	action := new(models.Action)

	err := request.ReadEntity(&action)
	if err != nil {
		glog.Errorf("ModifyAction request data error %+v.", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.ModifyActionRequest{
		ActionId:        action.ActionId,
		ActionName:      action.ActionName,
		TriggerStatus:   action.TriggerStatus,
		TriggerAction:   action.TriggerAction,
		PolicyId:        action.PolicyId,
		NfAddressListId: action.NfAddressListId,
	}

	resp, err := client.ModifyAction(ctx, req)
	if err != nil {
		glog.Errorf("ModifyAction failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("ModifyAction success: %+v", resp)

	response.WriteAsJson(resp)
}

func DeleteActions(request *restful.Request, response *restful.Response) {
	actionIds := strings.Split(request.QueryParameter("action_ids"), ",")

	client, err := alclient.NewClient()
	if err != nil {
		glog.Errorf("Failed to create alert grpc client %+v.", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grpcTimeoutSeconds*time.Second)
	defer cancel()

	var req = &pb.DeleteActionsRequest{
		ActionId: actionIds,
	}

	resp, err := client.DeleteActions(ctx, req)
	if err != nil {
		glog.Errorf("DeleteActions failed: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	glog.Infof("DeleteActions success: %+v", resp)

	response.WriteAsJson(resp)
}

func DescribeResourcesCluster(request *restful.Request, response *restful.Response) {
}

func DescribeResourcesNode(request *restful.Request, response *restful.Response) {
	resources := []string{}

	resourceSelector := []map[string]string{}
	err := json.Unmarshal([]byte(request.QueryParameter("selector")), &resourceSelector)
	if err != nil {
		glog.Errorf("Unmarshal DescribeResourcesNode Error: %+v", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	labelSelector := parseLabelSelector(resourceSelector)

	nodeList, err := k8sclient.NewK8sClient().CoreV1().Nodes().List(metaV1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		glog.Errorf("getResourceFilterURIBySelector list nodes error: %+v", err)
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	for _, node := range nodeList.Items {
		resources = append(resources, node.Name)
	}

	response.WriteAsJson(resources)
}

func DescribeResourcesWorkspace(request *restful.Request, response *restful.Response) {
}

func DescribeResourcesNamespace(request *restful.Request, response *restful.Response) {
}

func DescribeResourcesWorkload(request *restful.Request, response *restful.Response) {
	resources := []string{}

	resourceSelector := []map[string]string{}
	err := json.Unmarshal([]byte(request.QueryParameter("selector")), &resourceSelector)
	if err != nil {
		glog.Errorf("Unmarshal DescribeResourcesNode Error: %+v", err)
		response.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	labelSelector := parseLabelSelector(resourceSelector)

	namespace := request.PathParameter("ns_name")

	switch request.QueryParameter("workload_kind") {
	case "deployment":
		deploymentList, err := k8sclient.NewK8sClient().ExtensionsV1beta1().Deployments(namespace).List(metaV1.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			glog.Errorf("DescribeResourcesWorkload list deployments error: %+v", err)
			break
		}
		for _, deployment := range deploymentList.Items {
			resources = append(resources, deployment.Name)
		}
	case "statefulset":
		statefulsetList, err := k8sclient.NewK8sClient().AppsV1().StatefulSets(namespace).List(metaV1.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			glog.Errorf("DescribeResourcesWorkload list statefulsets error: %+v", err)
			break
		}
		for _, statefulset := range statefulsetList.Items {
			resources = append(resources, statefulset.Name)
		}
	case "daemonset":
		daemonsetList, err := k8sclient.NewK8sClient().ExtensionsV1beta1().DaemonSets(namespace).List(metaV1.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			glog.Errorf("DescribeResourcesWorkload list daemonsets error: %+v", err)
			break
		}
		for _, daemonset := range daemonsetList.Items {
			resources = append(resources, daemonset.Name)
		}
	}

	response.WriteAsJson(resources)
}

func DescribeResourcesPod(request *restful.Request, response *restful.Response) {
}

func DescribeResourcesContainer(request *restful.Request, response *restful.Response) {
}
