// Manage project tags
package sonargo

import "net/http"

type ProjectTagsService struct {
	client *Client
}

type ProjectTagsSearchOption struct {
	Ps int    `url:"ps,omitempty"` // Description:"Page size. Must be greater than 0 and less or equal than 100",ExampleValue:"20"
	Q  string `url:"q,omitempty"`  // Description:"Limit search to tags that contain the supplied string.",ExampleValue:"off"
}

// Search Search tags
func (s *ProjectTagsService) Search(opt *ProjectTagsSearchOption) (v *IssuesTagsObject, resp *http.Response, err error) {
	err = s.ValidateSearchOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "project_tags/search", opt)
	if err != nil {
		return
	}
	v = new(IssuesTagsObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type ProjectTagsSetOption struct {
	Project string  `url:"project,omitempty"` // Description:"Project key",ExampleValue:"my_project"
	Tags    *string `url:"tags,omitempty"`    // Description:"Comma-separated list of tags",ExampleValue:"finance, offshore"
}

// Set Set tags on a project.<br>Requires the following permission: 'Administer' rights on the specified project
func (s *ProjectTagsService) Set(opt *ProjectTagsSetOption) (resp *http.Response, err error) {
	err = s.ValidateSetOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "project_tags/set", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}
