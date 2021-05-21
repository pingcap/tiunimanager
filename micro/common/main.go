package main

import (
	"crypto/tls"
	"log"
	"net/http"

	mylogger "github.com/pingcap/tcp/addon/logger"
	"github.com/pingcap/tcp/addon/tracer"
	"github.com/pingcap/tcp/config"
	commonPb "github.com/pingcap/tcp/proto/common"
	"github.com/pingcap/tcp/router"
	"github.com/pingcap/tcp/service"

	"github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	mlogrus "github.com/asim/go-micro/plugins/logger/logrus/v3"
	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	mlog "github.com/asim/go-micro/v3/logger"
	"github.com/asim/go-micro/v3/transport"
)

func init() {
	// log
	mlog.DefaultLogger = mlogrus.NewLogger(mlogrus.WithLogger(mylogger.WithContext(nil)))
}

func main() {
	{
		// tls
		cert, err := tls.LoadX509KeyPair(config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath())
		if err != nil {
			log.Fatal(err)
			return
		}
		tlsConfigPtr := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
		//
		srv := micro.NewService(
			micro.Name("go.micro.tcp.common"),
			micro.WrapHandler(prometheus.NewHandlerWrapper()),
			micro.WrapHandler(opentracing.NewHandlerWrapper(tracer.GlobalTracer)),
			micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
		)
		srv.Init()
		{
			commonPb.RegisterCommonHandler(srv.Server(), new(service.Common))
		}
		go func() {
			if err := srv.Run(); err != nil {
				log.Fatal(err)
			}
		}()
		// start promhttp
		http.Handle("/metrics", promhttp.Handler())
		go func() {
			err := http.ListenAndServe(":8080", nil)
			if err != nil {
				log.Fatal("promhttp ListenAndServe err:", err)
			}
		}()
	}
	{
		g := router.SetUpRouter()
		if err := g.RunTLS(":443", config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath()); err != nil {
			log.Fatal(err)
		}
	}
}
