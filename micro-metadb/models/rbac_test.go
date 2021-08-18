package models

import (
	"reflect"
	"testing"
	"time"
)

func TestAddAccount(t *testing.T) {
	type args struct {
		tenantId  string
		name      string
		salt      string
		finalHash string
		status    int8
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wants      []func (args args, cluster *AccountDO) bool
	}{
		{"normal", args{tenantId: defaultTenantId, name: "TestAddAccount", salt: "TestAddAccount_salt", finalHash: "TestAddAccount_finalHash"},
			false,
			[]func (args args, cluster *AccountDO) bool{
				func (args args, cluster *AccountDO) bool{return len(cluster.ID) == ID_LENGTH},
				func (args args, cluster *AccountDO) bool{return cluster.Status == 0},
				func (args args, cluster *AccountDO) bool{return args.name == cluster.Name},
				func (args args, cluster *AccountDO) bool{return args.tenantId == cluster.TenantId},
				func (args args, cluster *AccountDO) bool{return args.salt == cluster.Salt},
				func (args args, cluster *AccountDO) bool{return args.finalHash == cluster.FinalHash},
				func (args args, cluster *AccountDO) bool{return cluster.CreatedAt.Add(time.Second + 2).After(time.Now())},
			},
		},
		{"without name", args{tenantId: defaultTenantId, salt: "TestAddAccount_salt", finalHash: "TestAddAccount_finalHash"},
			true,
			[]func (args args, cluster *AccountDO) bool{
			},
		},
		{"without tenantId", args{name: "TestAddAccount", salt: "TestAddAccount_salt", finalHash: "TestAddAccount_finalHash"},
			true,
			[]func (args args, cluster *AccountDO) bool{},
		},
		{"without salt", args{tenantId: defaultTenantId, name: "TestAddAccount", finalHash: "TestAddAccount_finalHash"},
			true,
			[]func (args args, cluster *AccountDO) bool{},
		},
		{"without finalHash", args{tenantId: defaultTenantId, name: "TestAddAccount", salt: "TestAddAccount_salt"},
			true,
			[]func (args args, cluster *AccountDO) bool{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := AddAccount(tt.args.tenantId, tt.args.name, tt.args.salt, tt.args.finalHash, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i, assert := range tt.wants {
				if !assert(tt.args, &gotResult) {
					t.Errorf("AddAccount() test error, testname = %v, assert %v, args = %v, gotResult = %v", tt.name, i, tt.args, gotResult)
				}
			}

		})
	}
}

func TestAddPermission(t *testing.T) {
	type args struct {
		tenantId       string
		code           string
		name           string
		desc           string
		permissionType int8
		status         int8
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wants      []func (args args, p *PermissionDO) bool
	}{
		{"normal", args{tenantId: defaultTenantId, code: "TestAddPermission_code_normal", name: "TestAddPermission_name_normal", desc: "desc", permissionType: 1, status: 99},
			false,
			[]func (args args, p *PermissionDO) bool{
				func (args args, p *PermissionDO) bool{return len(p.ID) == ID_LENGTH},
				func (args args, p *PermissionDO) bool{return p.Code == args.code},
				func (args args, p *PermissionDO) bool{return p.Name == args.name},
				func (args args, p *PermissionDO) bool{return p.Desc == args.desc},
				func (args args, p *PermissionDO) bool{return p.Type == args.permissionType},
				func (args args, p *PermissionDO) bool{return p.Status == 0},
				func (args args, p *PermissionDO) bool{return p.CreatedAt.Add(time.Second + 2).After(time.Now())},
			},
		},
		{"without name", args{tenantId: defaultTenantId, code: "TestAddPermission_code_without_name", desc: "desc", permissionType: 1, status: 99},
			true,
			[]func (args args, p *PermissionDO) bool{},
		},
		{"without tenantId", args{code: "TestAddPermission_code_withoutTenantId", name: "TestAddPermission_name_withoutTenantId", desc: "desc", permissionType: 1, status: 99},
			true,
			[]func (args args, p *PermissionDO) bool{},
		},
		{"without code", args{tenantId: defaultTenantId, name: "TestAddPermission_name_without_code", desc: "desc", permissionType: 1, status: 99},
			true,
			[]func (args args, p *PermissionDO) bool{},
		},
		{"without desc", args{tenantId: defaultTenantId, code: "TestAddPermission_code_without_desc", name: "TestAddPermission_name_without_desc", permissionType: 1, status: 99},
			false,
			[]func (args args, p *PermissionDO) bool{
				func (args args, p *PermissionDO) bool{return len(p.ID) == ID_LENGTH},
				func (args args, p *PermissionDO) bool{return p.CreatedAt.Add(time.Second + 2).After(time.Now())},
			},
		},
		{"without permissionType", args{tenantId: defaultTenantId, code: "TestAddPermission_code_without_type", name: "TestAddPermission_name_without_type", desc: "desc", status: 99},
			false,
			[]func (args args, p *PermissionDO) bool{
				func (args args, p *PermissionDO) bool{return len(p.ID) == ID_LENGTH},
				func (args args, p *PermissionDO) bool{return p.Type == 0},
				func (args args, p *PermissionDO) bool{return p.CreatedAt.Add(time.Second + 2).After(time.Now())},
			},
		},
		{"without status", args{tenantId: defaultTenantId, code: "TestAddPermission_code_without_status", name: "TestAddPermission_name_without_status", desc: "desc", permissionType: 1},
			false,
			[]func (args args, p *PermissionDO) bool{
				func (args args, p *PermissionDO) bool{return p.Status == 0},
				func (args args, p *PermissionDO) bool{return p.CreatedAt.Add(time.Second + 2).After(time.Now())},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := AddPermission(tt.args.tenantId, tt.args.code, tt.args.name, tt.args.desc, tt.args.permissionType, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i, assert := range tt.wants {
				if !assert(tt.args, &gotResult) {
					t.Errorf("AddPermission() test error, testname = %v, assert %v, args = %v, gotResult = %v", tt.name, i, tt.args, gotResult)
				}
			}

		})
	}
}

func TestAddPermissionBindings(t *testing.T) {
	type args struct {
		bindings []PermissionBindingDO
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{[]PermissionBindingDO{
			{Entity: Entity{TenantId: defaultTenantId}, RoleId: "TestAddPermissionBindings_roleId", PermissionId: "TestAddPermissionBindings_PermissionId"},
		}}, false},
		{"without tenantId", args{[]PermissionBindingDO{
			{RoleId: "TestAddPermissionBindings_roleId_withoutTenantId", PermissionId: "TestAddPermissionBindings_PermissionId_withoutTenantId"},
		}}, true},
		{"empty roleId", args{[]PermissionBindingDO{
			{Entity: Entity{TenantId: defaultTenantId}, PermissionId: "TestAddPermissionBindings_PermissionId_emptyRoleId"},
		}}, true},
		{"empty permissionId", args{[]PermissionBindingDO{
			{Entity: Entity{TenantId: defaultTenantId}, RoleId: "TestAddPermissionBindings_roleId_emptyPermissionId"},
		}}, true},
		{"batch", args{[]PermissionBindingDO{
			{Entity: Entity{TenantId: defaultTenantId}, RoleId: "batch_roleId1", PermissionId: "batch_permissionId1"},
			{Entity: Entity{TenantId: defaultTenantId}, RoleId: "batch_roleId1", PermissionId: "batch_permissionId2"},
			{Entity: Entity{TenantId: defaultTenantId}, RoleId: "batch_roleId1", PermissionId: "batch_permissionId3"},
			{Entity: Entity{TenantId: defaultTenantId}, RoleId: "batch_roleId2", PermissionId: "batch_permissionId1"},
			{Entity: Entity{TenantId: defaultTenantId}, RoleId: "batch_roleId2", PermissionId: "batch_permissionId2"},
			{Entity: Entity{TenantId: defaultTenantId}, RoleId: "batch_roleId3", PermissionId: "batch_permissionId1"},
		}}, false},
		{"conflict", args{[]PermissionBindingDO{
			{Entity: Entity{TenantId: defaultTenantId}, RoleId: "conflict_roleId1", PermissionId: "conflict_permissionId1"},
			{Entity: Entity{TenantId: defaultTenantId}, RoleId: "conflict_roleId1", PermissionId: "conflict_permissionId2"},
			{Entity: Entity{TenantId: defaultTenantId}, RoleId: "conflict_roleId2", PermissionId: "conflict_permissionId1"},
			{Entity: Entity{TenantId: defaultTenantId}, RoleId: "conflict_roleId2", PermissionId: "conflict_permissionId2"},
			{Entity: Entity{TenantId: defaultTenantId}, RoleId: "conflict_roleId1", PermissionId: "conflict_permissionId1"},
		}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AddPermissionBindings(tt.args.bindings); (err != nil) != tt.wantErr {
				t.Errorf("AddPermissionBindings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddRole(t *testing.T) {
	type args struct {
		tenantId string
		name     string
		desc     string
		status   int8
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wants      []func (args args, p *RoleDO) bool
	}{
		{"normal", args{tenantId: defaultTenantId, name: "TestAddPermission_name_normal", desc: "desc", status: 99},
			false,
			[]func (args args, r *RoleDO) bool{
				func (args args, r *RoleDO) bool{return len(r.ID) == ID_LENGTH},
				func (args args, r *RoleDO) bool{return r.Name == args.name},
				func (args args, r *RoleDO) bool{return r.Desc == args.desc},
				func (args args, r *RoleDO) bool{return r.Status == 0},
				func (args args, r *RoleDO) bool{return r.CreatedAt.Add(time.Second + 2).After(time.Now())},
			},
		},
		{"without name", args{tenantId: defaultTenantId, desc: "desc", status: 99},
			true,
			[]func (args args, p *RoleDO) bool{},
		},
		{"without tenantId", args{name: "TestAddPermission_name_withoutTenantId", desc: "desc", status: 99},
			true,
			[]func (args args, p *RoleDO) bool{},
		},
		{"without desc", args{tenantId: defaultTenantId, name: "TestAddPermission_name_withoutDesc", status: 99},
			false,
			[]func (args args, p *RoleDO) bool{
				func (args args, p *RoleDO) bool{return len(p.ID) == ID_LENGTH},
				func (args args, p *RoleDO) bool{return p.CreatedAt.Add(time.Second + 2).After(time.Now())},
			},
		},
		{"without status", args{tenantId: defaultTenantId, name: "TestAddPermission_name_withoutStatus", desc: "desc", status: 99},
			false,
			[]func (args args, p *RoleDO) bool{
				func (args args, p *RoleDO) bool{return p.Status == 0},
				func (args args, p *RoleDO) bool{return p.CreatedAt.Add(time.Second + 2).After(time.Now())},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := AddRole(tt.args.tenantId, tt.args.name, tt.args.desc, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i, assert := range tt.wants {
				if !assert(tt.args, &gotResult) {
					t.Errorf("AddPermission() test error, testname = %v, assert %v, args = %v, gotResult = %v", tt.name, i, tt.args, gotResult)
				}
			}

		})
	}
}

func TestAddRoleBindings(t *testing.T) {
	type args struct {
		bindings []RoleBindingDO
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
			if err := AddRoleBindings(tt.args.bindings); (err != nil) != tt.wantErr {
				t.Errorf("AddRoleBindings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFetchAllRolesByAccount(t *testing.T) {
	type args struct {
		tenantId  string
		accountId string
	}
	tests := []struct {
		name       string
		args       args
		wantResult []RoleDO
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := FetchAllRolesByAccount(tt.args.tenantId, tt.args.accountId)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchAllRolesByAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("FetchAllRolesByAccount() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestFetchAllRolesByPermission(t *testing.T) {
	type args struct {
		tenantId     string
		permissionId string
	}
	tests := []struct {
		name       string
		args       args
		wantResult []RoleDO
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := FetchAllRolesByPermission(tt.args.tenantId, tt.args.permissionId)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchAllRolesByPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("FetchAllRolesByPermission() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestFetchPermission(t *testing.T) {
	type args struct {
		tenantId string
		code     string
	}
	tests := []struct {
		name       string
		args       args
		wantResult PermissionDO
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := FetchPermission(tt.args.tenantId, tt.args.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("FetchPermission() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestFetchRole(t *testing.T) {
	type args struct {
		tenantId string
		name     string
	}
	tests := []struct {
		name       string
		args       args
		wantResult RoleDO
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := FetchRole(tt.args.tenantId, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("FetchRole() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestFetchRolesByIds(t *testing.T) {
	type args struct {
		roleIds []string
	}
	tests := []struct {
		name       string
		args       args
		wantResult []RoleDO
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := FetchRolesByIds(tt.args.roleIds)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchRolesByIds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("FetchRolesByIds() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestFindAccount(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name       string
		args       args
		wantResult AccountDO
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := FindAccount(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("FindAccount() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
