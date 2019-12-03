package registries

import (
	"testing"
)

func TestParseImage(t *testing.T) {
	type image struct {
		ImageName string
		ExDomain  string
		ExTag     string
		ExPath    string
	}

	testImages := []image{
		{ImageName: "dockerhub.qingcloud.com/kubesphere/test:v1", ExDomain: "dockerhub.qingcloud.com", ExTag: "v1", ExPath: "kubesphere/test"},
		{ImageName: "harbor.devops.kubesphere.local:30280/library/tomcat:latest", ExDomain: "harbor.devops.kubesphere.local:30280", ExTag: "latest", ExPath: "library/tomcat"},
		{ImageName: "zhuxiaoyang/nginx:v1", ExDomain: "docker.io", ExTag: "v1", ExPath: "zhuxiaoyang/nginx"},
		{ImageName: "nginx", ExDomain: "docker.io", ExTag: "latest", ExPath: "library/nginx"},
		{ImageName: "nginx:latest", ExDomain: "docker.io", ExTag: "latest", ExPath: "library/nginx"},
		{ImageName: "kubesphere/ks-account:v2.1.0", ExDomain: "docker.io", ExTag: "v2.1.0", ExPath: "kubesphere/ks-account"},
	}

	for _, image := range testImages {
		res, err := ParseImage(image.ImageName)
		if err != nil {
			t.Fatalf("Get err %s", err)
		}

		if res.Domain != image.ExDomain {
			t.Fatalf("Doamin got %v, expected %v", res.Domain, image.ExDomain)
		}

		if res.Tag != image.ExTag {
			t.Fatalf("Tag got %v, expected %v", res.Tag, image.ExTag)
		}

		if res.Path != image.ExPath {
			t.Fatalf("Path got %v, expected %v", res.Path, image.ExPath)
		}
	}

	invalidImage := []image{
		{ImageName: "http://docker.io/nginx:latest"},
		{ImageName: "https://harbor.devops.kubesphere.local:30280/library/tomcat:latest"},
		{ImageName: "docker.io/nginx:latest:latest"},
		{ImageName: "nginx:8000:latest"},
	}

	for _, image := range invalidImage {
		_, err := ParseImage(image.ImageName)
		if err == nil {
			t.Fatalf("Parse invalid image but without get any error")
		}
	}

}

func TestStringWithoutScheme(t *testing.T) {
	type testRawurl struct {
		Rawurl string
		ExUrl  string
	}
	testRawurls := []testRawurl{
		{"http://dockerhub.qingcloud.com/kubesphere/nginx:v1", "dockerhub.qingcloud.com/kubesphere/nginx:v1"},
		{"https://dockerhub.qingcloud.com/kubesphere/nginx:v1", "dockerhub.qingcloud.com/kubesphere/nginx:v1"},
		{"dockerhub.qingcloud.com/kubesphere/nginx:v1", "dockerhub.qingcloud.com/kubesphere/nginx:v1"},
		{"http://harbor.devops.kubesphere.local:30280/library/tomcat:latest", "harbor.devops.kubesphere.local:30280/library/tomcat:latest"},
		{"https://harbor.devops.kubesphere.local:30280/library/tomcat:latest", "harbor.devops.kubesphere.local:30280/library/tomcat:latest"},
	}

	for _, rawurl := range testRawurls {
		dockerurl, err := ParseDockerURL(rawurl.Rawurl)
		if err != nil {
			t.Fatalf("Get err %s", err)
		}

		imageName := dockerurl.StringWithoutScheme()

		if imageName != rawurl.ExUrl {
			t.Fatalf("imagename got %v, expected %v", imageName, rawurl.ExUrl)
		}
	}
}
