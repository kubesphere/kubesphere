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

package common

import (
	"fmt"
	"path"

	. "github.com/onsi/ginkgo"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/kubernetes/test/e2e/framework"
	imageutils "k8s.io/kubernetes/test/utils/image"
)

const (
	volumePath = "/test-volume"
)

var (
	testImageRootUid    = imageutils.GetE2EImage(imageutils.Mounttest)
	testImageNonRootUid = imageutils.GetE2EImage(imageutils.MounttestUser)
)

var _ = Describe("[sig-storage] EmptyDir volumes", func() {
	f := framework.NewDefaultFramework("emptydir")

	Context("when FSGroup is specified", func() {
		It("new files should be created with FSGroup ownership when container is root", func() {
			doTestSetgidFSGroup(f, testImageRootUid, v1.StorageMediumMemory)
		})

		It("new files should be created with FSGroup ownership when container is non-root", func() {
			doTestSetgidFSGroup(f, testImageNonRootUid, v1.StorageMediumMemory)
		})

		It("nonexistent volume subPath should have the correct mode and owner using FSGroup", func() {
			doTestSubPathFSGroup(f, testImageNonRootUid, v1.StorageMediumMemory)
		})

		It("files with FSGroup ownership should support (root,0644,tmpfs)", func() {
			doTest0644FSGroup(f, testImageRootUid, v1.StorageMediumMemory)
		})

		It("volume on default medium should have the correct mode using FSGroup", func() {
			doTestVolumeModeFSGroup(f, testImageRootUid, v1.StorageMediumDefault)
		})

		It("volume on tmpfs should have the correct mode using FSGroup", func() {
			doTestVolumeModeFSGroup(f, testImageRootUid, v1.StorageMediumMemory)
		})
	})

	/*
		    Testname: volume-emptydir-mode-tmpfs
		    Description: For a Pod created with an 'emptyDir' Volume with 'medium'
			of 'Memory', ensure the volume has 0777 unix file permissions and tmpfs
			mount type.
	*/
	framework.ConformanceIt("volume on tmpfs should have the correct mode", func() {
		doTestVolumeMode(f, testImageRootUid, v1.StorageMediumMemory)
	})

	/*
		    Testname: volume-emptydir-root-0644-tmpfs
		    Description: For a Pod created with an 'emptyDir' Volume with 'medium'
			of 'Memory', ensure a root owned file with 0644 unix file permissions
			is created correctly, has tmpfs mount type, and enforces the permissions.
	*/
	framework.ConformanceIt("should support (root,0644,tmpfs)", func() {
		doTest0644(f, testImageRootUid, v1.StorageMediumMemory)
	})

	/*
		    Testname: volume-emptydir-root-0666-tmpfs
		    Description: For a Pod created with an 'emptyDir' Volume with 'medium'
			of 'Memory', ensure a root owned file with 0666 unix file permissions
			is created correctly, has tmpfs mount type, and enforces the permissions.
	*/
	framework.ConformanceIt("should support (root,0666,tmpfs)", func() {
		doTest0666(f, testImageRootUid, v1.StorageMediumMemory)
	})

	/*
		    Testname: volume-emptydir-root-0777-tmpfs
		    Description: For a Pod created with an 'emptyDir' Volume with 'medium'
			of 'Memory', ensure a root owned file with 0777 unix file permissions
			is created correctly, has tmpfs mount type, and enforces the permissions.
	*/
	framework.ConformanceIt("should support (root,0777,tmpfs)", func() {
		doTest0777(f, testImageRootUid, v1.StorageMediumMemory)
	})

	/*
		    Testname: volume-emptydir-user-0644-tmpfs
		    Description: For a Pod created with an 'emptyDir' Volume with 'medium'
			of 'Memory', ensure a user owned file with 0644 unix file permissions
			is created correctly, has tmpfs mount type, and enforces the permissions.
	*/
	framework.ConformanceIt("should support (non-root,0644,tmpfs)", func() {
		doTest0644(f, testImageNonRootUid, v1.StorageMediumMemory)
	})

	/*
		    Testname: volume-emptydir-user-0666-tmpfs
		    Description: For a Pod created with an 'emptyDir' Volume with 'medium'
			of 'Memory', ensure a user owned file with 0666 unix file permissions
			is created correctly, has tmpfs mount type, and enforces the permissions.
	*/
	framework.ConformanceIt("should support (non-root,0666,tmpfs)", func() {
		doTest0666(f, testImageNonRootUid, v1.StorageMediumMemory)
	})

	/*
		    Testname: volume-emptydir-user-0777-tmpfs
		    Description: For a Pod created with an 'emptyDir' Volume with 'medium'
			of 'Memory', ensure a user owned file with 0777 unix file permissions
			is created correctly, has tmpfs mount type, and enforces the permissions.
	*/
	framework.ConformanceIt("should support (non-root,0777,tmpfs)", func() {
		doTest0777(f, testImageNonRootUid, v1.StorageMediumMemory)
	})

	/*
		    Testname: volume-emptydir-mode
		    Description: For a Pod created with an 'emptyDir' Volume, ensure the
			volume has 0777 unix file permissions.
	*/
	framework.ConformanceIt("volume on default medium should have the correct mode", func() {
		doTestVolumeMode(f, testImageRootUid, v1.StorageMediumDefault)
	})

	/*
		    Testname: volume-emptydir-root-0644
		    Description: For a Pod created with an 'emptyDir' Volume, ensure a
			root owned file with 0644 unix file permissions is created and enforced
			correctly.
	*/
	framework.ConformanceIt("should support (root,0644,default)", func() {
		doTest0644(f, testImageRootUid, v1.StorageMediumDefault)
	})

	/*
		    Testname: volume-emptydir-root-0666
		    Description: For a Pod created with an 'emptyDir' Volume, ensure a
			root owned file with 0666 unix file permissions is created and enforced
			correctly.
	*/
	framework.ConformanceIt("should support (root,0666,default)", func() {
		doTest0666(f, testImageRootUid, v1.StorageMediumDefault)
	})

	/*
		    Testname: volume-emptydir-root-0777
		    Description: For a Pod created with an 'emptyDir' Volume, ensure a
			root owned file with 0777 unix file permissions is created and enforced
			correctly.
	*/
	framework.ConformanceIt("should support (root,0777,default)", func() {
		doTest0777(f, testImageRootUid, v1.StorageMediumDefault)
	})

	/*
		    Testname: volume-emptydir-user-0644
		    Description: For a Pod created with an 'emptyDir' Volume, ensure a
			user owned file with 0644 unix file permissions is created and enforced
			correctly.
	*/
	framework.ConformanceIt("should support (non-root,0644,default)", func() {
		doTest0644(f, testImageNonRootUid, v1.StorageMediumDefault)
	})

	/*
		    Testname: volume-emptydir-user-0666
		    Description: For a Pod created with an 'emptyDir' Volume, ensure a
			user owned file with 0666 unix file permissions is created and enforced
			correctly.
	*/
	framework.ConformanceIt("should support (non-root,0666,default)", func() {
		doTest0666(f, testImageNonRootUid, v1.StorageMediumDefault)
	})

	/*
		    Testname: volume-emptydir-user-0777
		    Description: For a Pod created with an 'emptyDir' Volume, ensure a
			user owned file with 0777 unix file permissions is created and enforced
			correctly.
	*/
	framework.ConformanceIt("should support (non-root,0777,default)", func() {
		doTest0777(f, testImageNonRootUid, v1.StorageMediumDefault)
	})
})

const (
	containerName = "test-container"
	volumeName    = "test-volume"
)

func doTestSetgidFSGroup(f *framework.Framework, image string, medium v1.StorageMedium) {
	var (
		filePath = path.Join(volumePath, "test-file")
		source   = &v1.EmptyDirVolumeSource{Medium: medium}
		pod      = testPodWithVolume(testImageRootUid, volumePath, source)
	)

	pod.Spec.Containers[0].Args = []string{
		fmt.Sprintf("--fs_type=%v", volumePath),
		fmt.Sprintf("--new_file_0660=%v", filePath),
		fmt.Sprintf("--file_perm=%v", filePath),
		fmt.Sprintf("--file_owner=%v", filePath),
	}

	fsGroup := int64(123)
	pod.Spec.SecurityContext.FSGroup = &fsGroup

	msg := fmt.Sprintf("emptydir 0644 on %v", formatMedium(medium))
	out := []string{
		"perms of file \"/test-volume/test-file\": -rw-rw----",
		"content of file \"/test-volume/test-file\": mount-tester new file",
		"owner GID of \"/test-volume/test-file\": 123",
	}
	if medium == v1.StorageMediumMemory {
		out = append(out, "mount type of \"/test-volume\": tmpfs")
	}
	f.TestContainerOutput(msg, pod, 0, out)
}

func doTestSubPathFSGroup(f *framework.Framework, image string, medium v1.StorageMedium) {
	var (
		subPath = "test-sub"
		source  = &v1.EmptyDirVolumeSource{Medium: medium}
		pod     = testPodWithVolume(image, volumePath, source)
	)

	pod.Spec.Containers[0].Args = []string{
		fmt.Sprintf("--fs_type=%v", volumePath),
		fmt.Sprintf("--file_perm=%v", volumePath),
		fmt.Sprintf("--file_owner=%v", volumePath),
		fmt.Sprintf("--file_mode=%v", volumePath),
	}

	pod.Spec.Containers[0].VolumeMounts[0].SubPath = subPath

	fsGroup := int64(123)
	pod.Spec.SecurityContext.FSGroup = &fsGroup

	msg := fmt.Sprintf("emptydir subpath on %v", formatMedium(medium))
	out := []string{
		"perms of file \"/test-volume\": -rwxrwxrwx",
		"owner UID of \"/test-volume\": 0",
		"owner GID of \"/test-volume\": 123",
		"mode of file \"/test-volume\": dgtrwxrwxrwx",
	}
	if medium == v1.StorageMediumMemory {
		out = append(out, "mount type of \"/test-volume\": tmpfs")
	}
	f.TestContainerOutput(msg, pod, 0, out)
}

func doTestVolumeModeFSGroup(f *framework.Framework, image string, medium v1.StorageMedium) {
	var (
		source = &v1.EmptyDirVolumeSource{Medium: medium}
		pod    = testPodWithVolume(testImageRootUid, volumePath, source)
	)

	pod.Spec.Containers[0].Args = []string{
		fmt.Sprintf("--fs_type=%v", volumePath),
		fmt.Sprintf("--file_perm=%v", volumePath),
	}

	fsGroup := int64(1001)
	pod.Spec.SecurityContext.FSGroup = &fsGroup

	msg := fmt.Sprintf("emptydir volume type on %v", formatMedium(medium))
	out := []string{
		"perms of file \"/test-volume\": -rwxrwxrwx",
	}
	if medium == v1.StorageMediumMemory {
		out = append(out, "mount type of \"/test-volume\": tmpfs")
	}
	f.TestContainerOutput(msg, pod, 0, out)
}

func doTest0644FSGroup(f *framework.Framework, image string, medium v1.StorageMedium) {
	var (
		filePath = path.Join(volumePath, "test-file")
		source   = &v1.EmptyDirVolumeSource{Medium: medium}
		pod      = testPodWithVolume(image, volumePath, source)
	)

	pod.Spec.Containers[0].Args = []string{
		fmt.Sprintf("--fs_type=%v", volumePath),
		fmt.Sprintf("--new_file_0644=%v", filePath),
		fmt.Sprintf("--file_perm=%v", filePath),
	}

	fsGroup := int64(123)
	pod.Spec.SecurityContext.FSGroup = &fsGroup

	msg := fmt.Sprintf("emptydir 0644 on %v", formatMedium(medium))
	out := []string{
		"perms of file \"/test-volume/test-file\": -rw-r--r--",
		"content of file \"/test-volume/test-file\": mount-tester new file",
	}
	if medium == v1.StorageMediumMemory {
		out = append(out, "mount type of \"/test-volume\": tmpfs")
	}
	f.TestContainerOutput(msg, pod, 0, out)
}

func doTestVolumeMode(f *framework.Framework, image string, medium v1.StorageMedium) {
	var (
		source = &v1.EmptyDirVolumeSource{Medium: medium}
		pod    = testPodWithVolume(testImageRootUid, volumePath, source)
	)

	pod.Spec.Containers[0].Args = []string{
		fmt.Sprintf("--fs_type=%v", volumePath),
		fmt.Sprintf("--file_perm=%v", volumePath),
	}

	msg := fmt.Sprintf("emptydir volume type on %v", formatMedium(medium))
	out := []string{
		"perms of file \"/test-volume\": -rwxrwxrwx",
	}
	if medium == v1.StorageMediumMemory {
		out = append(out, "mount type of \"/test-volume\": tmpfs")
	}
	f.TestContainerOutput(msg, pod, 0, out)
}

func doTest0644(f *framework.Framework, image string, medium v1.StorageMedium) {
	var (
		filePath = path.Join(volumePath, "test-file")
		source   = &v1.EmptyDirVolumeSource{Medium: medium}
		pod      = testPodWithVolume(image, volumePath, source)
	)

	pod.Spec.Containers[0].Args = []string{
		fmt.Sprintf("--fs_type=%v", volumePath),
		fmt.Sprintf("--new_file_0644=%v", filePath),
		fmt.Sprintf("--file_perm=%v", filePath),
	}

	msg := fmt.Sprintf("emptydir 0644 on %v", formatMedium(medium))
	out := []string{
		"perms of file \"/test-volume/test-file\": -rw-r--r--",
		"content of file \"/test-volume/test-file\": mount-tester new file",
	}
	if medium == v1.StorageMediumMemory {
		out = append(out, "mount type of \"/test-volume\": tmpfs")
	}
	f.TestContainerOutput(msg, pod, 0, out)
}

func doTest0666(f *framework.Framework, image string, medium v1.StorageMedium) {
	var (
		filePath = path.Join(volumePath, "test-file")
		source   = &v1.EmptyDirVolumeSource{Medium: medium}
		pod      = testPodWithVolume(image, volumePath, source)
	)

	pod.Spec.Containers[0].Args = []string{
		fmt.Sprintf("--fs_type=%v", volumePath),
		fmt.Sprintf("--new_file_0666=%v", filePath),
		fmt.Sprintf("--file_perm=%v", filePath),
	}

	msg := fmt.Sprintf("emptydir 0666 on %v", formatMedium(medium))
	out := []string{
		"perms of file \"/test-volume/test-file\": -rw-rw-rw-",
		"content of file \"/test-volume/test-file\": mount-tester new file",
	}
	if medium == v1.StorageMediumMemory {
		out = append(out, "mount type of \"/test-volume\": tmpfs")
	}
	f.TestContainerOutput(msg, pod, 0, out)
}

func doTest0777(f *framework.Framework, image string, medium v1.StorageMedium) {
	var (
		filePath = path.Join(volumePath, "test-file")
		source   = &v1.EmptyDirVolumeSource{Medium: medium}
		pod      = testPodWithVolume(image, volumePath, source)
	)

	pod.Spec.Containers[0].Args = []string{
		fmt.Sprintf("--fs_type=%v", volumePath),
		fmt.Sprintf("--new_file_0777=%v", filePath),
		fmt.Sprintf("--file_perm=%v", filePath),
	}

	msg := fmt.Sprintf("emptydir 0777 on %v", formatMedium(medium))
	out := []string{
		"perms of file \"/test-volume/test-file\": -rwxrwxrwx",
		"content of file \"/test-volume/test-file\": mount-tester new file",
	}
	if medium == v1.StorageMediumMemory {
		out = append(out, "mount type of \"/test-volume\": tmpfs")
	}
	f.TestContainerOutput(msg, pod, 0, out)
}

func formatMedium(medium v1.StorageMedium) string {
	if medium == v1.StorageMediumMemory {
		return "tmpfs"
	}

	return "node default medium"
}

func testPodWithVolume(image, path string, source *v1.EmptyDirVolumeSource) *v1.Pod {
	podName := "pod-" + string(uuid.NewUUID())
	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  containerName,
					Image: image,
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      volumeName,
							MountPath: path,
						},
					},
				},
			},
			SecurityContext: &v1.PodSecurityContext{
				SELinuxOptions: &v1.SELinuxOptions{
					Level: "s0",
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
			Volumes: []v1.Volume{
				{
					Name: volumeName,
					VolumeSource: v1.VolumeSource{
						EmptyDir: source,
					},
				},
			},
		},
	}
}
