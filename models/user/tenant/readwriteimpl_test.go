package tenant

import (
	"context"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTenantReadWrite_AddTenant(t *testing.T) {
	type args struct {
		ctx        context.Context
		name       string
		tenantType int8
		status     int8
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{context.TODO(), "tenant", 0, 0}, false},
		{"without name", args{context.TODO(), "", 1, 0}, true},
		{"invalid status", args{context.TODO(), "tenant", 1, 1}, true},
		{"without status", args{context.TODO(), "tenant", 1, -1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testRW.AddTenant(tt.args.ctx, tt.args.name, tt.args.tenantType, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddTenant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.NotEmpty(t, got.ID)
			}
		})
	}
}

func TestAddTenant(t *testing.T) {
	type args struct {
		name       string
		tenantType int8
		status     int8
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, tenant Tenant) bool
	}{
		{"normal", args{name: "test_add_tenant_name"}, false, []func(a args, tenant Tenant) bool{
			func(a args, tenant Tenant) bool { return len(tenant.ID) == uuidutil.ENTITY_UUID_LENGTH },
			func(a args, tenant Tenant) bool { return tenant.CreatedAt.Before(time.Now()) },
		}},
		{"empty", args{}, true, []func(a args, tenant Tenant) bool{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTenant, err := testRW.AddTenant(context.TODO(), tt.args.name, tt.args.tenantType, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddTenant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i, assert := range tt.wants {
				if !assert(tt.args, *gotTenant) {
					t.Errorf("AddToken() test error, testname = %v, assert %v, args = %v, gotTenant = %v", tt.name, i, tt.args, gotTenant)
				}
			}
		})
	}
}


func TestTenantReadWrite_FindTenantByName(t *testing.T) {
	tenant := Tenant{
		Name:      "testName",
	}
	testRW.DB(context.TODO()).Create(&tenant)
	defer testRW.DB(context.TODO()).Delete(&tenant)

	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{context.TODO(), "testName"}, false},
		{"no record", args{context.TODO(), "test"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testRW.FindTenantByName(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindTenantByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.NotEmpty(t, got.ID)
			}
		})
	}
}

func TestTenantReadWrite_FindTenantById(t *testing.T) {
	tenant := Tenant{
		ID: "testID",
		Name:      "testName",
	}
	testRW.DB(context.TODO()).Create(&tenant)
	defer testRW.DB(context.TODO()).Delete(&tenant)

	type args struct {
		ctx      context.Context
		tenantId string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{context.TODO(), tenant.ID}, false},
		{"no record", args{context.TODO(), "test"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := testRW.FindTenantById(tt.args.ctx, tt.args.tenantId)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindTenantById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.NotEmpty(t, got.ID)
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
		if err != nil {
			t.Errorf("BeforeCreate() error = %v, wantErr nil", err)
		}

		if tenant.Status != 0 {
			t.Errorf("BeforeCreate() error, want status %v, got %v", 0, tenant.Status)
		}

		if tenant.ID == "" || len(tenant.ID) != uuidutil.ENTITY_UUID_LENGTH {
			t.Errorf("BeforeCreate() error, want id length %v, got %v", uuidutil.ENTITY_UUID_LENGTH, len(tenant.ID))
		}

		if tenant.Name != "name_test_tenant" {
			t.Errorf("BeforeCreate() want name = %v", "name_test_tenant")
		}

	})
}