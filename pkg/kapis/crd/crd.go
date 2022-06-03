/*
Copyright 2022 The KubeSphere Authors.

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

package crd

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	ksruntime "kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/models/crds"
)

const (
	ok = "OK"
)

//AddToContainer register GET and LIST API for CRD to the web service
func AddToContainer(c *restful.Container, cli client.Client, cache cache.Cache, crdList *extv1.CustomResourceDefinitionList) error {

	for _, crd := range crdList.Items {
		gvk := schema.GroupVersionKind{Group: crd.Spec.Group, Version: currentVersion(&crd), Kind: crd.Spec.Names.Kind}
		resource := crd.Spec.Names.Plural

		var h crds.Handler
		if cli.Scheme().Recognizes(gvk) {
			h = crds.NewTyped(cache, gvk, cli.Scheme())
		} else {
			h = crds.NewUnstructured(cache, gvk)
		}

		if containsRouter(c.RegisteredWebServices(), gvk.GroupVersion().String()) {
			//Skip existing Root service for now
			//TODO Register to existing WebService
			continue
		}

		//Create new WebService
		webservice := ksruntime.NewWebService(gvk.GroupVersion())

		listURL := fmt.Sprintf("/%s", resource)
		webservice.Route(webservice.GET(listURL).
			To(func(request *restful.Request, response *restful.Response) {
				h.ListResources(request, response)
			}).
			Doc(fmt.Sprintf("Cluster level resource %s query", resource)).
			Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
			Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
			Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
			Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
			Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
			Returns(http.StatusOK, ok, api.ListResult{}))

		if crd.Spec.Scope == extv1.ClusterScoped {
			getURL := fmt.Sprintf("/%s/{name}", resource)
			webservice.Route(webservice.GET(getURL).
				To(func(request *restful.Request, response *restful.Response) {
					h.GetResources(request, response)
				}).
				Doc(fmt.Sprintf("Cluster level resource %s", resource)).
				Param(webservice.PathParameter("name", "the name of the clustered resources")).
				Returns(http.StatusOK, api.StatusOK, nil))
		}

		if crd.Spec.Scope == extv1.NamespaceScoped {
			listNsURL := fmt.Sprintf("/namespaces/{namespace}/%s/", resource)
			webservice.Route(webservice.GET(listNsURL).
				To(func(request *restful.Request, response *restful.Response) {
					h.ListResources(request, response)
				}).
				Doc(fmt.Sprintf("Namespace level resource %s query", resource)).
				Param(webservice.PathParameter("namespace", "the name of the project")).
				Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
				Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
				Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
				Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
				Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
				Returns(http.StatusOK, ok, api.ListResult{}))

			getNsURL := fmt.Sprintf("/namespaces/{namespace}/%s/{name}", resource)
			webservice.Route(webservice.GET(getNsURL).
				To(func(request *restful.Request, response *restful.Response) {
					h.GetResources(request, response)
				}).
				Doc(fmt.Sprintf("Namespace level resource %s", resource)).
				Param(webservice.PathParameter("namespace", "the name of the project")).
				Param(webservice.PathParameter("name", "the name of resource")).
				Returns(http.StatusOK, ok, nil))
		}

		if crd.Spec.Scope == extv1.ClusterScoped && crd.Annotations != nil && crd.Annotations["kubesphere.io/resource-scope"] == "workspaced" {
			listWsURL := fmt.Sprintf("/workspaces/{workspace}/%s/", resource)
			webservice.Route(webservice.GET(listWsURL).
				To(func(request *restful.Request, response *restful.Response) {
					h.ListResources(request, response)
				}).
				Doc(fmt.Sprintf("Workspace level resource %s query", resource)).
				Param(webservice.PathParameter("workspace", "the name of the workspace")).
				Param(webservice.QueryParameter(query.ParameterName, "name used to do filtering").Required(false)).
				Param(webservice.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
				Param(webservice.QueryParameter(query.ParameterLimit, "limit").Required(false)).
				Param(webservice.QueryParameter(query.ParameterAscending, "sort parameters, e.g. reverse=true").Required(false).DefaultValue("ascending=false")).
				Param(webservice.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
				Returns(http.StatusOK, ok, api.ListResult{}))

			getWsURL := fmt.Sprintf("/workspaces/{workspace}/%s/{name}", resource)
			webservice.Route(webservice.GET(getWsURL).
				To(func(request *restful.Request, response *restful.Response) {
					h.GetResources(request, response)
				}).
				Doc(fmt.Sprintf("Workspace level resource %s", resource)).
				Param(webservice.PathParameter("workspace", "the name of the workspace")).
				Param(webservice.PathParameter("name", "the name of resource")).
				Returns(http.StatusOK, ok, nil))
		}
		c.Add(webservice)
	}

	return nil
}

func containsRouter(services []*restful.WebService, root string) bool {
	for _, svc := range services {
		if svc.RootPath() == "/kapis/"+root {
			return true
		}
	}
	return false
}

func currentVersion(crd *extv1.CustomResourceDefinition) string {
	for _, v := range crd.Spec.Versions {
		if v.Served && v.Storage {
			return v.Name
		}
	}
	return ""
}
