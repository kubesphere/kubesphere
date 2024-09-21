package application

import (
	helmrepo "helm.sh/helm/v3/pkg/repo"
	appv2 "kubesphere.io/api/application/v2"
)

func LoadRepoIndexFormOci(u string, cred appv2.RepoCredential) (idx helmrepo.IndexFile, err error) {
	return *helmrepo.NewIndexFile(), nil
}
