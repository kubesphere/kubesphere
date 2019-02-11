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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	"kubesphere.io/kubesphere/pkg/client"
)

type CRDResult struct {
	Status int                          `json:"status"`
	CRD    client.FluentBitOperatorSpec `json:"CRD,omitempty"`
}

type CRDDeleteResult struct {
	Status int `json:"status"`
}

type EnableResult struct {
	Status int    `json:"status"`
	Enable string `json:"Enable,omitempty"`
}

func createCRDClientSet() (*rest.RESTClient, *runtime.Scheme, error) {
	config, err := client.GetClientConfig("")
	if err != nil {
		//panic(err.Error())
		return nil, nil, err
	}

	// Create a new clientset which include our CRD schema
	return client.NewClient(config)
}

func getParameterValue(parameters []client.Parameter, name string) string {
	var value string

	value = ""
	for _, parameter := range parameters {
		if parameter.Name == name {
			value = parameter.Value
		}
	}

	return value
}

func CRDQuery(request *restful.Request) *CRDResult {
	var result CRDResult

	crdcs, scheme, err := createCRDClientSet()
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

	crdcs, scheme, err := createCRDClientSet()
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

func CRDDelete(request *restful.Request) *CRDDeleteResult {
	var result CRDDeleteResult

	crdcs, scheme, err := createCRDClientSet()
	if err != nil {
		//panic(err)
		result.Status = 400
		return &result
	}

	// Create a CRD client interface
	crdclient := client.CrdClient(crdcs, scheme, "default")

	err = crdclient.Delete("fluent-bit", nil)
	if err != nil {
		//panic(err)
		result.Status = 400
		return &result
	}

	result.Status = 200
	return &result
}

func EnableQuery(request *restful.Request) *EnableResult {
	var result EnableResult

	crdcs, scheme, err := createCRDClientSet()
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
		result.Enable = "true"
		result.Status = 200
		return &result
	}

	if len(item.Spec.Settings) > 0 {
		result.Enable = getParameterValue(item.Spec.Settings[0].Parameters, "Enable")
	} else {
		result.Enable = "true"
	}

	result.Status = 200

	return &result
}

func EnableUpdate(request *restful.Request) *EnableResult {
	var result EnableResult

	parameters := new([]client.Parameter)

	err := request.ReadEntity(&parameters)
	if err != nil {
		//panic(err.Error())
		result.Status = 400
		return &result
	}

	var settings []client.Plugin
	settings = append(settings, client.Plugin{"fluentbit_settings", "fluentbit-settings", *parameters})

	crdcs, scheme, err := createCRDClientSet()
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
		spec := new(client.FluentBitOperatorSpec)
		spec.Settings = settings

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

		result.Enable = getParameterValue(itemnew.Spec.Settings[0].Parameters, "Enable")
		result.Status = 200
	} else {
		item.Spec.Settings = settings

		itemnew, err := crdclient.Update("fluent-bit", item)
		if err != nil {
			//panic(err)
			result.Status = 400
			return &result
		}

		result.Enable = getParameterValue(itemnew.Spec.Settings[0].Parameters, "Enable")
		result.Status = 200
	}

	return &result
}
