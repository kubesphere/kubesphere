package internal

import (
	"github.com/beevik/etree"
	devopsv1alpha3 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	"strconv"
	"strings"
)

func AppendGitlabSourceToEtree(source *etree.Element, gitSource *devopsv1alpha3.GitlabSource) {
	source.CreateAttr("class", "io.jenkins.plugins.gitlabbranchsource.GitLabSCMSource")
	source.CreateAttr("plugin", "gitlab-branch-source")
	source.CreateElement("id").SetText(gitSource.ScmId)
	source.CreateElement("serverName").SetText(gitSource.ServerName)
	source.CreateElement("credentialsId").SetText(gitSource.CredentialId)
	source.CreateElement("projectOwner").SetText(gitSource.Owner)
	source.CreateElement("projectPath").SetText(gitSource.Repo)
	traits := source.CreateElement("traits")
	if gitSource.DiscoverBranches != 0 {
		traits.CreateElement("io.jenkins.plugins.gitlabbranchsource.BranchDiscoveryTrait").
			CreateElement("strategyId").SetText(strconv.Itoa(gitSource.DiscoverBranches))
	}
	if gitSource.DiscoverTags {
		traits.CreateElement("io.jenkins.plugins.gitlabbranchsource.TagDiscoveryTrait")
	}
	if gitSource.DiscoverPRFromOrigin != 0 {
		traits.CreateElement("io.jenkins.plugins.gitlabbranchsource.OriginMergeRequestDiscoveryTrait").
			CreateElement("strategyId").SetText(strconv.Itoa(gitSource.DiscoverPRFromOrigin))
	}
	if gitSource.DiscoverPRFromForks != nil {
		forkTrait := traits.CreateElement("io.jenkins.plugins.gitlabbranchsource.ForkMergeRequestDiscoveryTrait")
		forkTrait.CreateElement("strategyId").SetText(strconv.Itoa(gitSource.DiscoverPRFromForks.Strategy))
		trustClass := "io.jenkins.plugins.gitlabbranchsource.ForkMergeRequestDiscoveryTrait$"
		switch gitSource.DiscoverPRFromForks.Trust {
		case 1:
			trustClass += "TrustMembers" // it's difference with GitHub
		case 2:
			trustClass += "TrustEveryone"
		case 3:
			trustClass += "TrustPermission"
		case 4:
			trustClass += "TrustNobody"
		}
		forkTrait.CreateElement("trust").CreateAttr("class", trustClass)
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

func GetGitlabSourceFromEtree(source *etree.Element) (gitSource *devopsv1alpha3.GitlabSource) {
	gitSource = &devopsv1alpha3.GitlabSource{}
	if credential := source.SelectElement("credentialsId"); credential != nil {
		gitSource.CredentialId = credential.Text()
	}
	if serverName := source.SelectElement("serverName"); serverName != nil {
		gitSource.ServerName = serverName.Text()
	}
	if repoOwner := source.SelectElement("projectOwner"); repoOwner != nil {
		gitSource.Owner = repoOwner.Text()
	}
	if repository := source.SelectElement("projectPath"); repository != nil {
		gitSource.Repo = repository.Text()
	}
	traits := source.SelectElement("traits")
	if branchDiscoverTrait := traits.SelectElement(
		"io.jenkins.plugins.gitlabbranchsource.BranchDiscoveryTrait"); branchDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(branchDiscoverTrait.SelectElement("strategyId").Text())
		gitSource.DiscoverBranches = strategyId
	}
	if tagDiscoverTrait := traits.SelectElement(
		"io.jenkins.plugins.gitlabbranchsource.TagDiscoveryTrait"); tagDiscoverTrait != nil {
		gitSource.DiscoverTags = true
	}
	if originPRDiscoverTrait := traits.SelectElement(
		"io.jenkins.plugins.gitlabbranchsource.OriginMergeRequestDiscoveryTrait"); originPRDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(originPRDiscoverTrait.SelectElement("strategyId").Text())
		gitSource.DiscoverPRFromOrigin = strategyId
	}
	if forkPRDiscoverTrait := traits.SelectElement(
		"io.jenkins.plugins.gitlabbranchsource.ForkMergeRequestDiscoveryTrait"); forkPRDiscoverTrait != nil {
		strategyId, _ := strconv.Atoi(forkPRDiscoverTrait.SelectElement("strategyId").Text())
		trustClass := forkPRDiscoverTrait.SelectElement("trust").SelectAttr("class").Value
		trust := strings.Split(trustClass, "$")
		switch trust[1] {
		case "TrustMembers": // it's difference with GitHub
			gitSource.DiscoverPRFromForks = &devopsv1alpha3.DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    1,
			}
		case "TrustEveryone":
			gitSource.DiscoverPRFromForks = &devopsv1alpha3.DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    2,
			}
		case "TrustPermission":
			gitSource.DiscoverPRFromForks = &devopsv1alpha3.DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    3,
			}
		case "TrustNobody":
			gitSource.DiscoverPRFromForks = &devopsv1alpha3.DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    4,
			}
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
	}
	return
}
