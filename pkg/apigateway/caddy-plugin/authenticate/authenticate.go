package authenticate

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

type Auth struct {
	Rule Rule
	Next httpserver.Handler
}

type Rule struct {
	Secret       []byte
	Path         string
	ExceptedPath []string
}

type User struct {
	Username string                  `json:"username"`
	UID      string                  `json:"uid"`
	Groups   *[]string               `json:"groups,omitempty"`
	Extra    *map[string]interface{} `json:"extra,omitempty"`
}

func (h Auth) ServeHTTP(resp http.ResponseWriter, req *http.Request) (int, error) {
	for _, path := range h.Rule.ExceptedPath {
		if httpserver.Path(req.URL.Path).Matches(path) {
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

	payLoad, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return nil, errors.New("invalid payload")
	}

	for header := range req.Header {
		if strings.HasPrefix(header, "X-Token-") {
			req.Header.Del(header)
		}
	}

	username, ok := payLoad["username"].(string)

	if ok && username != "" {
		req.Header.Set("X-Token-Username", username)
	}

	uid := payLoad["uid"]

	if uid != nil {
		switch uid.(type) {
		case int:
			req.Header.Set("X-Token-UID", strconv.Itoa(uid.(int)))
			break
		case string:
			req.Header.Set("X-Token-UID", uid.(string))
			break
		}
	}

	groups, ok := payLoad["groups"].([]string)

	if ok && len(groups) > 0 {
		req.Header.Set("X-Token-Groups", strings.Join(groups, ","))
	}

	return req, nil
}

func (h Auth) Validate(uToken string) (*jwt.Token, error) {

	if len(uToken) == 0 {
		return nil, fmt.Errorf("token length is zero")
	}

	token, err := jwt.Parse(uToken, h.ProvideKey)

	if err != nil {
		return nil, err
	}

	return token, nil
}

func (h Auth) HandleUnauthorized(w http.ResponseWriter, err error) int {
	message := fmt.Sprintf("Unauthorized,%v", err)
	w.Header().Add("WWW-Authenticate", message)
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
