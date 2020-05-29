/*
Copyright 2020 The KubeSphere Authors.

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

package lib

import (
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apiserver/pkg/registry/rest"
	"log"
	"net"

	"github.com/go-openapi/spec"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	apiopenapi "k8s.io/apiserver/pkg/endpoints/openapi"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"

	"k8s.io/kube-openapi/pkg/builder"
	"k8s.io/kube-openapi/pkg/common"
)

type Config struct {
	Scheme *runtime.Scheme
	Codecs serializer.CodecFactory

	Info               spec.InfoProps
	OpenAPIDefinitions []common.GetOpenAPIDefinitions
	Resources          []schema.GroupVersionResource
	Mapper             *meta.DefaultRESTMapper
}

func (c *Config) GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	out := map[string]common.OpenAPIDefinition{}
	for _, def := range c.OpenAPIDefinitions {
		for k, v := range def(ref) {
			out[k] = v
		}
	}
	return out
}

func RenderOpenAPISpec(cfg Config) (string, error) {
	// we need to add the options to empty v1
	// TODO fix the server code to avoid this
	metav1.AddToGroupVersion(cfg.Scheme, schema.GroupVersion{Version: "v1"})

	// TODO: keep the generic API server from wanting this
	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	cfg.Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)

	recommendedOptions := genericoptions.NewRecommendedOptions("/registry/foo.com", cfg.Codecs.LegacyCodec(), nil)
	recommendedOptions.SecureServing.BindPort = 8443
	recommendedOptions.Etcd = nil
	recommendedOptions.Authentication = nil
	recommendedOptions.Authorization = nil
	recommendedOptions.CoreAPI = nil
	recommendedOptions.Admission = nil

	// TODO have a "real" external address
	if err := recommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		log.Fatal(fmt.Errorf("error creating self-signed certificates: %v", err))
	}

	serverConfig := genericapiserver.NewRecommendedConfig(cfg.Codecs)

	if err := recommendedOptions.ApplyTo(serverConfig); err != nil {
		log.Fatal(err)
		return "", err
	}
	serverConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(cfg.GetOpenAPIDefinitions, apiopenapi.NewDefinitionNamer(cfg.Scheme))
	serverConfig.OpenAPIConfig.Info.InfoProps = cfg.Info

	genericServer, err := serverConfig.Complete().New("stash-server", genericapiserver.NewEmptyDelegate()) // completion is done in Complete, no need for a second time
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	{
		// api router map
		table := map[string]map[string]ResourceInfo{}
		for _, gvr := range cfg.Resources {
			var resmap map[string]ResourceInfo
			// init ResourceInfo map
			if m, found := table[gvr.Group]; found {
				resmap = m
			} else {
				resmap = map[string]ResourceInfo{}
				table[gvr.Group] = resmap
			}

			gvk, err := cfg.Mapper.KindFor(gvr)
			if err != nil {
				log.Fatal(err)
				return "", err
			}
			obj, err := cfg.Scheme.New(gvk)
			if err != nil {
				return "", err
			}
			list, err := cfg.Scheme.New(gvk.GroupVersion().WithKind(gvk.Kind + "List"))
			if err != nil {
				log.Fatal(err)
				return "", err
			}

			resmap[gvr.Resource] = ResourceInfo{
				gvk:  gvk,
				obj:  obj,
				list: list,
			}
		}

		for g, resmap := range table {
			apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(g, cfg.Scheme, metav1.ParameterCodec, cfg.Codecs)
			storage := map[string]map[string]rest.Storage{}
			for r, stuff := range resmap {
				if storage[stuff.gvk.Version] == nil {
					storage[stuff.gvk.Version] = map[string]rest.Storage{}
				}
				storage[stuff.gvk.Version][r] = NewREST(stuff)
				storage[stuff.gvk.Version][r+"/status"] = NewStatusREST(
					StatusResourceInfo{
						gvk: struct {
							Group   string
							Version string
							Kind    string
						}{Group: stuff.gvk.Group, Version: stuff.gvk.Version, Kind: stuff.gvk.Kind},
						obj: stuff.obj,
					})
			}
			for version, s := range storage {
				apiGroupInfo.VersionedResourcesStorageMap[version] = s
			}

			if err := genericServer.InstallAPIGroup(&apiGroupInfo); err != nil {
				log.Fatal(err)
				return "", err
			}
		}
	}

	spec, err := builder.BuildOpenAPISpec(genericServer.Handler.GoRestfulContainer.RegisteredWebServices(), serverConfig.OpenAPIConfig)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	return string(data), nil
}
