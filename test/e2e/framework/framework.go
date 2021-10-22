/*
Copyright 2020 KubeSphere Authors

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

package framework

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo" //nolint:stylecheck
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"kubesphere.io/client-go/client"
	"kubesphere.io/client-go/client/generic"
	"kubesphere.io/client-go/restclient"

	"kubesphere.io/kubesphere/pkg/apis"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/test/e2e/framework/workspace"
)

const (
	// Using the same interval as integration should be fine given the
	// minimal load that the apiserver is likely to be under.
	PollInterval = 50 * time.Millisecond
	// How long to try single API calls (like 'get' or 'list'). Used to prevent
	// transient failures from failing tests.
	DefaultSingleCallTimeout = 30 * time.Second
)

type Framework struct {
	BaseName   string
	Workspace  string
	Namespaces []string
	Scheme     *runtime.Scheme
}

// KubeSphereFramework provides an interface to a test control plane so
// that the implementation can vary without affecting tests.
type KubeSphereFramework interface {
	GenericClient(userAgent string) client.Client
	RestClient(userAgent string) *restclient.RestClient
	KubeSphereSystemNamespace() string

	// Name of the workspace for the current test to target
	TestWorkSpaceName() string

	// Create a Namespace under current Worksapce
	CreateNamespace(name string) string
	// Get Names of the namespaces for the current test to target
	GetNamespaceNames() []string

	GetScheme() *runtime.Scheme
}

func NewKubeSphereFramework(baseName string) KubeSphereFramework {

	sch := runtime.NewScheme()
	if err := apis.AddToScheme(sch); err != nil {
		Failf("unable add KubeSphere APIs to scheme: %v", err)
	}
	if err := scheme.AddToScheme(sch); err != nil {
		Failf("unable add Kubernetes APIs to scheme: %v", err)
	}

	f := &Framework{
		BaseName: baseName,
		Scheme:   sch,
	}

	ginkgo.AfterEach(f.AfterEach)
	ginkgo.BeforeEach(f.BeforeEach)
	return f
}

// BeforeEach
func (f *Framework) BeforeEach() {

}

// AfterEach
func (f *Framework) AfterEach() {
}

func (f *Framework) TestWorkSpaceName() string {
	if f.Workspace == "" {
		f.Workspace = CreateTestWorkSpace(f.GenericClient(f.BaseName), f.BaseName)
	}
	return f.Workspace
}

func CreateTestWorkSpace(client client.Client, baseName string) string {
	ginkgo.By("Creating a WorkSpace to execute the test in")
	wspt := workspace.NewWorkspaceTemplate("", "admin")
	wspt.GenerateName = fmt.Sprintf("e2e-tests-%v-", baseName)
	wspt, err := workspace.CreateWorkspace(client, wspt)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	ginkgo.By(fmt.Sprintf("Created test workspace %s", wspt.Name))
	return wspt.Name
}

func (f *Framework) GetNamespaceNames() []string {
	return f.Namespaces
}

func (f *Framework) CreateNamespace(name string) string {
	name = fmt.Sprintf("%s-%s", f.Workspace, name)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				constants.WorkspaceLabelKey: f.TestWorkSpaceName(),
			},
		},
	}

	opts := &client.URLOptions{
		Group:   "tenant.kubesphere.io",
		Version: "v1alpha2",
	}

	err := f.GenericClient(f.BaseName).Create(context.TODO(), ns, opts, &client.WorkspaceOptions{Name: f.TestWorkSpaceName()})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return ns.Name
}

func (f *Framework) KubeSphereSystemNamespace() string {
	return "Kubesphere-system"
}

func (f *Framework) GenericClient(userAgent string) client.Client {

	ctx := TestContext

	config := &rest.Config{
		Host:     ctx.Host,
		Username: ctx.Username,
		Password: ctx.Password,
		ContentConfig: rest.ContentConfig{
			ContentType: runtime.ContentTypeJSON,
		},
	}

	rest.AddUserAgent(config, userAgent)

	return generic.NewForConfigOrDie(config, client.Options{Scheme: f.Scheme})
}

func (f *Framework) RestClient(userAgent string) *restclient.RestClient {
	ctx := TestContext
	config := &rest.Config{
		Host:     ctx.Host,
		Username: ctx.Username,
		Password: ctx.Password,
	}
	c, err := restclient.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return c
}

func (f *Framework) GetScheme() *runtime.Scheme {
	return f.Scheme
}
