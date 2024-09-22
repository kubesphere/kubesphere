package application

import (
	"testing"

	appv2 "kubesphere.io/api/application/v2"
)

func TestLoadRepoIndexFormOci(t *testing.T) {

	url := "oci://localhost:5000/shipper-artifacts/demo/test02/helm-gen"
	cred := appv2.RepoCredential{
		Username: "",
		Password: "",
	}
	index, err := LoadRepoIndexFormOci(url, cred)
	if err != nil {
		t.Errorf("LoadRepoIndexFormOci() error: %s", err)
	}
	t.Log(len(index.Entries))
}
