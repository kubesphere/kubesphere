package jenkins

import (
	"encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"k8s.io/klog"
	authtoken "kubesphere.io/kubesphere/pkg/apiserver/authentication/token"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TODO: deprecated, use SendJenkinsRequestWithHeaderResp() instead
func (j *Jenkins) SendPureRequest(path string, httpParameters *devops.HttpParameters) ([]byte, error) {
	resBody, _, err := j.SendPureRequestWithHeaderResp(path, httpParameters)

	return resBody, err
}

// provider request header to call jenkins api.
// transfer bearer token to basic token for inner Oauth and Jeknins
func (j *Jenkins) SendPureRequestWithHeaderResp(path string, httpParameters *devops.HttpParameters) ([]byte, http.Header, error) {
	Url, err := url.Parse(j.Server + path)
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}

	header := httpParameters.Header
	bearTokenArray := strings.Split(header.Get("Authorization"), " ")
	bearFlag := bearTokenArray[0]
	if strings.ToLower(bearFlag) == "bearer" {
		bearToken := bearTokenArray[1]
		if err != nil {
			klog.Error(err)
			return nil, nil, err
		}
		claim := authtoken.Claims{}
		parser := jwt.Parser{}
		_, _, err = parser.ParseUnverified(bearToken, &claim)
		if err != nil {
			return nil, nil, err
		}
		creds := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", claim.Username, bearToken)))
		header.Set("Authorization", fmt.Sprintf("Basic %s", creds))
	}

	newRequest := &http.Request{
		Method:   httpParameters.Method,
		URL:      Url,
		Header:   header,
		Body:     httpParameters.Body,
		Form:     httpParameters.Form,
		PostForm: httpParameters.PostForm,
	}

	resp, err := client.Do(newRequest)
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	resBody, _ := getRespBody(resp)
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		klog.Errorf("%+v", string(resBody))
		jkerr := new(JkError)
		jkerr.Code = resp.StatusCode
		jkerr.Message = string(resBody)
		return nil, nil, jkerr
	}

	return resBody, resp.Header, nil
}
