/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package registries

import (
	"strings"
	"testing"
)

type authServiceMock struct {
	service string
	realm   string
	scope   []string
}

type challengeTestCase struct {
	header      string
	errorString string
	value       authServiceMock
}

func (asm authServiceMock) equalTo(v *authService) bool {
	if asm.service != v.Service {
		return false
	}
	for i, v := range v.Scope {
		if v != asm.scope[i] {
			return false
		}
	}

	return asm.realm == v.Realm.String()
}

func TestToken(t *testing.T) {
	testImage := Image{Domain: "docker.io", Path: "library/alpine", Tag: "latest"}
	r, err := CreateRegistryClient("", "", "docker.io", true, false)
	if err != nil {
		t.Fatalf("Could not get registry client: %s", err)
	}

	digestUrl := r.GetDigestUrl(testImage)

	// Get token.
	token, err := r.Token(digestUrl)
	if err != nil || token == "" {
		t.Fatalf("Could not get token: %s", err)
	}

}

func TestParseChallenge(t *testing.T) {
	challengeHeaderCases := []challengeTestCase{
		{
			header: `Bearer realm="https://foobar.com/api/v1/token",service=foobar.com,scope=""`,
			value: authServiceMock{
				service: "foobar.com",
				realm:   "https://foobar.com/api/v1/token",
			},
		},
		{
			header: `Bearer realm="https://r.j3ss.co/auth",service="Docker registry",scope="repository:chrome:pull"`,
			value: authServiceMock{
				service: "Docker registry",
				realm:   "https://r.j3ss.co/auth",
				scope:   []string{"repository:chrome:pull"},
			},
		},
		{
			header:      `Basic realm="https://r.j3ss.co/auth",service="Docker registry"`,
			errorString: "basic auth required",
		},
		{
			header:      `Basic realm="Registry Realm",service="Docker registry"`,
			errorString: "basic auth required",
		},
	}

	for _, tc := range challengeHeaderCases {
		val, err := parseChallenge(tc.header)
		if err != nil && !strings.Contains(err.Error(), tc.errorString) {
			t.Fatalf("expected error to contain %v,  got %s", tc.errorString, err)
		}
		if err == nil && !tc.value.equalTo(val) {
			t.Fatalf("got %v, expected %v", val, tc.value)
		}

	}
}
