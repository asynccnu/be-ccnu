//go:build wireinject

package main

import (
	"github.com/asynccnu/be-ccnu/grpc"
	"github.com/asynccnu/be-ccnu/ioc"
	"github.com/asynccnu/be-ccnu/pkg/grpcx"
	"github.com/asynccnu/be-ccnu/service"
	"github.com/google/wire"
)

func InitGRPCServer() grpcx.Server {
	wire.Build(
		ioc.InitGRPCxKratosServer,
		grpc.NewCCNUServiceServer,
		service.NewCCNUService,
		ioc.InitLogger,
		ioc.InitEtcdClient,
	)
	return grpcx.Server(nil)
}
