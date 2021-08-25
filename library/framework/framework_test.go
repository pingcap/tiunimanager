package framework

import (
	"github.com/asim/go-micro/v3"
	"github.com/pingcap-inc/tiem/library/thirdparty/logger"
	"reflect"
	"testing"
)

func TestDefaultServiceFramework_GetDefaultLogger(t *testing.T) {
	type fields struct {
		serviceEnum MicroServiceEnum
		flags       micro.Option
		initOpts    []Opt
		log         *logger.LogRecord
		service     micro.Service
	}
	tests := []struct {
		name   string
		fields fields
		want   *logger.LogRecord
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &DefaultServiceFramework{
				serviceEnum: tt.fields.serviceEnum,
				flags:       tt.fields.flags,
				initOpts:    tt.fields.initOpts,
				log:         tt.fields.log,
				service:     tt.fields.service,
			}
			if got := p.GetDefaultLogger(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDefaultLogger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultServiceFramework_GetRegistryAddress(t *testing.T) {
	type fields struct {
		serviceEnum MicroServiceEnum
		flags       micro.Option
		initOpts    []Opt
		log         *logger.LogRecord
		service     micro.Service
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &DefaultServiceFramework{
				serviceEnum: tt.fields.serviceEnum,
				flags:       tt.fields.flags,
				initOpts:    tt.fields.initOpts,
				log:         tt.fields.log,
				service:     tt.fields.service,
			}
			if got := p.GetRegistryAddress(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRegistryAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultServiceFramework_Init(t *testing.T) {
	type fields struct {
		serviceEnum MicroServiceEnum
		flags       micro.Option
		initOpts    []Opt
		log         *logger.LogRecord
		service     micro.Service
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &DefaultServiceFramework{
				serviceEnum: tt.fields.serviceEnum,
				flags:       tt.fields.flags,
				initOpts:    tt.fields.initOpts,
				log:         tt.fields.log,
				service:     tt.fields.service,
			}
			if err := p.Init(); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultServiceFramework_StartService(t *testing.T) {
	type fields struct {
		serviceEnum MicroServiceEnum
		flags       micro.Option
		initOpts    []Opt
		log         *logger.LogRecord
		service     micro.Service
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &DefaultServiceFramework{
				serviceEnum: tt.fields.serviceEnum,
				flags:       tt.fields.flags,
				initOpts:    tt.fields.initOpts,
				log:         tt.fields.log,
				service:     tt.fields.service,
			}
			if err := p.StartService(); (err != nil) != tt.wantErr {
				t.Errorf("StartService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewDefaultFramework(t *testing.T) {
	type args struct {
		serviceName MicroServiceEnum
		initOpt     []Opt
	}
	tests := []struct {
		name string
		args args
		want *DefaultServiceFramework
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDefaultFramework(tt.args.serviceName, tt.args.initOpt...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDefaultFramework() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_initConfig(t *testing.T) {
	type args struct {
		p *DefaultServiceFramework
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := initConfig(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("initConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_initCurrentLogger(t *testing.T) {
	type args struct {
		p *DefaultServiceFramework
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := initCurrentLogger(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("initCurrentLogger() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_initKnowledge(t *testing.T) {
	type args struct {
		p *DefaultServiceFramework
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := initKnowledge(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("initKnowledge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_initShutdownFunc(t *testing.T) {
	type args struct {
		p *DefaultServiceFramework
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := initShutdownFunc(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("initShutdownFunc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_initTracer(t *testing.T) {
	type args struct {
		p *DefaultServiceFramework
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := initTracer(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("initTracer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
