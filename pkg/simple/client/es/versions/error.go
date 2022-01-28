package versions

type Error struct {
	Status  int           `json:"status"`
	Details *ErrorDetails `json:"error,omitempty"`
}

type ErrorDetails struct {
	Type         string                   `json:"type"`
	Reason       string                   `json:"reason"`
	ResourceType string                   `json:"resource.type,omitempty"`
	ResourceId   string                   `json:"resource.id,omitempty"`
	Index        string                   `json:"index,omitempty"`
	Phase        string                   `json:"phase,omitempty"`
	Grouped      bool                     `json:"grouped,omitempty"`
	CausedBy     map[string]interface{}   `json:"caused_by,omitempty"`
	RootCause    []*ErrorDetails          `json:"root_cause,omitempty"`
	FailedShards []map[string]interface{} `json:"failed_shards,omitempty"`
}
