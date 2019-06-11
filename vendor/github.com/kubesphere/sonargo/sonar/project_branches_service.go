// Manage branch (only available when the Branch plugin is installed)
package sonargo

import "net/http"

type ProjectBranchesService struct {
	client *Client
}

type ProjectBranchesListObject struct {
	Branches []*Branch `json:"branches,omitempty"`
}

type Branch struct {
	AnalysisDate string  `json:"analysisDate,omitempty"`
	IsMain       bool    `json:"isMain,omitempty"`
	MergeBranch  string  `json:"mergeBranch,omitempty"`
	Name         string  `json:"name,omitempty"`
	Status       *Status `json:"status,omitempty"`
	Type         string  `json:"type,omitempty"`
}

type Status struct {
	Bugs              int64  `json:"bugs,omitempty"`
	CodeSmells        int64  `json:"codeSmells,omitempty"`
	QualityGateStatus string `json:"qualityGateStatus,omitempty"`
	Vulnerabilities   int64  `json:"vulnerabilities,omitempty"`
}

type ProjectBranchesDeleteOption struct {
	Branch  string `url:"branch,omitempty"`  // Description:"Name of the branch",ExampleValue:"branch1"
	Project string `url:"project,omitempty"` // Description:"Project key",ExampleValue:"my_project"
}

// Delete Delete a non-main branch of a project.<br/>Requires 'Administer' rights on the specified project.
func (s *ProjectBranchesService) Delete(opt *ProjectBranchesDeleteOption) (resp *http.Response, err error) {
	err = s.ValidateDeleteOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "project_branches/delete", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type ProjectBranchesListOption struct {
	Project string `url:"project,omitempty"` // Description:"Project key",ExampleValue:"my_project"
}

// List List the branches of a project.<br/>Requires 'Browse' or 'Execute analysis' rights on the specified project.
func (s *ProjectBranchesService) List(opt *ProjectBranchesListOption) (v *ProjectBranchesListObject, resp *http.Response, err error) {
	err = s.ValidateListOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "project_branches/list", opt)
	if err != nil {
		return
	}
	v = new(ProjectBranchesListObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type ProjectBranchesRenameOption struct {
	Name    string `url:"name,omitempty"`    // Description:"New name of the main branch",ExampleValue:"branch1"
	Project string `url:"project,omitempty"` // Description:"Project key",ExampleValue:"my_project"
}

// Rename Rename the main branch of a project.<br/>Requires 'Administer' permission on the specified project.
func (s *ProjectBranchesService) Rename(opt *ProjectBranchesRenameOption) (resp *http.Response, err error) {
	err = s.ValidateRenameOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "project_branches/rename", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}
