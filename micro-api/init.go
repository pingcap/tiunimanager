package main

import (
	"crypto/tls"
	"fmt"
	"github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/transport"
	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiem/library/firstparty/config"
	"github.com/pingcap/tiem/library/thirdparty/tracer"
	"github.com/pingcap/tiem/micro-api/route"
	"github.com/pingcap/tiem/micro-cluster/client"
	"log"
)

var TiEMApiServiceName = "go.micro.tiem.api"

func initConfig() {
	config.InitForMonolith()
}

func initService() {
	cert, err := tls.LoadX509KeyPair(config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath())
	if err != nil {
		log.Fatal(err)
		return
	}
	tlsConfigPtr := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}

	srv := micro.NewService(
		micro.Name(TiEMApiServiceName),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.WrapClient(opentracing.NewClientWrapper(tracer.GlobalTracer)),
		micro.WrapHandler(opentracing.NewHandlerWrapper(tracer.GlobalTracer)),
		micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
		micro.Address(":" + config.GetStringWithDefault(config.KEY_API_PORT, "443")),
		micro.Registry(etcd.NewRegistry(registry.Addrs(config.GetRegistryAddress()...))),
	)
	srv.Init()

	go func() {
		if err := srv.Run(); err != nil {
			log.Fatal(err)
		}
	}()
}

func initClient() {
	client.InitManagerClient()
	client.InitClusterClient()
}

func initGinEngine() {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()

	route.Route(g)

	addr := fmt.Sprintf(":%d", 8080)
	if err := g.Run(addr); err != nil {
		log.Fatal(err)
	}

	//if err := g.RunTLS(addr, config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath()); err != nil {
	//	log.Fatal(err)
	//}
}
