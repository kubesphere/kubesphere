package registries

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/docker/distribution/manifest/schema2"
	log "github.com/golang/glog"
)

// Digest returns the digest for an image.
func (r *Registry) ImageManifest(image Image, token string) (*ImageManifest, error, int) {
	url := r.GetDigestUrl(image)
	log.Info("registry.manifests.get url=" + url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}

	req.Header.Add("Accept", schema2.MediaTypeManifest)
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err, http.StatusInternalServerError
	}
	defer resp.Body.Close()

	respBody, _ := GetRespBody(resp)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		log.Info("got response: " + string(resp.StatusCode) + string(respBody))
		return nil, fmt.Errorf("%s", respBody), resp.StatusCode
	}

	imageManifest := &ImageManifest{}
	json.Unmarshal(respBody, imageManifest)

	return imageManifest, nil, http.StatusOK
}

func (r *Registry) GetDigestUrl(image Image) string {
	url := r.url("/v2/%s/manifests/%s", image.Path, image.Tag)
	return url
}
