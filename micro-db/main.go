package main

import (
	mlogrus "github.com/asim/go-micro/plugins/logger/logrus/v3"
	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	mlog "github.com/asim/go-micro/v3/logger"
	mylogger "github.com/pingcap/ticp/addon/logger"
)

func init() {
	// log
	mlog.DefaultLogger = mlogrus.NewLogger(mlogrus.WithLogger(mylogger.WithContext(nil)))
}

func main() {
	//{
	//	// only use to init the config
	//	srv := micro.NewService(
	//		config.GetMicroCliArgsOption(),
	//	)
	//	srv.Init()
	//	config.Init()
	//	srv = nil
	//}
	//{
	//	// tls
	//	cert, err := tls.LoadX509KeyPair(config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath())
	//	if err != nil {
	//		log.Fatal(err)
	//		return
	//	}
	//	tlsConfigPtr := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	//	//
	//	srv := micro.NewService(
	//		micro.Name(adapt.TICP_DB_SERVICE_NAME),
	//		micro.WrapHandler(prometheus.NewHandlerWrapper()),
	//		micro.WrapClient(opentracing.NewClientWrapper(tracer.GlobalTracer)),
	//		micro.WrapHandler(opentracing.NewHandlerWrapper(tracer.GlobalTracer)),
	//		micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
	//	)
	//	srv.Init()
	//	temp.Init()
	//	temp.InitClient(srv)
	//	{
	//		proto.RegisterDbHandler(srv.Server(), new(adapt.Db))
	//	}
	//	go func() {
	//		if err := srv.Run(); err != nil {
	//			log.Fatal(err)
	//		}
	//	}()
	//	// start promhttp
	//	http.Handle("/metrics", promhttp.Handler())
	//	go func() {
	//		addr := fmt.Sprintf(":%d", config.GetPrometheusPort())
	//		err := http.ListenAndServe(addr, nil)
	//		if err != nil {
	//			log.Fatal("promhttp ListenAndServe err:", err)
	//		}
	//	}()
	//}
	//{
	//	g := api.SetUpRouter()
	//	addr := fmt.Sprintf(":%d", config.GetOpenApiPort())
	//	if err := g.RunTLS(addr, config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath()); err != nil {
	//		log.Fatal(err)
	//	}
	//}
}
