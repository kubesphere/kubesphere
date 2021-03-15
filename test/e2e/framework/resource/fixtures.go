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

package resource

import (
	"context"

	v1 "k8s.io/api/core/v1"

	"kubesphere.io/client-go/client"
)

// ListPods gets the Pods by namespace. If the returned error is nil, the returned PodList is valid.
func ListPods(c client.Client, namespace string) (*v1.PodList, error) {
	pods := &v1.PodList{}
	err := c.List(context.TODO(), pods, &client.ListOptions{Namespace: namespace})
	if err != nil {
		return nil, err
	}
	return pods, nil
}
