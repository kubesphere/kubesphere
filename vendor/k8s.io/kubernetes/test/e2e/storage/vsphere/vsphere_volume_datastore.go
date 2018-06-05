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

package vsphere

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/storage/utils"
)

const (
	InvalidDatastore = "invalidDatastore"
	DatastoreSCName  = "datastoresc"
)

/*
	Test to verify datastore specified in storage-class is being honored while volume creation.

	Steps
	1. Create StorageClass with invalid datastore.
	2. Create PVC which uses the StorageClass created in step 1.
	3. Expect the PVC to fail.
	4. Verify the error returned on PVC failure is the correct.
*/

var _ = utils.SIGDescribe("Volume Provisioning on Datastore [Feature:vsphere]", func() {
	f := framework.NewDefaultFramework("volume-datastore")
	var (
		client       clientset.Interface
		namespace    string
		scParameters map[string]string
	)
	BeforeEach(func() {
		framework.SkipUnlessProviderIs("vsphere")
		Bootstrap(f)
		client = f.ClientSet
		namespace = f.Namespace.Name
		scParameters = make(map[string]string)
		nodeList := framework.GetReadySchedulableNodesOrDie(f.ClientSet)
		if !(len(nodeList.Items) > 0) {
			framework.Failf("Unable to find ready and schedulable Node")
		}
	})

	It("verify dynamically provisioned pv using storageclass fails on an invalid datastore", func() {
		By("Invoking Test for invalid datastore")
		scParameters[Datastore] = InvalidDatastore
		scParameters[DiskFormat] = ThinDisk
		err := invokeInvalidDatastoreTestNeg(client, namespace, scParameters)
		Expect(err).To(HaveOccurred())
		errorMsg := `Failed to provision volume with StorageClass \"` + DatastoreSCName + `\": The specified datastore ` + InvalidDatastore + ` is not a shared datastore across node VMs`
		if !strings.Contains(err.Error(), errorMsg) {
			Expect(err).NotTo(HaveOccurred(), errorMsg)
		}
	})
})

func invokeInvalidDatastoreTestNeg(client clientset.Interface, namespace string, scParameters map[string]string) error {
	By("Creating Storage Class With Invalid Datastore")
	storageclass, err := client.StorageV1().StorageClasses().Create(getVSphereStorageClassSpec(DatastoreSCName, scParameters))
	Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Failed to create storage class with err: %v", err))
	defer client.StorageV1().StorageClasses().Delete(storageclass.Name, nil)

	By("Creating PVC using the Storage Class")
	pvclaim, err := framework.CreatePVC(client, namespace, getVSphereClaimSpecWithStorageClass(namespace, "2Gi", storageclass))
	Expect(err).NotTo(HaveOccurred())
	defer framework.DeletePersistentVolumeClaim(client, pvclaim.Name, namespace)

	By("Expect claim to fail provisioning volume")
	err = framework.WaitForPersistentVolumeClaimPhase(v1.ClaimBound, client, pvclaim.Namespace, pvclaim.Name, framework.Poll, 2*time.Minute)
	Expect(err).To(HaveOccurred())

	eventList, err := client.CoreV1().Events(pvclaim.Namespace).List(metav1.ListOptions{})
	return fmt.Errorf("Failure message: %+q", eventList.Items[0].Message)
}
