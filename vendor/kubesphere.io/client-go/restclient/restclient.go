/*
Copyright 2021 The KubeSphere Authors.

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

package restclient

import (
	rest "k8s.io/client-go/rest"
	iamv1alpha2 "kubesphere.io/client-go/restclient/versioned/iam/v1alpha2"
)

// NewForConfig returns a new Client using the provided config and Options.
func NewForConfig(c *rest.Config) (*RestClient, error) {
	var rc RestClient
	var err error
	rc.iamV1alpha2, err = iamv1alpha2.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	return &rc, nil
}

// RestClient is a set of restful API clients that doesn't compatible with
// Kube API machinery.
type RestClient struct {
	iamV1alpha2 *iamv1alpha2.IamV1alpha2Client
}

// IamV1alpha2 retrieves the IamV1alpha2Client
func (c *RestClient) IamV1alpha2() iamv1alpha2.IamV1alpha2Interface {
	return c.iamV1alpha2
}
