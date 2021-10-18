package application

import (
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/domain"
	"reflect"
	"testing"
)

func TestCreateAccount(t *testing.T) {
	type args struct {
		tenant *domain.Tenant
		name   string
		passwd string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(args args, account *domain.Account) bool
	}{
		{"normal", args{&domain.Tenant{Id: "111", Status: domain.Valid}, "name", "password"}, false, []func(args args, account *domain.Account) bool{
			func(args args, account *domain.Account) bool { return true },
		}},
		{"tenantInvalid", args{&domain.Tenant{Id: "111", Status: domain.Invalid}, "name", "password"}, true, []func(args args, account *domain.Account) bool{}},
		{"accountExisted", args{&domain.Tenant{Id: "111", Status: domain.Invalid}, "testMyName", "testMyPassword"}, true, []func(args args, account *domain.Account) bool{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := userManager.CreateAccount(tt.args.tenant, tt.args.name, tt.args.passwd)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.wants {
				if !assert(tt.args, got) {
					t.Errorf("CreateAccount() assert got false, index = %v, got = %v", i, got)
				}
			}
		})
	}
}

func Test_findAccountByName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *domain.Account
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := userManager.FindAccountByName(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAccountByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindAccountByName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRole_empower(t *testing.T) {
	type fields struct {
		TenantId string
		Id       string
		Name     string
		Desc     string
		Status   domain.CommonStatus
	}
	type args struct {
		permissions []domain.Permission
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"normal", fields{TenantId: "111"}, args{[]domain.Permission{{TenantId: "111", Code: "code111"}}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := &domain.Role{
				TenantId: tt.fields.TenantId,
				Id:       tt.fields.Id,
				Name:     tt.fields.Name,
				Desc:     tt.fields.Desc,
				Status:   tt.fields.Status,
			}
			if err := userManager.Empower(role, tt.args.permissions); (err != nil) != tt.wantErr {
				t.Errorf("Empower() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRole_persist(t *testing.T) {
	type fields struct {
		TenantId string
		Id       string
		Name     string
		Desc     string
		Status   domain.CommonStatus
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"normal", fields{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := &domain.Role{
				TenantId: tt.fields.TenantId,
				Id:       tt.fields.Id,
				Name:     tt.fields.Name,
				Desc:     tt.fields.Desc,
				Status:   tt.fields.Status,
			}

			if err := userManager.rbacRepo.AddRole(role); (err != nil) != tt.wantErr {
				t.Errorf("persist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createRole(t *testing.T) {
	type args struct {
		tenant *domain.Tenant
		name   string
		desc   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		asserts []func(args args, role *domain.Role) bool
	}{
		{"normal", args{&domain.Tenant{Id: "tenantId", Status: domain.Valid}, "notExisted", "role1"}, false, []func(args args, role *domain.Role) bool{
			func(args args, role *domain.Role) bool { return len(role.Id) > 0 },
			func(args args, role *domain.Role) bool { return role.Name == "notExisted" },
		}},
		{"existed", args{&domain.Tenant{Id: "tenantId", Status: domain.Valid}, "existed", "role1"}, true, []func(args args, role *domain.Role) bool{}},
		{"empty name", args{&domain.Tenant{Id: "tenantId", Status: domain.Valid}, "", "role1"}, true, []func(args args, role *domain.Role) bool{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := userManager.CreateRole(tt.args.tenant, tt.args.name, tt.args.desc)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.asserts {
				if !assert(tt.args, got) {
					t.Errorf("FindRoleByName() assert got false, index = %v, got = %v", i, got)
				}
			}
		})
	}
}

func Test_findRoleByName(t *testing.T) {
	type args struct {
		tenant *domain.Tenant
		name   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		asserts []func(args args, role *domain.Role) bool
	}{
		{"normal", args{tenant: &domain.Tenant{Id: "111"}, name: "existed"}, false, []func(args args, role *domain.Role) bool{
			func(args args, role *domain.Role) bool { return role.Name == "existed" },
			func(args args, role *domain.Role) bool { return len(role.Id) > 0 },
		}},
		{"not existed", args{tenant: &domain.Tenant{Id: "111"}, name: "notExisted"}, true, []func(args args, role *domain.Role) bool{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := userManager.FindRoleByName(tt.args.tenant, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindRoleByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.asserts {
				if !assert(tt.args, got) {
					t.Errorf("FindRoleByName() assert got false, index = %v, got = %v", i, got)
				}
			}
		})
	}
}
