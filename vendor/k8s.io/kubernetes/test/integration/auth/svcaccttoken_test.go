/*
Copyright 2017 The Kubernetes Authors.

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

package auth

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	authenticationv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/authentication/request/bearertoken"
	"k8s.io/apiserver/pkg/authorization/authorizerfactory"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	utilfeaturetesting "k8s.io/apiserver/pkg/util/feature/testing"
	clientset "k8s.io/client-go/kubernetes"
	externalclientset "k8s.io/client-go/kubernetes"
	certutil "k8s.io/client-go/util/cert"
	serviceaccountgetter "k8s.io/kubernetes/pkg/controller/serviceaccount"
	"k8s.io/kubernetes/pkg/features"
	"k8s.io/kubernetes/pkg/serviceaccount"
	"k8s.io/kubernetes/test/integration/framework"
)

const ecdsaPrivateKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIEZmTmUhuanLjPA2CLquXivuwBDHTt5XYwgIr/kA1LtRoAoGCCqGSM49
AwEHoUQDQgAEH6cuzP8XuD5wal6wf9M6xDljTOPLX2i8uIp/C/ASqiIGUeeKQtX0
/IR3qCXyThP/dbCiHrF3v1cuhBOHY8CLVg==
-----END EC PRIVATE KEY-----`

func TestServiceAccountTokenCreate(t *testing.T) {
	defer utilfeaturetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.TokenRequest, true)()

	// Build client config, clientset, and informers
	sk, err := certutil.ParsePrivateKeyPEM([]byte(ecdsaPrivateKey))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	pk := sk.(*ecdsa.PrivateKey).PublicKey

	const iss = "https://foo.bar.example.com"
	aud := []string{"api"}

	gcs := &clientset.Clientset{}

	// Start the server
	masterConfig := framework.NewIntegrationTestMasterConfig()
	masterConfig.GenericConfig.Authorization.Authorizer = authorizerfactory.NewAlwaysAllowAuthorizer()
	masterConfig.GenericConfig.Authentication.Authenticator = bearertoken.New(
		serviceaccount.JWTTokenAuthenticator(
			iss,
			[]interface{}{&pk},
			serviceaccount.NewValidator(aud, serviceaccountgetter.NewGetterFromClient(gcs)),
		),
	)
	masterConfig.ExtraConfig.ServiceAccountIssuer = serviceaccount.JWTTokenGenerator(iss, sk)
	masterConfig.ExtraConfig.ServiceAccountAPIAudiences = aud

	master, _, closeFn := framework.RunAMaster(masterConfig)
	defer closeFn()

	cs, err := clientset.NewForConfig(master.GenericAPIServer.LoopbackClientConfig)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	*gcs = *cs

	var (
		sa = &v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-svcacct",
				Namespace: "myns",
			},
		}
		pod = &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: sa.Namespace,
			},
			Spec: v1.PodSpec{
				ServiceAccountName: sa.Name,
				Containers:         []v1.Container{{Name: "test-container", Image: "nginx"}},
			},
		}
		otherpod = &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other-test-pod",
				Namespace: sa.Namespace,
			},
			Spec: v1.PodSpec{
				ServiceAccountName: "other-" + sa.Name,
				Containers:         []v1.Container{{Name: "test-container", Image: "nginx"}},
			},
		}
		secret = &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-secret",
				Namespace: sa.Namespace,
			},
		}

		one      = int64(1)
		wrongUID = types.UID("wrong")
		noUID    = types.UID("")
	)

	t.Run("bound to service account", func(t *testing.T) {
		treq := &authenticationv1.TokenRequest{
			Spec: authenticationv1.TokenRequestSpec{
				Audiences:         []string{"api"},
				ExpirationSeconds: &one,
			},
		}

		if resp, err := cs.CoreV1().ServiceAccounts(sa.Namespace).CreateToken(sa.Name, treq); err == nil {
			t.Fatalf("expected err creating token for nonexistant svcacct but got: %#v", resp)
		}
		sa, delSvcAcct := createDeleteSvcAcct(t, cs, sa)
		defer delSvcAcct()

		treq, err = cs.CoreV1().ServiceAccounts(sa.Namespace).CreateToken(sa.Name, treq)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		checkPayload(t, treq.Status.Token, `"system:serviceaccount:myns:test-svcacct"`, "sub")
		checkPayload(t, treq.Status.Token, `["api"]`, "aud")
		checkPayload(t, treq.Status.Token, "null", "kubernetes.io", "pod")
		checkPayload(t, treq.Status.Token, "null", "kubernetes.io", "secret")
		checkPayload(t, treq.Status.Token, `"myns"`, "kubernetes.io", "namespace")
		checkPayload(t, treq.Status.Token, `"test-svcacct"`, "kubernetes.io", "serviceaccount", "name")

		doTokenReview(t, cs, treq, false)
		delSvcAcct()
		doTokenReview(t, cs, treq, true)
	})

	t.Run("bound to service account and pod", func(t *testing.T) {
		treq := &authenticationv1.TokenRequest{
			Spec: authenticationv1.TokenRequestSpec{
				Audiences:         []string{"api"},
				ExpirationSeconds: &one,
				BoundObjectRef: &authenticationv1.BoundObjectReference{
					Kind:       "Pod",
					APIVersion: "v1",
					Name:       pod.Name,
				},
			},
		}

		if resp, err := cs.CoreV1().ServiceAccounts(sa.Namespace).CreateToken(sa.Name, treq); err == nil {
			t.Fatalf("expected err creating token for nonexistant svcacct but got: %#v", resp)
		}
		sa, del := createDeleteSvcAcct(t, cs, sa)
		defer del()

		if resp, err := cs.CoreV1().ServiceAccounts(sa.Namespace).CreateToken(sa.Name, treq); err == nil {
			t.Fatalf("expected err creating token bound to nonexistant pod but got: %#v", resp)
		}
		pod, delPod := createDeletePod(t, cs, pod)
		defer delPod()

		// right uid
		treq.Spec.BoundObjectRef.UID = pod.UID
		if _, err := cs.CoreV1().ServiceAccounts(sa.Namespace).CreateToken(sa.Name, treq); err != nil {
			t.Fatalf("err: %v", err)
		}
		// wrong uid
		treq.Spec.BoundObjectRef.UID = wrongUID
		if resp, err := cs.CoreV1().ServiceAccounts(sa.Namespace).CreateToken(sa.Name, treq); err == nil {
			t.Fatalf("expected err creating token bound to pod with wrong uid but got: %#v", resp)
		}
		// no uid
		treq.Spec.BoundObjectRef.UID = noUID
		treq, err = cs.CoreV1().ServiceAccounts(sa.Namespace).CreateToken(sa.Name, treq)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		checkPayload(t, treq.Status.Token, `"system:serviceaccount:myns:test-svcacct"`, "sub")
		checkPayload(t, treq.Status.Token, `["api"]`, "aud")
		checkPayload(t, treq.Status.Token, `"test-pod"`, "kubernetes.io", "pod", "name")
		checkPayload(t, treq.Status.Token, "null", "kubernetes.io", "secret")
		checkPayload(t, treq.Status.Token, `"myns"`, "kubernetes.io", "namespace")
		checkPayload(t, treq.Status.Token, `"test-svcacct"`, "kubernetes.io", "serviceaccount", "name")

		doTokenReview(t, cs, treq, false)
		delPod()
		doTokenReview(t, cs, treq, true)
	})

	t.Run("bound to service account and secret", func(t *testing.T) {
		treq := &authenticationv1.TokenRequest{
			Spec: authenticationv1.TokenRequestSpec{
				Audiences:         []string{"api"},
				ExpirationSeconds: &one,
				BoundObjectRef: &authenticationv1.BoundObjectReference{
					Kind:       "Secret",
					APIVersion: "v1",
					Name:       secret.Name,
					UID:        secret.UID,
				},
			},
		}

		if resp, err := cs.CoreV1().ServiceAccounts(sa.Namespace).CreateToken(sa.Name, treq); err == nil {
			t.Fatalf("expected err creating token for nonexistant svcacct but got: %#v", resp)
		}
		sa, del := createDeleteSvcAcct(t, cs, sa)
		defer del()

		if resp, err := cs.CoreV1().ServiceAccounts(sa.Namespace).CreateToken(sa.Name, treq); err == nil {
			t.Fatalf("expected err creating token bound to nonexistant secret but got: %#v", resp)
		}
		secret, delSecret := createDeleteSecret(t, cs, secret)
		defer delSecret()

		// right uid
		treq.Spec.BoundObjectRef.UID = secret.UID
		if _, err := cs.CoreV1().ServiceAccounts(sa.Namespace).CreateToken(sa.Name, treq); err != nil {
			t.Fatalf("err: %v", err)
		}
		// wrong uid
		treq.Spec.BoundObjectRef.UID = wrongUID
		if resp, err := cs.CoreV1().ServiceAccounts(sa.Namespace).CreateToken(sa.Name, treq); err == nil {
			t.Fatalf("expected err creating token bound to secret with wrong uid but got: %#v", resp)
		}
		// no uid
		treq.Spec.BoundObjectRef.UID = noUID
		treq, err = cs.CoreV1().ServiceAccounts(sa.Namespace).CreateToken(sa.Name, treq)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		checkPayload(t, treq.Status.Token, `"system:serviceaccount:myns:test-svcacct"`, "sub")
		checkPayload(t, treq.Status.Token, `["api"]`, "aud")
		checkPayload(t, treq.Status.Token, `null`, "kubernetes.io", "pod")
		checkPayload(t, treq.Status.Token, `"test-secret"`, "kubernetes.io", "secret", "name")
		checkPayload(t, treq.Status.Token, `"myns"`, "kubernetes.io", "namespace")
		checkPayload(t, treq.Status.Token, `"test-svcacct"`, "kubernetes.io", "serviceaccount", "name")

		doTokenReview(t, cs, treq, false)
		delSecret()
		doTokenReview(t, cs, treq, true)
	})

	t.Run("bound to service account and pod running as different service account", func(t *testing.T) {
		treq := &authenticationv1.TokenRequest{
			Spec: authenticationv1.TokenRequestSpec{
				Audiences:         []string{"api"},
				ExpirationSeconds: &one,
				BoundObjectRef: &authenticationv1.BoundObjectReference{
					Kind:       "Pod",
					APIVersion: "v1",
					Name:       otherpod.Name,
				},
			},
		}

		sa, del := createDeleteSvcAcct(t, cs, sa)
		defer del()
		_, del = createDeletePod(t, cs, otherpod)
		defer del()

		if resp, err := cs.CoreV1().ServiceAccounts(sa.Namespace).CreateToken(sa.Name, treq); err == nil {
			t.Fatalf("expected err but got: %#v", resp)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		treq := &authenticationv1.TokenRequest{
			Spec: authenticationv1.TokenRequestSpec{
				Audiences:         []string{"api"},
				ExpirationSeconds: &one,
			},
		}

		sa, del := createDeleteSvcAcct(t, cs, sa)
		defer del()

		treq, err = cs.CoreV1().ServiceAccounts(sa.Namespace).CreateToken(sa.Name, treq)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		doTokenReview(t, cs, treq, false)
		time.Sleep(63 * time.Second)
		doTokenReview(t, cs, treq, true)
	})

	t.Run("a token without an api audience is invalid", func(t *testing.T) {
		treq := &authenticationv1.TokenRequest{
			Spec: authenticationv1.TokenRequestSpec{
				Audiences: []string{"not-the-api"},
			},
		}

		sa, del := createDeleteSvcAcct(t, cs, sa)
		defer del()

		treq, err = cs.CoreV1().ServiceAccounts(sa.Namespace).CreateToken(sa.Name, treq)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		doTokenReview(t, cs, treq, true)
	})

	t.Run("a tokenrequest without an audience is valid against the api", func(t *testing.T) {
		treq := &authenticationv1.TokenRequest{
			Spec: authenticationv1.TokenRequestSpec{},
		}

		sa, del := createDeleteSvcAcct(t, cs, sa)
		defer del()

		treq, err = cs.CoreV1().ServiceAccounts(sa.Namespace).CreateToken(sa.Name, treq)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		checkPayload(t, treq.Status.Token, `["api"]`, "aud")

		doTokenReview(t, cs, treq, false)
	})

	t.Run("a token should be invalid after recreating same name pod", func(t *testing.T) {
		treq := &authenticationv1.TokenRequest{
			Spec: authenticationv1.TokenRequestSpec{
				Audiences:         []string{"api"},
				ExpirationSeconds: &one,
				BoundObjectRef: &authenticationv1.BoundObjectReference{
					Kind:       "Pod",
					APIVersion: "v1",
					Name:       pod.Name,
				},
			},
		}

		sa, del := createDeleteSvcAcct(t, cs, sa)
		defer del()
		originalPod, originalDelPod := createDeletePod(t, cs, pod)
		defer originalDelPod()

		treq.Spec.BoundObjectRef.UID = originalPod.UID
		if treq, err = cs.CoreV1().ServiceAccounts(sa.Namespace).CreateToken(sa.Name, treq); err != nil {
			t.Fatalf("err: %v", err)
		}

		checkPayload(t, treq.Status.Token, `"system:serviceaccount:myns:test-svcacct"`, "sub")
		checkPayload(t, treq.Status.Token, `["api"]`, "aud")
		checkPayload(t, treq.Status.Token, `"test-pod"`, "kubernetes.io", "pod", "name")
		checkPayload(t, treq.Status.Token, "null", "kubernetes.io", "secret")
		checkPayload(t, treq.Status.Token, `"myns"`, "kubernetes.io", "namespace")
		checkPayload(t, treq.Status.Token, `"test-svcacct"`, "kubernetes.io", "serviceaccount", "name")

		doTokenReview(t, cs, treq, false)
		originalDelPod()
		doTokenReview(t, cs, treq, true)

		_, recreateDelPod := createDeletePod(t, cs, pod)
		defer recreateDelPod()

		doTokenReview(t, cs, treq, true)
	})

	t.Run("a token should be invalid after recreating same name secret", func(t *testing.T) {
		treq := &authenticationv1.TokenRequest{
			Spec: authenticationv1.TokenRequestSpec{
				Audiences:         []string{"api"},
				ExpirationSeconds: &one,
				BoundObjectRef: &authenticationv1.BoundObjectReference{
					Kind:       "Secret",
					APIVersion: "v1",
					Name:       secret.Name,
					UID:        secret.UID,
				},
			},
		}

		sa, del := createDeleteSvcAcct(t, cs, sa)
		defer del()

		originalSecret, originalDelSecret := createDeleteSecret(t, cs, secret)
		defer originalDelSecret()

		treq.Spec.BoundObjectRef.UID = originalSecret.UID
		if treq, err = cs.CoreV1().ServiceAccounts(sa.Namespace).CreateToken(sa.Name, treq); err != nil {
			t.Fatalf("err: %v", err)
		}

		checkPayload(t, treq.Status.Token, `"system:serviceaccount:myns:test-svcacct"`, "sub")
		checkPayload(t, treq.Status.Token, `["api"]`, "aud")
		checkPayload(t, treq.Status.Token, `null`, "kubernetes.io", "pod")
		checkPayload(t, treq.Status.Token, `"test-secret"`, "kubernetes.io", "secret", "name")
		checkPayload(t, treq.Status.Token, `"myns"`, "kubernetes.io", "namespace")
		checkPayload(t, treq.Status.Token, `"test-svcacct"`, "kubernetes.io", "serviceaccount", "name")

		doTokenReview(t, cs, treq, false)
		originalDelSecret()
		doTokenReview(t, cs, treq, true)

		_, recreateDelSecret := createDeleteSecret(t, cs, secret)
		defer recreateDelSecret()

		doTokenReview(t, cs, treq, true)
	})
}

func doTokenReview(t *testing.T, cs externalclientset.Interface, treq *authenticationv1.TokenRequest, expectErr bool) {
	t.Helper()
	trev, err := cs.AuthenticationV1().TokenReviews().Create(&authenticationv1.TokenReview{
		Spec: authenticationv1.TokenReviewSpec{
			Token: treq.Status.Token,
		},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	t.Logf("status: %+v", trev.Status)
	if (trev.Status.Error != "") && !expectErr {
		t.Fatalf("expected no error but got: %v", trev.Status.Error)
	}
	if (trev.Status.Error == "") && expectErr {
		t.Fatalf("expected error but got: %+v", trev.Status)
	}
	if !trev.Status.Authenticated && !expectErr {
		t.Fatal("expected token to be authenticated but it wasn't")
	}
}

func checkPayload(t *testing.T, tok string, want string, parts ...string) {
	t.Helper()
	got := getSubObject(t, getPayload(t, tok), parts...)
	if got != want {
		t.Errorf("unexpected payload.\nsaw:\t%v\nwant:\t%v", got, want)
	}
}

func getSubObject(t *testing.T, b string, parts ...string) string {
	t.Helper()
	var obj interface{}
	obj = make(map[string]interface{})
	if err := json.Unmarshal([]byte(b), &obj); err != nil {
		t.Fatalf("err: %v", err)
	}
	for _, part := range parts {
		obj = obj.(map[string]interface{})[part]
	}
	out, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	return string(out)
}

func getPayload(t *testing.T, b string) string {
	t.Helper()
	parts := strings.Split(b, ".")
	if len(parts) != 3 {
		t.Fatalf("token did not have three parts: %v", b)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("failed to base64 decode token: %v", err)
	}
	return string(payload)
}

func createDeleteSvcAcct(t *testing.T, cs clientset.Interface, sa *v1.ServiceAccount) (*v1.ServiceAccount, func()) {
	t.Helper()
	sa, err := cs.CoreV1().ServiceAccounts(sa.Namespace).Create(sa)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	done := false
	return sa, func() {
		t.Helper()
		if done {
			return
		}
		done = true
		if err := cs.CoreV1().ServiceAccounts(sa.Namespace).Delete(sa.Name, nil); err != nil {
			t.Fatalf("err: %v", err)
		}
	}
}

func createDeletePod(t *testing.T, cs clientset.Interface, pod *v1.Pod) (*v1.Pod, func()) {
	t.Helper()
	pod, err := cs.CoreV1().Pods(pod.Namespace).Create(pod)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	done := false
	return pod, func() {
		t.Helper()
		if done {
			return
		}
		done = true
		if err := cs.CoreV1().Pods(pod.Namespace).Delete(pod.Name, nil); err != nil {
			t.Fatalf("err: %v", err)
		}
	}
}

func createDeleteSecret(t *testing.T, cs clientset.Interface, sec *v1.Secret) (*v1.Secret, func()) {
	t.Helper()
	sec, err := cs.CoreV1().Secrets(sec.Namespace).Create(sec)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	done := false
	return sec, func() {
		t.Helper()
		if done {
			return
		}
		done = true
		if err := cs.CoreV1().Secrets(sec.Namespace).Delete(sec.Name, nil); err != nil {
			t.Fatalf("err: %v", err)
		}
	}
}
