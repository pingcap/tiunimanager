package client

import (
	"crypto/tls"
	"fmt"
	"github.com/asim/go-micro/plugins/registry/etcd/v3"
	_ "github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	mlog "github.com/asim/go-micro/v3/logger"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/transport"
	"github.com/pingcap/ticp/library/firstparty/config"
	"github.com/pingcap/ticp/library/thirdparty/tracer"
	db "github.com/pingcap/ticp/micro-metadb/proto"
	"github.com/pingcap/ticp/micro-metadb/service"
)

var DBClient db.TiCPDBService

func InitDBClient() {
	cert, err := tls.LoadX509KeyPair(config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath())
	if err != nil {
		mlog.Fatal(err)
		return
	}
	tlsConfigPtr := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	srv := micro.NewService(
		micro.Name(service.TiCPMetaDBServiceName),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.WrapClient(opentracing.NewClientWrapper(tracer.GlobalTracer)),
		micro.WrapHandler(opentracing.NewHandlerWrapper(tracer.GlobalTracer)),
		micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
		micro.Registry(etcd.NewRegistry(registry.Addrs(config.GetRegistryAddress()...))),
	)
	srv.Init()

	DBClient = db.NewTiCPDBService(service.TiCPMetaDBServiceName, srv.Client())

	fmt.Println(DBClient)
}
