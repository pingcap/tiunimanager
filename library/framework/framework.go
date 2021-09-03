package framework

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/pingcap-inc/tiem/library/common"

	"github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/server"
	"github.com/asim/go-micro/v3/transport"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var Current Framework

type Framework interface {
	Init() error
	Shutdown() error

	GetClientArgs() *ClientArgs
	GetConfiguration() *Configuration
	GetRootLogger() *RootLogger
	LogWithContext(context.Context) *log.Entry
	GetTracer() *Tracer
	GetEtcdClient() *EtcdClient

	GetServiceMeta() *ServiceMeta
	StartService() error
	StopService() error
}

func GetRootLogger() *RootLogger {
	if Current != nil {
		return Current.GetRootLogger()
	} else {
		return DefaultRootLogger()
	}
}

func LogWithCaller() *log.Entry {
	return GetRootLogger().withCaller()
}

func LogWithContext(ctx context.Context) *log.Entry {
	id := GetTraceIDFromContext(ctx)
	return GetRootLogger().withCaller().WithField(TiEM_X_TRACE_ID_NAME, id)
}

type Opt func(d *BaseFramework) error
type ServiceHandler func(service micro.Service) error
type ClientHandler func(service micro.Service) error

type BaseFramework struct {
	args          *ClientArgs
	configuration *Configuration
	log           *RootLogger
	trace         *Tracer
	etcdClient    *EtcdClient
	certificate   *CertificateInfo

	serviceMeta  *ServiceMeta
	microService micro.Service

	initOpts     []Opt
	shutdownOpts []Opt

	clientHandler  map[ServiceNameEnum]ClientHandler
	serviceHandler ServiceHandler
}

func InitBaseFrameworkForUt(serviceName ServiceNameEnum, opts ...Opt) *BaseFramework {
	f := new(BaseFramework)

	f.args = &ClientArgs{
		Host:               "127.0.0.1",
		Port:               4116,
		MetricsPort:        4121,
		RegistryClientPort: 4101,
		RegistryPeerPort:   4102,
		RegistryAddress:    "127.0.0.1:4101",
		DeployDir:          "./../bin",
		DataDir:            "./testdata",
		LogLevel:           "info",
	}
	f.parseArgs(serviceName)

	f.initOpts = opts
	f.Init()

	f.shutdownOpts = []Opt{
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

	f.initEtcdClient()
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
		srv := micro.NewService(
			micro.Name(string(client)),
			micro.WrapHandler(prometheus.NewHandlerWrapper()),
			micro.WrapHandler(opentracing.NewHandlerWrapper(*b.trace)),
			micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(b.loadCert()))),
			micro.Registry(etcd.NewRegistry(registry.Addrs(b.GetServiceMeta().RegistryAddress...))),
			micro.WrapClient(opentracing.NewClientWrapper(*b.trace)),
		)
		srv.Init()

		handler(srv)
	}
}

func (b *BaseFramework) loadCert() *tls.Config {
	cert, err := tls.LoadX509KeyPair(b.certificate.CertificateCrtFilePath, b.certificate.CertificateKeyFilePath)
	if err != nil {
		panic("load certificate file failed")
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
}

func (b *BaseFramework) initMicroService() {
	server := server.NewServer(
		server.Name(string(b.serviceMeta.ServiceName)),
		server.WrapHandler(prometheus.NewHandlerWrapper()),
		server.WrapHandler(opentracing.NewHandlerWrapper(*b.trace)),
		server.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(b.loadCert()))),
		server.Address(b.serviceMeta.GetServiceAddress()),
		server.Registry(etcd.NewRegistry(registry.Addrs(b.serviceMeta.RegistryAddress...))),
	)

	srv := micro.NewService(
		micro.Server(server),
		micro.WrapClient(opentracing.NewClientWrapper(*b.trace)),
	)
	srv.Init()

	// listen prometheus metrics
	go b.prometheusBoot()

	b.microService = srv

	b.serviceHandler(b.microService)
}

func (b *BaseFramework) GetDataDir() string {
	return b.args.DataDir
}

func (b *BaseFramework) initEtcdClient() {
	b.etcdClient = InitEtcdClient(b.GetServiceMeta().RegistryAddress)
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

func (b *BaseFramework) LogWithContext(ctx context.Context) *log.Entry {
	id := GetTraceIDFromContext(ctx)
	return b.GetRootLogger().withCaller().WithField(TiEM_X_TRACE_ID_NAME, id)
}

func (b *BaseFramework) GetTracer() *Tracer {
	return b.trace
}

func (b *BaseFramework) GetEtcdClient() *EtcdClient {
	return b.etcdClient
}

func (b *BaseFramework) GetServiceMeta() *ServiceMeta {
	return b.serviceMeta
}

func (b *BaseFramework) StopService() error {
	return nil
}

func (b *BaseFramework) StartService() error {
	if err := b.microService.Run(); err != nil {
		b.GetRootLogger().ForkFile(common.LogFileSystem).Fatalf("Initialization micro service failed, error %v, listening address %s, etcd registry address %s", err, b.serviceMeta.GetServiceAddress(), b.serviceMeta.RegistryAddress)
		return errors.New("initialization micro service failed")
	}

	return nil
}

func (b *BaseFramework) prometheusBoot() {
	// add boot_time metrics
	bootTime := prom.NewGaugeVec(
		prom.GaugeOpts{
			Namespace: common.TiEM,
			Name:      "boot_time",
			Help:      "A gauge of micro service boot time.",
		},
		[]string{"service"},
	)
	prom.MustRegister(bootTime)
	bootTime.With(prom.Labels{"service": b.GetServiceMeta().ServiceName.ServerName()}).SetToCurrentTime()

	http.Handle("/metrics", promhttp.Handler())
	// 启动web服务，监听8085端口
	go func() {
		metricsPort := b.GetClientArgs().MetricsPort
		if metricsPort <= 0 {
			metricsPort = common.DefaultMetricsPort
		}
		LogWithCaller().Infof("prometheus listen address [0.0.0.0:%d]", metricsPort)
		err := http.ListenAndServe(common.LocalAddress+":"+strconv.Itoa(metricsPort), nil)
		if err != nil {
			LogWithCaller().Errorf("prometheus listen and serve error: %v", err)
			panic("ListenAndServe: " + err.Error())
		}
	}()
}
