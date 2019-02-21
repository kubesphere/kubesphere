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
	"strings"

	"github.com/emicklei/go-restful"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	"kubesphere.io/kubesphere/pkg/client"
)

type FluentbitCRDResult struct {
	Status int                  `json:"status"`
	CRD    client.FluentBitSpec `json:"CRD,omitempty"`
}

type FluentbitCRDDeleteResult struct {
	Status int `json:"status"`
}

type FluentbitSettingsResult struct {
	Status int    `json:"status"`
	Enable string `json:"Enable,omitempty"`
}

type FluentbitFilter struct {
	Type       string `json:"type"`
	Field      string `json:"field"`
	Expression string `json:"expression"`
}

type FluentbitFiltersResult struct {
	Status  int               `json:"status"`
	Filters []FluentbitFilter `json:"filters,omitempty"`
}

type FluentbitOutputsResult struct {
	Status  int             `json:"status"`
	Outputs []client.Plugin `json:"outputs,omitempty"`
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

func FluentbitCRDQuery(request *restful.Request) *FluentbitCRDResult {
	var result FluentbitCRDResult

	crdcs, scheme, err := createCRDClientSet()
	if err != nil {
		//panic(err)
		result.Status = 400
		return &result
	}

	// Create a CRD client interface
	crdclient := client.CrdClient(crdcs, scheme, "kubesphere-logging-system")

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

func FluentbitCRDUpdate(request *restful.Request) *FluentbitCRDResult {
	var result FluentbitCRDResult

	spec := new(client.FluentBitSpec)

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
	crdclient := client.CrdClient(crdcs, scheme, "kubesphere-logging-system")

	var item *client.FluentBit
	var err_read error

	item, err_read = crdclient.Get("fluent-bit")
	if err_read != nil {
		//panic(err)
		fluentBitOperator := &client.FluentBit{
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

func FluentbitCRDDelete(request *restful.Request) *FluentbitCRDDeleteResult {
	var result FluentbitCRDDeleteResult

	crdcs, scheme, err := createCRDClientSet()
	if err != nil {
		//panic(err)
		result.Status = 400
		return &result
	}

	// Create a CRD client interface
	crdclient := client.CrdClient(crdcs, scheme, "kubesphere-logging-system")

	err = crdclient.Delete("fluent-bit", nil)
	if err != nil {
		//panic(err)
		result.Status = 400
		return &result
	}

	result.Status = 200
	return &result
}

func FluentbitSettingsQuery(request *restful.Request) *FluentbitSettingsResult {
	var result FluentbitSettingsResult

	crdcs, scheme, err := createCRDClientSet()
	if err != nil {
		//panic(err)
		result.Status = 400
		return &result
	}

	// Create a CRD client interface
	crdclient := client.CrdClient(crdcs, scheme, "kubesphere-logging-system")

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

func FluentbitSettingsUpdate(request *restful.Request) *FluentbitSettingsResult {
	var result FluentbitSettingsResult

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
	crdclient := client.CrdClient(crdcs, scheme, "kubesphere-logging-system")

	var item *client.FluentBit
	var err_read error

	item, err_read = crdclient.Get("fluent-bit")
	if err_read != nil {
		//panic(err)
		spec := new(client.FluentBitSpec)
		spec.Settings = settings

		fluentBitOperator := &client.FluentBit{
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

func getFilters(result *FluentbitFiltersResult, Filters []client.Plugin) {
	for _, filter := range Filters {
		if strings.Compare(filter.Name, "fluentbit-filter-input-regex") == 0 {
			parameters := strings.Split(getParameterValue(filter.Parameters, "Regex"), " ")
			field := strings.TrimSuffix(strings.TrimPrefix(parameters[0], "kubernetes_"), "_name")
			expression := parameters[1]
			result.Filters = append(result.Filters, FluentbitFilter{"Regex", field, expression})
		}
		if strings.Compare(filter.Name, "fluentbit-filter-input-exclude") == 0 {
			parameters := strings.Split(getParameterValue(filter.Parameters, "Exclude"), " ")
			field := strings.TrimSuffix(strings.TrimPrefix(parameters[0], "kubernetes_"), "_name")
			expression := parameters[1]
			result.Filters = append(result.Filters, FluentbitFilter{"Exclude", field, expression})
		}
	}
}

func FluentbitFiltersQuery(request *restful.Request) *FluentbitFiltersResult {
	var result FluentbitFiltersResult

	crdcs, scheme, err := createCRDClientSet()
	if err != nil {
		//panic(err)
		result.Status = 400
		return &result
	}

	// Create a CRD client interface
	crdclient := client.CrdClient(crdcs, scheme, "kubesphere-logging-system")

	item, err := crdclient.Get("fluent-bit")
	if err != nil {
		//panic(err)
		result.Status = 200
		return &result
	}

	getFilters(&result, item.Spec.Filter)

	result.Status = 200

	return &result
}

func FluentbitFiltersUpdate(request *restful.Request) *FluentbitFiltersResult {
	var result FluentbitFiltersResult

	filters := new([]FluentbitFilter)

	err := request.ReadEntity(&filters)
	if err != nil {
		//panic(err.Error())
		result.Status = 400
		return &result
	}

	//Generate filter plugin config
	var filter []client.Plugin

	var para_kubernetes []client.Parameter
	para_kubernetes = append(para_kubernetes, client.Parameter{"Name", nil, "kubernetes"})
	para_kubernetes = append(para_kubernetes, client.Parameter{"Match", nil, "kube.*"})
	para_kubernetes = append(para_kubernetes, client.Parameter{"Kube_URL", nil, "https://kubernetes.default.svc:443"})
	para_kubernetes = append(para_kubernetes, client.Parameter{"Kube_CA_File", nil, "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"})
	para_kubernetes = append(para_kubernetes, client.Parameter{"Kube_Token_File", nil, "/var/run/secrets/kubernetes.io/serviceaccount/token"})
	filter = append(filter, client.Plugin{"fluentbit_filter", "fluentbit-filter-kubernetes", para_kubernetes})

	var para_lift []client.Parameter
	para_lift = append(para_lift, client.Parameter{"Name", nil, "nest"})
	para_lift = append(para_lift, client.Parameter{"Match", nil, "kube.*"})
	para_lift = append(para_lift, client.Parameter{"Operation", nil, "lift"})
	para_lift = append(para_lift, client.Parameter{"Nested_under", nil, "kubernetes"})
	para_lift = append(para_lift, client.Parameter{"Prefix_with", nil, "kubernetes_"})
	filter = append(filter, client.Plugin{"fluentbit_filter", "fluentbit-filter-input-lift", para_lift})

	var para_remove_stream []client.Parameter
	para_remove_stream = append(para_remove_stream, client.Parameter{"Name", nil, "modify"})
	para_remove_stream = append(para_remove_stream, client.Parameter{"Match", nil, "kube.*"})
	para_remove_stream = append(para_remove_stream, client.Parameter{"Remove", nil, "stream"})
	filter = append(filter, client.Plugin{"fluentbit_filter", "fluentbit-filter-input-remove-stream", para_remove_stream})

	var para_remove_labels []client.Parameter
	para_remove_labels = append(para_remove_labels, client.Parameter{"Name", nil, "modify"})
	para_remove_labels = append(para_remove_labels, client.Parameter{"Match", nil, "kube.*"})
	para_remove_labels = append(para_remove_labels, client.Parameter{"Remove", nil, "kubernetes_labels"})
	filter = append(filter, client.Plugin{"fluentbit_filter", "fluentbit-filter-input-remove-labels", para_remove_labels})

	var para_remove_annotations []client.Parameter
	para_remove_annotations = append(para_remove_annotations, client.Parameter{"Name", nil, "modify"})
	para_remove_annotations = append(para_remove_annotations, client.Parameter{"Match", nil, "kube.*"})
	para_remove_annotations = append(para_remove_annotations, client.Parameter{"Remove", nil, "kubernetes_annotations"})
	filter = append(filter, client.Plugin{"fluentbit_filter", "fluentbit-filter-input-remove-annotations", para_remove_annotations})

	var para_remove_pod_id []client.Parameter
	para_remove_pod_id = append(para_remove_pod_id, client.Parameter{"Name", nil, "modify"})
	para_remove_pod_id = append(para_remove_pod_id, client.Parameter{"Match", nil, "kube.*"})
	para_remove_pod_id = append(para_remove_pod_id, client.Parameter{"Remove", nil, "kubernetes_pod_id"})
	filter = append(filter, client.Plugin{"fluentbit_filter", "fluentbit-filter-input-remove-podid", para_remove_pod_id})

	var para_remove_docker_id []client.Parameter
	para_remove_docker_id = append(para_remove_docker_id, client.Parameter{"Name", nil, "modify"})
	para_remove_docker_id = append(para_remove_docker_id, client.Parameter{"Match", nil, "kube.*"})
	para_remove_docker_id = append(para_remove_docker_id, client.Parameter{"Remove", nil, "kubernetes_docker_id"})
	filter = append(filter, client.Plugin{"fluentbit_filter", "fluentbit-filter-input-remove-dockerid", para_remove_docker_id})

	if len(*filters) > 0 {
		for _, item := range *filters {
			if strings.Compare(item.Type, "Regex") == 0 {
				field := "kubernetes_" + strings.TrimSpace(item.Field) + "_name"
				expression := strings.TrimSpace(item.Expression)

				var para_regex []client.Parameter
				para_regex = append(para_regex, client.Parameter{"Name", nil, "grep"})
				para_regex = append(para_regex, client.Parameter{"Match", nil, "kube.*"})
				para_regex = append(para_regex, client.Parameter{"Regex", nil, field + " " + expression})
				filter = append(filter, client.Plugin{"fluentbit_filter", "fluentbit-filter-input-regex", para_regex})
			}

			if strings.Compare(item.Type, "Exclude") == 0 {
				field := "kubernetes_" + strings.TrimSpace(item.Field) + "_name"
				expression := strings.TrimSpace(item.Expression)

				var para_exclude []client.Parameter
				para_exclude = append(para_exclude, client.Parameter{"Name", nil, "grep"})
				para_exclude = append(para_exclude, client.Parameter{"Match", nil, "kube.*"})
				para_exclude = append(para_exclude, client.Parameter{"Exclude", nil, field + " " + expression})
				filter = append(filter, client.Plugin{"fluentbit_filter", "fluentbit-filter-input-exclude", para_exclude})
			}
		}
	}

	var para_nest []client.Parameter
	para_nest = append(para_nest, client.Parameter{"Name", nil, "nest"})
	para_nest = append(para_nest, client.Parameter{"Match", nil, "kube.*"})
	para_nest = append(para_nest, client.Parameter{"Operation", nil, "nest"})
	para_nest = append(para_nest, client.Parameter{"Wildcard", nil, "kubernetes_*"})
	para_nest = append(para_nest, client.Parameter{"Nested_under", nil, "kubernetes"})
	para_nest = append(para_nest, client.Parameter{"Remove_prefix", nil, "kubernetes_"})
	filter = append(filter, client.Plugin{"fluentbit_filter", "fluentbit-filter-input-nest", para_nest})

	crdcs, scheme, err := createCRDClientSet()
	if err != nil {
		//panic(err)
		result.Status = 400
		return &result
	}

	// Create a CRD client interface
	crdclient := client.CrdClient(crdcs, scheme, "kubesphere-logging-system")

	var item *client.FluentBit
	var err_read error

	item, err_read = crdclient.Get("fluent-bit")
	if err_read != nil {
		//panic(err)
		spec := new(client.FluentBitSpec)
		spec.Filter = filter

		fluentBitOperator := &client.FluentBit{
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

		getFilters(&result, itemnew.Spec.Filter)
		result.Status = 200
	} else {
		item.Spec.Filter = filter

		itemnew, err := crdclient.Update("fluent-bit", item)
		if err != nil {
			//panic(err)
			result.Status = 400
			return &result
		}

		getFilters(&result, itemnew.Spec.Filter)
		result.Status = 200
	}

	return &result
}

func FluentbitOutputsQuery(request *restful.Request) *FluentbitOutputsResult {
	var result FluentbitOutputsResult

	crdcs, scheme, err := createCRDClientSet()
	if err != nil {
		//panic(err)
		result.Status = 400
		return &result
	}

	// Create a CRD client interface
	crdclient := client.CrdClient(crdcs, scheme, "kubesphere-logging-system")

	item, err := crdclient.Get("fluent-bit")
	if err != nil {
		//panic(err)
		result.Status = 200
		return &result
	}

	result.Outputs = item.Spec.Output
	result.Status = 200

	return &result
}

func FluentbitOutputsUpdate(request *restful.Request) *FluentbitOutputsResult {
	var result FluentbitOutputsResult

	outputs := new([]client.Plugin)

	err := request.ReadEntity(&outputs)
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
	crdclient := client.CrdClient(crdcs, scheme, "kubesphere-logging-system")

	var item *client.FluentBit
	var err_read error

	item, err_read = crdclient.Get("fluent-bit")
	if err_read != nil {
		//panic(err)
		spec := new(client.FluentBitSpec)
		spec.Output = *outputs

		fluentBitOperator := &client.FluentBit{
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

		result.Outputs = itemnew.Spec.Output
		result.Status = 200
	} else {
		item.Spec.Output = *outputs

		itemnew, err := crdclient.Update("fluent-bit", item)
		if err != nil {
			//panic(err)
			result.Status = 400
			return &result
		}

		result.Outputs = itemnew.Spec.Output
		result.Status = 200
	}

	return &result
}
