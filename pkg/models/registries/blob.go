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
