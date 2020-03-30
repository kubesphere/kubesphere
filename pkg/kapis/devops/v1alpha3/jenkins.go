package v1alpha3

import "strings"

const (
	JenkinsJobStarted        = "jenkins.job.started"
	JenkinsJobCompleted      = "jenkins.job.completed"
	JenkinsJobFinalized      = "jenkins.job.finalized"
	JenkinsJobInputStarted   = "jenkins.job.input.started"
	JenkinsJobInputProceeded = "jenkins.job.input.proceeded"
	JenkinsJobInputAborted   = "jenkins.job.input.aborted"
)

var JenkinsEventTypeMap = map[string]string{
	JenkinsJobStarted:        PipelineStarted,
	JenkinsJobCompleted:      PipelineCompleted,
	JenkinsJobFinalized:      PipelineFinalized,
	JenkinsJobInputStarted:   PipelinePendingReview,
	JenkinsJobInputProceeded: PipelineReviewProceeded,
	JenkinsJobInputAborted:   PipelineReviewAborted,
}

type JenkinsBuildState struct {
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
	InputState  ReviewState         `json:"inputState,omitempty"`
}

type JenkinsJobState struct {
	Name                   string            `json:"name"`
	DisplayName            string            `json:"displayName"`
	Url                    string            `json:"url"`
	Build                  JenkinsBuildState `json:"build"`
	PreviousCompletedBuild JenkinsBuildState `json:"previousCompletedBuild"`
}

type JenkinsEventArgs struct {
	JobState JenkinsJobState `json:"jobState"`
}
type JenkinsEvent struct {
	Timestamp int64            `json:"timestamp"`
	Type      string           `json:"type"`
	Args      JenkinsEventArgs `json:"args"`
}

func (j JenkinsEvent) ToEvent() Event {
	return Event{
		Timestamp: j.Timestamp,
		Type:      JenkinsEventTypeMap[j.Type],
		Args:      EventArgs{PipelineState: j.Args.JobState.ToPipelineState()},
	}
}

func (j JenkinsBuildState) ToBuildState() BuildState {
	return BuildState{
		Artifacts:   j.Artifacts,
		FullUrl:     j.FullUrl,
		Number:      j.Number,
		QueueId:     j.QueueId,
		Phase:       j.Phase,
		Timestamp:   j.Timestamp,
		Status:      j.Status,
		Url:         j.Url,
		DisplayName: j.DisplayName,
		Parameters:  j.Parameters,
		TestSummary: j.TestSummary,
		ReviewState: j.InputState,
	}
}

func (j JenkinsJobState) ToPipelineState() PipelineState {
	s := strings.Split(j.Url, "job/")
	if len(s) != 3 {
		return PipelineState{}
	}
	return PipelineState{
		Name:                   j.Name,
		DisplayName:            j.DisplayName,
		Url:                    j.Url,
		ProjectId:              strings.Trim(s[1], "/"),
		Pipeline:               strings.Trim(s[2], "/"),
		Build:                  j.Build.ToBuildState(),
		PreviousCompletedBuild: j.PreviousCompletedBuild.ToBuildState(),
	}
}
