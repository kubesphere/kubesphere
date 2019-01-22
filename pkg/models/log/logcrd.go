/*
Copyright 2018 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package log

import (
	"github.com/emicklei/go-restful"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/client"
)

type CRDResult struct {
	Status int `json:"status"`
	CRD client.FluentBitOperatorSpec `json:"CRD,omitempty"`
}

func CRDQuery(request *restful.Request) *CRDResult {
	var result CRDResult

	config, err := client.GetClientConfig("")
	if err != nil {
		//panic(err.Error())
		result.Status = 400
		return &result
	}

	// Create a new clientset which include our CRD schema
	crdcs, scheme, err := client.NewClient(config)
	if err != nil {
		//panic(err)
		result.Status = 400
		return &result
	}

	// Create a CRD client interface
	crdclient := client.CrdClient(crdcs, scheme, "default")

	item, err := crdclient.Get("fluent-bit")
	if err != nil {
		//panic(err)
		result.Status = 200
		return &result
	}

	result.CRD = item.Spec
	result.Status = 200

	return &result
}

func CRDUpdate(request *restful.Request) *CRDResult {
	var result CRDResult

	spec := new(client.FluentBitOperatorSpec)

	err := request.ReadEntity(&spec)
	if err != nil {
		//panic(err.Error())
		result.Status = 400
		return &result
	}

	config, err := client.GetClientConfig("")
	if err != nil {
		//panic(err.Error())
		result.Status = 400
		return &result
	}

	// Create a new clientset which include our CRD schema
	crdcs, scheme, err := client.NewClient(config)
	if err != nil {
		//panic(err)
		result.Status = 400
		return &result
	}

	// Create a CRD client interface
	crdclient := client.CrdClient(crdcs, scheme, "default")

	var item *client.FluentBitOperator
	var err_read error

	item, err_read = crdclient.Get("fluent-bit")
	if err_read != nil {
		//panic(err)
		fluentBitOperator := &client.FluentBitOperator{
			ObjectMeta: metav1.ObjectMeta{
				Name: "fluent-bit",
			},
			Spec: *spec,
		}

		itemnew, err := crdclient.Create(fluentBitOperator)
		if err != nil {
			//panic(err)
			result.Status = 400
			return &result
		}

		result.CRD = itemnew.Spec
		result.Status = 200
	} else {
		item.Spec = *spec

		itemnew, err := crdclient.Update("fluent-bit", item)
		if err != nil {
			//panic(err)
			result.Status = 400
			return &result
		}

		result.CRD = itemnew.Spec
		result.Status = 200
	}

	return &result
}
