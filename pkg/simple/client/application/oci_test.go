package application

import (
	"testing"

	appv2 "kubesphere.io/api/application/v2"
)

func TestLoadRepoIndexFromOci(t *testing.T) {

	url := "oci://localhost:5000/shipper-artifacts/demo/test02/helm-gen"
	cred := appv2.RepoCredential{
		Username: "",
		Password: "",
	}
	index, err := LoadRepoIndexFromOci(url, cred)
	if err != nil {
		t.Errorf("LoadRepoIndexFromOci() error: %s", err)
	}
	t.Log(len(index.Entries))
}
