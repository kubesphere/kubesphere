package internal

import (
	"github.com/beevik/etree"
	"k8s.io/klog"

	devopsv1alpha3 "kubesphere.io/api/devops/v1alpha3"
)

func AppendSvnSourceToEtree(source *etree.Element, svnSource *devopsv1alpha3.SvnSource) {
	if svnSource == nil {
		klog.Warning("please provide SVN source when the sourceType is SVN")
		return
	}
	source.CreateAttr("class", "jenkins.scm.impl.subversion.SubversionSCMSource")
	source.CreateAttr("plugin", "subversion")
	source.CreateElement("id").SetText(svnSource.ScmId)
	if svnSource.CredentialId != "" {
		source.CreateElement("credentialsId").SetText(svnSource.CredentialId)
	}
	if svnSource.Remote != "" {
		source.CreateElement("remoteBase").SetText(svnSource.Remote)
	}
	if svnSource.Includes != "" {
		source.CreateElement("includes").SetText(svnSource.Includes)
	}
	if svnSource.Excludes != "" {
		source.CreateElement("excludes").SetText(svnSource.Excludes)
	}
	return
}

func AppendSingleSvnSourceToEtree(source *etree.Element, svnSource *devopsv1alpha3.SingleSvnSource) {
	if svnSource == nil {
		klog.Warning("please provide SingleSvn source when the sourceType is SingleSvn")
		return
	}
	source.CreateAttr("class", "jenkins.scm.impl.SingleSCMSource")
	source.CreateAttr("plugin", "scm-api")
	source.CreateElement("id").SetText(svnSource.ScmId)
	source.CreateElement("name").SetText("master")

	scm := source.CreateElement("scm")
	scm.CreateAttr("class", "hudson.scm.SubversionSCM")
	scm.CreateAttr("plugin", "subversion")

	location := scm.CreateElement("locations").CreateElement("hudson.scm.SubversionSCM_-ModuleLocation")
	if svnSource.Remote != "" {
		location.CreateElement("remote").SetText(svnSource.Remote)
	}
	if svnSource.CredentialId != "" {
		location.CreateElement("credentialsId").SetText(svnSource.CredentialId)
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

func GetSingleSvnSourceFromEtree(source *etree.Element) *devopsv1alpha3.SingleSvnSource {
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

func GetSvnSourcefromEtree(source *etree.Element) *devopsv1alpha3.SvnSource {
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
