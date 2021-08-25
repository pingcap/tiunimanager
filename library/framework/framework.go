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
	"github.com/pingcap/tiem/library/firstparty/util"
	"github.com/pingcap/tiem/library/firstparty/util/signal"
	"github.com/pingcap/tiem/library/knowledge"
	"github.com/pingcap/tiem/library/thirdparty/logger"
	"github.com/pingcap/tiem/library/thirdparty/tracer"
)

type Framework interface {
	CallFuncs() error
	Service() error
	Options() error
	Config() error
	Logger() error
	Tracer() error
	Knowledge() error

	StartService() error

	GetLogger() *logger.LogRecord
	GetServiceAddress(name string) string
	GetRegistryAddress() []string
	RegistryServiceHandler( name string) error

	Shutdown(func(bool)) error
}

type Func func(d *BaseFramework) error

type BaseFramework struct {
	log		*logger.LogRecord
	options micro.Option
	service micro.Service
	args config.ClientArgs
	fs 		[] Func

}

func (b* BaseFramework) Service (name string) error {
	cert, err := tls.LoadX509KeyPair(config.GetCertificateCrtFilePath(), config.GetCertificateKeyFilePath())
	if err != nil {
		panic("load certificate file failed")
	}
	tlsConfigPtr := &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	b.service = micro.NewService(
		micro.Name(name),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.WrapClient(opentracing.NewClientWrapper(tracer.GlobalTracer)),
		micro.WrapHandler(opentracing.NewHandlerWrapper(tracer.GlobalTracer)),
		micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(tlsConfigPtr))),
		micro.Address(b.GetServiceAddress(name)),
		micro.Registry(etcd.NewRegistry(registry.Addrs(b.GetRegistryAddress()...))),
	)
	b.service.Init()
	return nil
}

func (b* BaseFramework) Options (name string) error {
	panic ("Service must be overloaded implementation")
	return nil
}

func (b* BaseFramework) Config(mod config.Mod) error {
	config.InitForMonolith(mod)
	return nil
}

func (b* BaseFramework) Logger(key config.Key) error {
	b.log = logger.GetLogger(key)
	util.Assert(b.log != nil)
	return nil
}

func (b* BaseFramework) Tracer() error {
	tracer.InitTracer()
	return nil
}

func (b* BaseFramework) Knowledge() error {
	knowledge.LoadKnowledge()
	return nil
}

func (b* BaseFramework) GetLogger() *logger.LogRecord {
	util.Assert(b.log != nil)
	return b.log
}

func (b* BaseFramework) GetServiceAddress(name string) string {
	config.GetApiServiceAddress()
	return ""
}

func (b* BaseFramework) GetRegistryAddress() []string {
	return config.GetRegistryAddress()
}

func (b* BaseFramework) RegistryServiceHandler( name string) error {
	panic ("Service must be overloaded implementation")
	return nil
}

func (b* BaseFramework) Shutdown( fname func (bool)) {
	signal.SetupSignalHandler(fname)
}

func (b *BaseFramework) CallFuncs() error {
	for _, opt := range b.fs {
		util.AssertNoErr(opt(b))
	}
	return nil
}

//var f Framework

/*func GetDefaultLogger() *logger.LogRecord {
	return f.GetDefaultLogger()
}*/

/*type Opt func(d *DefaultServiceFramework) error

type DefaultServiceFramework struct {
	serviceEnum 		MicroServiceEnum
	flags       		micro.Option

	initOpts []Opt

	log 				*logger.LogRecord
	service 			micro.Service
}

func (p *DefaultServiceFramework) Init() error {
	for _, opt := range p.initOpts {
		util.AssertNoErr(opt(p))
	}
	return nil
}

func (p *DefaultServiceFramework) GetDefaultLogger() *logger.LogRecord {
	return p.log
}

func (p *DefaultServiceFramework) GetRegistryAddress() []string {
	return config.GetRegistryAddress()
}*/

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
	p.service = p.serviceEnum.BuildMicroService(p.GetRegistryAddress()...)

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
