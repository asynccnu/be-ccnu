//go:build wireinject

package main

import (
	"github.com/MuxiKeStack/be-ccnu/grpc"
	"github.com/MuxiKeStack/be-ccnu/ioc"
	"github.com/MuxiKeStack/be-ccnu/pkg/grpcx"
	"github.com/MuxiKeStack/be-ccnu/service"
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
