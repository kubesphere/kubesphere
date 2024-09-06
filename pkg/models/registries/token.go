/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package registries

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (t authToken) String() (string, error) {
	if t.Token != "" {
		return t.Token, nil
	}
	if t.AccessToken != "" {
		return t.AccessToken, nil
	}
	return "", errors.New("auth token cannot be empty")
}

func (a *authService) Request(username, password string) (*http.Request, error) {
	q := a.Realm.Query()
	q.Set("service", a.Service)
	for _, s := range a.Scope {
		q.Set("scope", s)
	}
	//	q.Set("scope", "repository:r.j3ss.co/htop:push,pull")
	a.Realm.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", a.Realm.String(), nil)

	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}

	return req, err
}

func isTokenDemand(resp *http.Response) (*authService, error) {
	if resp == nil {
		return nil, nil
	}
	if resp.StatusCode != http.StatusUnauthorized {
		return nil, nil
	}
	return parseAuthHeader(resp.Header)
}

// Token returns the required token for the specific resource url. If the registry requires basic authentication, this
// function returns ErrBasicAuth.
func (r *Registry) Token(url string) (str string, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden && gcrMatcher.MatchString(url) {
		// GCR is not sending HTTP 401 on missing credentials but a HTTP 403 without
		// any further information about why the request failed. Sending the credentials
		// from the Docker config fixes this.
		return "", ErrBasicAuth
	}

	authService, err := isTokenDemand(resp)
	if err != nil {
		return "", err
	}

	if authService == nil {
		return "", nil
	}

	authReq, err := authService.Request(r.Username, r.Password)
	if err != nil {
		return "", err
	}

	resp, err = r.Client.Do(authReq)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("getting image failed with secret")
	}

	var authToken authToken
	if err := json.NewDecoder(resp.Body).Decode(&authToken); err != nil {
		return "", err
	}

	token, err := authToken.String()
	return token, err
}

func parseAuthHeader(header http.Header) (*authService, error) {
	ch, err := parseChallenge(header.Get("www-authenticate"))
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func parseChallenge(challengeHeader string) (*authService, error) {
	if basicRegex.MatchString(challengeHeader) {
		return nil, ErrBasicAuth
	}

	match := bearerRegex.FindAllStringSubmatch(challengeHeader, -1)
	if d := len(match); d != 1 {
		return nil, fmt.Errorf("malformed auth challenge header: '%s', %d", challengeHeader, d)
	}
	parts := strings.SplitN(strings.TrimSpace(match[0][1]), ",", 3)

	var realm, service string
	var scope []string
	for _, s := range parts {
		p := strings.SplitN(s, "=", 2)
		if len(p) != 2 {
			return nil, fmt.Errorf("malformed auth challenge header: '%s'", challengeHeader)
		}
		key := p[0]
		value := strings.TrimSuffix(strings.TrimPrefix(p[1], `"`), `"`)
		switch key {
		case "realm":
			realm = value
		case "service":
			service = value
		case "scope":
			scope = strings.Fields(value)
		default:
			return nil, fmt.Errorf("unknown field in challenge header %s: %v", key, challengeHeader)
		}
	}
	parsedRealm, err := url.Parse(realm)
	if err != nil {
		return nil, err
	}

	a := &authService{
		Realm:   parsedRealm,
		Service: service,
		Scope:   scope,
	}

	return a, nil
}
