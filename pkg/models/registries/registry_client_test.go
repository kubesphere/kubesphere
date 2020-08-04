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

package registries

import (
	"testing"
)

const DockerHub = "docker.io"

func TestCreateRegistryClient(t *testing.T) {
	type imageInfo struct {
		Username string
		Password string
		Domain   string
		ExDomain string
		ExUrl    string
		UseSSL   bool
	}

	testImages := []imageInfo{
		{Domain: "kubesphere.io", ExDomain: "kubesphere.io", ExUrl: "https://kubesphere.io", UseSSL: true},
		{Domain: "127.0.0.1:5000", ExDomain: "127.0.0.1:5000", ExUrl: "http://127.0.0.1:5000", UseSSL: false},
		{Username: "Username", Password: "Password", Domain: DockerHub, ExDomain: "registry-1.docker.io", ExUrl: "https://registry-1.docker.io", UseSSL: true},
		{Domain: "harbor.devops.kubesphere.local:30280", ExDomain: "harbor.devops.kubesphere.local:30280", ExUrl: "http://harbor.devops.kubesphere.local:30280", UseSSL: false},
		{Domain: "dockerhub.qingcloud.com/zxytest/s2i-jj:jj", ExDomain: "dockerhub.qingcloud.com", ExUrl: "https://dockerhub.qingcloud.com/zxytest/s2i-jj:jj", UseSSL: true},
	}

	for _, testImage := range testImages {
		reg, err := CreateRegistryClient(testImage.Username, testImage.Password, testImage.Domain, testImage.UseSSL)
		if err != nil {
			t.Fatalf("Get err %s", err)
		}

		if reg.Domain != testImage.ExDomain {
			t.Fatalf("Doamin got %v, expected %v", reg.Domain, testImage.ExDomain)
		}

		if reg.URL != testImage.ExUrl {
			t.Fatalf("URL got %v, expected %v", reg.URL, testImage.ExUrl)
		}

	}

	testImage := Image{Domain: DockerHub, Path: "library/alpine", Tag: "latest"}
	r, err := CreateRegistryClient("", "", DockerHub, true)
	if err != nil {
		t.Fatalf("Could not get client: %s", err)
	}

	digestUrl := r.GetDigestUrl(testImage)

	// Get token.
	token, err := r.Token(digestUrl)
	if err != nil || token == "" {
		t.Fatalf("Could not get token: %s", err)
	}

	d, err := r.ImageManifest(testImage, token)
	if err != nil {
		t.Fatalf("Could not get digest: %s", err)
	}

	if d == nil {
		t.Error("Empty digest received")
	}
}
