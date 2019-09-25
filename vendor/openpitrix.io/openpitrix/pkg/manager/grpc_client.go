// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package manager

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

var ClientOptions = []grpc.DialOption{
	grpc.WithInsecure(),
	grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                30 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	}),
}

var clientCache sync.Map

func NewClient(host string, port int) (*grpc.ClientConn, error) {
	endpoint := fmt.Sprintf("%s:%d", host, port)
	if conn, ok := clientCache.Load(endpoint); ok {
		return conn.(*grpc.ClientConn), nil
	}
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, endpoint, ClientOptions...)
	if err != nil {
		return nil, err
	}
	clientCache.Store(endpoint, conn)
	return conn, nil
}

func NewTLSClient(host string, port int, tlsConfig *tls.Config) (*grpc.ClientConn, error) {
	endpoint := fmt.Sprintf("%s:%d", host, port)
	if conn, ok := clientCache.Load(endpoint); ok {
		return conn.(*grpc.ClientConn), nil
	}
	creds := credentials.NewTLS(tlsConfig)
	tlsClientOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	}
	conn, err := grpc.Dial(endpoint, tlsClientOptions...)
	if err != nil {
		return nil, err
	}
	clientCache.Store(endpoint, conn)
	return conn, nil
}
