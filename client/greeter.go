package client

import (
	"crypto/tls"
	"log"
	"tcp/addon/tracer"
	"tcp/config"
	greeterPb "tcp/proto/greeter"

	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/transport"
)

// Make request
/*
	rsp, err := GreeterClient.Hello(context.Background(), &pb.Request{
		Name: "Foo",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(rsp.Greeting)
*/
var GreeterClient greeterPb.GreeterService

func init() {
	// tls
	cert, err := tls.LoadX509KeyPair(config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath())
	if err != nil {
		log.Fatal(err)
		return
	}
	tlsConfigPtr := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	// create a new service
	service := micro.NewService(
		micro.WrapClient(opentracing.NewClientWrapper(tracer.GlobalTracer)),
		micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
	)

	// parse command line flags
	service.Init()

	// Use the generated client stub
	GreeterClient = greeterPb.NewGreeterService("go.micro.greeter", service.Client())
}
