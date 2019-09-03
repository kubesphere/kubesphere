package registries

import (
	"testing"
)

func TestCreateRegistryClient(t *testing.T) {
	type imageInfo struct {
		Username string
		Password string
		Domain   string
		ExDomain string
		ExUrl    string
	}

	testImages := []imageInfo{
		{Domain: "kubesphere.io", ExDomain: "kubesphere.io", ExUrl: "http://kubesphere.io"},
		{Domain: "127.0.0.1:5000", ExDomain: "127.0.0.1:5000", ExUrl: "http://127.0.0.1:5000"},
		{Username: "Username", Password: "Password", Domain: "docker.io", ExDomain: "registry-1.docker.io", ExUrl: "https://registry-1.docker.io"},
		{Domain: "harbor.devops.kubesphere.local:30280", ExDomain: "harbor.devops.kubesphere.local:30280", ExUrl: "http://harbor.devops.kubesphere.local:30280"},
	}

	for _, testImage := range testImages {
		reg, err := CreateRegistryClient(testImage.Username, testImage.Password, testImage.Domain)
		if err != nil {
			t.Fatalf("Get err %s", err)
		}

		if reg.Domain != testImage.ExDomain {
			t.Fatalf("Doamin got %v, expected %v", reg.Domain, testImage.ExDomain)
		}

		if reg.URL != testImage.ExUrl {
			t.Fatalf("URL got %v, expected %v", reg.URL, testImage.ExUrl)
		}

	}

	testImage := Image{Domain: "docker.io", Path: "library/alpine", Tag: "latest"}
	r, err := CreateRegistryClient("", "", "docker.io")
	if err != nil {
		t.Fatalf("Could not get client: %s", err)
	}

	digestUrl := r.GetDigestUrl(testImage)

	// Get token.
	token, err := r.Token(digestUrl)
	if err != nil || token == "" {
		t.Fatalf("Could not get token: %s", err)
	}

	d, err := r.ImageManifest(testImage, token)
	if err != nil {
		t.Fatalf("Could not get digest: %s", err)
	}

	if d == nil {
		t.Error("Empty digest received")
	}
}
