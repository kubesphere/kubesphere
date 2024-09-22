package application

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/registry"
	helmrepo "helm.sh/helm/v3/pkg/repo"
	"k8s.io/klog/v2"
	appv2 "kubesphere.io/api/application/v2"
	"kubesphere.io/kubesphere/pkg/simple/client/oci"
)

func LoadRepoIndexFormOci(u string, cred appv2.RepoCredential) (idx helmrepo.IndexFile, err error) {
	if !registry.IsOCI(u) {
		return idx, fmt.Errorf("invalid oci URL format: %s", u)
	}

	parsedURL, err := url.Parse(u)
	if err != nil {
		klog.Errorf("invalid repo URL format: %s, err:%v", u, err)
		return idx, err
	}

	skipTLS := true
	if cred.InsecureSkipTLSVerify != nil && !*cred.InsecureSkipTLSVerify {
		skipTLS = false
	}

	reg, err := oci.NewRegistry(parsedURL.Host,
		oci.WithTimeout(5*time.Second),
		oci.WithBasicAuth(cred.Username, cred.Password),
		oci.WithInsecureSkipVerifyTLS(skipTLS))
	if err != nil {
		return idx, err
	}

	ctx := context.Background()

	repoPath := strings.TrimSuffix(parsedURL.Path, "/")
	repoPath = strings.TrimPrefix(repoPath, "/")
	var repoCharts []string
	err = reg.Repositories(ctx, "", func(repos []string) error {
		cutPrefix := repoPath
		if cutPrefix != "" {
			cutPrefix = cutPrefix + "/"
		}
		for _, repo := range repos {
			if subRepo, found := strings.CutPrefix(repo, cutPrefix); found && subRepo != "" {
				if !strings.Contains(subRepo, "/") {
					repoCharts = append(repoCharts, fmt.Sprintf("%s/%s", repoPath, subRepo))
				}
			}
		}
		return nil
	})

	if err != nil {
		return idx, err
	}
	if len(repoCharts) == 0 {
		return idx, nil
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLS},
		Proxy:           http.ProxyFromEnvironment,
	}

	opts := []registry.ClientOption{registry.ClientOptHTTPClient(&http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	})}

	if reg.PlainHTTP {
		opts = append(opts, registry.ClientOptPlainHTTP())
	}

	client, err := registry.NewClient(opts...)

	if err != nil {
		return idx, err
	}

	if cred.Username != "" || cred.Password != "" {
		err = client.Login(parsedURL.Host,
			registry.LoginOptBasicAuth(cred.Username, cred.Password),
			registry.LoginOptInsecure(reg.PlainHTTP))

		if err != nil {
			return idx, err
		}
	}

	index := helmrepo.NewIndexFile()
	for _, repoChart := range repoCharts {
		tags, err := client.Tags(fmt.Sprintf("%s/%s", parsedURL.Host, repoChart))
		if err != nil {
			klog.Errorf("An error occurred to load tags from repository: %s/%s,err:%v", parsedURL.Host, repoChart, err)
			continue
		}
		if len(tags) == 0 {
			klog.Errorf("Unable to locate any tags in provided repository: %s/%s,err:%v", parsedURL.Host, repoChart, err)
			continue
		}

		for _, tag := range tags {
			pullRef := fmt.Sprintf("%s/%s:%s", parsedURL.Host, repoChart, tag)
			pullResult, err := client.Pull(pullRef)
			if err != nil {
				klog.Errorf("An error occurred to pull chart from repository: %s,err:%v", pullRef, err)
				continue
			}

			baseUrl := fmt.Sprintf("%s://%s", registry.OCIScheme, pullRef)
			hash := strings.TrimPrefix(pullResult.Chart.Digest, "sha256:")
			if err := index.MustAdd(pullResult.Chart.Meta, "", baseUrl, hash); err != nil {
				klog.Errorf("failed adding chart metadata to index with repository: %s,err:%v", pullRef, err)
				continue
			}
		}
	}

	return *index, nil
}
