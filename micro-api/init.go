package main

import (
	"crypto/tls"
	"fmt"
	"github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/transport"
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/addon/tracer"
	"github.com/pingcap/ticp/config"
	"github.com/pingcap/ticp/micro-api/route"
	clusterclient "github.com/pingcap/ticp/micro-cluster/client"
	managerclient "github.com/pingcap/ticp/micro-manager/client"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

var TiCPApiServiceName = "go.micro.ticp.api"

func initConfig() {
	{
		// only use to init the config
		srv := micro.NewService(
			config.GetMicroCliArgsOption(),
		)
		srv.Init()
		config.Init()
		srv = nil
	}
}
func initService() {
	cert, err := tls.LoadX509KeyPair(config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath())
	if err != nil {
		log.Fatal(err)
		return
	}
	tlsConfigPtr := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}

	srv := micro.NewService(
		micro.Name(TiCPApiServiceName),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.WrapClient(opentracing.NewClientWrapper(tracer.GlobalTracer)),
		micro.WrapHandler(opentracing.NewHandlerWrapper(tracer.GlobalTracer)),
		micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
	)
	srv.Init()

	go func() {
		if err := srv.Run(); err != nil {
			log.Fatal(err)
		}
	}()
}
func initClient() {
	managerclient.InitManagerClient()
	clusterclient.InitClusterClient()
}

func initPrometheus() {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		addr := fmt.Sprintf(":%d", config.GetPrometheusPort())
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			log.Fatal("promhttp ListenAndServe err:", err)
		}
	}()
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
