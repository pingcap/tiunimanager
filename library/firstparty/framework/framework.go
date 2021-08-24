package framework

import (
	"errors"
	"github.com/asim/go-micro/v3"
	"github.com/pingcap-inc/tiem/library/firstparty/config"
	"github.com/pingcap-inc/tiem/library/firstparty/util"
	mysignal "github.com/pingcap-inc/tiem/library/firstparty/util/signal"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/library/thirdparty/logger"
	"github.com/pingcap-inc/tiem/library/thirdparty/tracer"
	clusterPb "github.com/pingcap-inc/tiem/micro-cluster/proto"
	clusterSrv "github.com/pingcap-inc/tiem/micro-cluster/service"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
	dbSrv "github.com/pingcap-inc/tiem/micro-metadb/service"
)

type Framework interface {
	Init() error

	StartService() error

	GetDefaultLogger() *logger.LogRecord
	GetRegistryAddress() []string

	// registry center operator
	// config center operator
}

var f Framework

func GetDefaultLogger() *logger.LogRecord {
	return f.GetDefaultLogger()
}

type Opt func(d *DefaultServiceFramework) error

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
}

func NewDefaultFramework(serviceName MicroServiceEnum, initOpt ...Opt) *DefaultServiceFramework {
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
}

func initConfig(p *DefaultServiceFramework) error {
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
}
