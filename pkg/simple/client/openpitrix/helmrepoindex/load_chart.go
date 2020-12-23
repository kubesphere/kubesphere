/*
Copyright 2020 The KubeSphere Authors.

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

package helmrepoindex

import (
	"bytes"
	"context"
	"fmt"
	"helm.sh/helm/v3/pkg/getter"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"net/url"
	"strings"
	"time"
)

func parseS3Url(parse *url.URL) (region, endpoint, bucket, path string) {
	if strings.HasPrefix(parse.Host, "s3.") {
		region = strings.Split(parse.Host, ".")[1]
		endpoint = fmt.Sprintf("https://%s", parse.Host)
	} else {
		region = "us-east-1"
		endpoint = fmt.Sprintf("http://%s", parse.Host)
	}
	parts := strings.Split(strings.TrimPrefix(parse.Path, "/"), "/")
	if len(parts) > 0 {
		bucket = parts[0]
		path = strings.Join(parts[1:], "/")
	} else {
		bucket = parse.Path
	}

	return region, endpoint, bucket, path
}

func loadData(ctx context.Context, u string, cred *v1alpha1.HelmRepoCredential) (*bytes.Buffer, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	var resp *bytes.Buffer
	if strings.HasPrefix(u, "s3://") {
		region, endpoint, bucket, p := parseS3Url(parsedURL)
		client, err := s3.NewS3Client(&s3.Options{
			Endpoint:        endpoint,
			Bucket:          bucket,
			Region:          region,
			AccessKeyID:     cred.AccessKeyID,
			SecretAccessKey: cred.SecretAccessKey,
			DisableSSL:      !strings.HasPrefix(region, "https://"),
			ForcePathStyle:  true,
		})

		if err != nil {
			return nil, err
		}

		data, err := client.Read(p)
		if err != nil {
			return nil, err
		}

		resp = bytes.NewBuffer(data)
	} else {
		skipTLS := true
		if cred.InsecureSkipTLSVerify != nil && !*cred.InsecureSkipTLSVerify {
			skipTLS = false
		}

		indexURL := parsedURL.String()
		// TODO add user-agent
		g, _ := getter.NewHTTPGetter()
		resp, err = g.Get(indexURL,
			getter.WithTimeout(5*time.Minute),
			getter.WithURL(u),
			getter.WithInsecureSkipVerifyTLS(skipTLS),
			getter.WithTLSClientConfig(cred.CertFile, cred.KeyFile, cred.CAFile),
			getter.WithBasicAuth(cred.Username, cred.Password),
		)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func LoadChart(ctx context.Context, u string, cred *v1alpha1.HelmRepoCredential) (*bytes.Buffer, error) {
	return loadData(ctx, u, cred)
}
