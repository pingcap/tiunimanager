
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package domain

import (
	"testing"
)

func TestRole_empower(t *testing.T) {
	type fields struct {
		TenantId string
		Id       string
		Name     string
		Desc     string
		Status   CommonStatus
	}
	type args struct {
		permissions []Permission
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"normal", fields{TenantId: "111"}, args{[]Permission{{TenantId: "111", Code: "code111"}}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := &Role{
				TenantId: tt.fields.TenantId,
				Id:       tt.fields.Id,
				Name:     tt.fields.Name,
				Desc:     tt.fields.Desc,
				Status:   tt.fields.Status,
			}
			if err := role.empower(tt.args.permissions); (err != nil) != tt.wantErr {
				t.Errorf("empower() error = %v, wantErr %v", err, tt.wantErr)
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
		Status   CommonStatus
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
			role := &Role{
				TenantId: tt.fields.TenantId,
				Id:       tt.fields.Id,
				Name:     tt.fields.Name,
				Desc:     tt.fields.Desc,
				Status:   tt.fields.Status,
			}
			if err := role.persist(); (err != nil) != tt.wantErr {
				t.Errorf("persist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createRole(t *testing.T) {
	type args struct {
		tenant *Tenant
		name   string
		desc   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		asserts []func(args args, role *Role) bool
	}{
		{"normal", args{&Tenant{Id: "tenantId", Status: Valid}, "notExisted", "role1"}, false, []func(args args, role *Role) bool{
			func(args args, role *Role) bool { return len(role.Id) > 0 },
			func(args args, role *Role) bool { return role.Name == "notExisted" },
		}},
		{"existed", args{&Tenant{Id: "tenantId", Status: Valid}, "existed", "role1"}, true, []func(args args, role *Role) bool{}},
		{"empty name", args{&Tenant{Id: "tenantId", Status: Valid}, "", "role1"}, true, []func(args args, role *Role) bool{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createRole(tt.args.tenant, tt.args.name, tt.args.desc)
			if (err != nil) != tt.wantErr {
				t.Errorf("createRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.asserts {
				if !assert(tt.args, got) {
					t.Errorf("findRoleByName() assert got false, index = %v, got = %v", i, got)
				}
			}
		})
	}
}

func Test_findRoleByName(t *testing.T) {
	type args struct {
		tenant *Tenant
		name   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		asserts []func(args args, role *Role) bool
	}{
		{"normal", args{tenant: &Tenant{Id: "111"}, name: "existed"}, false, []func(args args, role *Role) bool{
			func(args args, role *Role) bool { return role.Name == "existed" },
			func(args args, role *Role) bool { return len(role.Id) > 0 },
		}},
		{"not existed", args{tenant: &Tenant{Id: "111"}, name: "notExisted"}, true, []func(args args, role *Role) bool{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findRoleByName(tt.args.tenant, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("findRoleByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.asserts {
				if !assert(tt.args, got) {
					t.Errorf("findRoleByName() assert got false, index = %v, got = %v", i, got)
				}
			}
		})
	}
}
