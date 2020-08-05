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
	"fmt"
	"github.com/beevik/etree"
	devopsv1alpha3 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	"strconv"
	"strings"
	"time"
)

func replaceXmlVersion(config, oldVersion, targetVersion string) string {
	lines := strings.Split(config, "\n")
	lines[0] = strings.Replace(lines[0], oldVersion, targetVersion, -1)
	output := strings.Join(lines, "\n")
	return output
}

func createPipelineConfigXml(pipeline *devopsv1alpha3.NoScmPipeline) (string, error) {
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
		appendParametersToEtree(properties, pipeline.Parameters)
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

func parsePipelineConfigXml(config string) (*devopsv1alpha3.NoScmPipeline, error) {
	pipeline := &devopsv1alpha3.NoScmPipeline{}
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
		pipeline.Discarder = &devopsv1alpha3.DiscarderProperty{
			DaysToKeep: strategy.SelectElement("daysToKeep").Text(),
			NumToKeep:  strategy.SelectElement("numToKeep").Text(),
		}
	}

	pipeline.Parameters = getParametersfromEtree(properties)
	if len(pipeline.Parameters) == 0 {
		pipeline.Parameters = nil
	}

	if triggerProperty := properties.
		SelectElement(
			"org.jenkinsci.plugins.workflow.job.properties.PipelineTriggersJobProperty"); triggerProperty != nil {
		triggers := triggerProperty.SelectElement("triggers")
		if timerTrigger := triggers.SelectElement("hudson.triggers.TimerTrigger"); timerTrigger != nil {
			pipeline.TimerTrigger = &devopsv1alpha3.TimerTrigger{
				Cron: timerTrigger.SelectElement("spec").Text(),
			}
		}
	}
	if authToken := flow.SelectElement("authToken"); authToken != nil {
		pipeline.RemoteTrigger = &devopsv1alpha3.RemoteTrigger{
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

func appendParametersToEtree(properties *etree.Element, parameters []devopsv1alpha3.Parameter) {
	parameterDefinitions := properties.CreateElement("hudson.model.ParametersDefinitionProperty").
		CreateElement("parameterDefinitions")
	for _, parameter := range parameters {
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

func getParametersfromEtree(properties *etree.Element) []devopsv1alpha3.Parameter {
	var parameters []devopsv1alpha3.Parameter
	if parametersProperty := properties.SelectElement("hudson.model.ParametersDefinitionProperty"); parametersProperty != nil {
		params := parametersProperty.SelectElement("parameterDefinitions").ChildElements()
		for _, param := range params {
			switch param.Tag {
			case "hudson.model.StringParameterDefinition":
				parameters = append(parameters, devopsv1alpha3.Parameter{
					Name:         param.SelectElement("name").Text(),
					Description:  param.SelectElement("description").Text(),
					DefaultValue: param.SelectElement("defaultValue").Text(),
					Type:         ParameterTypeMap["hudson.model.StringParameterDefinition"],
				})
			case "hudson.model.BooleanParameterDefinition":
				parameters = append(parameters, devopsv1alpha3.Parameter{
					Name:         param.SelectElement("name").Text(),
					Description:  param.SelectElement("description").Text(),
					DefaultValue: param.SelectElement("defaultValue").Text(),
					Type:         ParameterTypeMap["hudson.model.BooleanParameterDefinition"],
				})
			case "hudson.model.TextParameterDefinition":
				parameters = append(parameters, devopsv1alpha3.Parameter{
					Name:         param.SelectElement("name").Text(),
					Description:  param.SelectElement("description").Text(),
					DefaultValue: param.SelectElement("defaultValue").Text(),
					Type:         ParameterTypeMap["hudson.model.TextParameterDefinition"],
				})
			case "hudson.model.FileParameterDefinition":
				parameters = append(parameters, devopsv1alpha3.Parameter{
					Name:        param.SelectElement("name").Text(),
					Description: param.SelectElement("description").Text(),
					Type:        ParameterTypeMap["hudson.model.FileParameterDefinition"],
				})
			case "hudson.model.PasswordParameterDefinition":
				parameters = append(parameters, devopsv1alpha3.Parameter{
					Name:         param.SelectElement("name").Text(),
					Description:  param.SelectElement("description").Text(),
					DefaultValue: param.SelectElement("name").Text(),
					Type:         ParameterTypeMap["hudson.model.PasswordParameterDefinition"],
				})
			case "hudson.model.ChoiceParameterDefinition":
				choiceParameter := devopsv1alpha3.Parameter{
					Name:        param.SelectElement("name").Text(),
					Description: param.SelectElement("description").Text(),
					Type:        ParameterTypeMap["hudson.model.ChoiceParameterDefinition"],
				}
				choices := param.SelectElement("choices").SelectElement("a").SelectElements("string")
				for _, choice := range choices {
					choiceParameter.DefaultValue += fmt.Sprintf("%s\n", choice.Text())
				}
				choiceParameter.DefaultValue = strings.TrimSpace(choiceParameter.DefaultValue)
				parameters = append(parameters, choiceParameter)
			default:
				parameters = append(parameters, devopsv1alpha3.Parameter{
					Name:         param.SelectElement("name").Text(),
					Description:  param.SelectElement("description").Text(),
					DefaultValue: "unknown",
					Type:         param.Tag,
				})
			}
		}
	}
	return parameters
}

func appendGitSourceToEtree(source *etree.Element, gitSource *devopsv1alpha3.GitSource) {
	source.CreateAttr("class", "jenkins.plugins.git.GitSCMSource")
	source.CreateAttr("plugin", "git")
	source.CreateElement("id").SetText(gitSource.ScmId)
	source.CreateElement("remote").SetText(gitSource.Url)
	if gitSource.CredentialId != "" {
		source.CreateElement("credentialsId").SetText(gitSource.CredentialId)
	}
	traits := source.CreateElement("traits")
	if gitSource.DiscoverBranches {
		traits.CreateElement("jenkins.plugins.git.traits.BranchDiscoveryTrait")
	}
	if gitSource.CloneOption != nil {
		cloneExtension := traits.CreateElement("jenkins.plugins.git.traits.CloneOptionTrait").CreateElement("extension")
		cloneExtension.CreateAttr("class", "hudson.plugins.git.extensions.impl.CloneOption")
		cloneExtension.CreateElement("shallow").SetText(strconv.FormatBool(gitSource.CloneOption.Shallow))
		cloneExtension.CreateElement("noTags").SetText(strconv.FormatBool(false))
		cloneExtension.CreateElement("honorRefspec").SetText(strconv.FormatBool(true))
		cloneExtension.CreateElement("reference")
		if gitSource.CloneOption.Timeout >= 0 {
			cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(gitSource.CloneOption.Timeout))
		} else {
			cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(10))
		}

		if gitSource.CloneOption.Depth >= 0 {
			cloneExtension.CreateElement("depth").SetText(strconv.Itoa(gitSource.CloneOption.Depth))
		} else {
			cloneExtension.CreateElement("depth").SetText(strconv.Itoa(1))
		}
	}

	if gitSource.RegexFilter != "" {
		regexTraits := traits.CreateElement("jenkins.scm.impl.trait.RegexSCMHeadFilterTrait")
		regexTraits.CreateAttr("plugin", "scm-api@2.4.0")
		regexTraits.CreateElement("regex").SetText(gitSource.RegexFilter)
	}
	return
}

func getGitSourcefromEtree(source *etree.Element) *devopsv1alpha3.GitSource {
	var gitSource devopsv1alpha3.GitSource
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
			gitSource.CloneOption = &devopsv1alpha3.GitCloneOption{}
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
	return &gitSource
}

func getGithubSourcefromEtree(source *etree.Element) *devopsv1alpha3.GithubSource {
	var githubSource devopsv1alpha3.GithubSource
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
		strategyId, _ := strconv.Atoi(branchDiscoverTrait.SelectElement("strategyId").Text())
		githubSource.DiscoverBranches = strategyId
	}
	if originPRDiscoverTrait := traits.SelectElement(
		"org.jenkinsci.plugins.github__branch__source.OriginPullRequestDiscoveryTrait"); originPRDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(originPRDiscoverTrait.SelectElement("strategyId").Text())
		githubSource.DiscoverPRFromOrigin = strategyId
	}
	if forkPRDiscoverTrait := traits.SelectElement(
		"org.jenkinsci.plugins.github__branch__source.ForkPullRequestDiscoveryTrait"); forkPRDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(forkPRDiscoverTrait.SelectElement("strategyId").Text())
		trustClass := forkPRDiscoverTrait.SelectElement("trust").SelectAttr("class").Value
		trust := strings.Split(trustClass, "$")
		switch trust[1] {
		case "TrustContributors":
			githubSource.DiscoverPRFromForks = &devopsv1alpha3.DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    1,
			}
		case "TrustEveryone":
			githubSource.DiscoverPRFromForks = &devopsv1alpha3.DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    2,
			}
		case "TrustPermission":
			githubSource.DiscoverPRFromForks = &devopsv1alpha3.DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    3,
			}
		case "TrustNobody":
			githubSource.DiscoverPRFromForks = &devopsv1alpha3.DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    4,
			}
		}
		if cloneTrait := traits.SelectElement(
			"jenkins.plugins.git.traits.CloneOptionTrait"); cloneTrait != nil {
			if cloneExtension := cloneTrait.SelectElement(
				"extension"); cloneExtension != nil {
				githubSource.CloneOption = &devopsv1alpha3.GitCloneOption{}
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
	return &githubSource
}

func appendGithubSourceToEtree(source *etree.Element, githubSource *devopsv1alpha3.GithubSource) {
	source.CreateAttr("class", "org.jenkinsci.plugins.github_branch_source.GitHubSCMSource")
	source.CreateAttr("plugin", "github-branch-source")
	source.CreateElement("id").SetText(githubSource.ScmId)
	source.CreateElement("credentialsId").SetText(githubSource.CredentialId)
	source.CreateElement("repoOwner").SetText(githubSource.Owner)
	source.CreateElement("repository").SetText(githubSource.Repo)
	if githubSource.ApiUri != "" {
		source.CreateElement("apiUri").SetText(githubSource.ApiUri)
	}
	traits := source.CreateElement("traits")
	if githubSource.DiscoverBranches != 0 {
		traits.CreateElement("org.jenkinsci.plugins.github__branch__source.BranchDiscoveryTrait").
			CreateElement("strategyId").SetText(strconv.Itoa(githubSource.DiscoverBranches))
	}
	if githubSource.DiscoverPRFromOrigin != 0 {
		traits.CreateElement("org.jenkinsci.plugins.github__branch__source.OriginPullRequestDiscoveryTrait").
			CreateElement("strategyId").SetText(strconv.Itoa(githubSource.DiscoverPRFromOrigin))
	}
	if githubSource.DiscoverPRFromForks != nil {
		forkTrait := traits.CreateElement("org.jenkinsci.plugins.github__branch__source.ForkPullRequestDiscoveryTrait")
		forkTrait.CreateElement("strategyId").SetText(strconv.Itoa(githubSource.DiscoverPRFromForks.Strategy))
		trustClass := "org.jenkinsci.plugins.github_branch_source.ForkPullRequestDiscoveryTrait$"
		switch githubSource.DiscoverPRFromForks.Trust {
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
	if githubSource.CloneOption != nil {
		cloneExtension := traits.CreateElement("jenkins.plugins.git.traits.CloneOptionTrait").CreateElement("extension")
		cloneExtension.CreateAttr("class", "hudson.plugins.git.extensions.impl.CloneOption")
		cloneExtension.CreateElement("shallow").SetText(strconv.FormatBool(githubSource.CloneOption.Shallow))
		cloneExtension.CreateElement("noTags").SetText(strconv.FormatBool(false))
		cloneExtension.CreateElement("honorRefspec").SetText(strconv.FormatBool(true))
		cloneExtension.CreateElement("reference")
		if githubSource.CloneOption.Timeout >= 0 {
			cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(githubSource.CloneOption.Timeout))
		} else {
			cloneExtension.CreateElement("timeout").SetText(strconv.Itoa(10))
		}

		if githubSource.CloneOption.Depth >= 0 {
			cloneExtension.CreateElement("depth").SetText(strconv.Itoa(githubSource.CloneOption.Depth))
		} else {
			cloneExtension.CreateElement("depth").SetText(strconv.Itoa(1))
		}
	}
	if githubSource.RegexFilter != "" {
		regexTraits := traits.CreateElement("jenkins.scm.impl.trait.RegexSCMHeadFilterTrait")
		regexTraits.CreateAttr("plugin", "scm-api@2.4.0")
		regexTraits.CreateElement("regex").SetText(githubSource.RegexFilter)
	}
	return
}

func getBitbucketServerSourceFromEtree(source *etree.Element) *devopsv1alpha3.BitbucketServerSource {
	var s devopsv1alpha3.BitbucketServerSource
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
			s.DiscoverPRFromForks = &devopsv1alpha3.DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    1,
			}
		case "TrustTeamForks":
			s.DiscoverPRFromForks = &devopsv1alpha3.DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    2,
			}
		case "TrustNobody":
			s.DiscoverPRFromForks = &devopsv1alpha3.DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    3,
			}
		}
		if cloneTrait := traits.SelectElement(
			"jenkins.plugins.git.traits.CloneOptionTrait"); cloneTrait != nil {
			if cloneExtension := cloneTrait.SelectElement(
				"extension"); cloneExtension != nil {
				s.CloneOption = &devopsv1alpha3.GitCloneOption{}
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
	return &s
}

func appendBitbucketServerSourceToEtree(source *etree.Element, s *devopsv1alpha3.BitbucketServerSource) {
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
	return
}

func getSvnSourcefromEtree(source *etree.Element) *devopsv1alpha3.SvnSource {
	var s devopsv1alpha3.SvnSource
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
	return &s
}

func appendSvnSourceToEtree(source *etree.Element, s *devopsv1alpha3.SvnSource) {
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
	return
}

func getSingleSvnSourceFromEtree(source *etree.Element) *devopsv1alpha3.SingleSvnSource {
	var s devopsv1alpha3.SingleSvnSource
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
	return &s
}

func appendSingleSvnSourceToEtree(source *etree.Element, s *devopsv1alpha3.SingleSvnSource) {

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

	return
}

func appendMultiBranchJobTriggerToEtree(properties *etree.Element, s *devopsv1alpha3.MultiBranchJobTrigger) {
	triggerProperty := properties.CreateElement("org.jenkinsci.plugins.workflow.multibranch.PipelineTriggerProperty")
	triggerProperty.CreateAttr("plugin", "multibranch-action-triggers")
	triggerProperty.CreateElement("createActionJobsToTrigger").SetText(s.CreateActionJobsToTrigger)
	triggerProperty.CreateElement("deleteActionJobsToTrigger").SetText(s.DeleteActionJobsToTrigger)
	return
}

func getMultiBranchJobTriggerfromEtree(properties *etree.Element) *devopsv1alpha3.MultiBranchJobTrigger {
	var s devopsv1alpha3.MultiBranchJobTrigger
	triggerProperty := properties.SelectElement("org.jenkinsci.plugins.workflow.multibranch.PipelineTriggerProperty")
	if triggerProperty != nil {
		s.CreateActionJobsToTrigger = triggerProperty.SelectElement("createActionJobsToTrigger").Text()
		s.DeleteActionJobsToTrigger = triggerProperty.SelectElement("deleteActionJobsToTrigger").Text()
	}
	return &s
}
func createMultiBranchPipelineConfigXml(projectName string, pipeline *devopsv1alpha3.MultiBranchPipeline) (string, error) {
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
		appendMultiBranchJobTriggerToEtree(properties, pipeline.MultiBranchJobTrigger)
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
		appendGitSourceToEtree(source, pipeline.GitSource)
	case "github":
		appendGithubSourceToEtree(source, pipeline.GitHubSource)
	case "svn":
		appendSvnSourceToEtree(source, pipeline.SvnSource)
	case "single_svn":
		appendSingleSvnSourceToEtree(source, pipeline.SingleSvnSource)
	case "bitbucket_server":
		appendBitbucketServerSourceToEtree(source, pipeline.BitbucketServerSource)

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

func parseMultiBranchPipelineConfigXml(config string) (*devopsv1alpha3.MultiBranchPipeline, error) {
	pipeline := &devopsv1alpha3.MultiBranchPipeline{}
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
			pipeline.MultiBranchJobTrigger = getMultiBranchJobTriggerfromEtree(properties)
		}
	}
	pipeline.Description = project.SelectElement("description").Text()

	if discarder := project.SelectElement("orphanedItemStrategy"); discarder != nil {
		pipeline.Discarder = &devopsv1alpha3.DiscarderProperty{
			DaysToKeep: discarder.SelectElement("daysToKeep").Text(),
			NumToKeep:  discarder.SelectElement("numToKeep").Text(),
		}
	}
	if triggers := project.SelectElement("triggers"); triggers != nil {
		if timerTrigger := triggers.SelectElement(
			"com.cloudbees.hudson.plugins.folder.computed.PeriodicFolderTrigger"); timerTrigger != nil {
			pipeline.TimerTrigger = &devopsv1alpha3.TimerTrigger{
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
					pipeline.GitHubSource = getGithubSourcefromEtree(source)
					pipeline.SourceType = "github"
				case "com.cloudbees.jenkins.plugins.bitbucket.BitbucketSCMSource":
					pipeline.BitbucketServerSource = getBitbucketServerSourceFromEtree(source)
					pipeline.SourceType = "bitbucket_server"

				case "jenkins.plugins.git.GitSCMSource":
					pipeline.SourceType = "git"
					pipeline.GitSource = getGitSourcefromEtree(source)

				case "jenkins.scm.impl.SingleSCMSource":
					pipeline.SourceType = "single_svn"
					pipeline.SingleSvnSource = getSingleSvnSourceFromEtree(source)

				case "jenkins.scm.impl.subversion.SubversionSCMSource":
					pipeline.SourceType = "svn"
					pipeline.SvnSource = getSvnSourcefromEtree(source)
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
