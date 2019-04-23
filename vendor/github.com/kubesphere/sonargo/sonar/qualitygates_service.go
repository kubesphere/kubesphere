// Manage quality gates, including conditions and project association.
package sonargo

import "net/http"

type QualitygatesService struct {
	client *Client
}

type QualityGatesCondition struct {
	Error        string `json:"error,omitempty"`
	ID           int    `json:"id,omitempty"`
	Metric       string `json:"metric,omitempty"`
	Op           string `json:"op,omitempty"`
	Warning      string `json:"warning,omitempty"`
	Organization string `url:"organization,omitempty"` // Description:"Organization key. If no organization is provided, the default organization is used.",ExampleValue:"my-org
	Period       string `url:"period,omitempty"`       // Description:"Condition period. If not set, the absolute value is considered.",ExampleValue:"1"
}

type QualitygatesGetByProjectObject struct {
	QualityGate *QualityGate `json:"qualityGate,omitempty"`
}

type QualityGate struct {
	Default bool   `json:"default,omitempty"`
	ID      int    `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
}

type QualitygatesListObject struct {
	Actions      *Actions         `json:"actions,omitempty"`
	Default      int64            `json:"default,omitempty"`
	Qualitygates []*QualityGateV2 `json:"qualitygates,omitempty"`
}

type Actions struct {
	Create            bool `json:"create,omitempty"`
	AssociateProjects bool `json:"associateProjects,omitempty"`
	Copy              bool `json:"copy,omitempty"`
	Delete            bool `json:"delete,omitempty"`
	Edit              bool `json:"edit,omitempty"`
	ManageConditions  bool `json:"manageConditions,omitempty"`
	Rename            bool `json:"rename,omitempty"`
	SetAsDefault      bool `json:"setAsDefault,omitempty"`
}

type QualityGateV2 struct {
	Actions    *Actions                 `json:"actions,omitempty"`
	Conditions []*QualityGatesCondition `json:"conditions,omitempty"`
	ID         int                      `json:"id,omitempty"`
	IsBuiltIn  bool                     `json:"isBuiltIn,omitempty"`
	IsDefault  bool                     `json:"isDefault,omitempty"`
	Name       string                   `json:"name,omitempty"`
}

type QualitygatesProjectStatusObject struct {
	ProjectStatus *ProjectStatus `json:"projectStatus,omitempty"`
}

type Condition struct {
	ActualValue      string `json:"actualValue,omitempty"`
	Comparator       string `json:"comparator,omitempty"`
	ErrorThreshold   string `json:"errorThreshold,omitempty"`
	MetricKey        string `json:"metricKey,omitempty"`
	PeriodIndex      int64  `json:"periodIndex,omitempty"`
	Status           string `json:"status,omitempty"`
	WarningThreshold string `json:"warningThreshold,omitempty"`
}

type ProjectStatus struct {
	Conditions        []*Condition `json:"conditions,omitempty"`
	IgnoredConditions bool         `json:"ignoredConditions,omitempty"`
	Periods           []*Period    `json:"periods,omitempty"`
	Status            string       `json:"status,omitempty"`
}

type QualitygatesSearchObject struct {
	Paging  *Paging   `json:"paging,omitempty"`
	More    bool      `json:"more,omitempty"`
	Results []*Result `json:"results,omitempty"`
}

type Result struct {
	ID       int    `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Selected bool   `json:"selected,omitempty"`
}

type QualitygatesCopyOption struct {
	Id           int    `url:"id,omitempty"`           // Description:"The ID of the source quality gate",ExampleValue:"1"
	Name         string `url:"name,omitempty"`         // Description:"The name of the quality gate to create",ExampleValue:"My Quality Gate"
	Organization string `url:"organization,omitempty"` // Description:"Organization key. If no organization is provided, the default organization is used.",ExampleValue:"my-org"
}

// Copy Copy a Quality Gate.<br>Requires the 'Administer Quality Gates' permission.
func (s *QualitygatesService) Copy(opt *QualitygatesCopyOption) (v *QualityGate, resp *http.Response, err error) {
	err = s.ValidateCopyOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualitygates/copy", opt)
	if err != nil {
		return
	}
	v = new(QualityGate)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return
	}
	return
}

type QualitygatesCreateOption struct {
	Name         string `url:"name,omitempty"`         // Description:"The name of the quality gate to create",ExampleValue:"My Quality Gate"
	Organization string `url:"organization,omitempty"` // Description:"Organization key. If no organization is provided, the default organization is used.",ExampleValue:"my-org"
}

// Create Create a Quality Gate.<br>Requires the 'Administer Quality Gates' permission.
func (s *QualitygatesService) Create(opt *QualitygatesCreateOption) (v *QualityGate, resp *http.Response, err error) {
	err = s.ValidateCreateOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualitygates/create", opt)
	if err != nil {
		return
	}
	v = new(QualityGate)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type QualitygatesCreateConditionOption struct {
	Error        string `url:"error,omitempty"`        // Description:"Condition error threshold",ExampleValue:"10"
	GateId       int    `url:"gateId,omitempty"`       // Description:"ID of the quality gate",ExampleValue:"1"
	Metric       string `url:"metric,omitempty"`       // Description:"Condition metric",ExampleValue:"blocker_violations"
	Op           string `url:"op,omitempty"`           // Description:"Condition operator:<br/><ul><li>EQ = equals</li><li>NE = is not</li><li>LT = is lower than</li><li>GT = is greater than</li></ui>",ExampleValue:"EQ"
	Organization string `url:"organization,omitempty"` // Description:"Organization key. If no organization is provided, the default organization is used.",ExampleValue:"my-org"
	Period       string `url:"period,omitempty"`       // Description:"Condition period. If not set, the absolute value is considered.",ExampleValue:""
	Warning      string `url:"warning,omitempty"`      // Description:"Condition warning threshold",ExampleValue:"5"
}

// CreateCondition Add a new condition to a quality gate.<br>Requires the 'Administer Quality Gates' permission.
func (s *QualitygatesService) CreateCondition(opt *QualitygatesCreateConditionOption) (v *QualityGatesCondition, resp *http.Response, err error) {
	err = s.ValidateCreateConditionOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualitygates/create_condition", opt)
	if err != nil {
		return
	}
	v = new(QualityGatesCondition)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type QualitygatesDeleteConditionOption struct {
	ConditionID  int    `url:"id,omitempty"`           // Description:"Condition ID",ExampleValue:"2"
	Organization string `url:"organization,omitempty"` // Description:"Organization key. If no organization is provided, the default organization is used.",ExampleValue:"my-org"
}

// DeleteCondition Delete a condition from a quality gate.<br>Requires the 'Administer Quality Gates' permission.
func (s *QualitygatesService) DeleteCondition(opt *QualitygatesDeleteConditionOption) (resp *http.Response, err error) {
	err = s.ValidateDeleteConditionOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualitygates/delete_condition", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type QualitygatesDeselectOption struct {
	Organization string `url:"organization,omitempty"` // Description:"Organization key. If no organization is provided, the default organization is used.",ExampleValue:"my-org"
	ProjectId    string `url:"projectId,omitempty"`    // Description:"Project id",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	ProjectKey   string `url:"projectKey,omitempty"`   // Description:"Project key",ExampleValue:"my_project"
}

// Deselect Remove the association of a project from a quality gate.<br>Requires one of the following permissions:<ul><li>'Administer Quality Gates'</li><li>'Administer' rights on the project</li></ul>
func (s *QualitygatesService) Deselect(opt *QualitygatesDeselectOption) (resp *http.Response, err error) {
	err = s.ValidateDeselectOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualitygates/deselect", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type QualitygatesDestroyOption struct {
	Id           int    `url:"id,omitempty"`           // Description:"ID of the quality gate to delete",ExampleValue:"1"
	Organization string `url:"organization,omitempty"` // Description:"Organization key. If no organization is provided, the default organization is used.",ExampleValue:"my-org"
}

// Destroy Delete a Quality Gate.<br>Requires the 'Administer Quality Gates' permission.
func (s *QualitygatesService) Destroy(opt *QualitygatesDestroyOption) (resp *http.Response, err error) {
	err = s.ValidateDestroyOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualitygates/destroy", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type QualitygatesGetByProjectOption struct {
	Organization string `url:"organization,omitempty"` // Description:"Organization key. If no organization is provided, the default organization is used.",ExampleValue:"my-org"
	Project      string `url:"project,omitempty"`      // Description:"Project key",ExampleValue:"my_project"
}

// GetByProject Get the quality gate of a project.<br />Requires one of the following permissions:<ul><li>'Administer System'</li><li>'Administer' rights on the specified project</li><li>'Browse' on the specified project</li></ul>
func (s *QualitygatesService) GetByProject(opt *QualitygatesGetByProjectOption) (v *QualitygatesGetByProjectObject, resp *http.Response, err error) {
	err = s.ValidateGetByProjectOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "qualitygates/get_by_project", opt)
	if err != nil {
		return
	}
	v = new(QualitygatesGetByProjectObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type QualitygatesListOption struct {
	Organization string `url:"organization,omitempty"` // Description:"Organization key. If no organization is provided, the default organization is used.",ExampleValue:"my-org"
}

// List Get a list of quality gates
func (s *QualitygatesService) List(opt *QualitygatesListOption) (v *QualitygatesListObject, resp *http.Response, err error) {
	err = s.ValidateListOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "qualitygates/list", opt)
	if err != nil {
		return
	}
	v = new(QualitygatesListObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type QualitygatesProjectStatusOption struct {
	AnalysisId string `url:"analysisId,omitempty"` // Description:"Analysis id",ExampleValue:"AU-TpxcA-iU5OvuD2FL1"
	ProjectId  string `url:"projectId,omitempty"`  // Description:"Project id",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	ProjectKey string `url:"projectKey,omitempty"` // Description:"Project key",ExampleValue:"my_project"
}

// ProjectStatus Get the quality gate status of a project or a Compute Engine task.<br />Either 'analysisId', 'projectId' or 'projectKey' must be provided<br />The different statuses returned are: OK, WARN, ERROR, NONE. The NONE status is returned when there is no quality gate associated with the analysis.<br />Returns an HTTP code 404 if the analysis associated with the task is not found or does not exist.<br />Requires one of the following permissions:<ul><li>'Administer System'</li><li>'Administer' rights on the specified project</li><li>'Browse' on the specified project</li></ul>
func (s *QualitygatesService) ProjectStatus(opt *QualitygatesProjectStatusOption) (v *QualitygatesProjectStatusObject, resp *http.Response, err error) {
	err = s.ValidateProjectStatusOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "qualitygates/project_status", opt)
	if err != nil {
		return
	}
	v = new(QualitygatesProjectStatusObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type QualitygatesRenameOption QualitygatesCopyOption

// Rename Rename a Quality Gate.<br>Requires the 'Administer Quality Gates' permission.
func (s *QualitygatesService) Rename(opt *QualitygatesRenameOption) (v *QualityGate, resp *http.Response, err error) {
	err = s.ValidateRenameOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualitygates/rename", opt)
	if err != nil {
		return
	}
	v = new(QualityGate)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return
	}
	return
}

type QualitygatesSearchOption struct {
	GateId       int    `url:"gateId,omitempty"`       // Description:"Quality Gate ID",ExampleValue:"1"
	Organization string `url:"organization,omitempty"` // Description:"Organization key. If no organization is provided, the default organization is used.",ExampleValue:"my-org"
	Page         string `url:"page,omitempty"`         // Description:"Page number",ExampleValue:"2"
	PageSize     string `url:"pageSize,omitempty"`     // Description:"Page size",ExampleValue:"10"
	Query        string `url:"query,omitempty"`        // Description:"To search for projects containing this string. If this parameter is set, "selected" is set to "all".",ExampleValue:"abc"
	Selected     string `url:"selected,omitempty"`     // Description:"Depending on the value, show only selected items (selected=selected), deselected items (selected=deselected), or all items with their selection status (selected=all).",ExampleValue:""
}

// Search Search for projects associated (or not) to a quality gate.<br/>Only authorized projects for current user will be returned.
func (s *QualitygatesService) Search(opt *QualitygatesSearchOption) (v *QualitygatesSearchObject, resp *http.Response, err error) {
	err = s.ValidateSearchOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "qualitygates/search", opt)
	if err != nil {
		return
	}
	v = new(QualitygatesSearchObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type QualitygatesSelectOption struct {
	GateId       int    `url:"gateId,omitempty"`       // Description:"Quality gate id",ExampleValue:"1"
	Organization string `url:"organization,omitempty"` // Description:"Organization key. If no organization is provided, the default organization is used.",ExampleValue:"my-org"
	ProjectKey   string `url:"projectKey,omitempty"`   // Description:"Project key",ExampleValue:"my_project"
	ProjectID    string `url:"projectId,omitempty"`    // Description:"Project id",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
}

// Select Associate a project to a quality gate.<br>The 'projectId' or 'projectKey' must be provided.<br>Project id as a numeric value is deprecated since 6.1. Please use the id similar to 'AU-TpxcA-iU5OvuD2FLz'.<br>Requires the 'Administer Quality Gates' permission.
func (s *QualitygatesService) Select(opt *QualitygatesSelectOption) (resp *http.Response, err error) {
	err = s.ValidateSelectOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualitygates/select", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type QualitygatesSetAsDefaultOption QualitygatesDestroyOption

// SetAsDefault Set a quality gate as the default quality gate.<br>Requires the 'Administer Quality Gates' permission.
func (s *QualitygatesService) SetAsDefault(opt *QualitygatesSetAsDefaultOption) (resp *http.Response, err error) {
	err = s.ValidateSetAsDefaultOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualitygates/set_as_default", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type QualitygatesShowOption QualitygatesRenameOption

// Show Display the details of a quality gate
func (s *QualitygatesService) Show(opt *QualitygatesShowOption) (v *QualityGateV2, resp *http.Response, err error) {
	err = s.ValidateShowOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "qualitygates/show", opt)
	if err != nil {
		return
	}
	v = new(QualityGateV2)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type QualitygatesUpdateConditionOption struct {
	Error        string `url:"error,omitempty"`        // Description:"Condition error threshold",ExampleValue:"10"
	Id           int    `url:"id,omitempty"`           // Description:"Condition ID",ExampleValue:"10"
	Metric       string `url:"metric,omitempty"`       // Description:"Condition metric",ExampleValue:"blocker_violations"
	Op           string `url:"op,omitempty"`           // Description:"Condition operator:<br/><ul><li>EQ = equals</li><li>NE = is not</li><li>LT = is lower than</li><li>GT = is greater than</li></ui>",ExampleValue:"EQ"
	Organization string `url:"organization,omitempty"` // Description:"Organization key. If no organization is provided, the default organization is used.",ExampleValue:"my-org"
	Period       string `url:"period,omitempty"`       // Description:"Condition period. If not set, the absolute value is considered.",ExampleValue:""
	Warning      string `url:"warning,omitempty"`      // Description:"Condition warning threshold",ExampleValue:"5"
}

// UpdateCondition Update a condition attached to a quality gate.<br>Requires the 'Administer Quality Gates' permission.
func (s *QualitygatesService) UpdateCondition(opt *QualitygatesUpdateConditionOption) (v *QualityGatesCondition, resp *http.Response, err error) {
	err = s.ValidateUpdateConditionOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "qualitygates/update_condition", opt)
	if err != nil {
		return
	}
	v = new(QualityGatesCondition)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return
	}
	return
}
