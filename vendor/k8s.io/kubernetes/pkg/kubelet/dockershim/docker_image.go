/*
Copyright 2016 The Kubernetes Authors.

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

package dockershim

import (
	"context"
	"fmt"
	"net/http"

	dockertypes "github.com/docker/docker/api/types"
	dockerfilters "github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/jsonmessage"

	"github.com/golang/glog"
	runtimeapi "k8s.io/kubernetes/pkg/kubelet/apis/cri/runtime/v1alpha2"
	"k8s.io/kubernetes/pkg/kubelet/dockershim/libdocker"
)

// This file implements methods in ImageManagerService.

// ListImages lists existing images.
func (ds *dockerService) ListImages(_ context.Context, r *runtimeapi.ListImagesRequest) (*runtimeapi.ListImagesResponse, error) {
	filter := r.GetFilter()
	opts := dockertypes.ImageListOptions{}
	if filter != nil {
		if filter.GetImage().GetImage() != "" {
			opts.Filters = dockerfilters.NewArgs()
			opts.Filters.Add("reference", filter.GetImage().GetImage())
		}
	}

	images, err := ds.client.ListImages(opts)
	if err != nil {
		return nil, err
	}

	result := make([]*runtimeapi.Image, 0, len(images))
	for _, i := range images {
		apiImage, err := imageToRuntimeAPIImage(&i)
		if err != nil {
			glog.V(5).Infof("Failed to convert docker API image %+v to runtime API image: %v", i, err)
			continue
		}
		result = append(result, apiImage)
	}
	return &runtimeapi.ListImagesResponse{Images: result}, nil
}

// ImageStatus returns the status of the image, returns nil if the image doesn't present.
func (ds *dockerService) ImageStatus(_ context.Context, r *runtimeapi.ImageStatusRequest) (*runtimeapi.ImageStatusResponse, error) {
	image := r.GetImage()

	imageInspect, err := ds.client.InspectImageByRef(image.Image)
	if err != nil {
		if libdocker.IsImageNotFoundError(err) {
			return &runtimeapi.ImageStatusResponse{}, nil
		}
		return nil, err
	}

	imageStatus, err := imageInspectToRuntimeAPIImage(imageInspect)
	if err != nil {
		return nil, err
	}

	res := runtimeapi.ImageStatusResponse{Image: imageStatus}
	if r.GetVerbose() {
		res.Info = imageInspect.Config.Labels
	}
	return &res, nil
}

// PullImage pulls an image with authentication config.
func (ds *dockerService) PullImage(_ context.Context, r *runtimeapi.PullImageRequest) (*runtimeapi.PullImageResponse, error) {
	image := r.GetImage()
	auth := r.GetAuth()
	authConfig := dockertypes.AuthConfig{}

	if auth != nil {
		authConfig.Username = auth.Username
		authConfig.Password = auth.Password
		authConfig.ServerAddress = auth.ServerAddress
		authConfig.IdentityToken = auth.IdentityToken
		authConfig.RegistryToken = auth.RegistryToken
	}
	err := ds.client.PullImage(image.Image,
		authConfig,
		dockertypes.ImagePullOptions{},
	)
	if err != nil {
		return nil, filterHTTPError(err, image.Image)
	}

	imageRef, err := getImageRef(ds.client, image.Image)
	if err != nil {
		return nil, err
	}

	return &runtimeapi.PullImageResponse{ImageRef: imageRef}, nil
}

// RemoveImage removes the image.
func (ds *dockerService) RemoveImage(_ context.Context, r *runtimeapi.RemoveImageRequest) (*runtimeapi.RemoveImageResponse, error) {
	image := r.GetImage()
	// If the image has multiple tags, we need to remove all the tags
	// TODO: We assume image.Image is image ID here, which is true in the current implementation
	// of kubelet, but we should still clarify this in CRI.
	imageInspect, err := ds.client.InspectImageByID(image.Image)
	if err == nil && imageInspect != nil && len(imageInspect.RepoTags) > 1 {
		for _, tag := range imageInspect.RepoTags {
			if _, err := ds.client.RemoveImage(tag, dockertypes.ImageRemoveOptions{PruneChildren: true}); err != nil && !libdocker.IsImageNotFoundError(err) {
				return nil, err
			}
		}
		return &runtimeapi.RemoveImageResponse{}, nil
	}
	// dockerclient.InspectImageByID doesn't work with digest and repoTags,
	// it is safe to continue removing it since there is another check below.
	if err != nil && !libdocker.IsImageNotFoundError(err) {
		return nil, err
	}

	_, err = ds.client.RemoveImage(image.Image, dockertypes.ImageRemoveOptions{PruneChildren: true})
	if err != nil && !libdocker.IsImageNotFoundError(err) {
		return nil, err
	}
	return &runtimeapi.RemoveImageResponse{}, nil
}

// getImageRef returns the image digest if exists, or else returns the image ID.
func getImageRef(client libdocker.Interface, image string) (string, error) {
	img, err := client.InspectImageByRef(image)
	if err != nil {
		return "", err
	}
	if img == nil {
		return "", fmt.Errorf("unable to inspect image %s", image)
	}

	// Returns the digest if it exist.
	if len(img.RepoDigests) > 0 {
		return img.RepoDigests[0], nil
	}

	return img.ID, nil
}

func filterHTTPError(err error, image string) error {
	// docker/docker/pull/11314 prints detailed error info for docker pull.
	// When it hits 502, it returns a verbose html output including an inline svg,
	// which makes the output of kubectl get pods much harder to parse.
	// Here converts such verbose output to a concise one.
	jerr, ok := err.(*jsonmessage.JSONError)
	if ok && (jerr.Code == http.StatusBadGateway ||
		jerr.Code == http.StatusServiceUnavailable ||
		jerr.Code == http.StatusGatewayTimeout) {
		return fmt.Errorf("RegistryUnavailable: %v", err)
	}
	return err

}
