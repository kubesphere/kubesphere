package internal

import (
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"k8s.io/klog"

	devopsv1alpha3 "kubesphere.io/api/devops/v1alpha3"
)

func AppendBitbucketServerSourceToEtree(source *etree.Element, gitSource *devopsv1alpha3.BitbucketServerSource) {
	if gitSource == nil {
		klog.Warning("please provide BitbucketServer source when the sourceType is BitbucketServer")
		return
	}
	source.CreateAttr("class", "com.cloudbees.jenkins.plugins.bitbucket.BitbucketSCMSource")
	source.CreateAttr("plugin", "cloudbees-bitbucket-branch-source")
	source.CreateElement("id").SetText(gitSource.ScmId)
	source.CreateElement("credentialsId").SetText(gitSource.CredentialId)
	source.CreateElement("repoOwner").SetText(gitSource.Owner)
	source.CreateElement("repository").SetText(gitSource.Repo)
	source.CreateElement("serverUrl").SetText(gitSource.ApiUri)

	traits := source.CreateElement("traits")
	if gitSource.DiscoverBranches != 0 {
		traits.CreateElement("com.cloudbees.jenkins.plugins.bitbucket.BranchDiscoveryTrait>").
			CreateElement("strategyId").SetText(strconv.Itoa(gitSource.DiscoverBranches))
	}
	if gitSource.DiscoverPRFromOrigin != 0 {
		traits.CreateElement("com.cloudbees.jenkins.plugins.bitbucket.OriginPullRequestDiscoveryTrait").
			CreateElement("strategyId").SetText(strconv.Itoa(gitSource.DiscoverPRFromOrigin))
	}
	if gitSource.DiscoverPRFromForks != nil {
		forkTrait := traits.CreateElement("com.cloudbees.jenkins.plugins.bitbucket.ForkPullRequestDiscoveryTrait")
		forkTrait.CreateElement("strategyId").SetText(strconv.Itoa(gitSource.DiscoverPRFromForks.Strategy))
		trustClass := "com.cloudbees.jenkins.plugins.bitbucket.ForkPullRequestDiscoveryTrait$"

		if prTrust := PRDiscoverTrust(gitSource.DiscoverPRFromForks.Trust); prTrust.IsValid() {
			trustClass += prTrust.String()
		} else {
			klog.Warningf("invalid Bitbucket discover PR trust value: %d", prTrust.Value())
		}

		forkTrait.CreateElement("trust").CreateAttr("class", trustClass)
	}
	if gitSource.DiscoverTags {
		traits.CreateElement("com.cloudbees.jenkins.plugins.bitbucket.TagDiscoveryTrait")
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
		regexTraits.CreateAttr("plugin", "scm-api")
		regexTraits.CreateElement("regex").SetText(gitSource.RegexFilter)
	}
	return
}

func GetBitbucketServerSourceFromEtree(source *etree.Element) *devopsv1alpha3.BitbucketServerSource {
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
	if tagDiscoverTrait := traits.SelectElement(
		"com.cloudbees.jenkins.plugins.bitbucket.TagDiscoveryTrait"); tagDiscoverTrait != nil {
		s.DiscoverTags = true
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

		if prTrust := BitbucketPRDiscoverTrust(1).ParseFromString(trust[1]); prTrust.IsValid() {
			s.DiscoverPRFromForks = &devopsv1alpha3.DiscoverPRFromForks{
				Strategy: strategyId,
				Trust:    prTrust.Value(),
			}
		} else {
			klog.Warningf("invalid Bitbucket discover PR trust value: %s", trust[1])
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
