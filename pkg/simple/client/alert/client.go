// Copyright 2018 The KubeSphere Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file

package alert

import (
	"kubesphere.io/kubesphere/pkg/simple/client/alert/pb"
)

type Client struct {
	pb.AlertManagerClient
}

func NewClient() (*Client, error) {
	//cfg := config.GetInstance()
	managerHost := "alerting-manager-server.kubesphere-alerting-system.svc" //cfg.App.Host
	managerPort := 9201 //strconv.Atoi(cfg.App.Port)

	conn, err := NewGRPCClient(managerHost, managerPort)
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
	//cfg := config.GetInstance()
	managerHost := "alerting-manager-server.kubesphere-alerting-system.svc" //cfg.App.Host
	managerPort := 9201 //strconv.Atoi(cfg.App.Port)

	conn, err := NewGRPCClient(managerHost, managerPort)
	if err != nil {
		return nil, err
	}
	return &CustomClient{
		AlertManagerCustomClient: pb.NewAlertManagerCustomClient(conn),
	}, nil
}
