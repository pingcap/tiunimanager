package framework

import (
	"crypto/tls"
	"errors"
	"github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/server"
	"github.com/asim/go-micro/v3/transport"
	"github.com/pingcap-inc/tiem/library/common"
	log "github.com/sirupsen/logrus"
	"os"
)

var Current Framework

type Framework interface {
	Init() error
	Shutdown() error

	GetClientArgs() *ClientArgs
	GetConfiguration() *Configuration
	GetRootLogger() *RootLogger
	GetTracer() *Tracer

	GetServiceMeta() *ServiceMeta
	StartService() error
	StopService() error
}

func GetRootLogger() *RootLogger {
	if Current != nil {
		return Current.GetRootLogger()
	} else {
		return DefaultLogRecord()
	}
}

func Log() *log.Entry {
	if Current != nil {
		return Current.GetRootLogger().RecordFun()
	} else {
		return DefaultLogRecord().RecordFun()
	}
}

type Opt func(d *BaseFramework) error
type ServiceHandler func(service micro.Service) error
type ClientHandler func(service micro.Service) error

type BaseFramework struct {
	args          *ClientArgs
	configuration *Configuration
	log           *RootLogger
	trace         *Tracer
	certificate   *CertificateInfo

	serviceMeta  *ServiceMeta
	microService micro.Service

	initOpts 		[]Opt
	shutdownOpts 	[]Opt

	clientHandler  map[ServiceNameEnum]ClientHandler
	serviceHandler ServiceHandler
}

func InitBaseFrameworkForUt(serviceName ServiceNameEnum, opts ...Opt) *BaseFramework {
	f := new(BaseFramework)

	f.args = &ClientArgs{
		Host: "127.0.0.1",
		Port: 4116,
		MetricsPort: 4121,
		RegistryClientPort: 4101,
		RegistryPeerPort: 4102,
		RegistryAddress: "127.0.0.1:4101",
		DeployDir: "./../bin",
		DataDir: "./testdata",
		LogLevel: "info",
	}
	f.parseArgs(serviceName)

	f.initOpts = opts
	f.Init()

	f.shutdownOpts = []Opt {
		func(d *BaseFramework) error {
			return os.RemoveAll(d.GetDataDir())
		},
	}

	Current = f

	return f
}

func InitBaseFrameworkFromArgs(serviceName ServiceNameEnum, opts ...Opt) *BaseFramework {
	f := new(BaseFramework)

	f.acceptArgs()
	f.parseArgs(serviceName)

	f.initOpts = opts
	f.Init()

	Current = f
	return f
}

func (b *BaseFramework) acceptArgs() {
	b.args = new(ClientArgs)
	// receive all falgs
	srv := micro.NewService(
		micro.Flags(AllFlags(b.args)...),
	)
	srv.Init()
	srv = nil
}

func (b *BaseFramework) parseArgs(serviceName ServiceNameEnum) {
	b.serviceMeta = NewServiceMetaFromArgs(serviceName, b.args)
	b.log = NewLogRecordFromArgs(serviceName, b.args)
	b.certificate = NewCertificateFromArgs(b.args)
	b.trace = NewTracerFromArgs(b.args)
	// now empty
	b.configuration = &Configuration{}
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
			micro.WrapHandler(opentracing.NewHandlerWrapper(*b.trace)),
			micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
			micro.Registry(etcd.NewRegistry(registry.Addrs(b.GetServiceMeta().RegistryAddress...))),
			micro.WrapClient(opentracing.NewClientWrapper(*b.trace)),
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

	server := server.NewServer(
		server.Name(string(b.serviceMeta.ServiceName)),
		server.WrapHandler(prometheus.NewHandlerWrapper()),
		server.WrapHandler(opentracing.NewHandlerWrapper(*b.trace)),
		server.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
		server.Address(b.serviceMeta.GetServiceAddress()),
		server.Registry(etcd.NewRegistry(registry.Addrs(b.serviceMeta.RegistryAddress...))),
	)

	srv := micro.NewService(
		micro.Server(server),
		micro.WrapClient(opentracing.NewClientWrapper(*b.trace)),
	)
	srv.Init()

	b.microService = srv

	b.serviceHandler(b.microService)
}

func (b *BaseFramework) GetDataDir() string {
	return b.args.DataDir
}

func (b *BaseFramework) GetDeployDir() string {
	return b.args.DeployDir
}

func (b *BaseFramework) PrepareService(handler ServiceHandler) {
	b.serviceHandler = handler
	b.initMicroService()
}

func (b *BaseFramework) PrepareClientClient(clientHandlerMap map[ServiceNameEnum]ClientHandler) {
	b.clientHandler = clientHandlerMap
	b.initMicroClient()
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

func (b *BaseFramework) GetRootLogger() *RootLogger {
	return b.log
}

func (b *BaseFramework) GetTracer() *Tracer {
	return b.trace
}

func (b *BaseFramework) GetServiceMeta() *ServiceMeta {
	return b.serviceMeta
}

func (b *BaseFramework) StopService() error {
	panic("implement me")
}

func (b *BaseFramework) StartService() error {
	if err := b.microService.Run(); err != nil {
		b.GetRootLogger().ForkFile(common.LOG_FILE_SYSTEM).Fatalf("Initialization micro service failed, error %v, listening address %s, etcd registry address %s", err, b.serviceMeta.GetServiceAddress(), b.serviceMeta.RegistryAddress)
		return errors.New("initialization micro service failed")
	}

	return nil
}
