// Copyright 2018 The KubeSphere Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file

package alert

import (
	"kubesphere.io/kubesphere/pkg/simple/client/alert/pb"
)

const (
	ManagerHost = "alerting-manager-server.kubesphere-alerting-system.svc"
	ManagerPort = 9201
)

type Client struct {
	pb.AlertManagerClient
}

func NewClient() (*Client, error) {
	conn, err := NewGRPCClient(ManagerHost, ManagerPort)
	if err != nil {
		return nil, err
	}
	return &Client{
		AlertManagerClient: pb.NewAlertManagerClient(conn),
	}, nil
}

type CustomClient struct {
	pb.AlertManagerCustomClient
}

func NewCustomClient() (*CustomClient, error) {
	conn, err := NewGRPCClient(ManagerHost, ManagerPort)
	if err != nil {
		return nil, err
	}
	return &CustomClient{
		AlertManagerCustomClient: pb.NewAlertManagerCustomClient(conn),
	}, nil
}
