package helm

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/getter"
	helmrepo "helm.sh/helm/v3/pkg/repo"
	"kubesphere.io/utils/s3"
	"sigs.k8s.io/yaml"
)

const IndexYaml = "index.yaml"

func LoadRepoIndex(ctx context.Context, u string, cred RepoCredential) (*helmrepo.IndexFile, error) {

	if !strings.HasSuffix(u, "/") {
		u = fmt.Sprintf("%s/%s", u, IndexYaml)
	} else {
		u = fmt.Sprintf("%s%s", u, IndexYaml)
	}

	resp, err := LoadData(ctx, u, cred)
	if err != nil {
		return nil, err
	}

	indexFile, err := loadIndex(resp.Bytes())
	if err != nil {
		return nil, err
	}

	return indexFile, nil
}

// loadIndex loads an index file and does minimal validity checking.
//
// This will fail if API Version is not set (ErrNoAPIVersion) or if the unmarshal fails.
func loadIndex(data []byte) (*helmrepo.IndexFile, error) {
	i := &helmrepo.IndexFile{}
	if err := yaml.Unmarshal(data, i); err != nil {
		return i, err
	}
	i.SortEntries()
	if i.APIVersion == "" {
		return i, helmrepo.ErrNoAPIVersion
	}
	return i, nil
}

func LoadData(ctx context.Context, u string, cred RepoCredential) (*bytes.Buffer, error) {
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
		// TODO add user-agent
		g, _ := getter.NewHTTPGetter()
		resp, err = g.Get(parsedURL.String(),
			getter.WithTimeout(5*time.Minute),
			getter.WithInsecureSkipVerifyTLS(cred.InsecureSkipTLSVerify),
			getter.WithTLSClientConfig(cred.CertFile, cred.KeyFile, cred.CAFile),
			getter.WithBasicAuth(cred.Username, cred.Password),
		)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

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

type RepoCredential struct {
	// chart repository username
	Username string `json:"username,omitempty"`
	// chart repository password
	Password string `json:"password,omitempty"`
	// identify HTTPS client using this SSL certificate file
	CertFile string `json:"certFile,omitempty"`
	// identify HTTPS client using this SSL key file
	KeyFile string `json:"keyFile,omitempty"`
	// verify certificates of HTTPS-enabled servers using this CA bundle
	CAFile string `json:"caFile,omitempty"`
	// skip tls certificate checks for the repository, default is ture
	InsecureSkipTLSVerify bool `json:"insecureSkipTLSVerify,omitempty"`

	S3Config `json:",inline"`
}

type S3Config struct {
	AccessKeyID     string `json:"accessKeyID,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
}
