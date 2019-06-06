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
	Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability." `
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
	} `json:"_links,omitempty" description:"references the reachable path to this resource."`
	Actions         []interface{} `json:"actions,omitempty" description:"the list of actions."`
	Disabled        interface{}   `json:"disabled,omitempty" description:"disable or not"`
	DisplayName     string        `json:"displayName,omitempty" description:"display name"`
	FullDisplayName string        `json:"fullDisplayName,omitempty" description:"full display name"`
	FullName        string        `json:"fullName,omitempty" description:"full name"`
	Name            string        `json:"name,omitempty" description:"name"`
	Organization    string        `json:"organization,omitempty" description:"organization name"`
	Parameters      interface{}   `json:"parameters,omitempty" description:"parameters of pipeline"`
	Permissions     struct {
		Create    bool `json:"create,omitempty" description:"create action"`
		Configure bool `json:"configure,omitempty" description:"configure action"`
		Read      bool `json:"read,omitempty" description:"read action"`
		Start     bool `json:"start,omitempty" description:"start action"`
		Stop      bool `json:"stop,omitempty" description:"stop action"`
	} `json:"permissions,omitempty" description:"permissions"`
	EstimatedDurationInMillis      int           `json:"estimatedDurationInMillis,omitempty" description:"estimated duration time, unit is millis"`
	NumberOfFolders                int           `json:"numberOfFolders,omitempty" description:"number of folders"`
	NumberOfPipelines              int           `json:"numberOfPipelines,omitempty" description:"number of pipelines"`
	PipelineFolderNames            []interface{} `json:"pipelineFolderNames,omitempty" description:"pipeline folder names"`
	WeatherScore                   int           `json:"weatherScore,omitempty" description:"the score to description the result of pipeline"`
	BranchNames                    []string      `json:"branchNames,omitempty" description:"branch names"`
	NumberOfFailingBranches        int           `json:"numberOfFailingBranches,omitempty" description:"number of failing branches"`
	NumberOfFailingPullRequests    int           `json:"numberOfFailingPullRequests,omitempty" description:"number of failing pull requests"`
	NumberOfSuccessfulBranches     int           `json:"numberOfSuccessfulBranches,omitempty" description:"number of successful pull requests"`
	NumberOfSuccessfulPullRequests int           `json:"numberOfSuccessfulPullRequests,omitempty" description:"number of successful pull requests"`
	ScmSource                      struct {
		Class  string      `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		APIURL interface{} `json:"apiUrl,omitempty" description:"api url"`
		ID     string      `json:"id,omitempty" description:"scm source id"`
	} `json:"scmSource,omitempty"`
	TotalNumberOfBranches     int `json:"totalNumberOfBranches,omitempty" description:"total number of branches"`
	TotalNumberOfPullRequests int `json:"totalNumberOfPullRequests,omitempty" description:"total number of pull requests"`
}

// GetPipeBranchRun & SearchPipelineRuns
type BranchPipelineRun struct {
	Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
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
	} `json:"_links,omitempty" description:"references the reachable path to this resource"`
	Actions          []interface{} `json:"actions,omitempty" description:"the list of actions"`
	ArtifactsZipFile interface{}   `json:"artifactsZipFile,omitempty" description:"the artifacts zip file"`
	CauseOfBlockage  interface{}   `json:"causeOfBlockage,omitempty" description:"cause of blockage"`
	Causes           []struct {
		Class            string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		ShortDescription string `json:"shortDescription,omitempty" description:"short description"`
		UserID           string `json:"userId,omitempty" description:"user id"`
		UserName         string `json:"userName,omitempty" description:"user name"`
	} `json:"causes,omitempty"`
	ChangeSet                 []interface{} `json:"changeSet,omitempty" description:"change set"`
	Description               interface{}   `json:"description,omitempty" description:"description of resource"`
	DurationInMillis          int           `json:"durationInMillis,omitempty" description:"duration time in millis"`
	EnQueueTime               string        `json:"enQueueTime,omitempty" description:"enqueue time"`
	EndTime                   string        `json:"endTime,omitempty" description:"end time"`
	EstimatedDurationInMillis int           `json:"estimatedDurationInMillis,omitempty" description:"estimated duration time in millis"`
	ID                        string        `json:"id,omitempty" description:"id"`
	Name                      interface{}   `json:"name,omitempty" description:"name"`
	Organization              string        `json:"organization,omitempty" description:"organization name"`
	Pipeline                  string        `json:"pipeline,omitempty" description:"pipeline name"`
	Replayable                bool          `json:"replayable,omitempty" description:"replayable or not"`
	Result                    string        `json:"result,omitempty" description:"result"`
	RunSummary                string        `json:"runSummary,omitempty" description:"pipeline run summary"`
	StartTime                 string        `json:"startTime,omitempty" description:"start time"`
	State                     string        `json:"state,omitempty" description:"pipeline run state"`
	Type                      string        `json:"type,omitempty" description:"source type"`
	Branch                    struct {
		IsPrimary bool          `json:"isPrimary,omitempty" description:"primary or not"`
		Issues    []interface{} `json:"issues,omitempty" description:"issues"`
		URL       string        `json:"url,omitempty" description:"url"`
	} `json:"branch,omitempty"`
	CommitID    string      `json:"commitId,omitempty" description:"commit id"`
	CommitURL   interface{} `json:"commitUrl,omitempty" description:"commit url "`
	PullRequest interface{} `json:"pullRequest,omitempty" description:"pull request"`
}

// GetBranchPipeRunNodes
type BranchPipelineRunNodes struct {
	Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
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
	} `json:"_links,omitempty" description:"references the reachable path to this resource"`
	Actions            []interface{} `json:"actions,omitempty" description:"the list of actions"`
	DisplayDescription interface{}   `json:"displayDescription,omitempty" description:"display description"`
	DisplayName        string        `json:"displayName,omitempty" description:"display name"`
	DurationInMillis   int           `json:"durationInMillis,omitempty" description:"duration time in millis"`
	ID                 string        `json:"id,omitempty" description:"id"`
	Input              interface{}   `json:"input,omitempty" description:"input"`
	Result             string        `json:"result,omitempty" description:"result"`
	StartTime          string        `json:"startTime,omitempty" description:"start time"`
	State              string        `json:"state,omitempty" description:"statue"`
	Type               string        `json:"type,omitempty" description:"source type"`
	CauseOfBlockage    interface{}   `json:"causeOfBlockage,omitempty" description:"cause of blockage"`
	Edges              []struct {
		Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		ID    string `json:"id,omitempty" description:"id"`
		Type  string `json:"type,omitempty" description:"source type"`
	} `json:"edges,omitempty"`
	FirstParent interface{} `json:"firstParent,omitempty" description:"first parent resource"`
	Restartable bool        `json:"restartable,omitempty" description:"restartable or not"`
	Steps       []struct {
		Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
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
		} `json:"actions,omitempty" description:"references the reachable path to this resource"`
		DisplayDescription interface{} `json:"displayDescription,omitempty" description:"display description"`
		DisplayName        string      `json:"displayName,omitempty" description:"display name"`
		DurationInMillis   int         `json:"durationInMillis,omitempty" description:"duration time in millis"`
		ID                 string      `json:"id,omitempty" description:"id"`
		Input              interface{} `json:"input,omitempty" description:"input"`
		Result             string      `json:"result,omitempty" description:"result"`
		StartTime          string      `json:"startTime,omitempty" description:"start time"`
		State              string      `json:"state,omitempty" description:"source state"`
		Type               string      `json:"type,omitempty" description:"source type"`
	} `json:"steps,omitempty"`
}

// Validate
type Validates struct {
	CredentialID string `json:"credentialId,omitempty" description:"credential id"`
}

// GetSCMOrg
type SCMOrg struct {
	Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
	Links struct {
		Repositories struct {
			Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
			Href  string `json:"href,omitempty" description:"url in api"`
		} `json:"repositories,omitempty"`
		Self struct {
			Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
			Href  string `json:"href,omitempty" description:"self url in api"`
		} `json:"self,omitempty" description:"scm org self info"`
	} `json:"_links,omitempty" description:"references the reachable path to this resource"`
	Avatar                      string `json:"avatar,omitempty" description:"avatar url"`
	JenkinsOrganizationPipeline bool   `json:"jenkinsOrganizationPipeline,omitempty" description:"jenkins organization pipeline"`
	Name                        string `json:"name,omitempty" description:"org name "`
}

// GetOrgRepo
type OrgRepo struct {
	Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
	Links struct {
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
	} `json:"_links,omitempty" description:"references the reachable path to this resource"`
	Repositories struct {
		Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		Links struct {
			Self struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"self,omitempty"`
		} `json:"_links,omitempty" description:"references the reachable path to this resource"`
		Items []struct {
			Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
			Links struct {
				Self struct {
					Class string `json:"_class,omitempty"`
					Href  string `json:"href,omitempty"`
				} `json:"self,omitempty"`
			} `json:"_links,omitempty" description:"references the reachable path to this resource"`
			DefaultBranch string `json:"defaultBranch,omitempty" description:"default branch"`
			Description   string `json:"description,omitempty" description:"description"`
			Name          string `json:"name,omitempty" description:"name"`
			Permissions   struct {
				Admin bool `json:"admin,omitempty" description:"admin"`
				Push  bool `json:"push,omitempty" description:"push action"`
				Pull  bool `json:"pull,omitempty" description:"pull action"`
			} `json:"permissions,omitempty"`
			Private  bool   `json:"private,omitempty" description:"private"`
			FullName string `json:"fullName,omitempty" description:"full name"`
		} `json:"items,omitempty"`
		LastPage interface{} `json:"lastPage,omitempty" description:"last page"`
		NextPage interface{} `json:"nextPage,omitempty" description:"next page"`
		PageSize int         `json:"pageSize,omitempty" description:"page size"`
	} `json:"repositories,omitempty"`
}

// StopPipeline
type StopPipe struct {
	Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
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
	} `json:"_links,omitempty" description:"references the reachable path to this resource"`
	Actions          []interface{} `json:"actions,omitempty" description:"the list of actions."`
	ArtifactsZipFile interface{}   `json:"artifactsZipFile,omitempty" description:"the artifacts zip file"`
	CauseOfBlockage  interface{}   `json:"causeOfBlockage,omitempty" description:"cause of blockage"`
	Causes           []struct {
		Class            string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		ShortDescription string `json:"shortDescription,omitempty" description:"short description"`
	} `json:"causes,omitempty"`
	ChangeSet                 []interface{} `json:"changeSet,omitempty" description:"change set"`
	Description               interface{}   `json:"description,omitempty" description:"description"`
	DurationInMillis          int           `json:"durationInMillis,omitempty" description:"duration time in millis"`
	EnQueueTime               string        `json:"enQueueTime,omitempty" description:"enqueue time"`
	EndTime                   string        `json:"endTime,omitempty" description:"end time"`
	EstimatedDurationInMillis int           `json:"estimatedDurationInMillis,omitempty" description:"estimated duration time in millis"`
	ID                        string        `json:"id,omitempty" description:"id"`
	Name                      interface{}   `json:"name,omitempty" description:"name"`
	Organization              string        `json:"organization,omitempty" description:"organization"`
	Pipeline                  string        `json:"pipeline,omitempty" description:"pipeline"`
	Replayable                bool          `json:"replayable,omitempty" description:"replayable or not"`
	Result                    string        `json:"result,omitempty" description:"result"`
	RunSummary                string        `json:"runSummary,omitempty" description:"pipeline run summary"`
	StartTime                 string        `json:"startTime,omitempty" description:"start time"`
	State                     string        `json:"state,omitempty" description:"State"`
	Type                      string        `json:"type,omitempty" description:"type"`
	Branch                    struct {
		IsPrimary bool          `json:"isPrimary,omitempty" description:"primary or not"`
		Issues    []interface{} `json:"issues,omitempty" description:"issues"`
		URL       string        `json:"url,omitempty" description:"url"`
	} `json:"branch,omitempty"`
	CommitID    string      `json:"commitId,omitempty" description:"commit id"`
	CommitURL   interface{} `json:"commitUrl,omitempty" description:"commit url"`
	PullRequest interface{} `json:"pullRequest,omitempty" description:"pull request"`
}

// ReplayPipeline
type ReplayPipe struct {
	Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
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
	} `json:"_links,omitempty" description:"references the reachable path to this resource"`
	Actions          []interface{} `json:"actions,omitempty" description:"the list of actions."`
	ArtifactsZipFile interface{}   `json:"artifactsZipFile,omitempty" description:"the artifacts zip file"`
	CauseOfBlockage  string        `json:"causeOfBlockage,omitempty" description:"cause of blockage"`
	Causes           []struct {
		Class            string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		ShortDescription string `json:"shortDescription,omitempty" description:"short description"`
		UserID           string `json:"userId,omitempty" description:"user id"`
		UserName         string `json:"userName,omitempty" description:"user name"`
	} `json:"causes,omitempty"`
	ChangeSet                 []interface{} `json:"changeSet,omitempty" description:"change set"`
	Description               interface{}   `json:"description,omitempty" description:"description"`
	DurationInMillis          interface{}   `json:"durationInMillis,omitempty" description:"duration time in millis"`
	EnQueueTime               interface{}   `json:"enQueueTime,omitempty" description:"enqueue time"`
	EndTime                   interface{}   `json:"endTime,omitempty" description:"end time"`
	EstimatedDurationInMillis interface{}   `json:"estimatedDurationInMillis,omitempty" description:"estimated duration time, unit is millis"`
	ID                        string        `json:"id,omitempty" description:"id"`
	Name                      interface{}   `json:"name,omitempty" description:"name"`
	Organization              string        `json:"organization,omitempty" description:"organization"`
	Pipeline                  string        `json:"pipeline,omitempty" description:"pipeline"`
	Replayable                bool          `json:"replayable,omitempty" description:"replayable or not"`
	Result                    string        `json:"result,omitempty" description:"result"`
	RunSummary                interface{}   `json:"runSummary,omitempty" description:"pipeline run summary"`
	StartTime                 interface{}   `json:"startTime,omitempty" description:"start time"`
	State                     string        `json:"state,omitempty" description:"state"`
	Type                      string        `json:"type,omitempty" description:"type"`
	QueueID                   string        `json:"queueId,omitempty" description:"queue id"`
}

// GetArtifacts
type Artifacts struct {
	Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
	Links struct {
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
	} `json:"_links,omitempty" description:"references the reachable path to this resource"`
	Downloadable bool   `json:"downloadable,omitempty" description:"downloadable or not"`
	ID           string `json:"id,omitempty" description:"id"`
	Name         string `json:"name,omitempty" description:"name"`
	Path         string `json:"path,omitempty" description:"path"`
	Size         int    `json:"size,omitempty" description:"size"`
	URL          string `json:"url,omitempty" description:"The url for Download artifacts"`
}

// GetPipeBranch
type PipeBranch struct {
	Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
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
	} `json:"_links,omitempty" description:"references the reachable path to this resource"`
	Actions                   []interface{} `json:"actions,omitempty" description:"the list of actions."`
	Disabled                  bool          `json:"disabled,omitempty" description:"disable or not"`
	DisplayName               string        `json:"displayName,omitempty" description:"display name"`
	EstimatedDurationInMillis int           `json:"estimatedDurationInMillis,omitempty" description:"estimated duration time, unit is millis"`
	FullDisplayName           string        `json:"fullDisplayName,omitempty" description:"full display name"`
	FullName                  string        `json:"fullName,omitempty" description:"full name"`
	LatestRun                 struct {
		Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
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
		} `json:"_links,omitempty" description:"references the reachable path to this resource"`
		Actions          []interface{} `json:"actions,omitempty" description:"the list of actions"`
		ArtifactsZipFile string        `json:"artifactsZipFile,omitempty" description:"the artifacts zip file"`
		CauseOfBlockage  interface{}   `json:"causeOfBlockage,omitempty" description:"cause of blockage"`
		Causes           []struct {
			Class            string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
			ShortDescription string `json:"shortDescription,omitempty" description:"short description"`
		} `json:"causes,omitempty"`
		ChangeSet                 []interface{} `json:"changeSet,omitempty" description:"change set"`
		Description               interface{}   `json:"description,omitempty" description:"description"`
		DurationInMillis          int           `json:"durationInMillis,omitempty" description:"duration time in millis"`
		EnQueueTime               string        `json:"enQueueTime,omitempty" description:"enqueue time"`
		EndTime                   string        `json:"endTime,omitempty" description:"end time"`
		EstimatedDurationInMillis int           `json:"estimatedDurationInMillis,omitempty" description:"estimated duration time in millis"`
		ID                        string        `json:"id,omitempty" description:"id"`
		Name                      interface{}   `json:"name,omitempty" description:"name"`
		Organization              string        `json:"organization,omitempty" description:"organization"`
		Pipeline                  string        `json:"pipeline,omitempty" description:"pipeline"`
		Replayable                bool          `json:"replayable,omitempty" description:"replayable or not"`
		Result                    string        `json:"result,omitempty" description:"result"`
		RunSummary                string        `json:"runSummary,omitempty" description:"pipeline run summary"`
		StartTime                 string        `json:"startTime,omitempty" description:"start run"`
		State                     string        `json:"state,omitempty" description:"state"`
		Type                      string        `json:"type,omitempty" description:"type"`
	} `json:"latestRun,omitempty"`
	Name         string `json:"name,omitempty" description:"name"`
	Organization string `json:"organization,omitempty" description:"organization"`
	Parameters   []struct {
		Class                 string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		DefaultParameterValue struct {
			Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
			Name  string `json:"name,omitempty" description:"name"`
			Value string `json:"value,omitempty" description:"value"`
		} `json:"defaultParameterValue,omitempty"`
		Description string `json:"description,omitempty" description:"description"`
		Name        string `json:"name,omitempty" description:"name"`
		Type        string `json:"type,omitempty" description:"type"`
	} `json:"parameters,omitempty"`
	Permissions struct {
		Create    bool `json:"create,omitempty" description:"create action"`
		Configure bool `json:"configure,omitempty" description:"configure action"`
		Read      bool `json:"read,omitempty" description:"read action"`
		Start     bool `json:"start,omitempty" description:"start action"`
		Stop      bool `json:"stop,omitempty" description:"stop action"`
	} `json:"permissions,omitempty"`
	WeatherScore int `json:"weatherScore,omitempty" description:"the score to description the result of pipeline"`
	Branch       struct {
		IsPrimary bool          `json:"isPrimary,omitempty" description:"primary or not"`
		Issues    []interface{} `json:"issues,omitempty" description:"issues"`
		URL       string        `json:"url,omitempty" description:"url"`
	} `json:"branch,omitempty"`
}

// RunPipeline
type RunPayload struct {
	Parameters []struct {
		Name  string `json:"name,omitempty" description:"name"`
		Value string `json:"value,omitempty" description:"value"`
	} `json:"parameters,omitempty"`
}

type QueuedBlueRun struct {
	Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
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
	} `json:"_links,omitempty" description:"references the reachable path to this resource"`
	Actions          []interface{} `json:"actions,omitempty" description:"the list of actions"`
	ArtifactsZipFile interface{}   `json:"artifactsZipFile,omitempty" description:"the artifacts zip file"`
	CauseOfBlockage  string        `json:"causeOfBlockage,omitempty" description:"cause of blockage"`
	Causes           []struct {
		Class            string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		ShortDescription string `json:"shortDescription,omitempty" description:"short description"`
		UserID           string `json:"userId,omitempty" description:"user id"`
		UserName         string `json:"userName,omitempty" description:"user name"`
	} `json:"causes,omitempty"`
	ChangeSet                 []interface{} `json:"changeSet,omitempty" description:"change set"`
	Description               interface{}   `json:"description,omitempty" description:"description"`
	DurationInMillis          interface{}   `json:"durationInMillis,omitempty" description:"duration time in millis"`
	EnQueueTime               interface{}   `json:"enQueueTime,omitempty" description:"enqueue time"`
	EndTime                   interface{}   `json:"endTime,omitempty" description:"end time"`
	EstimatedDurationInMillis interface{}   `json:"estimatedDurationInMillis,omitempty" description:"estimated duration time in millis"`
	ID                        string        `json:"id,omitempty" description:"id"`
	Name                      interface{}   `json:"name,omitempty" description:"name"`
	Organization              string        `json:"organization,omitempty" description:"organization"`
	Pipeline                  string        `json:"pipeline,omitempty" description:"pipeline"`
	Replayable                bool          `json:"replayable,omitempty" description:"replayable or not"`
	Result                    string        `json:"result,omitempty" description:"result"`
	RunSummary                interface{}   `json:"runSummary,omitempty" description:"pipeline run summary"`
	StartTime                 interface{}   `json:"startTime,omitempty" description:"start time"`
	State                     string        `json:"state,omitempty" description:"state"`
	Type                      string        `json:"type,omitempty" description:"type"`
	QueueID                   string        `json:"queueId,omitempty" description:"queue id"`
}

// GetNodeStatus
type NodeStatus struct {
	Class string `json:"_class,omitempty" description:""`
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
	} `json:"_links,omitempty" description:"references the reachable path to this resource"`
	Actions            []interface{} `json:"actions,omitempty" description:"the list of actions"`
	DisplayDescription interface{}   `json:"displayDescription,omitempty" description:"display description"`
	DisplayName        string        `json:"displayName,omitempty" description:"display name"`
	DurationInMillis   int           `json:"durationInMillis,omitempty" description:"duration time in millis"`
	ID                 string        `json:"id,omitempty" description:"id"`
	Input              interface{}   `json:"input,omitempty" description:"input"`
	Result             string        `json:"result,omitempty" description:"result"`
	StartTime          string        `json:"startTime,omitempty" description:"start time"`
	State              string        `json:"state,omitempty" description:"state"`
	Type               string        `json:"type,omitempty" description:"type"`
	CauseOfBlockage    interface{}   `json:"causeOfBlockage,omitempty" description:"cause of blockage"`
	Edges              []struct {
		Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		ID    string `json:"id,omitempty" description:"id"`
		Type  string `json:"type,omitempty" description:"type"`
	} `json:"edges,omitempty"`
	FirstParent interface{} `json:"firstParent,omitempty" description:"first parent"`
	Restartable bool        `json:"restartable,omitempty" description:"restartable or not"`
	Steps       []struct {
		Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		Links struct {
			Self struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"self,omitempty"`
			Actions struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"actions,omitempty"`
		} `json:"_links,omitempty" description:"references the reachable path to this resource"`
		Actions []struct {
			Class string `json:"_class,omitempty" description:"references the reachable path to this resource"`
			Links struct {
				Self struct {
					Class string `json:"_class,omitempty"`
					Href  string `json:"href,omitempty"`
				} `json:"self,omitempty" description:""`
			} `json:"_links,omitempty" description:"references the reachable path to this resource"`
			URLName string `json:"urlName,omitempty" description:"url name"`
		} `json:"actions,omitempty"`
		DisplayDescription interface{} `json:"displayDescription,omitempty" description:"display description"`
		DisplayName        string      `json:"displayName,omitempty" description:"display name"`
		DurationInMillis   int         `json:"durationInMillis,omitempty" description:"duration time in millis"`
		ID                 string      `json:"id,omitempty" description:"id"`
		Input              interface{} `json:"input,omitempty" description:"input"`
		Result             string      `json:"result,omitempty" description:"result"`
		StartTime          string      `json:"startTime,omitempty" description:"start time"`
		State              string      `json:"state,omitempty" description:"state"`
		Type               string      `json:"type,omitempty" description:"type"`
	} `json:"steps,omitempty"`
}

// CheckPipeline
type CheckPlayload struct {
	ID         string `json:"id,omitempty" description:"id"`
	Parameters []struct {
		Name  string `json:"name,omitempty" description:"name"`
		Value string `json:"value,omitempty" description:"value"`
	} `json:"parameters,omitempty"`
	Abort bool `json:"abort,omitempty" description:"abort or not"`
}

// Getcrumb
type Crumb struct {
	Class             string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
	Crumb             string `json:"crumb,omitempty" description:"crumb data"`
	CrumbRequestField string `json:"crumbRequestField,omitempty" description:"crumb request field"`
}

// CheckScriptCompile
type CheckScript struct {
	Column  int    `json:"column,omitempty" description:"column e.g. 0"`
	Line    int    `json:"line,omitempty" description:"line e.g. 0"`
	Message string `json:"message,omitempty" description:"message e.g. success"`
	Status  string `json:"status,omitempty" description:"status e.g. success"`
}

// CheckCron
type CheckCronRes struct {
	Result  string `json:"result,omitempty" description:"result"`
	Message string `json:"message,omitempty" description:"message"`
}

// GetPipelineRun
type PipelineRun struct {
	Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
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
	} `json:"_links,omitempty" description:"references the reachable path to this resource"`
	Actions          []interface{} `json:"actions,omitempty" description:"the list of actions"`
	ArtifactsZipFile interface{}   `json:"artifactsZipFile,omitempty" description:"the artifacts zip file"`
	CauseOfBlockage  interface{}   `json:"causeOfBlockage,omitempty" description:"cause of blockage"`
	Causes           []struct {
		Class            string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		ShortDescription string `json:"shortDescription,omitempty" description:"short description"`
		UserID           string `json:"userId,omitempty" description:"user id"`
		UserName         string `json:"userName,omitempty" description:"user name"`
	} `json:"causes,omitempty"`
	ChangeSet                 []interface{} `json:"changeSet,omitempty" description:"change set"`
	Description               interface{}   `json:"description,omitempty" description:"description"`
	DurationInMillis          int           `json:"durationInMillis,omitempty" description:"duration time in millis"`
	EnQueueTime               string        `json:"enQueueTime,omitempty" description:"enqueue time"`
	EndTime                   string        `json:"endTime,omitempty" description:"end time"`
	EstimatedDurationInMillis int           `json:"estimatedDurationInMillis,omitempty" description:"estimated duration time in millis"`
	ID                        string        `json:"id,omitempty" description:"id"`
	Name                      interface{}   `json:"name,omitempty" description:"name"`
	Organization              string        `json:"organization,omitempty" description:"organization"`
	Pipeline                  string        `json:"pipeline,omitempty" description:"pipeline"`
	Replayable                bool          `json:"replayable,omitempty" description:"replayable or not"`
	Result                    string        `json:"result,omitempty" description:"result"`
	RunSummary                string        `json:"runSummary,omitempty" description:"pipeline run summary"`
	StartTime                 string        `json:"startTime,omitempty" description:"start time"`
	State                     string        `json:"state,omitempty" description:"state"`
	Type                      string        `json:"type,omitempty" description:"type"`
	Branch                    interface{}   `json:"branch,omitempty" description:"branch"`
	CommitID                  interface{}   `json:"commitId,omitempty" description:"commit id"`
	CommitURL                 interface{}   `json:"commitUrl,omitempty" description:"commit url"`
	PullRequest               interface{}   `json:"pullRequest,omitempty" description:"pull request"`
}

// GetBranchPipeRun
type BranchPipeline struct {
	Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
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
	} `json:"_links,omitempty" description:"references the reachable path to this resource"`
	Actions                   []interface{} `json:"actions,omitempty" description:"the list of actions"`
	Disabled                  bool          `json:"disabled,omitempty" description:"disable or not"`
	DisplayName               string        `json:"displayName,omitempty" description:"display name"`
	EstimatedDurationInMillis int           `json:"estimatedDurationInMillis,omitempty" description:"estimated duration time in millis"`
	FullDisplayName           string        `json:"fullDisplayName,omitempty" description:"full display name"`
	FullName                  string        `json:"fullName,omitempty" description:"full name"`
	LatestRun                 struct {
		Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
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
		} `json:"_links,omitempty" description:"references the reachable path to this resource"`
		Actions          []interface{} `json:"actions,omitempty" description:"the list of actions"`
		ArtifactsZipFile string        `json:"artifactsZipFile,omitempty" description:"the artifacts zip file"`
		CauseOfBlockage  interface{}   `json:"causeOfBlockage,omitempty" description:"cause of blockage"`
		Causes           []struct {
			Class            string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
			ShortDescription string `json:"shortDescription,omitempty" description:"short description"`
			UserID           string `json:"userId,omitempty" description:"user id"`
			UserName         string `json:"userName,omitempty" description:"user name"`
		} `json:"causes,omitempty"`
		ChangeSet                 []interface{} `json:"changeSet,omitempty" description:"change set"`
		Description               interface{}   `json:"description,omitempty" description:"description"`
		DurationInMillis          int           `json:"durationInMillis,omitempty" description:"duration time in millis"`
		EnQueueTime               string        `json:"enQueueTime,omitempty" description:"enqueue time"`
		EndTime                   string        `json:"endTime,omitempty" description:"end time"`
		EstimatedDurationInMillis int           `json:"estimatedDurationInMillis,omitempty" description:"estimated duration time in millis"`
		ID                        string        `json:"id,omitempty" description:"id"`
		Name                      interface{}   `json:"name,omitempty" description:"name"`
		Organization              string        `json:"organization,omitempty" description:"organization"`
		Pipeline                  string        `json:"pipeline,omitempty" description:"pipeline"`
		Replayable                bool          `json:"replayable,omitempty" description:"Replayable or not"`
		Result                    string        `json:"result,omitempty" description:"result"`
		RunSummary                string        `json:"runSummary,omitempty" description:"pipeline run summary"`
		StartTime                 string        `json:"startTime,omitempty" description:"start time"`
		State                     string        `json:"state,omitempty" description:"state"`
		Type                      string        `json:"type,omitempty" description:"type"`
	} `json:"latestRun,omitempty"`
	Name         string `json:"name,omitempty" description:"name"`
	Organization string `json:"organization,omitempty" description:"organization"`
	Parameters   []struct {
		Class                 string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		DefaultParameterValue struct {
			Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
			Name  string `json:"name,omitempty" description:"name"`
			Value string `json:"value,omitempty" description:"value"`
		} `json:"defaultParameterValue,omitempty" description:""`
		Description string `json:"description,omitempty" description:"description"`
		Name        string `json:"name,omitempty" description:"name"`
		Type        string `json:"type,omitempty" description:"type"`
	} `json:"parameters,omitempty"`
	Permissions struct {
		Create    bool `json:"create,omitempty" description:"create action"`
		Configure bool `json:"configure,omitempty" description:"configure action"`
		Read      bool `json:"read,omitempty" description:"read action"`
		Start     bool `json:"start,omitempty" description:"start action"`
		Stop      bool `json:"stop,omitempty" description:"stop action"`
	} `json:"permissions,omitempty"`
	WeatherScore int `json:"weatherScore,omitempty" description:"the score to description the result of pipeline"`
	Branch       struct {
		IsPrimary bool          `json:"isPrimary,omitempty" description:"primary or not"`
		Issues    []interface{} `json:"issues,omitempty" description:"issues"`
		URL       string        `json:"url,omitempty" description:"url"`
	} `json:"branch,omitempty"`
}

// GetPipelineRunNodes
type PipelineRunNodes struct {
	Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
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
	} `json:"_links,omitempty" description:"references the reachable path to this resource"`
	Actions            []interface{} `json:"actions,omitempty" description:"the list of actions"`
	DisplayDescription interface{}   `json:"displayDescription,omitempty" description:"display description"`
	DisplayName        string        `json:"displayName,omitempty" description:"display name"`
	DurationInMillis   int           `json:"durationInMillis,omitempty" description:"duration time in mullis"`
	ID                 string        `json:"id,omitempty" description:"id"`
	Input              interface{}   `json:"input,omitempty" description:"input"`
	Result             string        `json:"result,omitempty" description:"result"`
	StartTime          string        `json:"startTime,omitempty" description:"start time"`
	State              string        `json:"state,omitempty" description:"state"`
	Type               string        `json:"type,omitempty" description:"type"`
	CauseOfBlockage    interface{}   `json:"causeOfBlockage,omitempty" description:"cause 0f blockage"`
	Edges              []interface{} `json:"edges,omitempty" description:"edges"`
	FirstParent        interface{}   `json:"firstParent,omitempty" description:"first parent"`
	Restartable        bool          `json:"restartable,omitempty" description:"restartable or not"`
}

// GetNodeSteps
type NodeSteps struct {
	Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
	Links struct {
		Self struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
		Actions struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"actions,omitempty"`
	} `json:"_links,omitempty" description:"references the reachable path to this resource"`
	Actions []struct {
		Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		Links struct {
			Self struct {
				Class string `json:"_class,omitempty"`
				Href  string `json:"href,omitempty"`
			} `json:"self,omitempty"`
		} `json:"_links,omitempty" description:"references the reachable path to this resource"`
		URLName string `json:"urlName,omitempty" description:"url name"`
	} `json:"actions,omitempty"`
	DisplayDescription string      `json:"displayDescription,omitempty" description:"display description"`
	DisplayName        string      `json:"displayName,omitempty" description:"display name"`
	DurationInMillis   int         `json:"durationInMillis,omitempty" description:"duration time in mullis"`
	ID                 string      `json:"id,omitempty" description:"id"`
	Input              interface{} `json:"input,omitempty" description:"input"`
	Result             string      `json:"result,omitempty" description:"result"`
	StartTime          string      `json:"startTime,omitempty" description:"start times"`
	State              string      `json:"state,omitempty" description:"state"`
	Type               string      `json:"type,omitempty" description:"type"`
}

// CheckScriptCompile
type ReqScript struct {
	Value string `json:"value,omitempty" description:"check value"`
}

// ToJenkinsfile requests
type ReqJson struct {
	Json string `json:"json,omitempty" description:"json data"`
}

// ToJenkinsfile response
type ResJenkinsfile struct {
	Status string `json:"status,omitempty" description:"status"`
	Data   struct {
		Result      string `json:"result,omitempty" description:"result"`
		Jenkinsfile string `json:"jenkinsfile,omitempty" description:"jenkinsfile"`
		Errors      []struct {
			Location []string `json:"location,omitempty" description:"err location"`
			Error    string   `json:"error,omitempty" description:"error message"`
		} `json:"errors,omitempty"`
	} `json:"data,omitempty"`
}

type ReqJenkinsfile struct {
	Jenkinsfile string `json:"jenkinsfile,omitempty" description:"jenkinsfile"`
}

type ResJson struct {
	Status string `json:"status,omitempty" description:"status"`
	Data   struct {
		Result string `json:"result,omitempty" description:"result"`
		JSON   struct {
			Pipeline struct {
				Stages []interface{} `json:"stages,omitempty" description:"stages"`
				Agent  struct {
					Type      string `json:"type,omitempty" description:"type"`
					Arguments []struct {
						Key   string `json:"key,omitempty" description:"key"`
						Value struct {
							IsLiteral bool   `json:"isLiteral,omitempty" description:"is literal or not"`
							Value     string `json:"value,omitempty" description:"value"`
						} `json:"value,omitempty"`
					} `json:"arguments,omitempty"`
				} `json:"agent,omitempty"`
			} `json:"pipeline,omitempty"`
		} `json:"json,omitempty"`
	} `json:"data,omitempty"`
}

type NodesDetail struct {
	Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
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
	} `json:"_links,omitempty" description:"references the reachable path to this resource"`
	Actions            []interface{} `json:"actions,omitempty" description:"the list of actions"`
	DisplayDescription interface{}   `json:"displayDescription,omitempty" description:"display description"`
	DisplayName        string        `json:"displayName,omitempty" description:"display name"`
	DurationInMillis   int           `json:"durationInMillis,omitempty" description:"duration time in millis"`
	ID                 string        `json:"id,omitempty" description:"id"`
	Input              interface{}   `json:"input,omitempty" description:"input"`
	Result             string        `json:"result,omitempty" description:"result"`
	StartTime          string        `json:"startTime,omitempty" description:"start time"`
	State              string        `json:"state,omitempty" description:"statue"`
	Type               string        `json:"type,omitempty" description:"type"`
	CauseOfBlockage    interface{}   `json:"causeOfBlockage,omitempty" description:"cause of blockage"`
	Edges              []struct {
		Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		ID    string `json:"id,omitempty" description:"id"`
		Type  string `json:"type,omitempty" description:"type"`
	} `json:"edges,omitempty"`
	FirstParent interface{} `json:"firstParent,omitempty" description:"first parent"`
	Restartable bool        `json:"restartable,omitempty" description:"restartable or not"`
	Steps       []NodeSteps `json:"steps,omitempty" description:"steps"`
}

type NodesStepsIndex struct {
	Id    int         `json:"id,omitempty" description:"id"`
	Steps []NodeSteps `json:"steps,omitempty" description:"steps"`
}
