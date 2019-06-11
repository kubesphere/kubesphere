// Manage pull request (only available when the Branch plugin is installed)
package sonargo

import "net/http"

type ProjectPullRequestsService struct {
	client *Client
}

type ProjectPullRequestsListObject struct {
	PullRequests []*PullRequest `json:"pullRequests,omitempty"`
}

type PullRequest struct {
	AnalysisDate string `json:"analysisDate,omitempty"`
	Base         string `json:"base,omitempty"`
	Branch       string `json:"branch,omitempty"`
	Key          string `json:"key,omitempty"`
	Status       Status `json:"status,omitempty"`
	Title        string `json:"title,omitempty"`
	URL          string `json:"url,omitempty"`
}

type ProjectPullRequestsDeleteOption struct {
	Project     string `url:"project,omitempty"`     // Description:"Project key",ExampleValue:"my_project"
	PullRequest int    `url:"pullRequest,omitempty"` // Description:"Pull request id",ExampleValue:"1543"
}

// Delete Delete a pull request.<br/>Requires 'Administer' rights on the specified project.
func (s *ProjectPullRequestsService) Delete(opt *ProjectPullRequestsDeleteOption) (resp *http.Response, err error) {
	err = s.ValidateDeleteOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "project_pull_requests/delete", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type ProjectPullRequestsListOption struct {
	Project string `url:"project,omitempty"` // Description:"Project key",ExampleValue:"my_project"
}

// List List the pull requests of a project.<br/>One of the following permissions is required: <ul><li>'Browse' rights on the specified project</li><li>'Execute Analysis' rights on the specified project</li></ul>
func (s *ProjectPullRequestsService) List(opt *ProjectPullRequestsListOption) (v *ProjectPullRequestsListObject, resp *http.Response, err error) {
	err = s.ValidateListOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "project_pull_requests/list", opt)
	if err != nil {
		return
	}
	v = new(ProjectPullRequestsListObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}
