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

package log

import (
	"bytes"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/pborman/uuid"

	"k8s.io/apimachinery/pkg/apimachinery/registered"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	auditinternal "k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/apiserver/pkg/apis/audit/install"
	auditv1beta1 "k8s.io/apiserver/pkg/apis/audit/v1beta1"
	"k8s.io/apiserver/pkg/audit"
)

// NOTE: Copied from webhook backend to register auditv1beta1 to scheme
var (
	registry = registered.NewAPIRegistrationManager()
)

func init() {
	allGVs := []schema.GroupVersion{auditv1beta1.SchemeGroupVersion}
	registry.RegisterVersions(allGVs)
	install.Install(registry, audit.Scheme)
}

func TestLogEventsLegacy(t *testing.T) {
	for _, test := range []struct {
		event    *auditinternal.Event
		expected string
	}{
		{
			&auditinternal.Event{
				AuditID: types.UID(uuid.NewRandom().String()),
			},
			`[\d\:\-\.\+TZ]+ AUDIT: id="[\w-]+" stage="" ip="<unknown>" method="" user="<none>" groups="<none>" as="<self>" asgroups="<lookup>" namespace="<none>" uri="" response="<deferred>"`,
		},
		{
			&auditinternal.Event{
				ResponseStatus: &metav1.Status{
					Code: 200,
				},
				RequestURI: "/apis/rbac.authorization.k8s.io/v1/roles",
				SourceIPs: []string{
					"127.0.0.1",
				},
				RequestReceivedTimestamp: metav1.NewMicroTime(time.Now()),
				AuditID:                  types.UID(uuid.NewRandom().String()),
				Stage:                    auditinternal.StageRequestReceived,
				Verb:                     "get",
				User: auditinternal.UserInfo{
					Username: "admin",
					Groups: []string{
						"system:masters",
						"system:authenticated",
					},
				},
				ObjectRef: &auditinternal.ObjectReference{
					Namespace: "default",
				},
			},
			`[\d\:\-\.\+TZ]+ AUDIT: id="[\w-]+" stage="RequestReceived" ip="127.0.0.1" method="get" user="admin" groups="\\"system:masters\\",\\"system:authenticated\\"" as="<self>" asgroups="<lookup>" namespace="default" uri="/apis/rbac.authorization.k8s.io/v1/roles" response="200"`,
		},
		{
			&auditinternal.Event{
				AuditID: types.UID(uuid.NewRandom().String()),
				Level:   auditinternal.LevelMetadata,
				ObjectRef: &auditinternal.ObjectReference{
					Resource:    "foo",
					APIVersion:  "v1",
					Subresource: "bar",
				},
			},
			`[\d\:\-\.\+TZ]+ AUDIT: id="[\w-]+" stage="" ip="<unknown>" method="" user="<none>" groups="<none>" as="<self>" asgroups="<lookup>" namespace="<none>" uri="" response="<deferred>"`,
		},
	} {
		var buf bytes.Buffer
		backend := NewBackend(&buf, FormatLegacy, auditv1beta1.SchemeGroupVersion)
		backend.ProcessEvents(test.event)
		match, err := regexp.MatchString(test.expected, buf.String())
		if err != nil {
			t.Errorf("Unexpected error matching line %v", err)
			continue
		}
		if !match {
			t.Errorf("Unexpected line of audit: %s", buf.String())
		}
	}
}

func TestLogEventsJson(t *testing.T) {
	for _, event := range []*auditinternal.Event{
		{
			AuditID: types.UID(uuid.NewRandom().String()),
		},
		{
			ResponseStatus: &metav1.Status{
				Code: 200,
			},
			RequestURI: "/apis/rbac.authorization.k8s.io/v1/roles",
			SourceIPs: []string{
				"127.0.0.1",
			},
			RequestReceivedTimestamp: metav1.NewMicroTime(time.Now().Truncate(time.Microsecond)),
			StageTimestamp:           metav1.NewMicroTime(time.Now().Truncate(time.Microsecond)),
			AuditID:                  types.UID(uuid.NewRandom().String()),
			Stage:                    auditinternal.StageRequestReceived,
			Verb:                     "get",
			User: auditinternal.UserInfo{
				Username: "admin",
				Groups: []string{
					"system:masters",
					"system:authenticated",
				},
			},
			ObjectRef: &auditinternal.ObjectReference{
				Namespace: "default",
			},
		},
		{
			AuditID: types.UID(uuid.NewRandom().String()),
			Level:   auditinternal.LevelMetadata,
			ObjectRef: &auditinternal.ObjectReference{
				Resource:    "foo",
				APIVersion:  "v1",
				Subresource: "bar",
			},
		},
	} {
		var buf bytes.Buffer
		backend := NewBackend(&buf, FormatJson, auditv1beta1.SchemeGroupVersion)
		backend.ProcessEvents(event)
		// decode events back and compare with the original one.
		result := &auditinternal.Event{}
		decoder := audit.Codecs.UniversalDecoder(auditv1beta1.SchemeGroupVersion)
		if err := runtime.DecodeInto(decoder, buf.Bytes(), result); err != nil {
			t.Errorf("failed decoding buf: %s", buf.String())
			continue
		}
		if !reflect.DeepEqual(event, result) {
			t.Errorf("The result event should be the same with the original one, \noriginal: \n%#v\n result: \n%#v", event, result)
		}
	}
}
