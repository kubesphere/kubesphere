// Copyright 2018 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package manager

import (
	"context"
	"fmt"
	"net"
	"runtime/debug"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"openpitrix.io/openpitrix/pkg/config"
	"openpitrix.io/openpitrix/pkg/db"
	"openpitrix.io/openpitrix/pkg/gerr"
	"openpitrix.io/openpitrix/pkg/logger"
	"openpitrix.io/openpitrix/pkg/util/ctxutil"
	"openpitrix.io/openpitrix/pkg/version"
)

type checkerT func(ctx context.Context, req interface{}) error
type builderT func(ctx context.Context, req interface{}) interface{}

var (
	defaultChecker checkerT
	defaultBuilder builderT
)

type GrpcServer struct {
	ServiceName    string
	Port           int
	showErrorCause bool
	checker        checkerT
	builder        builderT
	mysqlConfig    config.MysqlConfig
}

type RegisterCallback func(*grpc.Server)

func NewGrpcServer(serviceName string, port int) *GrpcServer {
	return &GrpcServer{
		ServiceName:    serviceName,
		Port:           port,
		showErrorCause: false,
		checker:        defaultChecker,
		builder:        defaultBuilder,
	}
}

func (g *GrpcServer) ShowErrorCause(b bool) *GrpcServer {
	g.showErrorCause = b
	return g
}

func (g *GrpcServer) WithChecker(c checkerT) *GrpcServer {
	g.checker = c
	return g
}

func (g *GrpcServer) WithBuilder(b builderT) *GrpcServer {
	g.builder = b
	return g
}

func (g *GrpcServer) WithMysqlConfig(cfg config.MysqlConfig) *GrpcServer {
	g.mysqlConfig = cfg
	return g
}

func (g *GrpcServer) Serve(callback RegisterCallback, opt ...grpc.ServerOption) {
	version.PrintVersionInfo(func(s string, i ...interface{}) {
		logger.Info(nil, s, i)
	})
	logger.Info(nil, "Service [%s] start listen at port [%d]", g.ServiceName, g.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", g.Port))
	if err != nil {
		err = errors.WithStack(err)
		logger.Critical(nil, "failed to listen: %+v", err)
	}

	builtinOptions := []grpc.ServerOption{
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc_middleware.WithUnaryServerChain(
			grpc_validator.UnaryServerInterceptor(),
			g.unaryServerLogInterceptor(),
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				ctx = db.NewContext(ctx, g.mysqlConfig)

				if g.checker != nil {
					err = g.checker(ctx, req)
					if err != nil {
						return
					}
				}

				return handler(ctx, req)
			},
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				if g.builder != nil {
					req = g.builder(ctx, req)
				}
				return handler(ctx, req)
			},
			grpc_recovery.UnaryServerInterceptor(
				grpc_recovery.WithRecoveryHandler(func(p interface{}) error {
					logger.Critical(nil, "GRPC server recovery with error: %+v", p)
					logger.Critical(nil, string(debug.Stack()))
					if e, ok := p.(error); ok {
						return gerr.NewWithDetail(nil, gerr.Internal, e, gerr.ErrorInternalError)
					}
					return gerr.New(nil, gerr.Internal, gerr.ErrorInternalError)
				}),
			),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_recovery.StreamServerInterceptor(
				grpc_recovery.WithRecoveryHandler(func(p interface{}) error {
					logger.Critical(nil, "GRPC server recovery with error: %+v", p)
					logger.Critical(nil, string(debug.Stack()))
					if e, ok := p.(error); ok {
						return gerr.NewWithDetail(nil, gerr.Internal, e, gerr.ErrorInternalError)
					}
					return gerr.New(nil, gerr.Internal, gerr.ErrorInternalError)
				}),
			),
		),
	}

	grpcServer := grpc.NewServer(append(opt, builtinOptions...)...)
	reflection.Register(grpcServer)
	callback(grpcServer)

	if err = grpcServer.Serve(lis); err != nil {
		err = errors.WithStack(err)
		logger.Critical(nil, "%+v", err)
	}
}

var (
	jsonPbMarshaller = &jsonpb.Marshaler{
		OrigName: true,
	}
)

func (g *GrpcServer) unaryServerLogInterceptor() grpc.UnaryServerInterceptor {
	showErrorCause := g.showErrorCause

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var err error
		s := ctxutil.GetSender(ctx)
		requestId := ctxutil.GetRequestId(ctx)
		ctx = ctxutil.SetRequestId(ctx, requestId)
		ctx = ctxutil.ContextWithSender(ctx, s)
		locale := ctxutil.GetLocale(ctx)
		ctx = ctxutil.SetLocale(ctx, locale)

		method := strings.Split(info.FullMethod, "/")
		action := method[len(method)-1]
		if p, ok := req.(proto.Message); ok {
			if content, err := jsonPbMarshaller.MarshalToString(p); err != nil {
				logger.Error(ctx, "Failed to marshal proto message to string [%s] [%+v] [%+v]", action, s, err)
			} else {
				logger.Info(ctx, "Request received [%s] [%+v] [%s]", action, s, content)
			}
		}
		start := time.Now()

		resp, err := handler(ctx, req)

		elapsed := time.Since(start)
		logger.Info(ctx, "Handled request [%s] [%+v] exec_time is [%s]", action, s, elapsed)
		if e, ok := status.FromError(err); ok {
			if e.Code() != codes.OK {
				logger.Debug(ctx, "Response is error: %s, %s", e.Code().String(), e.Message())
				if !showErrorCause {
					err = gerr.ClearErrorCause(err)
				}
			}
		}
		return resp, err
	}
}
