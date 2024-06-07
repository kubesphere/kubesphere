/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"sort"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type Registryer interface {
	// list repository tags
	ListRepositoryTags(image string) (RepositoryTags, error)

	// get image config
	Config(image string) (*v1.ConfigFile, error)
}

type registryer struct {
	opts options
}

func NewRegistryer(opts ...Option) *registryer {
	return &registryer{
		opts: makeOptions(opts...),
	}
}

func (r *registryer) ListRepositoryTags(src string) (RepositoryTags, error) {
	repo, err := name.NewRepository(src, r.opts.name...)
	if err != nil {
		return RepositoryTags{}, err
	}

	tags, err := remote.List(repo, r.opts.remote...)
	if err != nil {
		return RepositoryTags{}, err
	}
	sort.SliceStable(tags, func(i, j int) bool {
		return i > j
	})

	return RepositoryTags{
		Registry:   repo.RegistryStr(),
		Repository: repo.RepositoryStr(),
		Tags:       tags,
		Total:      len(tags),
	}, nil
}

func (r *registryer) Config(image string) (*v1.ConfigFile, error) {
	img, _, err := r.getImage(image)
	if err != nil {
		return nil, err
	}

	configFile, err := img.ConfigFile()
	if err != nil {
		return nil, err
	}

	return configFile, nil
}

func (r *registryer) getImage(reference string) (v1.Image, name.Reference, error) {
	ref, err := name.ParseReference(reference, r.opts.name...)
	if err != nil {
		return nil, nil, err
	}

	img, err := remote.Image(ref, r.opts.remote...)
	if err != nil {
		return nil, nil, err
	}

	return img, ref, nil
}
