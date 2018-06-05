/*
Copyright 2018 The Kubernetes Authors.

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

package apiserver

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	batchv2alpha1 "k8s.io/api/batch/v2alpha1"
	rbacv1alpha1 "k8s.io/api/rbac/v1alpha1"
	schedulingv1alpha1 "k8s.io/api/scheduling/v1alpha1"
	settingsv1alpha1 "k8s.io/api/settings/v1alpha1"
	storagev1alpha1 "k8s.io/api/storage/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/gengo/examples/set-gen/sets"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/printers"
	printersinternal "k8s.io/kubernetes/pkg/printers/internalversion"
	"k8s.io/kubernetes/test/integration/framework"
)

var kindWhiteList = sets.NewString(
	// k8s.io/api/core
	"APIGroup",
	"APIVersions",
	"Binding",
	"DeleteOptions",
	"ExportOptions",
	"GetOptions",
	"ListOptions",
	"NodeConfigSource",
	"NodeProxyOptions",
	"PodAttachOptions",
	"PodExecOptions",
	"PodPortForwardOptions",
	"PodLogOptions",
	"PodProxyOptions",
	"PodStatusResult",
	"RangeAllocation",
	"ServiceProxyOptions",
	"SerializedReference",
	// --

	// k8s.io/api/admission
	"AdmissionReview",
	// --

	// k8s.io/api/admissionregistration
	"InitializerConfiguration",
	// --

	// k8s.io/api/authentication
	"TokenRequest",
	"TokenReview",
	// --

	// k8s.io/api/authorization
	"LocalSubjectAccessReview",
	"SelfSubjectAccessReview",
	"SelfSubjectRulesReview",
	"SubjectAccessReview",
	// --

	// k8s.io/api/autoscaling
	"Scale",
	// --

	// k8s.io/api/apps
	"DeploymentRollback",
	// --

	// k8s.io/api/batch
	"JobTemplate",
	// --

	// k8s.io/api/extensions
	"ReplicationControllerDummy",
	// --

	// k8s.io/api/imagepolicy
	"ImageReview",
	// --

	// k8s.io/api/policy
	"Eviction",
	// --

	// k8s.io/kubernetes/pkg/apis/componentconfig
	"KubeSchedulerConfiguration",
	// --

	// k8s.io/apimachinery/pkg/apis/meta
	"WatchEvent",
	"Status",
	// --
)

// TODO (soltysh): this list has to go down to 0!
var missingHanlders = sets.NewString(
	"ClusterRole",
	"LimitRange",
	"MutatingWebhookConfiguration",
	"ResourceQuota",
	"Role",
	"ValidatingWebhookConfiguration",
	"VolumeAttachment",
	"PriorityClass",
	"PodPreset",
)

func TestServerSidePrint(t *testing.T) {
	s, _, closeFn := setup(t,
		// additional groupversions needed for the test to run
		batchv2alpha1.SchemeGroupVersion,
		rbacv1alpha1.SchemeGroupVersion,
		settingsv1alpha1.SchemeGroupVersion,
		schedulingv1alpha1.SchemeGroupVersion,
		storagev1alpha1.SchemeGroupVersion)
	defer closeFn()

	ns := framework.CreateTestingNamespace("server-print", s, t)
	defer framework.DeleteTestingNamespace(ns, s, t)

	tableParam := fmt.Sprintf("application/json;as=Table;g=%s;v=%s, application/json", metav1beta1.GroupName, metav1beta1.SchemeGroupVersion.Version)
	printer := newFakePrinter(printersinternal.AddHandlers)

	factory := util.NewFactory(clientcmd.NewDefaultClientConfig(*createKubeConfig(s.URL), &clientcmd.ConfigOverrides{}))
	mapper, err := factory.RESTMapper()
	if err != nil {
		t.Errorf("unexpected error getting mapper: %v", err)
		return
	}
	for gvk, apiType := range legacyscheme.Scheme.AllKnownTypes() {
		// we do not care about internal objects or lists // TODO make sure this is always true
		if gvk.Version == runtime.APIVersionInternal || strings.HasSuffix(apiType.Name(), "List") {
			continue
		}
		if kindWhiteList.Has(gvk.Kind) || missingHanlders.Has(gvk.Kind) {
			continue
		}

		t.Logf("Checking %s", gvk)
		// read table definition as returned by the server
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			t.Errorf("unexpected error getting mapping for GVK %s: %v", gvk, err)
			continue
		}
		client, err := factory.ClientForMapping(mapping)
		if err != nil {
			t.Errorf("unexpected error getting client for GVK %s: %v", gvk, err)
			continue
		}
		req := client.Get()
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			req = req.Namespace(ns.Name)
		}
		body, err := req.Resource(mapping.Resource.Resource).SetHeader("Accept", tableParam).Do().Raw()
		if err != nil {
			t.Errorf("unexpected error getting %s: %v", gvk, err)
			continue
		}
		actual, err := decodeIntoTable(body)
		if err != nil {
			t.Errorf("unexpected error decoding %s: %v", gvk, err)
			continue
		}

		// get table definition used in printers
		obj, err := legacyscheme.Scheme.New(gvk)
		if err != nil {
			t.Errorf("unexpected error creating %s: %v", gvk, err)
			continue
		}
		intGV := gvk.GroupKind().WithVersion(runtime.APIVersionInternal).GroupVersion()
		intObj, err := legacyscheme.Scheme.ConvertToVersion(obj, intGV)
		if err != nil {
			t.Errorf("unexpected error converting %s to internal: %v", gvk, err)
			continue
		}
		expectedColumnDefinitions, ok := printer.handlers[reflect.TypeOf(intObj)]
		if !ok {
			t.Errorf("missing handler for type %v", gvk)
			continue
		}

		for _, e := range expectedColumnDefinitions {
			for _, a := range actual.ColumnDefinitions {
				if a.Name == e.Name && !reflect.DeepEqual(a, e) {
					t.Errorf("unexpected difference in column definition %s for %s:\nexpected:\n%#v\nactual:\n%#v\n", e.Name, gvk, e, a)
				}
			}
		}
	}
}

type fakePrinter struct {
	handlers map[reflect.Type][]metav1beta1.TableColumnDefinition
}

var _ printers.PrintHandler = &fakePrinter{}

func (f *fakePrinter) Handler(columns, columnsWithWide []string, printFunc interface{}) error {
	return nil
}

func (f *fakePrinter) TableHandler(columns []metav1beta1.TableColumnDefinition, printFunc interface{}) error {
	printFuncValue := reflect.ValueOf(printFunc)
	objType := printFuncValue.Type().In(0)
	f.handlers[objType] = columns
	return nil
}

func (f *fakePrinter) DefaultTableHandler(columns []metav1beta1.TableColumnDefinition, printFunc interface{}) error {
	return nil
}

func newFakePrinter(fns ...func(printers.PrintHandler)) *fakePrinter {
	handlers := make(map[reflect.Type][]metav1beta1.TableColumnDefinition, len(fns))
	p := &fakePrinter{handlers: handlers}
	for _, fn := range fns {
		fn(p)
	}
	return p
}

func decodeIntoTable(body []byte) (*metav1beta1.Table, error) {
	table := &metav1beta1.Table{}
	err := json.Unmarshal(body, table)
	if err != nil {
		return nil, err
	}
	return table, nil
}

func createKubeConfig(url string) *clientcmdapi.Config {
	clusterNick := "cluster"
	userNick := "user"
	contextNick := "context"

	config := clientcmdapi.NewConfig()

	cluster := clientcmdapi.NewCluster()
	cluster.Server = url
	cluster.InsecureSkipTLSVerify = true
	config.Clusters[clusterNick] = cluster

	context := clientcmdapi.NewContext()
	context.Cluster = clusterNick
	context.AuthInfo = userNick
	config.Contexts[contextNick] = context
	config.CurrentContext = contextNick

	return config
}
