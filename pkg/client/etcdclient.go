/*
Copyright 2018 The KubeSphere Authors.

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

package client

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"

	"context"
	"fmt"
	"time"

	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/options"
)

type EtcdClient struct {
	cli *clientv3.Client
}

func (cli EtcdClient) Put(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	_, err := cli.cli.Put(ctx, key, value)
	cancel()
	if err != nil {
		switch err {
		case context.Canceled:
			glog.Errorf("ctx is canceled by another routine: %v\n", err)
		case context.DeadlineExceeded:
			glog.Errorf("ctx is attached with a deadline is exceeded: %v\n", err)
		case rpctypes.ErrEmptyKey:
			glog.Errorf("client-side error: %v\n", err)
		default:
			glog.Errorf("bad cluster endpoints, which are not etcd servers: %v\n", err)
		}
		return err
	}

	return nil
}

func (cli EtcdClient) Get(key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	resp, err := cli.cli.Get(ctx, key)
	cancel()
	if err != nil {
		glog.Error(err)
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("empty value of key: %s", key)
	}
	return resp.Kvs[0].Value, nil
}

func (cli EtcdClient) Close() {

	cli.cli.Close()
}

func newEtcdClientWithHttps(certFile, keyFile, caFile string, endpoints []string, dialTimeout int) (*clientv3.Client, error) {
	tlsInfo := transport.TLSInfo{
		CertFile:      certFile,
		KeyFile:       keyFile,
		TrustedCAFile: caFile,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		glog.Errorln(err)
		return nil, err
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Duration(dialTimeout) * time.Second,
		TLS:         tlsConfig,
	})
	if err != nil {
		glog.Errorln(err)
		return nil, err
	}
	return cli, nil // make sure to close the client
}

func newEtcdClient(endpoints []string, dialTimeout int) (*clientv3.Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Duration(dialTimeout) * time.Second,
	})
	if err != nil {
		glog.Errorln(err)
		return nil, err
	}
	return cli, nil // make sure to close the client
}

func NewEtcdClient() (*EtcdClient, error) {
	var cli *clientv3.Client
	var err error
	cert := options.ServerOptions.GetEtcdCertFile()
	key := options.ServerOptions.GetEtcdKeyFile()
	ca := options.ServerOptions.GetEtcdCaFile()
	endpoints := options.ServerOptions.GetEtcdEndPoints()
	if len(cert) > 0 && len(key) > 0 && len(ca) > 0 {
		cli, err = newEtcdClientWithHttps(cert, key, ca, endpoints, 20)
	} else {
		cli, err = newEtcdClient(endpoints, 20)
	}
	if err != nil {
		return nil, err
	}
	return &EtcdClient{cli}, nil
}
