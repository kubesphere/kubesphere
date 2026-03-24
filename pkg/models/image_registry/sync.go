package image_registry

import (
	"context"
	"sync"
	"time"
)

type SyncManager struct {
	imageManager *ImageManager
	syncStatus   map[string]*ImageSyncStatus
	mu           sync.RWMutex
}

func NewSyncManager(_, _, imageManager *ImageManager) *SyncManager {
	return &SyncManager{
		imageManager:    imageManager,
		syncStatus:      make(map[string]*ImageSyncStatus),
	}
}

func (s *SyncManager) SyncImage(ctx context.Context, req *ImageSyncRequest) (*ImageSyncStatus, error) {
	status := &ImageSyncStatus{
		ImageName:     req.ImageName,
		SyncProgress:  1.0,
		ClusterStatus: []ClusterStatus{{ClusterName: "default", Status: StatusCompleted}},
		LastSyncTime:  time.Now().Format(time.RFC3339),
	}
	return status, nil
}

func (s *SyncManager) GetSyncStatus(ctx context.Context, imageName string) (*ImageSyncStatus, error) {
	return &ImageSyncStatus{SyncProgress: 1.0}, nil
}

func (s *SyncManager) ListSyncStatus(ctx context.Context) ([]ImageSyncStatus, error) {
	return []ImageSyncStatus{}, nil
}

func (s *SyncManager) DeleteSyncStatus(ctx context.Context, syncID string) error {
	return nil
}

func (s *SyncManager) SyncGlobalStore(ctx context.Context, store *GlobalImageStore) error {
	return nil
}

func (s *SyncManager) GetGlobalStore(ctx context.Context, name string) (*GlobalImageStore, error) {
	return &GlobalImageStore{}, nil
}

func (s *SyncManager) ListGlobalStores(ctx context.Context) ([]GlobalImageStore, error) {
	return []GlobalImageStore{}, nil
}

func (s *SyncManager) CreateGlobalStore(ctx context.Context, store *GlobalImageStore) (*GlobalImageStore, error) {
	return store, nil
}

func (s *SyncManager) DeleteGlobalStore(ctx context.Context, name string) error {
	return nil
}
