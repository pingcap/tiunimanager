package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	mlogrus "github.com/asim/go-micro/plugins/logger/logrus/v3"
	"github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	mlog "github.com/asim/go-micro/v3/logger"
	"github.com/asim/go-micro/v3/transport"
	mylogger "github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/addon/tracer"
	"github.com/pingcap/ticp/config"
	manager "github.com/pingcap/ticp/micro-manager/proto"
	"github.com/pingcap/ticp/micro-manager/service"
	"github.com/pingcap/ticp/micro-manager/service/tenant/adapt"
	dbclient "github.com/pingcap/ticp/micro-metadb/client"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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

func initLogger() {
	// log
	mlog.DefaultLogger = mlogrus.NewLogger(mlogrus.WithLogger(mylogger.WithContext(nil)))
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

func initService() {
	cert, err := tls.LoadX509KeyPair(config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath())
	if err != nil {
		mlog.Fatal(err)
		return
	}
	tlsConfigPtr := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	srv := micro.NewService(
		micro.Name(service.TiCPManagerServiceName),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.WrapClient(opentracing.NewClientWrapper(tracer.GlobalTracer)),
		micro.WrapHandler(opentracing.NewHandlerWrapper(tracer.GlobalTracer)),
		micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
	)
	srv.Init()

	manager.RegisterTiCPManagerServiceHandler(srv.Server(), new(service.ManagerServiceHandler))

	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}

func initClient() {
	dbclient.InitDBClient()
}

func initPort() {
	adapt.InjectionMetaDbRepo()
}
