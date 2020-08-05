/*
Copyright 2020 KubeSphere Authors

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

package registries

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/docker/distribution/manifest/schema2"
	"github.com/emicklei/go-restful"
	log "k8s.io/klog"
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
			log.Error(statusUnauthorized)
			return nil, restful.NewError(resp.StatusCode, statusUnauthorized)
		}
		log.Error("got response: " + string(resp.StatusCode) + string(respBody))
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
