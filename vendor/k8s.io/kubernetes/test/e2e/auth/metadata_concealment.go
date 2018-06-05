/*
Copyright 2017 The Kubernetes Authors.

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

package auth

import (
	batch "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/test/e2e/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	imageutil "k8s.io/kubernetes/test/utils/image"
)

var _ = SIGDescribe("Metadata Concealment", func() {
	f := framework.NewDefaultFramework("metadata-concealment")

	It("should run a check-metadata-concealment job to completion", func() {
		framework.SkipUnlessProviderIs("gce")
		By("Creating a job")
		job := &batch.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name: "check-metadata-concealment",
			},
			Spec: batch.JobSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name: "check-metadata-concealment",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "check-metadata-concealment",
								Image: imageutil.GetE2EImage(imageutil.CheckMetadataConcealment),
							},
						},
						RestartPolicy: v1.RestartPolicyOnFailure,
					},
				},
			},
		}
		job, err := framework.CreateJob(f.ClientSet, f.Namespace.Name, job)
		Expect(err).NotTo(HaveOccurred())

		By("Ensuring job reaches completions")
		err = framework.WaitForJobFinish(f.ClientSet, f.Namespace.Name, job.Name, int32(1))
		Expect(err).NotTo(HaveOccurred())
	})
})
