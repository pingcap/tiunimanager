package models

import (
	"reflect"
	"testing"
)

func TestAddTenant(t *testing.T) {
	type args struct {
		name       string
		tenantType int8
		status     int8
	}
	tests := []struct {
		name       string
		args       args
		wantTenant Tenant
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTenant, err := AddTenant(tt.args.name, tt.args.tenantType, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddTenant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTenant, tt.wantTenant) {
				t.Errorf("AddTenant() gotTenant = %v, want %v", gotTenant, tt.wantTenant)
			}
		})
	}
}

func TestFindTenantById(t *testing.T) {
	type args struct {
		tenantId string
	}
	tests := []struct {
		name       string
		args       args
		wantTenant Tenant
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTenant, err := FindTenantById(tt.args.tenantId)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindTenantById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTenant, tt.wantTenant) {
				t.Errorf("FindTenantById() gotTenant = %v, want %v", gotTenant, tt.wantTenant)
			}
		})
	}
}

func TestFindTenantByName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name       string
		args       args
		wantTenant Tenant
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTenant, err := FindTenantByName(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindTenantByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTenant, tt.wantTenant) {
				t.Errorf("FindTenantByName() gotTenant = %v, want %v", gotTenant, tt.wantTenant)
			}
		})
	}
}

func TestTenant_BeforeCreate(t *testing.T) {

		t.Run("normal", func(t *testing.T) {
			tenant := Tenant{
				Name: "name_test_tenant",
			}
			err := tenant.BeforeCreate(nil)
			if err != nil  {
				t.Errorf("BeforeCreate() error = %v, wantErr nil", err)
			}

			if tenant.Status != 0 {
				t.Errorf("BeforeCreate() error, want status %v, got %v", 0, tenant.Status)
			}

			if tenant.ID == "" || len(tenant.ID) != 18 {
				t.Errorf("BeforeCreate() error, want id length %v, got %v", 18, len(tenant.ID))
			}

			if tenant.Name != "name_test_tenant" {
				t.Errorf("BeforeCreate() want name = %v", "name_test_tenant")
			}

		})
}
