package registries

import (
	"testing"
)

func TestRegistryVerify(t *testing.T) {
	type testRegistry struct {
		Auth   AuthInfo
		Result bool
	}

	// some registry can not login with guest.
	registries := []testRegistry{
		{Auth: AuthInfo{Username: "guest", Password: "guest", ServerHost: "docker.io"}, Result: true},
		{Auth: AuthInfo{Username: "guest", Password: "guest", ServerHost: "dockerhub.qingcloud.com"}, Result: true},
		{Auth: AuthInfo{Username: "guest", Password: "guest", ServerHost: "https://dockerhub.qingcloud.com"}, Result: true},
		{Auth: AuthInfo{Username: "guest", Password: "guest", ServerHost: "http://dockerhub.qingcloud.com"}, Result: false},
		{Auth: AuthInfo{Username: "guest", Password: "guest", ServerHost: "registry.cn-hangzhou.aliyuncs.com"}, Result: false},
	}

	for _, registry := range registries {
		err := RegistryVerify(registry.Auth)
		if registry.Result == true && err != nil {
			t.Fatalf("Get err %s", err)
		}

		if registry.Result == false && err == nil {
			t.Fatalf("Input Wrong data but without any error.")
		}
	}
}
