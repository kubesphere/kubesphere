package v1alpha3

const (
	PipelineStarted         = "pipeline.started"
	PipelineCompleted       = "pipeline.completed"
	PipelineFinalized       = "pipeline.finalized"
	PipelinePendingReview   = "pipeline.pendingReview"
	PipelineReviewProceeded = "pipeline.reviewProceeded"
	PipelineReviewAborted   = "pipeline.reviewAborted"
)

type Artifact struct {
	Archive string `json:"archive,omitempty"`
}

type TestState struct {
	Total       int      `json:"total,omitempty"`
	Failed      int      `json:"failed,omitempty"`
	Passed      int      `json:"passed,omitempty"`
	Skipped     int      `json:"skipped,omitempty"`
	FailedTests []string `json:"failedTests,omitempty"`
}

type ReviewState struct {
	Message   string   `json:"message,omitempty"`
	Id        string   `json:"id"`
	Submitter []string `json:"submitter,omitempty"`
	Approver  string   `json:"approver,omitempty"`
}

type BuildState struct {
	Artifacts   map[string]Artifact `json:"artifacts,omitempty"`
	FullUrl     string              `json:"fullUrl"`
	Number      int                 `json:"number"`
	QueueId     int                 `json:"queueId"`
	Phase       string              `json:"phase"`
	Timestamp   int64               `json:"timestamp"`
	Status      string              `json:"status"`
	Url         string              `json:"url"`
	DisplayName string              `json:"displayName"`
	Parameters  map[string]string   `json:"parameters"`
	TestSummary TestState           `json:"testSummary,omitempty"`
	ReviewState ReviewState         `json:"reviewState,omitempty"`
}

type PipelineState struct {
	Name                   string     `json:"name"`
	DisplayName            string     `json:"displayName"`
	Url                    string     `json:"url"`
	ProjectId              string     `json:"projectId"`
	Pipeline               string     `json:"pipeline"`
	Build                  BuildState `json:"build"`
	PreviousCompletedBuild BuildState `json:"previousCompletedBuild"`
}

type EventArgs struct {
	PipelineState PipelineState `json:"pipelineState"`
}

type Event struct {
	Timestamp int64     `json:"timestamp"`
	Type      string    `json:"type"`
	Args      EventArgs `json:"args"`
}
