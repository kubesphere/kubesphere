package sdk

// AnnotationResponse is representation of an existing annotation
type AnnotationResponse struct {
	ID          uint                   `json:"id"`
	AlertID     uint                   `json:"alertId"`
	DashboardID uint                   `json:"dashboardId,omitempty"`
	PanelID     uint                   `json:"panelId,omitempty"`
	UserID      uint                   `json:"userId,omitempty"`
	UserName    string                 `json:"userName,omitempty"`
	NewState    string                 `json:"newState,omitempty"`
	PrevState   string                 `json:"prevState,omitempty"`
	Time        int64                  `json:"time,omitempty"`
	TimeEnd     int64                  `json:"timeEnd,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Text        string                 `json:"text,omitempty"`
	Type        string                 `json:"type"`
	Data        map[string]interface{} `json:"data"`
}

// CreateAnnotationRequest is a request to create a new annotation
type CreateAnnotationRequest struct {
	DashboardID uint     `json:"dashboardId,omitempty"`
	PanelID     uint     `json:"panelId,omitempty"`
	Time        int64    `json:"time,omitempty"`
	TimeEnd     int64    `json:"timeEnd,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Text        string   `json:"text,omitempty"`
}

// PatchAnnotationRequest is a request to patch an existing annotation
type PatchAnnotationRequest struct {
	Time    int64    `json:"time,omitempty"`
	TimeEnd int64    `json:"timeEnd,omitempty"`
	Tags    []string `json:"tags,omitempty"`
	Text    string   `json:"text,omitempty"`
}
