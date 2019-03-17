package log

import (
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/labels"
	es "kubesphere.io/kubesphere/pkg/simple/client/elasticsearch"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"reflect"
	"strconv"
	"strings"
)

func intersection(s1, s2 []string) (inter []string) {
	hash := make(map[string]bool)
	for _, e := range s1 {
		hash[e] = true
	}
	for _, e := range s2 {
		// If elements present in the hashmap then append intersection list.
		if hash[e] {
			inter = append(inter, e)
		}
	}
	//Remove dups from slice.
	inter = removeDups(inter)
	return
}

//Remove dups from slice.
func removeDups(elements []string) (nodups []string) {
	encountered := make(map[string]bool)
	for _, element := range elements {
		if !encountered[element] {
			nodups = append(nodups, element)
			encountered[element] = true
		}
	}
	return
}

func in(value interface{}, container interface{}) int {
	if container == nil {
		return -1
	}
	containerValue := reflect.ValueOf(container)
	switch reflect.TypeOf(container).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < containerValue.Len(); i++ {
			if containerValue.Index(i).Interface() == value {
				return i
			}
		}
	case reflect.Map:
		if containerValue.MapIndex(reflect.ValueOf(value)).IsValid() {
			return -1
		}
	default:
		return -1
	}
	return -1
}

func getWorkloadName(name string, kind string) string {
	if kind == "ReplicaSet" {
		lastIndex := strings.LastIndex(name, "-")
		if lastIndex >= 0 {
			return name[:lastIndex]
		}
	}

	return name
}

func matchLabel(label string, labelsMatch []string) bool {
	var result = false

	for _, labelMatch := range labelsMatch {
		if strings.Compare(label, labelMatch) == 0 {
			result = true
			break
		}
	}

	return result
}

func queryLabel(label string, labelsQuery []string) bool {
	var result = false

	for _, labelQuery := range labelsQuery {
		if strings.Contains(label, labelQuery) {
			result = true
			break
		}
	}

	return result
}

func queryWorkspace(workspaceMatch string, workspaceQuery string) (bool, []string) {
	if workspaceMatch == "" && workspaceQuery == "" {
		return false, nil
	}

	nsLister := informers.SharedInformerFactory().Core().V1().Namespaces().Lister()
	nsList, err := nsLister.List(labels.Everything())
	if err != nil {
		glog.Error("failed to list namespace, error: ", err)
		return true, nil
	}

	var namespaces []string

	var hasMatch = false
	var workspacesMatch []string
	if workspaceMatch != "" {
		workspacesMatch = strings.Split(strings.Replace(workspaceMatch, ",", " ", -1), " ")
		hasMatch = true
	}

	var hasQuery = false
	var workspacesQuery []string
	if workspaceQuery != "" {
		workspacesQuery = strings.Split(strings.ToLower(strings.Replace(workspaceQuery, ",", " ", -1)), " ")
		hasQuery = true
	}

	for _, ns := range nsList {
		labels := ns.GetLabels()
		_, ok := labels[constants.WorkspaceLabelKey]
		if ok {
			var namespaceCanAppend = true
			if hasMatch {
				if !matchLabel(labels[constants.WorkspaceLabelKey], workspacesMatch) {
					namespaceCanAppend = false
				}
			}
			if hasQuery {
				if !queryLabel(strings.ToLower(labels[constants.WorkspaceLabelKey]), workspacesQuery) {
					namespaceCanAppend = false
				}
			}

			if namespaceCanAppend {
				namespaces = append(namespaces, ns.GetName())
			}
		}
	}

	return true, namespaces
}

func matchNamespace(namespaceMatch string, namespaceFilled bool, namespaces []string) (bool, []string) {
	if namespaceMatch == "" {
		return namespaceFilled, namespaces
	}

	namespacesMatch := strings.Split(strings.Replace(namespaceMatch, ",", " ", -1), " ")

	if namespaceFilled {
		return true, intersection(namespacesMatch, namespaces)
	}

	return true, namespacesMatch
}

func queryWorkload(workloadMatch string, workloadQuery string, namespaces []string) (bool, []string) {
	if workloadMatch == "" && workloadQuery == "" {
		return false, nil
	}

	podLister := informers.SharedInformerFactory().Core().V1().Pods().Lister()
	podList, err := podLister.List(labels.Everything())
	if err != nil {
		glog.Error("failed to list pods, error: ", err)
		return true, nil
	}

	var pods []string

	var hasMatch = false
	var workloadsMatch []string
	if workloadMatch != "" {
		workloadsMatch = strings.Split(strings.Replace(workloadMatch, ",", " ", -1), " ")
		hasMatch = true
	}

	var hasQuery = false
	var workloadsQuery []string
	if workloadQuery != "" {
		workloadsQuery = strings.Split(strings.ToLower(strings.Replace(workloadQuery, ",", " ", -1)), " ")
		hasQuery = true
	}

	if namespaces == nil {
		for _, pod := range podList {
			/*if len(pod.ObjectMeta.OwnerReferences) > 0 {
				glog.Infof("List Pod %v:%v:%v", pod.Name, pod.ObjectMeta.OwnerReferences[0].Name, pod.ObjectMeta.OwnerReferences[0].Kind)
			}*/
			if len(pod.ObjectMeta.OwnerReferences) > 0 {
				var podCanAppend = true
				workloadName := getWorkloadName(pod.ObjectMeta.OwnerReferences[0].Name, pod.ObjectMeta.OwnerReferences[0].Kind)
				if hasMatch {
					if !matchLabel(workloadName, workloadsMatch) {
						podCanAppend = false
					}
				}
				if hasQuery {
					if !queryLabel(strings.ToLower(workloadName), workloadsQuery) {
						podCanAppend = false
					}
				}

				if podCanAppend {
					pods = append(pods, pod.Name)
				}
			}
		}
	} else {
		for _, pod := range podList {
			/*if len(pod.ObjectMeta.OwnerReferences) > 0 {
				glog.Infof("List Pod %v:%v:%v", pod.Name, pod.ObjectMeta.OwnerReferences[0].Name, pod.ObjectMeta.OwnerReferences[0].Kind)
			}*/
			if len(pod.ObjectMeta.OwnerReferences) > 0 && in(pod.Namespace, namespaces) >= 0 {
				var podCanAppend = true
				workloadName := getWorkloadName(pod.ObjectMeta.OwnerReferences[0].Name, pod.ObjectMeta.OwnerReferences[0].Kind)
				if hasMatch {
					if !matchLabel(workloadName, workloadsMatch) {
						podCanAppend = false
					}
				}
				if hasQuery {
					if !queryLabel(strings.ToLower(workloadName), workloadsQuery) {
						podCanAppend = false
					}
				}

				if podCanAppend {
					pods = append(pods, pod.Name)
				}
			}
		}
	}

	return true, pods
}

func matchPod(podMatch string, podFilled bool, pods []string) (bool, []string) {
	if podMatch == "" {
		return podFilled, pods
	}

	podsMatch := strings.Split(strings.Replace(podMatch, ",", " ", -1), " ")

	if podFilled {
		return true, intersection(podsMatch, pods)
	}

	return true, podsMatch
}

func matchContainer(containerMatch string) (bool, []string) {
	if containerMatch == "" {
		return false, nil
	}

	return true, strings.Split(strings.Replace(containerMatch, ",", " ", -1), " ")
}

func getWorkspaceOfNamesapce(namespace string) string {
	var workspace string
	workspace = ""

	nsLister := informers.SharedInformerFactory().Core().V1().Namespaces().Lister()
	nsList, err := nsLister.List(labels.Everything())
	if err != nil {
		glog.Error("failed to list namespace, error: ", err)
		return workspace
	}

	for _, ns := range nsList {
		if ns.GetName() == namespace {
			labels := ns.GetLabels()
			_, ok := labels[constants.WorkspaceLabelKey]
			if ok {
				workspace = labels[constants.WorkspaceLabelKey]
			}
		}
	}

	return workspace
}

func LogQuery(level constants.LogQueryLevel, request *restful.Request) *es.QueryResult {
	var param es.QueryParameters

	param.Level = level
	param.Operation = request.QueryParameter("operation")

	switch level {
	case constants.QueryLevelCluster:
		{
			param.NamespaceFilled, param.Namespaces = queryWorkspace(request.QueryParameter("workspaces"), request.QueryParameter("workspace_query"))
			param.NamespaceFilled, param.Namespaces = matchNamespace(request.QueryParameter("namespaces"), param.NamespaceFilled, param.Namespaces)
			param.NamespaceQuery = request.QueryParameter("namespace_query")
			param.PodFilled, param.Pods = queryWorkload(request.QueryParameter("workloads"), request.QueryParameter("workload_query"), param.Namespaces)
			param.PodFilled, param.Pods = matchPod(request.QueryParameter("pods"), param.PodFilled, param.Pods)
			param.PodQuery = request.QueryParameter("pod_query")
			param.ContainerFilled, param.Containers = matchContainer(request.QueryParameter("containers"))
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case constants.QueryLevelWorkspace:
		{
			param.NamespaceFilled, param.Namespaces = queryWorkspace(request.PathParameter("workspace"), "")
			param.NamespaceFilled, param.Namespaces = matchNamespace(request.QueryParameter("namespaces"), param.NamespaceFilled, param.Namespaces)
			param.NamespaceQuery = request.QueryParameter("namespace_query")
			param.PodFilled, param.Pods = queryWorkload(request.QueryParameter("workloads"), request.QueryParameter("workload_query"), param.Namespaces)
			param.PodFilled, param.Pods = matchPod(request.QueryParameter("pods"), param.PodFilled, param.Pods)
			param.PodQuery = request.QueryParameter("pod_query")
			param.ContainerFilled, param.Containers = matchContainer(request.QueryParameter("containers"))
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case constants.QueryLevelNamespace:
		{
			param.NamespaceFilled, param.Namespaces = matchNamespace(request.PathParameter("namespace"), false, nil)
			param.PodFilled, param.Pods = queryWorkload(request.QueryParameter("workloads"), request.QueryParameter("workload_query"), param.Namespaces)
			param.PodFilled, param.Pods = matchPod(request.QueryParameter("pods"), param.PodFilled, param.Pods)
			param.PodQuery = request.QueryParameter("pod_query")
			param.ContainerFilled, param.Containers = matchContainer(request.QueryParameter("containers"))
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case constants.QueryLevelWorkload:
		{
			param.NamespaceFilled, param.Namespaces = matchNamespace(request.PathParameter("namespace"), false, nil)
			param.PodFilled, param.Pods = queryWorkload(request.PathParameter("workload"), "", param.Namespaces)
			param.PodFilled, param.Pods = matchPod(request.QueryParameter("pods"), param.PodFilled, param.Pods)
			param.PodQuery = request.QueryParameter("pod_query")
			param.ContainerFilled, param.Containers = matchContainer(request.QueryParameter("containers"))
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case constants.QueryLevelPod:
		{
			param.NamespaceFilled, param.Namespaces = matchNamespace(request.PathParameter("namespace"), false, nil)
			param.PodFilled, param.Pods = matchPod(request.PathParameter("pod"), false, nil)
			param.ContainerFilled, param.Containers = matchContainer(request.QueryParameter("containers"))
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case constants.QueryLevelContainer:
		{
			param.NamespaceFilled, param.Namespaces = matchNamespace(request.PathParameter("namespace"), false, nil)
			param.PodFilled, param.Pods = matchPod(request.PathParameter("pod"), false, nil)
			param.ContainerFilled, param.Containers = matchContainer(request.PathParameter("container"))
		}
	}

	if len(param.Namespaces) == 1 {
		param.Workspace = getWorkspaceOfNamesapce(param.Namespaces[0])
	}

	param.Interval = request.QueryParameter("interval")

	param.LogQuery = request.QueryParameter("log_query")
	param.StartTime = request.QueryParameter("start_time")
	param.EndTime = request.QueryParameter("end_time")
	param.Sort = request.QueryParameter("sort")

	var err error
	param.From, err = strconv.ParseInt(request.QueryParameter("from"), 10, 64)
	if err != nil {
		param.From = 0
	}
	param.Size, err = strconv.ParseInt(request.QueryParameter("size"), 10, 64)
	if err != nil {
		param.Size = 10
	}

	glog.Infof("LogQuery with %v", param)

	return es.Query(param)
}