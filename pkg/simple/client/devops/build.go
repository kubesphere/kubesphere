/*
Copyright 2020 KubeSphere Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package devops

const (
	LastBuild             = "lastBuild"
	LastCompletedBuild    = "lastCompletedBuild"
	LastFailedBuild       = "lastFailedBuild"
	LastStableBuild       = "lastStableBuild"
	LastSuccessfulBuild   = "lastSuccessfulBuild"
	LastUnstableBuild     = "lastUnstableBuild"
	LastUnsuccessfulBuild = "lastUnsuccessfulBuild"
	FirstBuild            = "firstBuild"
)

type GeneralParameter struct {
	Name  string
	Value string
}

type Branch struct {
	SHA1 string `json:",omitempty"`
	Name string `json:",omitempty"`
}

type BuildRevision struct {
	SHA1   string   `json:"SHA1,omitempty"`
	Branch []Branch `json:"Branch,omitempty"`
}

type Builds struct {
	BuildNumber int64         `json:"buildNumber"`
	BuildResult interface{}   `json:"buildResult"`
	Marked      BuildRevision `json:"marked"`
	Revision    BuildRevision `json:"revision"`
}

type Culprit struct {
	AbsoluteUrl string
	FullName    string
}

type GeneralAction struct {
	Parameters         []GeneralParameter       `json:"parameters,omitempty"`
	Causes             []map[string]interface{} `json:"causes,omitempty"`
	BuildsByBranchName map[string]Builds        `json:"buildsByBranchName,omitempty"`
	LastBuiltRevision  *BuildRevision           `json:"lastBuiltRevision,omitempty"`
	RemoteUrls         []string                 `json:"remoteUrls,omitempty"`
	ScmName            string                   `json:"scmName,omitempty"`
	Subdir             interface{}              `json:"subdir,omitempty"`
	ClassName          string                   `json:"_class,omitempty"`
	SonarTaskId        string                   `json:"ceTaskId,omitempty"`
	SonarServerUrl     string                   `json:"serverUrl,omitempty"`
	SonarDashboardUrl  string                   `json:"sonarqubeDashboardUrl,omitempty"`
	TotalCount         int64                    `json:",omitempty"`
	UrlName            string                   `json:",omitempty"`
}

type Build struct {
	Actions   []GeneralAction
	Artifacts []struct {
		DisplayPath  string `json:"displayPath"`
		FileName     string `json:"fileName"`
		RelativePath string `json:"relativePath"`
	} `json:"artifacts"`
	Building  bool   `json:"building"`
	BuiltOn   string `json:"builtOn"`
	ChangeSet struct {
		Items []struct {
			AffectedPaths []string `json:"affectedPaths"`
			Author        struct {
				AbsoluteUrl string `json:"absoluteUrl"`
				FullName    string `json:"fullName"`
			} `json:"author"`
			Comment  string `json:"comment"`
			CommitID string `json:"commitId"`
			Date     string `json:"date"`
			ID       string `json:"id"`
			Msg      string `json:"msg"`
			Paths    []struct {
				EditType string `json:"editType"`
				File     string `json:"file"`
			} `json:"paths"`
			Timestamp int64 `json:"timestamp"`
		} `json:"items"`
		Kind      string `json:"kind"`
		Revisions []struct {
			Module   string
			Revision int
		} `json:"revision"`
	} `json:"changeSet"`
	Culprits          []Culprit   `json:"culprits"`
	Description       interface{} `json:"description"`
	Duration          int64       `json:"duration"`
	EstimatedDuration int64       `json:"estimatedDuration"`
	Executor          interface{} `json:"executor"`
	FullDisplayName   string      `json:"fullDisplayName"`
	ID                string      `json:"id"`
	KeepLog           bool        `json:"keepLog"`
	Number            int64       `json:"number"`
	QueueID           int64       `json:"queueId"`
	Result            string      `json:"result"`
	Timestamp         int64       `json:"timestamp"`
	URL               string      `json:"url"`
	Runs              []struct {
		Number int64
		URL    string
	} `json:"runs"`
}

type BuildGetter interface {
	// GetProjectPipelineBuildByType get the last build of the pipeline, status can specify the status of the last build.
	GetProjectPipelineBuildByType(projectId, pipelineId string, status string) (*Build, error)

	// GetMultiBranchPipelineBuildByType get the last build of the pipeline, status can specify the status of the last build.
	GetMultiBranchPipelineBuildByType(projectId, pipelineId, branch string, status string) (*Build, error)
}
