package domain

import (
	"reflect"
	"testing"
)

func TestCommonStatusFromStatus(t *testing.T) {
	type args struct {
		status int32
	}
	tests := []struct {
		name string
		args args
		want CommonStatus
	}{
		{"valid", args{0}, Valid},
		{"Invalid", args{1}, Invalid},
		{"Deleted", args{2}, Deleted},
		{"UnrecognizedStatus", args{99}, UnrecognizedStatus},
		{"UnrecognizedStatus", args{-1}, UnrecognizedStatus},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CommonStatusFromStatus(tt.args.status); got != tt.want {
				t.Errorf("CommonStatusFromStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommonStatus_IsValid(t *testing.T) {
	tests := []struct {
		name string
		s    CommonStatus
		want bool
	}{
		{"normal", Valid, true},
		{"invalid", Invalid, false},
		{"Deleted", Deleted, false},
		{"UnrecognizedStatus", UnrecognizedStatus, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateTenant(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *Tenant
		wantErr bool
	}{
		{"normal", args{"notExisted"}, &Tenant{Name: "notExisted", Type: InstanceWorkspace}, false},
		{"existed", args{"existed"}, &Tenant{Name: "existed"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateTenant(tt.args.name)
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
		want    *Tenant
		wantErr bool
	}{
		{"normal", args{name: "111"}, &Tenant{Name: "111"}, false},
		{"notExisted", args{name: "notExisted"}, &Tenant{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindTenant(tt.args.name)
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
		want    *Tenant
		wantErr bool
	}{
		{"normal", args{id: "111"}, &Tenant{Id: "111"}, false},
		{"notExisted", args{id: "notExisted"}, &Tenant{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindTenantById(tt.args.id)
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

func TestTenant_persist(t1 *testing.T) {
	type fields struct {
		Name   string
		Id     string
		Type   TenantType
		Status CommonStatus
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"normal", fields{}, false},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Tenant{
				Name:   tt.fields.Name,
				Id:     tt.fields.Id,
				Type:   tt.fields.Type,
				Status: tt.fields.Status,
			}
			if err := t.persist(); (err != nil) != tt.wantErr {
				t1.Errorf("persist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
