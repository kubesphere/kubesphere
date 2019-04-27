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
	Class string `json:"_class"`
	Links struct {
		Self struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"self"`
		Scm struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"scm"`
		Branches struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"branches"`
		Actions struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"actions"`
		Runs struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"runs"`
		Trends struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"trends"`
		Queue struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"queue"`
	} `json:"_links"`
	Actions         []interface{} `json:"actions"`
	Disabled        interface{}   `json:"disabled"`
	DisplayName     string        `json:"displayName"`
	FullDisplayName string        `json:"fullDisplayName"`
	FullName        string        `json:"fullName"`
	Name            string        `json:"name"`
	Organization    string        `json:"organization"`
	Parameters      interface{}   `json:"parameters"`
	Permissions     struct {
		Create    bool `json:"create"`
		Configure bool `json:"configure"`
		Read      bool `json:"read"`
		Start     bool `json:"start"`
		Stop      bool `json:"stop"`
	} `json:"permissions"`
	EstimatedDurationInMillis      int           `json:"estimatedDurationInMillis"`
	NumberOfFolders                int           `json:"numberOfFolders"`
	NumberOfPipelines              int           `json:"numberOfPipelines"`
	PipelineFolderNames            []interface{} `json:"pipelineFolderNames"`
	WeatherScore                   int           `json:"weatherScore"`
	BranchNames                    []string      `json:"branchNames"`
	NumberOfFailingBranches        int           `json:"numberOfFailingBranches"`
	NumberOfFailingPullRequests    int           `json:"numberOfFailingPullRequests"`
	NumberOfSuccessfulBranches     int           `json:"numberOfSuccessfulBranches"`
	NumberOfSuccessfulPullRequests int           `json:"numberOfSuccessfulPullRequests"`
	ScmSource                      struct {
		Class  string      `json:"_class"`
		APIURL interface{} `json:"apiUrl"`
		ID     string      `json:"id"`
	} `json:"scmSource"`
	TotalNumberOfBranches     int `json:"totalNumberOfBranches"`
	TotalNumberOfPullRequests int `json:"totalNumberOfPullRequests"`
}

// GetPipelineRun & SearchPipelineRuns
type PipelineRun struct {
	Class string `json:"_class"`
	Links struct {
		PrevRun struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"prevRun"`
		Parent struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"parent"`
		Tests struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"tests"`
		Nodes struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"nodes"`
		Log struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"log"`
		Self struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"self"`
		BlueTestSummary struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"blueTestSummary"`
		Actions struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"actions"`
		Steps struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"steps"`
		Artifacts struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"artifacts"`
		NextRun struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"nextRun"`
	} `json:"_links"`
	Actions          []interface{} `json:"actions"`
	ArtifactsZipFile interface{}   `json:"artifactsZipFile"`
	CauseOfBlockage  interface{}   `json:"causeOfBlockage"`
	Causes           []struct {
		Class            string `json:"_class"`
		ShortDescription string `json:"shortDescription"`
		UserID           string `json:"userId"`
		UserName         string `json:"userName"`
	} `json:"causes"`
	ChangeSet                 []interface{} `json:"changeSet"`
	Description               interface{}   `json:"description"`
	DurationInMillis          int           `json:"durationInMillis"`
	EnQueueTime               string        `json:"enQueueTime"`
	EndTime                   string        `json:"endTime"`
	EstimatedDurationInMillis int           `json:"estimatedDurationInMillis"`
	ID                        string        `json:"id"`
	Name                      interface{}   `json:"name"`
	Organization              string        `json:"organization"`
	Pipeline                  string        `json:"pipeline"`
	Replayable                bool          `json:"replayable"`
	Result                    string        `json:"result"`
	RunSummary                string        `json:"runSummary"`
	StartTime                 string        `json:"startTime"`
	State                     string        `json:"state"`
	Type                      string        `json:"type"`
	Branch                    struct {
		IsPrimary bool          `json:"isPrimary"`
		Issues    []interface{} `json:"issues"`
		URL       string        `json:"url"`
	} `json:"branch"`
	CommitID    string      `json:"commitId"`
	CommitURL   interface{} `json:"commitUrl"`
	PullRequest interface{} `json:"pullRequest"`
}

// GetPipelineRunNodes
type Nodes []struct {
	Class string `json:"_class"`
	Links struct {
		Self struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"self"`
		Actions struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"actions"`
		Steps struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"steps"`
	} `json:"_links"`
	Actions            []interface{} `json:"actions"`
	DisplayDescription interface{}   `json:"displayDescription"`
	DisplayName        string        `json:"displayName"`
	DurationInMillis   int           `json:"durationInMillis"`
	ID                 string        `json:"id"`
	Input              interface{}   `json:"input"`
	Result             string        `json:"result"`
	StartTime          string        `json:"startTime"`
	State              string        `json:"state"`
	Type               string        `json:"type"`
	CauseOfBlockage    interface{}   `json:"causeOfBlockage"`
	Edges              []struct {
		Class string `json:"_class"`
		ID    string `json:"id"`
		Type  string `json:"type"`
	} `json:"edges"`
	FirstParent interface{} `json:"firstParent"`
	Restartable bool        `json:"restartable"`
	Steps       []struct {
		Class string `json:"_class"`
		Links struct {
			Self struct {
				Class string `json:"_class"`
				Href  string `json:"href"`
			} `json:"self"`
			Actions struct {
				Class string `json:"_class"`
				Href  string `json:"href"`
			} `json:"actions"`
		} `json:"_links"`
		Actions []struct {
			Class string `json:"_class"`
			Links struct {
				Self struct {
					Class string `json:"_class"`
					Href  string `json:"href"`
				} `json:"self"`
			} `json:"_links"`
			URLName string `json:"urlName"`
		} `json:"actions"`
		DisplayDescription interface{} `json:"displayDescription"`
		DisplayName        string      `json:"displayName"`
		DurationInMillis   int         `json:"durationInMillis"`
		ID                 string      `json:"id"`
		Input              interface{} `json:"input"`
		Result             string      `json:"result"`
		StartTime          string      `json:"startTime"`
		State              string      `json:"state"`
		Type               string      `json:"type"`
	} `json:"steps"`
}

// Validate
type Validates struct {
	CredentialID string `json:"credentialId,omitempty"`
}

// GetSCMOrg
type SCMOrg struct {
	Class string `json:"_class"`
	Links struct {
		Repositories struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"repositories"`
		Self struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
	Avatar                      string `json:"avatar"`
	JenkinsOrganizationPipeline bool   `json:"jenkinsOrganizationPipeline"`
	Name                        string `json:"name"`
}

// GetOrgRepo
type OrgRepo struct {
	Class string `json:"_class"`
	Links struct {
		Self struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
	Repositories struct {
		Class string `json:"_class"`
		Links struct {
			Self struct {
				Class string `json:"_class"`
				Href  string `json:"href"`
			} `json:"self"`
		} `json:"_links"`
		Items []struct {
			Class string `json:"_class"`
			Links struct {
				Self struct {
					Class string `json:"_class"`
					Href  string `json:"href"`
				} `json:"self"`
			} `json:"_links"`
			DefaultBranch string `json:"defaultBranch"`
			Description   string `json:"description"`
			Name          string `json:"name"`
			Permissions   struct {
				Admin bool `json:"admin"`
				Push  bool `json:"push"`
				Pull  bool `json:"pull"`
			} `json:"permissions"`
			Private  bool   `json:"private"`
			FullName string `json:"fullName"`
		} `json:"items"`
		LastPage interface{} `json:"lastPage"`
		NextPage interface{} `json:"nextPage"`
		PageSize int         `json:"pageSize"`
	} `json:"repositories"`
}

// StopPipeline
type StopPipe struct {
	Class string `json:"_class"`
	Links struct {
		Parent struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"parent"`
		Tests struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"tests"`
		Nodes struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"nodes"`
		Log struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"log"`
		Self struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"self"`
		BlueTestSummary struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"blueTestSummary"`
		Actions struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"actions"`
		Steps struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"steps"`
		Artifacts struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"artifacts"`
	} `json:"_links"`
	Actions          []interface{} `json:"actions"`
	ArtifactsZipFile interface{}   `json:"artifactsZipFile"`
	CauseOfBlockage  interface{}   `json:"causeOfBlockage"`
	Causes           []struct {
		Class            string `json:"_class"`
		ShortDescription string `json:"shortDescription"`
	} `json:"causes"`
	ChangeSet                 []interface{} `json:"changeSet"`
	Description               interface{}   `json:"description"`
	DurationInMillis          int           `json:"durationInMillis"`
	EnQueueTime               string        `json:"enQueueTime"`
	EndTime                   string        `json:"endTime"`
	EstimatedDurationInMillis int           `json:"estimatedDurationInMillis"`
	ID                        string        `json:"id"`
	Name                      interface{}   `json:"name"`
	Organization              string        `json:"organization"`
	Pipeline                  string        `json:"pipeline"`
	Replayable                bool          `json:"replayable"`
	Result                    string        `json:"result"`
	RunSummary                string        `json:"runSummary"`
	StartTime                 string        `json:"startTime"`
	State                     string        `json:"state"`
	Type                      string        `json:"type"`
	Branch                    struct {
		IsPrimary bool          `json:"isPrimary"`
		Issues    []interface{} `json:"issues"`
		URL       string        `json:"url"`
	} `json:"branch"`
	CommitID    string      `json:"commitId"`
	CommitURL   interface{} `json:"commitUrl"`
	PullRequest interface{} `json:"pullRequest"`
}

// ReplayPipeline
type ReplayPipe struct {
	Class string `json:"_class"`
	Links struct {
		Parent struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"parent"`
		Tests struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"tests"`
		Log struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"log"`
		Self struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"self"`
		BlueTestSummary struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"blueTestSummary"`
		Actions struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"actions"`
		Artifacts struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"artifacts"`
	} `json:"_links"`
	Actions          []interface{} `json:"actions"`
	ArtifactsZipFile interface{}   `json:"artifactsZipFile"`
	CauseOfBlockage  string        `json:"causeOfBlockage"`
	Causes           []struct {
		Class            string `json:"_class"`
		ShortDescription string `json:"shortDescription"`
		UserID           string `json:"userId,omitempty"`
		UserName         string `json:"userName,omitempty"`
	} `json:"causes"`
	ChangeSet                 []interface{} `json:"changeSet"`
	Description               interface{}   `json:"description"`
	DurationInMillis          interface{}   `json:"durationInMillis"`
	EnQueueTime               interface{}   `json:"enQueueTime"`
	EndTime                   interface{}   `json:"endTime"`
	EstimatedDurationInMillis interface{}   `json:"estimatedDurationInMillis"`
	ID                        string        `json:"id"`
	Name                      interface{}   `json:"name"`
	Organization              string        `json:"organization"`
	Pipeline                  string        `json:"pipeline"`
	Replayable                bool          `json:"replayable"`
	Result                    string        `json:"result"`
	RunSummary                interface{}   `json:"runSummary"`
	StartTime                 interface{}   `json:"startTime"`
	State                     string        `json:"state"`
	Type                      string        `json:"type"`
	QueueID                   string        `json:"queueId"`
}

// GetArtifacts
type Artifacts struct {
	Class string `json:"_class"`
	Links struct {
		Self struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
	Downloadable bool   `json:"downloadable"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	Size         int    `json:"size"`
	URL          string `json:"url"` // The url for Download artifacts
}

// GetPipeBranch
type PipeBranch struct {
	Class string `json:"_class"`
	Links struct {
		Self struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"self"`
		Scm struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"scm"`
		Actions struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"actions"`
		Runs struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"runs"`
		Trends struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"trends"`
		Queue struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"queue"`
	} `json:"_links"`
	Actions                   []interface{} `json:"actions"`
	Disabled                  bool          `json:"disabled"`
	DisplayName               string        `json:"displayName"`
	EstimatedDurationInMillis int           `json:"estimatedDurationInMillis"`
	FullDisplayName           string        `json:"fullDisplayName"`
	FullName                  string        `json:"fullName"`
	LatestRun                 struct {
		Class string `json:"_class"`
		Links struct {
			PrevRun struct {
				Class string `json:"_class"`
				Href  string `json:"href"`
			} `json:"prevRun"`
			Parent struct {
				Class string `json:"_class"`
				Href  string `json:"href"`
			} `json:"parent"`
			Tests struct {
				Class string `json:"_class"`
				Href  string `json:"href"`
			} `json:"tests"`
			Log struct {
				Class string `json:"_class"`
				Href  string `json:"href"`
			} `json:"log"`
			Self struct {
				Class string `json:"_class"`
				Href  string `json:"href"`
			} `json:"self"`
			BlueTestSummary struct {
				Class string `json:"_class"`
				Href  string `json:"href"`
			} `json:"blueTestSummary"`
			Actions struct {
				Class string `json:"_class"`
				Href  string `json:"href"`
			} `json:"actions"`
			Artifacts struct {
				Class string `json:"_class"`
				Href  string `json:"href"`
			} `json:"artifacts"`
		} `json:"_links"`
		Actions          []interface{} `json:"actions"`
		ArtifactsZipFile string        `json:"artifactsZipFile"`
		CauseOfBlockage  interface{}   `json:"causeOfBlockage"`
		Causes           []struct {
			Class            string `json:"_class"`
			ShortDescription string `json:"shortDescription"`
		} `json:"causes"`
		ChangeSet                 []interface{} `json:"changeSet"`
		Description               interface{}   `json:"description"`
		DurationInMillis          int           `json:"durationInMillis"`
		EnQueueTime               string        `json:"enQueueTime"`
		EndTime                   string        `json:"endTime"`
		EstimatedDurationInMillis int           `json:"estimatedDurationInMillis"`
		ID                        string        `json:"id"`
		Name                      interface{}   `json:"name"`
		Organization              string        `json:"organization"`
		Pipeline                  string        `json:"pipeline"`
		Replayable                bool          `json:"replayable"`
		Result                    string        `json:"result"`
		RunSummary                string        `json:"runSummary"`
		StartTime                 string        `json:"startTime"`
		State                     string        `json:"state"`
		Type                      string        `json:"type"`
	} `json:"latestRun"`
	Name         string `json:"name"`
	Organization string `json:"organization"`
	Parameters   []struct {
		Class                 string `json:"_class"`
		DefaultParameterValue struct {
			Class string `json:"_class"`
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"defaultParameterValue"`
		Description string `json:"description"`
		Name        string `json:"name"`
		Type        string `json:"type"`
	} `json:"parameters"`
	Permissions struct {
		Create    bool `json:"create"`
		Configure bool `json:"configure"`
		Read      bool `json:"read"`
		Start     bool `json:"start"`
		Stop      bool `json:"stop"`
	} `json:"permissions"`
	WeatherScore int `json:"weatherScore"`
	Branch       struct {
		IsPrimary bool          `json:"isPrimary"`
		Issues    []interface{} `json:"issues"`
		URL       string        `json:"url"`
	} `json:"branch"`
}

// RunPipeline
type RunPayload struct {
	Parameters []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"parameters"`
}

type QueuedBlueRun struct {
	Class string `json:"_class"`
	Links struct {
		Parent struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"parent"`
		Tests struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"tests"`
		Log struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"log"`
		Self struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"self"`
		BlueTestSummary struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"blueTestSummary"`
		Actions struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"actions"`
		Artifacts struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"artifacts"`
	} `json:"_links"`
	Actions          []interface{} `json:"actions"`
	ArtifactsZipFile interface{}   `json:"artifactsZipFile"`
	CauseOfBlockage  string        `json:"causeOfBlockage"`
	Causes           []struct {
		Class            string `json:"_class"`
		ShortDescription string `json:"shortDescription"`
		UserID           string `json:"userId"`
		UserName         string `json:"userName"`
	} `json:"causes"`
	ChangeSet                 []interface{} `json:"changeSet"`
	Description               interface{}   `json:"description"`
	DurationInMillis          interface{}   `json:"durationInMillis"`
	EnQueueTime               interface{}   `json:"enQueueTime"`
	EndTime                   interface{}   `json:"endTime"`
	EstimatedDurationInMillis interface{}   `json:"estimatedDurationInMillis"`
	ID                        string        `json:"id"`
	Name                      interface{}   `json:"name"`
	Organization              string        `json:"organization"`
	Pipeline                  string        `json:"pipeline"`
	Replayable                bool          `json:"replayable"`
	Result                    string        `json:"result"`
	RunSummary                interface{}   `json:"runSummary"`
	StartTime                 interface{}   `json:"startTime"`
	State                     string        `json:"state"`
	Type                      string        `json:"type"`
	QueueID                   string        `json:"queueId"`
}

// GetNodeStatus
type NodeStatus []struct {
	Class string `json:"_class"`
	Links struct {
		Self struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"self"`
		Actions struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"actions"`
		Steps struct {
			Class string `json:"_class"`
			Href  string `json:"href"`
		} `json:"steps"`
	} `json:"_links"`
	Actions            []interface{} `json:"actions"`
	DisplayDescription interface{}   `json:"displayDescription"`
	DisplayName        string        `json:"displayName"`
	DurationInMillis   int           `json:"durationInMillis"`
	ID                 string        `json:"id"`
	Input              interface{}   `json:"input"`
	Result             string        `json:"result"`
	StartTime          string        `json:"startTime"`
	State              string        `json:"state"`
	Type               string        `json:"type"`
	CauseOfBlockage    interface{}   `json:"causeOfBlockage"`
	Edges              []struct {
		Class string `json:"_class"`
		ID    string `json:"id"`
		Type  string `json:"type"`
	} `json:"edges"`
	FirstParent interface{} `json:"firstParent"`
	Restartable bool        `json:"restartable"`
	Steps       []struct {
		Class string `json:"_class"`
		Links struct {
			Self struct {
				Class string `json:"_class"`
				Href  string `json:"href"`
			} `json:"self"`
			Actions struct {
				Class string `json:"_class"`
				Href  string `json:"href"`
			} `json:"actions"`
		} `json:"_links"`
		Actions []struct {
			Class string `json:"_class"`
			Links struct {
				Self struct {
					Class string `json:"_class"`
					Href  string `json:"href"`
				} `json:"self"`
			} `json:"_links"`
			URLName string `json:"urlName"`
		} `json:"actions"`
		DisplayDescription interface{} `json:"displayDescription"`
		DisplayName        string      `json:"displayName"`
		DurationInMillis   int         `json:"durationInMillis"`
		ID                 string      `json:"id"`
		Input              interface{} `json:"input"`
		Result             string      `json:"result"`
		StartTime          string      `json:"startTime"`
		State              string      `json:"state"`
		Type               string      `json:"type"`
	} `json:"steps"`
}

// CheckPipeline
type CheckPlayload struct {
	ID         string `json:"id"`
	Parameters []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"parameters"`
}

// Getcrumb
type Crumb struct {
	Class             string `json:"_class"`
	Crumb             string `json:"crumb"`
	CrumbRequestField string `json:"crumbRequestField"`
}
