package models

import (
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
	"net/http"
)

type pvcListBySc struct {
	Claims []simplePvcList `json:"persistentvolumeclaims"`
	Name   string          `json:"name"`
}

type simplePvcList struct {
	Claim     string `json:"name"`
	Namespace string `json:"namespace"`
}

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
		Data:       pvcListBySc{Claims: claims, Name: scName}}

	response.WriteAsJson(result)
}

func getPvcListBySc(storageclass string) (res []simplePvcList, err error) {

	cli := client.NewK8sClient()
	claimList, err := cli.CoreV1().PersistentVolumeClaims("").List(v1.ListOptions{})
	if err != nil {
		glog.Error("Read all PVC err: ", err)
		return nil, err
	}

	for _, claim := range claimList.Items {
		if *claim.Spec.StorageClassName != storageclass {
			continue
		}
		res = append(res, simplePvcList{Claim: claim.Name, Namespace: claim.Namespace})
	}
	return res, nil
}
