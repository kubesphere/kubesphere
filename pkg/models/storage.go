package models

import (
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
)

type storageListClaims struct {
	Claims []simpleClaimList `json:"persistentvolumeclaims"`
	Name   string            `json:"name"`
}

type simpleClaimList struct {
	Claim     string `json:"name"`
	Namespace string `json:"namespace"`
}

func GetStorageListClaims(request *restful.Request, response *restful.Response) {

	scName := request.PathParameter("storageclass")
	glog.Infof("Run GetStorageListClaims, StorageClass: %s", scName)

	claims := getStorageListClaims(scName)

	result := constants.ResultMessage{
		Kind:       constants.KIND,
		ApiVersion: constants.APIVERSION,
		Data:       storageListClaims{Claims: claims, Name: scName}}

	response.WriteAsJson(result)
}

func getStorageListClaims(storageclass string) (res []simpleClaimList) {

	cli := client.NewK8sClient()
	claimList, err := cli.CoreV1().PersistentVolumeClaims("").List(v1.ListOptions{})
	if err != nil {
		glog.Error("Read all PVC err: ", err)
		return nil
	}

	for _, claim := range claimList.Items {
		if *claim.Spec.StorageClassName != storageclass {
			continue
		}
		res = append(res, simpleClaimList{Claim: claim.Name, Namespace: claim.Namespace})
	}
	return res
}
