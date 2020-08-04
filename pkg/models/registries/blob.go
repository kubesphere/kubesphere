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
	"github.com/docker/distribution/manifest/schema2"
	log "k8s.io/klog"
	"net/http"
)

// Digest returns the digest for an image.
func (r *Registry) ImageBlob(image Image, token string) (*ImageBlob, error) {
	if image.Path == "" {
		return nil, fmt.Errorf("image is required")
	}
	url := r.GetBlobUrl(image)

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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		log.Error("got response: " + string(resp.StatusCode) + string(respBody))
		return nil, fmt.Errorf("got image blob faild")
	}

	imageBlob := &ImageBlob{}
	err = json.Unmarshal(respBody, imageBlob)

	return imageBlob, err
}

func (r *Registry) GetBlobUrl(image Image) string {
	url := r.url("/v2/%s/blobs/%s", image.Path, image.Digest)
	return url
}
