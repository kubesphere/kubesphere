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

// List persistent volume claims of a specific storage class
func GetPvcListBySc(storageclass string) (res []v12.PersistentVolumeClaim, err error) {
	// Create Kubernetes client
	cli := client.NewK8sClient()
	// Get all persistent volume claims
	claimList, err := cli.CoreV1().PersistentVolumeClaims("").List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	// Select persistent volume claims which
	// storage class name is equal to the specific storage class.
	for _, claim := range claimList.Items {
		if claim.Spec.StorageClassName != nil &&
			*claim.Spec.StorageClassName == storageclass {
			res = append(res, claim)
		} else {
			continue
		}
	}
	return res, nil
}

// Get metrics of a specific storage class
func GetScMetrics(storageclass string) (res StorageMetrics, err error) {
	// Create Kubernetes client
	cli := client.NewK8sClient()
	// Get persistent volumes
	pvList, err := cli.CoreV1().PersistentVolumes().List(v1.ListOptions{})
	if err != nil {
		return StorageMetrics{}, err
	}

	var total resource.Quantity
	// Gathering metrics of a specific storage class
	for _, volume := range pvList.Items {
		if volume.Spec.StorageClassName != storageclass {
			continue
		}
		total.Add(volume.Spec.Capacity[v12.ResourceStorage])
	}
	return StorageMetrics{Usage: total.String()}, nil
}
