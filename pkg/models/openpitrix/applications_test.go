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
	"k8s.io/client-go/kubernetes"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	fakeks "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/s3/fake"
	"testing"
)

var rawChartData = "H4sIFAAAAAAA/ykAK2FIUjBjSE02THk5NWIzVjBkUzVpWlM5Nk9WVjZNV2xqYW5keVRRbz1IZWxtAOxYUW/bNhDOM3/FTVmBNltoubEdQEAfirTAim1pMA/ZwzAUtHSS2FAkS1JOvLT77QNJ2XGUZEmxJukw34NEkcfj3fG+41EOrRsc1Mw4umCN2LoPStM0nYxG4Z2maf+dDif7W8NROhyP0v3R/t5WOnw+nAy3IL0XbXrUWsfMVvqv1+ob9x8hpvkxGsuVzGD+nDCtV59DOpzQlBRoc8O1C30v4QcUDeQ+YKBUBn5sZ2gkOrREsgYz8AFF3EJjBkxrwXPmZ5L5UmpKhzQlj232hjoK+J8z0aK9twRwG/6fD8d9/I+GG/w/CBkMGD1QrXQZDAnhDaswIwAGtbLcKbPIQFZcnhEA3QpxpATPFxm8KQ+VOzJoUToC4FiVQZJ0Ao5aIaaYG3Q2g9//CLnh7RyN4QUGtrIV4krnYzvjf0gB/w4bLZhDO3hXo9BoLHX6y6WCW/C/t7c/6eF/NN6c/w9D5+eDHZjzJgOLDkou0J/dLxrvlrzGDHYGnz4Rz0Ven2kmC3A1gkcuqDK0Qy1ASce3CwWWXCIkPrKoZ0xg92KItcIBjQXnoZdCj+Phs54M4CM408ocJnuhyZtpW5b8DJLdBDpZKAvfjKodGGQOga1W8OllAR9aJnjJsfClSFCakt8wyg78zq/gDbAww5y1FsGqBteqmmhqyVEUFphBELzhDgtwClzNLTydLYIbXh1OPS+XFViN+TNK3pRgUCCznb9yJR3j0nbVU+jjDk65EDBDaK3X0wILynfaXu/VZfK88CwvV47sZ9alw24cv4uzhV3J+TYonr24+25e6LhyQRRCf4n+iXOXel7q/EzltOHSlZA8sbtPbNKTFRe9e2xd37wUcWtb6bHRVbl+G8N2drERuQSbobhpSwPLxX727Vh3cWx3ZTp89Ae1YDlC8l0Cybvk88GjmkbJqJ69Qb04GPWrUTTU1oOgcgbn58BlLtqiZwqNi/UGLQrMnTI/dQLpWnR0lr1c3UH8GNOanqzgSLkarK4S5+fXTPkIH1rlsGfpVSkNk6zCYne2iIKWkTJFM+d5f3701LRT/p991Tdx99r1423pin8irOn1OnNpHZM5XtZ4HTzXxWg/YdvOQpbnvurzmay1eKMxgfll5D28KelcZqN5XLmX9p9eNvUii9FnNwmS67at4XwpMukayZ0EXMHyY5++j0+9+i9XsuRVw/SXvAze+v9nnPbqv3E63tR/D0InXBYZHIRt/5lp0qBjBXPM3wBXKWoZH1eBG/PU2i+kIVnO9qwZ+C8CsEHaV0oB/9Qf6bySyuB9rHEb/sd7V/7/7E3GG/w/BG3DEXMOjbS+DogxAKc1Spi1XBT+OqNZfsIqtJRsw6/+ymNbrZVxFmyNQkAl1Awa5vKay+p7f+dhjs8RNHP1Wj+TBdkGiVX4IQxPtcGSn2EBp9zV8M0zCm+lWICSYaZXCTQaEFwiJfTV9N3UKYNkG7p69fhgCgU3ltCKu0F4RvUJnf1pBuG57KirgX8sP+1cDi4EzVh+0upw97Vkh9pTTXbojJ2QHeoa31aGV2TnL7INx8xw1Vp48+q1JVQb9R5zRygvkA0iu1HvCZ3bXBU42CS9DW1oQ18z/R0AAP//GfF7tgAeAAA="

func TestOpenPitrixApp(t *testing.T) {
	appOperator := prepareAppOperator()

	chartData, _ := base64.RawStdEncoding.DecodeString(rawChartData)

	validateReq := &ValidatePackageRequest{
		VersionPackage: chartData,
	}
	// validate package
	validateResp, err := appOperator.ValidatePackage(validateReq)
	if err != nil || validateResp.Error != "" {
		klog.Errorf("validate pacakge failed, error: %s", err)
		t.FailNow()
	}

	validateReq = &ValidatePackageRequest{
		VersionPackage: strfmt.Base64(""),
	}

	// validate corrupted package
	validateResp, err = appOperator.ValidatePackage(validateReq)
	if err == nil {
		klog.Errorf("validate pacakge failed, error: %s", err)
		t.FailNow()
	}

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

	// describe app
	app, err := appOperator.DescribeApp(createAppResp.AppID)
	if err != nil {
		klog.Errorf("describe app failed, err: %s", err)
		t.FailNow()
	}
	_ = app

	cond := &params.Conditions{Match: map[string]string{
		WorkspaceLabel: testWorkspace,
	}}
	// list apps
	listApps, err := appOperator.ListApps(cond, "", false, 10, 0)
	if err != nil {
		klog.Errorf("list app failed")
		t.FailNow()
	}
	_ = listApps

	// describe app
	describeAppVersion, err := appOperator.DescribeAppVersion(createAppResp.VersionID)
	if err != nil {
		klog.Errorf("describe app version failed, error: %s", err)
		t.FailNow()
	}
	_ = describeAppVersion

	cond.Match[AppId] = createAppResp.AppID
	// list app version
	_, err = appOperator.ListAppVersions(cond, "", false, 10, 0)
	if err != nil {
		klog.Errorf("list app version failed")
		t.FailNow()
	}

	// get app version file
	getAppVersionFilesRequest := &GetAppVersionFilesRequest{}
	_, err = appOperator.GetAppVersionFiles(createAppResp.VersionID, getAppVersionFilesRequest)

	if err != nil {
		klog.Errorf("get app version files failed")
		t.FailNow()
	}

	//delete app
	err = appOperator.DeleteApp(createAppResp.AppID)

	if err == nil {
		klog.Errorf("we should delete application version first")
		t.FailNow()
	}

	//delete app
	err = appOperator.DeleteAppVersion(createAppResp.VersionID)

	if err != nil {
		klog.Errorf("delete application version failed, err: %s", err)
		t.FailNow()
	}

}

var (
	ksClient            versioned.Interface
	k8sClient           kubernetes.Interface
	fakeInformerFactory informers.InformerFactory
	testWorkspace       = "test-workspace"
)

func prepareAppOperator() ApplicationInterface {
	ksClient = fakeks.NewSimpleClientset()
	k8sClient = fakek8s.NewSimpleClientset()
	fakeInformerFactory = informers.NewInformerFactories(k8sClient, ksClient, nil, nil, nil, nil)

	return newApplicationOperator(cachedReposData, fakeInformerFactory.KubeSphereSharedInformerFactory(), ksClient, fake.NewFakeS3())
}
