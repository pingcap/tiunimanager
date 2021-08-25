package framework

import (
	"crypto/tls"
	"github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/transport"
	"github.com/pingcap/tiem/library/firstparty/config"
	"github.com/pingcap/tiem/library/thirdparty/logger"
	"github.com/pingcap/tiem/library/thirdparty/tracer"
)

type MicroServiceEnum string

const (
	MetaDBService  MicroServiceEnum = "go.micro.tiem.db"
	ClusterService MicroServiceEnum = "go.micro.tiem.cluster"
	ApiService     MicroServiceEnum = "go.micro.tiem.api"
)

func (p MicroServiceEnum) ToString() string {
	return string(p)
}

func (p MicroServiceEnum) buildArgsOption() micro.Option {
	switch p {
	case MetaDBService:
		return config.GetMicroMetaDBCliArgsOption()
	case ClusterService:
		return config.GetMicroClusterCliArgsOption()
	case ApiService:
		return config.GetMicroApiCliArgsOption()
	default:
		panic("Illegal MicroServiceEnum")
	}
}

func (p MicroServiceEnum) buildLogger() *logger.LogRecord {
	switch p {
	case MetaDBService: return logger.GetLogger(config.KEY_METADB_LOG)
	case ClusterService: return logger.GetLogger(config.KEY_CLUSTER_LOG)
	case ApiService: return logger.GetLogger(config.KEY_API_LOG)
	default:
		panic("Illegal MicroServiceEnum")
	}
}

func (p MicroServiceEnum) BuildMicroService(addrs ...string) micro.Service {
	cert, err := tls.LoadX509KeyPair(config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath())
	if err != nil {
		panic("load certificate file failed")
	}
	tlsConfigPtr := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	srv := micro.NewService(
		micro.Name(string(p)),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.WrapClient(opentracing.NewClientWrapper(tracer.GlobalTracer)),
		micro.WrapHandler(opentracing.NewHandlerWrapper(tracer.GlobalTracer)),
		micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
		micro.Address(p.buildServiceAddress()),
		micro.Registry(etcd.NewRegistry(registry.Addrs(addrs...))),
	)
	srv.Init()

	return srv
}

func (p MicroServiceEnum) buildServiceAddress() string {
	switch p {
	case MetaDBService: return config.GetMetaDBServiceAddress()
	case ClusterService: return config.GetClusterServiceAddress()
	case ApiService: return config.GetApiServiceAddress()
	default:
		panic("Illegal MicroServiceEnum")
	}
}

func (p MicroServiceEnum) clientList() []MicroServiceEnum {
	switch p {
	case MetaDBService: return []MicroServiceEnum{}
	case ClusterService: return []MicroServiceEnum{MetaDBService}
	case ApiService: return []MicroServiceEnum{ClusterService}
	default:
		panic("Illegal MicroServiceEnum")
	}
}

func (p MicroServiceEnum) logMod() config.Mod {
	switch p {
	case MetaDBService: return config.MicroMetaDBMod
	case ClusterService: return config.MicroClusterMod
	case ApiService: return config.MicroApiMod
	default:
		panic("Illegal MicroServiceEnum")
	}
}
