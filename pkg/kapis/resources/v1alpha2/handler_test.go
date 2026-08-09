/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha2

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emicklei/go-restful/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiserver/pkg/authentication/user"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizerfactory"
	requestctx "kubesphere.io/kubesphere/pkg/apiserver/request"
)

// TestVerifyGitCredential_RejectsUnauthorizedSecretRef reproduces the second half of
// https://github.com/kubesphere/kubesphere/issues/6653: without an authorization check,
// any authenticated caller can pass a secretRef pointing at a Secret in a namespace they
// have no access to (e.g. kube-system) and have ks-apiserver read it and send its
// credentials to the caller-supplied remoteUrl. The handler must reject the request
// before the referenced Secret is ever read.
func TestVerifyGitCredential_RejectsUnauthorizedSecretRef(t *testing.T) {
	h := &handler{authorizer: authorizerfactory.NewAlwaysDenyAuthorizer()}

	body, err := json.Marshal(api.GitCredential{
		RemoteUrl: "https://example.invalid/repo.git",
		SecretRef: &corev1.SecretReference{Namespace: "kube-system", Name: "some-secret"},
	})
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}

	httpReq := httptest.NewRequest(http.MethodPost, "/kapis/resources.kubesphere.io/v1alpha2/git/verify", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", restful.MIME_JSON)
	httpReq = httpReq.WithContext(requestctx.WithUser(httpReq.Context(), &user.DefaultInfo{Name: "alice"}))

	req := restful.NewRequest(httpReq)
	recorder := httptest.NewRecorder()
	resp := restful.NewResponse(recorder)

	h.VerifyGitCredential(req, resp)

	if resp.StatusCode() != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, resp.StatusCode())
	}
}
