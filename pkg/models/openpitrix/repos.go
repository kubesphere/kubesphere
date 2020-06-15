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
	"fmt"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"openpitrix.io/openpitrix/pkg/pb"
	"strings"
)

type RepoInterface interface {
	CreateRepo(request *CreateRepoRequest) (*CreateRepoResponse, error)
	DeleteRepo(id string) error
	ModifyRepo(id string, request *ModifyRepoRequest) error
	DescribeRepo(id string) (*Repo, error)
	ListRepos(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
	ValidateRepo(request *ValidateRepoRequest) (*ValidateRepoResponse, error)
	DoRepoAction(repoId string, request *RepoActionRequest) error
	ListEvents(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error)
	ListRepoEvents(repoId string, conditions *params.Conditions, limit, offset int) (*models.PageableResponse, error)
}

type repoOperator struct {
	opClient openpitrix.Client
}

func newRepoOperator(opClient openpitrix.Client) RepoInterface {
	return &repoOperator{
		opClient: opClient,
	}
}

func (c *repoOperator) CreateRepo(request *CreateRepoRequest) (*CreateRepoResponse, error) {
	createRepoRequest := &pb.CreateRepoRequest{
		Name:             &wrappers.StringValue{Value: request.Name},
		Description:      &wrappers.StringValue{Value: request.Description},
		Type:             &wrappers.StringValue{Value: request.Type},
		Url:              &wrappers.StringValue{Value: request.URL},
		Credential:       &wrappers.StringValue{Value: request.Credential},
		Visibility:       &wrappers.StringValue{Value: request.Visibility},
		CategoryId:       &wrappers.StringValue{Value: request.CategoryId},
		AppDefaultStatus: &wrappers.StringValue{Value: request.AppDefaultStatus},
	}

	if request.Providers != nil {
		createRepoRequest.Providers = request.Providers
	}
	if request.Workspace != nil {
		createRepoRequest.Labels = &wrappers.StringValue{Value: fmt.Sprintf("workspace=%s", *request.Workspace)}
	}

	resp, err := c.opClient.CreateRepo(openpitrix.SystemContext(), createRepoRequest)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return &CreateRepoResponse{
		RepoID: resp.GetRepoId().GetValue(),
	}, nil
}

func (c *repoOperator) DeleteRepo(id string) error {
	_, err := c.opClient.DeleteRepos(openpitrix.SystemContext(), &pb.DeleteReposRequest{
		RepoId: []string{id},
	})
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (c *repoOperator) ModifyRepo(id string, request *ModifyRepoRequest) error {
	modifyRepoRequest := &pb.ModifyRepoRequest{
		RepoId: &wrappers.StringValue{Value: id},
	}

	if request.Name != nil {
		modifyRepoRequest.Name = &wrappers.StringValue{Value: *request.Name}
	}
	if request.Description != nil {
		modifyRepoRequest.Description = &wrappers.StringValue{Value: *request.Description}
	}
	if request.Type != nil {
		modifyRepoRequest.Type = &wrappers.StringValue{Value: *request.Type}
	}
	if request.URL != nil {
		modifyRepoRequest.Url = &wrappers.StringValue{Value: *request.URL}
	}
	if request.Credential != nil {
		modifyRepoRequest.Credential = &wrappers.StringValue{Value: *request.Credential}
	}
	if request.Visibility != nil {
		modifyRepoRequest.Visibility = &wrappers.StringValue{Value: *request.Visibility}
	}

	if request.CategoryID != nil {
		modifyRepoRequest.CategoryId = &wrappers.StringValue{Value: *request.CategoryID}
	}
	if request.AppDefaultStatus != nil {
		modifyRepoRequest.AppDefaultStatus = &wrappers.StringValue{Value: *request.AppDefaultStatus}
	}
	if request.Providers != nil {
		modifyRepoRequest.Providers = request.Providers
	}
	if request.Workspace != nil {
		modifyRepoRequest.Labels = &wrappers.StringValue{Value: fmt.Sprintf("workspace=%s", *request.Workspace)}
	}

	_, err := c.opClient.ModifyRepo(openpitrix.SystemContext(), modifyRepoRequest)
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (c *repoOperator) DescribeRepo(id string) (*Repo, error) {
	resp, err := c.opClient.DescribeRepos(openpitrix.SystemContext(), &pb.DescribeReposRequest{
		RepoId: []string{id},
		Limit:  1,
	})
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	var repo *Repo

	if len(resp.RepoSet) > 0 {
		repo = convertRepo(resp.RepoSet[0])
		return repo, nil
	} else {
		err := status.New(codes.NotFound, "resource not found").Err()
		klog.Error(err)
		return nil, err
	}
}

func (c *repoOperator) ListRepos(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	req := &pb.DescribeReposRequest{}

	if keyword := conditions.Match[Keyword]; keyword != "" {
		req.SearchWord = &wrappers.StringValue{Value: keyword}
	}
	if status := conditions.Match[Status]; status != "" {
		req.Status = strings.Split(status, "|")
	}
	if typeStr := conditions.Match[Type]; typeStr != "" {
		req.Type = strings.Split(typeStr, "|")
	}
	if visibility := conditions.Match[Visibility]; visibility != "" {
		req.Visibility = strings.Split(visibility, "|")
	}
	if status := conditions.Match[Status]; status != "" {
		req.Status = strings.Split(status, "|")
	}
	if workspace := conditions.Match[WorkspaceLabel]; workspace != "" {
		req.Label = &wrappers.StringValue{Value: fmt.Sprintf("workspace=%s", workspace)}
	}
	if orderBy != "" {
		req.SortKey = &wrappers.StringValue{Value: orderBy}
	}
	req.Reverse = &wrappers.BoolValue{Value: reverse}
	req.Limit = uint32(limit)
	req.Offset = uint32(offset)
	resp, err := c.opClient.DescribeRepos(openpitrix.SystemContext(), req)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	items := make([]interface{}, 0)

	for _, item := range resp.RepoSet {
		items = append(items, convertRepo(item))
	}

	return &models.PageableResponse{Items: items, TotalCount: int(resp.TotalCount)}, nil
}

func (c *repoOperator) ValidateRepo(request *ValidateRepoRequest) (*ValidateRepoResponse, error) {
	resp, err := c.opClient.ValidateRepo(openpitrix.SystemContext(), &pb.ValidateRepoRequest{
		Type:       &wrappers.StringValue{Value: request.Type},
		Credential: &wrappers.StringValue{Value: request.Credential},
		Url:        &wrappers.StringValue{Value: request.Url},
	})

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &ValidateRepoResponse{
		ErrorCode: int64(resp.ErrorCode),
		Ok:        resp.Ok.Value,
	}, nil
}

func (c *repoOperator) DoRepoAction(repoId string, request *RepoActionRequest) error {
	var err error
	switch request.Action {
	case ActionIndex:
		indexRepoRequest := &pb.IndexRepoRequest{
			RepoId: &wrappers.StringValue{Value: repoId},
		}
		_, err := c.opClient.IndexRepo(openpitrix.SystemContext(), indexRepoRequest)

		if err != nil {
			klog.Error(err)
			return err
		}

		return nil
	default:
		err = status.New(codes.InvalidArgument, "action not support").Err()
		klog.Error(err)
		return err
	}
}

func (c *repoOperator) ListRepoEvents(repoId string, conditions *params.Conditions, limit, offset int) (*models.PageableResponse, error) {
	describeRepoEventsRequest := &pb.DescribeRepoEventsRequest{
		RepoId: []string{repoId},
	}
	if eventId := conditions.Match["repo_event_id"]; eventId != "" {
		describeRepoEventsRequest.RepoEventId = strings.Split(eventId, "|")
	}
	if status := conditions.Match["status"]; status != "" {
		describeRepoEventsRequest.Status = strings.Split(status, "|")
	}
	describeRepoEventsRequest.Limit = uint32(limit)
	describeRepoEventsRequest.Offset = uint32(offset)

	resp, err := c.opClient.DescribeRepoEvents(openpitrix.SystemContext(), describeRepoEventsRequest)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	items := make([]interface{}, 0)

	for _, item := range resp.RepoEventSet {
		items = append(items, convertRepoEvent(item))
	}

	return &models.PageableResponse{Items: items, TotalCount: int(resp.TotalCount)}, nil
}

func (c *repoOperator) ListEvents(conditions *params.Conditions, orderBy string, reverse bool, limit, offset int) (*models.PageableResponse, error) {
	describeRepoEventsRequest := &pb.DescribeRepoEventsRequest{}
	if repoId := conditions.Match["repo_id"]; repoId != "" {
		describeRepoEventsRequest.RepoId = strings.Split(repoId, "|")
	}
	if eventId := conditions.Match["repo_event_id"]; eventId != "" {
		describeRepoEventsRequest.RepoEventId = strings.Split(eventId, "|")
	}
	if status := conditions.Match["status"]; status != "" {
		describeRepoEventsRequest.Status = strings.Split(status, "|")
	}
	describeRepoEventsRequest.Limit = uint32(limit)
	describeRepoEventsRequest.Offset = uint32(offset)

	resp, err := c.opClient.DescribeRepoEvents(openpitrix.SystemContext(), describeRepoEventsRequest)

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	items := make([]interface{}, 0)

	for _, item := range resp.RepoEventSet {
		items = append(items, convertRepoEvent(item))
	}

	return &models.PageableResponse{Items: items, TotalCount: int(resp.TotalCount)}, nil
}
