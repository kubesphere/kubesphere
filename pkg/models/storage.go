package models

import (
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
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
	glog.Infof("Run GetStorageMetrics")
	scName := request.PathParameter("storageclass")

	claims := getStorageListClaims(scName)

	result := constants.ResultMessage{
		Kind:       constants.KIND,
		ApiVersion: constants.APIVERSION,
		Data:       storageListClaims{Claims: claims, Name: scName}}

	response.WriteAsJson(result)
}

func getStorageListClaims(storageclass string) []simpleClaimList {
	message := GetApiserver("/api/v1/persistentvolumeclaims")
	// read PVC lists
	items := jsonRawMessage(message).Find("items").ToList()

	var res []simpleClaimList
	for _, item := range items {
		curSC := item.Find("spec").Find("storageClassName").ToString()
		if curSC != storageclass {
			continue
		}
		pvc := item.Find("metadata").Find("name").ToString()
		ns := item.Find("metadata").Find("namespace").ToString()
		res = append(res, simpleClaimList{Claim: pvc, Namespace: ns})
	}
	return res
}
