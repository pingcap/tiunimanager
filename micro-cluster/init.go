package main

import (
	"crypto/tls"
	"fmt"
	mlogrus "github.com/asim/go-micro/plugins/logger/logrus/v3"
	"github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	mlog "github.com/asim/go-micro/v3/logger"
	"github.com/asim/go-micro/v3/transport"
	mylogger "github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/addon/tracer"
	"github.com/pingcap/ticp/config"
	"github.com/pingcap/ticp/micro-cluster/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
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

func initService() {
	cert, err := tls.LoadX509KeyPair(config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath())
	if err != nil {
		mlog.Fatal(err)
		return
	}
	tlsConfigPtr := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	srv := micro.NewService(
		micro.Name(service.TiCPClusterServiceName),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.WrapClient(opentracing.NewClientWrapper(tracer.GlobalTracer)),
		micro.WrapHandler(opentracing.NewHandlerWrapper(tracer.GlobalTracer)),
		micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
	)
	srv.Init()
}

func initPrometheus() {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		addr := fmt.Sprintf(":%d", config.GetPrometheusPort())
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			mlog.Fatal("promhttp ListenAndServe err:", err)
		}
	}()
}

func initClient() {
	// 远程DB服务client
	// 远程平台管理服务client
}
