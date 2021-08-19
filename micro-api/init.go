package main

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/transport"
	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiem/library/firstparty/config"
	"github.com/pingcap/tiem/library/knowledge"
	"github.com/pingcap/tiem/library/thirdparty/tracer"
	"github.com/pingcap/tiem/micro-api/route"
	"github.com/pingcap/tiem/micro-cluster/client"
)

var TiEMApiServiceName = "go.micro.tiem.api"

func initConfig() {
	srv := micro.NewService(
		config.GetMicroApiCliArgsOption(),
	)
	srv.Init()
	srv = nil

	config.InitForMonolith(config.MicroApiMod)
}

func initTracer() {
	tracer.InitTracer()
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
		micro.Address(config.GetApiServiceAddress()),
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
	client.InitClusterClient()
	client.InitManagerClient()
}

func initGinEngine() {
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()

	route.Route(g)

	port := config.GetClientArgs().RestPort
	if port <= 0 {
		port = config.DefaultRestPort
	}
	addr := fmt.Sprintf(":%d", port)
	if err := g.Run(addr); err != nil {
		log.Fatal(err)
	}

	//if err := g.RunTLS(addr, config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath()); err != nil {
	//	log.Fatal(err)
	//}
}

func initKnowledge() {
	knowledge.LoadKnowledge()
}
