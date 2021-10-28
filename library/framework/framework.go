/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

package framework

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/pingcap-inc/tiem/library/thirdparty/metrics"
	prom "github.com/prometheus/client_golang/prometheus"

	"github.com/pingcap-inc/tiem/library/common"

	"github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/server"
	"github.com/asim/go-micro/v3/transport"
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
	Log() *log.Entry
	LogWithContext(context.Context) *log.Entry
	GetTracer() *Tracer
	GetEtcdClient() *EtcdClient
	GetElasticsearchClient() *ElasticSearchClient
	GetMetrics() *metrics.Metrics

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

func Log() *log.Entry {
	return GetRootLogger().defaultLogEntry
}

func LogWithContext(ctx context.Context) *log.Entry {
	id := GetTraceIDFromContext(ctx)
	return GetRootLogger().defaultLogEntry.WithField(TiEM_X_TRACE_ID_NAME, id)
}

func LogForkFile(fileName string) *log.Entry {
	return GetRootLogger().ForkFile(fileName)
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

	elasticsearchClient *ElasticSearchClient

	metrics *metrics.Metrics

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

	f.metrics = metrics.InitMetricsForUT()
	f.serviceMeta = NewServiceMetaFromArgs(serviceName, f.args)
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
	f.initElasticsearchClient()
	f.initMetrics()
	// listen prometheus metrics
	go f.prometheusBoot()
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

	b.microService = srv

	b.serviceHandler(b.microService)
}

func (b *BaseFramework) GetDataDir() string {
	return b.args.DataDir
}

func (b *BaseFramework) initEtcdClient() {
	b.etcdClient = InitEtcdClient(b.GetServiceMeta().RegistryAddress)
}

func (b *BaseFramework) initElasticsearchClient() {
	b.elasticsearchClient = InitElasticsearch(b.GetClientArgs().ElasticsearchAddress)
}

func (b *BaseFramework) initMetrics() {
	b.metrics = metrics.InitMetrics()
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

func (b *BaseFramework) Log() *log.Entry {
	return b.GetRootLogger().defaultLogEntry
}

func (b *BaseFramework) LogWithContext(ctx context.Context) *log.Entry {
	id := GetTraceIDFromContext(ctx)
	return b.Log().WithField(TiEM_X_TRACE_ID_NAME, id)
}

func (b *BaseFramework) GetTracer() *Tracer {
	return b.trace
}

func (b *BaseFramework) GetCertificateInfo() *CertificateInfo {
	return b.certificate
}

func (b *BaseFramework) GetEtcdClient() *EtcdClient {
	return b.etcdClient
}

func (b *BaseFramework) GetElasticsearchClient() *ElasticSearchClient {
	return b.elasticsearchClient
}

func (b *BaseFramework) GetMetrics() *metrics.Metrics {
	return b.metrics
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
	b.metrics.BootTimeGaugeMetric.
		With(prom.Labels{metrics.ServiceLabel: b.GetServiceMeta().ServiceName.ServerName()}).
		SetToCurrentTime()

	http.Handle("/metrics", promhttp.Handler())
	// 启动web服务，监听8085端口
	go func() {
		metricsPort := b.GetClientArgs().MetricsPort
		if metricsPort <= 0 {
			metricsPort = common.DefaultMetricsPort
		}
		Log().Infof("prometheus listen address [0.0.0.0:%d]", metricsPort)
		err := http.ListenAndServe(common.LocalAddress+":"+strconv.Itoa(metricsPort), nil)
		if err != nil {
			Log().Errorf("prometheus listen and serve error: %v", err)
			panic("ListenAndServe: " + err.Error())
		}
	}()
}
