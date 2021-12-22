package tenant

import (
	"context"
	"github.com/pingcap-inc/tiem/models/common"
	"reflect"
	"testing"
)

func TestTenantReadWrite_AddTenant(t *testing.T) {
	type fields struct {
		GormDB common.GormDB
	}
	type args struct {
		ctx        context.Context
		name       string
		tenantType int8
		status     int8
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Tenant
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &TenantReadWrite{
				GormDB: tt.fields.GormDB,
			}
			got, err := g.AddTenant(tt.args.ctx, tt.args.name, tt.args.tenantType, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddTenant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddTenant() got = %v, want %v", got, tt.want)
			}
		})
	}
}
