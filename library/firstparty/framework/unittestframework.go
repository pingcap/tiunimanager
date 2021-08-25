package framework

import (
	"github.com/pingcap-inc/tiem/library/firstparty/config"
	"github.com/pingcap-inc/tiem/library/firstparty/util"
	"github.com/pingcap-inc/tiem/library/thirdparty/logger"
	"github.com/pingcap-inc/tiem/library/thirdparty/tracer"
)

type UtOpt func(d *UtFramework) error

type UtFramework struct {
	serviceEnum MicroServiceEnum

	initOpts []UtOpt
	log      *logger.LogRecord
}

func NewUtFramework(serviceName MicroServiceEnum, initOpt ...UtOpt) *UtFramework {
	p := &UtFramework{
		serviceEnum: serviceName,
		initOpts: []UtOpt{
			func(d *UtFramework) error {
				config.InitConfigForDev(d.serviceEnum.logMod())
				return nil
			},
			func(d *UtFramework) error {
				d.log = d.serviceEnum.buildLogger()
				return nil
			},
			func(d *UtFramework) error {
				initUTTracer(d)
				return nil
			},
		},
	}

	p.initOpts = append(p.initOpts, initOpt...)

	p.Init()

	return p
}

func (u UtFramework) Init() error {
	for _, opt := range u.initOpts {
		util.AssertNoErr(opt(&u))
	}
	return nil
}

func (u UtFramework) StartService() error {
	return nil
}

func (u UtFramework) GetDefaultLogger() *logger.LogRecord {
	return u.log
}

func (u UtFramework) GetRegistryAddress() []string {
	panic("implement me")
}

func initUTTracer(u *UtFramework) error {
	tracer.InitTracer()
	return nil
}
