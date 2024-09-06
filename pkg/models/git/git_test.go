/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package git

import (
	"testing"

	runtimefakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"kubesphere.io/kubesphere/pkg/scheme"
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
	client := runtimefakeclient.NewClientBuilder().
		WithScheme(scheme.Scheme).
		Build()

	verifier := gitVerifier{cache: client}

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
