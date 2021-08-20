package framework

import (
	"fmt"

	"github.com/micro/cli/v2"

	"github.com/asim/go-micro/v3/server"
	"github.com/pingcap/tiem/library/firstparty/util"
	mysignal "github.com/pingcap/tiem/library/firstparty/util/signal"
)

type Opt func(*framework)

type Framework interface {
	InitLogger() error
	InitConfig() error
	InitClient() error
	ArgsParse() error
	SetupQuitSignalHandler(quitSignalHandler func()) error

	MustGetLogger() Logger
	MustGetConfig() Config
	MustGetClient() Client

	Run(opts ...Opt) error
	AddOpts(opts ...Opt)
}

type framework struct {
	initLoggerFp func() (Logger, error)
	initConfigFp func() (Config, error)
	initClientFp func() (Client, error)

	argFlags          []cli.Flag
	quitSignalHandler func()

	certificateCrtFilePath string
	certificateKeyFilePath string
	serviceName            string
	serviceListenAddr      string
	registryAddrs          []string

	registerServiceHandlerFp func(server server.Server, opts ...server.HandlerOption) error

	config Config
	logger Logger
	client Client
}

func WithCertificateCrtFilePath(path string) Opt {
	return func(p *framework) {
		p.certificateCrtFilePath = path
	}
}

func WithCertificateKeyFilePath(path string) Opt {
	return func(p *framework) {
		p.certificateKeyFilePath = path
	}
}

func WithServiceName(serviceName string) Opt {
	return func(p *framework) {
		p.serviceName = serviceName
	}
}

func WithArgFlags(argFlags []cli.Flag) Opt {
	return func(p *framework) {
		p.argFlags = argFlags
	}
}

func WithQuitSignalHandler(quitSignalHandler func()) Opt {
	return func(p *framework) {
		p.quitSignalHandler = quitSignalHandler
	}
}

func WithServiceListenAddr(addr string) Opt {
	return func(p *framework) {
		p.serviceListenAddr = addr
	}
}
func WithRegistryAddrs(registryAddrs []string) Opt {
	return func(p *framework) {
		p.registryAddrs = registryAddrs
	}
}

func WithRegisterServiceHandlerFp(
	fp func(server server.Server, opts ...server.HandlerOption) error) Opt {

	return func(p *framework) {
		p.registerServiceHandlerFp = fp
	}
}

func (p *framework) InitLogger() error {
	var err error
	p.logger, err = p.initLoggerFp()
	return err
}

func (p *framework) InitConfig() error {
	var err error
	p.config, err = p.initConfigFp()
	return err
}

func (p *framework) InitClient() error {
	var err error
	p.client, err = p.initClientFp()
	return err
}

func (p *framework) ArgsParse() error {
	argsParse(p.argFlags)
	return nil
}

func (p *framework) SetupQuitSignalHandler(quitSignalHandler func()) error {
	if quitSignalHandler == nil {
		return fmt.Errorf("quitSignalHandler should not be nil")
	}
	p.quitSignalHandler = quitSignalHandler
	mysignal.SetupSignalHandler(func(bool) {
		quitSignalHandler()
	})
	return nil
}

func (p *framework) MustGetLogger() Logger {
	ret := p.logger
	util.Assert(ret != nil)
	return ret
}

func (p *framework) MustGetConfig() Config {
	ret := p.config
	util.Assert(ret != nil)
	return ret
}

func (p *framework) MustGetClient() Client {
	ret := p.client
	util.Assert(ret != nil)
	return ret
}

func (p *framework) Run(opts ...Opt) error {
	fmt.Println(p)
	for _, opt := range opts {
		opt(p)
	}
	// TODO: setup srv
	return nil
}

func (p *framework) AddOpts(opts ...Opt) {
	for _, opt := range opts {
		opt(p)
	}
}

// there is a rather rough usage example in the test
func NewFramework(opts ...Opt) Framework {
	p := &framework{
		initLoggerFp: initLogger,
		initConfigFp: initConfig,
		initClientFp: initClient,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}
