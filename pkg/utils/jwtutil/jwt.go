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
package jwtutil

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
)

const secretEnv = "JWT_SECRET"

var secret []byte

func Setup(key string) {
	secret = []byte(key)
}

func MustSigned(claims jwt.MapClaims) string {
	uToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := uToken.SignedString(secret)
	if err != nil {
		panic(err)
	}
	return token
}

func provideKey(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
		return secret, nil
	} else {
		return nil, fmt.Errorf("expect token signed with HMAC but got %v", token.Header["alg"])
	}
}

func ValidateToken(uToken string) (*jwt.Token, error) {

	if len(uToken) == 0 {
		return nil, fmt.Errorf("token length is zero")
	}

	token, err := jwt.Parse(uToken, provideKey)

	if err != nil {
		return nil, err
	}

	return token, nil
}
