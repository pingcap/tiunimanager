package main

import (
	"log"

	"tcp/addon/tracer"
	greeterPb "tcp/proto/greeter"
	"tcp/router"
	"tcp/service"

	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
)

func main() {
	{
		srv := micro.NewService(
			micro.Name("go.micro.greeter"),
			micro.WrapHandler(opentracing.NewHandlerWrapper(tracer.GlobalTracer)),
		)
		srv.Init()
		{
			greeterPb.RegisterGreeterHandler(srv.Server(), new(service.Greeter))
		}
		go func() {
			if err := srv.Run(); err != nil {
				log.Fatal(err)
			}
		}()
	}
	{
		g := router.SetUpRouter()
		if err := g.RunTLS(":443", "config/example/server.crt", "config/example/server.key"); err != nil {
			log.Fatal(err)
		}
	}
}
