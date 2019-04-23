// Manage project existence.
package sonargo

import "net/http"

type ProjectsService struct {
	client *Client
}

const (
	ProjectVisibilityPublic  = "public"
	ProjectVisibilityPrivate = "private"
)

type ProjectsBulkUpdateKeyObject struct {
	Keys []*Key `json:"keys,omitempty"`
}

type Key struct {
	Duplicate bool   `json:"duplicate,omitempty"`
	Key       string `json:"key,omitempty"`
	NewKey    string `json:"newKey,omitempty"`
}

type Project struct {
	CreationDate string `json:"creationDate,omitempty"`
	Key          string `json:"key,omitempty"`
	Name         string `json:"name,omitempty"`
	Qualifier    string `json:"qualifier,omitempty"`
	UUID         string `json:"uuid,omitempty"`
	Visibility   string `json:"visibility,omitempty"`
}

type ProjectsCreateObject struct {
	Project *Project `json:"project,omitempty"`
}

type ProjectsBulkDeleteOption struct {
	AnalyzedBefore    string `url:"analyzedBefore,omitempty"`    // Description:"Filter the projects for which last analysis is older than the given date (exclusive).<br> Either a date (server timezone) or datetime can be provided.",ExampleValue:"2017-10-19 or 2017-10-19T13:00:00+0200"
	OnProvisionedOnly string `url:"onProvisionedOnly,omitempty"` // Description:"Filter the projects that are provisioned",ExampleValue:""
	ProjectIds        string `url:"projectIds,omitempty"`        // Description:"Comma-separated list of project ids. Only the 1'000 first ids are used. Others are silently ignored.",ExampleValue:"AU-Tpxb--iU5OvuD2FLy,AU-TpxcA-iU5OvuD2FLz"
	Projects          string `url:"projects,omitempty"`          // Description:"Comma-separated list of project keys",ExampleValue:"my_project,another_project"
	Q                 string `url:"q,omitempty"`                 // Description:"Limit to: <ul><li>component names that contain the supplied string</li><li>component keys that contain the supplied string</li></ul>",ExampleValue:"sonar"
	Qualifiers        string `url:"qualifiers,omitempty"`        // Description:"Comma-separated list of component qualifiers. Filter the results with the specified qualifiers",ExampleValue:""
}

// BulkDelete Delete one or several projects.<br />Requires 'Administer System' permission.
func (s *ProjectsService) BulkDelete(opt *ProjectsBulkDeleteOption) (resp *http.Response, err error) {
	err = s.ValidateBulkDeleteOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "projects/bulk_delete", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type ProjectsBulkUpdateKeyOption struct {
	DryRun    string `url:"dryRun,omitempty"`    // Description:"Simulate bulk update. No component key is updated.",ExampleValue:""
	From      string `url:"from,omitempty"`      // Description:"String to match in components keys",ExampleValue:"_old"
	Project   string `url:"project,omitempty"`   // Description:"Project or module key",ExampleValue:"my_old_project"
	ProjectId string `url:"projectId,omitempty"` // Description:"Project or module ID",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	To        string `url:"to,omitempty"`        // Description:"String replacement in components keys",ExampleValue:"_new"
}

// BulkUpdateKey Bulk update a project or module key and all its sub-components keys. The bulk update allows to replace a part of the current key by another string on the current project and all its sub-modules.<br>It's possible to simulate the bulk update by setting the parameter 'dryRun' at true. No key is updated with a dry run.<br>Ex: to rename a project with key 'my_project' to 'my_new_project' and all its sub-components keys, call the WS with parameters:<ul>  <li>project: my_project</li>  <li>from: my_</li>  <li>to: my_new_</li></ul>Either 'projectId' or 'project' must be provided.<br> Requires one of the following permissions: <ul><li>'Administer System'</li><li>'Administer' rights on the specified project</li></ul>
func (s *ProjectsService) BulkUpdateKey(opt *ProjectsBulkUpdateKeyOption) (v *ProjectsBulkUpdateKeyObject, resp *http.Response, err error) {
	err = s.ValidateBulkUpdateKeyOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "projects/bulk_update_key", opt)
	if err != nil {
		return
	}
	v = new(ProjectsBulkUpdateKeyObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type ProjectsCreateOption struct {
	Branch     string `url:"branch,omitempty"`     // Description:"SCM Branch of the project. The key of the project will become key:branch, for instance 'SonarQube:branch-5.0'",ExampleValue:"branch-5.0"
	Name       string `url:"name,omitempty"`       // Description:"Name of the project",ExampleValue:"SonarQube"
	Project    string `url:"project,omitempty"`    // Description:"Key of the project",ExampleValue:"my_project"
	Visibility string `url:"visibility,omitempty"` // Description:"Whether the created project should be visible to everyone, or only specific user/groups.<br/>If no visibility is specified, the default project visibility of the organization will be used.",ExampleValue:""
}

// Create Create a project.<br/>Requires 'Create Projects' permission
func (s *ProjectsService) Create(opt *ProjectsCreateOption) (v *ProjectsCreateObject, resp *http.Response, err error) {
	err = s.ValidateCreateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "projects/create", opt)
	if err != nil {
		return
	}
	v = new(ProjectsCreateObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type ProjectsDeleteOption struct {
	Project   string `url:"project,omitempty"`   // Description:"Project key",ExampleValue:"my_project"
	ProjectId string `url:"projectId,omitempty"` // Description:"Project ID",ExampleValue:"ce4c03d6-430f-40a9-b777-ad877c00aa4d"
}

// Delete Delete a project.<br> Requires 'Administer System' permission or 'Administer' permission on the project.
func (s *ProjectsService) Delete(opt *ProjectsDeleteOption) (resp *http.Response, err error) {
	err = s.ValidateDeleteOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "projects/delete", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type ProjectsSearchOption struct {
	AnalyzedBefore    string `url:"analyzedBefore,omitempty"`    // Description:"Filter the projects for which last analysis is older than the given date (exclusive).<br> Either a date (server timezone) or datetime can be provided.",ExampleValue:"2017-10-19 or 2017-10-19T13:00:00+0200"
	OnProvisionedOnly string `url:"onProvisionedOnly,omitempty"` // Description:"Filter the projects that are provisioned",ExampleValue:""
	P                 string `url:"p,omitempty"`                 // Description:"1-based page number",ExampleValue:"42"
	ProjectIds        string `url:"projectIds,omitempty"`        // Description:"Comma-separated list of project ids",ExampleValue:"AU-Tpxb--iU5OvuD2FLy,AU-TpxcA-iU5OvuD2FLz"
	Projects          string `url:"projects,omitempty"`          // Description:"Comma-separated list of project keys",ExampleValue:"my_project,another_project"
	Ps                string `url:"ps,omitempty"`                // Description:"Page size. Must be greater than 0 and less or equal than 500",ExampleValue:"20"
	Q                 string `url:"q,omitempty"`                 // Description:"Limit search to: <ul><li>component names that contain the supplied string</li><li>component keys that contain the supplied string</li></ul>",ExampleValue:"sonar"
	Qualifiers        string `url:"qualifiers,omitempty"`        // Description:"Comma-separated list of component qualifiers. Filter the results with the specified qualifiers",ExampleValue:""
}
type ProjectSearchObject ComponentsSearchObject

// Search Search for projects or views to administrate them.<br>Requires 'System Administrator' permission
func (s *ProjectsService) Search(opt *ProjectsSearchOption) (v *ProjectSearchObject, resp *http.Response, err error) {
	err = s.ValidateSearchOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "projects/search", opt)
	if err != nil {
		return
	}
	v = new(ProjectSearchObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type ProjectsUpdateKeyOption struct {
	From      string `url:"from,omitempty"`      // Description:"Project or module key",ExampleValue:"my_old_project"
	ProjectId string `url:"projectId,omitempty"` // Description:"Project or module id",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	To        string `url:"to,omitempty"`        // Description:"New component key",ExampleValue:"my_new_project"
}

// UpdateKey Update a project or module key and all its sub-components keys.<br>Either 'from' or 'projectId' must be provided.<br> Requires one of the following permissions: <ul><li>'Administer System'</li><li>'Administer' rights on the specified project</li></ul>
func (s *ProjectsService) UpdateKey(opt *ProjectsUpdateKeyOption) (resp *http.Response, err error) {
	err = s.ValidateUpdateKeyOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "projects/update_key", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type ProjectsUpdateVisibilityOption struct {
	Project    string `url:"project,omitempty"`    // Description:"Project key",ExampleValue:"my_project"
	Visibility string `url:"visibility,omitempty"` // Description:"New visibility",ExampleValue:""
}

// UpdateVisibility Updates visibility of a project.<br>Requires 'Project administer' permission on the specified project
func (s *ProjectsService) UpdateVisibility(opt *ProjectsUpdateVisibilityOption) (resp *http.Response, err error) {
	err = s.ValidateUpdateVisibilityOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "projects/update_visibility", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}
