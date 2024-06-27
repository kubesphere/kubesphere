package overview

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/Masterminds/semver/v3"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"

	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
)

type Counter interface {
	RegisterResource(options ...RegisterOption)
	GetMetrics(metricNames []string, namespace, workspace, prefix string) (*MetricResults, error)
}

var _ Counter = &metricsCounter{}

type metricsCounter struct {
	registerTypeMap map[string]reflect.Type
	client          client.Client
}

func New(client client.Client) Counter {
	collector := &metricsCounter{
		registerTypeMap: make(map[string]reflect.Type),
		client:          client,
	}

	return collector
}

func (r *metricsCounter) RegisterResource(options ...RegisterOption) {
	for _, o := range options {
		r.register(o)
	}
}

func (r *metricsCounter) register(option RegisterOption) {
	r.registerTypeMap[option.MetricsName] = reflect.TypeOf(option.Type)
}

func (r *metricsCounter) GetMetrics(metricNames []string, namespace, workspace, prefix string) (*MetricResults, error) {
	result := &MetricResults{Results: make([]Metric, 0)}
	for _, n := range metricNames {
		metric, err := r.collect(n, prefix, namespace, workspace)
		if err != nil {
			return nil, err
		}

		result.AddMetric(metric)
	}

	return result, nil
}

func CustomMetric(metricName, prefix string, count int) *Metric {
	return newMetric(metricName, prefix, count)
}

func newMetric(metricName, prefix string, count int) *Metric {
	if prefix != "" {
		metricName = fmt.Sprintf("%s_%s", prefix, metricName)
	}
	metric := &Metric{
		MetricName: metricName,
		Data: MetricData{
			ResultType: "vector",
			Result:     make([]MetricValue, 0),
		},
	}

	value := MetricValue{
		Value: []interface{}{
			time.Now().Unix(),
			count,
		},
	}
	metric.Data.Result = append(metric.Data.Result, value)

	return metric
}

type RegisterOption struct {
	MetricsName string
	Type        client.ObjectList
}

func (r *metricsCounter) collect(metricName, prefix, namespace, workspace string) (*Metric, error) {
	t, exist := r.registerTypeMap[metricName]
	if !exist {
		return nil, errors.New("can not find metric type")
	}
	objVal := reflect.New(t.Elem())
	objList := objVal.Interface().(client.ObjectList)

	opts := make([]client.ListOption, 0)
	if workspace != "" {
		opt := client.MatchingLabels(map[string]string{tenantv1beta1.WorkspaceLabel: workspace})
		opts = append(opts, opt)
	}
	if namespace != "" {
		opt := client.InNamespace(namespace)
		opts = append(opts, opt)
	}

	err := r.client.List(context.Background(), objList, opts...)

	if err != nil {
		return nil, err
	}

	return newMetric(metricName, prefix, meta.LenList(objList)), nil
}

type MetricResults struct {
	Results []Metric `json:"results"`
}

type Metric struct {
	MetricName string     `json:"metric_name"`
	Data       MetricData `json:"data"`
}

type MetricData struct {
	ResultType string        `json:"resultType"`
	Result     []MetricValue `json:"result"`
}

type MetricValue struct {
	Value []interface{} `json:"value"`
}

func (r *MetricResults) AddMetric(metric *Metric) {
	r.Results = append(r.Results, *metric)
}

func NewDefaultRegisterOptions(k8sVersion *semver.Version) []RegisterOption {
	options := []RegisterOption{
		{
			MetricsName: NamespaceCount,
			Type:        &v12.NamespaceList{},
		},
		{
			MetricsName: WorkspaceCount,
			Type:        &tenantv1beta1.WorkspaceTemplateList{},
		},
		{
			MetricsName: ClusterCount,
			Type:        &clusterv1alpha1.ClusterList{},
		},
		{
			MetricsName: PodCount,
			Type:        &v12.PodList{},
		},
		{
			MetricsName: DeploymentCount,
			Type:        &appsv1.DeploymentList{},
		},
		{
			MetricsName: StatefulSetCount,
			Type:        &appsv1.StatefulSetList{},
		},
		{
			MetricsName: DaemonSetCount,
			Type:        &appsv1.DaemonSetList{},
		},
		{
			MetricsName: JobCount,
			Type:        &batchv1.JobList{},
		},
		{
			MetricsName: ServiceCount,
			Type:        &v12.ServiceList{},
		},
		{
			MetricsName: IngressCount,
			Type:        &v1.IngressList{},
		},
		{
			MetricsName: PersistentVolumeCount,
			Type:        &v12.PersistentVolumeList{},
		},
		{
			MetricsName: PersistentVolumeClaimCount,
			Type:        &v12.PersistentVolumeClaimList{},
		},
		{
			MetricsName: GlobalRoleCount,
			Type:        &iamv1beta1.GlobalRoleList{},
		},
		{
			MetricsName: WorkspaceRoleCount,
			Type:        &iamv1beta1.WorkspaceRoleList{},
		},
		{
			MetricsName: RoleCount,
			Type:        &iamv1beta1.RoleList{},
		},
		{
			MetricsName: ClusterRoleCount,
			Type:        &iamv1beta1.ClusterRoleList{},
		},
		{
			MetricsName: UserCount,
			Type:        &iamv1beta1.UserList{},
		},
		{
			MetricsName: GlobalRoleBindingCount,
			Type:        &iamv1beta1.GlobalRoleBindingList{},
		},
		{
			MetricsName: ClusterRoleBindingCount,
			Type:        &iamv1beta1.ClusterRoleBindingList{},
		},
		{
			MetricsName: RoleBindingCount,
			Type:        &iamv1beta1.RoleBindingList{},
		},
		{
			MetricsName: WorkspaceRoleBindingCount,
			Type:        &iamv1beta1.WorkspaceRoleBindingList{},
		},
		{
			MetricsName: ExtensionCount,
			Type:        &corev1alpha1.ExtensionList{},
		},
	}

	var cronJobType client.ObjectList
	if k8sutil.ServeBatchV1beta1(k8sVersion) {
		cronJobType = &batchv1beta1.CronJobList{}
	} else {
		cronJobType = &batchv1.CronJobList{}
	}
	options = append(options, RegisterOption{
		MetricsName: CronJobCount,
		Type:        cronJobType,
	})

	return options
}
