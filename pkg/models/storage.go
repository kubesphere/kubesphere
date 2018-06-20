package models

import (
	"strconv"

	"github.com/golang/glog"
	v12 "k8s.io/api/core/v1"
	v13 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/kubesphere/pkg/client"
)

type PvcListBySc struct {
	Name   string                      `json:"name"`
	Claims []v12.PersistentVolumeClaim `json:"items"`
}

type ScMetrics struct {
	Capacity  string `json:"capacity,omitempty"`
	Usage     string `json:"usage,omitempty"`
	PvcNumber string `json:"pvcNumber"`
}

// StorageClass metrics item
type ScMetricsItem struct {
	Name    string    `json:"name"`
	Metrics ScMetrics `json:"metrics"`
}

// StorageClass metrics items list
type ScMetricsItemList struct {
	Items []ScMetricsItem `json:"items"`
}

// List persistent volume claims of a specific storage class
func GetPvcListBySc(storageclass string) (res []v12.PersistentVolumeClaim, err error) {
	// Create Kubernetes client
	cli := client.NewK8sClient()
	// Get all persistent volume claims
	claimList, err := cli.CoreV1().PersistentVolumeClaims("").List(v1.ListOptions{})
	if err != nil {
		glog.Errorf("list persistent volumes error: name: \"%s\", error msg: \"%s\"", storageclass, err.Error())
		return nil, err
	}
	// Select persistent volume claims which
	// storage class name is equal to the specific storage class.
	for _, claim := range claimList.Items {
		if claim.Spec.StorageClassName != nil {
			if *claim.Spec.StorageClassName == storageclass {
				res = append(res, claim)
			}
		} else if claim.GetAnnotations()[v12.BetaStorageClassAnnotation] == storageclass {
			res = append(res, claim)
		}
	}
	return res, nil
}

// Get info of metrics
func GetScEntityMetrics(scname string) (ScMetrics, error) {
	// Create Kubernetes client
	cli := client.NewK8sClient()

	// Get PV
	pvList, err := cli.CoreV1().PersistentVolumes().List(v1.ListOptions{})
	if err != nil {
		glog.Errorf("list persistent volume request error: error msg: \"%s\"", err.Error())
		return ScMetrics{}, err
	}
	// Get PVC
	pvcList, err := GetPvcListBySc(scname)
	if err != nil {
		return ScMetrics{}, err
	}

	// Get storage usage
	// Gathering usage of a specific StorageClass
	var total resource.Quantity
	for _, volume := range pvList.Items {
		if volume.Spec.StorageClassName != scname {
			continue
		}
		total.Add(volume.Spec.Capacity[v12.ResourceStorage])
	}
	usage := total.String()

	// Get PVC number
	pvcNum := len(pvcList)

	return ScMetrics{Usage: usage, PvcNumber: strconv.Itoa(pvcNum)}, nil
}

// Get raw information of a SC
func GetScEntity(scname string) (res v13.StorageClass, err error) {
	// Create Kubernetes client
	cli := client.NewK8sClient()
	// Get SC
	sc, err := cli.StorageV1().StorageClasses().Get(scname, v1.GetOptions{})
	if err != nil {
		glog.Errorf("get storage class request error: name: \"%s\", error msg: \"%s\"", scname, err.Error())
		return v13.StorageClass{}, err
	}
	return *sc, nil
}

// Get SC item
func GetScItemMetrics(scname string) (res ScMetricsItem, err error) {
	// Check SC exist
	_, err = GetScEntity(scname)
	if err != nil {
		return ScMetricsItem{}, err
	}

	metrics, err := GetScEntityMetrics(scname)
	if err != nil {
		return ScMetricsItem{}, err
	}

	result := ScMetricsItem{scname, metrics}
	return result, nil
}

// Get SC item list
func GetScItemMetricsList() (res ScMetricsItemList, err error) {
	// Create Kubernetes client
	cli := client.NewK8sClient()
	// Get StorageClass list
	scList, err := cli.StorageV1().StorageClasses().List(v1.ListOptions{})
	if err != nil {
		glog.Errorf("list storage classes request error: error msg: \"%s\"", err.Error())
		return ScMetricsItemList{}, err
	}
	if scList == nil {
		return ScMetricsItemList{}, nil
	}
	// Set return value
	res = ScMetricsItemList{}
	for _, v := range scList.Items {
		item, err := GetScItemMetrics(v.GetName())
		if err != nil {
			return ScMetricsItemList{}, err
		}
		res.Items = append(res.Items, item)
	}
	return res, nil
}
