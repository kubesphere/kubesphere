package models

import (
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
	"net/http"
)

type pvcListBySc struct {
	Name   string                      `json:"name"`
	Claims []v12.PersistentVolumeClaim `json:"persistentvolumeclaims"`
}

type scMetrics struct {
	Name    string         `json:"name"`
	Metrics storageMetrics `json:"metrics"`
}

type storageMetrics struct {
	Capacity string `json:"capacity,omitempty"`
	Usage    string `json:"usage,omitempty"`
}

// List all PersistentVolumeClaims of a specific StorageClass
// Extended API URL: "GET /api/v1alpha/storage/storageclasses/{name}/persistentvolumeclaims"
func GetPvcListBySc(request *restful.Request, response *restful.Response) {

	scName := request.PathParameter("storageclass")
	glog.Infof("Run GetPvcListBySc: SC = %s", scName)
	claims, err := getPvcListBySc(scName)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	result := constants.ResultMessage{
		Kind:       constants.KIND,
		ApiVersion: constants.APIVERSION,
		Data:       pvcListBySc{scName, claims}}

	response.WriteAsJson(result)
}

// Get metrics of a specific StorageClass
// Extended API URL: "GET /api/v1alpha/storage/storageclasses/{name}/metrics"
func GetScMetrics(request *restful.Request, response *restful.Response) {
	scName := request.PathParameter("storageclass")
	glog.Infof("Run GetPvcListBySc: SC = %s", scName)

	metrics, err := getScMetrics(scName)
	if err != nil {
		response.WriteError(http.StatusInternalServerError, err)
	}
	result := constants.ResultMessage{
		Kind:       constants.KIND,
		ApiVersion: constants.APIVERSION,
		Data:       scMetrics{Name: scName, Metrics: metrics},
	}
	response.WriteAsJson(result)
}

func getPvcListBySc(storageclass string) (res []v12.PersistentVolumeClaim, err error) {

	cli := client.NewK8sClient()
	claimList, err := cli.CoreV1().PersistentVolumeClaims("").List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, claim := range claimList.Items {
		if *claim.Spec.StorageClassName != storageclass {
			continue
		}
		res = append(res, claim)
	}
	return res, nil
}

func getScMetrics(storageclass string) (res storageMetrics, err error) {
	cli := client.NewK8sClient()
	pvList, err := cli.CoreV1().PersistentVolumes().List(v1.ListOptions{})
	if err != nil {
		return storageMetrics{}, err
	}

	var total resource.Quantity
	for _, volume := range pvList.Items {
		if volume.Spec.StorageClassName != storageclass {
			continue
		}
		total.Add(volume.Spec.Capacity[v12.ResourceStorage])
	}
	return storageMetrics{Usage: total.String()}, nil
}
