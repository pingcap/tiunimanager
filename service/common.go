package service

import (
	"context"

	commonPb "github.com/pingcap/ticp/proto/common"
)

const TICP_COMMON_SERVICE_NAME = "go.micro.ticp.ticp-server"

type Common struct{}

func (c *Common) Hello(ctx context.Context, req *commonPb.HelloRequest, rsp *commonPb.HelloResponse) error {
	rsp.Greeting = "Hello " + req.Name
	return nil
}
