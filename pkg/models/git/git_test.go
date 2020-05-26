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

package git

import (
	"testing"
)

func TestGitReadVerifyWithBasicAuth(t *testing.T) {
	shouldSuccess := []map[string]string{
		{
			"username": "",
			"password": "",
			"remote":   "https://github.com/kubesphere/kubesphere",
		},
	}
	shouldFailed := []map[string]string{
		{
			"username": "",
			"password": "",
			"remote":   "https://github.com/kubesphere/kubesphere12222",
		},
		{
			"username": "",
			"password": "",
			"remote":   "git@github.com:kubesphere/kubesphere.git",
		},
		{
			"username": "runzexia",
			"password": "",
			"remote":   "git@github.com:kubesphere/kubesphere.git",
		},
		{
			"username": "",
			"password": "",
			"remote":   "git@fdsfs41342`@@@2414!!!!github.com:kubesphere/kubesphere.git",
		},
	}
	verifier := gitVerifier{informers: nil}

	for _, item := range shouldSuccess {
		err := verifier.gitReadVerifyWithBasicAuth(item["username"], item["password"], item["remote"])
		if err != nil {

			t.Errorf("should could access repo [%s] with %s:%s, %v", item["username"], item["password"], item["remote"], err)
		}
	}

	for _, item := range shouldFailed {
		err := verifier.gitReadVerifyWithBasicAuth(item["username"], item["password"], item["remote"])
		if err == nil {
			t.Errorf("should could access repo [%s] with %s:%s ", item["username"], item["password"], item["remote"])
		}
	}
}
