package framework

import (
	"crypto/tls"
	"errors"
	"github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/transport"
	"github.com/pingcap-inc/tiem/library/firstparty/util"
	"github.com/pingcap-inc/tiem/library/framework/args"
	"github.com/pingcap-inc/tiem/library/framework/certificate"
	"github.com/pingcap-inc/tiem/library/framework/config"
	"github.com/pingcap-inc/tiem/library/framework/logger"
	"github.com/pingcap-inc/tiem/library/framework/servicemeta"
	"github.com/pingcap-inc/tiem/library/framework/tracer"
)

var Current Framework

type Framework interface {
	Init() error
	Shutdown() error

	GetClientArgs() *args.ClientArgs
	GetConfiguration() *config.Configuration
	GetLogger() *logger.LogRecord
	GetTracer() *tracer.Tracer

	GetServiceMeta() servicemeta.ServiceMeta
	StartService() error
	StopService() error

}

type Opt func(d *BaseFramework) error
type ServiceHandler func(service micro.Service) error
type ClientHandler func(service micro.Service) error

type BaseFramework struct {
	args          *args.ClientArgs
	configuration *config.Configuration
	log           *logger.LogRecord
	trace         *tracer.Tracer
	certificate   *certificate.CertificateInfo

	serviceMeta  *servicemeta.ServiceMeta
	microService micro.Service

	initOpts 		[]Opt
	shutdownOpts 	[]Opt

	clientHandler  map[*servicemeta.ServiceMeta]ClientHandler
	serviceHandler ServiceHandler
}

func InitBaseFrameworkFromArgs(serviceName servicemeta.ServiceNameEnum, opts ...Opt) *BaseFramework {
	f := new(BaseFramework)

	f.args = new(args.ClientArgs)
	// receive all falgs
	micro.Flags(args.AllFlags(f.args)...)

	f.serviceMeta = servicemeta.NewServiceMetaFromArgs(serviceName, f.args)
	f.log = logger.NewLogRecordFromArgs(f.args)
	f.certificate = certificate.NewCertificateFromArgs(f.args)
	f.trace = tracer.NewTracerFromArgs(f.args)
	// now empty
	f.configuration = &config.Configuration{}

	f.Init()

	return f
}

func (b *BaseFramework) PrepareServiceHandler(handler ServiceHandler) {
	b.serviceHandler = handler
}

func (b *BaseFramework) PrepareClientHandler(clientHandlerMap map[*servicemeta.ServiceMeta]ClientHandler) {
	b.clientHandler = clientHandlerMap
}

func (b *BaseFramework) Init() error {
	for _, opt := range b.initOpts {
		util.AssertNoErr(opt(b))
	}
	return nil
}

func (b *BaseFramework) Shutdown() error {
	for _, opt := range b.shutdownOpts {
		util.AssertNoErr(opt(b))
	}
	return nil
}

func (b *BaseFramework) GetClientArgs() *args.ClientArgs {
	return b.args
}

func (b *BaseFramework) GetConfiguration() *config.Configuration {
	return b.configuration
}

func (b *BaseFramework) GetLogger() *logger.LogRecord {
	return b.log
}

func (b *BaseFramework) GetTracer() *tracer.Tracer {
	return b.trace
}

func (b *BaseFramework) GetServiceMeta() *servicemeta.ServiceMeta {
	return b.serviceMeta
}

func (b *BaseFramework) initMicroClient() {
	for client, handler := range b.clientHandler {
		cert, err := tls.LoadX509KeyPair(b.certificate.CertificateCrtFilePath, b.certificate.CertificateKeyFilePath)
		if err != nil {
			panic("load certificate file failed")
		}
		tlsConfigPtr := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
		srv := micro.NewService(
			micro.Name(string(client.ServiceName)),
			micro.WrapHandler(prometheus.NewHandlerWrapper()),
			micro.WrapClient(opentracing.NewClientWrapper(*b.trace)),
			micro.WrapHandler(opentracing.NewHandlerWrapper(*b.trace)),
			micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
			micro.Address(client.GetServiceAddress()),
			micro.Registry(etcd.NewRegistry(registry.Addrs(client.RegistryAddress...))),
		)
		srv.Init()
		handler(srv)
	}
}

func (b *BaseFramework) initMicroService() {
	cert, err := tls.LoadX509KeyPair(b.certificate.CertificateCrtFilePath, b.certificate.CertificateKeyFilePath)
	if err != nil {
		panic("load certificate file failed")
	}
	tlsConfigPtr := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	srv := micro.NewService(
		micro.Name(string(b.serviceMeta.ServiceName)),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.WrapClient(opentracing.NewClientWrapper(*b.trace)),
		micro.WrapHandler(opentracing.NewHandlerWrapper(*b.trace)),
		micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
		micro.Address(b.serviceMeta.GetServiceAddress()),
		micro.Registry(etcd.NewRegistry(registry.Addrs(b.serviceMeta.RegistryAddress...))),
	)
	srv.Init()

	b.microService = srv
}

func (b *BaseFramework) StopService() error {
	panic("implement me")
}

func (b *BaseFramework) StartService() error {
	if err := b.microService.Run(); err != nil {
		b.GetLogger().Fatalf("Initialization micro service failed, error %v, listening address %s, etcd registry address %s", err, b.serviceMeta.GetServiceAddress(), b.serviceMeta.RegistryAddress)
		return errors.New("initialization micro service failed")
	}

	return nil
}

/*func NewDefaultFramework(serviceName MicroServiceEnum, initOpt ...Opt) *DefaultServiceFramework {
	p := &DefaultServiceFramework{
		serviceEnum: serviceName,
		initOpts: []Opt{
			initConfig,
			initCurrentLogger,
			initKnowledge,
			initTracer,
			initShutdownFunc,
		},
	}

	p.initOpts = append(p.initOpts, initOpt...)

	p.Init()

	f = p
	return p
}

func (p *DefaultServiceFramework) StartService() error {
	p.service = p.serviceEnum.initMicroService(p.GetRegistryAddress()...)

	switch p.serviceEnum {
	case MetaDBService:
		util.AssertNoErr(dbPb.RegisterTiEMDBServiceHandler(p.service.Server(), new(dbSrv.DBServiceHandler)))
	case ClusterService:
		util.AssertNoErr(clusterPb.RegisterClusterServiceHandler(p.service.Server(), new(clusterSrv.ClusterServiceHandler)))
	default:
		panic("Illegal MicroServiceEnum")
	}

	util.Assert(p.service != nil)

	if err := p.service.Run(); err != nil {
		p.GetDefaultLogger().Fatalf("Initialization micro service failed, error %v, listening address %s, etcd registry address %s", err, config.GetMetaDBServiceAddress(), config.GetRegistryAddress())
		return errors.New("initialization micro service failed")
	}

	return nil
}*/

/*func initConfig(p *DefaultServiceFramework) error {
	p.flags = p.serviceEnum.buildArgsOption()
	srv := micro.NewService(
		p.flags,
	)
	srv.Init()
	srv = nil
	config.InitForMonolith(p.serviceEnum.logMod())
	return nil
}

func initCurrentLogger(p *DefaultServiceFramework) error {
	p.log = p.serviceEnum.buildLogger()
	// use log
	p.log.Debug("init logger completed!")
	return nil
}

func initKnowledge(p *DefaultServiceFramework) error {
	knowledge.LoadKnowledge()
	return nil

}

func initTracer(p *DefaultServiceFramework) error {
	tracer.InitTracer()
	return nil
}

func initShutdownFunc(p *DefaultServiceFramework) error {
	mysignal.SetupSignalHandler(func(bool) {
		// todo do something before quit
	})
	return nil
}*/
