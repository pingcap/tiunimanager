package adapt

import (
	"context"
	"github.com/pingcap/ticp/cluster/interfaces/proto"
)

const TICP_COMMON_SERVICE_NAME = "go.micro.ticp.common"

type Common struct{}

func (c *Common) Hello(ctx context.Context, req *proto.HelloRequest, rsp *proto.HelloResponse) error {
	rsp.Greeting = "Hello " + req.Name
	return nil
}
