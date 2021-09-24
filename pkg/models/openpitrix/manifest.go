package openpitrix

import (
	"kubesphere.io/api/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
)

type ManifestCustomResourceInterface interface {
	CreateManifest(repo *v1alpha1.Manifest) (*CreateRepoResponse, error)
	DeleteManifest(id string) error
	ValidateManifest(u string, request *v1alpha1.Manifest) (*ValidateRepoResponse, error)
	ModifyManifest(id string, request *v1alpha1.Manifest) error
	ListManifest(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
}

