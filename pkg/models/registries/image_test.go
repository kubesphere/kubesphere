/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package registries

import (
	"testing"
)

func TestParseImage(t *testing.T) {
	type testImage struct {
		inputImageName string
		ExImage        Image
	}

	testImages := []testImage{
		{inputImageName: "dockerhub.qingcloud.com/kubesphere/test:v1", ExImage: Image{Domain: "dockerhub.qingcloud.com", Tag: "v1", Path: "kubesphere/test"}},
		{inputImageName: "harbor.devops.kubesphere.local:30280/library/tomcat:latest", ExImage: Image{Domain: "harbor.devops.kubesphere.local:30280", Tag: "latest", Path: "library/tomcat"}},
		{inputImageName: "zhuxiaoyang/nginx:v1", ExImage: Image{Domain: "docker.io", Tag: "v1", Path: "zhuxiaoyang/nginx"}},
		{inputImageName: "nginx", ExImage: Image{Domain: "docker.io", Tag: "latest", Path: "library/nginx"}},
		{inputImageName: "nginx:latest", ExImage: Image{Domain: "docker.io", Tag: "latest", Path: "library/nginx"}},
		{inputImageName: "kubesphere/ks-account:v2.1.0", ExImage: Image{Domain: "docker.io", Tag: "v2.1.0", Path: "kubesphere/ks-account"}},
		{inputImageName: "http://docker.io/nginx:latest", ExImage: Image{}},
		{inputImageName: "https://harbor.devops.kubesphere.local:30280/library/tomcat:latest", ExImage: Image{}},
		{inputImageName: "docker.io/nginx:latest:latest", ExImage: Image{}},
		{inputImageName: "nginx:8000:latest", ExImage: Image{}},
	}

	for _, image := range testImages {
		res, err := ParseImage(image.inputImageName)
		if err != nil {
			if res != image.ExImage {
				t.Fatalf("Get err %s", err)
			}
		}
		if res.Domain != image.ExImage.Domain {
			t.Fatalf("Doamin got %v, expected %v", res.Domain, image.ExImage.Domain)
		}

		if res.Tag != image.ExImage.Tag {
			t.Fatalf("Tag got %v, expected %v", res.Tag, image.ExImage.Tag)
		}

		if res.Path != image.ExImage.Path {
			t.Fatalf("Path got %v, expected %v", res.Path, image.ExImage.Path)
		}

	}

}

func TestStringWithoutScheme(t *testing.T) {
	type testRawUrl struct {
		Rawurl string
		ExUrl  string
	}
	testRawurls := []testRawUrl{
		{"http://dockerhub.qingcloud.com/kubesphere/nginx:v1", "dockerhub.qingcloud.com/kubesphere/nginx:v1"},
		{"https://dockerhub.qingcloud.com/kubesphere/nginx:v1", "dockerhub.qingcloud.com/kubesphere/nginx:v1"},
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
