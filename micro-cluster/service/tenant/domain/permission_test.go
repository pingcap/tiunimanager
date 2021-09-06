package domain

import (
	"reflect"
	"testing"
)

func TestPermissionTypeFromType(t *testing.T) {
	type args struct {
		pType int32
	}
	tests := []struct {
		name string
		args args
		want PermissionType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PermissionTypeFromType(tt.args.pType); got != tt.want {
				t.Errorf("PermissionTypeFromType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermission_listAllRoles(t *testing.T) {
	type fields struct {
		TenantId string
		Code     string
		Name     string
		Type     PermissionType
		Desc     string
		Status   CommonStatus
	}
	tests := []struct {
		name    string
		fields  fields
		want    []Role
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permission := &Permission{
				TenantId: tt.fields.TenantId,
				Code:     tt.fields.Code,
				Name:     tt.fields.Name,
				Type:     tt.fields.Type,
				Desc:     tt.fields.Desc,
				Status:   tt.fields.Status,
			}
			got, err := permission.listAllRoles()
			if (err != nil) != tt.wantErr {
				t.Errorf("listAllRoles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listAllRoles() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermission_persist(t *testing.T) {
	type fields struct {
		TenantId string
		Code     string
		Name     string
		Type     PermissionType
		Desc     string
		Status   CommonStatus
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
			permission := &Permission{
				TenantId: tt.fields.TenantId,
				Code:     tt.fields.Code,
				Name:     tt.fields.Name,
				Type:     tt.fields.Type,
				Desc:     tt.fields.Desc,
				Status:   tt.fields.Status,
			}
			if err := permission.persist(); (err != nil) != tt.wantErr {
				t.Errorf("persist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createPermission(t *testing.T) {
	type args struct {
		tenant         *Tenant
		code           string
		name           string
		desc           string
		permissionType PermissionType
	}
	tests := []struct {
		name    string
		args    args
		want    *Permission
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createPermission(tt.args.tenant, tt.args.code, tt.args.name, tt.args.desc, tt.args.permissionType)
			if (err != nil) != tt.wantErr {
				t.Errorf("createPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createPermission() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findPermissionByCode(t *testing.T) {
	type args struct {
		tenantId string
		code     string
	}
	tests := []struct {
		name    string
		args    args
		want    *Permission
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findPermissionByCode(tt.args.tenantId, tt.args.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("findPermissionByCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findPermissionByCode() got = %v, want %v", got, tt.want)
			}
		})
	}
}
