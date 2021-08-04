package main

import (
	"crypto/tls"
	mlogrus "github.com/asim/go-micro/plugins/logger/logrus/v3"
	"github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	mlog "github.com/asim/go-micro/v3/logger"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/transport"
	"github.com/pingcap/ticp/library/firstparty/config"
	"github.com/pingcap/ticp/library/secondparty/libtiup"
	"github.com/pingcap/ticp/library/thirdparty/logger"
	"github.com/pingcap/ticp/library/thirdparty/tracer"
	cluster "github.com/pingcap/ticp/micro-cluster/proto"
	"github.com/pingcap/ticp/micro-cluster/service"
	"github.com/pingcap/ticp/micro-cluster/service/tenant/adapt"
	dbclient "github.com/pingcap/ticp/micro-metadb/client"
	"log"
)

func initConfig() {
	config.InitForMonolith()
}

func initLogger() {
	// log
	mlog.DefaultLogger = mlogrus.NewLogger(mlogrus.WithLogger(logger.WithContext(nil)))
}

func initClusterOperator() {
	libtiup.MicroInit("./tiupmgr/tiupmgr", "tiup", "")
}

func initService() {
	cert, err := tls.LoadX509KeyPair(config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath())
	if err != nil {
		mlog.Fatal(err)
		return
	}
	tlsConfigPtr := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	srv1 := micro.NewService(
		micro.Name(service.TiCPClusterServiceName),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.WrapClient(opentracing.NewClientWrapper(tracer.GlobalTracer)),
		micro.WrapHandler(opentracing.NewHandlerWrapper(tracer.GlobalTracer)),
		micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
		micro.Registry(etcd.NewRegistry(registry.Addrs(config.GetRegistryAddress()...))),
	)
	srv1.Init()

	cluster.RegisterClusterServiceHandler(srv1.Server(), new(service.ClusterServiceHandler))

	if err := srv1.Run(); err != nil {
		log.Fatal(err)
	}

	srv2 := micro.NewService(
		micro.Name(service.TiCPManagerServiceName),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.WrapClient(opentracing.NewClientWrapper(tracer.GlobalTracer)),
		micro.WrapHandler(opentracing.NewHandlerWrapper(tracer.GlobalTracer)),
		micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
		micro.Registry(etcd.NewRegistry(registry.Addrs(config.GetRegistryAddress()...))),
	)
	srv2.Init()

	cluster.RegisterTiCPManagerServiceHandler(srv2.Server(), new(service.ManagerServiceHandler))

	if err := srv2.Run(); err != nil {
		log.Fatal(err)
	}
}

func initClient() {
	dbclient.InitDBClient()
}

func initPort() {
	adapt.InjectionMetaDbRepo()
}

