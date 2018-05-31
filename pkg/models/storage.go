package models

import (
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/client"
)

type PvcListBySc struct {
	Name   string                      `json:"name"`
	Claims []v12.PersistentVolumeClaim `json:"items"`
}

type ScMetrics struct {
	Name    string         `json:"name"`
	Metrics StorageMetrics `json:"metrics"`
}

type StorageMetrics struct {
	Capacity string `json:"capacity,omitempty"`
	Usage    string `json:"usage,omitempty"`
}

func GetPvcListBySc(storageclass string) (res []v12.PersistentVolumeClaim, err error) {

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

func GetScMetrics(storageclass string) (res StorageMetrics, err error) {
	cli := client.NewK8sClient()
	pvList, err := cli.CoreV1().PersistentVolumes().List(v1.ListOptions{})
	if err != nil {
		return StorageMetrics{}, err
	}

	var total resource.Quantity
	for _, volume := range pvList.Items {
		if volume.Spec.StorageClassName != storageclass {
			continue
		}
		total.Add(volume.Spec.Capacity[v12.ResourceStorage])
	}
	return StorageMetrics{Usage: total.String()}, nil
}
