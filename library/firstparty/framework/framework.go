package framework

import (
	"fmt"

	"github.com/asim/go-micro/v3/server"
)

type Framework interface {
	Run() error
}

type framework struct {
	initFp func() error

	certificateCrtFilePath string
	certificateKeyFilePath string
	serviceName            string
	serviceListenAddr      string
	registryAddrs          []string

	registerServiceHandlerFp func(server server.Server, opts ...server.HandlerOption) error
}

func WithInitFp(initFp func() error) func(*framework) {
	return func(p *framework) {
		p.initFp = initFp
	}
}

func WithCertificateCrtFilePath(path string) func(*framework) {
	return func(p *framework) {
		p.certificateCrtFilePath = path
	}
}

func WithCertificateKeyFilePath(path string) func(*framework) {
	return func(p *framework) {
		p.certificateKeyFilePath = path
	}
}

func WithServiceName(serviceName string) func(*framework) {
	return func(p *framework) {
		p.serviceName = serviceName
	}
}

func WithServiceListenAddr(addr string) func(*framework) {
	return func(p *framework) {
		p.serviceListenAddr = addr
	}
}
func WithRegistryAddrs(registryAddrs []string) func(*framework) {
	return func(p *framework) {
		p.registryAddrs = registryAddrs
	}
}

func WithRegisterServiceHandlerFp(
	fp func(server server.Server, opts ...server.HandlerOption) error) func(*framework) {

	return func(p *framework) {
		p.registerServiceHandlerFp = fp
	}
}

func (p *framework) Run() error {
	if p.initFp == nil {
	} else {
		err := p.initFp()
		if err != nil {
			return fmt.Errorf("meet an error when initFp is executing: %s", err)
		}
	}
	fmt.Println(p)
	// TODO: setup srv
	return nil
}

// there is a rather rough usage example in the test
func NewFramework(opts ...func(*framework)) Framework {
	p := &framework{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}
