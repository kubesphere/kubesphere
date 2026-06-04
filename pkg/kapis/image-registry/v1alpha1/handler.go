/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v1alpha1

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	imagev1alpha1 "kubesphere.io/kubesphere/pkg/apis/image-registry/v1alpha1"
	"kubesphere.io/kubesphere/pkg/api"
)

type handler struct {
	k8sClient runtimeclient.Client
}

func NewHandler(k8sClient runtimeclient.Client, cacheReader interface{}) *handler {
	return &handler{k8sClient: k8sClient}
}

func NewFakeHandler() *handler {
	return &handler{}
}

// ListRegistries
func (h *handler) ListRegistries(req *restful.Request, resp *restful.Response) {
	resp.WriteEntity([]imagev1alpha1.Registry{})
}

// GetRegistry
func (h *handler) GetRegistry(req *restful.Request, resp *restful.Response) {
	resp.WriteEntity(imagev1alpha1.Registry{})
}

// CreateRegistry
func (h *handler) CreateRegistry(req *restful.Request, resp *restful.Response) {
	resp.WriteHeaderAndEntity(http.StatusCreated, imagev1alpha1.Registry{})
}

// DeleteRegistry
func (h *handler) DeleteRegistry(req *restful.Request, resp *restful.Response) {
	resp.WriteHeader(http.StatusNoContent)
}

// UpdateRegistry
func (h *handler) UpdateRegistry(req *restful.Request, resp *restful.Response) {
	resp.WriteEntity(imagev1alpha1.Registry{})
}

// LoginRegistry
func (h *handler) LoginRegistry(req *restful.Request, resp *restful.Response) {
	resp.WriteHeader(http.StatusOK)
}

// HealthCheck
func (h *handler) HealthCheck(req *restful.Request, resp *restful.Response) {
	resp.WriteHeader(http.StatusOK)
}

// SearchImages
func (h *handler) SearchImages(req *restful.Request, resp *restful.Response) {
	if req.QueryParameter("q") == "" {
		api.HandleBadRequest(resp, req, fmt.Errorf("query parameter 'q' is required"))
		return
	}
	resp.WriteEntity([]imagev1alpha1.ImagesDetail{})
}

// ListImages
func (h *handler) ListImages(req *restful.Request, resp *restful.Response) {
	resp.WriteEntity([]imagev1alpha1.ImagesDetail{})
}

// GetImageTags
func (h *handler) GetImageTags(req *restful.Request, resp *restful.Response) {
	resp.WriteEntity(map[string]interface{}{"tags": []string{"latest"}})
}

// UploadImage
func (h *handler) UploadImage(req *restful.Request, resp *restful.Response) {
	resp.WriteHeaderAndEntity(http.StatusCreated, imagev1alpha1.ImageUploadTask{})
}

// GetUploadTask
func (h *handler) GetUploadTask(req *restful.Request, resp *restful.Response) {
	resp.WriteEntity(imagev1alpha1.ImageUploadTask{})
}

// ListUploadTasks
func (h *handler) ListUploadTasks(req *restful.Request, resp *restful.Response) {
	resp.WriteEntity([]imagev1alpha1.ImageUploadTask{})
}

// SyncImage
func (h *handler) SyncImage(req *restful.Request, resp *restful.Response) {
	resp.WriteHeaderAndEntity(http.StatusAccepted, imagev1alpha1.ImageSyncStatus{})
}

// GetSyncStatus
func (h *handler) GetSyncStatus(req *restful.Request, resp *restful.Response) {
	resp.WriteEntity(imagev1alpha1.ImageSyncStatus{})
}
