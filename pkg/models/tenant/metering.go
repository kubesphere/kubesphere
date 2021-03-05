package tenant

import (
	"context"
	"fmt"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/metering"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	meteringv1alpha1 "kubesphere.io/kubesphere/pkg/api/metering/v1alpha1"
	tenantv1alpha2 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha2"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	monitoringmodel "kubesphere.io/kubesphere/pkg/models/monitoring"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

type QueryOptions struct {
	MetricFilter string
	NamedMetrics []string

	Start time.Time
	End   time.Time
	Time  time.Time
	Step  time.Duration

	Target     string
	Identifier string
	Order      string
	Page       int
	Limit      int

	Option monitoring.QueryOption
}

func (q QueryOptions) isRangeQuery() bool {
	return q.Time.IsZero()
}

func (q QueryOptions) shouldSort() bool {
	return q.Target != "" && q.Identifier != ""
}

func (t *tenantOperator) makeQueryOptions(user user.Info, q meteringv1alpha1.Query, lvl monitoring.Level) (qo QueryOptions, err error) {
	if q.ResourceFilter == "" {
		q.ResourceFilter = meteringv1alpha1.DefaultFilter
	}

	qo.MetricFilter = q.MetricFilter
	if q.MetricFilter == "" {
		qo.MetricFilter = meteringv1alpha1.DefaultFilter
	}

	var decision authorizer.Decision
	switch lvl {
	case monitoring.LevelCluster:
		clusterOption := monitoring.ClusterOption{}
		qo.NamedMetrics = monitoringmodel.ClusterMetrics

		listPods := authorizer.AttributesRecord{
			User:            user,
			Verb:            "list",
			APIGroup:        "",
			APIVersion:      "v1",
			Resource:        "pods",
			ResourceRequest: true,
			ResourceScope:   request.ClusterScope,
		}
		decision, _, err = t.authorizer.Authorize(listPods)
		if err != nil {
			klog.Error(err)
			return
		}

		// only cluster admin is allowed
		if decision != authorizer.DecisionAllow {
			return qo, errors.New(fmt.Sprintf(meteringv1alpha1.ErrScopeNotAllowed, request.ClusterScope))
		}
		qo.Option = clusterOption

	case monitoring.LevelNode:
		qo.Identifier = monitoringmodel.IdentifierNode
		nodeOption := monitoring.NodeOption{
			ResourceFilter:   q.ResourceFilter,
			NodeName:         q.NodeName,
			PVCFilter:        q.PVCFilter,
			StorageClassName: q.StorageClassName,
		}
		qo.NamedMetrics = monitoringmodel.NodeMetrics

		listPods := authorizer.AttributesRecord{
			User:            user,
			Verb:            "list",
			APIGroup:        "",
			APIVersion:      "v1",
			Resource:        "pods",
			ResourceRequest: true,
			ResourceScope:   request.ClusterScope,
		}
		decision, _, err = t.authorizer.Authorize(listPods)
		if err != nil {
			klog.Error(err)
			return
		}

		// only cluster admin is allowed
		if decision != authorizer.DecisionAllow {
			return qo, errors.New(fmt.Sprintf(meteringv1alpha1.ErrScopeNotAllowed, request.ClusterScope))
		}
		qo.Option = nodeOption

	case monitoring.LevelWorkspace:
		qo.Identifier = monitoringmodel.IdentifierWorkspace

		// at least one of WorkspaceName, ResourceFilter isn't empty
		wsOption := monitoring.WorkspaceOption{
			ResourceFilter:   q.ResourceFilter, // ws filter
			WorkspaceName:    q.WorkspaceName,
			PVCFilter:        q.PVCFilter,
			StorageClassName: q.StorageClassName,
		}
		qo.NamedMetrics = monitoringmodel.WorkspaceMetrics

		wsScope := request.ClusterScope
		if q.WorkspaceName != "" {
			wsScope = request.WorkspaceScope
		}

		listPods := authorizer.AttributesRecord{
			User:            user,
			Verb:            "list",
			Resource:        "pods",
			Workspace:       q.WorkspaceName,
			ResourceScope:   wsScope,
			ResourceRequest: true,
		}
		decision, _, err = t.authorizer.Authorize(listPods)
		if err != nil {
			klog.Error(err)
			return
		}
		if decision != authorizer.DecisionAllow {
			// specified by WorkspaceName and not allowed
			if q.WorkspaceName != "" {
				return qo, errors.New(fmt.Sprintf(meteringv1alpha1.ErrScopeNotAllowed, wsScope))
			}

			// not specified by ResourceFilter or WorkspaceName
			if q.ResourceFilter == "" {
				return qo, errors.New(fmt.Sprintf(meteringv1alpha1.ErrScopeNotAllowed, wsScope))
			}
		}

		// apply ResourceFilter if necessary
		if q.ResourceFilter != "" {
			var wsList *api.ListResult
			qu := query.New()
			qu.LabelSelector = q.LabelSelector
			wsList, err = t.ListWorkspaces(user, qu)
			if err != nil {
				return qo, err
			}

			targetWs := []string{}
			for _, item := range wsList.Items {
				ws := item.(*tenantv1alpha2.WorkspaceTemplate)
				if ok, _ := regexp.MatchString(q.ResourceFilter, ws.ObjectMeta.GetName()); ok {
					listPods = authorizer.AttributesRecord{
						User:            user,
						Verb:            "list",
						Resource:        "pods",
						Workspace:       ws.ObjectMeta.GetName(),
						ResourceScope:   request.WorkspaceScope,
						ResourceRequest: true,
					}
					decision, _, err = t.authorizer.Authorize(listPods)
					if err != nil {
						klog.Error(err)
						return
					}
					if decision == authorizer.DecisionAllow {
						targetWs = append(targetWs, ws.ObjectMeta.GetName())
					}
				}
			}
			wsOption.ResourceFilter = strings.Join(targetWs, "|")
		}

		qo.Option = wsOption

	case monitoring.LevelNamespace:
		qo.Identifier = monitoringmodel.IdentifierNamespace
		nsOption := monitoring.NamespaceOption{
			ResourceFilter:   q.ResourceFilter, // ns filter
			WorkspaceName:    q.WorkspaceName,
			NamespaceName:    q.NamespaceName,
			PVCFilter:        q.PVCFilter,
			StorageClassName: q.StorageClassName,
		}
		qo.NamedMetrics = monitoringmodel.NamespaceMetrics

		nsScope := request.ClusterScope
		if q.WorkspaceName != "" {
			nsScope = request.WorkspaceScope
		}
		if q.NamespaceName != "" {
			nsScope = request.NamespaceScope
		}

		listPods := authorizer.AttributesRecord{
			User:            user,
			Verb:            "list",
			APIGroup:        "",
			APIVersion:      "v1",
			Resource:        "pods",
			ResourceRequest: true,
			Workspace:       q.WorkspaceName,
			Namespace:       q.NamespaceName,
			ResourceScope:   nsScope,
		}
		decision, _, err = t.authorizer.Authorize(listPods)
		if err != nil {
			klog.Error(err)
			return
		}
		if decision != authorizer.DecisionAllow {
			if q.WorkspaceName != "" {
				// specified by WorkspaceName & NamespaceName and not allowed
				if q.NamespaceName != "" {
					return qo, errors.New(fmt.Sprintf(meteringv1alpha1.ErrScopeNotAllowed, nsScope))
				}
			} else {
				// specified by NamespaceName & NamespaceName and not allowed
				if q.NamespaceName != "" {
					return qo, errors.New(fmt.Sprintf(meteringv1alpha1.ErrScopeNotAllowed, nsScope))
				}
			}

			if q.ResourceFilter == "" {
				return qo, errors.New(fmt.Sprintf(meteringv1alpha1.ErrScopeNotAllowed, nsScope))
			}
		}

		if q.NamespaceName != "" {
			nsOption.ResourceFilter = q.NamespaceName
		} else {
			var nsList *api.ListResult
			qu := query.New()
			qu.LabelSelector = q.LabelSelector
			nsList, err = t.ListNamespaces(user, q.WorkspaceName, qu)
			if err != nil {
				return qo, err
			}

			targetNs := []string{}
			for _, item := range nsList.Items {
				ns := item.(*corev1.Namespace)
				if ok, _ := regexp.MatchString(q.ResourceFilter, ns.ObjectMeta.GetName()); ok {
					listPods = authorizer.AttributesRecord{
						User:            user,
						Verb:            "list",
						Resource:        "pods",
						Namespace:       ns.ObjectMeta.GetName(),
						ResourceScope:   request.NamespaceScope,
						ResourceRequest: true,
					}
					decision, _, err = t.authorizer.Authorize(listPods)
					if err != nil {
						klog.Error(err)
						return
					}
					if decision == authorizer.DecisionAllow {
						targetNs = append(targetNs, ns.ObjectMeta.GetName())
					}
				}
			}
			nsOption.ResourceFilter = strings.Join(targetNs, "|")
		}

		qo.Option = nsOption

	case monitoring.LevelApplication:
		qo.Identifier = monitoringmodel.IdentifierApplication
		if q.NamespaceName == "" {
			return qo, errors.New(fmt.Sprintf(meteringv1alpha1.ErrParameterNotfound, "namespace"))
		}

		appScope := request.NamespaceScope

		listPods := authorizer.AttributesRecord{
			User:            user,
			Verb:            "list",
			APIGroup:        "",
			APIVersion:      "v1",
			Resource:        "pods",
			ResourceRequest: true,
			Namespace:       q.NamespaceName,
			ResourceScope:   appScope,
		}
		decision, _, err = t.authorizer.Authorize(listPods)
		if err != nil {
			klog.Error(err)
			return
		}
		if decision != authorizer.DecisionAllow {
			return qo, errors.New(fmt.Sprintf(meteringv1alpha1.ErrScopeNotAllowed, appScope))
		}

		qo.Option = monitoring.ApplicationsOption{
			NamespaceName:    q.NamespaceName,
			Applications:     strings.Split(q.Applications, "|"),
			StorageClassName: q.StorageClassName,
		}
		qo.NamedMetrics = monitoringmodel.ApplicationMetrics

	case monitoring.LevelWorkload:
		qo.Identifier = monitoringmodel.IdentifierWorkload
		if q.NamespaceName == "" {
			return qo, errors.New(fmt.Sprintf(meteringv1alpha1.ErrParameterNotfound, "namespace"))
		}

		qo.Option = monitoring.WorkloadOption{
			ResourceFilter: q.ResourceFilter, // workload filter
			NamespaceName:  q.NamespaceName,
			WorkloadKind:   q.WorkloadKind,
		}

		workloadScope := request.NamespaceScope

		listPods := authorizer.AttributesRecord{
			User:            user,
			Verb:            "list",
			APIGroup:        "",
			APIVersion:      "v1",
			Resource:        "pods",
			ResourceRequest: true,
			Namespace:       q.NamespaceName,
			ResourceScope:   workloadScope,
		}
		decision, _, err = t.authorizer.Authorize(listPods)
		if err != nil {
			klog.Error(err)
			return
		}
		if decision != authorizer.DecisionAllow {
			return qo, errors.New(fmt.Sprintf(meteringv1alpha1.ErrScopeNotAllowed, workloadScope))
		}

		qo.NamedMetrics = monitoringmodel.WorkloadMetrics

	case monitoring.LevelPod:
		qo.Identifier = monitoringmodel.IdentifierPod
		qo.Option = monitoring.PodOption{
			ResourceFilter: q.ResourceFilter,
			NodeName:       q.NodeName,
			NamespaceName:  q.NamespaceName,
			WorkloadKind:   q.WorkloadKind,
			WorkloadName:   q.WorkloadName,
			PodName:        q.PodName,
		}

		podScope := request.NamespaceScope

		listPods := authorizer.AttributesRecord{
			User:            user,
			Verb:            "list",
			APIGroup:        "",
			APIVersion:      "v1",
			Resource:        "pods",
			ResourceRequest: true,
			Namespace:       q.NamespaceName,
			ResourceScope:   podScope,
		}
		decision, _, err = t.authorizer.Authorize(listPods)
		if err != nil {
			klog.Error(err)
			return
		}
		if decision != authorizer.DecisionAllow {
			return qo, errors.New(fmt.Sprintf(meteringv1alpha1.ErrScopeNotAllowed, podScope))
		}

		qo.NamedMetrics = monitoringmodel.PodMetrics

	case monitoring.LevelService:
		qo.Identifier = monitoringmodel.IdentifierService
		qo.Option = monitoring.ServicesOption{
			NamespaceName: q.NamespaceName,
			Services:      strings.Split(q.Services, "|"),
		}

		if q.NamespaceName == "" {
			return qo, errors.New(fmt.Sprintf(meteringv1alpha1.ErrParameterNotfound, "namespace"))
		}
		serviceScope := request.NamespaceScope

		listPods := authorizer.AttributesRecord{
			User:            user,
			Verb:            "list",
			APIGroup:        "",
			APIVersion:      "v1",
			Resource:        "pods",
			ResourceRequest: true,
			Namespace:       q.NamespaceName,
			ResourceScope:   serviceScope,
		}
		decision, _, err = t.authorizer.Authorize(listPods)
		if err != nil {
			klog.Error(err)
			return
		}
		if decision != authorizer.DecisionAllow {
			return qo, errors.New(fmt.Sprintf(meteringv1alpha1.ErrScopeNotAllowed, serviceScope))
		}

		// TODO: list services if q.Services are empty

		qo.NamedMetrics = monitoringmodel.ServiceMetrics

	default:
		return qo, errors.New(fmt.Sprintf(meteringv1alpha1.ErrParameterNotfound, "level"))
	}

	// Parse time params
	if q.Start != "" && q.End != "" {
		startInt, err := strconv.ParseInt(q.Start, 10, 64)
		if err != nil {
			return qo, err
		}
		qo.Start = time.Unix(startInt, 0)

		endInt, err := strconv.ParseInt(q.End, 10, 64)
		if err != nil {
			return qo, err
		}
		qo.End = time.Unix(endInt, 0)

		if q.Step == "" {
			qo.Step = meteringv1alpha1.DefaultStep
		} else {
			qo.Step, err = time.ParseDuration(q.Step)
			if err != nil {
				return qo, err
			}
		}

		if qo.Start.After(qo.End) {
			return qo, errors.New(meteringv1alpha1.ErrInvalidStartEnd)
		}
	} else if q.Start == "" && q.End == "" {
		if q.Time == "" {
			qo.Time = time.Now()
		} else {
			timeInt, err := strconv.ParseInt(q.Time, 10, 64)
			if err != nil {
				return qo, err
			}
			qo.Time = time.Unix(timeInt, 0)
		}
	} else {
		return qo, errors.Errorf(meteringv1alpha1.ErrParamConflict)
	}

	if q.NamespaceName != "" {

		queryParameter := query.New()
		queryParameter.Filters[query.FieldName] = query.Value(q.NamespaceName)

		listResult, err := t.ListNamespaces(user, q.WorkspaceName, queryParameter)
		if err != nil {
			return qo, err
		}
		if listResult.TotalItems == 0 {
			return qo, errors.New(meteringv1alpha1.ErrResourceNotfound)
		}
		ns := listResult.Items[0].(*corev1.Namespace)
		cts := ns.CreationTimestamp.Time

		// Query should happen no earlier than namespace's creation time.
		// For range query, check and mutate `start`. For instant query, check `time`.
		// In range query, if `start` and `end` are both before namespace's creation time, it causes no hit.
		if !qo.isRangeQuery() {
			if qo.Time.Before(cts) {
				return qo, errors.New(meteringv1alpha1.ErrNoHit)
			}
		} else {
			if qo.End.Before(cts) {
				return qo, errors.New(meteringv1alpha1.ErrNoHit)
			}
			if qo.Start.Before(cts) {
				qo.Start = qo.End
				for qo.Start.Add(-qo.Step).After(cts) {
					qo.Start = qo.Start.Add(-qo.Step)
				}
			}
		}
	}

	// Parse sorting and paging params
	if q.Target != "" {
		qo.Target = q.Target
		qo.Page = meteringv1alpha1.DefaultPage
		qo.Limit = meteringv1alpha1.DefaultLimit
		qo.Order = q.Order
		if q.Order != monitoringmodel.OrderAscending {
			qo.Order = meteringv1alpha1.DefaultOrder
		}
		if q.Page != "" {
			qo.Page, err = strconv.Atoi(q.Page)
			if err != nil || qo.Page <= 0 {
				return qo, errors.New(meteringv1alpha1.ErrInvalidPage)
			}
		}
		if q.Limit != "" {
			qo.Limit, err = strconv.Atoi(q.Limit)
			if err != nil || qo.Limit <= 0 {
				return qo, errors.New(meteringv1alpha1.ErrInvalidLimit)
			}
		}
	}

	return qo, nil
}

func (t *tenantOperator) ProcessNamedMetersQuery(q QueryOptions) (metrics monitoringmodel.Metrics, err error) {
	var meters []string
	for _, meter := range q.NamedMetrics {
		if !strings.HasPrefix(meter, monitoringmodel.MetricMeterPrefix) {
			// skip non-meter metric
			continue
		}

		ok, _ := regexp.MatchString(q.MetricFilter, meter)
		if ok {
			meters = append(meters, meter)
		}
	}

	if len(meters) == 0 {
		klog.Info("no meters found")
		return
	}

	_, ok := q.Option.(monitoring.ApplicationsOption)
	if ok {
		metrics, err = t.processApplicationMetersQuery(meters, q)
		return
	}

	_, ok = q.Option.(monitoring.ServicesOption)
	if ok {
		metrics, err = t.processServiceMetersQuery(meters, q)
		return
	}

	if q.isRangeQuery() {
		metrics, err = t.mo.GetNamedMetersOverTime(meters, q.Start, q.End, q.Step, q.Option)
	} else {
		metrics, err = t.mo.GetNamedMeters(meters, q.Time, q.Option)
		if q.shouldSort() {
			metrics = *metrics.Sort(q.Target, q.Order, q.Identifier).Page(q.Page, q.Limit)
		}
	}

	return
}

func getMetricPosMap(metrics []monitoring.Metric) map[string]int {
	var metricMap = make(map[string]int)

	for i, m := range metrics {
		metricMap[m.MetricName] = i
	}

	return metricMap
}

func (t *tenantOperator) processApplicationMetersQuery(meters []string, q QueryOptions) (res monitoringmodel.Metrics, err error) {
	var metricMap = make(map[string]int)
	var current_res monitoringmodel.Metrics

	aso, ok := q.Option.(monitoring.ApplicationsOption)
	if !ok {
		err = errors.New("invalid application option")
		klog.Error(err.Error())
		return
	}
	componentsMap := t.mo.GetAppComponentsMap(aso.NamespaceName, aso.Applications)

	for k, _ := range componentsMap {
		opt := monitoring.ApplicationOption{
			NamespaceName:         aso.NamespaceName,
			Application:           k,
			ApplicationComponents: componentsMap[k],
			StorageClassName:      aso.StorageClassName,
		}

		if q.isRangeQuery() {
			current_res, err = t.mo.GetNamedMetersOverTime(meters, q.Start, q.End, q.Step, opt)
		} else {
			current_res, err = t.mo.GetNamedMeters(meters, q.Time, opt)
		}

		if res.Results == nil {
			res = current_res
			metricMap = getMetricPosMap(res.Results)
		} else {
			for _, cur_res := range current_res.Results {
				pos, ok := metricMap[cur_res.MetricName]
				if ok {
					res.Results[pos].MetricValues = append(res.Results[pos].MetricValues, cur_res.MetricValues...)
				} else {
					res.Results = append(res.Results, cur_res)
				}
			}
		}
	}

	if !q.isRangeQuery() && q.shouldSort() {
		res = *res.Sort(q.Target, q.Order, q.Identifier).Page(q.Page, q.Limit)
	}

	return
}

func (t *tenantOperator) processServiceMetersQuery(meters []string, q QueryOptions) (res monitoringmodel.Metrics, err error) {
	var metricMap = make(map[string]int)
	var current_res monitoringmodel.Metrics

	sso, ok := q.Option.(monitoring.ServicesOption)
	if !ok {
		err = errors.New("invalid service option")
		klog.Error(err.Error())
		return
	}
	svcPodsMap := t.mo.GetSerivePodsMap(sso.NamespaceName, sso.Services)

	for k, _ := range svcPodsMap {
		opt := monitoring.ServiceOption{
			NamespaceName: sso.NamespaceName,
			ServiceName:   k,
			PodNames:      svcPodsMap[k],
		}

		if q.isRangeQuery() {
			current_res, err = t.mo.GetNamedMetersOverTime(meters, q.Start, q.End, q.Step, opt)
		} else {
			current_res, err = t.mo.GetNamedMeters(meters, q.Time, opt)
		}

		if res.Results == nil {
			res = current_res
			metricMap = getMetricPosMap(res.Results)
		} else {
			for _, cur_res := range current_res.Results {
				pos, ok := metricMap[cur_res.MetricName]
				if ok {
					res.Results[pos].MetricValues = append(res.Results[pos].MetricValues, cur_res.MetricValues...)
				} else {
					res.Results = append(res.Results, cur_res)
				}
			}
		}
	}

	if !q.isRangeQuery() && q.shouldSort() {
		res = *res.Sort(q.Target, q.Order, q.Identifier).Page(q.Page, q.Limit)
	}

	return
}

func (t *tenantOperator) transformMetricData(metrics monitoringmodel.Metrics) metering.PodsStats {
	podsStats := make(metering.PodsStats)

	for _, metric := range metrics.Results {
		metricName := metric.MetricName
		for _, metricValue := range metric.MetricValues {
			//metricValue.SumValue
			podName := metricValue.Metadata["pod"]
			podsStats.Set(podName, metricName, metricValue.SumValue)
		}
	}

	return podsStats
}

func (t *tenantOperator) classifyPodStats(user user.Info, ns string, podsStats metering.PodsStats) (resourceStats metering.ResourceStatistic, err error) {

	if err = t.updateServicesStats(user, ns, podsStats, &resourceStats); err != nil {
		return
	}

	if err = t.updateDeploysStats(user, ns, podsStats, &resourceStats); err != nil {
		return
	}

	if err = t.updateDaemonsetsStats(user, ns, podsStats, &resourceStats); err != nil {
		return
	}

	if err = t.updateStatefulsetsStats(user, ns, podsStats, &resourceStats); err != nil {
		return
	}

	return
}

func (t *tenantOperator) updateServicesStats(user user.Info, ns string, podsStats metering.PodsStats, resourceStats *metering.ResourceStatistic) error {

	svcList, err := t.listServices(user, ns)
	if err != nil {
		return err
	}

	for _, svc := range svcList.Items {
		if svc.Annotations[constants.ApplicationReleaseName] != "" &&
			svc.Annotations[constants.ApplicationReleaseNS] != "" &&
			t.isOpNamespace(ns) {
			// for op svc
			// currently we do NOT include op svc
			continue
		} else {
			appName, nameOK := svc.Labels[constants.ApplicationName]
			appVersion, versionOK := svc.Labels[constants.ApplicationVersion]

			svcPodsMap := t.mo.GetSerivePodsMap(ns, []string{svc.Name})
			pods := svcPodsMap[svc.Name]

			if nameOK && versionOK {
				// for app crd svc
				for _, pod := range pods {
					podStat := podsStats[pod]
					if podStat == nil {
						klog.Warningf("%v not found", pod)
						continue
					}

					appFullName := appName + ":" + appVersion
					if err := resourceStats.GetAppStats(appFullName).GetServiceStats(svc.Name).SetPodStats(pod, podsStats[pod]); err != nil {
						klog.Error(err)
						return err
					}
				}
			} else {
				// for k8s svc
				for _, pod := range pods {
					if err := resourceStats.GetServiceStats(svc.Name).SetPodStats(pod, podsStats[pod]); err != nil {
						klog.Error(err)
						return err
					}
				}
			}
		}
	}

	// aggregate svc data
	for _, app := range resourceStats.Apps {
		for _, svc := range app.Services {
			svc.Aggregate()
		}
		app.Aggregate()
	}

	for _, svc := range resourceStats.Services {
		svc.Aggregate()
	}

	return nil
}

func (t *tenantOperator) listServices(user user.Info, ns string) (*corev1.ServiceList, error) {

	svcScope := request.NamespaceScope

	listSvc := authorizer.AttributesRecord{
		User:            user,
		Verb:            "list",
		Resource:        "services",
		Namespace:       ns,
		ResourceRequest: true,
		ResourceScope:   svcScope,
	}

	decision, _, err := t.authorizer.Authorize(listSvc)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if decision != authorizer.DecisionAllow {
		_, err := t.am.ListRoleBindings(user.GetName(), nil, ns)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
	}

	svcs, err := t.k8sclient.CoreV1().Services(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return svcs, nil
}

func (t *tenantOperator) updateDeploysStats(user user.Info, ns string, podsStats metering.PodsStats, resourceStats *metering.ResourceStatistic) error {
	deployList, err := t.listDeploys(user, ns)
	if err != nil {
		return err
	}

	for _, deploy := range deployList.Items {

		if deploy.Annotations[constants.ApplicationReleaseName] != "" &&
			deploy.Annotations[constants.ApplicationReleaseNS] != "" &&
			t.isOpNamespace(ns) {
			// for op deploy
			// currently we do NOT include op deploy
			continue
		} else {
			_, appNameOK := deploy.Labels[constants.ApplicationName]
			_, appVersionOK := deploy.Labels[constants.ApplicationVersion]

			pods, err := t.listPods(user, ns, deploy.Spec.Selector)
			if err != nil {
				klog.Error(err)
				return err
			}

			if appNameOK && appVersionOK {
				// for app crd svc
				continue
			} else {
				// for k8s svc
				for _, pod := range pods {
					if err := resourceStats.GetDeployStats(deploy.Name).SetPodStats(pod, podsStats[pod]); err != nil {
						klog.Error(err)
						return err
					}
				}
			}
		}
	}

	for _, deploy := range resourceStats.Deploys {
		deploy.Aggregate()
	}
	return nil
}

func (t *tenantOperator) updateDaemonsetsStats(user user.Info, ns string, podsStats metering.PodsStats, resourceStats *metering.ResourceStatistic) error {
	daemonsetList, err := t.listDaemonsets(user, ns)
	if err != nil {
		return err
	}

	for _, daemonset := range daemonsetList.Items {

		if daemonset.Annotations["meta.helm.sh/release-name"] != "" &&
			daemonset.Annotations["meta.helm.sh/release-namespace"] != "" &&
			t.isOpNamespace(ns) {
			// for op deploy
			// currently we do NOT include op deploy
			continue
		} else {
			appName := daemonset.Labels[constants.ApplicationName]
			appVersion := daemonset.Labels[constants.ApplicationVersion]

			pods, err := t.listPods(user, ns, daemonset.Spec.Selector)
			if err != nil {
				klog.Error(err)
				return err
			}

			if appName != "" && appVersion != "" {
				// for app crd svc
				continue
			} else {
				// for k8s svc
				for _, pod := range pods {
					if err := resourceStats.GetDaemonsetStats(daemonset.Name).SetPodStats(pod, podsStats[pod]); err != nil {
						klog.Error(err)
						return err
					}
				}
			}
		}
	}

	for _, daemonset := range resourceStats.Daemonsets {
		daemonset.Aggregate()
	}
	return nil
}

func (t *tenantOperator) isOpNamespace(ns string) bool {

	nsObj, err := t.k8sclient.CoreV1().Namespaces().Get(context.Background(), ns, metav1.GetOptions{})
	if err != nil {
		return false
	}

	ws := nsObj.Labels[constants.WorkspaceLabelKey]

	if len(ws) != 0 && ws != "system-workspace" {
		return true
	}
	return false
}

func (t *tenantOperator) updateStatefulsetsStats(user user.Info, ns string, podsStats metering.PodsStats, resourceStats *metering.ResourceStatistic) error {
	statefulsetsList, err := t.listStatefulsets(user, ns)
	if err != nil {
		return err
	}

	for _, statefulset := range statefulsetsList.Items {

		if statefulset.Annotations[constants.ApplicationReleaseName] != "" &&
			statefulset.Annotations[constants.ApplicationReleaseNS] != "" &&
			t.isOpNamespace(ns) {
			// for op deploy
			// currently we do NOT include op deploy
			continue
		} else {
			appName := statefulset.Labels[constants.ApplicationName]
			appVersion := statefulset.Labels[constants.ApplicationVersion]

			pods, err := t.listPods(user, ns, statefulset.Spec.Selector)
			if err != nil {
				klog.Error(err)
				return err
			}

			if appName != "" && appVersion != "" {
				// for app crd svc
				continue
			} else {
				// for k8s svc
				for _, pod := range pods {
					if err := resourceStats.GetStatefulsetStats(statefulset.Name).SetPodStats(pod, podsStats[pod]); err != nil {
						klog.Error(err)
						return err
					}
				}
			}
		}
	}

	for _, statefulset := range resourceStats.Statefulsets {
		statefulset.Aggregate()
	}
	return nil
}

func (t *tenantOperator) listPods(user user.Info, ns string, selector *metav1.LabelSelector) ([]string, error) {
	podScope := request.NamespaceScope

	listPods := authorizer.AttributesRecord{
		User:            user,
		Verb:            "list",
		Resource:        "pods",
		ResourceRequest: true,
		Namespace:       ns,
		ResourceScope:   podScope,
	}

	decision, _, err := t.authorizer.Authorize(listPods)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if decision != authorizer.DecisionAllow {
		_, err := t.am.ListRoleBindings(user.GetName(), nil, ns)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
	}

	var labelFilter []string
	for k, v := range selector.MatchLabels {
		labelFilter = append(labelFilter, fmt.Sprintf("%v=%v", k, v))
	}

	opt := metav1.ListOptions{LabelSelector: strings.Join(labelFilter, ",")}

	pods, err := t.k8sclient.CoreV1().Pods(ns).List(context.Background(), opt)
	if err != nil {
		return nil, err
	}

	ret := []string{}
	for _, pod := range pods.Items {
		ret = append(ret, pod.Name)
	}

	return ret, nil
}

func (t *tenantOperator) listDeploys(user user.Info, ns string) (*appv1.DeploymentList, error) {

	deployScope := request.NamespaceScope

	listSvc := authorizer.AttributesRecord{
		User:            user,
		Verb:            "list",
		Resource:        "deployments",
		ResourceRequest: true,
		Namespace:       ns,
		ResourceScope:   deployScope,
	}

	decision, _, err := t.authorizer.Authorize(listSvc)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if decision != authorizer.DecisionAllow {
		_, err := t.am.ListRoleBindings(user.GetName(), nil, ns)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
	}

	deploys, err := t.k8sclient.AppsV1().Deployments(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return deploys, nil
}

func (t *tenantOperator) listDaemonsets(user user.Info, ns string) (*appv1.DaemonSetList, error) {

	dsScope := request.NamespaceScope

	listSvc := authorizer.AttributesRecord{
		User:            user,
		Verb:            "list",
		Resource:        "daemonsets",
		ResourceRequest: true,
		Namespace:       ns,
		ResourceScope:   dsScope,
	}

	decision, _, err := t.authorizer.Authorize(listSvc)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if decision != authorizer.DecisionAllow {
		_, err := t.am.ListRoleBindings(user.GetName(), nil, ns)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
	}

	ds, err := t.k8sclient.AppsV1().DaemonSets(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func (t *tenantOperator) listStatefulsets(user user.Info, ns string) (*appv1.StatefulSetList, error) {

	stsScope := request.NamespaceScope

	listSvc := authorizer.AttributesRecord{
		User:            user,
		Verb:            "list",
		Resource:        "statefulsets",
		Namespace:       ns,
		ResourceRequest: true,
		ResourceScope:   stsScope,
	}

	decision, _, err := t.authorizer.Authorize(listSvc)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if decision != authorizer.DecisionAllow {
		_, err := t.am.ListRoleBindings(user.GetName(), nil, ns)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
	}

	stss, err := t.k8sclient.AppsV1().StatefulSets(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return stss, nil
}
