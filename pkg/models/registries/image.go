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
	"fmt"
	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
	log "k8s.io/klog"
	"net/url"
	"strings"
)

// Image holds information about an image.
type Image struct {
	Domain string
	Path   string
	Tag    string
	Digest digest.Digest
	named  reference.Named
}

// String returns the string representation of an image.
func (i *Image) String() string {
	return i.named.String()
}

// Reference returns either the digest if it is non-empty or the tag for the image.
func (i *Image) Reference() string {
	if len(i.Digest.String()) > 1 {
		return i.Digest.String()
	}

	return i.Tag
}

type DockerURL struct {
	*url.URL
}

func (u *DockerURL) StringWithoutScheme() string {
	u.Scheme = ""
	s := u.String()
	return strings.Trim(s, "//")
}

func ParseDockerURL(rawurl string) (*DockerURL, error) {
	url, err := url.Parse(rawurl)
	if err != nil {
		log.Errorf("%+v", err)
		return nil, err
	}
	return &DockerURL{URL: url}, nil
}

// ParseImage returns an Image struct with all the values filled in for a given image.
// example : localhost:5000/nginx:latest, nginx:perl etc.
func ParseImage(image string) (i Image, err error) {
	// Parse the image name and tag.
	named, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return Image{}, fmt.Errorf("parsing image %q failed: %v", image, err)
	}
	// Add the latest lag if they did not provide one.
	named = reference.TagNameOnly(named)

	i = Image{
		named:  named,
		Domain: reference.Domain(named),
		Path:   reference.Path(named),
	}

	// Add the tag if there was one.
	if tagged, ok := named.(reference.Tagged); ok {
		i.Tag = tagged.Tag()
	}

	// Add the digest if there was one.
	if canonical, ok := named.(reference.Canonical); ok {
		i.Digest = canonical.Digest()
	}

	return i, nil
}
