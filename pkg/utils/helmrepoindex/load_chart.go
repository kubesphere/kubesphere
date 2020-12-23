package helmrepoindex

import (
	"bytes"
	"context"
	"helm.sh/helm/v3/pkg/getter"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"net/url"
)

func LoadChart(ctx context.Context, u string, cred *v1alpha1.HelmRepoCredential) (*bytes.Buffer, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	//skipTLS := true
	//if cred.InsecureSkipTLSVerify != nil && !*cred.InsecureSkipTLSVerify {
	//	skipTLS = false
	//}

	indexURL := parsedURL.String()
	// TODO add user-agent
	g, _ := getter.NewHTTPGetter()
	resp, err := g.Get(indexURL,
		//getter.WithTimeout(5*time.Minute),
		getter.WithURL(u),
		//getter.WithInsecureSkipVerifyTLS(skipTLS),
		getter.WithTLSClientConfig(cred.CertFile, cred.KeyFile, cred.CAFile),
		getter.WithBasicAuth(cred.Username, cred.Password),
	)

	return resp, err
}
