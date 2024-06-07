/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package oauth

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

func TestMarshalInto(t *testing.T) {
	want := &Client{
		Name:                                "test",
		Secret:                              "test",
		Trusted:                             false,
		GrantMethod:                         "auto",
		RedirectURIs:                        []string{"test"},
		AccessTokenMaxAgeSeconds:            10000,
		AccessTokenInactivityTimeoutSeconds: 10000,
	}
	secret := &v1.Secret{}
	if err := MarshalInto(want, secret); err != nil {
		t.Errorf("Error: %v", err)
	}

	got, err := UnmarshalFrom(secret)
	if err != nil {
		klog.Errorf("failed to unmarshal secret data: %s", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("got %v, want %v", got, want)
	}
}
