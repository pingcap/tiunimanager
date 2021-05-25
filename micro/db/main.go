package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	mylogger "github.com/pingcap/tcp/addon/logger"
	"github.com/pingcap/tcp/addon/tracer"
	"github.com/pingcap/tcp/client"
	"github.com/pingcap/tcp/config"
	"github.com/pingcap/tcp/models"
	dbPb "github.com/pingcap/tcp/proto/db"
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
		// only use to init the config
		srv := micro.NewService(
			config.GetMicroCliArgsOption(),
		)
		srv.Init()
		config.Init()
		srv = nil
	}
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
			micro.Name(service.TCP_DB_SERVICE_NAME),
			micro.WrapHandler(prometheus.NewHandlerWrapper()),
			micro.WrapHandler(opentracing.NewHandlerWrapper(tracer.GlobalTracer)),
			micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
		)
		srv.Init()
		models.Init()
		client.InitClient(srv)
		{
			dbPb.RegisterDbHandler(srv.Server(), new(service.Db))
		}
		go func() {
			if err := srv.Run(); err != nil {
				log.Fatal(err)
			}
		}()
		// start promhttp
		http.Handle("/metrics", promhttp.Handler())
		go func() {
			addr := fmt.Sprintf(":%d", config.GetPrometheusPort())
			err := http.ListenAndServe(addr, nil)
			if err != nil {
				log.Fatal("promhttp ListenAndServe err:", err)
			}
		}()
	}
	{
		g := router.SetUpRouter()
		addr := fmt.Sprintf(":%d", config.GetOpenApiPort())
		if err := g.RunTLS(addr, config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath()); err != nil {
			log.Fatal(err)
		}
	}
}
