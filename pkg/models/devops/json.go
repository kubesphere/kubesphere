/*

 Copyright 2019 The KubeSphere Authors.

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

type Pipeline struct {
	Class                          string        `json:"_class,omitempty"`
	Links                          Links         `json:"_links,omitempty"`
	Actions                        []interface{} `json:"actions,omitempty"`
	DisplayName                    string        `json:"displayName,omitempty"`
	FullDisplayName                string        `json:"fullDisplayName,omitempty"`
	FullName                       string        `json:"fullName,omitempty"`
	Name                           interface{}   `json:"name,omitempty"`
	Organization                   string        `json:"organization,omitempty"`
	Parameters                     interface{}   `json:"parameters,omitempty"`
	Permissions                    Permissions   `json:"permissions,omitempty"`
	EstimatedDurationInMillis      int           `json:"estimatedDurationInMillis,omitempty"`
	NumberOfFolders                int           `json:"numberOfFolders,omitempty"`
	NumberOfPipelines              int           `json:"numberOfPipelines,omitempty"`
	PipelineFolderNames            []interface{} `json:"pipelineFolderNames,omitempty"`
	WeatherScore                   int           `json:"weatherScore,omitempty"`
	BranchNames                    []string      `json:"branchNames,omitempty"`
	NumberOfFailingBranches        int           `json:"numberOfFailingBranches,omitempty"`
	NumberOfFailingPullRequests    int           `json:"numberOfFailingPullRequests,omitempty"`
	NumberOfSuccessfulBranches     int           `json:"numberOfSuccessfulBranches,omitempty"`
	NumberOfSuccessfulPullRequests int           `json:"numberOfSuccessfulPullRequests,omitempty"`
	ScmSource                      ScmSource     `json:"scmSource,omitempty"`
	TotalNumberOfBranches          int           `json:"totalNumberOfBranches,omitempty"`
	TotalNumberOfPullRequests      int           `json:"totalNumberOfPullRequests,omitempty"`
	ArtifactsZipFile               interface{}   `json:"artifactsZipFile,omitempty"`
	CauseOfBlockage                interface{}   `json:"causeOfBlockage,omitempty"`
	Causes                         []Causes      `json:"causes,omitempty"`
	ChangeSet                      []interface{} `json:"changeSet,omitempty"`
	Description                    interface{}   `json:"description,omitempty"`
	DurationInMillis               int           `json:"durationInMillis,omitempty"`
	EnQueueTime                    string        `json:"enQueueTime,omitempty"`
	EndTime                        string        `json:"endTime,omitempty"`
	ID                             string        `json:"id,omitempty"`
	Pipeline                       string        `json:"pipeline,omitempty"`
	Replayable                     bool          `json:"replayable,omitempty"`
	Result                         string        `json:"result,omitempty"`
	RunSummary                     string        `json:"runSummary,omitempty"`
	StartTime                      string        `json:"startTime,omitempty"`
	State                          string        `json:"state,omitempty"`
	Type                           string        `json:"type,omitempty"`
	Branch                         Branch        `json:"branch,omitempty"`
	CommitID                       string        `json:"commitId,omitempty"`
	CommitURL                      interface{}   `json:"commitUrl,omitempty"`
	PullRequest                    interface{}   `json:"pullRequest,omitempty"`
}
type Self struct {
	Class string `json:"_class,omitempty"`
	Href  string `json:"href,omitempty"`
}
type Scm struct {
	Class string `json:"_class,omitempty"`
	Href  string `json:"href,omitempty"`
}
type Branches struct {
	Class string `json:"_class,omitempty"`
	Href  string `json:"href,omitempty"`
}
type Actions struct {
	Class string `json:"_class,omitempty"`
	Href  string `json:"href,omitempty"`
}
type Runs struct {
	Class string `json:"_class,omitempty"`
	Href  string `json:"href,omitempty"`
}
type Trends struct {
	Class string `json:"_class,omitempty"`
	Href  string `json:"href,omitempty"`
}
type Queue struct {
	Class string `json:"_class,omitempty"`
	Href  string `json:"href,omitempty"`
}
type Links struct {
	Self            Self            `json:"self,omitempty"`
	Scm             Scm             `json:"scm,omitempty"`
	Branches        Branches        `json:"branches,omitempty"`
	Actions         Actions         `json:"actions,omitempty"`
	Runs            Runs            `json:"runs,omitempty"`
	Trends          Trends          `json:"trends,omitempty"`
	Queue           Queue           `json:"queue,omitempty"`
	PrevRun         PrevRun         `json:"prevRun"`
	Parent          Parent          `json:"parent"`
	Tests           Tests           `json:"tests"`
	Nodes           Nodes           `json:"nodes"`
	Log             Log             `json:"log"`
	BlueTestSummary BlueTestSummary `json:"blueTestSummary"`
	Steps           Steps           `json:"steps"`
	Artifacts       Artifacts       `json:"artifacts"`
}
type Permissions struct {
	Create    bool `json:"create,omitempty"`
	Configure bool `json:"configure,omitempty"`
	Read      bool `json:"read,omitempty"`
	Start     bool `json:"start,omitempty"`
	Stop      bool `json:"stop,omitempty"`
}
type ScmSource struct {
	Class  string      `json:"_class,omitempty"`
	APIURL interface{} `json:"apiUrl,omitempty"`
	ID     string      `json:"id,omitempty"`
}
type PrevRun struct {
	Class string `json:"_class"`
	Href  string `json:"href"`
}
type Parent struct {
	Class string `json:"_class"`
	Href  string `json:"href"`
}
type Tests struct {
	Class string `json:"_class"`
	Href  string `json:"href"`
}
type Nodes struct {
	Class string `json:"_class"`
	Href  string `json:"href"`
}
type Log struct {
	Class string `json:"_class"`
	Href  string `json:"href"`
}
type BlueTestSummary struct {
	Class string `json:"_class"`
	Href  string `json:"href"`
}
type Steps struct {
	Class string `json:"_class"`
	Href  string `json:"href"`
}
type Artifacts struct {
	Class string `json:"_class"`
	Href  string `json:"href"`
}
type Causes struct {
	Class            string `json:"_class"`
	ShortDescription string `json:"shortDescription"`
	UserID           string `json:"userId,omitempty"`
	UserName         string `json:"userName,omitempty"`
}
type Branch struct {
	IsPrimary bool          `json:"isPrimary"`
	Issues    []interface{} `json:"issues"`
	URL       string        `json:"url"`
}
