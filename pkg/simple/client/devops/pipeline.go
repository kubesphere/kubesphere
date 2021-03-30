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

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type PipelineList struct {
	Items []Pipeline `json:"items"`
	Total int        `json:"total_count"`
}

// GetPipeline & SearchPipelines
type Pipeline struct {
	Annotations map[string]string `json:"annotations,omitempty" description:"Add annotations from crd" `
	Class       string            `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability." `
	Links       struct {
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
	Actions         []interface{} `json:"actions,omitempty" description:"the list of all actions."`
	Disabled        interface{}   `json:"disabled,omitempty" description:"disable or not, if disabled, can not do any action."`
	DisplayName     string        `json:"displayName,omitempty" description:"display name"`
	FullDisplayName string        `json:"fullDisplayName,omitempty" description:"full display name"`
	FullName        string        `json:"fullName,omitempty" description:"full name"`
	Name            string        `json:"name,omitempty" description:"name"`
	Organization    string        `json:"organization,omitempty" description:"the name of organization"`
	Parameters      interface{}   `json:"parameters,omitempty" description:"parameters of pipeline, a pipeline can define list of parameters pipeline job expects."`
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
	WeatherScore                   int           `json:"weatherScore" description:"the score to description the result of pipeline activity"`
	BranchNames                    []string      `json:"branchNames,omitempty" description:"branch names"`
	NumberOfFailingBranches        int           `json:"numberOfFailingBranches,omitempty" description:"number of failing branches"`
	NumberOfFailingPullRequests    int           `json:"numberOfFailingPullRequests,omitempty" description:"number of failing pull requests"`
	NumberOfSuccessfulBranches     int           `json:"numberOfSuccessfulBranches,omitempty" description:"number of successful pull requests"`
	NumberOfSuccessfulPullRequests int           `json:"numberOfSuccessfulPullRequests,omitempty" description:"number of successful pull requests"`
	ScmSource                      struct {
		Class  string      `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		APIURL interface{} `json:"apiUrl,omitempty" description:"api url"`
		ID     string      `json:"id,omitempty" description:"The id of the source configuration management (SCM)."`
	} `json:"scmSource,omitempty"`
	TotalNumberOfBranches     int `json:"totalNumberOfBranches,omitempty" description:"total number of branches"`
	TotalNumberOfPullRequests int `json:"totalNumberOfPullRequests,omitempty" description:"total number of pull requests"`
}

// UnmarshalPipeline unmarshal data into the Pipeline list
func UnmarshalPipeline(total int, data []byte) (pipelineList *PipelineList, err error) {
	pipelineList = &PipelineList{Total: total}
	pipelineList.Items = make([]Pipeline, total)
	for i, _ := range pipelineList.Items {
		pipelineList.Items[i].WeatherScore = 100
	}
	err = json.Unmarshal(data, &pipelineList.Items)
	return
}

// GetPipeBranchRun & SearchPipelineRuns
type PipelineRunList struct {
	Items []PipelineRun `json:"items"`
	Total int           `json:"totalItems"`
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
	Actions            []interface{} `json:"actions,omitempty" description:"the list of all actions"`
	DisplayDescription interface{}   `json:"displayDescription,omitempty" description:"display description"`
	DisplayName        string        `json:"displayName,omitempty" description:"display name"`
	DurationInMillis   int           `json:"durationInMillis,omitempty" description:"duration time in millis"`
	ID                 string        `json:"id,omitempty" description:"id"`
	Input              *Input        `json:"input,omitempty" description:"the action should user input"`
	Result             string        `json:"result,omitempty" description:"the result of pipeline run. e.g. SUCCESS. e.g. SUCCESS"`
	StartTime          string        `json:"startTime,omitempty" description:"the time of start"`
	State              string        `json:"state,omitempty" description:"run state. e.g. RUNNING"`
	Type               string        `json:"type,omitempty" description:"source type, e.g. \"WorkflowRun\""`
	CauseOfBlockage    interface{}   `json:"causeOfBlockage,omitempty" description:"the cause of blockage"`
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
		Input              *Input      `json:"input,omitempty" description:"the action should user input"`
		Result             string      `json:"result,omitempty" description:"result"`
		StartTime          string      `json:"startTime,omitempty" description:"the time of start"`
		State              string      `json:"state,omitempty" description:"run state. e.g. RUNNING"`
		Type               string      `json:"type,omitempty" description:"source type"`
	} `json:"steps,omitempty"`
}

// Validate
type Validates struct {
	CredentialID string `json:"credentialId,omitempty" description:"the id of credential"`
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
	Avatar                      string `json:"avatar,omitempty" description:"the url of organization avatar"`
	Key                         string `json:"key,omitempty" description:"the key of a Bitbucket organization"`
	JenkinsOrganizationPipeline bool   `json:"jenkinsOrganizationPipeline,omitempty" description:"weather or not already have jenkins pipeline."`
	Name                        string `json:"name,omitempty" description:"organization name"`
}

type SCMServer struct {
	Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
	Links struct {
		Self struct {
			Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
			Href  string `json:"href,omitempty" description:"self url in api"`
		} `json:"self,omitempty" description:"scm server self info"`
	} `json:"_links,omitempty" description:"references the reachable path to this resource"`
	ID     string `json:"id,omitempty" description:"server id of scm server"`
	Name   string `json:"name,omitempty" description:"name of scm server"`
	ApiURL string `json:"apiUrl,omitempty"  description:"url of scm server"`
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
			Private  bool   `json:"private,omitempty" description:"private or not"`
			FullName string `json:"fullName,omitempty" description:"full name"`
		} `json:"items,omitempty"`
		LastPage interface{} `json:"lastPage,omitempty" description:"last page"`
		NextPage interface{} `json:"nextPage,omitempty" description:"next page"`
		PageSize int         `json:"pageSize,omitempty" description:"page size"`
	} `json:"repositories,omitempty"`
}

// StopPipeline
type StopPipeline struct {
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
	Actions          []interface{} `json:"actions,omitempty" description:"the list of all actions."`
	ArtifactsZipFile interface{}   `json:"artifactsZipFile,omitempty" description:"the artifacts zip file"`
	CauseOfBlockage  interface{}   `json:"causeOfBlockage,omitempty" description:"the cause of blockage"`
	Causes           []struct {
		Class            string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		ShortDescription string `json:"shortDescription,omitempty" description:"short description"`
	} `json:"causes,omitempty"`
	ChangeSet                 []interface{} `json:"changeSet,omitempty" description:"changeset information"`
	Description               interface{}   `json:"description,omitempty" description:"description"`
	DurationInMillis          int           `json:"durationInMillis,omitempty" description:"duration time in millis"`
	EnQueueTime               string        `json:"enQueueTime,omitempty" description:"the time of enter the queue"`
	EndTime                   string        `json:"endTime,omitempty" description:"the time of end"`
	EstimatedDurationInMillis int           `json:"estimatedDurationInMillis,omitempty" description:"estimated duration time in millis"`
	ID                        string        `json:"id,omitempty" description:"id"`
	Name                      interface{}   `json:"name,omitempty" description:"name"`
	Organization              string        `json:"organization,omitempty" description:"the name of organization"`
	Pipeline                  string        `json:"pipeline,omitempty" description:"pipeline"`
	Replayable                bool          `json:"replayable,omitempty" description:"replayable or not"`
	Result                    string        `json:"result,omitempty" description:"the result of pipeline run. e.g. SUCCESS"`
	RunSummary                string        `json:"runSummary,omitempty" description:"pipeline run summary"`
	StartTime                 string        `json:"startTime,omitempty" description:"the time of start"`
	State                     string        `json:"state,omitempty" description:"run state. e.g. RUNNING"`
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
type ReplayPipeline struct {
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
	Actions          []interface{} `json:"actions,omitempty" description:"the list of all actions."`
	ArtifactsZipFile interface{}   `json:"artifactsZipFile,omitempty" description:"the artifacts zip file"`
	CauseOfBlockage  string        `json:"causeOfBlockage,omitempty" description:"the cause of blockage"`
	Causes           []struct {
		Class            string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		ShortDescription string `json:"shortDescription,omitempty" description:"short description"`
		UserID           string `json:"userId,omitempty" description:"user id"`
		UserName         string `json:"userName,omitempty" description:"user name"`
	} `json:"causes,omitempty"`
	ChangeSet                 []interface{} `json:"changeSet,omitempty" description:"changeset information"`
	Description               interface{}   `json:"description,omitempty" description:"description"`
	DurationInMillis          interface{}   `json:"durationInMillis,omitempty" description:"duration time in millis"`
	EnQueueTime               interface{}   `json:"enQueueTime,omitempty" description:"the time of enter the queue"`
	EndTime                   interface{}   `json:"endTime,omitempty" description:"the time of end"`
	EstimatedDurationInMillis interface{}   `json:"estimatedDurationInMillis,omitempty" description:"estimated duration time, unit is millis"`
	ID                        string        `json:"id,omitempty" description:"id"`
	Name                      interface{}   `json:"name,omitempty" description:"name"`
	Organization              string        `json:"organization,omitempty" description:"the name of organization"`
	Pipeline                  string        `json:"pipeline,omitempty" description:"pipeline"`
	Replayable                bool          `json:"replayable,omitempty" description:"replayable or not"`
	Result                    string        `json:"result,omitempty" description:"the result of pipeline run. e.g. SUCCESS"`
	RunSummary                interface{}   `json:"runSummary,omitempty" description:"pipeline run summary"`
	StartTime                 interface{}   `json:"startTime,omitempty" description:"the time of start"`
	State                     string        `json:"state,omitempty" description:"run state. e.g. RUNNING"`
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
type PipelineBranch []PipelineBranchItem

type PipelineBranchItem struct {
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
	Actions                   []interface{} `json:"actions,omitempty" description:"the list of all actions."`
	Disabled                  bool          `json:"disabled,omitempty" description:"disable or not, if disabled, can not do any action"`
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
		Actions          []interface{} `json:"actions,omitempty" description:"the list of all actions"`
		ArtifactsZipFile string        `json:"artifactsZipFile,omitempty" description:"the artifacts zip file"`
		CauseOfBlockage  interface{}   `json:"causeOfBlockage,omitempty" description:"the cause of blockage"`
		Causes           []struct {
			Class            string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
			ShortDescription string `json:"shortDescription,omitempty" description:"short description"`
		} `json:"causes,omitempty"`
		ChangeSet                 []interface{} `json:"changeSet,omitempty" description:"changeset information"`
		Description               interface{}   `json:"description,omitempty" description:"description"`
		DurationInMillis          int           `json:"durationInMillis,omitempty" description:"duration time in millis"`
		EnQueueTime               string        `json:"enQueueTime,omitempty" description:"the time of enter the queue"`
		EndTime                   string        `json:"endTime,omitempty" description:"the time of end"`
		EstimatedDurationInMillis int           `json:"estimatedDurationInMillis,omitempty" description:"estimated duration time in millis"`
		ID                        string        `json:"id,omitempty" description:"id"`
		Name                      interface{}   `json:"name,omitempty" description:"name"`
		Organization              string        `json:"organization,omitempty" description:"the name of organization"`
		Pipeline                  string        `json:"pipeline,omitempty" description:"pipeline"`
		Replayable                bool          `json:"replayable,omitempty" description:"replayable or not"`
		Result                    string        `json:"result,omitempty" description:"the result of pipeline run. e.g. SUCCESS"`
		RunSummary                string        `json:"runSummary,omitempty" description:"pipeline run summary"`
		StartTime                 string        `json:"startTime,omitempty" description:"start run"`
		State                     string        `json:"state,omitempty" description:"run state. e.g. RUNNING"`
		Type                      string        `json:"type,omitempty" description:"type"`
	} `json:"latestRun,omitempty"`
	Name         string `json:"name,omitempty" description:"name"`
	Organization string `json:"organization,omitempty" description:"the name of organization"`
	Parameters   []struct {
		Class                 string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		DefaultParameterValue struct {
			Class string      `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
			Name  string      `json:"name,omitempty" description:"name"`
			Value interface{} `json:"value,omitempty" description:"value"`
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
	PullRequest struct {
		Author string `json:"author,omitempty" description:"author of pull request"`
		ID     string `json:"id,omitempty" description:"id of pull request"`
		Title  string `json:"title,omitempty" description:"title of pull request"`
		URL    string `json:"url,omitempty" description:"url of pull request"`
	} `json:"pullRequest,omitempty"`
}

// RunPipeline
type RunPayload struct {
	Parameters []struct {
		Name  string      `json:"name,omitempty" description:"name"`
		Value interface{} `json:"value,omitempty" description:"value"`
	} `json:"parameters,omitempty"`
}

type RunPipeline struct {
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
	Actions          []interface{} `json:"actions,omitempty" description:"the list of all actions"`
	ArtifactsZipFile interface{}   `json:"artifactsZipFile,omitempty" description:"the artifacts zip file"`
	CauseOfBlockage  string        `json:"causeOfBlockage,omitempty" description:"the cause of blockage"`
	Causes           []struct {
		Class            string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		ShortDescription string `json:"shortDescription,omitempty" description:"short description"`
		UserID           string `json:"userId,omitempty" description:"user id"`
		UserName         string `json:"userName,omitempty" description:"user name"`
	} `json:"causes,omitempty"`
	ChangeSet                 []interface{} `json:"changeSet,omitempty" description:"changeset information"`
	Description               interface{}   `json:"description,omitempty" description:"description"`
	DurationInMillis          interface{}   `json:"durationInMillis,omitempty" description:"duration time in millis"`
	EnQueueTime               interface{}   `json:"enQueueTime,omitempty" description:"the time of enter the queue"`
	EndTime                   interface{}   `json:"endTime,omitempty" description:"the time of end"`
	EstimatedDurationInMillis interface{}   `json:"estimatedDurationInMillis,omitempty" description:"estimated duration time in millis"`
	ID                        string        `json:"id,omitempty" description:"id"`
	Name                      interface{}   `json:"name,omitempty" description:"name"`
	Organization              string        `json:"organization,omitempty" description:"the name of organization"`
	Pipeline                  string        `json:"pipeline,omitempty" description:"pipeline"`
	Replayable                bool          `json:"replayable,omitempty" description:"replayable or not"`
	Result                    string        `json:"result,omitempty" description:"the result of pipeline run. e.g. SUCCESS"`
	RunSummary                interface{}   `json:"runSummary,omitempty" description:"pipeline run summary"`
	StartTime                 interface{}   `json:"startTime,omitempty" description:"the time of start"`
	State                     string        `json:"state,omitempty" description:"run state. e.g. RUNNING"`
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
	Actions            []interface{} `json:"actions,omitempty" description:"the list of all actions"`
	DisplayDescription interface{}   `json:"displayDescription,omitempty" description:"display description"`
	DisplayName        string        `json:"displayName,omitempty" description:"display name"`
	DurationInMillis   int           `json:"durationInMillis,omitempty" description:"duration time in millis"`
	ID                 string        `json:"id,omitempty" description:"id"`
	Input              *Input        `json:"input,omitempty" description:"the action should user input"`
	Result             string        `json:"result,omitempty" description:"the result of pipeline run. e.g. SUCCESS"`
	StartTime          string        `json:"startTime,omitempty" description:"the time of start"`
	State              string        `json:"state,omitempty" description:"run state. e.g. RUNNING"`
	Type               string        `json:"type,omitempty" description:"type"`
	CauseOfBlockage    interface{}   `json:"causeOfBlockage,omitempty" description:"the cause of blockage"`
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
		Input              *Input      `json:"input,omitempty" description:"the action should user input"`
		Result             string      `json:"result,omitempty" description:"the result of pipeline run. e.g. SUCCESS"`
		StartTime          string      `json:"startTime,omitempty" description:"the time of start"`
		State              string      `json:"state,omitempty" description:"run state. e.g. RUNNING"`
		Type               string      `json:"type,omitempty" description:"type"`
	} `json:"steps,omitempty"`
}

// CheckPipeline
type CheckPlayload struct {
	ID         string                    `json:"id,omitempty" description:"id"`
	Parameters []CheckPlayloadParameters `json:"parameters,omitempty"`
	Abort      bool                      `json:"abort,omitempty" description:"abort or not"`
}

type CreateScmServerReq struct {
	Name   string `json:"name,omitempty" description:"name of scm server"`
	ApiURL string `json:"apiUrl,omitempty"  description:"url of scm server"`
}

type CheckPlayloadParameters struct {
	Name  string      `json:"name,omitempty" description:"name"`
	Value interface{} `json:"value,omitempty" description:"value"`
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
	Message string `json:"message,omitempty" description:"message e.g. unexpected char: '#'"`
	Status  string `json:"status,omitempty" description:"status e.g. fail"`
}

// CheckCron
type CronData struct {
	PipelineName string `json:"pipelineName,omitempty" description:"Pipeline name, if pipeline haven't created, not required'"`
	Cron         string `json:"cron" description:"Cron script data."`
}

type CheckCronRes struct {
	Result   string `json:"result,omitempty" description:"result e.g. ok, error"`
	Message  string `json:"message,omitempty" description:"message"`
	LastTime string `json:"lastTime,omitempty" description:"last run time."`
	NextTime string `json:"nextTime,omitempty" description:"next run time."`
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
	Actions          []interface{} `json:"actions,omitempty" description:"the list of all actions"`
	ArtifactsZipFile interface{}   `json:"artifactsZipFile,omitempty" description:"the artifacts zip file"`
	CauseOfBlockage  interface{}   `json:"causeOfBlockage,omitempty" description:"the cause of blockage"`
	Causes           []struct {
		Class            string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		ShortDescription string `json:"shortDescription,omitempty" description:"short description"`
		UserID           string `json:"userId,omitempty" description:"user id"`
		UserName         string `json:"userName,omitempty" description:"user name"`
	} `json:"causes,omitempty"`
	ChangeSet                 []interface{} `json:"changeSet,omitempty" description:"changeset information"`
	Description               interface{}   `json:"description,omitempty" description:"description"`
	DurationInMillis          int           `json:"durationInMillis,omitempty" description:"duration time in millis"`
	EnQueueTime               string        `json:"enQueueTime,omitempty" description:"the time of enter the queue"`
	EndTime                   string        `json:"endTime,omitempty" description:"the time of end"`
	EstimatedDurationInMillis int           `json:"estimatedDurationInMillis,omitempty" description:"estimated duration time in millis"`
	ID                        string        `json:"id,omitempty" description:"id"`
	Name                      interface{}   `json:"name,omitempty" description:"name"`
	Organization              string        `json:"organization,omitempty" description:"the name of organization"`
	Pipeline                  string        `json:"pipeline,omitempty" description:"the name of pipeline"`
	Replayable                bool          `json:"replayable,omitempty" description:"replayable or not"`
	Result                    string        `json:"result,omitempty" description:"the result of pipeline run. e.g. SUCCESS"`
	RunSummary                string        `json:"runSummary,omitempty" description:"pipeline run summary"`
	StartTime                 string        `json:"startTime,omitempty" description:"the time of start"`
	State                     string        `json:"state,omitempty" description:"run state. e.g. RUNNING"`
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
	Actions                   []interface{} `json:"actions,omitempty" description:"the list of all actions"`
	Disabled                  bool          `json:"disabled,omitempty" description:"disable or not, if disabled, can not do any action"`
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
		Actions          []interface{} `json:"actions,omitempty" description:"the list of all actions"`
		ArtifactsZipFile string        `json:"artifactsZipFile,omitempty" description:"the artifacts zip file"`
		CauseOfBlockage  interface{}   `json:"causeOfBlockage,omitempty" description:"the cause of blockage"`
		Causes           []struct {
			Class            string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
			ShortDescription string `json:"shortDescription,omitempty" description:"short description"`
			UserID           string `json:"userId,omitempty" description:"user id"`
			UserName         string `json:"userName,omitempty" description:"user name"`
		} `json:"causes,omitempty"`
		ChangeSet                 []interface{} `json:"changeSet,omitempty" description:"changeset information"`
		Description               interface{}   `json:"description,omitempty" description:"description"`
		DurationInMillis          int           `json:"durationInMillis,omitempty" description:"duration time in millis"`
		EnQueueTime               string        `json:"enQueueTime,omitempty" description:"the time of enter the queue"`
		EndTime                   string        `json:"endTime,omitempty" description:"the time of end"`
		EstimatedDurationInMillis int           `json:"estimatedDurationInMillis,omitempty" description:"estimated duration time in millis"`
		ID                        string        `json:"id,omitempty" description:"id"`
		Name                      interface{}   `json:"name,omitempty" description:"name"`
		Organization              string        `json:"organization,omitempty" description:"the name of organization"`
		Pipeline                  string        `json:"pipeline,omitempty" description:"pipeline"`
		Replayable                bool          `json:"replayable,omitempty" description:"Replayable or not"`
		Result                    string        `json:"result,omitempty" description:"the result of pipeline run. e.g. SUCCESS"`
		RunSummary                string        `json:"runSummary,omitempty" description:"pipeline run summary"`
		StartTime                 string        `json:"startTime,omitempty" description:"the time of start"`
		State                     string        `json:"state,omitempty" description:"run state. e.g. RUNNING"`
		Type                      string        `json:"type,omitempty" description:"type"`
	} `json:"latestRun,omitempty"`
	Name         string `json:"name,omitempty" description:"name"`
	Organization string `json:"organization,omitempty" description:"the name of organization"`
	Parameters   []struct {
		Class                 string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
		DefaultParameterValue struct {
			Class string      `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
			Name  string      `json:"name,omitempty" description:"name"`
			Value interface{} `json:"value,omitempty" description:"value, string or bool type"`
		} `json:"defaultParameterValue,omitempty" description:""`
		Description string        `json:"description,omitempty" description:"description"`
		Name        string        `json:"name,omitempty" description:"name"`
		Type        string        `json:"type,omitempty" description:"type"`
		Choices     []interface{} `json:"choices,omitempty" description:"choices"`
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
	Actions            []interface{} `json:"actions,omitempty" description:"the list of all actions"`
	DisplayDescription interface{}   `json:"displayDescription,omitempty" description:"display description"`
	DisplayName        string        `json:"displayName,omitempty" description:"display name"`
	DurationInMillis   int           `json:"durationInMillis,omitempty" description:"duration time in mullis"`
	ID                 string        `json:"id,omitempty" description:"id"`
	Input              *Input        `json:"input,omitempty" description:"the action should user input"`
	Result             string        `json:"result,omitempty" description:"the result of pipeline run. e.g. SUCCESS"`
	StartTime          string        `json:"startTime,omitempty" description:"the time of start"`
	State              string        `json:"state,omitempty" description:"run state. e.g. FINISHED"`
	Type               string        `json:"type,omitempty" description:"type"`
	CauseOfBlockage    interface{}   `json:"causeOfBlockage,omitempty" description:"the cause of blockage"`
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
	DisplayDescription string `json:"displayDescription,omitempty" description:"display description"`
	DisplayName        string `json:"displayName,omitempty" description:"display name"`
	DurationInMillis   int    `json:"durationInMillis,omitempty" description:"duration time in mullis"`
	ID                 string `json:"id,omitempty" description:"id"`
	Input              *Input `json:"input,omitempty" description:"the action should user input"`
	Result             string `json:"result,omitempty" description:"the result of pipeline run. e.g. SUCCESS"`
	StartTime          string `json:"startTime,omitempty" description:"the time of starts"`
	State              string `json:"state,omitempty" description:"run state. e.g. SKIPPED"`
	Type               string `json:"type,omitempty" description:"type"`
	// Approvable indicates if this step can be approved by current user
	Approvable bool `json:"aprovable" description:"indicate if this step can be approved by current user"`
}

// CheckScriptCompile
type ReqScript struct {
	Value string `json:"value,omitempty" description:"Pipeline script data"`
}

// ToJenkinsfile requests
type ReqJson struct {
	Json string `json:"json,omitempty" description:"json data"`
}

// ToJenkinsfile response
type ResJenkinsfile struct {
	Status string `json:"status,omitempty" description:"status e.g. ok"`
	Data   struct {
		Result      string `json:"result,omitempty" description:"result e.g. success"`
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
	Actions            []interface{} `json:"actions,omitempty" description:"the list of all actions"`
	DisplayDescription interface{}   `json:"displayDescription,omitempty" description:"display description"`
	DisplayName        string        `json:"displayName,omitempty" description:"display name"`
	DurationInMillis   int           `json:"durationInMillis,omitempty" description:"duration time in mullis"`
	ID                 string        `json:"id,omitempty" description:"id"`
	Input              *Input        `json:"input,omitempty" description:"the action should user input"`
	Result             string        `json:"result,omitempty" description:"the result of pipeline run. e.g. SUCCESS"`
	StartTime          string        `json:"startTime,omitempty" description:"the time of start"`
	State              string        `json:"state,omitempty" description:"run state. e.g. FINISHED"`
	Type               string        `json:"type,omitempty" description:"type"`
	CauseOfBlockage    interface{}   `json:"causeOfBlockage,omitempty" description:"the cause of blockage"`
	Edges              []interface{} `json:"edges,omitempty" description:"edges"`
	FirstParent        interface{}   `json:"firstParent,omitempty" description:"first parent"`
	Restartable        bool          `json:"restartable,omitempty" description:"restartable or not"`
	Steps              []NodeSteps   `json:"steps,omitempty" description:"steps"`
}

const (
	// StatePaused indicates a node or a step was paused, for example it's waiting for an iput
	StatePaused = "PAUSED"
)

type NodesStepsIndex struct {
	Id    int         `json:"id,omitempty" description:"id"`
	Steps []NodeSteps `json:"steps,omitempty" description:"steps"`
}

type Input struct {
	Class string `json:"_class,omitempty" description:"It’s a fully qualified name and is an identifier of the producer of this resource's capability."`
	Links *struct {
		Self *struct {
			Class string `json:"_class,omitempty"`
			Href  string `json:"href,omitempty"`
		} `json:"self,omitempty"`
	} `json:"_links,omitempty" description:"references the reachable path to this resource"`
	ID         string        `json:"id,omitempty" description:"the id of check action"`
	Message    string        `json:"message,omitempty" description:"the message of check action"`
	Ok         string        `json:"ok,omitempty" description:"check status. e.g. \"Proceed\""`
	Parameters []interface{} `json:"parameters,omitempty" description:"the parameters of check action"`
	Submitter  interface{}   `json:"submitter,omitempty" description:"check submitter"`
}

// GetSubmitters returns the all submitters related to this input
func (i *Input) GetSubmitters() (submitters []string) {
	if i.Submitter == nil {
		return
	}

	submitterArray := strings.Split(fmt.Sprintf("%v", i.Submitter), ",")
	submitters = make([]string, len(submitterArray))
	for i, submitter := range submitterArray {
		submitters[i] = strings.TrimSpace(submitter)
	}
	return
}

// Approvable returns the result if the given identify (username or group name) can approve this input
func (i *Input) Approvable(identify string) (ok bool) {
	submitters := i.GetSubmitters()

	for _, submitter := range submitters {
		if submitter == identify {
			ok = true
			break
		}
	}
	return
}

type HttpParameters struct {
	Method   string        `json:"method,omitempty"`
	Header   http.Header   `json:"header,omitempty"`
	Body     io.ReadCloser `json:"body,omitempty"`
	Form     url.Values    `json:"form,omitempty"`
	PostForm url.Values    `json:"postForm,omitempty"`
	Url      *url.URL      `json:"url,omitempty"`
}

type PipelineOperator interface {
	// Pipelinne operator interface
	GetPipeline(projectName, pipelineName string, httpParameters *HttpParameters) (*Pipeline, error)
	ListPipelines(httpParameters *HttpParameters) (*PipelineList, error)
	GetPipelineRun(projectName, pipelineName, runId string, httpParameters *HttpParameters) (*PipelineRun, error)
	ListPipelineRuns(projectName, pipelineName string, httpParameters *HttpParameters) (*PipelineRunList, error)
	StopPipeline(projectName, pipelineName, runId string, httpParameters *HttpParameters) (*StopPipeline, error)
	ReplayPipeline(projectName, pipelineName, runId string, httpParameters *HttpParameters) (*ReplayPipeline, error)
	RunPipeline(projectName, pipelineName string, httpParameters *HttpParameters) (*RunPipeline, error)
	GetArtifacts(projectName, pipelineName, runId string, httpParameters *HttpParameters) ([]Artifacts, error)
	GetRunLog(projectName, pipelineName, runId string, httpParameters *HttpParameters) ([]byte, error)
	GetStepLog(projectName, pipelineName, runId, nodeId, stepId string, httpParameters *HttpParameters) ([]byte, http.Header, error)
	GetNodeSteps(projectName, pipelineName, runId, nodeId string, httpParameters *HttpParameters) ([]NodeSteps, error)
	GetPipelineRunNodes(projectName, pipelineName, runId string, httpParameters *HttpParameters) ([]PipelineRunNodes, error)
	SubmitInputStep(projectName, pipelineName, runId, nodeId, stepId string, httpParameters *HttpParameters) ([]byte, error)

	//BranchPipelinne operator interface
	GetBranchPipeline(projectName, pipelineName, branchName string, httpParameters *HttpParameters) (*BranchPipeline, error)
	GetBranchPipelineRun(projectName, pipelineName, branchName, runId string, httpParameters *HttpParameters) (*PipelineRun, error)
	StopBranchPipeline(projectName, pipelineName, branchName, runId string, httpParameters *HttpParameters) (*StopPipeline, error)
	ReplayBranchPipeline(projectName, pipelineName, branchName, runId string, httpParameters *HttpParameters) (*ReplayPipeline, error)
	RunBranchPipeline(projectName, pipelineName, branchName string, httpParameters *HttpParameters) (*RunPipeline, error)
	GetBranchArtifacts(projectName, pipelineName, branchName, runId string, httpParameters *HttpParameters) ([]Artifacts, error)
	GetBranchRunLog(projectName, pipelineName, branchName, runId string, httpParameters *HttpParameters) ([]byte, error)
	GetBranchStepLog(projectName, pipelineName, branchName, runId, nodeId, stepId string, httpParameters *HttpParameters) ([]byte, http.Header, error)
	GetBranchNodeSteps(projectName, pipelineName, branchName, runId, nodeId string, httpParameters *HttpParameters) ([]NodeSteps, error)
	GetBranchPipelineRunNodes(projectName, pipelineName, branchName, runId string, httpParameters *HttpParameters) ([]BranchPipelineRunNodes, error)
	SubmitBranchInputStep(projectName, pipelineName, branchName, runId, nodeId, stepId string, httpParameters *HttpParameters) ([]byte, error)
	GetPipelineBranch(projectName, pipelineName string, httpParameters *HttpParameters) (*PipelineBranch, error)
	ScanBranch(projectName, pipelineName string, httpParameters *HttpParameters) ([]byte, error)

	// Common pipeline operator interface
	GetConsoleLog(projectName, pipelineName string, httpParameters *HttpParameters) ([]byte, error)
	GetCrumb(httpParameters *HttpParameters) (*Crumb, error)

	// SCM operator interface
	GetSCMServers(scmId string, httpParameters *HttpParameters) ([]SCMServer, error)
	GetSCMOrg(scmId string, httpParameters *HttpParameters) ([]SCMOrg, error)
	GetOrgRepo(scmId, organizationId string, httpParameters *HttpParameters) (OrgRepo, error)
	CreateSCMServers(scmId string, httpParameters *HttpParameters) (*SCMServer, error)
	Validate(scmId string, httpParameters *HttpParameters) (*Validates, error)

	//Webhook operator interface
	GetNotifyCommit(httpParameters *HttpParameters) ([]byte, error)
	GithubWebhook(httpParameters *HttpParameters) ([]byte, error)

	CheckScriptCompile(projectName, pipelineName string, httpParameters *HttpParameters) (*CheckScript, error)
	CheckCron(projectName string, httpParameters *HttpParameters) (*CheckCronRes, error)
	ToJenkinsfile(httpParameters *HttpParameters) (*ResJenkinsfile, error)
	ToJson(httpParameters *HttpParameters) (map[string]interface{}, error)
}
