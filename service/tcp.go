package service

import (
	"context"

	tcpPb "tcp/proto/tcp"
)

type Tcp struct{}

func (g *Tcp) Hello(ctx context.Context, req *tcpPb.HelloRequest, rsp *tcpPb.HelloResponse) error {
	rsp.Greeting = "Hello " + req.Name
	return nil
}
