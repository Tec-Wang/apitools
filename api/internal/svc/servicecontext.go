package svc

import (
	"apitools/api/internal/config"
	"apitools/rpc/pb"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config   config.Config
	EmailRpc pb.EmailServiceClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:   c,
		EmailRpc: pb.NewEmailServiceClient(zrpc.MustNewClient(c.EmailRpc).Conn()),
	}
}
