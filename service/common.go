package service

import (
	"context"

	commonPb "github.com/pingcap/tcp/proto/common"
)

const TCP_COMMON_SERVICE_NAME = "go.micro.tcp.common"

type Common struct{}

func (c *Common) Hello(ctx context.Context, req *commonPb.HelloRequest, rsp *commonPb.HelloResponse) error {
	rsp.Greeting = "Hello " + req.Name
	return nil
}
