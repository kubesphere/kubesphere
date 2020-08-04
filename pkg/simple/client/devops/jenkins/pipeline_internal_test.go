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

package jenkins

import (
	devopsv1alpha3 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	"reflect"
	"testing"
)

func Test_NoScmPipelineConfig(t *testing.T) {
	inputs := []*devopsv1alpha3.NoScmPipeline{
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
		},
		{
			Name:        "",
			Description: "",
			Jenkinsfile: "node{echo 'hello'}",
		},
		{
			Name:              "",
			Description:       "",
			Jenkinsfile:       "node{echo 'hello'}",
			DisableConcurrent: true,
		},
	}
	for _, input := range inputs {
		outputString, err := createPipelineConfigXml(input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parsePipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}
}

func Test_NoScmPipelineConfig_Discarder(t *testing.T) {
	inputs := []*devopsv1alpha3.NoScmPipeline{
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			Discarder: &devopsv1alpha3.DiscarderProperty{
				DaysToKeep: "3", NumToKeep: "5",
			},
		},
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			Discarder: &devopsv1alpha3.DiscarderProperty{
				DaysToKeep: "3", NumToKeep: "",
			},
		},
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			Discarder: &devopsv1alpha3.DiscarderProperty{
				DaysToKeep: "", NumToKeep: "21321",
			},
		},
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			Discarder: &devopsv1alpha3.DiscarderProperty{
				DaysToKeep: "", NumToKeep: "",
			},
		},
	}
	for _, input := range inputs {
		outputString, err := createPipelineConfigXml(input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parsePipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}
}

func Test_NoScmPipelineConfig_Param(t *testing.T) {
	inputs := []*devopsv1alpha3.NoScmPipeline{
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			Parameters: []devopsv1alpha3.Parameter{
				{
					Name:         "d",
					DefaultValue: "a\nb",
					Type:         "choice",
					Description:  "fortest",
				},
			},
		},
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			Parameters: []devopsv1alpha3.Parameter{
				{
					Name:         "a",
					DefaultValue: "abc",
					Type:         "string",
					Description:  "fortest",
				},
				{
					Name:         "b",
					DefaultValue: "false",
					Type:         "boolean",
					Description:  "fortest",
				},
				{
					Name:         "c",
					DefaultValue: "password \n aaa",
					Type:         "text",
					Description:  "fortest",
				},
				{
					Name:         "d",
					DefaultValue: "a\nb",
					Type:         "choice",
					Description:  "fortest",
				},
			},
		},
	}
	for _, input := range inputs {
		outputString, err := createPipelineConfigXml(input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parsePipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}
}

func Test_NoScmPipelineConfig_Trigger(t *testing.T) {
	inputs := []*devopsv1alpha3.NoScmPipeline{
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			TimerTrigger: &devopsv1alpha3.TimerTrigger{
				Cron: "1 1 1 * * *",
			},
		},

		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			RemoteTrigger: &devopsv1alpha3.RemoteTrigger{
				Token: "abc",
			},
		},
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			TimerTrigger: &devopsv1alpha3.TimerTrigger{
				Cron: "1 1 1 * * *",
			},
			RemoteTrigger: &devopsv1alpha3.RemoteTrigger{
				Token: "abc",
			},
		},
	}

	for _, input := range inputs {
		outputString, err := createPipelineConfigXml(input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parsePipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}
}

func Test_MultiBranchPipelineConfig(t *testing.T) {

	inputs := []*devopsv1alpha3.MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "git",
			GitSource:   &devopsv1alpha3.GitSource{},
		},
		{
			Name:         "",
			Description:  "for test",
			ScriptPath:   "Jenkinsfile",
			SourceType:   "github",
			GitHubSource: &devopsv1alpha3.GithubSource{},
		},
		{
			Name:            "",
			Description:     "for test",
			ScriptPath:      "Jenkinsfile",
			SourceType:      "single_svn",
			SingleSvnSource: &devopsv1alpha3.SingleSvnSource{},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "svn",
			SvnSource:   &devopsv1alpha3.SvnSource{},
		},
	}
	for _, input := range inputs {
		outputString, err := createMultiBranchPipelineConfigXml("", input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parseMultiBranchPipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}
}

func Test_MultiBranchPipelineConfig_Discarder(t *testing.T) {

	inputs := []*devopsv1alpha3.MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "git",
			Discarder: &devopsv1alpha3.DiscarderProperty{
				DaysToKeep: "1",
				NumToKeep:  "2",
			},
			GitSource: &devopsv1alpha3.GitSource{},
		},
	}
	for _, input := range inputs {
		outputString, err := createMultiBranchPipelineConfigXml("", input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parseMultiBranchPipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}
}

func Test_MultiBranchPipelineConfig_TimerTrigger(t *testing.T) {
	inputs := []*devopsv1alpha3.MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "git",
			TimerTrigger: &devopsv1alpha3.TimerTrigger{
				Interval: "12345566",
			},
			GitSource: &devopsv1alpha3.GitSource{},
		},
	}
	for _, input := range inputs {
		outputString, err := createMultiBranchPipelineConfigXml("", input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parseMultiBranchPipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}
}

func Test_MultiBranchPipelineConfig_Source(t *testing.T) {

	inputs := []*devopsv1alpha3.MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "git",
			TimerTrigger: &devopsv1alpha3.TimerTrigger{
				Interval: "12345566",
			},
			GitSource: &devopsv1alpha3.GitSource{
				Url:              "https://github.com/kubesphere/devops",
				CredentialId:     "git",
				DiscoverBranches: true,
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "github",
			TimerTrigger: &devopsv1alpha3.TimerTrigger{
				Interval: "12345566",
			},
			GitHubSource: &devopsv1alpha3.GithubSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "bitbucket_server",
			TimerTrigger: &devopsv1alpha3.TimerTrigger{
				Interval: "12345566",
			},
			BitbucketServerSource: &devopsv1alpha3.BitbucketServerSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
			},
		},

		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "svn",
			TimerTrigger: &devopsv1alpha3.TimerTrigger{
				Interval: "12345566",
			},
			SvnSource: &devopsv1alpha3.SvnSource{
				Remote:       "https://api.svn.com/bcd",
				CredentialId: "svn",
				Excludes:     "truck",
				Includes:     "tag/*",
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "single_svn",
			TimerTrigger: &devopsv1alpha3.TimerTrigger{
				Interval: "12345566",
			},
			SingleSvnSource: &devopsv1alpha3.SingleSvnSource{
				Remote:       "https://api.svn.com/bcd",
				CredentialId: "svn",
			},
		},
	}

	for _, input := range inputs {
		outputString, err := createMultiBranchPipelineConfigXml("", input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parseMultiBranchPipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}
}

func Test_MultiBranchPipelineCloneConfig(t *testing.T) {

	inputs := []*devopsv1alpha3.MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "git",
			GitSource: &devopsv1alpha3.GitSource{
				Url:              "https://github.com/kubesphere/devops",
				CredentialId:     "git",
				DiscoverBranches: true,
				CloneOption: &devopsv1alpha3.GitCloneOption{
					Shallow: false,
					Depth:   3,
					Timeout: 20,
				},
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "github",
			GitHubSource: &devopsv1alpha3.GithubSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
				CloneOption: &devopsv1alpha3.GitCloneOption{
					Shallow: false,
					Depth:   3,
					Timeout: 20,
				},
			},
		},
	}

	for _, input := range inputs {
		outputString, err := createMultiBranchPipelineConfigXml("", input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parseMultiBranchPipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}

}

func Test_MultiBranchPipelineRegexFilter(t *testing.T) {

	inputs := []*devopsv1alpha3.MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "git",
			GitSource: &devopsv1alpha3.GitSource{
				Url:              "https://github.com/kubesphere/devops",
				CredentialId:     "git",
				DiscoverBranches: true,
				RegexFilter:      ".*",
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "github",
			GitHubSource: &devopsv1alpha3.GithubSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
				RegexFilter: ".*",
			},
		},
	}

	for _, input := range inputs {
		outputString, err := createMultiBranchPipelineConfigXml("", input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parseMultiBranchPipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}

}

func Test_MultiBranchPipelineMultibranchTrigger(t *testing.T) {

	inputs := []*devopsv1alpha3.MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "github",
			GitHubSource: &devopsv1alpha3.GithubSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
				RegexFilter: ".*",
			},
			MultiBranchJobTrigger: &devopsv1alpha3.MultiBranchJobTrigger{
				CreateActionJobsToTrigger: "abc",
				DeleteActionJobsToTrigger: "ddd",
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "github",
			GitHubSource: &devopsv1alpha3.GithubSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
				RegexFilter: ".*",
			},
			MultiBranchJobTrigger: &devopsv1alpha3.MultiBranchJobTrigger{
				CreateActionJobsToTrigger: "abc",
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "github",
			GitHubSource: &devopsv1alpha3.GithubSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &devopsv1alpha3.DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
				RegexFilter: ".*",
			},
			MultiBranchJobTrigger: &devopsv1alpha3.MultiBranchJobTrigger{
				DeleteActionJobsToTrigger: "ddd",
			},
		},
	}

	for _, input := range inputs {
		outputString, err := createMultiBranchPipelineConfigXml("", input)
		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		output, err := parseMultiBranchPipelineConfigXml(outputString)

		if err != nil {
			t.Fatalf("should not get error %+v", err)
		}
		if !reflect.DeepEqual(input, output) {
			t.Fatalf("input [%+v] output [%+v] should equal ", input, output)
		}
	}

}
