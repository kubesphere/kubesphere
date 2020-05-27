/*
Copyright 2019 The KubeSphere Authors.

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

package ldap

import (
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"testing"
)

func TestSimpleLdap(t *testing.T) {
	ldapClient := NewSimpleLdap()

	foo := &iamv1alpha2.User{
		TypeMeta: metav1.TypeMeta{APIVersion: iamv1alpha2.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name: "jerry",
		},
		Spec: iamv1alpha2.UserSpec{
			Email:             "jerry@kubesphere.io",
			Lang:              "en",
			Description:       "Jerry is kind and gentle.",
			Groups:            []string{},
			EncryptedPassword: "P@88w0rd",
		},
	}

	t.Run("should create user", func(t *testing.T) {
		err := ldapClient.Create(foo)
		if err != nil {
			t.Fatal(err)
		}

		// check if user really created
		user, err := ldapClient.Get(foo.Name)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(user, foo); len(diff) != 0 {
			t.Fatalf("%T differ (-got, +want): %s", user, diff)
		}

		_ = ldapClient.Delete(foo.Name)
	})

	t.Run("should update user", func(t *testing.T) {
		err := ldapClient.Create(foo)
		if err != nil {
			t.Fatal(err)
		}

		foo.Spec.Description = "Jerry needs some drinks."
		err = ldapClient.Update(foo)
		if err != nil {
			t.Fatal(err)
		}

		// check if user really created
		user, err := ldapClient.Get(foo.Name)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(user, foo); len(diff) != 0 {
			t.Fatalf("%T differ (-got, +want): %s", user, diff)
		}

		_ = ldapClient.Delete(foo.Name)
	})

	t.Run("should delete user", func(t *testing.T) {
		err := ldapClient.Create(foo)
		if err != nil {
			t.Fatal(err)
		}

		err = ldapClient.Delete(foo.Name)
		if err != nil {
			t.Fatal(err)
		}

		_, err = ldapClient.Get(foo.Name)
		if err == nil || err != ErrUserNotExists {
			t.Fatalf("expected ErrUserNotExists error, got %v", err)
		}
	})

	t.Run("should verify username and password", func(t *testing.T) {
		err := ldapClient.Create(foo)
		if err != nil {
			t.Fatal(err)
		}

		err = ldapClient.Authenticate(foo.Name, foo.Spec.EncryptedPassword)
		if err != nil {
			t.Fatalf("should pass but got an error %v", err)
		}

		err = ldapClient.Authenticate(foo.Name, "gibberish")
		if err == nil || err != ErrInvalidCredentials {
			t.Fatalf("expected error ErrInvalidCrenentials but got %v", err)
		}
	})
}
