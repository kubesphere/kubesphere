/*
Copyright 2021 KubeSphere Authors

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

package v1alpha1

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/apimachinery/pkg/runtime/schema"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/rest"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
)

const (
	GroupName = "package.kubesphere.io"
	Version   = "v1alpha1"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

func NewHandler(cache runtimeclient.Reader) rest.Handler {
	return &handler{cache: cache}
}

func NewFakeHandler() rest.Handler {
	return &handler{}
}

func (h *handler) AddToContainer(container *restful.Container) error {
	ws := runtime.NewWebService(GroupVersion)
	ws.Route(ws.GET("/extensionversions/{version}/files").
		To(h.ListFiles).
		Doc("List all files").
		Operation("list-extension-version-files").
		Param(ws.PathParameter("version", "The specified extension version name.")).
		Returns(http.StatusOK, api.StatusOK, []loader.BufferedFile{}))
	container.Add(ws)
	return nil
}
