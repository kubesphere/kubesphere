package oci

import (
	"context"
	"strings"
	"testing"
	"time"

	"oras.land/oras-go/pkg/registry"
)

func TestRegistry_Api(t *testing.T) {
	reg, err := NewRegistry("registry.cn-shanghai.aliyuncs.com",
		WithTimeout(5*time.Second),
		WithBasicAuth("", ""),
		WithInsecureSkipVerifyTLS(true))
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	ctx := context.Background()
	err = reg.Ping(ctx)
	if err != nil {
		t.Fatalf("Registry.Ping() error = %v", err)
	}

	var registryTags []string

	repo, err := reg.Repository(ctx, "kube-shipper/shipper-ui")
	if err != nil {
		t.Fatalf("Registry.Repository() error = %v", err)
	}
	registryTags, err = registry.Tags(ctx, repo)
	if err != nil {
		t.Fatalf("Registry.Repository().Tags() error = %v", err)
	}
	t.Log(len(registryTags))

	err = reg.Repositories(ctx, "", func(repos []string) error {
		for _, repo := range repos {
			if subRepo, found := strings.CutPrefix(repo, ""); found {
				t.Log(subRepo)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Registry.Repositories() error = %v", err)
	}

}
