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
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/metrics"

	prom "github.com/prometheus/client_golang/prometheus"

	"github.com/asim/go-micro/plugins/registry/etcd/v3"
	"github.com/asim/go-micro/plugins/wrapper/monitoring/prometheus/v3"
	"github.com/asim/go-micro/plugins/wrapper/trace/opentracing/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/server"
	"github.com/asim/go-micro/v3/transport"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	transport2 "go.etcd.io/etcd/client/pkg/v3/transport"
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
	GetEtcdClient() *EtcdClientV3
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
	return GetRootLogger().defaultLogEntry.WithField(TiEM_X_TRACE_ID_KEY, id)
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
	etcdClient    *EtcdClientV3
	certificate   *CertificateInfo

	elasticsearchClient *ElasticSearchClient

	serviceMeta  *ServiceMeta
	microService micro.Service
	metrics      *metrics.Metrics

	initOpts     []Opt
	shutdownOpts []Opt

	clientHandler  map[ServiceNameEnum]ClientHandler
	serviceHandler ServiceHandler
}

func InitBaseFrameworkForUt(serviceName ServiceNameEnum, opts ...Opt) *BaseFramework {
	f := new(BaseFramework)

	Current = f
	f.args = &ClientArgs{
		Host:                "127.0.0.1",
		Port:                4116,
		MetricsPort:         4121,
		RegistryClientPort:  4101,
		RegistryPeerPort:    4102,
		RegistryAddress:     "127.0.0.1:4101",
		DeployDir:           "./../bin",
		DataDir:             "./testdata",
		LogLevel:            "info",
		EMClusterName:       "em-test",
		EMVersion:           "InTesting",
		DeployUser:          "test-user",
		DeployGroup:         "test-group",
		LoginHostUser:       "root",
		LoginPrivateKeyPath: "/fake/private/key/path",
		LoginPublicKeyPath:  "/fake/public/key/path",
	}
	f.parseArgs(serviceName)

	f.serviceMeta = NewServiceMetaFromArgs(serviceName, f.args)
	f.initOpts = opts
	f.Init()

	f.shutdownOpts = []Opt{
		func(d *BaseFramework) error {
			return os.RemoveAll(d.GetDataDir())
		},
	}

	return f
}

func InitBaseFrameworkFromArgs(serviceName ServiceNameEnum, opts ...Opt) *BaseFramework {
	f := new(BaseFramework)
	Current = f

	f.acceptArgs()
	f.parseArgs(serviceName)
	f.setEtcdCertConfig()
	f.initOpts = opts
	f.Init()
	f.initEtcdClient()
	f.initElasticsearchClient()
	f.initMetrics()
	// listen prometheus metrics
	go f.prometheusBoot()
	return f
}

func (b *BaseFramework) setEtcdCertConfig() {
	EtcdCert = EtcdCertTransport{
		PeerTLSInfo:   b.genTlsInfo(constants.ETCDPeerCertFile),
		ClientTLSInfo: b.genTlsInfo(constants.ETCDClientCertFile),
		ServerTLSInfo: b.genTlsInfo(constants.ETCDServerCertFile),
	}
}

func (b *BaseFramework) genTlsInfo(certFilePath constants.EtcdCertFileType) transport2.TLSInfo {
	return transport2.TLSInfo{
		CertFile:            b.GetClientArgs().DeployDir + constants.CertDirPrefix + certFilePath[0],
		KeyFile:             b.GetClientArgs().DeployDir + constants.CertDirPrefix + certFilePath[1],
		TrustedCAFile:       b.GetClientArgs().DeployDir + constants.CertDirPrefix + constants.ETCDCAFileName,
		ClientCertAuth:      true,
		InsecureSkipVerify:  true,
		SkipClientSANVerify: true,
	}
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
	LocalConfig = make(map[Key]Instance)
}

func (b *BaseFramework) initMicroClient() {
	for client, handler := range b.clientHandler {
		srv := micro.NewService(
			micro.Name(string(client)),
			micro.WrapHandler(prometheus.NewHandlerWrapper()),
			micro.WrapHandler(opentracing.NewHandlerWrapper(*b.trace)),
			micro.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(b.loadCert("grpc")))),
			micro.Registry(etcd.NewRegistry(registry.Addrs(b.GetServiceMeta().RegistryAddress...),
				registry.Secure(true), registry.TLSConfig(b.loadCert("etcd")))),
			micro.WrapClient(opentracing.NewClientWrapper(*b.trace)),
		)
		srv.Init()

		handler(srv)
	}
}

func (b *BaseFramework) loadCert(crtType string) *tls.Config {
	var cert tls.Certificate
	var err error
	switch crtType {
	case "etcd":
		cert, err = tls.LoadX509KeyPair(b.GetDeployDir()+constants.CertDirPrefix+constants.ETCDClientCertFile[0],
			b.GetDeployDir()+constants.CertDirPrefix+constants.ETCDClientCertFile[1])
	default:
		cert, err = tls.LoadX509KeyPair(b.certificate.CertificateCrtFilePath, b.certificate.CertificateKeyFilePath)
	}

	if err != nil {
		panic("load " + crtType + " certificate file failed")
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true} // #nosec G402
}

func (b *BaseFramework) initMicroService() {
	server := server.NewServer(
		server.Name(string(b.serviceMeta.ServiceName)),
		server.WrapHandler(NewMicroHandlerWrapper()),
		server.WrapHandler(prometheus.NewHandlerWrapper()),
		server.WrapHandler(opentracing.NewHandlerWrapper(*b.trace)),
		server.Transport(transport.NewHTTPTransport(transport.Secure(true), transport.TLSConfig(b.loadCert("grpc")))),
		server.Address(b.serviceMeta.GetServiceAddress()),
		server.Registry(etcd.NewRegistry(registry.Addrs(b.serviceMeta.RegistryAddress...),
			registry.Secure(true), registry.TLSConfig(b.loadCert("etcd")))),
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
	b.metrics = metrics.GetMetrics()
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

func GetCurrentDeployUser() string {
	return Current.GetClientArgs().DeployUser
}

func GetCurrentDeployGroup() string {
	return Current.GetClientArgs().DeployGroup
}

func GetCurrentSpecifiedUser() string {
	return Current.GetClientArgs().LoginHostUser
}

func GetCurrentSpecifiedPrivateKeyPath() string {
	return Current.GetClientArgs().LoginPrivateKeyPath
}

func GetCurrentSpecifiedPublicKeyPath() string {
	return Current.GetClientArgs().LoginPublicKeyPath
}

func GetPrivateKeyFilePath(userName string) (keyPath string) {
	useSpecifiedKeyPair := GetBoolWithDefault(UsingSpecifiedKeyPair, false)
	if useSpecifiedKeyPair {
		keyPath = GetCurrentSpecifiedPrivateKeyPath()
	} else {
		// use default private key under deploy user ssh dir
		keyPath = fmt.Sprintf("/home/%s/.ssh/tiup_rsa", userName)
	}
	return
}

func GetPublicKeyFilePath(userName string) (keyPath string) {
	useSpecifiedKeyPair := GetBoolWithDefault(UsingSpecifiedKeyPair, false)
	if useSpecifiedKeyPair {
		keyPath = GetCurrentSpecifiedPublicKeyPath()
	} else {
		// use default public key under deploy user ssh dir
		keyPath = fmt.Sprintf("/home/%s/.ssh/id_rsa.pub", userName)
	}
	return
}

func GetTiupHomePathForTiem() string {
	userName := GetCurrentDeployUser()
	return fmt.Sprintf("/home/%s/.em", userName)
}

func GetTiupHomePathForTidb() string {
	userName := GetCurrentDeployUser()
	return fmt.Sprintf("/home/%s/.tiup", userName)
}

func GetTiupAuthorizaitonFlag() (flags []string) {
	userName := GetCurrentDeployUser()
	keyPath := GetPrivateKeyFilePath(userName)
	flags = append(flags, "--user")
	flags = append(flags, userName)
	flags = append(flags, "-i")
	flags = append(flags, keyPath)
	return
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
	return b.Log().WithField(TiEM_X_TRACE_ID_KEY, id)
}

func (b *BaseFramework) GetTracer() *Tracer {
	return b.trace
}

func (b *BaseFramework) GetCertificateInfo() *CertificateInfo {
	return b.certificate
}

func (b *BaseFramework) GetEtcdClient() *EtcdClientV3 {
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
		b.GetRootLogger().ForkFile(constants.LogFileSystem).Fatalf("Initialization micro service failed, error %v, listening address %s, etcd registry address %s", err, b.serviceMeta.GetServiceAddress(), b.serviceMeta.RegistryAddress)
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
	go func() {
		metricsPort := b.GetClientArgs().MetricsPort
		if metricsPort <= 0 {
			metricsPort = constants.DefaultMetricsPort
		}
		LogForkFile(constants.LogFileSystem).Infof("prometheus listen address [0.0.0.0:%d]", metricsPort)
		err := http.ListenAndServe("0.0.0.0:"+strconv.Itoa(metricsPort), nil)
		if err != nil {
			Log().Errorf("prometheus listen and serve error: %v", err)
			panic("ListenAndServe: " + err.Error())
		}
	}()
}