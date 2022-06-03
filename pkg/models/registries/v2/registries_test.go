// Copyright 2022 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package v2

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/google/go-cmp/cmp"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

func TestRegistryerConfig(t *testing.T) {
	testCases := []struct {
		name       string
		secret     *corev1.Secret
		image      string
		configFile *v1.ConfigFile
		expectErr  bool
	}{
		{
			name:      "Should fetch image config with public registry",
			secret:    nil,
			image:     "kubesphere/ks-apiserver:v3.1.0",
			expectErr: false,
			configFile: &v1.ConfigFile{
				Config: v1.Config{
					Image: "sha256:a3006bafb7702494227f7fa69720e65e1324b7bf8ece8ac03d9fe1d0134e7341",
				},
			},
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {
			secretAuthenticator, err := NewSecretAuthenticator(testCase.secret)
			if err != nil {
				t.Error(err)
			}

			// fix platform to linux/amd64 so we could compare config
			platform := &v1.Platform{
				OS:           "linux",
				Architecture: "amd64",
			}

			options := secretAuthenticator.Options()
			options = append(options, WithPlatform(platform))

			registryer := NewRegistryer(options...)

			config, err := registryer.Config(testCase.image)
			if testCase.expectErr && err == nil {
				t.Errorf("expected error, but got nil")
			}

			if !testCase.expectErr && err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(testCase.configFile.Config.Image, config.Config.Image); len(diff) != 0 {
				t.Errorf("expected %v, but got %v", testCase.configFile.Config.Image, config.Config.Image)
			}
		})
	}

}

func TestRegistryerListRepoTags(t *testing.T) {
	testCases := []struct {
		name           string
		secret         *corev1.Secret
		image          string
		repositoryTags RepositoryTags
		expectErr      bool
	}{
		{
			name:      "Should fetch config with public registry",
			secret:    nil,
			image:     "kubesphere/ks-apiserver",
			expectErr: false,
			repositoryTags: RepositoryTags{
				Registry: "index.docker.io",
				Tags: []string{
					"v3.1.1",
					"v3.1.0",
					"latest",
				},
			},
		},
		{
			name:      "Should fetch config from public registry with credential",
			secret:    buildSecret("dockerhub.qingcloud.com", "guest", "guest", false),
			image:     "dockerhub.qingcloud.com/calico/cni",
			expectErr: false,
			repositoryTags: RepositoryTags{
				Registry: "dockerhub.qingcloud.com",
				Tags: []string{
					"v1.11.4",
					"v3.1.3",
					"v3.3.2",
					"v3.3.3",
					"v3.3.6",
					"v3.7.3",
					"v3.8.4",
				},
			},
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {
			secretAuthenticator, err := NewSecretAuthenticator(testCase.secret)
			if err != nil {
				t.Error(err)
			}

			// fix platform to linux/amd64 so we could compare config
			platform := &v1.Platform{
				OS:           "linux",
				Architecture: "amd64",
			}

			options := secretAuthenticator.Options()
			options = append(options, WithPlatform(platform))

			registryer := NewRegistryer(options...)

			tags, err := registryer.ListRepositoryTags(testCase.image)
			if testCase.expectErr && err == nil {
				t.Errorf("expected error, but got nil")
			}

			if !testCase.expectErr && err != nil {
				t.Error(err)
			}

			cotains := func(s []string, e string) bool {
				for _, a := range s {
					if a == e {
						return true
					}
				}
				return false
			}

			for _, tag := range testCase.repositoryTags.Tags {
				if !cotains(tags.Tags, tag) {
					t.Errorf("no expected tag %s in result %v", tag, tags.Tags)
				}
			}
		})
	}
}
