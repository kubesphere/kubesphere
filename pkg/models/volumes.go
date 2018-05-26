package models

import (
	"github.com/emicklei/go-restful"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
	"net/http"
)

type podListByPvc struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Pods      []v12.Pod `json:"pods"`
}

// List all pods of a specific PVC
// Extended API URL: "GET /api/v1alpha/volumes/namespaces/{namespace}/persistentvolumeclaims/{name}/pods"
func GetPodListByPvc(request *restful.Request, response *restful.Response) {

	pvcName := request.PathParameter("pvc")
	nsName := request.PathParameter("namespace")
	pods, err := getPodListByPvc(pvcName, nsName)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	result := constants.ResultMessage{

		Data: podListByPvc{
			Name: pvcName, Namespace: nsName, Pods: pods}}

	response.WriteAsJson(result)
}

func getPodListByPvc(pvc string, ns string) (res []v12.Pod, err error) {
	cli := client.NewK8sClient()
	podList, err := cli.CoreV1().Pods(ns).List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pod := range podList.Items {
		if isPvcInPod(pod, pvc) == true {
			res = append(res, pod)
		}
	}
	return res, nil
}

func isPvcInPod(pod v12.Pod, pvcname string) bool {
	for _, v := range pod.Spec.Volumes {
		if v.VolumeSource.PersistentVolumeClaim != nil &&
			v.VolumeSource.PersistentVolumeClaim.ClaimName == pvcname {
			return true
		}
	}
	return false
}
