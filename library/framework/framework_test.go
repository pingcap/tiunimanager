package framework

import (
	"github.com/asim/go-micro/v3"
	"reflect"
	"testing"
)

func TestBaseFramework_GetClientArgs(t *testing.T) {
	type fields struct {
		args           *ClientArgs
		configuration  *Configuration
		log            *LogRecord
		trace          *Tracer
		certificate    *CertificateInfo
		serviceMeta    *ServiceMeta
		microService   micro.Service
		initOpts       []Opt
		shutdownOpts   []Opt
		clientHandler  map[ServiceNameEnum]ClientHandler
		serviceHandler ServiceHandler
	}
	tests := []struct {
		name   string
		fields fields
		want   *ClientArgs
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BaseFramework{
				args:           tt.fields.args,
				configuration:  tt.fields.configuration,
				log:            tt.fields.log,
				trace:          tt.fields.trace,
				certificate:    tt.fields.certificate,
				serviceMeta:    tt.fields.serviceMeta,
				microService:   tt.fields.microService,
				initOpts:       tt.fields.initOpts,
				shutdownOpts:   tt.fields.shutdownOpts,
				clientHandler:  tt.fields.clientHandler,
				serviceHandler: tt.fields.serviceHandler,
			}
			if got := b.GetClientArgs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetClientArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseFramework_GetConfiguration(t *testing.T) {
	type fields struct {
		args           *ClientArgs
		configuration  *Configuration
		log            *LogRecord
		trace          *Tracer
		certificate    *CertificateInfo
		serviceMeta    *ServiceMeta
		microService   micro.Service
		initOpts       []Opt
		shutdownOpts   []Opt
		clientHandler  map[ServiceNameEnum]ClientHandler
		serviceHandler ServiceHandler
	}
	tests := []struct {
		name   string
		fields fields
		want   *Configuration
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BaseFramework{
				args:           tt.fields.args,
				configuration:  tt.fields.configuration,
				log:            tt.fields.log,
				trace:          tt.fields.trace,
				certificate:    tt.fields.certificate,
				serviceMeta:    tt.fields.serviceMeta,
				microService:   tt.fields.microService,
				initOpts:       tt.fields.initOpts,
				shutdownOpts:   tt.fields.shutdownOpts,
				clientHandler:  tt.fields.clientHandler,
				serviceHandler: tt.fields.serviceHandler,
			}
			if got := b.GetConfiguration(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConfiguration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseFramework_GetDataDir(t *testing.T) {
	type fields struct {
		args           *ClientArgs
		configuration  *Configuration
		log            *LogRecord
		trace          *Tracer
		certificate    *CertificateInfo
		serviceMeta    *ServiceMeta
		microService   micro.Service
		initOpts       []Opt
		shutdownOpts   []Opt
		clientHandler  map[ServiceNameEnum]ClientHandler
		serviceHandler ServiceHandler
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BaseFramework{
				args:           tt.fields.args,
				configuration:  tt.fields.configuration,
				log:            tt.fields.log,
				trace:          tt.fields.trace,
				certificate:    tt.fields.certificate,
				serviceMeta:    tt.fields.serviceMeta,
				microService:   tt.fields.microService,
				initOpts:       tt.fields.initOpts,
				shutdownOpts:   tt.fields.shutdownOpts,
				clientHandler:  tt.fields.clientHandler,
				serviceHandler: tt.fields.serviceHandler,
			}
			if got := b.GetDataDir(); got != tt.want {
				t.Errorf("GetDataDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseFramework_GetDeployDir(t *testing.T) {
	type fields struct {
		args           *ClientArgs
		configuration  *Configuration
		log            *LogRecord
		trace          *Tracer
		certificate    *CertificateInfo
		serviceMeta    *ServiceMeta
		microService   micro.Service
		initOpts       []Opt
		shutdownOpts   []Opt
		clientHandler  map[ServiceNameEnum]ClientHandler
		serviceHandler ServiceHandler
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BaseFramework{
				args:           tt.fields.args,
				configuration:  tt.fields.configuration,
				log:            tt.fields.log,
				trace:          tt.fields.trace,
				certificate:    tt.fields.certificate,
				serviceMeta:    tt.fields.serviceMeta,
				microService:   tt.fields.microService,
				initOpts:       tt.fields.initOpts,
				shutdownOpts:   tt.fields.shutdownOpts,
				clientHandler:  tt.fields.clientHandler,
				serviceHandler: tt.fields.serviceHandler,
			}
			if got := b.GetDeployDir(); got != tt.want {
				t.Errorf("GetDeployDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseFramework_GetLogger(t *testing.T) {
	type fields struct {
		args           *ClientArgs
		configuration  *Configuration
		log            *LogRecord
		trace          *Tracer
		certificate    *CertificateInfo
		serviceMeta    *ServiceMeta
		microService   micro.Service
		initOpts       []Opt
		shutdownOpts   []Opt
		clientHandler  map[ServiceNameEnum]ClientHandler
		serviceHandler ServiceHandler
	}
	tests := []struct {
		name   string
		fields fields
		want   *LogRecord
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BaseFramework{
				args:           tt.fields.args,
				configuration:  tt.fields.configuration,
				log:            tt.fields.log,
				trace:          tt.fields.trace,
				certificate:    tt.fields.certificate,
				serviceMeta:    tt.fields.serviceMeta,
				microService:   tt.fields.microService,
				initOpts:       tt.fields.initOpts,
				shutdownOpts:   tt.fields.shutdownOpts,
				clientHandler:  tt.fields.clientHandler,
				serviceHandler: tt.fields.serviceHandler,
			}
			if got := b.GetLogger(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLogger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseFramework_GetServiceMeta(t *testing.T) {
	type fields struct {
		args           *ClientArgs
		configuration  *Configuration
		log            *LogRecord
		trace          *Tracer
		certificate    *CertificateInfo
		serviceMeta    *ServiceMeta
		microService   micro.Service
		initOpts       []Opt
		shutdownOpts   []Opt
		clientHandler  map[ServiceNameEnum]ClientHandler
		serviceHandler ServiceHandler
	}
	tests := []struct {
		name   string
		fields fields
		want   *ServiceMeta
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BaseFramework{
				args:           tt.fields.args,
				configuration:  tt.fields.configuration,
				log:            tt.fields.log,
				trace:          tt.fields.trace,
				certificate:    tt.fields.certificate,
				serviceMeta:    tt.fields.serviceMeta,
				microService:   tt.fields.microService,
				initOpts:       tt.fields.initOpts,
				shutdownOpts:   tt.fields.shutdownOpts,
				clientHandler:  tt.fields.clientHandler,
				serviceHandler: tt.fields.serviceHandler,
			}
			if got := b.GetServiceMeta(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetServiceMeta() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseFramework_GetTracer(t *testing.T) {
	type fields struct {
		args           *ClientArgs
		configuration  *Configuration
		log            *LogRecord
		trace          *Tracer
		certificate    *CertificateInfo
		serviceMeta    *ServiceMeta
		microService   micro.Service
		initOpts       []Opt
		shutdownOpts   []Opt
		clientHandler  map[ServiceNameEnum]ClientHandler
		serviceHandler ServiceHandler
	}
	tests := []struct {
		name   string
		fields fields
		want   *Tracer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BaseFramework{
				args:           tt.fields.args,
				configuration:  tt.fields.configuration,
				log:            tt.fields.log,
				trace:          tt.fields.trace,
				certificate:    tt.fields.certificate,
				serviceMeta:    tt.fields.serviceMeta,
				microService:   tt.fields.microService,
				initOpts:       tt.fields.initOpts,
				shutdownOpts:   tt.fields.shutdownOpts,
				clientHandler:  tt.fields.clientHandler,
				serviceHandler: tt.fields.serviceHandler,
			}
			if got := b.GetTracer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTracer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseFramework_Init(t *testing.T) {
	type fields struct {
		args           *ClientArgs
		configuration  *Configuration
		log            *LogRecord
		trace          *Tracer
		certificate    *CertificateInfo
		serviceMeta    *ServiceMeta
		microService   micro.Service
		initOpts       []Opt
		shutdownOpts   []Opt
		clientHandler  map[ServiceNameEnum]ClientHandler
		serviceHandler ServiceHandler
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
			b := &BaseFramework{
				args:           tt.fields.args,
				configuration:  tt.fields.configuration,
				log:            tt.fields.log,
				trace:          tt.fields.trace,
				certificate:    tt.fields.certificate,
				serviceMeta:    tt.fields.serviceMeta,
				microService:   tt.fields.microService,
				initOpts:       tt.fields.initOpts,
				shutdownOpts:   tt.fields.shutdownOpts,
				clientHandler:  tt.fields.clientHandler,
				serviceHandler: tt.fields.serviceHandler,
			}
			if err := b.Init(); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBaseFramework_PrepareClientClient(t *testing.T) {
	type fields struct {
		args           *ClientArgs
		configuration  *Configuration
		log            *LogRecord
		trace          *Tracer
		certificate    *CertificateInfo
		serviceMeta    *ServiceMeta
		microService   micro.Service
		initOpts       []Opt
		shutdownOpts   []Opt
		clientHandler  map[ServiceNameEnum]ClientHandler
		serviceHandler ServiceHandler
	}
	type args struct {
		clientHandlerMap map[ServiceNameEnum]ClientHandler
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BaseFramework{
				args:           tt.fields.args,
				configuration:  tt.fields.configuration,
				log:            tt.fields.log,
				trace:          tt.fields.trace,
				certificate:    tt.fields.certificate,
				serviceMeta:    tt.fields.serviceMeta,
				microService:   tt.fields.microService,
				initOpts:       tt.fields.initOpts,
				shutdownOpts:   tt.fields.shutdownOpts,
				clientHandler:  tt.fields.clientHandler,
				serviceHandler: tt.fields.serviceHandler,
			}
			b.Shutdown()
		})
	}
}

func TestBaseFramework_PrepareService(t *testing.T) {
	type fields struct {
		args           *ClientArgs
		configuration  *Configuration
		log            *LogRecord
		trace          *Tracer
		certificate    *CertificateInfo
		serviceMeta    *ServiceMeta
		microService   micro.Service
		initOpts       []Opt
		shutdownOpts   []Opt
		clientHandler  map[ServiceNameEnum]ClientHandler
		serviceHandler ServiceHandler
	}
	type args struct {
		handler ServiceHandler
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BaseFramework{
				args:           tt.fields.args,
				configuration:  tt.fields.configuration,
				log:            tt.fields.log,
				trace:          tt.fields.trace,
				certificate:    tt.fields.certificate,
				serviceMeta:    tt.fields.serviceMeta,
				microService:   tt.fields.microService,
				initOpts:       tt.fields.initOpts,
				shutdownOpts:   tt.fields.shutdownOpts,
				clientHandler:  tt.fields.clientHandler,
				serviceHandler: tt.fields.serviceHandler,
			}
			b.Shutdown()

		})
	}
}

func TestBaseFramework_Shutdown(t *testing.T) {
	type fields struct {
		args           *ClientArgs
		configuration  *Configuration
		log            *LogRecord
		trace          *Tracer
		certificate    *CertificateInfo
		serviceMeta    *ServiceMeta
		microService   micro.Service
		initOpts       []Opt
		shutdownOpts   []Opt
		clientHandler  map[ServiceNameEnum]ClientHandler
		serviceHandler ServiceHandler
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
			b := &BaseFramework{
				args:           tt.fields.args,
				configuration:  tt.fields.configuration,
				log:            tt.fields.log,
				trace:          tt.fields.trace,
				certificate:    tt.fields.certificate,
				serviceMeta:    tt.fields.serviceMeta,
				microService:   tt.fields.microService,
				initOpts:       tt.fields.initOpts,
				shutdownOpts:   tt.fields.shutdownOpts,
				clientHandler:  tt.fields.clientHandler,
				serviceHandler: tt.fields.serviceHandler,
			}
			if err := b.Shutdown(); (err != nil) != tt.wantErr {
				t.Errorf("Shutdown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBaseFramework_StartService(t *testing.T) {
	type fields struct {
		args           *ClientArgs
		configuration  *Configuration
		log            *LogRecord
		trace          *Tracer
		certificate    *CertificateInfo
		serviceMeta    *ServiceMeta
		microService   micro.Service
		initOpts       []Opt
		shutdownOpts   []Opt
		clientHandler  map[ServiceNameEnum]ClientHandler
		serviceHandler ServiceHandler
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
			b := &BaseFramework{
				args:           tt.fields.args,
				configuration:  tt.fields.configuration,
				log:            tt.fields.log,
				trace:          tt.fields.trace,
				certificate:    tt.fields.certificate,
				serviceMeta:    tt.fields.serviceMeta,
				microService:   tt.fields.microService,
				initOpts:       tt.fields.initOpts,
				shutdownOpts:   tt.fields.shutdownOpts,
				clientHandler:  tt.fields.clientHandler,
				serviceHandler: tt.fields.serviceHandler,
			}
			if err := b.StartService(); (err != nil) != tt.wantErr {
				t.Errorf("StartService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBaseFramework_StopService(t *testing.T) {
	type fields struct {
		args           *ClientArgs
		configuration  *Configuration
		log            *LogRecord
		trace          *Tracer
		certificate    *CertificateInfo
		serviceMeta    *ServiceMeta
		microService   micro.Service
		initOpts       []Opt
		shutdownOpts   []Opt
		clientHandler  map[ServiceNameEnum]ClientHandler
		serviceHandler ServiceHandler
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
			b := &BaseFramework{
				args:           tt.fields.args,
				configuration:  tt.fields.configuration,
				log:            tt.fields.log,
				trace:          tt.fields.trace,
				certificate:    tt.fields.certificate,
				serviceMeta:    tt.fields.serviceMeta,
				microService:   tt.fields.microService,
				initOpts:       tt.fields.initOpts,
				shutdownOpts:   tt.fields.shutdownOpts,
				clientHandler:  tt.fields.clientHandler,
				serviceHandler: tt.fields.serviceHandler,
			}
			if err := b.StopService(); (err != nil) != tt.wantErr {
				t.Errorf("StopService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBaseFramework_acceptArgs(t *testing.T) {
	type fields struct {
		args           *ClientArgs
		configuration  *Configuration
		log            *LogRecord
		trace          *Tracer
		certificate    *CertificateInfo
		serviceMeta    *ServiceMeta
		microService   micro.Service
		initOpts       []Opt
		shutdownOpts   []Opt
		clientHandler  map[ServiceNameEnum]ClientHandler
		serviceHandler ServiceHandler
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BaseFramework{
				args:           tt.fields.args,
				configuration:  tt.fields.configuration,
				log:            tt.fields.log,
				trace:          tt.fields.trace,
				certificate:    tt.fields.certificate,
				serviceMeta:    tt.fields.serviceMeta,
				microService:   tt.fields.microService,
				initOpts:       tt.fields.initOpts,
				shutdownOpts:   tt.fields.shutdownOpts,
				clientHandler:  tt.fields.clientHandler,
				serviceHandler: tt.fields.serviceHandler,
			}
			b.Shutdown()

		})
	}
}

func TestBaseFramework_initMicroClient(t *testing.T) {
	type fields struct {
		args           *ClientArgs
		configuration  *Configuration
		log            *LogRecord
		trace          *Tracer
		certificate    *CertificateInfo
		serviceMeta    *ServiceMeta
		microService   micro.Service
		initOpts       []Opt
		shutdownOpts   []Opt
		clientHandler  map[ServiceNameEnum]ClientHandler
		serviceHandler ServiceHandler
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BaseFramework{
				args:           tt.fields.args,
				configuration:  tt.fields.configuration,
				log:            tt.fields.log,
				trace:          tt.fields.trace,
				certificate:    tt.fields.certificate,
				serviceMeta:    tt.fields.serviceMeta,
				microService:   tt.fields.microService,
				initOpts:       tt.fields.initOpts,
				shutdownOpts:   tt.fields.shutdownOpts,
				clientHandler:  tt.fields.clientHandler,
				serviceHandler: tt.fields.serviceHandler,
			}
			b.Shutdown()

		})
	}
}

func TestBaseFramework_initMicroService(t *testing.T) {
	type fields struct {
		args           *ClientArgs
		configuration  *Configuration
		log            *LogRecord
		trace          *Tracer
		certificate    *CertificateInfo
		serviceMeta    *ServiceMeta
		microService   micro.Service
		initOpts       []Opt
		shutdownOpts   []Opt
		clientHandler  map[ServiceNameEnum]ClientHandler
		serviceHandler ServiceHandler
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BaseFramework{
				args:           tt.fields.args,
				configuration:  tt.fields.configuration,
				log:            tt.fields.log,
				trace:          tt.fields.trace,
				certificate:    tt.fields.certificate,
				serviceMeta:    tt.fields.serviceMeta,
				microService:   tt.fields.microService,
				initOpts:       tt.fields.initOpts,
				shutdownOpts:   tt.fields.shutdownOpts,
				clientHandler:  tt.fields.clientHandler,
				serviceHandler: tt.fields.serviceHandler,
			}
			b.Shutdown()

		})
	}
}

func TestBaseFramework_parseArgs(t *testing.T) {
	type fields struct {
		args           *ClientArgs
		configuration  *Configuration
		log            *LogRecord
		trace          *Tracer
		certificate    *CertificateInfo
		serviceMeta    *ServiceMeta
		microService   micro.Service
		initOpts       []Opt
		shutdownOpts   []Opt
		clientHandler  map[ServiceNameEnum]ClientHandler
		serviceHandler ServiceHandler
	}
	type args struct {
		serviceName ServiceNameEnum
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BaseFramework{
				args:           tt.fields.args,
				configuration:  tt.fields.configuration,
				log:            tt.fields.log,
				trace:          tt.fields.trace,
				certificate:    tt.fields.certificate,
				serviceMeta:    tt.fields.serviceMeta,
				microService:   tt.fields.microService,
				initOpts:       tt.fields.initOpts,
				shutdownOpts:   tt.fields.shutdownOpts,
				clientHandler:  tt.fields.clientHandler,
				serviceHandler: tt.fields.serviceHandler,
			}
			b.Shutdown()

		})
	}
}

func TestGetLogger(t *testing.T) {
	tests := []struct {
		name string
		want *LogRecord
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetLogger(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLogger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitBaseFrameworkForUt(t *testing.T) {
	type args struct {
		serviceName ServiceNameEnum
		opts        []Opt
	}
	tests := []struct {
		name string
		args args
		want *BaseFramework
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitBaseFrameworkForUt(tt.args.serviceName, tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitBaseFrameworkForUt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitBaseFrameworkFromArgs(t *testing.T) {
	type args struct {
		serviceName ServiceNameEnum
		opts        []Opt
	}
	tests := []struct {
		name string
		args args
		want *BaseFramework
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InitBaseFrameworkFromArgs(tt.args.serviceName, tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitBaseFrameworkFromArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
