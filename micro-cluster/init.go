package main

import (
	"crypto/tls"

	"github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/server"
	"github.com/asim/go-micro/v3/transport"
	"github.com/pingcap/tiem/library/firstparty/config"
	"github.com/pingcap/tiem/library/secondparty/libtiup"
	"github.com/pingcap/tiem/library/thirdparty/logger"
	"github.com/pingcap/tiem/library/thirdparty/tracer"
	cluster "github.com/pingcap/tiem/micro-cluster/proto"
	"github.com/pingcap/tiem/micro-cluster/service"
	clusterAdapt "github.com/pingcap/tiem/micro-cluster/service/cluster/adapt"
	tenantAdapt "github.com/pingcap/tiem/micro-cluster/service/tenant/adapt"

	dbclient "github.com/pingcap/tiem/micro-metadb/client"
)

// Global LogRecord object
var log *logger.LogRecord

func initConfig() {
	config.InitForMonolith()
}

func initLogger() {
	log = logger.GetLogger()
	service.InitClusterLogger()
	service.InitHostLogger()
	log.Debug("init logger completed!")
}

func initClusterOperator() {
	libtiup.MicroInit("./tiupmgr/tiupmgr", "tiup", "")
}

func initService() {
	cert, err := tls.LoadX509KeyPair(config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath())
	if err != nil {
		log.Fatal(err)
		return
	}
	tlsConfigPtr := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}

	serv1 := server.NewServer(
		server.Name(service.TiEMClusterServiceName),
		server.WrapHandler(prometheus.NewHandlerWrapper()),
		server.WrapHandler(opentracing.NewHandlerWrapper(tracer.GlobalTracer)),
		server.Address(config.GetClusterServiceAddress()),
		server.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
		server.Registry(etcd.NewRegistry(registry.Addrs(config.GetRegistryAddress()...))),
	)

	srv1 := micro.NewService(
		micro.Server(serv1),
		micro.WrapClient(opentracing.NewClientWrapper(tracer.GlobalTracer)),
	)

	srv1.Init()

	cluster.RegisterClusterServiceHandler(srv1.Server(), new(service.ClusterServiceHandler))

	go func() {
		if err := srv1.Run(); err != nil {
			log.Fatal(err)
		}
	}()

	serv2 := server.NewServer(
		server.Name(service.TiEMManagerServiceName),
		server.WrapHandler(prometheus.NewHandlerWrapper()),
		server.WrapHandler(opentracing.NewHandlerWrapper(tracer.GlobalTracer)),
		server.Address(config.GetManagerServiceAddress()),
		server.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
		server.Registry(etcd.NewRegistry(registry.Addrs(config.GetRegistryAddress()...))),
	)

	srv2 := micro.NewService(
		micro.Server(serv2),
		micro.WrapClient(opentracing.NewClientWrapper(tracer.GlobalTracer)),
	)
	srv2.Init()

	cluster.RegisterTiEMManagerServiceHandler(srv2.Server(), new(service.ManagerServiceHandler))

	if err := srv2.Run(); err != nil {
		log.Fatal(err)
	}
}

func initClient() {
	dbclient.InitDBClient()
}

func initPort() {
	tenantAdapt.InjectionMetaDbRepo()
	clusterAdapt.InjectionMetaDbRepo()
}
