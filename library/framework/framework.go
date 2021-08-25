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
)

var Current Framework

type Framework interface {
	Init() error
	Shutdown() error

	GetClientArgs() *ClientArgs
	GetConfiguration() *Configuration
	GetLogger() *LogRecord
	GetTracer() *Tracer

	GetServiceMeta() ServiceMeta
	StartService() error
	StopService() error
}

func GetLogger() *LogRecord {
	if Current != nil {
		return Current.GetLogger()
	} else {
		panic("framework not ")
	}
}

type Opt func(d *BaseFramework) error
type ServiceHandler func(service micro.Service) error
type ClientHandler func(service micro.Service) error

type BaseFramework struct {
	args          *ClientArgs
	configuration *Configuration
	log           *LogRecord
	trace         *Tracer
	certificate   *CertificateInfo

	serviceMeta  *ServiceMeta
	microService micro.Service

	initOpts 		[]Opt
	shutdownOpts 	[]Opt

	clientHandler  map[ServiceNameEnum]ClientHandler
	serviceHandler ServiceHandler
}

func InitBaseFrameworkFromArgs(serviceName ServiceNameEnum, opts ...Opt) *BaseFramework {
	f := new(BaseFramework)

	f.args = new(ClientArgs)
	// receive all falgs
	micro.Flags(AllFlags(f.args)...)

	f.serviceMeta = NewServiceMetaFromArgs(serviceName, f.args)
	f.log = NewLogRecordFromArgs(f.args)
	f.certificate = NewCertificateFromArgs(f.args)
	f.trace = NewTracerFromArgs(f.args)
	// now empty
	f.configuration = &Configuration{}

	f.Init()

	return f
}

func (b *BaseFramework) GetDataDir() string {
	return b.args.DataDir
}

func (b *BaseFramework) GetDeployDir() string {
	return b.args.DeployDir
}

func (b *BaseFramework) PrepareServiceHandler(handler ServiceHandler) {
	b.serviceHandler = handler
}

func (b *BaseFramework) PrepareClientHandler(clientHandlerMap map[ServiceNameEnum]ClientHandler) {
	b.clientHandler = clientHandlerMap
}

func (b *BaseFramework) Init() error {
	for _, opt := range b.initOpts {
		AssertNoErr(opt(b))
	}
	return nil
}

func (b *BaseFramework) Shutdown() error {
	for _, opt := range b.shutdownOpts {
		AssertNoErr(opt(b))
	}
	return nil
}

func (b *BaseFramework) GetClientArgs() *ClientArgs {
	return b.args
}

func (b *BaseFramework) GetConfiguration() *Configuration {
	return b.configuration
}

func (b *BaseFramework) GetLogger() *LogRecord {
	return b.log
}

func (b *BaseFramework) GetTracer() *Tracer {
	return b.trace
}

func (b *BaseFramework) GetServiceMeta() *ServiceMeta {
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
			micro.Name(string(client)),
			micro.WrapHandler(prometheus.NewHandlerWrapper()),
			micro.WrapClient(opentracing.NewClientWrapper(*b.trace)),
			micro.WrapHandler(opentracing.NewHandlerWrapper(*b.trace)),
			micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
			micro.Registry(etcd.NewRegistry(registry.Addrs(b.GetServiceMeta().RegistryAddress...))),
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
	p.getLogger().Debug("init logger completed!")
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
