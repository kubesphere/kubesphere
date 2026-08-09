/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package git

import (
	"net/http"
	"net/http/httptest"
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

// TestGitReadVerifyWithBasicAuth_BlocksSSRF reproduces
// https://github.com/kubesphere/kubesphere/issues/6653: without the dial-time
// guard, VerifyGitCredential makes ks-apiserver issue an outbound request to any
// caller-supplied remoteUrl and reflects the upstream response body in the API
// error, which lets an authenticated user reach loopback/internal services. The
// local server below stands in for such an internal service; the fix must reject
// the connection before the server ever receives the request.
func TestGitReadVerifyWithBasicAuth_BlocksSSRF(t *testing.T) {
	reached := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := runtimefakeclient.NewClientBuilder().
		WithScheme(scheme.Scheme).
		Build()
	verifier := gitVerifier{cache: client}

	err := verifier.gitReadVerifyWithBasicAuth("", "", server.URL+"/internal/repo.git")
	if err == nil {
		t.Fatalf("expected connection to loopback address %s to be rejected", server.URL)
	}
	if reached {
		t.Fatalf("internal server received a request; SSRF guard did not block the dial: %v", err)
	}
}
