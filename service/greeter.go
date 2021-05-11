package service

import (
	"context"

	greeterPb "tcp/proto/greeter"
)

type Greeter struct{}

func (g *Greeter) Hello(ctx context.Context, req *greeterPb.Request, rsp *greeterPb.Response) error {
	rsp.Greeting = "Hello " + req.Name
	return nil
}
