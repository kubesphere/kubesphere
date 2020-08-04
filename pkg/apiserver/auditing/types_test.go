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

package auditing

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/auditregistration/v1alpha1"
	v1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/apiserver/pkg/authentication/user"
	k8srequest "k8s.io/apiserver/pkg/endpoints/request"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	auditingv1alpha1 "kubesphere.io/kubesphere/pkg/apis/auditing/v1alpha1"
	v1alpha12 "kubesphere.io/kubesphere/pkg/apiserver/auditing/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/request"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/utils/iputil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

var noResyncPeriodFunc = func() time.Duration { return 0 }

func TestGetAuditLevel(t *testing.T) {
	webhook := &auditingv1alpha1.Webhook{
		TypeMeta: metav1.TypeMeta{
			APIVersion: auditingv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-auditing-webhook",
		},
		Spec: auditingv1alpha1.WebhookSpec{
			AuditLevel: v1alpha1.LevelRequestResponse,
		},
	}

	ksClient := fake.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	fakeInformerFactory := informers.NewInformerFactories(k8sClient, ksClient, nil, nil, nil, nil)

	a := auditing{
		webhookLister: fakeInformerFactory.KubeSphereSharedInformerFactory().Auditing().V1alpha1().Webhooks().Lister(),
	}

	err := fakeInformerFactory.KubeSphereSharedInformerFactory().Auditing().V1alpha1().Webhooks().Informer().GetIndexer().Add(webhook)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, string(webhook.Spec.AuditLevel), string(a.getAuditLevel()))
}

func TestAuditing_Enabled(t *testing.T) {
	webhook := &auditingv1alpha1.Webhook{
		TypeMeta: metav1.TypeMeta{
			APIVersion: auditingv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-auditing-webhook",
		},
		Spec: auditingv1alpha1.WebhookSpec{
			AuditLevel: v1alpha1.LevelNone,
		},
	}

	ksClient := fake.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	fakeInformerFactory := informers.NewInformerFactories(k8sClient, ksClient, nil, nil, nil, nil)

	a := auditing{
		webhookLister: fakeInformerFactory.KubeSphereSharedInformerFactory().Auditing().V1alpha1().Webhooks().Lister(),
	}

	err := fakeInformerFactory.KubeSphereSharedInformerFactory().Auditing().V1alpha1().Webhooks().Informer().GetIndexer().Add(webhook)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, false, a.Enabled())
}

func TestAuditing_K8sAuditingEnabled(t *testing.T) {
	webhook := &auditingv1alpha1.Webhook{
		TypeMeta: metav1.TypeMeta{
			APIVersion: auditingv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-auditing-webhook",
		},
		Spec: auditingv1alpha1.WebhookSpec{
			AuditLevel:         v1alpha1.LevelNone,
			K8sAuditingEnabled: true,
		},
	}

	ksClient := fake.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	fakeInformerFactory := informers.NewInformerFactories(k8sClient, ksClient, nil, nil, nil, nil)

	a := auditing{
		webhookLister: fakeInformerFactory.KubeSphereSharedInformerFactory().Auditing().V1alpha1().Webhooks().Lister(),
	}

	err := fakeInformerFactory.KubeSphereSharedInformerFactory().Auditing().V1alpha1().Webhooks().Informer().GetIndexer().Add(webhook)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, true, a.K8sAuditingEnabled())
}

func TestAuditing_LogRequestObject(t *testing.T) {
	webhook := &auditingv1alpha1.Webhook{
		TypeMeta: metav1.TypeMeta{
			APIVersion: auditingv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-auditing-webhook",
		},
		Spec: auditingv1alpha1.WebhookSpec{
			AuditLevel:         v1alpha1.LevelRequestResponse,
			K8sAuditingEnabled: true,
		},
	}

	ksClient := fake.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	fakeInformerFactory := informers.NewInformerFactories(k8sClient, ksClient, nil, nil, nil, nil)

	a := auditing{
		webhookLister: fakeInformerFactory.KubeSphereSharedInformerFactory().Auditing().V1alpha1().Webhooks().Lister(),
	}

	err := fakeInformerFactory.KubeSphereSharedInformerFactory().Auditing().V1alpha1().Webhooks().Informer().GetIndexer().Add(webhook)
	if err != nil {
		panic(err)
	}

	req := &http.Request{}
	u, err := url.Parse("http://139.198.121.143:32306//kapis/tenant.kubesphere.io/v1alpha2/workspaces")
	if err != nil {
		panic(err)
	}

	req.URL = u
	req.Header = http.Header{}
	req.Header.Add(iputil.XClientIP, "192.168.0.2")
	req = req.WithContext(request.WithUser(req.Context(), &user.DefaultInfo{
		Name: "admin",
		Groups: []string{
			"system",
		},
	}))

	info := &request.RequestInfo{
		RequestInfo: &k8srequest.RequestInfo{
			IsResourceRequest: false,
			Path:              "/kapis/tenant.kubesphere.io/v1alpha2/workspaces",
			Verb:              "create",
			APIGroup:          "tenant.kubesphere.io",
			APIVersion:        "v1alpha2",
			Resource:          "workspaces",
			Name:              "test",
		},
	}

	e := a.LogRequestObject(req, info)

	expectedEvent := &v1alpha12.Event{
		Event: audit.Event{
			AuditID: e.AuditID,
			Level:   "RequestResponse",
			Verb:    "create",
			Stage:   "ResponseComplete",
			User: v1.UserInfo{
				Username: "admin",
				Groups: []string{
					"system",
				},
			},
			SourceIPs: []string{
				"192.168.0.2",
			},
			RequestURI:               "/kapis/tenant.kubesphere.io/v1alpha2/workspaces",
			RequestReceivedTimestamp: e.RequestReceivedTimestamp,
			ObjectRef: &audit.ObjectReference{
				Resource:        "workspaces",
				Namespace:       "",
				Name:            "test",
				UID:             "",
				APIGroup:        "tenant.kubesphere.io",
				APIVersion:      "v1alpha2",
				ResourceVersion: "",
				Subresource:     "",
			},
		},
	}

	assert.Equal(t, expectedEvent, e)
}

func TestAuditing_LogResponseObject(t *testing.T) {
	webhook := &auditingv1alpha1.Webhook{
		TypeMeta: metav1.TypeMeta{
			APIVersion: auditingv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-auditing-webhook",
		},
		Spec: auditingv1alpha1.WebhookSpec{
			AuditLevel:         v1alpha1.LevelMetadata,
			K8sAuditingEnabled: true,
		},
	}

	ksClient := fake.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	fakeInformerFactory := informers.NewInformerFactories(k8sClient, ksClient, nil, nil, nil, nil)

	a := auditing{
		webhookLister: fakeInformerFactory.KubeSphereSharedInformerFactory().Auditing().V1alpha1().Webhooks().Lister(),
	}

	err := fakeInformerFactory.KubeSphereSharedInformerFactory().Auditing().V1alpha1().Webhooks().Informer().GetIndexer().Add(webhook)
	if err != nil {
		panic(err)
	}

	req := &http.Request{}
	u, err := url.Parse("http://139.198.121.143:32306//kapis/tenant.kubesphere.io/v1alpha2/workspaces")
	if err != nil {
		panic(err)
	}

	req.URL = u
	req.Header = http.Header{}
	req.Header.Add(iputil.XClientIP, "192.168.0.2")
	req = req.WithContext(request.WithUser(req.Context(), &user.DefaultInfo{
		Name: "admin",
		Groups: []string{
			"system",
		},
	}))

	info := &request.RequestInfo{
		RequestInfo: &k8srequest.RequestInfo{
			IsResourceRequest: false,
			Path:              "/kapis/tenant.kubesphere.io/v1alpha2/workspaces",
			Verb:              "create",
			APIGroup:          "tenant.kubesphere.io",
			APIVersion:        "v1alpha2",
			Resource:          "workspaces",
			Name:              "test",
		},
	}

	e := a.LogRequestObject(req, info)

	resp := NewResponseCapture(httptest.NewRecorder())
	resp.WriteHeader(200)

	a.LogResponseObject(e, resp)

	expectedEvent := &v1alpha12.Event{
		Event: audit.Event{
			Verb:    "create",
			AuditID: e.AuditID,
			Level:   "Metadata",
			Stage:   "ResponseComplete",
			User: v1.UserInfo{
				Username: "admin",
				Groups: []string{
					"system",
				},
			},
			SourceIPs: []string{
				"192.168.0.2",
			},
			ObjectRef: &audit.ObjectReference{
				Resource:   "workspaces",
				Name:       "test",
				APIGroup:   "tenant.kubesphere.io",
				APIVersion: "v1alpha2",
			},

			RequestReceivedTimestamp: e.RequestReceivedTimestamp,
			StageTimestamp:           e.StageTimestamp,
			RequestURI:               "/kapis/tenant.kubesphere.io/v1alpha2/workspaces",
			ResponseStatus: &metav1.Status{
				Code: 200,
			},
		},
	}

	expectedBs, err := json.Marshal(expectedEvent)
	if err != nil {
		panic(err)
	}
	bs, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	assert.EqualValues(t, string(expectedBs), string(bs))
}

func TestResponseCapture_WriteHeader(t *testing.T) {
	record := httptest.NewRecorder()
	resp := NewResponseCapture(record)

	resp.WriteHeader(404)

	assert.EqualValues(t, 404, resp.StatusCode())
	assert.EqualValues(t, 404, record.Code)
}

func TestResponseCapture_Write(t *testing.T) {

	record := httptest.NewRecorder()
	resp := NewResponseCapture(record)

	body := []byte("123")

	_, err := resp.Write(body)
	if err != nil {
		panic(err)
	}

	assert.EqualValues(t, body, resp.Bytes())
	assert.EqualValues(t, body, record.Body.Bytes())
}
