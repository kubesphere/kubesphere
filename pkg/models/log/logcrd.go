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
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/json-iterator/go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"kubesphere.io/kubesphere/pkg/informers"
	es "kubesphere.io/kubesphere/pkg/simple/client/elasticsearch"
	fb "kubesphere.io/kubesphere/pkg/simple/client/fluentbit"
	"net/http"
	"strings"
	"time"
)

var jsonIter = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	ConfigMapName    = "fluent-bit-output-config"
	ConfigMapData    = "outputs"
	LoggingNamespace = "kubesphere-logging-system"
)

func createCRDClientSet() (*rest.RESTClient, *runtime.Scheme, error) {
	config, err := fb.GetClientConfig("")
	if err != nil {
		//panic(err.Error())
		return nil, nil, err
	}

	// Create a new clientset which include our CRD schema
	return fb.NewFluentbitCRDClient(config)
}

func getParameterValue(parameters []fb.Parameter, name string) string {
	var value string

	value = ""
	for _, parameter := range parameters {
		if parameter.Name == name {
			value = parameter.Value
		}
	}

	return value
}

func getFilters(result *FluentbitFiltersResult, Filters []fb.Plugin) {
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

func FluentbitFiltersQuery() *FluentbitFiltersResult {
	var result FluentbitFiltersResult

	crdcs, scheme, err := createCRDClientSet()
	if err != nil {
		result.Status = http.StatusInternalServerError
		return &result
	}

	// Create a CRD client interface
	crdclient := fb.CrdClient(crdcs, scheme, LoggingNamespace)

	item, err := crdclient.Get("fluent-bit")
	if err != nil {
		result.Status = http.StatusInternalServerError
		return &result
	}

	getFilters(&result, item.Spec.Filter)

	result.Status = http.StatusOK

	return &result
}

func FluentbitFiltersUpdate(filters *[]FluentbitFilter) *FluentbitFiltersResult {
	var result FluentbitFiltersResult

	//Generate filter plugin config
	var filter []fb.Plugin

	var para_kubernetes []fb.Parameter
	para_kubernetes = append(para_kubernetes, fb.Parameter{Name: "Name", Value: "kubernetes"})
	para_kubernetes = append(para_kubernetes, fb.Parameter{Name: "Match", Value: "kube.*"})
	para_kubernetes = append(para_kubernetes, fb.Parameter{Name: "Kube_URL", Value: "https://kubernetes.default.svc:443"})
	para_kubernetes = append(para_kubernetes, fb.Parameter{Name: "Kube_CA_File", Value: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"})
	para_kubernetes = append(para_kubernetes, fb.Parameter{Name: "Kube_Token_File", Value: "/var/run/secrets/kubernetes.io/serviceaccount/token"})
	filter = append(filter, fb.Plugin{Type: "fluentbit_filter", Name: "fluentbit-filter-kubernetes", Parameters: para_kubernetes})

	var para_lift []fb.Parameter
	para_lift = append(para_lift, fb.Parameter{Name: "Name", Value: "nest"})
	para_lift = append(para_lift, fb.Parameter{Name: "Match", Value: "kube.*"})
	para_lift = append(para_lift, fb.Parameter{Name: "Operation", Value: "lift"})
	para_lift = append(para_lift, fb.Parameter{Name: "Nested_under", Value: "kubernetes"})
	para_lift = append(para_lift, fb.Parameter{Name: "Prefix_with", Value: "kubernetes_"})
	filter = append(filter, fb.Plugin{Type: "fluentbit_filter", Name: "fluentbit-filter-input-lift", Parameters: para_lift})

	var para_remove_stream []fb.Parameter
	para_remove_stream = append(para_remove_stream, fb.Parameter{Name: "Name", Value: "modify"})
	para_remove_stream = append(para_remove_stream, fb.Parameter{Name: "Match", Value: "kube.*"})
	para_remove_stream = append(para_remove_stream, fb.Parameter{Name: "Remove", Value: "stream"})
	filter = append(filter, fb.Plugin{Type: "fluentbit_filter", Name: "fluentbit-filter-input-remove-stream", Parameters: para_remove_stream})

	var para_remove_labels []fb.Parameter
	para_remove_labels = append(para_remove_labels, fb.Parameter{Name: "Name", Value: "modify"})
	para_remove_labels = append(para_remove_labels, fb.Parameter{Name: "Match", Value: "kube.*"})
	para_remove_labels = append(para_remove_labels, fb.Parameter{Name: "Remove", Value: "kubernetes_labels"})
	filter = append(filter, fb.Plugin{Type: "fluentbit_filter", Name: "fluentbit-filter-input-remove-labels", Parameters: para_remove_labels})

	var para_remove_annotations []fb.Parameter
	para_remove_annotations = append(para_remove_annotations, fb.Parameter{Name: "Name", Value: "modify"})
	para_remove_annotations = append(para_remove_annotations, fb.Parameter{Name: "Match", Value: "kube.*"})
	para_remove_annotations = append(para_remove_annotations, fb.Parameter{Name: "Remove", Value: "kubernetes_annotations"})
	filter = append(filter, fb.Plugin{Type: "fluentbit_filter", Name: "fluentbit-filter-input-remove-annotations", Parameters: para_remove_annotations})

	var para_remove_pod_id []fb.Parameter
	para_remove_pod_id = append(para_remove_pod_id, fb.Parameter{Name: "Name", Value: "modify"})
	para_remove_pod_id = append(para_remove_pod_id, fb.Parameter{Name: "Match", Value: "kube.*"})
	para_remove_pod_id = append(para_remove_pod_id, fb.Parameter{Name: "Remove", Value: "kubernetes_pod_id"})
	filter = append(filter, fb.Plugin{Type: "fluentbit_filter", Name: "fluentbit-filter-input-remove-podid", Parameters: para_remove_pod_id})

	var para_remove_docker_id []fb.Parameter
	para_remove_docker_id = append(para_remove_docker_id, fb.Parameter{Name: "Name", Value: "modify"})
	para_remove_docker_id = append(para_remove_docker_id, fb.Parameter{Name: "Match", Value: "kube.*"})
	para_remove_docker_id = append(para_remove_docker_id, fb.Parameter{Name: "Remove", Value: "kubernetes_docker_id"})
	filter = append(filter, fb.Plugin{Type: "fluentbit_filter", Name: "fluentbit-filter-input-remove-dockerid", Parameters: para_remove_docker_id})

	if len(*filters) > 0 {
		for _, item := range *filters {
			if strings.Compare(item.Type, "Regex") == 0 {
				field := "kubernetes_" + strings.TrimSpace(item.Field) + "_name"
				expression := strings.TrimSpace(item.Expression)

				var para_regex []fb.Parameter
				para_regex = append(para_regex, fb.Parameter{Name: "Name", Value: "grep"})
				para_regex = append(para_regex, fb.Parameter{Name: "Match", Value: "kube.*"})
				para_regex = append(para_regex, fb.Parameter{Name: "Regex", Value: field + " " + expression})
				filter = append(filter, fb.Plugin{Type: "fluentbit_filter", Name: "fluentbit-filter-input-regex", Parameters: para_regex})
			}

			if strings.Compare(item.Type, "Exclude") == 0 {
				field := "kubernetes_" + strings.TrimSpace(item.Field) + "_name"
				expression := strings.TrimSpace(item.Expression)

				var para_exclude []fb.Parameter
				para_exclude = append(para_exclude, fb.Parameter{Name: "Name", Value: "grep"})
				para_exclude = append(para_exclude, fb.Parameter{Name: "Match", Value: "kube.*"})
				para_exclude = append(para_exclude, fb.Parameter{Name: "Exclude", Value: field + " " + expression})
				filter = append(filter, fb.Plugin{Type: "fluentbit_filter", Name: "fluentbit-filter-input-exclude", Parameters: para_exclude})
			}
		}
	}

	var para_nest []fb.Parameter
	para_nest = append(para_nest, fb.Parameter{Name: "Name", Value: "nest"})
	para_nest = append(para_nest, fb.Parameter{Name: "Match", Value: "kube.*"})
	para_nest = append(para_nest, fb.Parameter{Name: "Operation", Value: "nest"})
	para_nest = append(para_nest, fb.Parameter{Name: "Wildcard", Value: "kubernetes_*"})
	para_nest = append(para_nest, fb.Parameter{Name: "Nested_under", Value: "kubernetes"})
	para_nest = append(para_nest, fb.Parameter{Name: "Remove_prefix", Value: "kubernetes_"})
	filter = append(filter, fb.Plugin{Type: "fluentbit_filter", Name: "fluentbit-filter-input-nest", Parameters: para_nest})

	crdcs, scheme, err := createCRDClientSet()
	if err != nil {
		result.Status = http.StatusInternalServerError
		return &result
	}

	// Create a CRD client interface
	crdclient := fb.CrdClient(crdcs, scheme, LoggingNamespace)

	var item *fb.FluentBit
	var err_read error

	item, err_read = crdclient.Get("fluent-bit")
	if err_read != nil {
		result.Status = http.StatusInternalServerError
		return &result
	}

	item.Spec.Filter = filter

	itemnew, err := crdclient.Update("fluent-bit", item)
	if err != nil {
		result.Status = http.StatusInternalServerError
		return &result
	}

	getFilters(&result, itemnew.Spec.Filter)
	result.Status = http.StatusOK

	return &result
}

func FluentbitOutputsQuery() *FluentbitOutputsResult {
	var result FluentbitOutputsResult

	outputs, err := GetFluentbitOutputFromConfigMap()
	if err != nil {
		result.Status = http.StatusNotFound
		return &result
	}

	result.Outputs = outputs
	result.Status = http.StatusOK

	return &result
}

func FluentbitOutputInsert(output fb.OutputPlugin) *FluentbitOutputsResult {
	var result FluentbitOutputsResult

	// 1. Update ConfigMap
	var outputs []fb.OutputPlugin
	outputs, err := GetFluentbitOutputFromConfigMap()
    if err != nil {
		// If the ConfigMap doesn't exist, a new one will be created later
		glog.Errorln(err)
	}

	// When adding a new output for the first time, one should always set it disabled
	output.Enable = true
	output.Id = uuid.New().String()
	output.Updatetime = time.Now()

	outputs = append(outputs, output)

	err = updateFluentbitOutputConfigMap(outputs)
	if err != nil {
		result.Status = http.StatusInternalServerError
		return &result
	}

	// 2. Keep CRD in inline with ConfigMap
	err = syncFluentbitCRDOutputWithConfigMap(outputs)
	if err != nil {
		result.Status = http.StatusInternalServerError
		return &result
	}

	// 3. If it's an configs output added, reset configs client configs
	configs := ParseEsOutputParams(output.Parameters)
	if configs != nil {
		configs.WriteESConfigs()
	}

	result.Status = http.StatusOK
	return &result
}

func FluentbitOutputUpdate(output fb.OutputPlugin, id string) *FluentbitOutputsResult {
	var result FluentbitOutputsResult

	// 1. Update ConfigMap
	var outputs []fb.OutputPlugin
	outputs, err := GetFluentbitOutputFromConfigMap()
	if err != nil {
		// If the ConfigMap doesn't exist, a new one will be created later
		glog.Errorln(err)
	}

	index := 0
	for _, output := range outputs {
		if output.Id == id {
			break
		}
		index++
	}

	if index >= len(outputs) {
		result.Status = http.StatusNotFound
		return &result
	}

	output.Updatetime = time.Now()
	outputs = append(append(outputs[:index], outputs[index+1:]...), output)

	err = updateFluentbitOutputConfigMap(outputs)
	if err != nil {
		result.Status = http.StatusInternalServerError
		return &result
	}

	// 2. Keep CRD in inline with ConfigMap
	err = syncFluentbitCRDOutputWithConfigMap(outputs)
	if err != nil {
		result.Status = http.StatusInternalServerError
		return &result
	}

	// 3. If it's an configs output updated, reset configs client configs
	configs := ParseEsOutputParams(output.Parameters)
	if configs != nil {
		configs.WriteESConfigs()
	}

	result.Status = http.StatusOK
	return &result
}

func FluentbitOutputDelete(id string) *FluentbitOutputsResult {
	var result FluentbitOutputsResult

	// 1. Update ConfigMap
	// If the ConfigMap doesn't exist, a new one will be created
	outputs, _ := GetFluentbitOutputFromConfigMap()

	index := 0
	for _, output := range outputs {
		if output.Id == id {
			break
		}
		index++
	}

	if index >= len(outputs) {
		result.Status = http.StatusNotFound
		return &result
	}

	outputs = append(outputs[:index], outputs[index+1:]...)

	err := updateFluentbitOutputConfigMap(outputs)
	if err != nil {
		result.Status = http.StatusInternalServerError
		return &result
	}

	// 2. Keep CRD in inline with DB
	err = syncFluentbitCRDOutputWithConfigMap(outputs)
	if err != nil {
		result.Status = http.StatusInternalServerError
		return &result
	}

	result.Status = http.StatusOK
	return &result
}

func GetFluentbitOutputFromConfigMap() ([]fb.OutputPlugin, error) {
	configMap, err := informers.SharedInformerFactory().Core().V1().ConfigMaps().Lister().ConfigMaps(LoggingNamespace).Get(ConfigMapName)
	if err != nil {
		return nil, err
	}

	data := configMap.Data[ConfigMapData]

	var outputs []fb.OutputPlugin
	if err = jsonIter.UnmarshalFromString(data, &outputs); err != nil {
		return nil, err
	}

	return outputs, nil
}

func updateFluentbitOutputConfigMap(outputs []fb.OutputPlugin) error {

	var data string
	data, err := jsonIter.MarshalToString(outputs)
	if err != nil {
		glog.Errorln(err)
		return err
	}

	// Update the ConfigMap
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Errorln(err)
		return err
	}

	// Creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Errorln(err)
		return err
	}

	configMapClient := clientset.CoreV1().ConfigMaps(LoggingNamespace)

	configMap, err := configMapClient.Get(ConfigMapName, metav1.GetOptions{})
	if err != nil {

		// If the ConfigMap doesn't exist, create a new one
		newConfigMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: ConfigMapName,
			},
			Data: map[string]string{ConfigMapData: data},
		}

		_, err = configMapClient.Create(newConfigMap)
		if err != nil {
			glog.Errorln(err)
			return err
		}
	} else {

		// update
		configMap.Data = map[string]string{ConfigMapData: data}
		_, err = configMapClient.Update(configMap)
		if err != nil {
			glog.Errorln(err)
			return err
		}
	}

	return nil
}

func syncFluentbitCRDOutputWithConfigMap(outputs []fb.OutputPlugin) error {

	var enabledOutputs []fb.Plugin
	for _, output := range outputs {
		if output.Enable {
			enabledOutputs = append(enabledOutputs, fb.Plugin{Type: output.Type, Name: output.Name, Parameters: output.Parameters})
		}
	}

	// Empty output is not allowed, must specify a null-type output
	if len(enabledOutputs) == 0 {
		enabledOutputs = []fb.Plugin{
			{
				Type: "fluentbit_output",
				Name: "fluentbit-output-null",
				Parameters: []fb.Parameter{
					{
						Name:  "Name",
						Value: "null",
					},
					{
						Name:  "Match",
						Value: "*",
					},
				},
			},
		}
	}

	crdcs, scheme, err := createCRDClientSet()
	if err != nil {
		return err
	}

	// Create a CRD client interface
	crdclient := fb.CrdClient(crdcs, scheme, LoggingNamespace)

	fluentbit, err := crdclient.Get("fluent-bit")
	if err != nil {
		return err
	}

	fluentbit.Spec.Output = enabledOutputs
	_, err = crdclient.Update("fluent-bit", fluentbit)
	if err != nil {
		return err
	}

	return nil
}

// Parse es host, port and index
func ParseEsOutputParams(params []fb.Parameter) *es.ESConfigs {

	var (
		isEsFound bool

		host           = "127.0.0.1"
		port           = "9200"
		index          = "logstash"
		logstashFormat string
		logstashPrefix string
	)

	for _, param := range params {
		switch param.Name {
		case "Name":
			if param.Value == "es" {
				isEsFound = true
			}
		case "Host":
			host = param.Value
		case "Port":
			port = param.Value
		case "Index":
			index = param.Value
		case "Logstash_Format":
			logstashFormat = strings.ToLower(param.Value)
		case "Logstash_Prefix":
			logstashPrefix = param.Value
		}
	}

	if !isEsFound {
		return nil
	}

	// If Logstash_Format is On/True, ignore Index
	if logstashFormat == "on" || logstashFormat == "true" {
		if logstashPrefix != "" {
			index = logstashPrefix
		} else {
			index = "logstash"
		}
	}

	return &es.ESConfigs{Host: host, Port: port, Index: index}
}
