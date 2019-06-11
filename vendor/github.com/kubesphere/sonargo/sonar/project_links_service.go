// Manage projects links.
package sonargo

import "net/http"

type ProjectLinksService struct {
	client *Client
}

type ProjectLinksCreateObject struct {
	Link *Link `json:"link,omitempty"`
}

type Link struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
	Type string `json:"type,omitempty"`
}

type ProjectLinksSearchObject struct {
	Links []*Link `json:"links,omitempty"`
}

type ProjectLinksCreateOption struct {
	Name       string `url:"name,omitempty"`       // Description:"Link name",ExampleValue:"Custom"
	ProjectId  string `url:"projectId,omitempty"`  // Description:"Project id",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	ProjectKey string `url:"projectKey,omitempty"` // Description:"Project key",ExampleValue:"my_project"
	Url        string `url:"url,omitempty"`        // Description:"Link url",ExampleValue:"http://example.com"
}

// Create Create a new project link.<br>Requires 'Administer' permission on the specified project, or global 'Administer' permission.
func (s *ProjectLinksService) Create(opt *ProjectLinksCreateOption) (v *ProjectLinksCreateObject, resp *http.Response, err error) {
	err = s.ValidateCreateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "project_links/create", opt)
	if err != nil {
		return
	}
	v = new(ProjectLinksCreateObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type ProjectLinksDeleteOption struct {
	Id string `url:"id,omitempty"` // Description:"Link id",ExampleValue:"17"
}

// Delete Delete existing project link.<br>Requires 'Administer' permission on the specified project, or global 'Administer' permission.
func (s *ProjectLinksService) Delete(opt *ProjectLinksDeleteOption) (resp *http.Response, err error) {
	err = s.ValidateDeleteOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "project_links/delete", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type ProjectLinksSearchOption struct {
	ProjectId  string `url:"projectId,omitempty"`  // Description:"Project Id",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	ProjectKey string `url:"projectKey,omitempty"` // Description:"Project Key",ExampleValue:"my_project"
}

// Search List links of a project.<br>The 'projectId' or 'projectKey' must be provided.<br>Requires one of the following permissions:<ul><li>'Administer System'</li><li>'Administer' rights on the specified project</li><li>'Browse' on the specified project</li></ul>
func (s *ProjectLinksService) Search(opt *ProjectLinksSearchOption) (v *ProjectLinksSearchObject, resp *http.Response, err error) {
	err = s.ValidateSearchOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "project_links/search", opt)
	if err != nil {
		return
	}
	v = new(ProjectLinksSearchObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}
