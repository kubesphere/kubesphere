package registries

import (
	"encoding/json"
	"fmt"
	"github.com/docker/distribution/manifest/schema2"
	log "github.com/golang/glog"
	"net/http"
)

// Digest returns the digest for an image.
func (r *Registry) ImageBlob(image Image, token string) (*ImageBlob, error) {

	url := r.GetBlobUrl(image)
	log.Info("registry.blobs.get url=" + url)

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
		log.Info("got response: " + string(resp.StatusCode) + string(respBody))
		return nil, fmt.Errorf("got response: %s", respBody)
	}

	imageBlob := &ImageBlob{}
	json.Unmarshal(respBody, imageBlob)

	return imageBlob, nil
}

func (r *Registry) GetBlobUrl(image Image) string {
	url := r.url("/v2/%s/blobs/%s", image.Path, image.Digest)
	return url
}
