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

// GetPipeline & SearchPipelines
type Pipeline struct {
	Class string `json:"_class,omitempty"`
	Links struct {
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
		Scm struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"scm,omitempty"`
		Branches struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"branches,omitempty"`
		Actions struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"actions,omitempty"`
		Runs struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"runs,omitempty"`
		Trends struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"trends,omitempty"`
		Queue struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"queue,omitempty"`
	} `json:"_links,omitempty"`
	Actions         []interface{} `json:"actions,omitempty"`
	Disabled        interface{}   `json:"disabled,omitempty"`
	DisplayName     string        `json:"displayName,omitempty"`
	FullDisplayName string        `json:"fullDisplayName,omitempty"`
	FullName        string        `json:"fullName,omitempty"`
	Name            string        `json:"name,omitempty"`
	Organization    string        `json:"organization,omitempty"`
	Parameters      interface{}   `json:"parameters,omitempty"`
	Permissions     struct {
		Create    bool `json:"create,omitempty"`
		Configure bool `json:"configure,omitempty"`
		Read      bool `json:"read,omitempty"`
		Start     bool `json:"start,omitempty"`
		Stop      bool `json:"stop,omitempty"`
	} `json:"permissions,omitempty"`
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
	ScmSource                      struct {
		Class  string      `json:"_class,omitempty"`
		APIURL interface{} `json:"apiUrl,omitempty"`
		ID     string      `json:"id,omitempty"`
	} `json:"scmSource,omitempty"`
	TotalNumberOfBranches     int `json:"totalNumberOfBranches,omitempty"`
	TotalNumberOfPullRequests int `json:"totalNumberOfPullRequests,omitempty"`
}

// GetPipeBranchRun & SearchPipelineRuns
type BranchPipelineRun struct {
	Class string `json:"_class,omitempty"`
	Links struct {
		PrevRun struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"prevRun,omitempty"`
		Parent struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"parent,omitempty"`
		Tests struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"tests,omitempty"`
		Nodes struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"nodes,omitempty"`
		Log struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"log,omitempty"`
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
		BlueTestSummary struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"blueTestSummary,omitempty"`
		Actions struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"actions,omitempty"`
		Steps struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"steps,omitempty"`
		Artifacts struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"artifacts,omitempty"`
		NextRun struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"nextRun,omitempty"`
	} `json:"_links,omitempty"`
	Actions          []interface{} `json:"actions,omitempty"`
	ArtifactsZipFile interface{}   `json:"artifactsZipFile,omitempty"`
	CauseOfBlockage  interface{}   `json:"causeOfBlockage,omitempty"`
	Causes           []struct {
		Class            string `json:"_class,omitempty"`
		ShortDescription string `json:"shortDescription,omitempty"`
		UserID           string `json:"userId,omitempty"`
		UserName         string `json:"userName,omitempty"`
	} `json:"causes,omitempty"`
	ChangeSet                 []interface{} `json:"changeSet,omitempty"`
	Description               interface{}   `json:"description,omitempty"`
	DurationInMillis          int           `json:"durationInMillis,omitempty"`
	EnQueueTime               string        `json:"enQueueTime,omitempty"`
	EndTime                   string        `json:"endTime,omitempty"`
	EstimatedDurationInMillis int           `json:"estimatedDurationInMillis,omitempty"`
	ID                        string        `json:"id,omitempty"`
	Name                      interface{}   `json:"name,omitempty"`
	Organization              string        `json:"organization,omitempty"`
	Pipeline                  string        `json:"pipeline,omitempty"`
	Replayable                bool          `json:"replayable,omitempty"`
	Result                    string        `json:"result,omitempty"`
	RunSummary                string        `json:"runSummary,omitempty"`
	StartTime                 string        `json:"startTime,omitempty"`
	State                     string        `json:"state,omitempty"`
	Type                      string        `json:"type,omitempty"`
	Branch                    struct {
		IsPrimary bool          `json:"isPrimary,omitempty"`
		Issues    []interface{} `json:"issues,omitempty"`
		URL       string        `json:"url,omitempty"`
	} `json:"branch,omitempty"`
	CommitID    string      `json:"commitId,omitempty"`
	CommitURL   interface{} `json:"commitUrl,omitempty"`
	PullRequest interface{} `json:"pullRequest,omitempty"`
}

// GetBranchPipeRunNodes
type BranchPipelineRunNodes struct {
	Class string `json:"_class,omitempty"`
	Links struct {
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
		Actions struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"actions,omitempty"`
		Steps struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"steps,omitempty"`
	} `json:"_links,omitempty"`
	Actions            []interface{} `json:"actions,omitempty"`
	DisplayDescription interface{}   `json:"displayDescription,omitempty"`
	DisplayName        string        `json:"displayName,omitempty"`
	DurationInMillis   int           `json:"durationInMillis,omitempty"`
	ID                 string        `json:"id,omitempty"`
	Input              interface{}   `json:"input,omitempty"`
	Result             string        `json:"result,omitempty"`
	StartTime          string        `json:"startTime,omitempty"`
	State              string        `json:"state,omitempty"`
	Type               string        `json:"type,omitempty"`
	CauseOfBlockage    interface{}   `json:"causeOfBlockage,omitempty"`
	Edges              []struct {
		Class string `json:"_class,omitempty"`
		ID    string `json:"id,omitempty"`
		Type  string `json:"type,omitempty"`
	} `json:"edges,omitempty"`
	FirstParent interface{} `json:"firstParent,omitempty"`
	Restartable bool        `json:"restartable,omitempty"`
	Steps       []struct {
		Class string `json:"_class,omitempty"`
		Links struct {
			Self struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"self,omitempty"`
			Actions struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"actions,omitempty"`
		} `json:"_links,omitempty"`
		Actions []struct {
			Class string `json:"_class,omitempty"`
			Links struct {
				Self struct {
					Class string `json:"_class,omitempty"`
					Href  string `json:"href,omitempty"`
				} `json:"self,omitempty"`
			} `json:"_links,omitempty"`
			URLName string `json:"urlName,omitempty"`
		} `json:"actions,omitempty"`
		DisplayDescription interface{} `json:"displayDescription,omitempty"`
		DisplayName        string      `json:"displayName,omitempty"`
		DurationInMillis   int         `json:"durationInMillis,omitempty"`
		ID                 string      `json:"id,omitempty"`
		Input              interface{} `json:"input,omitempty"`
		Result             string      `json:"result,omitempty"`
		StartTime          string      `json:"startTime,omitempty"`
		State              string      `json:"state,omitempty"`
		Type               string      `json:"type,omitempty"`
	} `json:"steps,omitempty"`
}

// Validate
type Validates struct {
	CredentialID string `json:"credentialId,omitempty"`
}

// GetSCMOrg
type SCMOrg struct {
	Class string `json:"_class,omitempty"`
	Links struct {
		Repositories struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"repositories,omitempty"`
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
	} `json:"_links,omitempty"`
	Avatar                      string `json:"avatar,omitempty"`
	JenkinsOrganizationPipeline bool   `json:"jenkinsOrganizationPipeline,omitempty"`
	Name                        string `json:"name,omitempty"`
}

// GetOrgRepo
type OrgRepo struct {
	Class string `json:"_class,omitempty"`
	Links struct {
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
	} `json:"_links,omitempty"`
	Repositories struct {
		Class string `json:"_class,omitempty"`
		Links struct {
			Self struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"self,omitempty"`
		} `json:"_links,omitempty"`
		Items []struct {
			Class string `json:"_class,omitempty"`
			Links struct {
				Self struct {
					Class string `json:"_class,omitempty"`
					Href  string `json:"href,omitempty"`
				} `json:"self,omitempty"`
			} `json:"_links,omitempty"`
			DefaultBranch string `json:"defaultBranch,omitempty"`
			Description   string `json:"description,omitempty"`
			Name          string `json:"name,omitempty"`
			Permissions   struct {
				Admin bool `json:"admin,omitempty"`
				Push  bool `json:"push,omitempty"`
				Pull  bool `json:"pull,omitempty"`
			} `json:"permissions,omitempty"`
			Private  bool   `json:"private,omitempty"`
			FullName string `json:"fullName,omitempty"`
		} `json:"items,omitempty"`
		LastPage interface{} `json:"lastPage,omitempty"`
		NextPage interface{} `json:"nextPage,omitempty"`
		PageSize int         `json:"pageSize,omitempty"`
	} `json:"repositories,omitempty"`
}

// StopPipeline
type StopPipe struct {
	Class string `json:"_class,omitempty"`
	Links struct {
		Parent struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"parent,omitempty"`
		Tests struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"tests,omitempty"`
		Nodes struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"nodes,omitempty"`
		Log struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"log,omitempty"`
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
		BlueTestSummary struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"blueTestSummary,omitempty"`
		Actions struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"actions,omitempty"`
		Steps struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"steps,omitempty"`
		Artifacts struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"artifacts,omitempty"`
	} `json:"_links,omitempty"`
	Actions          []interface{} `json:"actions,omitempty"`
	ArtifactsZipFile interface{}   `json:"artifactsZipFile,omitempty"`
	CauseOfBlockage  interface{}   `json:"causeOfBlockage,omitempty"`
	Causes           []struct {
		Class            string `json:"_class,omitempty"`
		ShortDescription string `json:"shortDescription,omitempty"`
	} `json:"causes,omitempty"`
	ChangeSet                 []interface{} `json:"changeSet,omitempty"`
	Description               interface{}   `json:"description,omitempty"`
	DurationInMillis          int           `json:"durationInMillis,omitempty"`
	EnQueueTime               string        `json:"enQueueTime,omitempty"`
	EndTime                   string        `json:"endTime,omitempty"`
	EstimatedDurationInMillis int           `json:"estimatedDurationInMillis,omitempty"`
	ID                        string        `json:"id,omitempty"`
	Name                      interface{}   `json:"name,omitempty"`
	Organization              string        `json:"organization,omitempty"`
	Pipeline                  string        `json:"pipeline,omitempty"`
	Replayable                bool          `json:"replayable,omitempty"`
	Result                    string        `json:"result,omitempty"`
	RunSummary                string        `json:"runSummary,omitempty"`
	StartTime                 string        `json:"startTime,omitempty"`
	State                     string        `json:"state,omitempty"`
	Type                      string        `json:"type,omitempty"`
	Branch                    struct {
		IsPrimary bool          `json:"isPrimary,omitempty"`
		Issues    []interface{} `json:"issues,omitempty"`
		URL       string        `json:"url,omitempty"`
	} `json:"branch,omitempty"`
	CommitID    string      `json:"commitId,omitempty"`
	CommitURL   interface{} `json:"commitUrl,omitempty"`
	PullRequest interface{} `json:"pullRequest,omitempty"`
}

// ReplayPipeline
type ReplayPipe struct {
	Class string `json:"_class,omitempty"`
	Links struct {
		Parent struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"parent,omitempty"`
		Tests struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"tests,omitempty"`
		Log struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"log,omitempty"`
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
		BlueTestSummary struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"blueTestSummary,omitempty"`
		Actions struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"actions,omitempty"`
		Artifacts struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"artifacts,omitempty"`
	} `json:"_links,omitempty"`
	Actions          []interface{} `json:"actions,omitempty"`
	ArtifactsZipFile interface{}   `json:"artifactsZipFile,omitempty"`
	CauseOfBlockage  string        `json:"causeOfBlockage,omitempty"`
	Causes           []struct {
		Class            string `json:"_class,omitempty"`
		ShortDescription string `json:"shortDescription,omitempty"`
		UserID           string `json:"userId,omitempty"`
		UserName         string `json:"userName,omitempty"`
	} `json:"causes,omitempty"`
	ChangeSet                 []interface{} `json:"changeSet,omitempty"`
	Description               interface{}   `json:"description,omitempty"`
	DurationInMillis          interface{}   `json:"durationInMillis,omitempty"`
	EnQueueTime               interface{}   `json:"enQueueTime,omitempty"`
	EndTime                   interface{}   `json:"endTime,omitempty"`
	EstimatedDurationInMillis interface{}   `json:"estimatedDurationInMillis,omitempty"`
	ID                        string        `json:"id,omitempty"`
	Name                      interface{}   `json:"name,omitempty"`
	Organization              string        `json:"organization,omitempty"`
	Pipeline                  string        `json:"pipeline,omitempty"`
	Replayable                bool          `json:"replayable,omitempty"`
	Result                    string        `json:"result,omitempty"`
	RunSummary                interface{}   `json:"runSummary,omitempty"`
	StartTime                 interface{}   `json:"startTime,omitempty"`
	State                     string        `json:"state,omitempty"`
	Type                      string        `json:"type,omitempty"`
	QueueID                   string        `json:"queueId,omitempty"`
}

// GetArtifacts
type Artifacts struct {
	Class string `json:"_class,omitempty"`
	Links struct {
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
	} `json:"_links,omitempty"`
	Downloadable bool   `json:"downloadable,omitempty"`
	ID           string `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	Path         string `json:"path,omitempty"`
	Size         int    `json:"size,omitempty"`
	URL          string `json:"url,omitempty"` // The url for Download artifacts
}

// GetPipeBranch
type PipeBranch struct {
	Class string `json:"_class,omitempty"`
	Links struct {
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
		Scm struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"scm,omitempty"`
		Actions struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"actions,omitempty"`
		Runs struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"runs,omitempty"`
		Trends struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"trends,omitempty"`
		Queue struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"queue,omitempty"`
	} `json:"_links,omitempty"`
	Actions                   []interface{} `json:"actions,omitempty"`
	Disabled                  bool          `json:"disabled,omitempty"`
	DisplayName               string        `json:"displayName,omitempty"`
	EstimatedDurationInMillis int           `json:"estimatedDurationInMillis,omitempty"`
	FullDisplayName           string        `json:"fullDisplayName,omitempty"`
	FullName                  string        `json:"fullName,omitempty"`
	LatestRun                 struct {
		Class string `json:"_class,omitempty"`
		Links struct {
			PrevRun struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"prevRun,omitempty"`
			Parent struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"parent,omitempty"`
			Tests struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"tests,omitempty"`
			Log struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"log,omitempty"`
			Self struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"self,omitempty"`
			BlueTestSummary struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"blueTestSummary,omitempty"`
			Actions struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"actions,omitempty"`
			Artifacts struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"artifacts,omitempty"`
		} `json:"_links,omitempty"`
		Actions          []interface{} `json:"actions,omitempty"`
		ArtifactsZipFile string        `json:"artifactsZipFile,omitempty"`
		CauseOfBlockage  interface{}   `json:"causeOfBlockage,omitempty"`
		Causes           []struct {
			Class            string `json:"_class,omitempty"`
			ShortDescription string `json:"shortDescription,omitempty"`
		} `json:"causes,omitempty"`
		ChangeSet                 []interface{} `json:"changeSet,omitempty"`
		Description               interface{}   `json:"description,omitempty"`
		DurationInMillis          int           `json:"durationInMillis,omitempty"`
		EnQueueTime               string        `json:"enQueueTime,omitempty"`
		EndTime                   string        `json:"endTime,omitempty"`
		EstimatedDurationInMillis int           `json:"estimatedDurationInMillis,omitempty"`
		ID                        string        `json:"id,omitempty"`
		Name                      interface{}   `json:"name,omitempty"`
		Organization              string        `json:"organization,omitempty"`
		Pipeline                  string        `json:"pipeline,omitempty"`
		Replayable                bool          `json:"replayable,omitempty"`
		Result                    string        `json:"result,omitempty"`
		RunSummary                string        `json:"runSummary,omitempty"`
		StartTime                 string        `json:"startTime,omitempty"`
		State                     string        `json:"state,omitempty"`
		Type                      string        `json:"type,omitempty"`
	} `json:"latestRun,omitempty"`
	Name         string `json:"name,omitempty"`
	Organization string `json:"organization,omitempty"`
	Parameters   []struct {
		Class                 string `json:"_class,omitempty"`
		DefaultParameterValue struct {
			Class string `json:"_class,omitempty"`
			Name  string `json:"name,omitempty"`
			Value string `json:"value,omitempty"`
		} `json:"defaultParameterValue,omitempty"`
		Description string `json:"description,omitempty"`
		Name        string `json:"name,omitempty"`
		Type        string `json:"type,omitempty"`
	} `json:"parameters,omitempty"`
	Permissions struct {
		Create    bool `json:"create,omitempty"`
		Configure bool `json:"configure,omitempty"`
		Read      bool `json:"read,omitempty"`
		Start     bool `json:"start,omitempty"`
		Stop      bool `json:"stop,omitempty"`
	} `json:"permissions,omitempty"`
	WeatherScore int `json:"weatherScore,omitempty"`
	Branch       struct {
		IsPrimary bool          `json:"isPrimary,omitempty"`
		Issues    []interface{} `json:"issues,omitempty"`
		URL       string        `json:"url,omitempty"`
	} `json:"branch,omitempty"`
}

// RunPipeline
type RunPayload struct {
	Parameters []struct {
		Name  string `json:"name,omitempty"`
		Value string `json:"value,omitempty"`
	} `json:"parameters,omitempty"`
}

type QueuedBlueRun struct {
	Class string `json:"_class,omitempty"`
	Links struct {
		Parent struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"parent,omitempty"`
		Tests struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"tests,omitempty"`
		Log struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"log,omitempty"`
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
		BlueTestSummary struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"blueTestSummary,omitempty"`
		Actions struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"actions,omitempty"`
		Artifacts struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"artifacts,omitempty"`
	} `json:"_links,omitempty"`
	Actions          []interface{} `json:"actions,omitempty"`
	ArtifactsZipFile interface{}   `json:"artifactsZipFile,omitempty"`
	CauseOfBlockage  string        `json:"causeOfBlockage,omitempty"`
	Causes           []struct {
		Class            string `json:"_class,omitempty"`
		ShortDescription string `json:"shortDescription,omitempty"`
		UserID           string `json:"userId,omitempty"`
		UserName         string `json:"userName,omitempty"`
	} `json:"causes,omitempty"`
	ChangeSet                 []interface{} `json:"changeSet,omitempty"`
	Description               interface{}   `json:"description,omitempty"`
	DurationInMillis          interface{}   `json:"durationInMillis,omitempty"`
	EnQueueTime               interface{}   `json:"enQueueTime,omitempty"`
	EndTime                   interface{}   `json:"endTime,omitempty"`
	EstimatedDurationInMillis interface{}   `json:"estimatedDurationInMillis,omitempty"`
	ID                        string        `json:"id,omitempty"`
	Name                      interface{}   `json:"name,omitempty"`
	Organization              string        `json:"organization,omitempty"`
	Pipeline                  string        `json:"pipeline,omitempty"`
	Replayable                bool          `json:"replayable,omitempty"`
	Result                    string        `json:"result,omitempty"`
	RunSummary                interface{}   `json:"runSummary,omitempty"`
	StartTime                 interface{}   `json:"startTime,omitempty"`
	State                     string        `json:"state,omitempty"`
	Type                      string        `json:"type,omitempty"`
	QueueID                   string        `json:"queueId,omitempty"`
}

// GetNodeStatus
type NodeStatus struct {
	Class string `json:"_class,omitempty"`
	Links struct {
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
		Actions struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"actions,omitempty"`
		Steps struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"steps,omitempty"`
	} `json:"_links,omitempty"`
	Actions            []interface{} `json:"actions,omitempty"`
	DisplayDescription interface{}   `json:"displayDescription,omitempty"`
	DisplayName        string        `json:"displayName,omitempty"`
	DurationInMillis   int           `json:"durationInMillis,omitempty"`
	ID                 string        `json:"id,omitempty"`
	Input              interface{}   `json:"input,omitempty"`
	Result             string        `json:"result,omitempty"`
	StartTime          string        `json:"startTime,omitempty"`
	State              string        `json:"state,omitempty"`
	Type               string        `json:"type,omitempty"`
	CauseOfBlockage    interface{}   `json:"causeOfBlockage,omitempty"`
	Edges              []struct {
		Class string `json:"_class,omitempty"`
		ID    string `json:"id,omitempty"`
		Type  string `json:"type,omitempty"`
	} `json:"edges,omitempty"`
	FirstParent interface{} `json:"firstParent,omitempty"`
	Restartable bool        `json:"restartable,omitempty"`
	Steps       []struct {
		Class string `json:"_class,omitempty"`
		Links struct {
			Self struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"self,omitempty"`
			Actions struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"actions,omitempty"`
		} `json:"_links,omitempty"`
		Actions []struct {
			Class string `json:"_class,omitempty"`
			Links struct {
				Self struct {
					Class string `json:"_class,omitempty"`
					Href  string `json:"href,omitempty"`
				} `json:"self,omitempty"`
			} `json:"_links,omitempty"`
			URLName string `json:"urlName,omitempty"`
		} `json:"actions,omitempty"`
		DisplayDescription interface{} `json:"displayDescription,omitempty"`
		DisplayName        string      `json:"displayName,omitempty"`
		DurationInMillis   int         `json:"durationInMillis,omitempty"`
		ID                 string      `json:"id,omitempty"`
		Input              interface{} `json:"input,omitempty"`
		Result             string      `json:"result,omitempty"`
		StartTime          string      `json:"startTime,omitempty"`
		State              string      `json:"state,omitempty"`
		Type               string      `json:"type,omitempty"`
	} `json:"steps,omitempty"`
}

// CheckPipeline
type CheckPlayload struct {
	ID         string `json:"id,omitempty"`
	Parameters []struct {
		Name  string `json:"name,omitempty"`
		Value string `json:"value,omitempty"`
	} `json:"parameters,omitempty"`
	Abort bool `json:"abort,omitempty"`
}

// Getcrumb
type Crumb struct {
	Class             string `json:"_class,omitempty"`
	Crumb             string `json:"crumb,omitempty"`
	CrumbRequestField string `json:"crumbRequestField,omitempty"`
}

// CheckScriptCompile
type CheckScript struct {
	Column  int    `json:"column,omitempty"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message,omitempty"`
	Status  string `json:"status,omitempty"`
}

// CheckCron
type CheckCronRes struct {
	Result  string `json:"result,omitempty"`
	Message string `json:"message,omitempty"`
}

// GetPipelineRun
type PipelineRun struct {
	Class string `json:"_class,omitempty"`
	Links struct {
		PrevRun struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"prevRun,omitempty"`
		Parent struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"parent,omitempty"`
		Tests struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"tests,omitempty"`
		Nodes struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"nodes,omitempty"`
		Log struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"log,omitempty"`
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
		BlueTestSummary struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"blueTestSummary,omitempty"`
		Actions struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"actions,omitempty"`
		Steps struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"steps,omitempty"`
		Artifacts struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"artifacts,omitempty"`
	} `json:"_links,omitempty"`
	Actions          []interface{} `json:"actions,omitempty"`
	ArtifactsZipFile interface{}   `json:"artifactsZipFile,omitempty"`
	CauseOfBlockage  interface{}   `json:"causeOfBlockage,omitempty"`
	Causes           []struct {
		Class            string `json:"_class,omitempty"`
		ShortDescription string `json:"shortDescription,omitempty"`
		UserID           string `json:"userId,omitempty"`
		UserName         string `json:"userName,omitempty"`
	} `json:"causes,omitempty"`
	ChangeSet                 []interface{} `json:"changeSet,omitempty"`
	Description               interface{}   `json:"description,omitempty"`
	DurationInMillis          int           `json:"durationInMillis,omitempty"`
	EnQueueTime               string        `json:"enQueueTime,omitempty"`
	EndTime                   string        `json:"endTime,omitempty"`
	EstimatedDurationInMillis int           `json:"estimatedDurationInMillis,omitempty"`
	ID                        string        `json:"id,omitempty"`
	Name                      interface{}   `json:"name,omitempty"`
	Organization              string        `json:"organization,omitempty"`
	Pipeline                  string        `json:"pipeline,omitempty"`
	Replayable                bool          `json:"replayable,omitempty"`
	Result                    string        `json:"result,omitempty"`
	RunSummary                string        `json:"runSummary,omitempty"`
	StartTime                 string        `json:"startTime,omitempty"`
	State                     string        `json:"state,omitempty"`
	Type                      string        `json:"type,omitempty"`
	Branch                    interface{}   `json:"branch,omitempty"`
	CommitID                  interface{}   `json:"commitId,omitempty"`
	CommitURL                 interface{}   `json:"commitUrl,omitempty"`
	PullRequest               interface{}   `json:"pullRequest,omitempty"`
}

// GetBranchPipeRun
type BranchPipeline struct {
	Class string `json:"_class,omitempty"`
	Links struct {
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
		Scm struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"scm,omitempty"`
		Actions struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"actions,omitempty"`
		Runs struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"runs,omitempty"`
		Trends struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"trends,omitempty"`
		Queue struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"queue,omitempty"`
	} `json:"_links,omitempty"`
	Actions                   []interface{} `json:"actions,omitempty"`
	Disabled                  bool          `json:"disabled,omitempty"`
	DisplayName               string        `json:"displayName,omitempty"`
	EstimatedDurationInMillis int           `json:"estimatedDurationInMillis,omitempty"`
	FullDisplayName           string        `json:"fullDisplayName,omitempty"`
	FullName                  string        `json:"fullName,omitempty"`
	LatestRun                 struct {
		Class string `json:"_class,omitempty"`
		Links struct {
			PrevRun struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"prevRun,omitempty"`
			Parent struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"parent,omitempty"`
			Tests struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"tests,omitempty"`
			Log struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"log,omitempty"`
			Self struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"self,omitempty"`
			BlueTestSummary struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"blueTestSummary,omitempty"`
			Actions struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"actions,omitempty"`
			Artifacts struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"artifacts,omitempty"`
		} `json:"_links,omitempty"`
		Actions          []interface{} `json:"actions,omitempty"`
		ArtifactsZipFile string        `json:"artifactsZipFile,omitempty"`
		CauseOfBlockage  interface{}   `json:"causeOfBlockage,omitempty"`
		Causes           []struct {
			Class            string `json:"_class,omitempty"`
			ShortDescription string `json:"shortDescription,omitempty"`
			UserID           string `json:"userId,omitempty"`
			UserName         string `json:"userName,omitempty"`
		} `json:"causes,omitempty"`
		ChangeSet                 []interface{} `json:"changeSet,omitempty"`
		Description               interface{}   `json:"description,omitempty"`
		DurationInMillis          int           `json:"durationInMillis,omitempty"`
		EnQueueTime               string        `json:"enQueueTime,omitempty"`
		EndTime                   string        `json:"endTime,omitempty"`
		EstimatedDurationInMillis int           `json:"estimatedDurationInMillis,omitempty"`
		ID                        string        `json:"id,omitempty"`
		Name                      interface{}   `json:"name,omitempty"`
		Organization              string        `json:"organization,omitempty"`
		Pipeline                  string        `json:"pipeline,omitempty"`
		Replayable                bool          `json:"replayable,omitempty"`
		Result                    string        `json:"result,omitempty"`
		RunSummary                string        `json:"runSummary,omitempty"`
		StartTime                 string        `json:"startTime,omitempty"`
		State                     string        `json:"state,omitempty"`
		Type                      string        `json:"type,omitempty"`
	} `json:"latestRun,omitempty"`
	Name         string `json:"name,omitempty"`
	Organization string `json:"organization,omitempty"`
	Parameters   []struct {
		Class                 string `json:"_class,omitempty"`
		DefaultParameterValue struct {
			Class string `json:"_class,omitempty"`
			Name  string `json:"name,omitempty"`
			Value string `json:"value,omitempty"`
		} `json:"defaultParameterValue,omitempty"`
		Description string `json:"description,omitempty"`
		Name        string `json:"name,omitempty"`
		Type        string `json:"type,omitempty"`
	} `json:"parameters,omitempty"`
	Permissions struct {
		Create    bool `json:"create,omitempty"`
		Configure bool `json:"configure,omitempty"`
		Read      bool `json:"read,omitempty"`
		Start     bool `json:"start,omitempty"`
		Stop      bool `json:"stop,omitempty"`
	} `json:"permissions,omitempty"`
	WeatherScore int `json:"weatherScore,omitempty"`
	Branch       struct {
		IsPrimary bool          `json:"isPrimary,omitempty"`
		Issues    []interface{} `json:"issues,omitempty"`
		URL       string        `json:"url,omitempty"`
	} `json:"branch,omitempty"`
}

// GetPipelineRunNodes
type PipelineRunNodes struct {
	Class string `json:"_class,omitempty"`
	Links struct {
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
		Actions struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"actions,omitempty"`
		Steps struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"steps,omitempty"`
	} `json:"_links,omitempty"`
	Actions            []interface{} `json:"actions,omitempty"`
	DisplayDescription interface{}   `json:"displayDescription,omitempty"`
	DisplayName        string        `json:"displayName,omitempty"`
	DurationInMillis   int           `json:"durationInMillis,omitempty"`
	ID                 string        `json:"id,omitempty"`
	Input              interface{}   `json:"input,omitempty"`
	Result             string        `json:"result,omitempty"`
	StartTime          string        `json:"startTime,omitempty"`
	State              string        `json:"state,omitempty"`
	Type               string        `json:"type,omitempty"`
	CauseOfBlockage    interface{}   `json:"causeOfBlockage,omitempty"`
	Edges              []interface{} `json:"edges,omitempty"`
	FirstParent        interface{}   `json:"firstParent,omitempty"`
	Restartable        bool          `json:"restartable,omitempty"`
}

// GetNodeSteps
type NodeSteps struct {
	Class string `json:"_class,omitempty"`
	Links struct {
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
		Actions struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"actions,omitempty"`
	} `json:"_links,omitempty"`
	Actions []struct {
		Class string `json:"_class,omitempty"`
		Links struct {
			Self struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"self,omitempty"`
		} `json:"_links,omitempty"`
		URLName string `json:"urlName,omitempty"`
	} `json:"actions,omitempty"`
	DisplayDescription string      `json:"displayDescription,omitempty"`
	DisplayName        string      `json:"displayName,omitempty"`
	DurationInMillis   int         `json:"durationInMillis,omitempty"`
	ID                 string      `json:"id,omitempty"`
	Input              interface{} `json:"input,omitempty"`
	Result             string      `json:"result,omitempty"`
	StartTime          string      `json:"startTime,omitempty"`
	State              string      `json:"state,omitempty"`
	Type               string      `json:"type,omitempty"`
}

// ToJenkinsfile requests
type ReqJson struct {
	Json string `json:"json,omitempty"`
}

// ToJenkinsfile response
type ResJenkinsfile struct {
	Status string `json:"status,omitempty"`
	Data   struct {
		Result      string `json:"result,omitempty"`
		Jenkinsfile string `json:"jenkinsfile,omitempty"`
		Errors      []struct {
			Location []string `json:"location,omitempty"`
			Error    string   `json:"error,omitempty"`
		} `json:"errors,omitempty"`
	} `json:"data,omitempty"`
}

type ReqJenkinsfile struct {
	Jenkinsfile string `json:"jenkinsfile,omitempty"`
}

type ResJson struct {
	Status string `json:"status,omitempty"`
	Data   struct {
		Result string `json:"result,omitempty"`
		JSON   struct {
			Pipeline struct {
				Stages []interface{} `json:"stages,omitempty"`
				Agent  struct {
					Type      string `json:"type,omitempty"`
					Arguments []struct {
						Key   string `json:"key,omitempty"`
						Value struct {
							IsLiteral bool   `json:"isLiteral,omitempty"`
							Value     string `json:"value,omitempty"`
						} `json:"value,omitempty"`
					} `json:"arguments,omitempty"`
				} `json:"agent,omitempty"`
			} `json:"pipeline,omitempty"`
		} `json:"json,omitempty"`
	} `json:"data,omitempty"`
}

type NodesDetail struct {
	Class string `json:"_class,omitempty"`
	Links struct {
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
		Actions struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"actions,omitempty"`
		Steps struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"steps,omitempty"`
	} `json:"_links,omitempty"`
	Actions            []interface{} `json:"actions,omitempty"`
	DisplayDescription interface{}   `json:"displayDescription,omitempty"`
	DisplayName        string        `json:"displayName,omitempty"`
	DurationInMillis   int           `json:"durationInMillis,omitempty"`
	ID                 string        `json:"id,omitempty"`
	Input              interface{}   `json:"input,omitempty"`
	Result             string        `json:"result,omitempty"`
	StartTime          string        `json:"startTime,omitempty"`
	State              string        `json:"state,omitempty"`
	Type               string        `json:"type,omitempty"`
	CauseOfBlockage    interface{}   `json:"causeOfBlockage,omitempty"`
	Edges              []struct {
		Class string `json:"_class,omitempty"`
		ID    string `json:"id,omitempty"`
		Type  string `json:"type,omitempty"`
	} `json:"edges,omitempty"`
	FirstParent interface{} `json:"firstParent,omitempty"`
	Restartable bool        `json:"restartable,omitempty"`
	Steps       []NodeSteps `json:"steps,omitempty"`
}

type NodesStepsIndex struct {
	Id    int         `json:"id,omitempty"`
	Steps []NodeSteps `json:"steps,omitempty"`
}
