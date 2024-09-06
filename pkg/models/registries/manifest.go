/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package registries

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/docker/distribution/manifest/schema2"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"
)

var statusUnauthorized = "Not found or unauthorized"

// Digest returns the digest for an image.
func (r *Registry) ImageManifest(image Image, token string) (*ImageManifest, error) {
	url := r.GetDigestUrl(image)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", schema2.MediaTypeManifest)
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := GetRespBody(resp)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusUnauthorized {
			klog.Error(statusUnauthorized)
			return nil, restful.NewError(resp.StatusCode, statusUnauthorized)
		}
		klog.Errorf("got response: statusCode is '%d', body is '%s'\n", resp.StatusCode, respBody)
		return nil, restful.NewError(resp.StatusCode, "got image manifest failed")
	}

	imageManifest := &ImageManifest{}
	err = json.Unmarshal(respBody, imageManifest)

	return imageManifest, err
}

func (r *Registry) GetDigestUrl(image Image) string {
	url := r.url("/v2/%s/manifests/%s", image.Path, image.Tag)
	return url
}
