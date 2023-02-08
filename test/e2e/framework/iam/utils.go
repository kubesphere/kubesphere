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

package iam

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"kubesphere.io/client-go/client"
	"kubesphere.io/client-go/client/generic"
	"kubesphere.io/client-go/restclient"

	"kubesphere.io/kubesphere/test/e2e/framework"
)

// NewClient creates a new client with user authencation
func NewClient(s *runtime.Scheme, user, passsword string) (client.Client, error) {

	ctx := framework.TestContext
	config := &rest.Config{
		Host:     ctx.Host,
		Username: user,
		Password: passsword,
	}

	return generic.New(config, client.Options{Scheme: s})
}

func NewRestClient(user, passsword string) (*restclient.RestClient, error) {
	ctx := framework.TestContext
	config := &rest.Config{
		Host:     ctx.Host,
		Username: user,
		Password: passsword,
	}

	return restclient.NewForConfig(config)
}
