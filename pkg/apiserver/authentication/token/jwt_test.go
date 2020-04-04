/*
 *
 * Copyright 2020 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package token

import (
	"github.com/google/go-cmp/cmp"
	"k8s.io/apiserver/pkg/authentication/user"
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	"kubesphere.io/kubesphere/pkg/simple/client/cache"
	"testing"
)

func TestJwtTokenIssuer(t *testing.T) {
	options := authoptions.NewAuthenticateOptions()
	options.JwtSecret = "kubesphere"
	issuer := NewJwtTokenIssuer(DefaultIssuerName, options, cache.NewSimpleCache())

	testCases := []struct {
		description string
		name        string
		uid         string
		email       string
	}{
		{
			name: "admin",
			uid:  "b8be6edd-2c92-4535-9b2a-df6326474458",
		},
		{
			name: "bar",
			uid:  "b8be6edd-2c92-4535-9b2a-df6326474452",
		},
	}

	for _, testCase := range testCases {
		user := &user.DefaultInfo{
			Name: testCase.name,
			UID:  testCase.uid,
		}

		t.Run(testCase.description, func(t *testing.T) {
			token, err := issuer.IssueTo(user, 0)
			if err != nil {
				t.Fatal(err)
			}

			got, err := issuer.Verify(token)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(user, got); len(diff) != 0 {
				t.Errorf("%T differ (-got, +expected), %s", user, diff)
			}
		})
	}
}
