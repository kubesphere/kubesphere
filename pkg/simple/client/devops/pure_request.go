package devops

import "net/http"

func (j *Jenkins) SendJenkinsRequest(baseUrl string, req *http.Request) {

}

func (j *Jenkins) SendJenkinsRequestWithHeaderResp(baseUrl string, req *http.Request) ([]byte, http.Header, error) {

}
