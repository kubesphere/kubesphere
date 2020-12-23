/*
Copyright 2020 The KubeSphere Authors.

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

package openpitrix

import (
	"context"
	"encoding/base64"
	"github.com/go-openapi/strfmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/server/params"
	"testing"
)

func TestOpenPitrixRelease(t *testing.T) {
	appOperator := prepareAppOperator()

	chartData, _ := base64.RawStdEncoding.DecodeString(rawChartData)

	appReq := &CreateAppRequest{
		Isv:            testWorkspace,
		Name:           "test-chart",
		VersionName:    "0.1.0",
		VersionPackage: strfmt.Base64(chartData),
	}

	// create app
	createAppResp, err := appOperator.CreateApp(appReq)
	if err != nil {
		klog.Errorf("create app failed")
		t.Fail()
	}

	// add app to indexer
	apps, err := ksClient.ApplicationV1alpha1().HelmApplications().List(context.TODO(), metav1.ListOptions{})
	for _, app := range apps.Items {
		err := fakeInformerFactory.KubeSphereSharedInformerFactory().Application().V1alpha1().HelmApplications().
			Informer().GetIndexer().Add(&app)
		if err != nil {
			klog.Errorf("failed to add app to indexer")
			t.FailNow()
		}
	}

	// add app version to indexer
	appvers, err := ksClient.ApplicationV1alpha1().HelmApplicationVersions().List(context.TODO(), metav1.ListOptions{})
	for _, ver := range appvers.Items {
		err := fakeInformerFactory.KubeSphereSharedInformerFactory().Application().V1alpha1().HelmApplicationVersions().
			Informer().GetIndexer().Add(&ver)
		if err != nil {
			klog.Errorf("failed to add app version to indexer")
			t.Fail()
		}
	}

	rlsOperator := newReleaseOperator(cachedReposData, fakeInformerFactory.KubernetesSharedInformerFactory(), fakeInformerFactory.KubeSphereSharedInformerFactory(), ksClient)

	req := CreateClusterRequest{
		Name:      "test-rls",
		AppId:     createAppResp.AppID,
		VersionId: createAppResp.VersionID,
		Workspace: testWorkspace,
	}
	err = rlsOperator.CreateApplication(testWorkspace, "", "default", req)

	if err != nil {
		klog.Errorf("create release failed, error: %s", err)
		t.FailNow()
	}

	// add app version to indexer
	rls, err := ksClient.ApplicationV1alpha1().HelmReleases().List(context.TODO(), metav1.ListOptions{})
	for _, item := range rls.Items {
		err := fakeInformerFactory.KubeSphereSharedInformerFactory().Application().V1alpha1().HelmReleases().
			Informer().GetIndexer().Add(&item)
		if err != nil {
			klog.Errorf("failed to add release to indexer")
			t.FailNow()
		}
	}

	cond := &params.Conditions{Match: map[string]string{
		WorkspaceLabel: testWorkspace,
	}}
	rlsList, err := rlsOperator.ListApplications(testWorkspace, "", "default", cond, 10, 0, "", false)

	if err != nil {
		klog.Errorf("failed to list release, error: %s", err)
		t.FailNow()
	}

	var rlsId string
	for _, item := range rlsList.Items {
		app := item.(*Application)
		rlsId = app.Cluster.ClusterId
		break
	}

	//describe release
	describeRls, err := rlsOperator.DescribeApplication(testWorkspace, "", "default", rlsId)
	if err != nil {
		klog.Errorf("failed to describe release, error: %s", err)
		t.FailNow()
	}
	_ = describeRls

	//delete release
	err = rlsOperator.DeleteApplication(testWorkspace, "", "default", rlsId)
	if err != nil {
		klog.Errorf("failed to delete release, error: %s", err)
		t.FailNow()
	}
}
