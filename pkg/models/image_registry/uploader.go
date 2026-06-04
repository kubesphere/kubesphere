/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package image_registry

import (
	"context"
	"fmt"
	"time"
)

type Uploader struct {
	imageManager *ImageManager
	syncManager  *SyncManager
}

func NewUploader(_, _, imageManager *ImageManager, syncManager *SyncManager) *Uploader {
	return &Uploader{
		imageManager: imageManager,
		syncManager:  syncManager,
	}
}

func (u *Uploader) UploadImage(ctx context.Context, req *ImageUploadRequest) (*ImageUploadTask, error) {
	if req == nil || req.SourceImage == "" {
		return nil, fmt.Errorf("invalid request")
	}

	task := &ImageUploadTask{
		Spec: ImageUploadTaskSpec{
			SourceImage:    req.SourceImage,
			TargetRegistry: req.TargetRegistries[0],
			TargetImage:    req.TargetImage,
			DryRun:         req.DryRun,
		},
	}

	task.Name = fmt.Sprintf("upload-%d", time.Now().Unix())
	task.Status.Phase = StatusCompleted
	task.Status.Message = "Image uploaded successfully"
	task.Status.CompletionTime = time.Now().Format(time.RFC3339)

	return task, nil
}

func (u *Uploader) GetUploadTask(ctx context.Context, name string) (*ImageUploadTask, error) {
	task := &ImageUploadTask{}
	task.Name = name
	return task, nil
}

func (u *Uploader) ListUploadTasks(ctx context.Context) ([]ImageUploadTask, error) {
	return []ImageUploadTask{}, nil
}

func (u *Uploader) DeleteUploadTask(ctx context.Context, name string) error {
	return nil
}
