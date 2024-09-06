/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package registries

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/distribution/reference"
	"github.com/opencontainers/go-digest"
	"k8s.io/klog/v2"
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
	return strings.Trim(s, "/")
}

func ParseDockerURL(rawurl string) (*DockerURL, error) {
	url, err := url.Parse(rawurl)
	if err != nil {
		klog.Errorf("%+v", err)
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
