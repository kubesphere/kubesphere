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
package authenticate

import (
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apigateway/caddy-plugin/internal"
	"kubesphere.io/kubesphere/pkg/simple/client/redis"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

type Auth struct {
	Rule *Rule
	Next httpserver.Handler
}

type Rule struct {
	Secret           []byte
	Path             string
	RedisOptions     *redis.RedisOptions
	TokenIdleTimeout time.Duration
	RedisClient      *redis.RedisClient
	ExclusionRules   []internal.ExclusionRule
}

type User struct {
	Username string                  `json:"username"`
	UID      string                  `json:"uid"`
	Groups   *[]string               `json:"groups,omitempty"`
	Extra    *map[string]interface{} `json:"extra,omitempty"`
}

var requestInfoFactory = request.RequestInfoFactory{
	APIPrefixes:          sets.NewString("api", "apis", "kapis", "kapi"),
	GrouplessAPIPrefixes: sets.NewString("api")}

func (h Auth) ServeHTTP(resp http.ResponseWriter, req *http.Request) (int, error) {
	for _, rule := range h.Rule.ExclusionRules {
		if httpserver.Path(req.URL.Path).Matches(rule.Path) && (rule.Method == internal.AllMethod || req.Method == rule.Method) {
			return h.Next.ServeHTTP(resp, req)
		}
	}

	if httpserver.Path(req.URL.Path).Matches(h.Rule.Path) {

		uToken, err := h.ExtractToken(req)

		if err != nil {
			return h.HandleUnauthorized(resp, err), nil
		}

		token, err := h.Validate(uToken)

		if err != nil {
			return h.HandleUnauthorized(resp, err), nil
		}

		req, err = h.InjectContext(req, token)

		if err != nil {
			return h.HandleUnauthorized(resp, err), nil
		}
	}

	return h.Next.ServeHTTP(resp, req)
}

func (h Auth) InjectContext(req *http.Request, token *jwt.Token) (*http.Request, error) {

	payload, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return nil, errors.New("invalid payload")
	}

	for header := range req.Header {
		if strings.HasPrefix(header, "X-Token-") {
			req.Header.Del(header)
		}
	}

	usr := &user.DefaultInfo{}

	username, ok := payload["username"].(string)

	if ok && username != "" {
		req.Header.Set("X-Token-Username", username)
		usr.Name = username
	}

	uid := payload["uid"]

	if uid != nil {
		switch uid.(type) {
		case int:
			req.Header.Set("X-Token-UID", strconv.Itoa(uid.(int)))
			usr.UID = strconv.Itoa(uid.(int))
			break
		case string:
			req.Header.Set("X-Token-UID", uid.(string))
			usr.UID = uid.(string)
			break
		}
	}

	groups, ok := payload["groups"].([]string)
	if ok && len(groups) > 0 {
		req.Header.Set("X-Token-Groups", strings.Join(groups, ","))
		usr.Groups = groups
	}

	// hard code, support jenkins auth plugin
	if httpserver.Path(req.URL.Path).Matches("/kapis/jenkins.kubesphere.io") ||
		httpserver.Path(req.URL.Path).Matches("job") ||
		httpserver.Path(req.URL.Path).Matches("/kapis/devops.kubesphere.io/v1alpha2") {
		req.SetBasicAuth(username, token.Raw)
	}

	context := request.WithUser(req.Context(), usr)

	requestInfo, err := requestInfoFactory.NewRequestInfo(req)

	if err == nil {
		context = request.WithRequestInfo(context, requestInfo)
	} else {
		return nil, err
	}

	req = req.WithContext(context)

	return req, nil
}

func (h Auth) Validate(uToken string) (*jwt.Token, error) {

	if len(uToken) == 0 {
		return nil, fmt.Errorf("token length is zero")
	}

	token, err := jwt.Parse(uToken, h.ProvideKey)

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	payload, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		err := fmt.Errorf("invalid payload")
		klog.Errorln(err)
		return nil, err
	}

	username, ok := payload["username"].(string)

	if !ok {
		err := fmt.Errorf("invalid payload")
		klog.Errorln(err)
		return nil, err
	}

	if _, ok = payload["exp"]; ok {
		// allow static token has expiration time
		return token, nil
	}

	tokenKey := fmt.Sprintf("kubesphere:users:%s:token:%s", username, uToken)

	exist, err := h.Rule.RedisClient.Redis().Exists(tokenKey).Result()
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	if exist == 1 {
		// reset expiration time if token exist
		h.Rule.RedisClient.Redis().Expire(tokenKey, h.Rule.TokenIdleTimeout)
		return token, nil
	} else {
		return nil, errors.New("illegal token")
	}
}

func (h Auth) HandleUnauthorized(w http.ResponseWriter, err error) int {
	message := fmt.Sprintf("Unauthorized,%v", err)
	w.Header().Add("WWW-Authenticate", message)
	log.Println(message)
	return http.StatusUnauthorized
}

func (h Auth) ExtractToken(r *http.Request) (string, error) {

	jwtHeader := strings.Split(r.Header.Get("Authorization"), " ")

	if jwtHeader[0] == "Bearer" && len(jwtHeader) == 2 {
		return jwtHeader[1], nil
	}

	jwtCookie, err := r.Cookie("token")

	if err == nil {
		return jwtCookie.Value, nil
	}

	jwtQuery := r.URL.Query().Get("token")

	if jwtQuery != "" {
		return jwtQuery, nil
	}

	return "", fmt.Errorf("no token found")
}

func (h Auth) ProvideKey(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
		return h.Rule.Secret, nil
	} else {
		return nil, fmt.Errorf("expect token signed with HMAC but got %v", token.Header["alg"])
	}
}
