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

package provider

import (
	"fmt"

	"github.com/projectcalico/kube-controllers/pkg/converter"
	api "github.com/projectcalico/libcalico-go/lib/apis/v3"
	constants "github.com/projectcalico/libcalico-go/lib/backend/k8s/conversion"
	v1 "k8s.io/api/networking/v1"
)

func NewFakeNetworkProvider() *FakeNetworkProvider {
	f := new(FakeNetworkProvider)
	f.NSNPData = make(map[string]*api.NetworkPolicy)
	f.policyConverter = converter.NewPolicyConverter()
	return f
}

type FakeNetworkProvider struct {
	NSNPData        map[string]*api.NetworkPolicy
	policyConverter converter.Converter
}

func (f *FakeNetworkProvider) Delete(key string) {
	delete(f.NSNPData, key)
}

func (f *FakeNetworkProvider) Start(stopCh <-chan struct{}) {

}

func (f *FakeNetworkProvider) Set(np *v1.NetworkPolicy) error {
	policy, err := f.policyConverter.Convert(np)
	if err != nil {
		return err
	}

	// Add to cache.
	k := f.policyConverter.GetKey(policy)
	tmp := policy.(api.NetworkPolicy)
	f.NSNPData[k] = &tmp

	return nil
}

func (f *FakeNetworkProvider) GetKey(name, nsname string) string {
	policyName := fmt.Sprintf(constants.K8sNetworkPolicyNamePrefix + name)
	return fmt.Sprintf("%s/%s", nsname, policyName)
}
