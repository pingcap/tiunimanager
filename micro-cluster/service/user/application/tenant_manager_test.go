package application

import (
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/domain"
	"reflect"
	"testing"
)

func TestCreateTenant(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *domain.Tenant
		wantErr bool
	}{
		{"normal", args{"notExisted"}, &domain.Tenant{Name: "notExisted", Type: domain.InstanceWorkspace}, false},
		{"existed", args{"existed"}, &domain.Tenant{Name: "existed"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tenantManager.CreateTenant(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTenant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateTenant() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindTenant(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *domain.Tenant
		wantErr bool
	}{
		{"normal", args{name: "111"}, &domain.Tenant{Name: "111"}, false},
		{"notExisted", args{name: "notExisted"}, &domain.Tenant{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tenantManager.FindTenant(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindTenant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindTenant() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindTenantById(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    *domain.Tenant
		wantErr bool
	}{
		{"normal", args{id: "111"}, &domain.Tenant{Id: "111"}, false},
		{"notExisted", args{id: "notExisted"}, &domain.Tenant{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tenantManager.FindTenantById(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindTenantById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindTenantById() got = %v, want %v", got, tt.want)
			}
		})
	}
}
