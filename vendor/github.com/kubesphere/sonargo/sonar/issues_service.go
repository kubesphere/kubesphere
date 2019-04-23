// Read and update issues.
package sonargo

import "net/http"

type IssuesService struct {
	client *Client
}

const (
	SeverityINFO     = "INFO"
	SeverityMINOR    = "MINOR"
	SeverityMAJOR    = "MAJOR"
	SeverityCRITICAL = "CRITICAL"
	SeverityBLOCKER  = "BLOCKER"

	IssueTypeBug           = "BUG"
	IssueTypeCodeSmell     = "CODE_SMELL"
	IssueTypeVulnerability = "VULNERABILITY"
)

type IssuesAddCommentObject struct {
	Components []*Component `json:"components,omitempty"`
	Issue      *Issue       `json:"issue,omitempty"`
	Rules      []*Rule      `json:"rules,omitempty"`
	Users      []*User      `json:"users,omitempty"`
}

type IssuesSearchObject struct {
	Components  []*Component `json:"components,omitempty"`
	EffortTotal int          `json:"effortTotal,omitempty"`
	DebtTotal   int          `json:"debtTotal,omitempty"`
	Issues      []*Issue     `json:"issues,omitempty"`
	P           int          `json:"p,omitempty"`
	Ps          int          `json:"ps,omitempty"`
	Paging      *Paging      `json:"paging,omitempty"`
	Rules       []*Rule      `json:"rules,omitempty"`
	Total       int          `json:"total,omitempty"`
	Users       []*User      `json:"users,omitempty"`
	Facets      []*Facet     `json:"facets,omitempty"`
}

type Comment struct {
	CreatedAt string `json:"createdAt,omitempty"`
	HTMLText  string `json:"htmlText,omitempty"`
	Key       string `json:"key,omitempty"`
	Login     string `json:"login,omitempty"`
	Markdown  string `json:"markdown,omitempty"`
	Updatable bool   `json:"updatable,omitempty"`
}

type Issue struct {
	Actions      []string      `json:"actions,omitempty"`
	Assignee     string        `json:"assignee,omitempty"`
	Author       string        `json:"author,omitempty"`
	Comments     []*Comment    `json:"comments,omitempty"`
	Component    string        `json:"component,omitempty"`
	CreationDate string        `json:"creationDate,omitempty"`
	Debt         string        `json:"debt,omitempty"`
	Effort       string        `json:"effort,omitempty"`
	Flows        []interface{} `json:"flows,omitempty"`
	Hash         string        `json:"hash,omitempty"`
	Key          string        `json:"key,omitempty"`
	Line         int           `json:"line,omitempty"`
	Message      string        `json:"message,omitempty"`
	Organization string        `json:"organization,omitempty"`
	Project      string        `json:"project,omitempty"`
	Rule         string        `json:"rule,omitempty"`
	Severity     string        `json:"severity,omitempty"`
	Status       string        `json:"status,omitempty"`
	Tags         []string      `json:"tags,omitempty"`
	TextRange    *TextRange    `json:"textRange,omitempty"`
	Transitions  []string      `json:"transitions,omitempty"`
	Type         string        `json:"type,omitempty"`
	UpdateDate   string        `json:"updateDate,omitempty"`
	FromHotspot  bool          `json:"fromHotspot,omitempty"`
	Resolution   string        `json:"resolution,omitempty"`
	CloseDate    string        `json:"closeDate,omitempty"`
}

type TextRange struct {
	EndLine     int `json:"endLine,omitempty"`
	EndOffset   int `json:"endOffset,omitempty"`
	StartLine   int `json:"startLine,omitempty"`
	StartOffset int `json:"startOffset,omitempty"`
}

type IssuesAuthorsObject struct {
	Authors []string `json:"authors,omitempty"`
}

type IssuesBulkChangeObject struct {
	Failures int `json:"failures,omitempty"`
	Ignored  int `json:"ignored,omitempty"`
	Success  int `json:"success,omitempty"`
	Total    int `json:"total,omitempty"`
}

type IssuesChangelogObject struct {
	Changelog []*Changelog `json:"changelog,omitempty"`
}

type Changelog struct {
	Avatar       string  `json:"avatar,omitempty"`
	CreationDate string  `json:"creationDate,omitempty"`
	Diffs        []*Diff `json:"diffs,omitempty"`
	User         string  `json:"user,omitempty"`
	UserName     string  `json:"userName,omitempty"`
}

type Diff struct {
	Key      string `json:"key,omitempty"`
	NewValue string `json:"newValue,omitempty"`
	OldValue string `json:"oldValue,omitempty"`
}

type IssuesTagsObject struct {
	Tags []string `json:"tags,omitempty"`
}

type IssuesAddCommentOption struct {
	Issue string `url:"issue,omitempty"` // Description:"Issue key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Text  string `url:"text,omitempty"`  // Description:"Comment text",ExampleValue:"Won't fix because it doesn't apply to the context"
}

// AddComment Add a comment.<br/>Requires authentication and the following permission: 'Browse' on the project of the specified issue.
func (s *IssuesService) AddComment(opt *IssuesAddCommentOption) (v *IssuesAddCommentObject, resp *http.Response, err error) {
	err = s.ValidateAddCommentOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "issues/add_comment", opt)
	if err != nil {
		return
	}
	v = new(IssuesAddCommentObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type IssuesAssignOption struct {
	Assignee string `url:"assignee,omitempty"` // Description:"Login of the assignee. When not set, it will unassign the issue. Use '_me' to assign to current user",ExampleValue:"admin"
	Issue    string `url:"issue,omitempty"`    // Description:"Issue key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
}

// Assign Assign/Unassign an issue. Requires authentication and Browse permission on project
func (s *IssuesService) Assign(opt *IssuesAssignOption) (v *IssuesAddCommentObject, resp *http.Response, err error) {
	err = s.ValidateAssignOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "issues/assign", opt)
	if err != nil {
		return
	}
	v = new(IssuesAddCommentObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type IssuesAuthorsOption struct {
	Ps int    `url:"ps,omitempty"` // Description:"The size of the list to return",ExampleValue:"25"
	Q  string `url:"q,omitempty"`  // Description:"A pattern to match SCM accounts against",ExampleValue:"luke"
}

// Authors Search SCM accounts which match a given query
func (s *IssuesService) Authors(opt *IssuesAuthorsOption) (v *IssuesAuthorsObject, resp *http.Response, err error) {
	err = s.ValidateAuthorsOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "issues/authors", opt)
	if err != nil {
		return
	}
	v = new(IssuesAuthorsObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type IssuesBulkChangeOption struct {
	AddTags           string `url:"add_tags,omitempty"`          // Description:"Add tags",ExampleValue:"security,java8"
	Assign            string `url:"assign,omitempty"`            // Description:"To assign the list of issues to a specific user (login), or un-assign all the issues",ExampleValue:"john.smith"
	Comment           string `url:"comment,omitempty"`           // Description:"To add a comment to a list of issues",ExampleValue:"Here is my comment"
	DoTransition      string `url:"do_transition,omitempty"`     // Description:"Transition",ExampleValue:"reopen"
	Issues            string `url:"issues,omitempty"`            // Description:"Comma-separated list of issue keys",ExampleValue:"AU-Tpxb--iU5OvuD2FLy,AU-TpxcA-iU5OvuD2FLz"
	Plan              string `url:"plan,omitempty"`              // Description:"In 5.5, action plans are dropped. Has no effect. To plan the list of issues to a specific action plan (key), or unlink all the issues from an action plan",ExampleValue:""
	RemoveTags        string `url:"remove_tags,omitempty"`       // Description:"Remove tags",ExampleValue:"security,java8"
	SendNotifications string `url:"sendNotifications,omitempty"` // Description:"",ExampleValue:""
	SetSeverity       string `url:"set_severity,omitempty"`      // Description:"To change the severity of the list of issues",ExampleValue:"BLOCKER"
	SetType           string `url:"set_type,omitempty"`          // Description:"To change the type of the list of issues",ExampleValue:"BUG"
}

// BulkChange Bulk change on issues.<br/>Requires authentication.
func (s *IssuesService) BulkChange(opt *IssuesBulkChangeOption) (v *IssuesBulkChangeObject, resp *http.Response, err error) {
	err = s.ValidateBulkChangeOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "issues/bulk_change", opt)
	if err != nil {
		return
	}
	v = new(IssuesBulkChangeObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type IssuesChangelogOption struct {
	Issue string `url:"issue,omitempty"` // Description:"Issue key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
}

// Changelog Display changelog of an issue.<br/>Requires the 'Browse' permission on the project of the specified issue.
func (s *IssuesService) Changelog(opt *IssuesChangelogOption) (v *IssuesChangelogObject, resp *http.Response, err error) {
	err = s.ValidateChangelogOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "issues/changelog", opt)
	if err != nil {
		return
	}
	v = new(IssuesChangelogObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type IssuesDeleteCommentOption struct {
	Comment string `url:"comment,omitempty"` // Description:"Comment key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
}

// DeleteComment Delete a comment.<br/>Requires authentication and the following permission: 'Browse' on the project of the specified issue.
func (s *IssuesService) DeleteComment(opt *IssuesDeleteCommentOption) (v *IssuesAddCommentObject, resp *http.Response, err error) {
	err = s.ValidateDeleteCommentOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "issues/delete_comment", opt)
	if err != nil {
		return
	}
	v = new(IssuesAddCommentObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type IssuesDoTransitionOption struct {
	Issue      string `url:"issue,omitempty"`      // Description:"Issue key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Transition string `url:"transition,omitempty"` // Description:"Transition",ExampleValue:""
}

// DoTransition Do workflow transition on an issue. Requires authentication and Browse permission on project.<br/>The transitions 'wontfix' and 'falsepositive' require the permission 'Administer Issues'.
func (s *IssuesService) DoTransition(opt *IssuesDoTransitionOption) (v *IssuesAddCommentObject, resp *http.Response, err error) {
	err = s.ValidateDoTransitionOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "issues/do_transition", opt)
	if err != nil {
		return
	}
	v = new(IssuesAddCommentObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type IssuesEditCommentOption struct {
	Comment string `url:"comment,omitempty"` // Description:"Comment key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Text    string `url:"text,omitempty"`    // Description:"Comment text",ExampleValue:"Won't fix because it doesn't apply to the context"
}

// EditComment Edit a comment.<br/>Requires authentication and the following permission: 'Browse' on the project of the specified issue.
func (s *IssuesService) EditComment(opt *IssuesEditCommentOption) (v *IssuesAddCommentObject, resp *http.Response, err error) {
	err = s.ValidateEditCommentOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "issues/edit_comment", opt)
	if err != nil {
		return
	}
	v = new(IssuesAddCommentObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type IssuesSearchOption struct {
	AdditionalFields   string `url:"additionalFields,omitempty"`   // Description:"Comma-separated list of the optional fields to be returned in response. Action plans are dropped in 5.5, it is not returned in the response.",ExampleValue:""
	Asc                string `url:"asc,omitempty"`                // Description:"Ascending sort",ExampleValue:""
	Assigned           string `url:"assigned,omitempty"`           // Description:"To retrieve assigned or unassigned issues",ExampleValue:""
	Assignees          string `url:"assignees,omitempty"`          // Description:"Comma-separated list of assignee logins. The value '__me__' can be used as a placeholder for user who performs the request",ExampleValue:"admin,usera,__me__"
	Authors            string `url:"authors,omitempty"`            // Description:"Comma-separated list of SCM accounts",ExampleValue:"torvalds@linux-foundation.org"
	ComponentKeys      string `url:"componentKeys,omitempty"`      // Description:"Comma-separated list of component keys. Retrieve issues associated to a specific list of components (and all its descendants). A component can be a portfolio, project, module, directory or file.",ExampleValue:"my_project"
	ComponentRootUuids string `url:"componentRootUuids,omitempty"` // Description:"If used, will have the same meaning as componentUuids AND onComponentOnly=false.",ExampleValue:""
	ComponentRoots     string `url:"componentRoots,omitempty"`     // Description:"If used, will have the same meaning as componentKeys AND onComponentOnly=false.",ExampleValue:""
	ComponentUuids     string `url:"componentUuids,omitempty"`     // Description:"To retrieve issues associated to a specific list of components their sub-components (comma-separated list of component IDs). This parameter is mostly used by the Issues page, please prefer usage of the componentKeys parameter. A component can be a project, module, directory or file.",ExampleValue:"584a89f2-8037-4f7b-b82c-8b45d2d63fb2"
	Components         string `url:"components,omitempty"`         // Description:"If used, will have the same meaning as componentKeys AND onComponentOnly=true.",ExampleValue:""
	CreatedAfter       string `url:"createdAfter,omitempty"`       // Description:"To retrieve issues created after the given date (inclusive). <br>Either a date (server timezone) or datetime can be provided. <br>If this parameter is set, createdSince must not be set",ExampleValue:"2017-10-19 or 2017-10-19T13:00:00+0200"
	CreatedAt          string `url:"createdAt,omitempty"`          // Description:"Datetime to retrieve issues created during a specific analysis",ExampleValue:"2017-10-19T13:00:00+0200"
	CreatedBefore      string `url:"createdBefore,omitempty"`      // Description:"To retrieve issues created before the given date (inclusive). <br>Either a date (server timezone) or datetime can be provided.",ExampleValue:"2017-10-19 or 2017-10-19T13:00:00+0200"
	CreatedInLast      string `url:"createdInLast,omitempty"`      // Description:"To retrieve issues created during a time span before the current time (exclusive). Accepted units are 'y' for year, 'm' for month, 'w' for week and 'd' for day. If this parameter is set, createdAfter must not be set",ExampleValue:"1m2w (1 month 2 weeks)"
	Facets             string `url:"facets,omitempty"`             // Description:"Comma-separated list of the facets to be computed. No facet is computed by default."
	Issues             string `url:"issues,omitempty"`             // Description:"Comma-separated list of issue keys",ExampleValue:"5bccd6e8-f525-43a2-8d76-fcb13dde79ef"
	Languages          string `url:"languages,omitempty"`          // Description:"Comma-separated list of languages. Available since 4.4",ExampleValue:"java,js"
	P                  string `url:"p,omitempty"`                  // Description:"1-based page number",ExampleValue:"42"
	Ps                 string `url:"ps,omitempty"`                 // Description:"Page size. Must be greater than 0 and less or equal than 500",ExampleValue:"20"
	Resolutions        string `url:"resolutions,omitempty"`        // Description:"Comma-separated list of resolutions",ExampleValue:"FIXED,REMOVED"
	Resolved           string `url:"resolved,omitempty"`           // Description:"To match resolved or unresolved issues",ExampleValue:""
	Rules              string `url:"rules,omitempty"`              // Description:"Comma-separated list of coding rule keys. Format is &lt;repository&gt;:&lt;rule&gt;",ExampleValue:"squid:AvoidCycles"
	S                  string `url:"s,omitempty"`                  // Description:"Sort field",ExampleValue:""
	Severities         string `url:"severities,omitempty"`         // Description:"Comma-separated list of severities",ExampleValue:"BLOCKER,CRITICAL"
	SinceLeakPeriod    string `url:"sinceLeakPeriod,omitempty"`    // Description:"To retrieve issues created since the leak period.<br>If this parameter is set to a truthy value, createdAfter must not be set and one component id or key must be provided.",ExampleValue:""
	Statuses           string `url:"statuses,omitempty"`           // Description:"Comma-separated list of statuses",ExampleValue:"OPEN,REOPENED"
	Tags               string `url:"tags,omitempty"`               // Description:"Comma-separated list of tags.",ExampleValue:"security,convention"
	Types              string `url:"types,omitempty"`              // Description:"Comma-separated list of types.",ExampleValue:"CODE_SMELL,BUG"
}

// Search Search for issues.<br>At most one of the following parameters can be provided at the same time: componentKeys, componentUuids, components, componentRootUuids, componentRoots.<br>Requires the 'Browse' permission on the specified project(s).
func (s *IssuesService) Search(opt *IssuesSearchOption) (v *IssuesSearchObject, resp *http.Response, err error) {
	err = s.ValidateSearchOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "issues/search", opt)
	if err != nil {
		return
	}
	v = new(IssuesSearchObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type IssuesSetSeverityOption struct {
	Issue    string `url:"issue,omitempty"`    // Description:"Issue key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Severity string `url:"severity,omitempty"` // Description:"New severity",ExampleValue:""
}

// SetSeverity Change severity.<br/>Requires the following permissions:<ul>  <li>'Authentication'</li>  <li>'Browse' rights on project of the specified issue</li>  <li>'Administer Issues' rights on project of the specified issue</li></ul>
func (s *IssuesService) SetSeverity(opt *IssuesSetSeverityOption) (v *IssuesAddCommentObject, resp *http.Response, err error) {
	err = s.ValidateSetSeverityOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "issues/set_severity", opt)
	if err != nil {
		return
	}
	v = new(IssuesAddCommentObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type IssuesSetTagsOption struct {
	Issue string `url:"issue,omitempty"` // Description:"Issue key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Tags  string `url:"tags,omitempty"`  // Description:"Comma-separated list of tags. All tags are removed if parameter is empty or not set.",ExampleValue:"security,cwe,misra-c"
}

// SetTags Set tags on an issue. <br/>Requires authentication and Browse permission on project
func (s *IssuesService) SetTags(opt *IssuesSetTagsOption) (v *IssuesAddCommentObject, resp *http.Response, err error) {
	err = s.ValidateSetTagsOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "issues/set_tags", opt)
	if err != nil {
		return
	}
	v = new(IssuesAddCommentObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type IssuesSetTypeOption struct {
	Issue string `url:"issue,omitempty"` // Description:"Issue key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Type  string `url:"type,omitempty"`  // Description:"New type",ExampleValue:""
}

// SetType Change type of issue, for instance from 'code smell' to 'bug'.<br/>Requires the following permissions:<ul>  <li>'Authentication'</li>  <li>'Browse' rights on project of the specified issue</li>  <li>'Administer Issues' rights on project of the specified issue</li></ul>
func (s *IssuesService) SetType(opt *IssuesSetTypeOption) (v *IssuesAddCommentObject, resp *http.Response, err error) {
	err = s.ValidateSetTypeOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "issues/set_type", opt)
	if err != nil {
		return
	}
	v = new(IssuesAddCommentObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type IssuesTagsOption struct {
	Ps int    `url:"ps,omitempty"` // Description:"Page size. Must be greater than 0 and less or equal than 100",ExampleValue:"20"
	Q  string `url:"q,omitempty"`  // Description:"Limit search to tags that contain the supplied string.",ExampleValue:"misra"
}

// Tags List tags matching a given query
func (s *IssuesService) Tags(opt *IssuesTagsOption) (v *IssuesTagsObject, resp *http.Response, err error) {
	err = s.ValidateTagsOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "issues/tags", opt)
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
