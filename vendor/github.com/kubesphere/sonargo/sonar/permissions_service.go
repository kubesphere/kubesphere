// Manage permission templates, and the granting and revoking of permissions at the global and project levels.
package sonargo

import "net/http"

type PermissionsService struct {
	client *Client
}

type PermissionsCreateTemplateObject struct {
	PermissionTemplate *PermissionTemplate `json:"permissionTemplate,omitempty"`
}

type PermissionTemplate struct {
	ID                string        `json:"id,omitempty"`
	CreatedAt         string        `json:"createdAt,omitempty"`
	UpdatedAt         string        `json:"updatedAt,omitempty"`
	Description       string        `json:"description,omitempty"`
	Name              string        `json:"name,omitempty"`
	ProjectKeyPattern string        `json:"projectKeyPattern,omitempty"`
	Permissions       []*Permission `json:"permissions,omitempty"`
}

type Permission struct {
	Description        string `json:"description,omitempty"`
	GroupsCount        int64  `json:"groupsCount,omitempty"`
	Key                string `json:"key,omitempty"`
	Name               string `json:"name,omitempty"`
	UsersCount         int64  `json:"usersCount,omitempty"`
	WithProjectCreator bool   `json:"withProjectCreator,omitempty"`
}

type PermissionsSearchTemplatesObject struct {
	DefaultTemplates    []*DefaultTemplate    `json:"defaultTemplates,omitempty"`
	PermissionTemplates []*PermissionTemplate `json:"permissionTemplates,omitempty"`
	Permissions         []*Permission         `json:"permissions,omitempty"`
}

type DefaultTemplate struct {
	Qualifier  string `json:"qualifier,omitempty"`
	TemplateID string `json:"templateId,omitempty"`
}

type PermissionsAddGroupOption struct {
	GroupId    int    `url:"groupId,omitempty"`    // Description:"Group id",ExampleValue:"42"
	GroupName  string `url:"groupName,omitempty"`  // Description:"Group name or 'anyone' (case insensitive)",ExampleValue:"sonar-administrators"
	Permission string `url:"permission,omitempty"` // Description:"Permission<ul><li>Possible values for global permissions: admin, profileadmin, gateadmin, scan, provisioning</li><li>Possible values for project permissions admin, codeviewer, issueadmin, scan, user</li></ul>",ExampleValue:""
	ProjectId  string `url:"projectId,omitempty"`  // Description:"Project id",ExampleValue:"ce4c03d6-430f-40a9-b777-ad877c00aa4d"
	ProjectKey string `url:"projectKey,omitempty"` // Description:"Project key",ExampleValue:"my_project"
}

// AddGroup Add permission to a group.<br /> This service defaults to global permissions, but can be limited to project permissions by providing project id or project key.<br /> The group name or group id must be provided. <br />Requires one of the following permissions:<ul><li>'Administer System'</li><li>'Administer' rights on the specified project</li></ul>
func (s *PermissionsService) AddGroup(opt *PermissionsAddGroupOption) (resp *http.Response, err error) {
	err = s.ValidateAddGroupOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "permissions/add_group", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type PermissionsAddGroupToTemplateOption struct {
	GroupId      int    `url:"groupId,omitempty"`      // Description:"Group id",ExampleValue:"42"
	GroupName    string `url:"groupName,omitempty"`    // Description:"Group name or 'anyone' (case insensitive)",ExampleValue:"sonar-administrators"
	Permission   string `url:"permission,omitempty"`   // Description:"Permission<ul><li>Possible values for project permissions admin, codeviewer, issueadmin, scan, user</li></ul>",ExampleValue:""
	TemplateId   string `url:"templateId,omitempty"`   // Description:"Template id",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	TemplateName string `url:"templateName,omitempty"` // Description:"Template name",ExampleValue:"Default Permission Template for Projects"
}

// AddGroupToTemplate Add a group to a permission template.<br /> The group id or group name must be provided. <br />Requires the following permission: 'Administer System'.
func (s *PermissionsService) AddGroupToTemplate(opt *PermissionsAddGroupToTemplateOption) (resp *http.Response, err error) {
	err = s.ValidateAddGroupToTemplateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "permissions/add_group_to_template", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type PermissionsAddProjectCreatorToTemplateOption struct {
	Permission   string `url:"permission,omitempty"`   // Description:"Permission<ul><li>Possible values for project permissions admin, codeviewer, issueadmin, scan, user</li></ul>",ExampleValue:""
	TemplateId   string `url:"templateId,omitempty"`   // Description:"Template id",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	TemplateName string `url:"templateName,omitempty"` // Description:"Template name",ExampleValue:"Default Permission Template for Projects"
}

// AddProjectCreatorToTemplate Add a project creator to a permission template.<br>Requires the following permission: 'Administer System'.
func (s *PermissionsService) AddProjectCreatorToTemplate(opt *PermissionsAddProjectCreatorToTemplateOption) (resp *http.Response, err error) {
	err = s.ValidateAddProjectCreatorToTemplateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "permissions/add_project_creator_to_template", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type PermissionsAddUserOption struct {
	Login      string `url:"login,omitempty"`      // Description:"User login",ExampleValue:"g.hopper"
	Permission string `url:"permission,omitempty"` // Description:"Permission<ul><li>Possible values for global permissions: admin, profileadmin, gateadmin, scan, provisioning</li><li>Possible values for project permissions admin, codeviewer, issueadmin, scan, user</li></ul>",ExampleValue:""
	ProjectId  string `url:"projectId,omitempty"`  // Description:"Project id",ExampleValue:"ce4c03d6-430f-40a9-b777-ad877c00aa4d"
	ProjectKey string `url:"projectKey,omitempty"` // Description:"Project key",ExampleValue:"my_project"
}

// AddUser Add permission to a user.<br /> This service defaults to global permissions, but can be limited to project permissions by providing project id or project key.<br />Requires one of the following permissions:<ul><li>'Administer System'</li><li>'Administer' rights on the specified project</li></ul>
func (s *PermissionsService) AddUser(opt *PermissionsAddUserOption) (resp *http.Response, err error) {
	err = s.ValidateAddUserOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "permissions/add_user", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type PermissionsAddUserToTemplateOption struct {
	Login        string `url:"login,omitempty"`        // Description:"User login",ExampleValue:"g.hopper"
	Permission   string `url:"permission,omitempty"`   // Description:"Permission<ul><li>Possible values for project permissions admin, codeviewer, issueadmin, scan, user</li></ul>",ExampleValue:""
	TemplateId   string `url:"templateId,omitempty"`   // Description:"Template id",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	TemplateName string `url:"templateName,omitempty"` // Description:"Template name",ExampleValue:"Default Permission Template for Projects"
}

// AddUserToTemplate Add a user to a permission template.<br /> Requires the following permission: 'Administer System'.
func (s *PermissionsService) AddUserToTemplate(opt *PermissionsAddUserToTemplateOption) (resp *http.Response, err error) {
	err = s.ValidateAddUserToTemplateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "permissions/add_user_to_template", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type PermissionsApplyTemplateOption struct {
	ProjectId    string `url:"projectId,omitempty"`    // Description:"Project id",ExampleValue:"ce4c03d6-430f-40a9-b777-ad877c00aa4d"
	ProjectKey   string `url:"projectKey,omitempty"`   // Description:"Project key",ExampleValue:"my_project"
	TemplateId   string `url:"templateId,omitempty"`   // Description:"Template id",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	TemplateName string `url:"templateName,omitempty"` // Description:"Template name",ExampleValue:"Default Permission Template for Projects"
}

// ApplyTemplate Apply a permission template to one project.<br>The project id or project key must be provided.<br>The template id or name must be provided.<br>Requires the following permission: 'Administer System'.
func (s *PermissionsService) ApplyTemplate(opt *PermissionsApplyTemplateOption) (resp *http.Response, err error) {
	err = s.ValidateApplyTemplateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "permissions/apply_template", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type PermissionsBulkApplyTemplateOption struct {
	AnalyzedBefore    string `url:"analyzedBefore,omitempty"`    // Description:"Filter the projects for which last analysis is older than the given date (exclusive).<br> Either a date (server timezone) or datetime can be provided.",ExampleValue:"2017-10-19 or 2017-10-19T13:00:00+0200"
	OnProvisionedOnly string `url:"onProvisionedOnly,omitempty"` // Description:"Filter the projects that are provisioned",ExampleValue:""
	Projects          string `url:"projects,omitempty"`          // Description:"Comma-separated list of project keys",ExampleValue:"my_project,another_project"
	Q                 string `url:"q,omitempty"`                 // Description:"Limit search to: <ul><li>project names that contain the supplied string</li><li>project keys that are exactly the same as the supplied string</li></ul>",ExampleValue:"apac"
	Qualifiers        string `url:"qualifiers,omitempty"`        // Description:"Comma-separated list of component qualifiers. Filter the results with the specified qualifiers. Possible values are:<ul><li>TRK - Projects</li></ul>",ExampleValue:""
	TemplateId        string `url:"templateId,omitempty"`        // Description:"Template id",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	TemplateName      string `url:"templateName,omitempty"`      // Description:"Template name",ExampleValue:"Default Permission Template for Projects"
}

// BulkApplyTemplate Apply a permission template to several projects.<br />The template id or name must be provided.<br />Requires the following permission: 'Administer System'.
func (s *PermissionsService) BulkApplyTemplate(opt *PermissionsBulkApplyTemplateOption) (resp *http.Response, err error) {
	err = s.ValidateBulkApplyTemplateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "permissions/bulk_apply_template", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type PermissionsCreateTemplateOption struct {
	Description       string `url:"description,omitempty"`       // Description:"Description",ExampleValue:"Permissions for all projects related to the financial service"
	Name              string `url:"name,omitempty"`              // Description:"Name",ExampleValue:"Financial Service Permissions"
	ProjectKeyPattern string `url:"projectKeyPattern,omitempty"` // Description:"Project key pattern. Must be a valid Java regular expression",ExampleValue:".*\.finance\..*"
}

// CreateTemplate Create a permission template.<br />Requires the following permission: 'Administer System'.
func (s *PermissionsService) CreateTemplate(opt *PermissionsCreateTemplateOption) (v *PermissionsCreateTemplateObject, resp *http.Response, err error) {
	err = s.ValidateCreateTemplateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "permissions/create_template", opt)
	if err != nil {
		return
	}
	v = new(PermissionsCreateTemplateObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type PermissionsDeleteTemplateOption struct {
	TemplateId   string `url:"templateId,omitempty"`   // Description:"Template id",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	TemplateName string `url:"templateName,omitempty"` // Description:"Template name",ExampleValue:"Default Permission Template for Projects"
}

// DeleteTemplate Delete a permission template.<br />Requires the following permission: 'Administer System'.
func (s *PermissionsService) DeleteTemplate(opt *PermissionsDeleteTemplateOption) (resp *http.Response, err error) {
	err = s.ValidateDeleteTemplateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "permissions/delete_template", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type PermissionsRemoveGroupOption PermissionsAddGroupOption

// RemoveGroup Remove a permission from a group.<br /> This service defaults to global permissions, but can be limited to project permissions by providing project id or project key.<br /> The group id or group name must be provided, not both.<br />Requires one of the following permissions:<ul><li>'Administer System'</li><li>'Administer' rights on the specified project</li></ul>
func (s *PermissionsService) RemoveGroup(opt *PermissionsRemoveGroupOption) (resp *http.Response, err error) {
	err = s.ValidateRemoveGroupOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "permissions/remove_group", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type PermissionsRemoveGroupFromTemplateOption PermissionsAddGroupToTemplateOption

// RemoveGroupFromTemplate Remove a group from a permission template.<br /> The group id or group name must be provided. <br />Requires the following permission: 'Administer System'.
func (s *PermissionsService) RemoveGroupFromTemplate(opt *PermissionsRemoveGroupFromTemplateOption) (resp *http.Response, err error) {
	err = s.ValidateRemoveGroupFromTemplateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "permissions/remove_group_from_template", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type PermissionsRemoveProjectCreatorFromTemplateOption PermissionsAddProjectCreatorToTemplateOption

// RemoveProjectCreatorFromTemplate Remove a project creator from a permission template.<br>Requires the following permission: 'Administer System'.
func (s *PermissionsService) RemoveProjectCreatorFromTemplate(opt *PermissionsRemoveProjectCreatorFromTemplateOption) (resp *http.Response, err error) {
	err = s.ValidateRemoveProjectCreatorFromTemplateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "permissions/remove_project_creator_from_template", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type PermissionsRemoveUserOption PermissionsAddUserOption

// RemoveUser Remove permission from a user.<br /> This service defaults to global permissions, but can be limited to project permissions by providing project id or project key.<br /> Requires one of the following permissions:<ul><li>'Administer System'</li><li>'Administer' rights on the specified project</li></ul>
func (s *PermissionsService) RemoveUser(opt *PermissionsRemoveUserOption) (resp *http.Response, err error) {
	err = s.ValidateRemoveUserOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "permissions/remove_user", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type PermissionsRemoveUserFromTemplateOption PermissionsAddUserToTemplateOption

// RemoveUserFromTemplate Remove a user from a permission template.<br /> Requires the following permission: 'Administer System'.
func (s *PermissionsService) RemoveUserFromTemplate(opt *PermissionsRemoveUserFromTemplateOption) (resp *http.Response, err error) {
	err = s.ValidateRemoveUserFromTemplateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "permissions/remove_user_from_template", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type PermissionsSearchTemplatesOption struct {
	Q string `url:"q,omitempty"` // Description:"Limit search to permission template names that contain the supplied string.",ExampleValue:"defau"
}

// SearchTemplates List permission templates.<br />Requires the following permission: 'Administer System'.
func (s *PermissionsService) SearchTemplates(opt *PermissionsSearchTemplatesOption) (v *PermissionsSearchTemplatesObject, resp *http.Response, err error) {
	err = s.ValidateSearchTemplatesOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "permissions/search_templates", opt)
	if err != nil {
		return
	}
	v = new(PermissionsSearchTemplatesObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type PermissionsSetDefaultTemplateOption struct {
	Qualifier    string `url:"qualifier,omitempty"`    // Description:"Project qualifier. Filter the results with the specified qualifier. Possible values are:<ul><li>TRK - Projects</li></ul>",ExampleValue:""
	TemplateId   string `url:"templateId,omitempty"`   // Description:"Template id",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	TemplateName string `url:"templateName,omitempty"` // Description:"Template name",ExampleValue:"Default Permission Template for Projects"
}

// SetDefaultTemplate Set a permission template as default.<br />Requires the following permission: 'Administer System'.
func (s *PermissionsService) SetDefaultTemplate(opt *PermissionsSetDefaultTemplateOption) (resp *http.Response, err error) {
	err = s.ValidateSetDefaultTemplateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "permissions/set_default_template", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type PermissionsUpdateTemplateOption struct {
	Description       string `url:"description,omitempty"`       // Description:"Description",ExampleValue:"Permissions for all projects related to the financial service"
	Id                string `url:"id,omitempty"`                // Description:"Id",ExampleValue:"af8cb8cc-1e78-4c4e-8c00-ee8e814009a5"
	Name              string `url:"name,omitempty"`              // Description:"Name",ExampleValue:"Financial Service Permissions"
	ProjectKeyPattern string `url:"projectKeyPattern,omitempty"` // Description:"Project key pattern. Must be a valid Java regular expression",ExampleValue:".*\.finance\..*"
}

// UpdateTemplate Update a permission template.<br />Requires the following permission: 'Administer System'.
func (s *PermissionsService) UpdateTemplate(opt *PermissionsUpdateTemplateOption) (v *PermissionsCreateTemplateObject, resp *http.Response, err error) {
	err = s.ValidateUpdateTemplateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "permissions/update_template", opt)
	if err != nil {
		return
	}
	v = new(PermissionsCreateTemplateObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}
