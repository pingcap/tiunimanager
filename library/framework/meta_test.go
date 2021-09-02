package framework

import (
	"reflect"
	"testing"
)

func TestNewServiceMetaFromArgs(t *testing.T) {
	type args struct {
		serviceName ServiceNameEnum
		args        *ClientArgs
	}
	tests := []struct {
		name string
		args args
		want *ServiceMeta
	}{
		{"normal", args{"TestNewServiceMetaFromArgs", &ClientArgs{
			Port: 111,
			Host: "127.0.0.1",
			RegistryAddress: "11111,22222",
		}}, &ServiceMeta{
			ServiceName: "TestNewServiceMetaFromArgs",
			ServicePort: 111,
			ServiceHost: "127.0.0.1",
			RegistryAddress: []string{"11111", "22222"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewServiceMetaFromArgs(tt.args.serviceName, tt.args.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServiceMetaFromArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServiceMeta_GetServiceAddress(t *testing.T) {
	type fields struct {
		ServiceName     ServiceNameEnum
		RegistryAddress []string
		ServiceHost     string
		ServicePort     int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"normal", fields{ServicePort: 999}, ":999"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ServiceMeta{
				ServiceName:     tt.fields.ServiceName,
				RegistryAddress: tt.fields.RegistryAddress,
				ServiceHost:     tt.fields.ServiceHost,
				ServicePort:     tt.fields.ServicePort,
			}
			if got := s.GetServiceAddress(); got != tt.want {
				t.Errorf("GetServiceAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServiceNameEnum_ServerName(t *testing.T) {
	tests := []struct {
		name string
		s    ServiceNameEnum
		want string
	}{
		{"MetaDBService", MetaDBService, "metadb-server"},
		{"ClusterService", ClusterService, "cluster-server"},
		{"ApiService", ApiService, "openapi-server"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.ServerName(); got != tt.want {
				t.Errorf("ServerName() = %v, want %v", got, tt.want)
			}
		})
	}
}

