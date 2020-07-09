package jenkins

import (
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
	"net/url"
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
	SetBasicBearTokenHeader(&header)

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
