// Manage project analyses.
package sonargo

import "net/http"

type ProjectAnalysesService struct {
	client *Client
}

type EventCatagory string

const (
	VersionEventCatagory        EventCatagory = "VERSION"
	OtherEventCatagory          EventCatagory = "OTHER"
	QualityProfileEventCatagory EventCatagory = "QUALITY_PROFILE"
	QualityGateEventCatagory    EventCatagory = "QUALITY_GATE"
)

type Event struct {
	Analysis    string        `json:"analysis,omitempty"`
	Key         string        `json:"key,omitempty"`
	Category    EventCatagory `json:"category,omitempty"`
	Name        string        `json:"name,omitempty"`
	Description string        `json:"description,omitempty"`
}

type ProjectAnalysesEventObject struct {
	Event *Event `json:"event,omitempty"`
}

type Analysis struct {
	Key    string   `json:"key,omitempty"`
	Date   string   `json:"date,omitempty"`
	Events []*Event `json:"events,omitempty"`
}

type ProjectAnalysesSearchObject struct {
	Paging   *Paging     `json:"paging,omitempty"`
	Analyses []*Analysis `json:"analyses,omitempty"`
}

type ProjectAnalysesCreateEventOption struct {
	Analysis string        `url:"analysis,omitempty"` // Description:"Analysis key",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
	Category EventCatagory `url:"category,omitempty"` // Description:"Category",ExampleValue:""
	Name     string        `url:"name,omitempty"`     // Description:"Name",ExampleValue:"5.6"
}

// CreateEvent Create a project analysis event.<br>Only event of category 'VERSION' and 'OTHER' can be created.<br>Requires one of the following permissions:<ul>  <li>'Administer System'</li>  <li>'Administer' rights on the specified project</li></ul>
func (s *ProjectAnalysesService) CreateEvent(opt *ProjectAnalysesCreateEventOption) (v *ProjectAnalysesEventObject, resp *http.Response, err error) {
	err = s.ValidateCreateEventOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "project_analyses/create_event", opt)
	if err != nil {
		return
	}
	v = new(ProjectAnalysesEventObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type ProjectAnalysesDeleteOption struct {
	Analysis string `url:"analysis,omitempty"` // Description:"Analysis key",ExampleValue:"AU-TpxcA-iU5OvuD2FL1"
}

// Delete Delete a project analysis.<br>Requires one of the following permissions:<ul>  <li>'Administer System'</li>  <li>'Administer' rights on the project of the specified analysis</li></ul>
func (s *ProjectAnalysesService) Delete(opt *ProjectAnalysesDeleteOption) (resp *http.Response, err error) {
	err = s.ValidateDeleteOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "project_analyses/delete", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type ProjectAnalysesDeleteEventOption struct {
	Event string `url:"event,omitempty"` // Description:"Event key",ExampleValue:"AU-TpxcA-iU5OvuD2FLz"
}

// DeleteEvent Delete a project analysis event.<br>Only event of category 'VERSION' and 'OTHER' can be deleted.<br>Requires one of the following permissions:<ul>  <li>'Administer System'</li>  <li>'Administer' rights on the specified project</li></ul>
func (s *ProjectAnalysesService) DeleteEvent(opt *ProjectAnalysesDeleteEventOption) (resp *http.Response, err error) {
	err = s.ValidateDeleteEventOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "project_analyses/delete_event", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type ProjectAnalysesSearchOption struct {
	Category string `url:"category,omitempty"` // Description:"Event category. Filter analyses that have at least one event of the category specified.",ExampleValue:"OTHER"
	From     string `url:"from,omitempty"`     // Description:"Filter analyses created after the given date (inclusive). <br>Either a date (server timezone) or datetime can be provided",ExampleValue:"2013-05-01"
	P        int    `url:"p,omitempty"`        // Description:"1-based page number",ExampleValue:"42"
	Project  string `url:"project,omitempty"`  // Description:"Project key",ExampleValue:"my_project"
	Ps       int    `url:"ps,omitempty"`       // Description:"Page size. Must be greater than 0 and less or equal than 500",ExampleValue:"20"
	To       string `url:"to,omitempty"`       // Description:"Filter analyses created before the given date (inclusive). <br>Either a date (server timezone) or datetime can be provided",ExampleValue:"2017-10-19 or 2017-10-19T13:00:00+0200"
}

// Search Search a project analyses and attached events.<br>Requires the following permission: 'Browse' on the specified project
func (s *ProjectAnalysesService) Search(opt *ProjectAnalysesSearchOption) (v *ProjectAnalysesSearchObject, resp *http.Response, err error) {
	err = s.ValidateSearchOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "project_analyses/search", opt)
	if err != nil {
		return
	}
	v = new(ProjectAnalysesSearchObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type ProjectAnalysesUpdateEventOption struct {
	Event string `url:"event,omitempty"` // Description:"Event key",ExampleValue:"AU-TpxcA-iU5OvuD2FL5"
	Name  string `url:"name,omitempty"`  // Description:"New name",ExampleValue:"5.6"
}

// UpdateEvent Update a project analysis event.<br>Only events of category 'VERSION' and 'OTHER' can be updated.<br>Requires one of the following permissions:<ul>  <li>'Administer System'</li>  <li>'Administer' rights on the specified project</li></ul>
func (s *ProjectAnalysesService) UpdateEvent(opt *ProjectAnalysesUpdateEventOption) (v *ProjectAnalysesEventObject, resp *http.Response, err error) {
	err = s.ValidateUpdateEventOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "project_analyses/update_event", opt)
	if err != nil {
		return
	}
	v = new(ProjectAnalysesEventObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}
