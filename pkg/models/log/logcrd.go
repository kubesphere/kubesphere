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
	"github.com/google/uuid"
	"github.com/json-iterator/go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api/logging/v1alpha2"
	"kubesphere.io/kubesphere/pkg/informers"
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

func FluentbitOutputsQuery() *FluentbitOutputsResult {
	var result FluentbitOutputsResult

	outputs, err := GetFluentbitOutputFromConfigMap()
	if err != nil {
		result.Status = http.StatusInternalServerError
		result.Error = err.Error()
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
		klog.Errorln(err)
	}

	// When adding a new output for the first time, one should always set it enabled
	output.Enable = true
	output.Id = uuid.New().String()
	output.Updatetime = time.Now()

	outputs = append(outputs, output)

	err = updateFluentbitOutputConfigMap(outputs)
	if err != nil {
		result.Status = http.StatusInternalServerError
		result.Error = err.Error()
		return &result
	}

	// 2. Keep CRD in inline with ConfigMap
	err = syncFluentbitCRDOutputWithConfigMap(outputs)
	if err != nil {
		result.Status = http.StatusInternalServerError
		result.Error = err.Error()
		return &result
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
		klog.Errorln(err)
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
		result.Error = "The output plugin to update doesn't exist. Please check the output id you provide."
		return &result
	}

	output.Updatetime = time.Now()
	outputs = append(append(outputs[:index], outputs[index+1:]...), output)

	err = updateFluentbitOutputConfigMap(outputs)
	if err != nil {
		result.Status = http.StatusInternalServerError
		result.Error = err.Error()
		return &result
	}

	// 2. Keep CRD in inline with ConfigMap
	err = syncFluentbitCRDOutputWithConfigMap(outputs)
	if err != nil {
		result.Status = http.StatusInternalServerError
		result.Error = err.Error()
		return &result
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
		result.Error = "The output plugin to delete doesn't exist. Please check the output id you provide."
		return &result
	}

	outputs = append(outputs[:index], outputs[index+1:]...)

	err := updateFluentbitOutputConfigMap(outputs)
	if err != nil {
		result.Status = http.StatusInternalServerError
		result.Error = err.Error()
		return &result
	}

	// 2. Keep CRD in inline with DB
	err = syncFluentbitCRDOutputWithConfigMap(outputs)
	if err != nil {
		result.Status = http.StatusInternalServerError
		result.Error = err.Error()
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
		klog.Errorln(err)
		return err
	}

	// Update the ConfigMap
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Errorln(err)
		return err
	}

	// Creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorln(err)
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
			klog.Errorln(err)
			return err
		}
	} else {

		// update
		configMap.Data = map[string]string{ConfigMapData: data}
		_, err = configMapClient.Update(configMap)
		if err != nil {
			klog.Errorln(err)
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
		klog.Errorln(err)
		return err
	}

	// Create a CRD client interface
	crdclient := fb.CrdClient(crdcs, scheme, LoggingNamespace)

	fluentbit, err := crdclient.Get("fluent-bit")
	if err != nil {
		klog.Errorln(err)
		return err
	}

	fluentbit.Spec.Output = enabledOutputs
	_, err = crdclient.Update("fluent-bit", fluentbit)
	if err != nil {
		klog.Errorln(err)
		return err
	}

	return nil
}

// Parse es host, port and index
func ParseEsOutputParams(params []fb.Parameter) *v1alpha2.Config {

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

	return &v1alpha2.Config{Host: host, Port: port, Index: index}
}
