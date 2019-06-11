//
package sonargo

import "net/http"

type ServerService struct {
	client *Client
}

// Version Version of SonarQube in plain text
func (s *ServerService) Version() (v *string, resp *http.Response, err error) {
	req, err := s.client.NewRequest("GET", "server/version", nil)
	if err != nil {
		return
	}
	v = new(string)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}
