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
	"fmt"
	"github.com/beevik/etree"
	"github.com/kubesphere/sonargo/sonar"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/gojenkins"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"strconv"
	"strings"
	"time"
)

const (
	NoScmPipelineType       = "pipeline"
	MultiBranchPipelineType = "multi-branch-pipeline"
)

type Parameters []*Parameter

var ParameterTypeMap = map[string]string{
	"hudson.model.StringParameterDefinition":   "string",
	"hudson.model.ChoiceParameterDefinition":   "choice",
	"hudson.model.TextParameterDefinition":     "text",
	"hudson.model.BooleanParameterDefinition":  "boolean",
	"hudson.model.FileParameterDefinition":     "file",
	"hudson.model.PasswordParameterDefinition": "password",
}

const (
	SonarAnalysisActionClass = "hudson.plugins.sonar.action.SonarAnalysisAction"
	SonarMetricKeys          = "alert_status,quality_gate_details,bugs,new_bugs,reliability_rating,new_reliability_rating,vulnerabilities,new_vulnerabilities,security_rating,new_security_rating,code_smells,new_code_smells,sqale_rating,new_maintainability_rating,sqale_index,new_technical_debt,coverage,new_coverage,new_lines_to_cover,tests,duplicated_lines_density,new_duplicated_lines_density,duplicated_blocks,ncloc,ncloc_language_distribution,projects,new_lines"
	SonarAdditionalFields    = "metrics,periods"
)

type SonarStatus struct {
	Measures      *sonargo.MeasuresComponentObject `json:"measures,omitempty"`
	Issues        *sonargo.IssuesSearchObject      `json:"issues,omitempty"`
	JenkinsAction *gojenkins.GeneralObj            `json:"jenkinsAction,omitempty"`
	Task          *sonargo.CeTaskObject            `json:"task,omitempty"`
}

type ProjectPipeline struct {
	Type                string               `json:"type" description:"type of devops pipeline, in scm or no scm"`
	Pipeline            *NoScmPipeline       `json:"pipeline,omitempty" description:"no scm pipeline structs"`
	MultiBranchPipeline *MultiBranchPipeline `json:"multi_branch_pipeline,omitempty" description:"in scm pipeline structs"`
}

type NoScmPipeline struct {
	Name              string             `json:"name" description:"name of pipeline"`
	Description       string             `json:"descriptio,omitempty" description:"description of pipeline"`
	Discarder         *DiscarderProperty `json:"discarder,omitempty" description:"Discarder of pipeline, managing when to drop a pipeline"`
	Parameters        *Parameters        `json:"parameters,omitempty" description:"Parameters define of pipeline,user could pass param when run pipeline"`
	DisableConcurrent bool               `json:"disable_concurrent,omitempty" mapstructure:"disable_concurrent" description:"Whether to prohibit the pipeline from running in parallel"`
	TimerTrigger      *TimerTrigger      `json:"timer_trigger,omitempty" mapstructure:"timer_trigger" description:"Timer to trigger pipeline run"`
	RemoteTrigger     *RemoteTrigger     `json:"remote_trigger,omitempty" mapstructure:"remote_trigger" description:"Remote api define to trigger pipeline run"`
	Jenkinsfile       string             `json:"jenkinsfile,omitempty" description:"Jenkinsfile's content'"`
}

type MultiBranchPipeline struct {
	Name                  string                 `json:"name" description:"name of pipeline"`
	Description           string                 `json:"descriptio,omitempty" description:"description of pipeline"`
	Discarder             *DiscarderProperty     `json:"discarder,omitempty" description:"Discarder of pipeline, managing when to drop a pipeline"`
	TimerTrigger          *TimerTrigger          `json:"timer_trigger,omitempty" mapstructure:"timer_trigger" description:"Timer to trigger pipeline run"`
	SourceType            string                 `json:"source_type" description:"type of scm, such as github/git/svn"`
	GitSource             *GitSource             `json:"git_source,omitempty" description:"git scm define"`
	GitHubSource          *GithubSource          `json:"github_source,omitempty" description:"github scm define"`
	SvnSource             *SvnSource             `json:"svn_source,omitempty" description:"multi branch svn scm define"`
	SingleSvnSource       *SingleSvnSource       `json:"single_svn_source,omitempty" description:"single branch svn scm define"`
	BitbucketServerSource *BitbucketServerSource `json:"bitbucket_server_source,omitempty" description:"bitbucket server scm defile"`
	ScriptPath            string                 `json:"script_path" mapstructure:"script_path" description:"script path in scm"`
	MultiBranchJobTrigger *MultiBranchJobTrigger `json:"multibranch_job_trigger,omitempty" mapstructure:"multibranch_job_trigger" description:"Pipeline tasks that need to be triggered when branch creation/deletion"`
}

type GitSource struct {
	ScmId            string          `json:"scm_id,omitempty" description:"uid of scm"`
	Url              string          `json:"url,omitempty" mapstructure:"url" description:"url of git source"`
	CredentialId     string          `json:"credential_id,omitempty" mapstructure:"credential_id" description:"credential id to access git source"`
	DiscoverBranches bool            `json:"discover_branches,omitempty" mapstructure:"discover_branches" description:"Whether to discover a branch"`
	CloneOption      *GitCloneOption `json:"git_clone_option,omitempty" mapstructure:"git_clone_option" description:"advavced git clone options"`
	RegexFilter      string          `json:"regex_filter,omitempty" mapstructure:"regex_filter" description:"Regex used to match the name of the branch that needs to be run"`
}

type GithubSource struct {
	ScmId                string               `json:"scm_id,omitempty" description:"uid of scm"`
	Owner                string               `json:"owner,omitempty" mapstructure:"owner" description:"owner of github repo"`
	Repo                 string               `json:"repo,omitempty" mapstructure:"repo" description:"repo name of github repo"`
	CredentialId         string               `json:"credential_id,omitempty" mapstructure:"credential_id" description:"credential id to access github source"`
	ApiUri               string               `json:"api_uri,omitempty" mapstructure:"api_uri" description:"The api url can specify the location of the github apiserver.For private cloud configuration"`
	DiscoverBranches     int                  `json:"discover_branches,omitempty" mapstructure:"discover_branches" description:"Discover branch configuration"`
	DiscoverPRFromOrigin int                  `json:"discover_pr_from_origin,omitempty" mapstructure:"discover_pr_from_origin" description:"Discover origin PR configuration"`
	DiscoverPRFromForks  *DiscoverPRFromForks `json:"discover_pr_from_forks,omitempty" mapstructure:"discover_pr_from_forks" description:"Discover fork PR configuration"`
	CloneOption          *GitCloneOption      `json:"git_clone_option,omitempty" mapstructure:"git_clone_option" description:"advavced git clone options"`
	RegexFilter          string               `json:"regex_filter,omitempty" mapstructure:"regex_filter" description:"Regex used to match the name of the branch that needs to be run"`
}

type MultiBranchJobTrigger struct {
	CreateActionJobsToTrigger string `json:"create_action_job_to_trigger,omitempty" description:"pipeline name to trigger"`
	DeleteActionJobsToTrigger string `json:"delete_action_job_to_trigger,omitempty" description:"pipeline name to trigger"`
}

type BitbucketServerSource struct {
	ScmId                string               `json:"scm_id,omitempty" description:"uid of scm"`
	Owner                string               `json:"owner,omitempty" mapstructure:"owner" description:"owner of github repo"`
	Repo                 string               `json:"repo,omitempty" mapstructure:"repo" description:"repo name of github repo"`
	CredentialId         string               `json:"credential_id,omitempty" mapstructure:"credential_id" description:"credential id to access github source"`
	ApiUri               string               `json:"api_uri,omitempty" mapstructure:"api_uri" description:"The api url can specify the location of the github apiserver.For private cloud configuration"`
	DiscoverBranches     int                  `json:"discover_branches,omitempty" mapstructure:"discover_branches" description:"Discover branch configuration"`
	DiscoverPRFromOrigin int                  `json:"discover_pr_from_origin,omitempty" mapstructure:"discover_pr_from_origin" description:"Discover origin PR configuration"`
	DiscoverPRFromForks  *DiscoverPRFromForks `json:"discover_pr_from_forks,omitempty" mapstructure:"discover_pr_from_forks" description:"Discover fork PR configuration"`
	CloneOption          *GitCloneOption      `json:"git_clone_option,omitempty" mapstructure:"git_clone_option" description:"advavced git clone options"`
	RegexFilter          string               `json:"regex_filter,omitempty" mapstructure:"regex_filter" description:"Regex used to match the name of the branch that needs to be run"`
}

type GitCloneOption struct {
	Shallow bool `json:"shallow,omitempty" mapstructure:"shallow" description:"Whether to use git shallow clone"`
	Timeout int  `json:"timeout,omitempty" mapstructure:"timeout" description:"git clone timeout mins"`
	Depth   int  `json:"depth,omitempty" mapstructure:"depth" description:"git clone depth"`
}

type SvnSource struct {
	ScmId        string `json:"scm_id,omitempty" description:"uid of scm"`
	Remote       string `json:"remote,omitempty" description:"remote address url"`
	CredentialId string `json:"credential_id,omitempty" mapstructure:"credential_id" description:"credential id to access svn source"`
	Includes     string `json:"includes,omitempty" description:"branches to run pipeline"`
	Excludes     string `json:"excludes,omitempty" description:"branches do not run pipeline"`
}
type SingleSvnSource struct {
	ScmId        string `json:"scm_id,omitempty" description:"uid of scm"`
	Remote       string `json:"remote,omitempty" description:"remote address url"`
	CredentialId string `json:"credential_id,omitempty" mapstructure:"credential_id" description:"credential id to access svn source"`
}

type DiscoverPRFromForks struct {
	Strategy int `json:"strategy,omitempty" mapstructure:"strategy" description:"github discover strategy"`
	Trust    int `json:"trust,omitempty" mapstructure:"trust" description:"trust user type"`
}

type DiscarderProperty struct {
	DaysToKeep string `json:"days_to_keep,omitempty" mapstructure:"days_to_keep" description:"days to keep pipeline"`
	NumToKeep  string `json:"num_to_keep,omitempty" mapstructure:"num_to_keep" description:"nums to keep pipeline"`
}

type Parameter struct {
	Name         string `json:"name" description:"name of param"`
	DefaultValue string `json:"default_value,omitempty" mapstructure:"default_value" description:"default value of param"`
	Type         string `json:"type" description:"type of param"`
	Description  string `json:"description,omitempty" description:"description of pipeline"`
}

type TimerTrigger struct {
	// user in no scm job
	Cron string `json:"cron,omitempty" description:"jenkins cron script"`

	// use in multi-branch job
	Interval string `json:"interval,omitempty" description:"interval ms"`
}

type RemoteTrigger struct {
	Token string `json:"token,omitempty" description:"remote trigger token"`
}

func replaceXmlVersion(config, oldVersion, targetVersion string) string {
	lines := strings.Split(config, "\n")
	lines[0] = strings.Replace(lines[0], oldVersion, targetVersion, -1)
	output := strings.Join(lines, "\n")
	return output
}

func createPipelineConfigXml(pipeline *NoScmPipeline) (string, error) {
	doc := etree.NewDocument()
	xmlString := `<?xml version='1.0' encoding='UTF-8'?>
<flow-definition plugin="workflow-job">
  <actions>
    <org.jenkinsci.plugins.pipeline.modeldefinition.actions.DeclarativeJobAction plugin="pipeline-model-definition"/>
    <org.jenkinsci.plugins.pipeline.modeldefinition.actions.DeclarativeJobPropertyTrackerAction plugin="pipeline-model-definition">
      <jobProperties/>
      <triggers/>
      <parameters/>
      <options/>
    </org.jenkinsci.plugins.pipeline.modeldefinition.actions.DeclarativeJobPropertyTrackerAction>
  </actions>
</flow-definition>
`
	doc.ReadFromString(xmlString)
	flow := doc.SelectElement("flow-definition")
	flow.CreateElement("description").SetText(pipeline.Description)
	properties := flow.CreateElement("properties")

	if pipeline.DisableConcurrent {
		properties.CreateElement("org.jenkinsci.plugins.workflow.job.properties.DisableConcurrentBuildsJobProperty")
	}

	if pipeline.Discarder != nil {
		discarder := properties.CreateElement("jenkins.model.BuildDiscarderProperty")
		strategy := discarder.CreateElement("strategy")
		strategy.CreateAttr("class", "hudson.tasks.LogRotator")
		strategy.CreateElement("daysToKeep").SetText(pipeline.Discarder.DaysToKeep)
		strategy.CreateElement("numToKeep").SetText(pipeline.Discarder.NumToKeep)
		strategy.CreateElement("artifactDaysToKeep").SetText("-1")
		strategy.CreateElement("artifactNumToKeep").SetText("-1")
	}
	if pipeline.Parameters != nil {
		pipeline.Parameters.appendToEtree(properties)
	}

	if pipeline.TimerTrigger != nil {
		triggers := properties.
			CreateElement("org.jenkinsci.plugins.workflow.job.properties.PipelineTriggersJobProperty").
			CreateElement("triggers")
		triggers.CreateElement("hudson.triggers.TimerTrigger").CreateElement("spec").SetText(pipeline.TimerTrigger.Cron)
	}

	pipelineDefine := flow.CreateElement("definition")
	pipelineDefine.CreateAttr("class", "org.jenkinsci.plugins.workflow.cps.CpsFlowDefinition")
	pipelineDefine.CreateAttr("plugin", "workflow-cps")
	pipelineDefine.CreateElement("script").SetText(pipeline.Jenkinsfile)

	pipelineDefine.CreateElement("sandbox").SetText("true")

	flow.CreateElement("triggers")

	if pipeline.RemoteTrigger != nil {
		flow.CreateElement("authToken").SetText(pipeline.RemoteTrigger.Token)
	}
	flow.CreateElement("disabled").SetText("false")

	doc.Indent(2)
	stringXml, err := doc.WriteToString()
	if err != nil {
		return "", err
	}
	return replaceXmlVersion(stringXml, "1.0", "1.1"), err
}

func parsePipelineConfigXml(config string) (*NoScmPipeline, error) {
	pipeline := &NoScmPipeline{}
	config = replaceXmlVersion(config, "1.1", "1.0")
	doc := etree.NewDocument()
	err := doc.ReadFromString(config)
	if err != nil {
		return nil, err
	}
	flow := doc.SelectElement("flow-definition")
	if flow == nil {
		return nil, fmt.Errorf("can not find pipeline definition")
	}
	pipeline.Description = flow.SelectElement("description").Text()

	properties := flow.SelectElement("properties")
	if properties.
		SelectElement(
			"org.jenkinsci.plugins.workflow.job.properties.DisableConcurrentBuildsJobProperty") != nil {
		pipeline.DisableConcurrent = true
	}
	if properties.SelectElement("jenkins.model.BuildDiscarderProperty") != nil {
		strategy := properties.
			SelectElement("jenkins.model.BuildDiscarderProperty").
			SelectElement("strategy")
		pipeline.Discarder = &DiscarderProperty{
			DaysToKeep: strategy.SelectElement("daysToKeep").Text(),
			NumToKeep:  strategy.SelectElement("numToKeep").Text(),
		}
	}
	pipeline.Parameters = &Parameters{}
	pipeline.Parameters = pipeline.Parameters.fromEtree(properties)
	if len(*pipeline.Parameters) == 0 {
		pipeline.Parameters = nil
	}

	if triggerProperty := properties.
		SelectElement(
			"org.jenkinsci.plugins.workflow.job.properties.PipelineTriggersJobProperty"); triggerProperty != nil {
		triggers := triggerProperty.SelectElement("triggers")
		if timerTrigger := triggers.SelectElement("hudson.triggers.TimerTrigger"); timerTrigger != nil {
			pipeline.TimerTrigger = &TimerTrigger{
				Cron: timerTrigger.SelectElement("spec").Text(),
			}
		}
	}
	if authToken := flow.SelectElement("authToken"); authToken != nil {
		pipeline.RemoteTrigger = &RemoteTrigger{
			Token: authToken.Text(),
		}
	}
	if definition := flow.SelectElement("definition"); definition != nil {
		if script := definition.SelectElement("script"); script != nil {
			pipeline.Jenkinsfile = script.Text()
		}
	}
	return pipeline, nil
}

func (s *Parameters) appendToEtree(properties *etree.Element) *Parameters {
	parameterDefinitions := properties.CreateElement("hudson.model.ParametersDefinitionProperty").
		CreateElement("parameterDefinitions")
	for _, parameter := range *s {
		for className, typeName := range ParameterTypeMap {
			if typeName == parameter.Type {
				paramDefine := parameterDefinitions.CreateElement(className)
				paramDefine.CreateElement("name").SetText(parameter.Name)
				paramDefine.CreateElement("description").SetText(parameter.Description)
				switch parameter.Type {
				case "choice":
					choices := paramDefine.CreateElement("choices")
					choices.CreateAttr("class", "java.util.Arrays$ArrayList")
					a := choices.CreateElement("a")
					a.CreateAttr("class", "string-array")
					choiceValues := strings.Split(parameter.DefaultValue, "\n")
					for _, choiceValue := range choiceValues {
						a.CreateElement("string").SetText(choiceValue)
					}
				case "file":
					break
				default:
					paramDefine.CreateElement("defaultValue").SetText(parameter.DefaultValue)
				}
			}
		}
	}
	return s
}

func (s *Parameters) fromEtree(properties *etree.Element) *Parameters {

	if parametersProperty := properties.SelectElement("hudson.model.ParametersDefinitionProperty"); parametersProperty != nil {
		params := parametersProperty.SelectElement("parameterDefinitions").ChildElements()
		if *s == nil {
			*s = make([]*Parameter, 0)
		}
		for _, param := range params {
			switch param.Tag {
			case "hudson.model.StringParameterDefinition":
				*s = append(*s, &Parameter{
					Name:         param.SelectElement("name").Text(),
					Description:  param.SelectElement("description").Text(),
					DefaultValue: param.SelectElement("defaultValue").Text(),
					Type:         ParameterTypeMap["hudson.model.StringParameterDefinition"],
				})
			case "hudson.model.BooleanParameterDefinition":
				*s = append(*s, &Parameter{
					Name:         param.SelectElement("name").Text(),
					Description:  param.SelectElement("description").Text(),
					DefaultValue: param.SelectElement("defaultValue").Text(),
					Type:         ParameterTypeMap["hudson.model.BooleanParameterDefinition"],
				})
			case "hudson.model.TextParameterDefinition":
				*s = append(*s, &Parameter{
					Name:         param.SelectElement("name").Text(),
					Description:  param.SelectElement("description").Text(),
					DefaultValue: param.SelectElement("defaultValue").Text(),
					Type:         ParameterTypeMap["hudson.model.TextParameterDefinition"],
				})
			case "hudson.model.FileParameterDefinition":
				*s = append(*s, &Parameter{
					Name:        param.SelectElement("name").Text(),
					Description: param.SelectElement("description").Text(),
					Type:        ParameterTypeMap["hudson.model.FileParameterDefinition"],
				})
			case "hudson.model.PasswordParameterDefinition":
				*s = append(*s, &Parameter{
					Name:         param.SelectElement("name").Text(),
					Description:  param.SelectElement("description").Text(),
					DefaultValue: param.SelectElement("name").Text(),
					Type:         ParameterTypeMap["hudson.model.PasswordParameterDefinition"],
				})
			case "hudson.model.ChoiceParameterDefinition":
				choiceParameter := &Parameter{
					Name:        param.SelectElement("name").Text(),
					Description: param.SelectElement("description").Text(),
					Type:        ParameterTypeMap["hudson.model.ChoiceParameterDefinition"],
				}
				choices := param.SelectElement("choices").SelectElement("a").SelectElements("string")
				for _, choice := range choices {
					choiceParameter.DefaultValue += fmt.Sprintf("%s\n", choice.Text())
				}
				choiceParameter.DefaultValue = strings.TrimSpace(choiceParameter.DefaultValue)
				*s = append(*s, choiceParameter)
			default:
				*s = append(*s, &Parameter{
					Name:         param.SelectElement("name").Text(),
					Description:  param.SelectElement("description").Text(),
					DefaultValue: "unknown",
					Type:         param.Tag,
				})
			}
		}
	}
	return s
}

func (s *GitSource) appendToEtree(source *etree.Element) *GitSource {
	source.CreateAttr("class", "jenkins.plugins.git.GitSCMSource")
	source.CreateAttr("plugin", "git")
	source.CreateElement("id").SetText(s.ScmId)
	source.CreateElement("remote").SetText(s.Url)
	if s.CredentialId != "" {
		source.CreateElement("credentialsId").SetText(s.CredentialId)
	}
	traits := source.CreateElement("traits")
	if s.DiscoverBranches {
		traits.CreateElement("jenkins.plugins.git.traits.BranchDiscoveryTrait")
	}
	if s.CloneOption != nil {
		cloneExtension := traits.CreateElement("jenkins.plugins.git.traits.CloneOptionTrait").CreateElement("extension")
		cloneExtension.CreateAttr("class", "hudson.plugins.git.extensions.impl.CloneOption")
		cloneExtension.CreateElement("shallow").SetText(strconv.FormatBool(s.CloneOption.Shallow))
		cloneExtension.CreateElement("noTags").SetText(strconv.FormatBool(false))
		cloneExtension.CreateElement("honorRefspec").SetText(strconv.FormatBool(true))
		cloneExtension.CreateElement("reference")
		if s.CloneOption.Timeout >= 0 {
			cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(s.CloneOption.Timeout))
		} else {
			cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(10))
		}

		if s.CloneOption.Depth >= 0 {
			cloneExtension.CreateElement("depth").SetText(strconv.Itoa(s.CloneOption.Depth))
		} else {
			cloneExtension.CreateElement("depth").SetText(strconv.Itoa(1))
		}
	}

	if s.RegexFilter != "" {
		regexTraits := traits.CreateElement("jenkins.scm.impl.trait.RegexSCMHeadFilterTrait")
		regexTraits.CreateAttr("plugin", "scm-api@2.4.0")
		regexTraits.CreateElement("regex").SetText(s.RegexFilter)
	}
	return s
}

func (s *GitSource) fromEtree(source *etree.Element) *GitSource {
	if credential := source.SelectElement("credentialsId"); credential != nil {
		s.CredentialId = credential.Text()
	}
	if remote := source.SelectElement("remote"); remote != nil {
		s.Url = remote.Text()
	}

	traits := source.SelectElement("traits")
	if branchDiscoverTrait := traits.SelectElement(
		"jenkins.plugins.git.traits.BranchDiscoveryTrait"); branchDiscoverTrait != nil {
		s.DiscoverBranches = true
	}
	if cloneTrait := traits.SelectElement(
		"jenkins.plugins.git.traits.CloneOptionTrait"); cloneTrait != nil {
		if cloneExtension := cloneTrait.SelectElement(
			"extension"); cloneExtension != nil {
			s.CloneOption = &GitCloneOption{}
			if value, err := strconv.ParseBool(cloneExtension.SelectElement("shallow").Text()); err == nil {
				s.CloneOption.Shallow = value
			}
			if value, err := strconv.ParseInt(cloneExtension.SelectElement("timeout").Text(), 10, 32); err == nil {
				s.CloneOption.Timeout = int(value)
			}
			if value, err := strconv.ParseInt(cloneExtension.SelectElement("depth").Text(), 10, 32); err == nil {
				s.CloneOption.Depth = int(value)
			}
		}
	}
	if regexTrait := traits.SelectElement(
		"jenkins.scm.impl.trait.RegexSCMHeadFilterTrait"); regexTrait != nil {
		if regex := regexTrait.SelectElement("regex"); regex != nil {
			s.RegexFilter = regex.Text()
		}
	}
	return s
}

func (s *GithubSource) fromEtree(source *etree.Element) *GithubSource {
	if credential := source.SelectElement("credentialsId"); credential != nil {
		s.CredentialId = credential.Text()
	}
	if repoOwner := source.SelectElement("repoOwner"); repoOwner != nil {
		s.Owner = repoOwner.Text()
	}
	if repository := source.SelectElement("repository"); repository != nil {
		s.Repo = repository.Text()
	}
	if apiUri := source.SelectElement("apiUri"); apiUri != nil {
		s.ApiUri = apiUri.Text()
	}
	traits := source.SelectElement("traits")
	if branchDiscoverTrait := traits.SelectElement(
		"org.jenkinsci.plugins.github__branch__source.BranchDiscoveryTrait"); branchDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(branchDiscoverTrait.SelectElement("strategyId").Text())
		s.DiscoverBranches = strategyId
	}
	if originPRDiscoverTrait := traits.SelectElement(
		"org.jenkinsci.plugins.github__branch__source.OriginPullRequestDiscoveryTrait"); originPRDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(originPRDiscoverTrait.SelectElement("strategyId").Text())
		s.DiscoverPRFromOrigin = strategyId
	}
	if forkPRDiscoverTrait := traits.SelectElement(
		"org.jenkinsci.plugins.github__branch__source.ForkPullRequestDiscoveryTrait"); forkPRDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(forkPRDiscoverTrait.SelectElement("strategyId").Text())
		trustClass := forkPRDiscoverTrait.SelectElement("trust").SelectAttr("class").Value
		trust := strings.Split(trustClass, "$")
		switch trust[1] {
		case "TrustContributors":
			s.DiscoverPRFromForks = &DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    1,
			}
		case "TrustEveryone":
			s.DiscoverPRFromForks = &DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    2,
			}
		case "TrustPermission":
			s.DiscoverPRFromForks = &DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    3,
			}
		case "TrustNobody":
			s.DiscoverPRFromForks = &DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    4,
			}
		}
		if cloneTrait := traits.SelectElement(
			"jenkins.plugins.git.traits.CloneOptionTrait"); cloneTrait != nil {
			if cloneExtension := cloneTrait.SelectElement(
				"extension"); cloneExtension != nil {
				s.CloneOption = &GitCloneOption{}
				if value, err := strconv.ParseBool(cloneExtension.SelectElement("shallow").Text()); err == nil {
					s.CloneOption.Shallow = value
				}
				if value, err := strconv.ParseInt(cloneExtension.SelectElement("timeout").Text(), 10, 32); err == nil {
					s.CloneOption.Timeout = int(value)
				}
				if value, err := strconv.ParseInt(cloneExtension.SelectElement("depth").Text(), 10, 32); err == nil {
					s.CloneOption.Depth = int(value)
				}
			}
		}

		if regexTrait := traits.SelectElement(
			"jenkins.scm.impl.trait.RegexSCMHeadFilterTrait"); regexTrait != nil {
			if regex := regexTrait.SelectElement("regex"); regex != nil {
				s.RegexFilter = regex.Text()
			}
		}
	}
	return s
}

func (s *GithubSource) appendToEtree(source *etree.Element) *GithubSource {
	source.CreateAttr("class", "org.jenkinsci.plugins.github_branch_source.GitHubSCMSource")
	source.CreateAttr("plugin", "github-branch-source")
	source.CreateElement("id").SetText(s.ScmId)
	source.CreateElement("credentialsId").SetText(s.CredentialId)
	source.CreateElement("repoOwner").SetText(s.Owner)
	source.CreateElement("repository").SetText(s.Repo)
	if s.ApiUri != "" {
		source.CreateElement("apiUri").SetText(s.ApiUri)
	}
	traits := source.CreateElement("traits")
	if s.DiscoverBranches != 0 {
		traits.CreateElement("org.jenkinsci.plugins.github__branch__source.BranchDiscoveryTrait").
			CreateElement("strategyId").SetText(strconv.Itoa(s.DiscoverBranches))
	}
	if s.DiscoverPRFromOrigin != 0 {
		traits.CreateElement("org.jenkinsci.plugins.github__branch__source.OriginPullRequestDiscoveryTrait").
			CreateElement("strategyId").SetText(strconv.Itoa(s.DiscoverPRFromOrigin))
	}
	if s.DiscoverPRFromForks != nil {
		forkTrait := traits.CreateElement("org.jenkinsci.plugins.github__branch__source.ForkPullRequestDiscoveryTrait")
		forkTrait.CreateElement("strategyId").SetText(strconv.Itoa(s.DiscoverPRFromForks.Strategy))
		trustClass := "org.jenkinsci.plugins.github_branch_source.ForkPullRequestDiscoveryTrait$"
		switch s.DiscoverPRFromForks.Trust {
		case 1:
			trustClass += "TrustContributors"
		case 2:
			trustClass += "TrustEveryone"
		case 3:
			trustClass += "TrustPermission"
		case 4:
			trustClass += "TrustNobody"
		default:
		}
		forkTrait.CreateElement("trust").CreateAttr("class", trustClass)
	}
	if s.CloneOption != nil {
		cloneExtension := traits.CreateElement("jenkins.plugins.git.traits.CloneOptionTrait").CreateElement("extension")
		cloneExtension.CreateAttr("class", "hudson.plugins.git.extensions.impl.CloneOption")
		cloneExtension.CreateElement("shallow").SetText(strconv.FormatBool(s.CloneOption.Shallow))
		cloneExtension.CreateElement("noTags").SetText(strconv.FormatBool(false))
		cloneExtension.CreateElement("honorRefspec").SetText(strconv.FormatBool(true))
		cloneExtension.CreateElement("reference")
		if s.CloneOption.Timeout >= 0 {
			cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(s.CloneOption.Timeout))
		} else {
			cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(10))
		}

		if s.CloneOption.Depth >= 0 {
			cloneExtension.CreateElement("depth").SetText(strconv.Itoa(s.CloneOption.Depth))
		} else {
			cloneExtension.CreateElement("depth").SetText(strconv.Itoa(1))
		}
	}
	if s.RegexFilter != "" {
		regexTraits := traits.CreateElement("jenkins.scm.impl.trait.RegexSCMHeadFilterTrait")
		regexTraits.CreateAttr("plugin", "scm-api@2.4.0")
		regexTraits.CreateElement("regex").SetText(s.RegexFilter)
	}
	return s
}

func (s *BitbucketServerSource) fromEtree(source *etree.Element) *BitbucketServerSource {
	if credential := source.SelectElement("credentialsId"); credential != nil {
		s.CredentialId = credential.Text()
	}
	if repoOwner := source.SelectElement("repoOwner"); repoOwner != nil {
		s.Owner = repoOwner.Text()
	}
	if repository := source.SelectElement("repository"); repository != nil {
		s.Repo = repository.Text()
	}
	if apiUri := source.SelectElement("serverUrl"); apiUri != nil {
		s.ApiUri = apiUri.Text()
	}
	traits := source.SelectElement("traits")
	if branchDiscoverTrait := traits.SelectElement(
		"com.cloudbees.jenkins.plugins.bitbucket.BranchDiscoveryTrait"); branchDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(branchDiscoverTrait.SelectElement("strategyId").Text())
		s.DiscoverBranches = strategyId
	}
	if originPRDiscoverTrait := traits.SelectElement(
		"com.cloudbees.jenkins.plugins.bitbucket.OriginPullRequestDiscoveryTrait"); originPRDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(originPRDiscoverTrait.SelectElement("strategyId").Text())
		s.DiscoverPRFromOrigin = strategyId
	}
	if forkPRDiscoverTrait := traits.SelectElement(
		"com.cloudbees.jenkins.plugins.bitbucket.ForkPullRequestDiscoveryTrait"); forkPRDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(forkPRDiscoverTrait.SelectElement("strategyId").Text())
		trustClass := forkPRDiscoverTrait.SelectElement("trust").SelectAttr("class").Value
		trust := strings.Split(trustClass, "$")
		switch trust[1] {
		case "TrustEveryone":
			s.DiscoverPRFromForks = &DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    1,
			}
		case "TrustTeamForks":
			s.DiscoverPRFromForks = &DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    2,
			}
		case "TrustNobody":
			s.DiscoverPRFromForks = &DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    3,
			}
		}
		if cloneTrait := traits.SelectElement(
			"jenkins.plugins.git.traits.CloneOptionTrait"); cloneTrait != nil {
			if cloneExtension := cloneTrait.SelectElement(
				"extension"); cloneExtension != nil {
				s.CloneOption = &GitCloneOption{}
				if value, err := strconv.ParseBool(cloneExtension.SelectElement("shallow").Text()); err == nil {
					s.CloneOption.Shallow = value
				}
				if value, err := strconv.ParseInt(cloneExtension.SelectElement("timeout").Text(), 10, 32); err == nil {
					s.CloneOption.Timeout = int(value)
				}
				if value, err := strconv.ParseInt(cloneExtension.SelectElement("depth").Text(), 10, 32); err == nil {
					s.CloneOption.Depth = int(value)
				}
			}
		}

		if regexTrait := traits.SelectElement(
			"jenkins.scm.impl.trait.RegexSCMHeadFilterTrait"); regexTrait != nil {
			if regex := regexTrait.SelectElement("regex"); regex != nil {
				s.RegexFilter = regex.Text()
			}
		}
	}
	return s
}

func (s *BitbucketServerSource) appendToEtree(source *etree.Element) *BitbucketServerSource {
	source.CreateAttr("class", "com.cloudbees.jenkins.plugins.bitbucket.BitbucketSCMSource")
	source.CreateAttr("plugin", "cloudbees-bitbucket-branch-source")
	source.CreateElement("id").SetText(s.ScmId)
	source.CreateElement("credentialsId").SetText(s.CredentialId)
	source.CreateElement("repoOwner").SetText(s.Owner)
	source.CreateElement("repository").SetText(s.Repo)
	source.CreateElement("serverUrl").SetText(s.ApiUri)

	traits := source.CreateElement("traits")
	if s.DiscoverBranches != 0 {
		traits.CreateElement("com.cloudbees.jenkins.plugins.bitbucket.BranchDiscoveryTrait>").
			CreateElement("strategyId").SetText(strconv.Itoa(s.DiscoverBranches))
	}
	if s.DiscoverPRFromOrigin != 0 {
		traits.CreateElement("com.cloudbees.jenkins.plugins.bitbucket.OriginPullRequestDiscoveryTrait").
			CreateElement("strategyId").SetText(strconv.Itoa(s.DiscoverPRFromOrigin))
	}
	if s.DiscoverPRFromForks != nil {
		forkTrait := traits.CreateElement("com.cloudbees.jenkins.plugins.bitbucket.ForkPullRequestDiscoveryTrait")
		forkTrait.CreateElement("strategyId").SetText(strconv.Itoa(s.DiscoverPRFromForks.Strategy))
		trustClass := "com.cloudbees.jenkins.plugins.bitbucket.ForkPullRequestDiscoveryTrait$"
		switch s.DiscoverPRFromForks.Trust {
		case 1:
			trustClass += "TrustEveryone"
		case 2:
			trustClass += "TrustTeamForks"
		case 3:
			trustClass += "TrustNobody"
		default:
			trustClass += "TrustEveryone"
		}
		forkTrait.CreateElement("trust").CreateAttr("class", trustClass)
	}
	if s.CloneOption != nil {
		cloneExtension := traits.CreateElement("jenkins.plugins.git.traits.CloneOptionTrait").CreateElement("extension")
		cloneExtension.CreateAttr("class", "hudson.plugins.git.extensions.impl.CloneOption")
		cloneExtension.CreateElement("shallow").SetText(strconv.FormatBool(s.CloneOption.Shallow))
		cloneExtension.CreateElement("noTags").SetText(strconv.FormatBool(false))
		cloneExtension.CreateElement("honorRefspec").SetText(strconv.FormatBool(true))
		cloneExtension.CreateElement("reference")
		if s.CloneOption.Timeout >= 0 {
			cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(s.CloneOption.Timeout))
		} else {
			cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(10))
		}

		if s.CloneOption.Depth >= 0 {
			cloneExtension.CreateElement("depth").SetText(strconv.Itoa(s.CloneOption.Depth))
		} else {
			cloneExtension.CreateElement("depth").SetText(strconv.Itoa(1))
		}
	}
	if s.RegexFilter != "" {
		regexTraits := traits.CreateElement("jenkins.scm.impl.trait.RegexSCMHeadFilterTrait")
		regexTraits.CreateAttr("plugin", "scm-api@2.4.0")
		regexTraits.CreateElement("regex").SetText(s.RegexFilter)
	}
	return s
}

func (s *SvnSource) fromEtree(source *etree.Element) *SvnSource {
	if remote := source.SelectElement("remoteBase"); remote != nil {
		s.Remote = remote.Text()
	}

	if credentialsId := source.SelectElement("credentialsId"); credentialsId != nil {
		s.CredentialId = credentialsId.Text()
	}

	if includes := source.SelectElement("includes"); includes != nil {
		s.Includes = includes.Text()
	}

	if excludes := source.SelectElement("excludes"); excludes != nil {
		s.Excludes = excludes.Text()
	}
	return s
}

func (s *SvnSource) appendToEtree(source *etree.Element) *SvnSource {
	source.CreateAttr("class", "jenkins.scm.impl.subversion.SubversionSCMSource")
	source.CreateAttr("plugin", "subversion")
	source.CreateElement("id").SetText(s.ScmId)
	if s.CredentialId != "" {
		source.CreateElement("credentialsId").SetText(s.CredentialId)
	}
	if s.Remote != "" {
		source.CreateElement("remoteBase").SetText(s.Remote)
	}
	if s.Includes != "" {
		source.CreateElement("includes").SetText(s.Includes)
	}
	if s.Excludes != "" {
		source.CreateElement("excludes").SetText(s.Excludes)
	}
	return nil
}

func (s *SingleSvnSource) fromEtree(source *etree.Element) *SingleSvnSource {
	if scm := source.SelectElement("scm"); scm != nil {
		if locations := scm.SelectElement("locations"); locations != nil {
			if moduleLocations := locations.SelectElement("hudson.scm.SubversionSCM_-ModuleLocation"); moduleLocations != nil {
				if remote := moduleLocations.SelectElement("remote"); remote != nil {
					s.Remote = remote.Text()
				}
				if credentialId := moduleLocations.SelectElement("credentialsId"); credentialId != nil {
					s.CredentialId = credentialId.Text()
				}
			}
		}
	}
	return s
}

func (s *SingleSvnSource) appendToEtree(source *etree.Element) *SingleSvnSource {

	source.CreateAttr("class", "jenkins.scm.impl.SingleSCMSource")
	source.CreateAttr("plugin", "scm-api")
	source.CreateElement("id").SetText(s.ScmId)
	source.CreateElement("name").SetText("master")

	scm := source.CreateElement("scm")
	scm.CreateAttr("class", "hudson.scm.SubversionSCM")
	scm.CreateAttr("plugin", "subversion")

	location := scm.CreateElement("locations").CreateElement("hudson.scm.SubversionSCM_-ModuleLocation")
	if s.Remote != "" {
		location.CreateElement("remote").SetText(s.Remote)
	}
	if s.CredentialId != "" {
		location.CreateElement("credentialsId").SetText(s.CredentialId)
	}
	location.CreateElement("local").SetText(".")
	location.CreateElement("depthOption").SetText("infinity")
	location.CreateElement("ignoreExternalsOption").SetText("true")
	location.CreateElement("cancelProcessOnExternalsFail").SetText("true")

	source.CreateElement("excludedRegions")
	source.CreateElement("includedRegions")
	source.CreateElement("excludedUsers")
	source.CreateElement("excludedRevprop")
	source.CreateElement("excludedCommitMessages")
	source.CreateElement("workspaceUpdater").CreateAttr("class", "hudson.scm.subversion.UpdateUpdater")
	source.CreateElement("ignoreDirPropChanges").SetText("false")
	source.CreateElement("filterChangelog").SetText("false")
	source.CreateElement("quietOperation").SetText("true")

	return s
}

func (s *MultiBranchJobTrigger) appendToEtree(properties *etree.Element) *MultiBranchJobTrigger {
	triggerProperty := properties.CreateElement("org.jenkinsci.plugins.workflow.multibranch.PipelineTriggerProperty")
	triggerProperty.CreateAttr("plugin", "multibranch-action-triggers")
	triggerProperty.CreateElement("createActionJobsToTrigger").SetText(s.CreateActionJobsToTrigger)
	triggerProperty.CreateElement("deleteActionJobsToTrigger").SetText(s.DeleteActionJobsToTrigger)
	return s
}

func (s *MultiBranchJobTrigger) fromEtree(properties *etree.Element) *MultiBranchJobTrigger {
	triggerProperty := properties.SelectElement("org.jenkinsci.plugins.workflow.multibranch.PipelineTriggerProperty")
	if triggerProperty != nil {
		s.CreateActionJobsToTrigger = triggerProperty.SelectElement("createActionJobsToTrigger").Text()
		s.DeleteActionJobsToTrigger = triggerProperty.SelectElement("deleteActionJobsToTrigger").Text()
	}
	return s
}
func createMultiBranchPipelineConfigXml(projectName string, pipeline *MultiBranchPipeline) (string, error) {
	doc := etree.NewDocument()
	xmlString := `
<?xml version='1.0' encoding='UTF-8'?>
<org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject plugin="workflow-multibranch">
  <actions/>
  <properties>
    <org.jenkinsci.plugins.pipeline.modeldefinition.config.FolderConfig plugin="pipeline-model-definition">
      <dockerLabel></dockerLabel>
      <registry plugin="docker-commons"/>
    </org.jenkinsci.plugins.pipeline.modeldefinition.config.FolderConfig>
  </properties>
  <folderViews class="jenkins.branch.MultiBranchProjectViewHolder" plugin="branch-api">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </folderViews>
  <healthMetrics>
    <com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric plugin="cloudbees-folder">
      <nonRecursive>false</nonRecursive>
    </com.cloudbees.hudson.plugins.folder.health.WorstChildHealthMetric>
  </healthMetrics>
  <icon class="jenkins.branch.MetadataActionFolderIcon" plugin="branch-api">
    <owner class="org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject" reference="../.."/>
  </icon>
</org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject>`
	err := doc.ReadFromString(xmlString)
	if err != nil {
		return "", err
	}

	project := doc.SelectElement("org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject")
	project.CreateElement("description").SetText(pipeline.Description)

	if pipeline.MultiBranchJobTrigger != nil {
		properties := project.SelectElement("properties")
		pipeline.MultiBranchJobTrigger.appendToEtree(properties)
	}

	if pipeline.Discarder != nil {
		discarder := project.CreateElement("orphanedItemStrategy")
		discarder.CreateAttr("class", "com.cloudbees.hudson.plugins.folder.computed.DefaultOrphanedItemStrategy")
		discarder.CreateAttr("plugin", "cloudbees-folder")
		discarder.CreateElement("pruneDeadBranches").SetText("true")
		discarder.CreateElement("daysToKeep").SetText(pipeline.Discarder.DaysToKeep)
		discarder.CreateElement("numToKeep").SetText(pipeline.Discarder.NumToKeep)
	}

	triggers := project.CreateElement("triggers")
	if pipeline.TimerTrigger != nil {
		timeTrigger := triggers.CreateElement(
			"com.cloudbees.hudson.plugins.folder.computed.PeriodicFolderTrigger")
		timeTrigger.CreateAttr("plugin", "cloudbees-folder")
		millis, err := strconv.ParseInt(pipeline.TimerTrigger.Interval, 10, 64)
		if err != nil {
			return "", err
		}
		timeTrigger.CreateElement("spec").SetText(toCrontab(millis))
		timeTrigger.CreateElement("interval").SetText(pipeline.TimerTrigger.Interval)

		triggers.CreateElement("disabled").SetText("false")
	}

	sources := project.CreateElement("sources")
	sources.CreateAttr("class", "jenkins.branch.MultiBranchProject$BranchSourceList")
	sources.CreateAttr("plugin", "branch-api")
	sourcesOwner := sources.CreateElement("owner")
	sourcesOwner.CreateAttr("class", "org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject")
	sourcesOwner.CreateAttr("reference", "../..")

	branchSource := sources.CreateElement("data").CreateElement("jenkins.branch.BranchSource")
	branchSourceStrategy := branchSource.CreateElement("strategy")
	branchSourceStrategy.CreateAttr("class", "jenkins.branch.NamedExceptionsBranchPropertyStrategy")
	branchSourceStrategy.CreateElement("defaultProperties").CreateAttr("class", "empty-list")
	branchSourceStrategy.CreateElement("namedExceptions").CreateAttr("class", "empty-list")
	source := branchSource.CreateElement("source")

	switch pipeline.SourceType {
	case "git":
		gitDefine := pipeline.GitSource
		gitDefine.ScmId = projectName + pipeline.Name
		gitDefine.appendToEtree(source)
	case "github":
		githubDefine := pipeline.GitHubSource
		githubDefine.ScmId = projectName + pipeline.Name
		githubDefine.appendToEtree(source)
	case "svn":
		svnDefine := pipeline.SvnSource
		svnDefine.ScmId = projectName + pipeline.Name
		svnDefine.appendToEtree(source)

	case "single_svn":
		singSvnDefine := pipeline.SingleSvnSource
		singSvnDefine.ScmId = projectName + pipeline.Name
		singSvnDefine.appendToEtree(source)

	case "bitbucket_server":
		bitbucketServerDefine := pipeline.BitbucketServerSource
		bitbucketServerDefine.ScmId = projectName + pipeline.Name
		bitbucketServerDefine.appendToEtree(source)

	default:
		return "", fmt.Errorf("unsupport source type")
	}

	factory := project.CreateElement("factory")
	factory.CreateAttr("class", "org.jenkinsci.plugins.workflow.multibranch.WorkflowBranchProjectFactory")
	factoryOwner := factory.CreateElement("owner")
	factoryOwner.CreateAttr("class", "org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject")
	factoryOwner.CreateAttr("reference", "../..")
	factory.CreateElement("scriptPath").SetText(pipeline.ScriptPath)

	doc.Indent(2)
	stringXml, err := doc.WriteToString()
	return replaceXmlVersion(stringXml, "1.0", "1.1"), err
}

func parseMultiBranchPipelineConfigXml(config string) (*MultiBranchPipeline, error) {
	pipeline := &MultiBranchPipeline{}
	config = replaceXmlVersion(config, "1.1", "1.0")
	doc := etree.NewDocument()
	err := doc.ReadFromString(config)
	if err != nil {
		return nil, err
	}
	project := doc.SelectElement("org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject")
	if project == nil {
		return nil, fmt.Errorf("can not parse mutibranch pipeline config")
	}
	if properties := project.SelectElement("properties"); properties != nil {
		if multibranchTrigger := properties.SelectElement(
			"org.jenkinsci.plugins.workflow.multibranch.PipelineTriggerProperty"); multibranchTrigger != nil {
			trigger := &MultiBranchJobTrigger{}
			trigger.fromEtree(properties)
			pipeline.MultiBranchJobTrigger = trigger
		}
	}
	pipeline.Description = project.SelectElement("description").Text()

	if discarder := project.SelectElement("orphanedItemStrategy"); discarder != nil {
		pipeline.Discarder = &DiscarderProperty{
			DaysToKeep: discarder.SelectElement("daysToKeep").Text(),
			NumToKeep:  discarder.SelectElement("numToKeep").Text(),
		}
	}
	if triggers := project.SelectElement("triggers"); triggers != nil {
		if timerTrigger := triggers.SelectElement(
			"com.cloudbees.hudson.plugins.folder.computed.PeriodicFolderTrigger"); timerTrigger != nil {
			pipeline.TimerTrigger = &TimerTrigger{
				Interval: timerTrigger.SelectElement("interval").Text(),
			}
		}
	}

	if sources := project.SelectElement("sources"); sources != nil {
		if sourcesData := sources.SelectElement("data"); sourcesData != nil {
			if branchSource := sourcesData.SelectElement("jenkins.branch.BranchSource"); branchSource != nil {
				source := branchSource.SelectElement("source")
				switch source.SelectAttr("class").Value {
				case "org.jenkinsci.plugins.github_branch_source.GitHubSCMSource":
					githubSource := &GithubSource{}
					githubSource.fromEtree(source)
					pipeline.GitHubSource = githubSource
					pipeline.SourceType = "github"
				case "com.cloudbees.jenkins.plugins.bitbucket.BitbucketSCMSource":
					bitbucketServerSource := &BitbucketServerSource{}
					bitbucketServerSource.fromEtree(source)
					pipeline.BitbucketServerSource = bitbucketServerSource
					pipeline.SourceType = "bitbucket_server"

				case "jenkins.plugins.git.GitSCMSource":
					gitSource := &GitSource{}
					gitSource.fromEtree(source)
					pipeline.SourceType = "git"
					pipeline.GitSource = gitSource

				case "jenkins.scm.impl.SingleSCMSource":
					singleSvnSource := &SingleSvnSource{}
					singleSvnSource.fromEtree(source)
					pipeline.SourceType = "single_svn"
					pipeline.SingleSvnSource = singleSvnSource

				case "jenkins.scm.impl.subversion.SubversionSCMSource":
					svnSource := &SvnSource{}
					svnSource.fromEtree(source)
					pipeline.SourceType = "svn"
					pipeline.SvnSource = svnSource
				}
			}
		}
	}

	pipeline.ScriptPath = project.SelectElement("factory").SelectElement("scriptPath").Text()
	return pipeline, nil
}

func toCrontab(millis int64) string {
	if millis*time.Millisecond.Nanoseconds() <= 5*time.Minute.Nanoseconds() {
		return "* * * * *"
	}
	if millis*time.Millisecond.Nanoseconds() <= 30*time.Minute.Nanoseconds() {
		return "H/5 * * * *"
	}
	if millis*time.Millisecond.Nanoseconds() <= 1*time.Hour.Nanoseconds() {
		return "H/15 * * * *"
	}
	if millis*time.Millisecond.Nanoseconds() <= 8*time.Hour.Nanoseconds() {
		return "H/30 * * * *"
	}
	if millis*time.Millisecond.Nanoseconds() <= 24*time.Hour.Nanoseconds() {
		return "H H/4 * * *"
	}
	if millis*time.Millisecond.Nanoseconds() <= 48*time.Hour.Nanoseconds() {
		return "H H/12 * * *"
	}
	return "H H * * *"

}

func getBuildSonarResults(build *gojenkins.Build) ([]*SonarStatus, error) {

	sonarClient, err := client.ClientSets().SonarQube()
	if err != nil {
		return nil, err
	}

	actions := build.GetActions()
	sonarStatuses := make([]*SonarStatus, 0)
	for _, action := range actions {
		if action.ClassName == SonarAnalysisActionClass {
			sonarStatus := &SonarStatus{}
			taskOptions := &sonargo.CeTaskOption{
				Id: action.SonarTaskId,
			}
			ceTask, _, err := sonarClient.SonarQube().Ce.Task(taskOptions)
			if err != nil {
				klog.Errorf("get sonar task error [%+v]", err)
				continue
			}
			sonarStatus.Task = ceTask
			measuresComponentOption := &sonargo.MeasuresComponentOption{
				Component:        ceTask.Task.ComponentKey,
				AdditionalFields: SonarAdditionalFields,
				MetricKeys:       SonarMetricKeys,
			}
			measures, _, err := sonarClient.SonarQube().Measures.Component(measuresComponentOption)
			if err != nil {
				klog.Errorf("get sonar task error [%+v]", err)
				continue
			}
			sonarStatus.Measures = measures

			issuesSearchOption := &sonargo.IssuesSearchOption{
				AdditionalFields: "_all",
				ComponentKeys:    ceTask.Task.ComponentKey,
				Resolved:         "false",
				Ps:               "10",
				S:                "FILE_LINE",
				Facets:           "severities,types",
			}
			issuesSearch, _, err := sonarClient.SonarQube().Issues.Search(issuesSearchOption)
			sonarStatus.Issues = issuesSearch
			jenkinsAction := action
			sonarStatus.JenkinsAction = &jenkinsAction

			sonarStatuses = append(sonarStatuses, sonarStatus)
		}
	}
	return sonarStatuses, nil
}
