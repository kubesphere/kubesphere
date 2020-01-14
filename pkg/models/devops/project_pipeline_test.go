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

import (
	"reflect"
	"testing"
)

func Test_NoScmPipelineConfig(t *testing.T) {
	inputs := []*NoScmPipeline{
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
	inputs := []*NoScmPipeline{
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			Discarder: &DiscarderProperty{
				"3", "5",
			},
		},
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			Discarder: &DiscarderProperty{
				"3", "",
			},
		},
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			Discarder: &DiscarderProperty{
				"", "21321",
			},
		},
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			Discarder: &DiscarderProperty{
				"", "",
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
	inputs := []*NoScmPipeline{
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			Parameters: &Parameters{
				&Parameter{
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
			Parameters: &Parameters{
				&Parameter{
					Name:         "a",
					DefaultValue: "abc",
					Type:         "string",
					Description:  "fortest",
				},
				&Parameter{
					Name:         "b",
					DefaultValue: "false",
					Type:         "boolean",
					Description:  "fortest",
				},
				&Parameter{
					Name:         "c",
					DefaultValue: "password \n aaa",
					Type:         "text",
					Description:  "fortest",
				},
				&Parameter{
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
	inputs := []*NoScmPipeline{
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			TimerTrigger: &TimerTrigger{
				Cron: "1 1 1 * * *",
			},
		},

		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			RemoteTrigger: &RemoteTrigger{
				Token: "abc",
			},
		},
		{
			Name:        "",
			Description: "for test",
			Jenkinsfile: "node{echo 'hello'}",
			TimerTrigger: &TimerTrigger{
				Cron: "1 1 1 * * *",
			},
			RemoteTrigger: &RemoteTrigger{
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

	inputs := []*MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "git",
			GitSource:   &GitSource{},
		},
		{
			Name:         "",
			Description:  "for test",
			ScriptPath:   "Jenkinsfile",
			SourceType:   "github",
			GitHubSource: &GithubSource{},
		},
		{
			Name:            "",
			Description:     "for test",
			ScriptPath:      "Jenkinsfile",
			SourceType:      "single_svn",
			SingleSvnSource: &SingleSvnSource{},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "svn",
			SvnSource:   &SvnSource{},
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

	inputs := []*MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "git",
			Discarder: &DiscarderProperty{
				DaysToKeep: "1",
				NumToKeep:  "2",
			},
			GitSource: &GitSource{},
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
	inputs := []*MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "git",
			TimerTrigger: &TimerTrigger{
				Interval: "12345566",
			},
			GitSource: &GitSource{},
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

	inputs := []*MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "git",
			TimerTrigger: &TimerTrigger{
				Interval: "12345566",
			},
			GitSource: &GitSource{
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
			TimerTrigger: &TimerTrigger{
				Interval: "12345566",
			},
			GitHubSource: &GithubSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &DiscoverPRFromForks{
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
			TimerTrigger: &TimerTrigger{
				Interval: "12345566",
			},
			BitbucketServerSource: &BitbucketServerSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &DiscoverPRFromForks{
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
			TimerTrigger: &TimerTrigger{
				Interval: "12345566",
			},
			SvnSource: &SvnSource{
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
			TimerTrigger: &TimerTrigger{
				Interval: "12345566",
			},
			SingleSvnSource: &SingleSvnSource{
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

	inputs := []*MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "git",
			GitSource: &GitSource{
				Url:              "https://github.com/kubesphere/devops",
				CredentialId:     "git",
				DiscoverBranches: true,
				CloneOption: &GitCloneOption{
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
			GitHubSource: &GithubSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
				CloneOption: &GitCloneOption{
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

	inputs := []*MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "git",
			GitSource: &GitSource{
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
			GitHubSource: &GithubSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &DiscoverPRFromForks{
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

	inputs := []*MultiBranchPipeline{
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "github",
			GitHubSource: &GithubSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
				RegexFilter: ".*",
			},
			MultiBranchJobTrigger: &MultiBranchJobTrigger{
				CreateActionJobsToTrigger: "abc",
				DeleteActionJobsToTrigger: "ddd",
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "github",
			GitHubSource: &GithubSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
				RegexFilter: ".*",
			},
			MultiBranchJobTrigger: &MultiBranchJobTrigger{
				CreateActionJobsToTrigger: "abc",
			},
		},
		{
			Name:        "",
			Description: "for test",
			ScriptPath:  "Jenkinsfile",
			SourceType:  "github",
			GitHubSource: &GithubSource{
				Owner:                "kubesphere",
				Repo:                 "devops",
				CredentialId:         "github",
				ApiUri:               "https://api.github.com",
				DiscoverBranches:     1,
				DiscoverPRFromOrigin: 2,
				DiscoverPRFromForks: &DiscoverPRFromForks{
					Strategy: 1,
					Trust:    1,
				},
				RegexFilter: ".*",
			},
			MultiBranchJobTrigger: &MultiBranchJobTrigger{
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

func Test_ParsePipelineConfigChoiceParameters(t *testing.T) {

	case1 := `<flow-definition plugin="workflow-job@2.32">
<actions>
<org.jenkinsci.plugins.pipeline.modeldefinition.actions.DeclarativeJobAction plugin="pipeline-model-definition@1.3.9"/>
<org.jenkinsci.plugins.pipeline.modeldefinition.actions.DeclarativeJobPropertyTrackerAction plugin="pipeline-model-definition@1.3.9">
<jobProperties/>
<triggers/>
<parameters>
<string>BIOGRAPHY</string>
<string>PERSON</string>
<string>PASSWORD</string>
<string>TOGGLE</string>
</parameters>
<options/>
</org.jenkinsci.plugins.pipeline.modeldefinition.actions.DeclarativeJobPropertyTrackerAction>
</actions>
<description/>
<keepDependencies>false</keepDependencies>
<properties>
<hudson.plugins.jira.JiraProjectProperty plugin="jira@3.0.8"/>
<jenkins.model.BuildDiscarderProperty>
<strategy class="hudson.tasks.LogRotator">
<daysToKeep>7</daysToKeep>
<numToKeep>10</numToKeep>
<artifactDaysToKeep>-1</artifactDaysToKeep>
<artifactNumToKeep>-1</artifactNumToKeep>
</strategy>
</jenkins.model.BuildDiscarderProperty>
<hudson.model.ParametersDefinitionProperty>
<parameterDefinitions>
<hudson.model.StringParameterDefinition>
<name>PERSON</name>
<description>Who should I say hello to?</description>
<defaultValue>Mr Jenkins</defaultValue>
<trim>false</trim>
</hudson.model.StringParameterDefinition>
<hudson.model.TextParameterDefinition>
<name>BIOGRAPHY</name>
<description>Enter some information about the person</description>
<defaultValue/>
<trim>false</trim>
</hudson.model.TextParameterDefinition>
<hudson.model.BooleanParameterDefinition>
<name>TOGGLE</name>
<description>Toggle this value</description>
<defaultValue>true</defaultValue>
</hudson.model.BooleanParameterDefinition>
<hudson.model.PasswordParameterDefinition>
<name>PASSWORD</name>
<description>Enter a password</description>
<defaultValue>
{AQAAABAAAAAQHD/HoXKklrRKtlQ5ylr1lgN1dj7im57I8SGfkLIe17s=}
</defaultValue>
</hudson.model.PasswordParameterDefinition>
<hudson.model.ChoiceParameterDefinition>
<name>1</name>
<description>x</description>
<choices class="java.util.Arrays$ArrayList">
<a class="string-array">
<string>1</string>
<string>2</string>
<string>3</string>
</a>
</choices>
</hudson.model.ChoiceParameterDefinition>
</parameterDefinitions>
</hudson.model.ParametersDefinitionProperty>
</properties>
<definition class="org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition" plugin="workflow-cps@2.73">
<script>
pipeline { agent any parameters { string(name: 'PERSON', defaultValue: 'Mr Jenkins', description: 'Who should I say hello to?') text(name: 'BIOGRAPHY', defaultValue: '', description: 'Enter some information about the person') booleanParam(name: 'TOGGLE', defaultValue: true, description: 'Toggle this value') password(name: 'PASSWORD', defaultValue: 'SECRET', description: 'Enter a password') } stages { stage('Example') { steps { echo "Hello ${params.PERSON}" echo "Biography: ${params.BIOGRAPHY}" echo "Toggle: ${params.TOGGLE}" echo "Choice: ${params.CHOICE}" echo "Password: ${params.PASSWORD}" } } } }
</script>
<sandbox>true</sandbox>
</definition>
<triggers/>
<disabled>false</disabled>
</flow-definition>`
	case2 := `
<flow-definition plugin="workflow-job@2.32">
<actions>
<org.jenkinsci.plugins.pipeline.modeldefinition.actions.DeclarativeJobAction plugin="pipeline-model-definition@1.3.9"/>
<org.jenkinsci.plugins.pipeline.modeldefinition.actions.DeclarativeJobPropertyTrackerAction plugin="pipeline-model-definition@1.3.9">
<jobProperties/>
<triggers/>
<parameters>
<string>BIOGRAPHY</string>
<string>PERSON</string>
<string>PASSWORD</string>
<string>TOGGLE</string>
</parameters>
<options/>
</org.jenkinsci.plugins.pipeline.modeldefinition.actions.DeclarativeJobPropertyTrackerAction>
</actions>
<description/>
<keepDependencies>false</keepDependencies>
<properties>
<hudson.plugins.jira.JiraProjectProperty plugin="jira@3.0.8"/>
<jenkins.model.BuildDiscarderProperty>
<strategy class="hudson.tasks.LogRotator">
<daysToKeep>7</daysToKeep>
<numToKeep>10</numToKeep>
<artifactDaysToKeep>-1</artifactDaysToKeep>
<artifactNumToKeep>-1</artifactNumToKeep>
</strategy>
</jenkins.model.BuildDiscarderProperty>
<hudson.model.ParametersDefinitionProperty>
<parameterDefinitions>
<hudson.model.StringParameterDefinition>
<name>PERSON</name>
<description>Who should I say hello to?</description>
<defaultValue>Mr Jenkins</defaultValue>
<trim>false</trim>
</hudson.model.StringParameterDefinition>
<hudson.model.TextParameterDefinition>
<name>BIOGRAPHY</name>
<description>Enter some information about the person</description>
<defaultValue/>
<trim>false</trim>
</hudson.model.TextParameterDefinition>
<hudson.model.BooleanParameterDefinition>
<name>TOGGLE</name>
<description>Toggle this value</description>
<defaultValue>true</defaultValue>
</hudson.model.BooleanParameterDefinition>
<hudson.model.PasswordParameterDefinition>
<name>PASSWORD</name>
<description>Enter a password</description>
<defaultValue>
{AQAAABAAAAAQHD/HoXKklrRKtlQ5ylr1lgN1dj7im57I8SGfkLIe17s=}
</defaultValue>
</hudson.model.PasswordParameterDefinition>
<hudson.model.ChoiceParameterDefinition>
<name>1</name>
<description>x</description>
<choices class="java.util.Arrays$ArrayList">
<string>1</string>
<string>2</string>
<string>3</string>
</choices>
</hudson.model.ChoiceParameterDefinition>
</parameterDefinitions>
</hudson.model.ParametersDefinitionProperty>
</properties>
<definition class="org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition" plugin="workflow-cps@2.73">
<script>
pipeline { agent any parameters { string(name: 'PERSON', defaultValue: 'Mr Jenkins', description: 'Who should I say hello to?') text(name: 'BIOGRAPHY', defaultValue: '', description: 'Enter some information about the person') booleanParam(name: 'TOGGLE', defaultValue: true, description: 'Toggle this value') password(name: 'PASSWORD', defaultValue: 'SECRET', description: 'Enter a password') } stages { stage('Example') { steps { echo "Hello ${params.PERSON}" echo "Biography: ${params.BIOGRAPHY}" echo "Toggle: ${params.TOGGLE}" echo "Choice: ${params.CHOICE}" echo "Password: ${params.PASSWORD}" } } } }
</script>
<sandbox>true</sandbox>
</definition>
<triggers/>
<disabled>false</disabled>
</flow-definition>`

	case1Pipeline, err := parsePipelineConfigXml(case1)
	if err != nil {
		t.Fatal(err)
	}
	case2Pipeline, err := parsePipelineConfigXml(case2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(case1Pipeline, case2Pipeline) {
		t.Fatalf("case1 [%+v] case2 [%+v] should equal ", case1Pipeline, case2Pipeline)
	}

}
