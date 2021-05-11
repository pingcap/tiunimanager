package main

import (
	"log"

	greeterPb "tcp/proto/greeter"
	"tcp/router"
	"tcp/service"

	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/v3"
	"tcp/models"
)

func main() {
	{
		srv := micro.NewService(
			micro.Name("greeter"),
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
		if err := g.Run(); err != nil {
			log.Fatal(err)
		}
	}
}
