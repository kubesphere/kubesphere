package kapis

import (
	"github.com/emicklei/go-restful"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	devopsv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/devops/v1alpha2"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/iam/v1alpha2"
	loggingv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/logging/v1alpha2"
	monitoringv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/monitoring/v1alpha2"
	openpitrixv1 "kubesphere.io/kubesphere/pkg/kapis/openpitrix/v1"
	operationsv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/operations/v1alpha2"
	resourcesv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/resources/v1alpha2"
	servicemeshv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/servicemesh/metrics/v1alpha2"
	tenantv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/tenant/v1alpha2"
	terminalv1alpha2 "kubesphere.io/kubesphere/pkg/kapis/terminal/v1alpha2"
)

func InstallAPIs(container *restful.Container) {
	urlruntime.Must(servicemeshv1alpha2.AddToContainer(container))
	urlruntime.Must(devopsv1alpha2.AddToContainer(container))
	urlruntime.Must(loggingv1alpha2.AddToContainer(container))
	urlruntime.Must(monitoringv1alpha2.AddToContainer(container))
	urlruntime.Must(openpitrixv1.AddToContainer(container))
	urlruntime.Must(operationsv1alpha2.AddToContainer(container))
	urlruntime.Must(resourcesv1alpha2.AddToContainer(container))
	urlruntime.Must(tenantv1alpha2.AddToContainer(container))
	urlruntime.Must(terminalv1alpha2.AddToContainer(container))
}

func InstallAuthorizationAPIs(container *restful.Container) {
	urlruntime.Must(iamv1alpha2.AddToContainer(container))
}
