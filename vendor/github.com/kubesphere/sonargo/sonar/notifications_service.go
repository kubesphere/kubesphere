// Manage notifications of the authenticated user
package sonargo

import "net/http"

type NotificationsService struct {
	client *Client
}

type NotificationsListObject struct {
	Channels        []string        `json:"channels,omitempty"`
	GlobalTypes     []string        `json:"globalTypes,omitempty"`
	Notifications   []*Notification `json:"notifications,omitempty"`
	PerProjectTypes []string        `json:"perProjectTypes,omitempty"`
}

type Notification struct {
	Channel      string `json:"channel,omitempty"`
	Organization string `json:"organization,omitempty"`
	Project      string `json:"project,omitempty"`
	ProjectName  string `json:"projectName,omitempty"`
	Type         string `json:"type,omitempty"`
}

type NotificationsAddOption struct {
	Channel string `url:"channel,omitempty"` // Description:"Channel through which the notification is sent. For example, notifications can be sent by email.",ExampleValue:""
	Login   string `url:"login,omitempty"`   // Description:"User login",ExampleValue:""
	Project string `url:"project,omitempty"` // Description:"Project key",ExampleValue:"my_project"
	Type    string `url:"type,omitempty"`    // Description:"Notification type. Possible values are for:<ul>  <li>Global notifications: CeReportTaskFailure, ChangesOnMyIssue, NewAlerts, NewFalsePositiveIssue, NewIssues, SQ-MyNewIssues</li>  <li>Per project notifications: CeReportTaskFailure, ChangesOnMyIssue, NewAlerts, NewFalsePositiveIssue, NewIssues, SQ-MyNewIssues</li></ul>",ExampleValue:"SQ-MyNewIssues"
}

// Add Add a notification for the authenticated user.<br>Requires one of the following permissions:<ul> <li>Authentication if no login is provided. If a project is provided, requires the 'Browse' permission on the specified project.</li> <li>System administration if a login is provided. If a project is provided, requires the 'Browse' permission on the specified project.</li></ul>
func (s *NotificationsService) Add(opt *NotificationsAddOption) (resp *http.Response, err error) {
	err = s.ValidateAddOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "notifications/add", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}

type NotificationsListOption struct {
	Login string `url:"login,omitempty"` // Description:"User login",ExampleValue:""
}

// List List notifications of the authenticated user.<br>Requires one of the following permissions:<ul>  <li>Authentication if no login is provided</li>  <li>System administration if a login is provided</li></ul>
func (s *NotificationsService) List(opt *NotificationsListOption) (v *NotificationsListObject, resp *http.Response, err error) {
	err = s.ValidateListOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("GET", "notifications/list", opt)
	if err != nil {
		return
	}
	v = new(NotificationsListObject)
	resp, err = s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}
	return
}

type NotificationsRemoveOption struct {
	Channel string `url:"channel,omitempty"` // Description:"Channel through which the notification is sent. For example, notifications can be sent by email.",ExampleValue:""
	Login   string `url:"login,omitempty"`   // Description:"User login",ExampleValue:""
	Project string `url:"project,omitempty"` // Description:"Project key",ExampleValue:"my_project"
	Type    string `url:"type,omitempty"`    // Description:"Notification type. Possible values are for:<ul>  <li>Global notifications: CeReportTaskFailure, ChangesOnMyIssue, NewAlerts, NewFalsePositiveIssue, NewIssues, SQ-MyNewIssues</li>  <li>Per project notifications: CeReportTaskFailure, ChangesOnMyIssue, NewAlerts, NewFalsePositiveIssue, NewIssues, SQ-MyNewIssues</li></ul>",ExampleValue:"SQ-MyNewIssues"
}

// Remove Remove a notification for the authenticated user.<br>Requires one of the following permissions:<ul>  <li>Authentication if no login is provided</li>  <li>System administration if a login is provided</li></ul>
func (s *NotificationsService) Remove(opt *NotificationsRemoveOption) (resp *http.Response, err error) {
	err = s.ValidateRemoveOpt(opt)
	if err != nil {
		return
	}
	req, err := s.client.NewRequest("POST", "notifications/remove", opt)
	if err != nil {
		return
	}
	resp, err = s.client.Do(req, nil)
	if err != nil {
		return
	}
	return
}
