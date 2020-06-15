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
	"fmt"
	"github.com/golang/protobuf/ptypes/wrappers"
	"k8s.io/klog"
	"openpitrix.io/openpitrix/pkg/manager"
	"openpitrix.io/openpitrix/pkg/pb"
	"openpitrix.io/openpitrix/pkg/sender"
	"openpitrix.io/openpitrix/pkg/util/ctxutil"
	"strconv"
	"strings"
)

const (
	RuntimeAnnotationKey = "openpitrix_runtime"
	KubernetesProvider   = "kubernetes"
	Unknown              = "-"
	DeploySuffix         = "-Deployment"
	DaemonSuffix         = "-DaemonSet"
	StateSuffix          = "-StatefulSet"
	SystemUsername       = "system"
	SystemUserPath       = ":system"
)

type Client interface {
	pb.RuntimeManagerClient
	pb.ClusterManagerClient
	pb.AppManagerClient
	pb.RepoManagerClient
	pb.CategoryManagerClient
	pb.AttachmentManagerClient
	pb.RepoIndexerClient
	// upsert the openpitrix runtime when cluster is updated or created
	UpsertRuntime(cluster string, kubeConfig string) error
	// clean up the openpitrix runtime when cluster is deleted
	CleanupRuntime(cluster string) error
	// migrate the openpitrix runtime when upgrade ks2.x to ks3.x
	MigrateRuntime(runtimeId string, cluster string) error
}

type client struct {
	pb.RuntimeManagerClient
	pb.ClusterManagerClient
	pb.AppManagerClient
	pb.RepoManagerClient
	pb.CategoryManagerClient
	pb.AttachmentManagerClient
	pb.RepoIndexerClient
}

func (c *client) UpsertRuntime(cluster string, kubeConfig string) error {
	ctx := SystemContext()
	req := &pb.CreateRuntimeCredentialRequest{
		Name:                     &wrappers.StringValue{Value: fmt.Sprintf("kubeconfig-%s", cluster)},
		Provider:                 &wrappers.StringValue{Value: KubernetesProvider},
		Description:              &wrappers.StringValue{Value: "kubeconfig"},
		RuntimeUrl:               &wrappers.StringValue{Value: "kubesphere"},
		RuntimeCredentialContent: &wrappers.StringValue{Value: kubeConfig},
		RuntimeCredentialId:      &wrappers.StringValue{Value: cluster},
	}
	_, err := c.CreateRuntimeCredential(ctx, req)
	if err != nil {
		return err
	}
	_, err = c.CreateRuntime(ctx, &pb.CreateRuntimeRequest{
		Name:                &wrappers.StringValue{Value: cluster},
		RuntimeCredentialId: &wrappers.StringValue{Value: cluster},
		Provider:            &wrappers.StringValue{Value: KubernetesProvider},
		Zone:                &wrappers.StringValue{Value: cluster},
		RuntimeId:           &wrappers.StringValue{Value: cluster},
	})

	return err
}

func (c *client) CleanupRuntime(cluster string) error {
	ctx := SystemContext()
	_, err := c.DeleteClusterInRuntime(ctx, &pb.DeleteClusterInRuntimeRequest{
		RuntimeId: []string{cluster},
	})
	return err
}

func (c *client) MigrateRuntime(runtimeId string, cluster string) error {
	ctx := SystemContext()
	_, err := c.MigrateClusterInRuntime(ctx, &pb.MigrateClusterInRuntimeRequest{
		FromRuntimeId: runtimeId,
		ToRuntimeId:   cluster,
	})
	return err
}

func parseToHostPort(endpoint string) (string, int, error) {
	args := strings.Split(endpoint, ":")
	if len(args) != 2 {
		return "", 0, fmt.Errorf("invalid server host: %s", endpoint)
	}
	host := args[0]
	port, err := strconv.Atoi(args[1])
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}

func newRuntimeManagerClient(endpoint string) (pb.RuntimeManagerClient, error) {
	if len(endpoint) == 0 {
		return nil, nil
	}

	host, port, err := parseToHostPort(endpoint)
	if err != nil {
		return nil, err
	}

	conn, err := manager.NewClient(host, port)
	if err != nil {
		return nil, err
	}
	return pb.NewRuntimeManagerClient(conn), nil
}
func newClusterManagerClient(endpoint string) (pb.ClusterManagerClient, error) {
	if len(endpoint) == 0 {
		return nil, nil
	}

	host, port, err := parseToHostPort(endpoint)
	if err != nil {
		return nil, err
	}
	conn, err := manager.NewClient(host, port)
	if err != nil {
		return nil, err
	}
	return pb.NewClusterManagerClient(conn), nil
}
func newCategoryManagerClient(endpoint string) (pb.CategoryManagerClient, error) {
	if len(endpoint) == 0 {
		return nil, nil
	}

	host, port, err := parseToHostPort(endpoint)
	if err != nil {
		return nil, err
	}
	conn, err := manager.NewClient(host, port)
	if err != nil {
		return nil, err
	}
	return pb.NewCategoryManagerClient(conn), nil
}

func newAttachmentManagerClient(endpoint string) (pb.AttachmentManagerClient, error) {
	if len(endpoint) == 0 {
		return nil, nil
	}

	host, port, err := parseToHostPort(endpoint)
	if err != nil {
		return nil, err
	}
	conn, err := manager.NewClient(host, port)
	if err != nil {
		return nil, err
	}
	return pb.NewAttachmentManagerClient(conn), nil
}

func newRepoManagerClient(endpoint string) (pb.RepoManagerClient, error) {
	if len(endpoint) == 0 {
		return nil, nil
	}

	host, port, err := parseToHostPort(endpoint)
	if err != nil {
		return nil, err
	}
	conn, err := manager.NewClient(host, port)
	if err != nil {
		return nil, err
	}
	return pb.NewRepoManagerClient(conn), nil
}

func newRepoIndexer(endpoint string) (pb.RepoIndexerClient, error) {
	if len(endpoint) == 0 {
		return nil, nil
	}

	host, port, err := parseToHostPort(endpoint)
	if err != nil {
		return nil, err
	}
	conn, err := manager.NewClient(host, port)
	if err != nil {
		return nil, err
	}
	return pb.NewRepoIndexerClient(conn), nil
}

func newAppManagerClient(endpoint string) (pb.AppManagerClient, error) {
	if len(endpoint) == 0 {
		return nil, nil
	}

	host, port, err := parseToHostPort(endpoint)
	if err != nil {
		return nil, err
	}
	conn, err := manager.NewClient(host, port)
	if err != nil {
		return nil, err
	}
	return pb.NewAppManagerClient(conn), nil
}

// will return a nil client and nil error if endpoint is empty
func NewClient(options *Options) (Client, error) {
	if options.IsEmpty() {
		return nil, nil
	}

	runtimeMangerClient, err := newRuntimeManagerClient(options.RuntimeManagerEndpoint)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	clusterManagerClient, err := newClusterManagerClient(options.ClusterManagerEndpoint)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	repoManagerClient, err := newRepoManagerClient(options.RepoManagerEndpoint)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	repoIndexerClient, err := newRepoIndexer(options.RepoIndexerEndpoint)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	appManagerClient, err := newAppManagerClient(options.AppManagerEndpoint)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	categoryManagerClient, err := newCategoryManagerClient(options.CategoryManagerEndpoint)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	attachmentManagerClient, err := newAttachmentManagerClient(options.AttachmentManagerEndpoint)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	client := client{
		RuntimeManagerClient:    runtimeMangerClient,
		ClusterManagerClient:    clusterManagerClient,
		RepoManagerClient:       repoManagerClient,
		AppManagerClient:        appManagerClient,
		CategoryManagerClient:   categoryManagerClient,
		AttachmentManagerClient: attachmentManagerClient,
		RepoIndexerClient:       repoIndexerClient,
	}

	return &client, nil
}

func SystemContext() context.Context {
	ctx := context.Background()
	ctx = ctxutil.ContextWithSender(ctx, sender.New(SystemUsername, SystemUserPath, ""))
	return ctx
}
func ContextWithUsername(username string) context.Context {
	ctx := context.Background()
	if username == "" {
		username = SystemUsername
	}
	ctx = ctxutil.ContextWithSender(ctx, sender.New(username, SystemUserPath, ""))
	return ctx
}

func IsNotFound(err error) bool {
	if strings.Contains(err.Error(), "not exist") {
		return true
	}
	if strings.Contains(err.Error(), "not found") {
		return true
	}
	return false
}

func IsDeleted(err error) bool {
	if strings.Contains(err.Error(), "is [deleted]") {
		return true
	}
	return false
}
