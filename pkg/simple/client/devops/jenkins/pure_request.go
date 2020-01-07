package jenkins

import (
	"k8s.io/klog"
	"net/http"
	"net/url"
	"time"
)

// TODO: deprecated, use SendJenkinsRequestWithHeaderResp() instead
func (j *Jenkins) SendJenkinsRequest(reqUrl string, req *http.Request) ([]byte,error){
	resBody, _, err := j.SendJenkinsRequestWithHeaderResp(reqUrl, req)
	return resBody, err
}

func (j *Jenkins) SendJenkinsRequestWithHeaderResp(reqUrl string, req *http.Request) ([]byte, http.Header, error) {
	newReqUrl, err := url.Parse(j.Server + reqUrl)
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}

	newRequest := &http.Request{
		Method:   req.Method,
		URL:      newReqUrl,
		Header:   req.Header,
		Body:     req.Body,
		Form:     req.Form,
		PostForm: req.PostForm,
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