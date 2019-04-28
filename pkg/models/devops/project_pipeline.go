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
	"github.com/golang/glog"
	"github.com/kubesphere/sonargo/sonar"
	"kubesphere.io/kubesphere/pkg/gojenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"
	"strconv"
	"strings"
	"time"
)

const (
	NoScmPipelineType       = "pipeline"
	MultiBranchPipelineType = "multi-branch-pipeline"
)

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
	Type                string               `json:"type"`
	Pipeline            *NoScmPipeline       `json:"pipeline,omitempty"`
	MultiBranchPipeline *MultiBranchPipeline `json:"multi_branch_pipeline,omitempty"`
}

type NoScmPipeline struct {
	Name              string             `json:"name"`
	Description       string             `json:"descriptio,omitempty"`
	Discarder         *DiscarderProperty `json:"discarder,omitempty"`
	Parameters        []*Parameter       `json:"parameters,omitempty"`
	DisableConcurrent bool               `json:"disable_concurrent,omitempty" mapstructure:"disable_concurrent"`
	TimerTrigger      *TimerTrigger      `json:"timer_trigger,omitempty" mapstructure:"timer_trigger"`
	RemoteTrigger     *RemoteTrigger     `json:"remote_trigger,omitempty" mapstructure:"remote_trigger"`
	Jenkinsfile       string             `json:"jenkinsfile,omitempty"`
}

type MultiBranchPipeline struct {
	Name            string             `json:"name"`
	Description     string             `json:"description,omitempty"`
	Discarder       *DiscarderProperty `json:"discarder,omitempty"`
	TimerTrigger    *TimerTrigger      `json:"timer_trigger,omitempty" mapstructure:"timer_trigger"`
	SourceType      string             `json:"source_type"`
	GitSource       *GitSource         `json:"git_source,omitempty"`
	GitHubSource    *GithubSource      `json:"github_source,omitempty"`
	SvnSource       *SvnSource         `json:"svn_source,omitempty"`
	SingleSvnSource *SingleSvnSource   `json:"single_svn_source,omitempty"`
	ScriptPath      string             `json:"script_path" mapstructure:"script_path"`
}

type GitSource struct {
	Url              string          `json:"url,omitempty" mapstructure:"url"`
	CredentialId     string          `json:"credential_id,omitempty" mapstructure:"credential_id"`
	DiscoverBranches bool            `json:"discover_branches,omitempty" mapstructure:"discover_branches"`
	CloneOption      *GitCloneOption `json:"git_clone_option,omitempty" mapstructure:"git_clone_option"`
	RegexFilter      string          `json:"regex_filter,omitempty" mapstructure:"regex_filter"`
}

type GithubSource struct {
	Owner                string                     `json:"owner,omitempty" mapstructure:"owner"`
	Repo                 string                     `json:"repo,omitempty" mapstructure:"repo"`
	CredentialId         string                     `json:"credential_id,omitempty" mapstructure:"credential_id"`
	ApiUri               string                     `json:"api_uri,omitempty" mapstructure:"api_uri"`
	DiscoverBranches     int                        `json:"discover_branches,omitempty" mapstructure:"discover_branches"`
	DiscoverPRFromOrigin int                        `json:"discover_pr_from_origin,omitempty" mapstructure:"discover_pr_from_origin"`
	DiscoverPRFromForks  *GithubDiscoverPRFromForks `json:"discover_pr_from_forks,omitempty" mapstructure:"discover_pr_from_forks"`
	CloneOption          *GitCloneOption            `json:"git_clone_option,omitempty" mapstructure:"git_clone_option"`
	RegexFilter          string                     `json:"regex_filter,omitempty" mapstructure:"regex_filter"`
}

type GitCloneOption struct {
	Shallow bool `json:"shallow,omitempty" mapstructure:"shallow"`
	Timeout int  `json:"timeout,omitempty" mapstructure:"timeout"`
	Depth   int  `json:"depth,omitempty" mapstructure:"depth"`
}

type SvnSource struct {
	Remote       string `json:"remote,omitempty"`
	CredentialId string `json:"credential_id,omitempty" mapstructure:"credential_id"`
	Includes     string `json:"includes,omitempty"`
	Excludes     string `json:"excludes,omitempty"`
}
type SingleSvnSource struct {
	Remote       string `json:"remote,omitempty"`
	CredentialId string `json:"credential_id,omitempty" mapstructure:"credential_id"`
}

type GithubDiscoverPRFromForks struct {
	Strategy int `json:"strategy,omitempty" mapstructure:"strategy"`
	Trust    int `json:"trust,omitempty" mapstructure:"trust"`
}

type DiscarderProperty struct {
	DaysToKeep string `json:"days_to_keep,omitempty" mapstructure:"days_to_keep"`
	NumToKeep  string `json:"num_to_keep,omitempty" mapstructure:"num_to_keep"`
}

type Parameter struct {
	Name         string `json:"name"`
	DefaultValue string `json:"default_value,omitempty" mapstructure:"default_value"`
	Type         string `json:"type"`
	Description  string `json:"description,omitempty"`
}

type TimerTrigger struct {
	// user in no scm job
	Cron string `json:"cron,omitempty"`

	// use in multi-branch job
	Interval string `json:"interval,omitempty"`
}

type RemoteTrigger struct {
	Token string `json:"token,omitempty"`
}

func replaceXmlVersion(config, oldVersion, targetVersion string) string {
	lines := strings.Split(string(config), "\n")
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
		parameterDefinitions := properties.CreateElement("hudson.model.ParametersDefinitionProperty").
			CreateElement("parameterDefinitions")
		for _, parameter := range pipeline.Parameters {
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
	if parametersProperty := properties.SelectElement("hudson.model.ParametersDefinitionProperty"); parametersProperty != nil {
		params := parametersProperty.SelectElement("parameterDefinitions").ChildElements()
		for _, param := range params {
			switch param.Tag {
			case "hudson.model.StringParameterDefinition":
				pipeline.Parameters = append(pipeline.Parameters, &Parameter{
					Name:         param.SelectElement("name").Text(),
					Description:  param.SelectElement("description").Text(),
					DefaultValue: param.SelectElement("defaultValue").Text(),
					Type:         ParameterTypeMap["hudson.model.StringParameterDefinition"],
				})
			case "hudson.model.BooleanParameterDefinition":
				pipeline.Parameters = append(pipeline.Parameters, &Parameter{
					Name:         param.SelectElement("name").Text(),
					Description:  param.SelectElement("description").Text(),
					DefaultValue: param.SelectElement("defaultValue").Text(),
					Type:         ParameterTypeMap["hudson.model.BooleanParameterDefinition"],
				})
			case "hudson.model.TextParameterDefinition":
				pipeline.Parameters = append(pipeline.Parameters, &Parameter{
					Name:         param.SelectElement("name").Text(),
					Description:  param.SelectElement("description").Text(),
					DefaultValue: param.SelectElement("defaultValue").Text(),
					Type:         ParameterTypeMap["hudson.model.TextParameterDefinition"],
				})
			case "hudson.model.FileParameterDefinition":
				pipeline.Parameters = append(pipeline.Parameters, &Parameter{
					Name:        param.SelectElement("name").Text(),
					Description: param.SelectElement("description").Text(),
					Type:        ParameterTypeMap["hudson.model.FileParameterDefinition"],
				})
			case "hudson.model.PasswordParameterDefinition":
				pipeline.Parameters = append(pipeline.Parameters, &Parameter{
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
				pipeline.Parameters = append(pipeline.Parameters, choiceParameter)
			default:
				pipeline.Parameters = append(pipeline.Parameters, &Parameter{
					Name:         param.SelectElement("name").Text(),
					Description:  param.SelectElement("description").Text(),
					DefaultValue: "unknown",
					Type:         param.Tag,
				})
			}
		}
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

	switch pipeline.SourceType {
	case "git":
		gitDefine := pipeline.GitSource

		gitSource := branchSource.CreateElement("source")
		gitSource.CreateAttr("class", "jenkins.plugins.git.GitSCMSource")
		gitSource.CreateAttr("plugin", "git")
		gitSource.CreateElement("id").SetText(projectName + pipeline.Name)
		gitSource.CreateElement("remote").SetText(gitDefine.Url)
		if gitDefine.CredentialId != "" {
			gitSource.CreateElement("credentialsId").SetText(gitDefine.CredentialId)
		}
		traits := gitSource.CreateElement("traits")
		if gitDefine.DiscoverBranches {
			traits.CreateElement("jenkins.plugins.git.traits.BranchDiscoveryTrait")
		}
		if gitDefine.CloneOption != nil {
			cloneExtension := traits.CreateElement("jenkins.plugins.git.traits.CloneOptionTrait").CreateElement("extension")
			cloneExtension.CreateAttr("class", "hudson.plugins.git.extensions.impl.CloneOption")
			cloneExtension.CreateElement("shallow").SetText(strconv.FormatBool(gitDefine.CloneOption.Shallow))
			cloneExtension.CreateElement("noTags").SetText(strconv.FormatBool(false))
			cloneExtension.CreateElement("reference")
			if gitDefine.CloneOption.Timeout >= 0 {
				cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(gitDefine.CloneOption.Timeout))
			} else {
				cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(10))
			}

			if gitDefine.CloneOption.Depth >= 0 {
				cloneExtension.CreateElement("depth").SetText(strconv.Itoa(gitDefine.CloneOption.Depth))
			} else {
				cloneExtension.CreateElement("depth").SetText(strconv.Itoa(1))
			}
		}

		if gitDefine.RegexFilter != "" {
			regexTraits := traits.CreateElement("jenkins.scm.impl.trait.RegexSCMHeadFilterTrait")
			regexTraits.CreateAttr("plugin", "scm-api@2.4.0")
			regexTraits.CreateElement("regex").SetText(gitDefine.RegexFilter)
		}

	case "github":
		githubDefine := pipeline.GitHubSource

		githubSource := branchSource.CreateElement("source")
		githubSource.CreateAttr("class", "org.jenkinsci.plugins.github_branch_source.GitHubSCMSource")
		githubSource.CreateAttr("plugin", "github-branch-source")
		githubSource.CreateElement("id").SetText(projectName + pipeline.Name)
		githubSource.CreateElement("credentialsId").SetText(githubDefine.CredentialId)
		githubSource.CreateElement("repoOwner").SetText(githubDefine.Owner)
		githubSource.CreateElement("repository").SetText(githubDefine.Repo)
		if githubDefine.ApiUri != "" {
			githubSource.CreateElement("apiUri").SetText(githubDefine.ApiUri)
		}
		traits := githubSource.CreateElement("traits")
		if githubDefine.DiscoverBranches != 0 {
			traits.CreateElement("org.jenkinsci.plugins.github__branch__source.BranchDiscoveryTrait").
				CreateElement("strategyId").SetText(strconv.Itoa(githubDefine.DiscoverBranches))
		}
		if githubDefine.DiscoverPRFromOrigin != 0 {
			traits.CreateElement("org.jenkinsci.plugins.github__branch__source.OriginPullRequestDiscoveryTrait").
				CreateElement("strategyId").SetText(strconv.Itoa(githubDefine.DiscoverPRFromOrigin))
		}
		if githubDefine.DiscoverPRFromForks != nil {
			forkTrait := traits.CreateElement("org.jenkinsci.plugins.github__branch__source.ForkPullRequestDiscoveryTrait")
			forkTrait.CreateElement("strategyId").SetText(strconv.Itoa(githubDefine.DiscoverPRFromForks.Strategy))
			trustClass := "org.jenkinsci.plugins.github_branch_source.ForkPullRequestDiscoveryTrait$"
			switch githubDefine.DiscoverPRFromForks.Trust {
			case 1:
				trustClass += "TrustContributors"
			case 2:
				trustClass += "TrustEveryone"
			case 3:
				trustClass += "TrustPermission"
			case 4:
				trustClass += "TrustNobody"
			default:
				return "", fmt.Errorf("unsupport trust choice")
			}
			forkTrait.CreateElement("trust").CreateAttr("class", trustClass)
		}
		if githubDefine.CloneOption != nil {
			cloneExtension := traits.CreateElement("jenkins.plugins.git.traits.CloneOptionTrait").CreateElement("extension")
			cloneExtension.CreateAttr("class", "hudson.plugins.git.extensions.impl.CloneOption")
			cloneExtension.CreateElement("shallow").SetText(strconv.FormatBool(githubDefine.CloneOption.Shallow))
			cloneExtension.CreateElement("noTags").SetText(strconv.FormatBool(false))
			cloneExtension.CreateElement("reference")
			if githubDefine.CloneOption.Timeout >= 0 {
				cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(githubDefine.CloneOption.Timeout))
			} else {
				cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(10))
			}

			if githubDefine.CloneOption.Depth >= 0 {
				cloneExtension.CreateElement("depth").SetText(strconv.Itoa(githubDefine.CloneOption.Depth))
			} else {
				cloneExtension.CreateElement("depth").SetText(strconv.Itoa(1))
			}
		}
		if githubDefine.RegexFilter != "" {
			regexTraits := traits.CreateElement("jenkins.scm.impl.trait.RegexSCMHeadFilterTrait")
			regexTraits.CreateAttr("plugin", "scm-api@2.4.0")
			regexTraits.CreateElement("regex").SetText(githubDefine.RegexFilter)
		}

	case "svn":
		svnDefine := pipeline.SvnSource
		svnSource := branchSource.CreateElement("source")
		svnSource.CreateAttr("class", "jenkins.scm.impl.subversion.SubversionSCMSource")
		svnSource.CreateAttr("plugin", "subversion")
		svnSource.CreateElement("id").SetText(projectName + pipeline.Name)
		if svnDefine.CredentialId != "" {
			svnSource.CreateElement("credentialsId").SetText(svnDefine.CredentialId)
		}
		if svnDefine.Remote != "" {
			svnSource.CreateElement("remoteBase").SetText(svnDefine.Remote)
		}
		if svnDefine.Includes != "" {
			svnSource.CreateElement("includes").SetText(svnDefine.Includes)
		}
		if svnDefine.Excludes != "" {
			svnSource.CreateElement("excludes").SetText(svnDefine.Excludes)
		}

	case "single_svn":
		singleSvnDefine := pipeline.SingleSvnSource
		if err != nil {
			return "", err
		}
		svnSource := branchSource.CreateElement("source")
		svnSource.CreateAttr("class", "jenkins.scm.impl.SingleSCMSource")
		svnSource.CreateAttr("plugin", "scm-api")

		svnSource.CreateElement("id").SetText(projectName + pipeline.Name)
		svnSource.CreateElement("name").SetText("master")

		scm := svnSource.CreateElement("scm")
		scm.CreateAttr("class", "hudson.scm.SubversionSCM")
		scm.CreateAttr("plugin", "subversion")

		location := scm.CreateElement("locations").CreateElement("hudson.scm.SubversionSCM_-ModuleLocation")
		if singleSvnDefine.Remote != "" {
			location.CreateElement("remote").SetText(singleSvnDefine.Remote)
		}
		if singleSvnDefine.CredentialId != "" {
			location.CreateElement("credentialsId").SetText(singleSvnDefine.CredentialId)
		}
		location.CreateElement("local").SetText(".")
		location.CreateElement("depthOption").SetText("infinity")
		location.CreateElement("ignoreExternalsOption").SetText("true")
		location.CreateElement("cancelProcessOnExternalsFail").SetText("true")

		svnSource.CreateElement("excludedRegions")
		svnSource.CreateElement("includedRegions")
		svnSource.CreateElement("excludedUsers")
		svnSource.CreateElement("excludedRevprop")
		svnSource.CreateElement("excludedCommitMessages")
		svnSource.CreateElement("workspaceUpdater").CreateAttr("class", "hudson.scm.subversion.UpdateUpdater")
		svnSource.CreateElement("ignoreDirPropChanges").SetText("false")
		svnSource.CreateElement("filterChangelog").SetText("false")
		svnSource.CreateElement("quietOperation").SetText("true")

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
					if credential := source.SelectElement("credentialsId"); credential != nil {
						githubSource.CredentialId = credential.Text()
					}
					if repoOwner := source.SelectElement("repoOwner"); repoOwner != nil {
						githubSource.Owner = repoOwner.Text()
					}
					if repository := source.SelectElement("repository"); repository != nil {
						githubSource.Repo = repository.Text()
					}
					if apiUri := source.SelectElement("apiUri"); apiUri != nil {
						githubSource.ApiUri = apiUri.Text()
					}
					traits := source.SelectElement("traits")
					if branchDiscoverTrait := traits.SelectElement(
						"org.jenkinsci.plugins.github__branch__source.BranchDiscoveryTrait"); branchDiscoverTrait != nil {
						strategyId, err := strconv.Atoi(branchDiscoverTrait.SelectElement("strategyId").Text())
						if err != nil {
							return nil, err
						}
						githubSource.DiscoverBranches = strategyId
					}
					if originPRDiscoverTrait := traits.SelectElement(
						"org.jenkinsci.plugins.github__branch__source.OriginPullRequestDiscoveryTrait"); originPRDiscoverTrait != nil {
						strategyId, err := strconv.Atoi(originPRDiscoverTrait.SelectElement("strategyId").Text())
						if err != nil {
							return nil, err
						}
						githubSource.DiscoverPRFromOrigin = strategyId
					}
					if forkPRDiscoverTrait := traits.SelectElement(
						"org.jenkinsci.plugins.github__branch__source.ForkPullRequestDiscoveryTrait"); forkPRDiscoverTrait != nil {
						strategyId, err := strconv.Atoi(forkPRDiscoverTrait.SelectElement("strategyId").Text())
						if err != nil {
							return nil, err
						}
						trustClass := forkPRDiscoverTrait.SelectElement("trust").SelectAttr("class").Value
						trust := strings.Split(trustClass, "$")
						switch trust[1] {
						case "TrustContributors":
							githubSource.DiscoverPRFromForks = &GithubDiscoverPRFromForks{
								Strategy: strategyId,
								Trust:    1,
							}
						case "TrustEveryone":
							githubSource.DiscoverPRFromForks = &GithubDiscoverPRFromForks{
								Strategy: strategyId,
								Trust:    2,
							}
						case "TrustPermission":
							githubSource.DiscoverPRFromForks = &GithubDiscoverPRFromForks{
								Strategy: strategyId,
								Trust:    3,
							}
						case "TrustNobody":
							githubSource.DiscoverPRFromForks = &GithubDiscoverPRFromForks{
								Strategy: strategyId,
								Trust:    4,
							}
						}
						if cloneTrait := traits.SelectElement(
							"jenkins.plugins.git.traits.CloneOptionTrait"); cloneTrait != nil {
							if cloneExtension := cloneTrait.SelectElement(
								"extension"); cloneExtension != nil {
								githubSource.CloneOption = &GitCloneOption{}
								if value, err := strconv.ParseBool(cloneExtension.SelectElement("shallow").Text()); err == nil {
									githubSource.CloneOption.Shallow = value
								}
								if value, err := strconv.ParseInt(cloneExtension.SelectElement("timeout").Text(), 10, 32); err == nil {
									githubSource.CloneOption.Timeout = int(value)
								}
								if value, err := strconv.ParseInt(cloneExtension.SelectElement("depth").Text(), 10, 32); err == nil {
									githubSource.CloneOption.Depth = int(value)
								}
							}
						}

						if regexTrait := traits.SelectElement(
							"jenkins.scm.impl.trait.RegexSCMHeadFilterTrait"); regexTrait != nil {
							if regex := regexTrait.SelectElement("regex"); regex != nil {
								githubSource.RegexFilter = regex.Text()
							}
						}
					}

					pipeline.GitHubSource = githubSource
					pipeline.SourceType = "github"
				case "jenkins.plugins.git.GitSCMSource":
					gitSource := &GitSource{}
					if credential := source.SelectElement("credentialsId"); credential != nil {
						gitSource.CredentialId = credential.Text()
					}
					if remote := source.SelectElement("remote"); remote != nil {
						gitSource.Url = remote.Text()
					}

					traits := source.SelectElement("traits")
					if branchDiscoverTrait := traits.SelectElement(
						"jenkins.plugins.git.traits.BranchDiscoveryTrait"); branchDiscoverTrait != nil {
						gitSource.DiscoverBranches = true
					}
					if cloneTrait := traits.SelectElement(
						"jenkins.plugins.git.traits.CloneOptionTrait"); cloneTrait != nil {
						if cloneExtension := cloneTrait.SelectElement(
							"extension"); cloneExtension != nil {
							gitSource.CloneOption = &GitCloneOption{}
							if value, err := strconv.ParseBool(cloneExtension.SelectElement("shallow").Text()); err == nil {
								gitSource.CloneOption.Shallow = value
							}
							if value, err := strconv.ParseInt(cloneExtension.SelectElement("timeout").Text(), 10, 32); err == nil {
								gitSource.CloneOption.Timeout = int(value)
							}
							if value, err := strconv.ParseInt(cloneExtension.SelectElement("depth").Text(), 10, 32); err == nil {
								gitSource.CloneOption.Depth = int(value)
							}
						}
					}
					if regexTrait := traits.SelectElement(
						"jenkins.scm.impl.trait.RegexSCMHeadFilterTrait"); regexTrait != nil {
						if regex := regexTrait.SelectElement("regex"); regex != nil {
							gitSource.RegexFilter = regex.Text()
						}
					}

					pipeline.SourceType = "git"
					pipeline.GitSource = gitSource
				case "jenkins.scm.impl.SingleSCMSource":
					singleSvnSource := &SingleSvnSource{}

					if scm := source.SelectElement("scm"); scm != nil {
						if locations := scm.SelectElement("locations"); locations != nil {
							if moduleLocations := locations.SelectElement("hudson.scm.SubversionSCM_-ModuleLocation"); moduleLocations != nil {
								if remote := moduleLocations.SelectElement("remote"); remote != nil {
									singleSvnSource.Remote = remote.Text()
								}
								if credentialId := moduleLocations.SelectElement("credentialsId"); credentialId != nil {
									singleSvnSource.CredentialId = credentialId.Text()
								}
							}
						}
					}
					pipeline.SourceType = "single_svn"

					pipeline.SingleSvnSource = singleSvnSource

				case "jenkins.scm.impl.subversion.SubversionSCMSource":
					svnSource := &SvnSource{}

					if remote := source.SelectElement("remoteBase"); remote != nil {
						svnSource.Remote = remote.Text()
					}

					if credentialsId := source.SelectElement("credentialsId"); credentialsId != nil {
						svnSource.CredentialId = credentialsId.Text()
					}

					if includes := source.SelectElement("includes"); includes != nil {
						svnSource.Includes = includes.Text()
					}

					if excludes := source.SelectElement("excludes"); excludes != nil {
						svnSource.Excludes = excludes.Text()
					}

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
	sonarClient := sonarqube.Client()
	actions := build.GetActions()
	sonarStatuses := make([]*SonarStatus, 0)
	for _, action := range actions {
		if action.ClassName == SonarAnalysisActionClass {
			sonarStatus := &SonarStatus{}
			taskOptions := &sonargo.CeTaskOption{
				Id: action.SonarTaskId,
			}
			ceTask, _, err := sonarClient.Ce.Task(taskOptions)
			if err != nil {
				glog.Errorf("get sonar task error [%+v]", err)
				continue
			}
			sonarStatus.Task = ceTask
			measuresComponentOption := &sonargo.MeasuresComponentOption{
				Component:        ceTask.Task.ComponentKey,
				AdditionalFields: SonarAdditionalFields,
				MetricKeys:       SonarMetricKeys,
			}
			measures, _, err := sonarClient.Measures.Component(measuresComponentOption)
			if err != nil {
				glog.Errorf("get sonar task error [%+v]", err)
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
			issuesSearch, _, err := sonarClient.Issues.Search(issuesSearchOption)
			sonarStatus.Issues = issuesSearch
			jenkinsAction := action
			sonarStatus.JenkinsAction = &jenkinsAction

			sonarStatuses = append(sonarStatuses, sonarStatus)
		}
	}
	return sonarStatuses, nil
}
