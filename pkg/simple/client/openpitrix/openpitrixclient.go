/*

 Copyright 2019 The KubeSphere Authors.

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
	"k8s.io/klog"
	"openpitrix.io/openpitrix/pkg/manager"
	"openpitrix.io/openpitrix/pkg/pb"
	"openpitrix.io/openpitrix/pkg/sender"
	"openpitrix.io/openpitrix/pkg/util/ctxutil"
	"strconv"
	"strings"
)

const (
	KubernetesProvider = "kubernetes"
	Unknown            = "-"
	DeploySuffix       = "-Deployment"
	DaemonSuffix       = "-DaemonSet"
	StateSuffix        = "-StatefulSet"
	SystemUsername     = "system"
	SystemUserPath     = ":system"
)

type OpenPitrixClient struct {
	runtime     pb.RuntimeManagerClient
	cluster     pb.ClusterManagerClient
	app         pb.AppManagerClient
	repo        pb.RepoManagerClient
	category    pb.CategoryManagerClient
	attachment  pb.AttachmentManagerClient
	repoIndexer pb.RepoIndexerClient
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
	host, port, err := parseToHostPort(endpoint)
	conn, err := manager.NewClient(host, port)
	if err != nil {
		return nil, err
	}
	return pb.NewRuntimeManagerClient(conn), nil
}
func newClusterManagerClient(endpoint string) (pb.ClusterManagerClient, error) {
	host, port, err := parseToHostPort(endpoint)
	conn, err := manager.NewClient(host, port)
	if err != nil {
		return nil, err
	}
	return pb.NewClusterManagerClient(conn), nil
}
func newCategoryManagerClient(endpoint string) (pb.CategoryManagerClient, error) {
	host, port, err := parseToHostPort(endpoint)
	conn, err := manager.NewClient(host, port)
	if err != nil {
		return nil, err
	}
	return pb.NewCategoryManagerClient(conn), nil
}

func newAttachmentManagerClient(endpoint string) (pb.AttachmentManagerClient, error) {
	host, port, err := parseToHostPort(endpoint)
	conn, err := manager.NewClient(host, port)
	if err != nil {
		return nil, err
	}
	return pb.NewAttachmentManagerClient(conn), nil
}

func newRepoManagerClient(endpoint string) (pb.RepoManagerClient, error) {
	host, port, err := parseToHostPort(endpoint)
	conn, err := manager.NewClient(host, port)
	if err != nil {
		return nil, err
	}
	return pb.NewRepoManagerClient(conn), nil
}

func newRepoIndexer(endpoint string) (pb.RepoIndexerClient, error) {
	host, port, err := parseToHostPort(endpoint)
	conn, err := manager.NewClient(host, port)
	if err != nil {
		return nil, err
	}
	return pb.NewRepoIndexerClient(conn), nil
}

func newAppManagerClient(endpoint string) (pb.AppManagerClient, error) {
	host, port, err := parseToHostPort(endpoint)
	conn, err := manager.NewClient(host, port)
	if err != nil {
		return nil, err
	}
	return pb.NewAppManagerClient(conn), nil
}

func NewOpenPitrixClient(options *OpenPitrixOptions) (*OpenPitrixClient, error) {

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

	client := OpenPitrixClient{
		runtime:     runtimeMangerClient,
		cluster:     clusterManagerClient,
		repo:        repoManagerClient,
		app:         appManagerClient,
		category:    categoryManagerClient,
		attachment:  attachmentManagerClient,
		repoIndexer: repoIndexerClient,
	}

	return &client, nil
}
func (c *OpenPitrixClient) Runtime() pb.RuntimeManagerClient {
	return c.runtime
}
func (c *OpenPitrixClient) App() pb.AppManagerClient {
	return c.app
}
func (c *OpenPitrixClient) Cluster() pb.ClusterManagerClient {
	return c.cluster
}
func (c *OpenPitrixClient) Category() pb.CategoryManagerClient {
	return c.category
}

func (c *OpenPitrixClient) Repo() pb.RepoManagerClient {
	return c.repo
}

func (c *OpenPitrixClient) RepoIndexer() pb.RepoIndexerClient {
	return c.repoIndexer
}

func (c *OpenPitrixClient) Attachment() pb.AttachmentManagerClient {
	return c.attachment
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
