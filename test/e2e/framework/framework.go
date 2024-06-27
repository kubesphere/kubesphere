/*
Copyright 2020 KubeSphere Authors

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

package framework

import (
	"github.com/onsi/ginkgo/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/client-go/kubesphere/scheme"
	"kubesphere.io/client-go/rest"
)

type Framework struct {
	Workspace  string
	Namespaces []string
}

// KubeSphereFramework provides an interface to a test control plane so
// that the implementation can vary without affecting tests.
type KubeSphereFramework interface {
	RestClient() *rest.RESTClient
}

func NewKubeSphereFramework() KubeSphereFramework {
	f := &Framework{}
	ginkgo.AfterEach(f.AfterEach)
	ginkgo.BeforeEach(f.BeforeEach)
	return f
}

func (f *Framework) BeforeEach() {

}

func (f *Framework) AfterEach() {
}

func (f *Framework) RestClient() *rest.RESTClient {
	ctx := TestContext
	config := &rest.Config{
		Host:     ctx.Host,
		Username: ctx.Username,
		Password: ctx.Password,
		ContentConfig: rest.ContentConfig{
			NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
			ContentType:          runtime.ContentTypeJSON,
		},
	}
	c, err := rest.UnversionedRESTClientFor(config)
	if err != nil {
		panic(err)
	}
	return c
}
