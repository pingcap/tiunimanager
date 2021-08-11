package models

import (
	"testing"
	"time"
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
		wantErr    bool
		wants 	  []func(a args, tenant Tenant) bool
	}{
		{"normal", args{name: "test_add_tenant_name"}, false, []func(a args, tenant Tenant) bool{
			func(a args, tenant Tenant) bool{return len(tenant.ID) == 22},
			func(a args, tenant Tenant) bool{return tenant.CreatedAt.Before(time.Now())},
		}},
		{"empty", args{}, true, []func(a args, tenant Tenant) bool{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTenant, err := AddTenant(tt.args.name, tt.args.tenantType, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddTenant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i, assert := range tt.wants {
				if !assert(tt.args, gotTenant) {
					t.Errorf("AddToken() test error, testname = %v, assert %v, args = %v, gotTenant = %v", tt.name, i, tt.args, gotTenant)
				}
			}

		})
	}
}

func TestFindTenantById(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		tenant, _ := AddTenant("tenantName", 1, 0)

		gotTenant, err := FindTenantById(tenant.ID)
		if err != nil {
			t.Errorf("TestFindTenantById() error = %v", err)
			return
		}

		if gotTenant.ID != tenant.ID {
			t.Errorf("TestFindTenantById() want tenant id = %v, got = %v", tenant.ID, gotTenant.ID)
			return
		}
		if gotTenant.Name != tenant.Name {
			t.Errorf("TestFindTenantById() want tenant name = %v, got = %v", tenant.Name, gotTenant.Name)
			return
		}
	})
	t.Run("no result", func(t *testing.T) {
		AddTenant("tenantName", 1, 0)

		gotTenant, err := FindTenantById("dfsaf")
		if err == nil {
			t.Errorf("TestFindTenantById() want err")
			return
		}
		if gotTenant.ID != "" {
			t.Errorf("TestFindTenantById() want empty result, got = %v", gotTenant)
			return
		}
		gotTenant, err = FindTenantById("")
		if err == nil {
			t.Errorf("TestFindTenantById() want err")
			return
		}
		if gotTenant.ID != "" {
			t.Errorf("TestFindTenantById() want empty result, got = %v", gotTenant)
			return
		}

	})

}

func TestFindTenantByName(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		tenant, _ := AddTenant("testTenantName", 1, 0)

		gotTenant, err := FindTenantByName(tenant.Name)
		if err != nil {
			t.Errorf("TestFindTenantByName() error = %v", err)
			return
		}

		if gotTenant.ID != tenant.ID {
			t.Errorf("TestFindTenantByName() want tenant id = %v, got = %v", tenant.ID, gotTenant.ID)
			return
		}
		if gotTenant.Name != tenant.Name {
			t.Errorf("TestFindTenantByName() want tenant name = %v, got = %v", tenant.Name, gotTenant.Name)
			return
		}
	})
	t.Run("no result", func(t *testing.T) {
		AddTenant("tenantName", 1, 0)

		gotTenant, err := FindTenantByName("no_result_name")
		if err == nil {
			t.Errorf("TestFindTenantByName() want err")
			return
		}
		if gotTenant.ID != "" {
			t.Errorf("TestFindTenantByName() want empty result, got = %v", gotTenant)
			return
		}
		gotTenant, err = FindTenantByName("")
		if err == nil {
			t.Errorf("TestFindTenantByName() want err")
			return
		}
		if gotTenant.ID != "" {
			t.Errorf("TestFindTenantByName() want empty result, got = %v", gotTenant)
			return
		}
	})
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

			if tenant.ID == "" || len(tenant.ID) != 22 {
				t.Errorf("BeforeCreate() error, want id length %v, got %v", 22, len(tenant.ID))
			}

			if tenant.Name != "name_test_tenant" {
				t.Errorf("BeforeCreate() want name = %v", "name_test_tenant")
			}

		})
}
