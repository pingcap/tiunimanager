package framework

import (
	"github.com/pingcap-inc/tiem/library/firstparty/config"
	"github.com/pingcap-inc/tiem/library/firstparty/util"
	"github.com/pingcap-inc/tiem/library/thirdparty/logger"
)

type ItOpt func(d *ItFramework) error

type ItFramework struct {
	serviceEnum 		MicroServiceEnum

	initOpts 			[]ItOpt
	log 				*logger.LogRecord
}

func NewItFramework(serviceName MicroServiceEnum, initOpt ...ItOpt) *ItFramework {
	p := &ItFramework{
		serviceEnum: serviceName,
		initOpts: []ItOpt{
			func(d *ItFramework) error {
				config.InitConfigForDev(d.serviceEnum.logMod())
				return nil
			},
			func(d *ItFramework) error {
				d.log = d.serviceEnum.buildLogger()
				return nil
			},
		},
	}


	p.initOpts = append(p.initOpts, initOpt...)

	p.Init()

	f = p
	return p
}

func (i ItFramework) Init() error {
	for _, opt := range i.initOpts {
		util.AssertNoErr(opt(&i))
	}
	return nil
}

func (i ItFramework) StartService() error {
	return nil
}

func (i ItFramework) GetDefaultLogger() *logger.LogRecord {
	return i.log
}

func (i ItFramework) GetRegistryAddress() []string {
	panic("implement me")
}
